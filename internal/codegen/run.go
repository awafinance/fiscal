package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

type family struct {
	name          string
	normalizeDirs []string
	xgenJobs      []xgenJob
	postprocess   func(verbose bool) error
}

type xgenJob struct {
	input  string
	output string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "codegen: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	defaultJobs := min(8, max(1, runtime.NumCPU()))

	flags := flag.NewFlagSet("codegen", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	jobs := flags.Int("j", defaultJobs, "maximum parallel xgen jobs per family")
	list := flags.Bool("list", false, "list available codegen families")
	verbose := flags.Bool("v", false, "print detailed codegen progress")
	if err := flags.Parse(args); err != nil {
		return err
	}

	families := codegenFamilies()
	if *list {
		for _, family := range families {
			fmt.Println(family.name)
		}
		return nil
	}

	selected, err := selectFamilies(families, flags.Args())
	if err != nil {
		return err
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	start := time.Now()
	ctx := context.Background()
	xgenSlots := make(chan struct{}, max(1, *jobs))
	if err := runFamilies(ctx, repoRoot, selected, max(1, *jobs), xgenSlots, *verbose); err != nil {
		return err
	}
	fmt.Printf("[gen] completed %d family(s) in %s\n", len(selected), time.Since(start).Round(time.Millisecond))
	return nil
}

func runFamilies(ctx context.Context, repoRoot string, families []family, parallelism int, xgenSlots chan struct{}, verbose bool) error {
	if len(families) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(families))
	var wg sync.WaitGroup
	for _, family := range families {
		family := family
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runFamily(ctx, repoRoot, family, parallelism, xgenSlots, verbose); err != nil {
				errCh <- err
				cancel()
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var canceled error
	for err := range errCh {
		if errors.Is(err, context.Canceled) {
			canceled = err
			continue
		}
		return err
	}
	return canceled
}

func selectFamilies(families []family, names []string) ([]family, error) {
	if len(names) == 0 || slices.Contains(names, "all") {
		return families, nil
	}

	byName := make(map[string]family, len(families))
	for _, family := range families {
		byName[family.name] = family
	}

	selected := make([]family, 0, len(names))
	for _, name := range names {
		family, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown family %q; available: %s", name, familyNames(families))
		}
		selected = append(selected, family)
	}
	return selected, nil
}

func familyNames(families []family) string {
	names := make([]string, 0, len(families))
	for _, family := range families {
		names = append(names, family.name)
	}
	return strings.Join(names, ", ")
}

func runFamily(ctx context.Context, repoRoot string, family family, parallelism int, xgenSlots chan struct{}, verbose bool) error {
	start := time.Now()
	if verbose {
		fmt.Printf("[gen] %s normalize %d schema dir(s)\n", family.name, len(family.normalizeDirs))
	}
	if len(family.normalizeDirs) > 0 {
		stats, err := normalizeSchemas(repoRoot, family.normalizeDirs)
		if err != nil {
			return fmt.Errorf("%s normalize schemas: %w", family.name, err)
		}
		if verbose {
			fmt.Printf("[gen] %s normalized %d xsd file(s), %d inline type(s), %d optional sequence(s), %d rewritten\n",
				family.name, stats.files, stats.generatedTypes, stats.flattenedOptionalSequences, stats.rewritten)
		}
	}

	if verbose {
		fmt.Printf("[gen] %s xgen %d job(s), parallelism %d\n", family.name, len(family.xgenJobs), min(parallelism, max(1, len(family.xgenJobs))))
	}
	if err := runXGenJobs(ctx, repoRoot, family, parallelism, xgenSlots, verbose); err != nil {
		return fmt.Errorf("%s xgen: %w", family.name, err)
	}

	if verbose {
		fmt.Printf("[gen] %s postprocess\n", family.name)
	}
	if err := family.postprocess(verbose); err != nil {
		return fmt.Errorf("%s postprocess generated: %w", family.name, err)
	}

	if verbose {
		fmt.Printf("[gen] %s done in %s\n", family.name, time.Since(start).Round(time.Millisecond))
	}
	return nil
}

func runXGenJobs(ctx context.Context, repoRoot string, family family, parallelism int, xgenSlots chan struct{}, verbose bool) error {
	if len(family.xgenJobs) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobCh := make(chan xgenJob)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	workers := min(parallelism, len(family.xgenJobs))
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				if err := runXGenJob(ctx, repoRoot, family.name, job, xgenSlots, verbose); err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
					return
				}
			}
		}()
	}

dispatch:
	for _, job := range family.xgenJobs {
		select {
		case <-ctx.Done():
			break dispatch
		case jobCh <- job:
		}
	}
	close(jobCh)
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}
}

func runXGenJob(ctx context.Context, repoRoot, familyName string, job xgenJob, xgenSlots chan struct{}, verbose bool) error {
	input := filepath.Join("..", "schemas", job.input)
	output := filepath.Join(".", job.output)
	genDir := filepath.Join(repoRoot, "internal", familyName, "gen")

	start := time.Now()
	select {
	case xgenSlots <- struct{}{}:
		defer func() { <-xgenSlots }()
	case <-ctx.Done():
		return ctx.Err()
	}

	cmd := exec.CommandContext(ctx, "xgen", "-i", input, "-o", output, "-l", "Go")
	cmd.Dir = genDir

	var stderr bytes.Buffer
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s -> %s: %w\n%s", input, output, err, strings.TrimSpace(stderr.String()))
	}

	if verbose {
		fmt.Printf("[gen] %s xgen %s -> %s in %s\n", familyName, job.input, job.output, time.Since(start).Round(time.Millisecond))
	}
	return nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root from %s", dir)
		}
		dir = parent
	}
}
