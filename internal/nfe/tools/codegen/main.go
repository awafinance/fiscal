package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const xmlDSigSignatureTag = "`xml:\"ds:Signature\"`"

var anonComplexXMLName = regexp.MustCompile("`xml:\"TAnonComplex_([^\"_]+(?:_[^\"_]+)*)_\\d+\"`")

func main() {
	if len(os.Args) != 2 {
		fatalf("usage: go run ./internal/nfe/tools/codegen <normalize-schemas|postprocess-generated>")
	}

	switch os.Args[1] {
	case "normalize-schemas":
		if err := normalizeSchemas(); err != nil {
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

	root := filepath.Join(repoRoot, "internal", "nfe", "gen", "v4_0")
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		updated := strings.ReplaceAll(string(text), xmlDSigSignatureTag, "`xml:\"http://www.w3.org/2000/09/xmldsig# Signature\"`")
		updated = anonComplexXMLName.ReplaceAllString(updated, "`xml:\"$1\"`")
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

func normalizeSchemas() error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	root := filepath.Join(repoRoot, "internal", "nfe", "schemas", "v4_0", "nfe_proc")
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
		clone := inserted[i]
		newChildren = append(newChildren, &clone)
	}
	newChildren = append(newChildren, children[idx:]...)
	return newChildren
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
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
			return "", fmt.Errorf("could not locate repo root from %s", dir)
		}
		dir = parent
	}
}

type element struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Children []*element `xml:",any"`
	Content  string     `xml:",chardata"`
}

func (e *element) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	e.XMLName = start.Name
	e.Attrs = append([]xml.Attr(nil), start.Attr...)

	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			child := &element{}
			if err := child.UnmarshalXML(d, tok); err != nil {
				return err
			}
			e.Children = append(e.Children, child)
		case xml.EndElement:
			if tok.Name == start.Name {
				return nil
			}
		case xml.CharData:
			text := strings.TrimSpace(string(tok))
			if text != "" {
				if e.Content == "" {
					e.Content = text
				} else {
					e.Content += text
				}
			}
		}
	}
}

func (e *element) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return encodeElement(enc, e)
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
	for i, attr := range e.Attrs {
		if attr.Name.Local == name {
			e.Attrs[i].Value = value
			return
		}
	}

	e.Attrs = append(e.Attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

func (e *element) deepCopy() *element {
	copyValue := e.deepCopyValue()
	return &copyValue
}

func (e *element) deepCopyValue() element {
	clone := element{
		XMLName: e.XMLName,
		Attrs:   append([]xml.Attr(nil), e.Attrs...),
		Content: e.Content,
	}

	if len(e.Children) > 0 {
		clone.Children = make([]*element, 0, len(e.Children))
		for _, child := range e.Children {
			clone.Children = append(clone.Children, child.deepCopy())
		}
	}

	return clone
}
