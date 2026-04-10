package main

import "github.com/awafinance/fiscal/internal/codegen/postprocess"

type generatedWorkaround struct {
	name  string
	apply postprocess.Replacement
}

type generatedPostprocessSpec struct {
	genDir               string
	nestedImportPatterns []string
	removeFile           func(path string) (bool, string)
	workarounds          []generatedWorkaround
	addJSONTags          bool
}

func runGeneratedPostprocess(verbose bool, spec generatedPostprocessSpec) error {
	replacements := make([]postprocess.Replacement, 0, len(spec.workarounds))
	for _, workaround := range spec.workarounds {
		replacements = append(replacements, workaround.apply)
	}

	return postprocess.Generated(postprocess.Options{
		GenDir:               spec.genDir,
		NestedImportPatterns: spec.nestedImportPatterns,
		RemoveFile:           spec.removeFile,
		Replacements:         replacements,
		AddJSONTags:          spec.addJSONTags,
		Verbose:              verbose,
	})
}

func workaround(name string, apply postprocess.Replacement) generatedWorkaround {
	return generatedWorkaround{
		name:  name,
		apply: apply,
	}
}
