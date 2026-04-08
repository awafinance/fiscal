package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const xmlDSigSignatureTag = "`xml:\"ds:Signature\"`"
const infModalStruct = "type InfModal struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
const infModalStructWithInnerXML = "type InfModal struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tInnerXML        string   `xml:\",innerxml\"`\n}"
const anonInfModalStruct = "type TAnonComplexInfModal1 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
const anonInfModalStructWithInnerXML = "type TAnonComplexInfModal1 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tInnerXML        string   `xml:\",innerxml\"`\n}"
const detEventoStruct = "type DetEvento struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
const detEventoStructWithInnerXML = "type DetEvento struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\tInnerXML         string   `xml:\",innerxml\"`\n}"
const anonDetEventoStruct = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
const anonDetEventoStructWithInnerXML = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\tInnerXML         string   `xml:\",innerxml\"`\n}"

var anonComplexXMLName = regexp.MustCompile("`xml:\"TAnonComplex_([^\"_]+(?:_[^\"_]+)*)_\\d+\"`")

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: go run ./internal/mdfe/tools/codegen <normalize-schemas|postprocess-generated> [schema-dir ...]")
	}

	switch os.Args[1] {
	case "normalize-schemas":
		if err := normalizeSchemas(os.Args[2:]); err != nil {
			fatalf("normalize schemas: %v", err)
		}
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

	root := filepath.Join(repoRoot, "internal", "mdfe", "gen")
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

		updated := string(text)
		updated = strings.ReplaceAll(updated, xmlDSigSignatureTag, "`xml:\"http://www.w3.org/2000/09/xmldsig# Signature\"`")
		updated = anonComplexXMLName.ReplaceAllString(updated, "`xml:\"$1\"`")
		updated = strings.ReplaceAll(updated, "*TpAmb", "*string")
		updated = strings.Replace(updated, infModalStruct, infModalStructWithInnerXML, 1)
		updated = strings.Replace(updated, anonInfModalStruct, anonInfModalStructWithInnerXML, 1)
		if strings.HasSuffix(path, string(filepath.Separator)+"eventoMDFeTiposBasico_v3.00.xsd.go") {
			updated = strings.Replace(updated, detEventoStruct, detEventoStructWithInnerXML, 1)
			updated = strings.Replace(updated, anonDetEventoStruct, anonDetEventoStructWithInnerXML, 1)
		}
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
		string(filepath.Separator) + filepath.Join("nfelib", "nfelib", "mdfe", "schemas") + string(filepath.Separator),
		string(filepath.Separator) + filepath.Join("internal", "mdfe", "schemas") + string(filepath.Separator),
		string(filepath.Separator) + filepath.Join("internal", "mdfe", "gen") + string(filepath.Separator) + "v3_0" + string(filepath.Separator) + "schemas" + string(filepath.Separator),
	}
	for _, pattern := range patterns {
		if strings.Contains(clean, pattern) {
			return true
		}
	}
	return false
}

func normalizeSchemas(args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	roots := args
	if len(roots) == 0 {
		roots = []string{
			filepath.Join("internal", "mdfe", "schemas", "v3_0", "mdfe"),
			filepath.Join("internal", "mdfe", "schemas", "v3_0", "cons_nao_enc"),
			filepath.Join("internal", "mdfe", "schemas", "v3_0", "cons_reci"),
			filepath.Join("internal", "mdfe", "schemas", "v3_0", "evento"),
		}
	}

	for _, rootArg := range roots {
		root := rootArg
		if !filepath.IsAbs(root) {
			root = filepath.Join(repoRoot, rootArg)
		}

		entries, err := os.ReadDir(root)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".xsd" {
				continue
			}

			path := filepath.Join(root, entry.Name())
			changed, flattened, err := normalizeSchema(path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}

			fmt.Printf("normalized %d inline simpleType elements in %s\n", changed, path)
			fmt.Printf("flattened %d optional direct-element sequences in %s\n", flattened, path)
		}
	}

	return nil
}

func normalizeSchema(path string) (generatedTypes int, flattenedOptionalSequences int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}

	var schema element
	if err := xml.Unmarshal(data, &schema); err != nil {
		return 0, 0, err
	}

	typeInsertAt := 0
	for idx, child := range schema.Children {
		if isSchemaPrelude(child.XMLName.Local) {
			typeInsertAt = idx + 1
			continue
		}
		break
	}

	flattenedOptionalSequences = flattenOptionalDirectElementSequences(&schema)

	simpleCounters := map[string]int{}
	complexCounters := map[string]int{}
	var generated []element
	collectInlineTypes(&schema, simpleCounters, complexCounters, &generated)
	generatedTypes = len(generated)
	if generatedTypes > 0 {
		schema.Children = insertChildren(schema.Children, typeInsertAt, generated)
	}

	output, err := marshalXML(schema)
	if err != nil {
		return 0, 0, err
	}

	if err := os.WriteFile(path, output, 0o644); err != nil {
		return 0, 0, err
	}

	return generatedTypes, flattenedOptionalSequences, nil
}

func flattenOptionalDirectElementSequences(root *element) int {
	count := 0
	var walk func(*element)
	walk = func(node *element) {
		for _, child := range node.Children {
			walk(child)
		}

		for idx := 0; idx < len(node.Children); idx++ {
			child := node.Children[idx]
			if child.XMLName.Local != "sequence" || child.attr("minOccurs") != "0" {
				continue
			}
			if !sequenceHasOnlyDirectElements(child) {
				continue
			}

			replacement := make([]*element, 0, len(child.Children))
			for _, seqChild := range child.Children {
				clone := seqChild.deepCopy()
				clone.setAttr("minOccurs", "0")
				replacement = append(replacement, clone)
			}

			node.Children = replaceChild(node.Children, idx, replacement)
			count++
			idx += len(replacement) - 1
		}
	}

	walk(root)
	return count
}

func collectInlineTypes(node *element, simpleCounters, complexCounters map[string]int, generated *[]element) {
	for _, child := range node.Children {
		collectInlineTypes(child, simpleCounters, complexCounters, generated)
	}

	if node.XMLName.Local != "element" {
		return
	}

	name := node.attr("name")
	if name == "" {
		return
	}

	for idx := 0; idx < len(node.Children); idx++ {
		child := node.Children[idx]
		switch child.XMLName.Local {
		case "complexType":
			complexCounters[name]++
			typeName := fmt.Sprintf("TAnonComplex_%s_%d", name, complexCounters[name])
			clone := child.deepCopyValue()
			clone.setAttr("name", typeName)
			*generated = append(*generated, clone)
			node.Children = removeChild(node.Children, idx)
			node.setAttr("type", typeName)
			idx--
		case "simpleType":
			simpleCounters[name]++
			typeName := fmt.Sprintf("TAnon_%s_%d", name, simpleCounters[name])
			clone := child.deepCopyValue()
			clone.setAttr("name", typeName)
			*generated = append(*generated, clone)
			node.Children = removeChild(node.Children, idx)
			node.setAttr("type", typeName)
			idx--
		}
	}
}

func marshalXML(schema element) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encodeElement(encoder, &schema); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func encodeElement(enc *xml.Encoder, node *element) error {
	start := xml.StartElement{Name: node.XMLName, Attr: append([]xml.Attr(nil), node.Attrs...)}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	if strings.TrimSpace(node.Content) != "" {
		if err := enc.EncodeToken(xml.CharData([]byte(node.Content))); err != nil {
			return err
		}
	}

	for _, child := range node.Children {
		if err := encodeElement(enc, child); err != nil {
			return err
		}
	}

	return enc.EncodeToken(start.End())
}

func isSchemaPrelude(local string) bool {
	switch local {
	case "annotation", "import", "include", "redefine", "simpleType":
		return true
	default:
		return false
	}
}

func sequenceHasOnlyDirectElements(node *element) bool {
	if len(node.Children) == 0 {
		return false
	}

	for _, child := range node.Children {
		if child.XMLName.Local != "element" {
			return false
		}
	}

	return true
}

func replaceChild(children []*element, idx int, replacement []*element) []*element {
	updated := make([]*element, 0, len(children)-1+len(replacement))
	updated = append(updated, children[:idx]...)
	updated = append(updated, replacement...)
	updated = append(updated, children[idx+1:]...)
	return updated
}

func removeChild(children []*element, idx int) []*element {
	updated := make([]*element, 0, len(children)-1)
	updated = append(updated, children[:idx]...)
	updated = append(updated, children[idx+1:]...)
	return updated
}

func insertChildren(children []*element, idx int, inserted []element) []*element {
	newChildren := make([]*element, 0, len(children)+len(inserted))
	newChildren = append(newChildren, children[:idx]...)
	for i := range inserted {
		child := inserted[i]
		newChildren = append(newChildren, child.deepCopy())
	}
	newChildren = append(newChildren, children[idx:]...)
	return newChildren
}

type element struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Children []*element `xml:",any"`
	Content  string     `xml:",chardata"`
}

func (e *element) attr(name string) string {
	for _, attr := range e.Attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func (e *element) setAttr(name, value string) {
	for idx, attr := range e.Attrs {
		if attr.Name.Local == name {
			e.Attrs[idx].Value = value
			return
		}
	}
	e.Attrs = append(e.Attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

func (e *element) deepCopy() *element {
	if e == nil {
		return nil
	}
	copyValue := e.deepCopyValue()
	return &copyValue
}

func (e *element) deepCopyValue() element {
	if e == nil {
		return element{}
	}

	attrs := make([]xml.Attr, len(e.Attrs))
	copy(attrs, e.Attrs)

	children := make([]*element, len(e.Children))
	for idx, child := range e.Children {
		children[idx] = child.deepCopy()
	}

	return element{
		XMLName:  e.XMLName,
		Attrs:    attrs,
		Children: children,
		Content:  e.Content,
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if exists(filepath.Join(dir, "go.mod")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate repository root from %s", dir)
		}
		dir = parent
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
