package postprocess

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const xmlDSigSignatureTag = "`xml:\"ds:Signature\"`"
const xmlDSigSignatureNamespaceTag = "`xml:\"http://www.w3.org/2000/09/xmldsig# Signature\"`"

var anonComplexXMLName = regexp.MustCompile("`xml:\"TAnonComplex_([^\"_]+(?:_[^\"_]+)*)_\\d+\"`")

type Replacement func(path, text string) string
type FieldMatcher func(path string, field *ast.Field) bool

type Options struct {
	GenDir               string
	NestedImportPatterns []string
	RemoveFile           func(path string) (bool, string)
	Replacements         []Replacement
	AddJSONTags          bool
	Verbose              bool
}

func Generated(opts Options) error {
	repoRoot, err := FindRepoRoot()
	if err != nil {
		return err
	}

	root := opts.GenDir
	if !filepath.IsAbs(root) {
		root = filepath.Join(repoRoot, root)
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		if opts.RemoveFile != nil {
			remove, message := opts.RemoveFile(path)
			if remove {
				if err := os.Remove(path); err != nil {
					return err
				}
				if opts.Verbose {
					fmt.Printf("%s %s\n", message, path)
				}
				return nil
			}
		}

		if isNestedImportedSchema(path, opts.NestedImportPatterns) {
			if err := os.Remove(path); err != nil {
				return err
			}
			if opts.Verbose {
				fmt.Printf("removed duplicated imported schema package %s\n", path)
			}
			return nil
		}

		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		updated := strings.ReplaceAll(string(text), xmlDSigSignatureTag, xmlDSigSignatureNamespaceTag)
		updated = anonComplexXMLName.ReplaceAllString(updated, "`xml:\"$1\"`")
		for _, replacement := range opts.Replacements {
			updated = replacement(path, updated)
		}

		changedTags := false
		if opts.AddJSONTags {
			updatedBytes, tagsChanged, err := addJSONTags(path, []byte(updated))
			if err != nil {
				return err
			}
			updated = string(updatedBytes)
			changedTags = tagsChanged
		}

		if updated == string(text) {
			return nil
		}

		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return err
		}

		if changedTags {
			if opts.Verbose {
				fmt.Printf("postprocessed generated xml/json tags in %s\n", path)
			}
		} else {
			if opts.Verbose {
				fmt.Printf("postprocessed generated xml tags in %s\n", path)
			}
		}
		return nil
	})
}

func ReplaceAll(old, new string) Replacement {
	return func(_ string, text string) string {
		return strings.ReplaceAll(text, old, new)
	}
}

func Replace(old, new string, n int) Replacement {
	return func(_ string, text string) string {
		return strings.Replace(text, old, new, n)
	}
}

func RegexReplaceAll(re *regexp.Regexp, replacement string) Replacement {
	return func(_ string, text string) string {
		return re.ReplaceAllString(text, replacement)
	}
}

func ReplaceFieldType(match FieldMatcher, replacementType string) Replacement {
	return func(path string, text string) string {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, text, parser.ParseComments)
		if err != nil {
			return text
		}

		replacementExpr, err := parser.ParseExpr(replacementType)
		if err != nil {
			return text
		}

		changed := false
		ast.Inspect(file, func(node ast.Node) bool {
			structType, ok := node.(*ast.StructType)
			if !ok {
				return true
			}

			for _, field := range structType.Fields.List {
				if !match(path, field) {
					continue
				}
				field.Type = cloneExpr(replacementExpr)
				changed = true
			}

			return false
		})

		if !changed {
			return text
		}

		var buf bytes.Buffer
		if err := printer.Fprint(&buf, fset, file); err != nil {
			return text
		}
		return buf.String()
	}
}

func FieldNamed(names ...string) FieldMatcher {
	return func(_ string, field *ast.Field) bool {
		for _, name := range names {
			if fieldHasName(field, name) {
				return true
			}
		}
		return false
	}
}

func FieldTypeEquals(typeExpr string) FieldMatcher {
	return func(_ string, field *ast.Field) bool {
		return exprString(field.Type) == typeExpr
	}
}

func AllFields(matchers ...FieldMatcher) FieldMatcher {
	return func(path string, field *ast.Field) bool {
		for _, matcher := range matchers {
			if !matcher(path, field) {
				return false
			}
		}
		return true
	}
}

func IfPath(match func(path string) bool, replacements ...Replacement) Replacement {
	return func(path string, text string) string {
		if !match(path) {
			return text
		}
		for _, replacement := range replacements {
			text = replacement(path, text)
		}
		return text
	}
}

func PathContains(elem ...string) func(path string) bool {
	pattern := string(filepath.Separator) + filepath.Join(elem...) + string(filepath.Separator)
	return func(path string) bool {
		return strings.Contains(filepath.Clean(path), pattern)
	}
}

func FindRepoRoot() (string, error) {
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

func isNestedImportedSchema(path string, patterns []string) bool {
	clean := filepath.Clean(path)
	for _, pattern := range patterns {
		if strings.Contains(clean, pattern) {
			return true
		}
	}
	return false
}

func addJSONTags(path string, src []byte) ([]byte, bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	changed := false
	ast.Inspect(file, func(node ast.Node) bool {
		structType, ok := node.(*ast.StructType)
		if !ok {
			return true
		}

		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}

			rawTag, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				continue
			}

			updatedTag, ok := addJSONTag(rawTag, field)
			if !ok {
				continue
			}

			field.Tag.Value = "`" + updatedTag + "`"
			changed = true
		}

		return false
	})

	if !changed {
		return src, false, nil
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return nil, false, err
	}

	return buf.Bytes(), true, nil
}

func addJSONTag(rawTag string, field *ast.Field) (string, bool) {
	tag := reflect.StructTag(rawTag)
	if tag.Get("json") != "" {
		return rawTag, false
	}

	xmlTag := tag.Get("xml")
	if xmlTag == "" {
		return rawTag, false
	}

	jsonName := jsonNameFromXMLTag(xmlTag, field)
	if jsonName == "" {
		return rawTag, false
	}
	if jsonName != "-" {
		jsonName += ",omitempty"
	}

	return rawTag + ` json:"` + jsonName + `"`, true
}

func jsonNameFromXMLTag(xmlTag string, field *ast.Field) string {
	if xmlTag == "-" || fieldHasName(field, "XMLName") {
		return "-"
	}

	name, options, _ := strings.Cut(xmlTag, ",")
	if name == "" {
		switch optionSet := strings.Split(options, ","); {
		case slices.Contains(optionSet, "chardata"):
			return "value"
		case slices.Contains(optionSet, "innerxml"):
			return "innerXML"
		}
	}

	if idx := strings.LastIndexByte(name, ' '); idx >= 0 {
		name = name[idx+1:]
	}

	return name
}

func fieldHasName(field *ast.Field, name string) bool {
	for _, fieldName := range field.Names {
		if fieldName.Name == name {
			return true
		}
	}
	return false
}

func exprString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return buf.String()
}

func cloneExpr(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}

	cloned, err := parser.ParseExpr(exprString(expr))
	if err != nil {
		return expr
	}
	return cloned
}
