package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const xmlDSigSignatureTag = "`xml:\"ds:Signature\"`"
const xDescInterfaceField = "interface{} `xml:\"xDesc\"`"
const xDescStringField = "string `xml:\"xDesc\"`"

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: go run ./internal/nfse/tools/codegen <postprocess-generated>")
	}

	switch os.Args[1] {
	case "postprocess-generated":
		if err := postprocessGenerated(); err != nil {
			fatalf("postprocess generated: %v", err)
		}
	default:
		fatalf("unknown subcommand %q", os.Args[1])
	}
}

func postprocessGenerated() error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	root := filepath.Join(repoRoot, "internal", "nfse", "gen")
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		if isNestedImportedSchema(path) {
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Printf("removed duplicated imported schema package %s\n", path)
			return nil
		}

		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		updated := strings.ReplaceAll(string(text), xmlDSigSignatureTag, "`xml:\"http://www.w3.org/2000/09/xmldsig# Signature\"`")
		updated = strings.ReplaceAll(updated, xDescInterfaceField, xDescStringField)
		if updated == string(text) {
			return nil
		}

		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return err
		}

		fmt.Printf("postprocessed generated xml tags in %s\n", path)
		return nil
	})
}

func isNestedImportedSchema(path string) bool {
	clean := filepath.Clean(path)
	patterns := []string{
		string(filepath.Separator) + filepath.Join("nfelib", "nfelib", "nfse", "schemas") + string(filepath.Separator),
		string(filepath.Separator) + filepath.Join("internal", "nfse", "schemas") + string(filepath.Separator),
		string(filepath.Separator) + filepath.Join("internal", "nfse", "gen") + string(filepath.Separator) + "v1_0" + string(filepath.Separator) + "schemas" + string(filepath.Separator),
	}
	for _, pattern := range patterns {
		if strings.Contains(clean, pattern) {
			return true
		}
	}
	return false
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

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
