package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type normalizeStats struct {
	files                      int
	rewritten                  int
	generatedTypes             int
	flattenedOptionalSequences int
}

func normalizeSchemas(repoRoot string, roots []string) (normalizeStats, error) {
	var stats normalizeStats
	for _, rootArg := range roots {
		root := rootArg
		if !filepath.IsAbs(root) {
			root = filepath.Join(repoRoot, rootArg)
		}

		entries, err := os.ReadDir(root)
		if err != nil {
			return stats, err
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".xsd" {
				continue
			}

			path := filepath.Join(root, entry.Name())
			fileStats, err := normalizeSchema(path)
			if err != nil {
				return stats, fmt.Errorf("%s: %w", path, err)
			}

			stats.files++
			stats.rewritten += fileStats.rewritten
			stats.generatedTypes += fileStats.generatedTypes
			stats.flattenedOptionalSequences += fileStats.flattenedOptionalSequences
		}
	}

	return stats, nil
}

func normalizeSchema(path string) (normalizeStats, error) {
	var stats normalizeStats

	data, err := os.ReadFile(path)
	if err != nil {
		return stats, err
	}

	var schema element
	if err := xml.Unmarshal(data, &schema); err != nil {
		return stats, err
	}

	typeInsertAt := 0
	for idx, child := range schema.Children {
		if isSchemaPrelude(child.XMLName.Local) {
			typeInsertAt = idx + 1
			continue
		}
		break
	}

	stats.flattenedOptionalSequences = flattenOptionalDirectElementSequences(&schema)

	simpleCounters := map[string]int{}
	complexCounters := map[string]int{}
	var generated []element
	collectInlineTypes(&schema, simpleCounters, complexCounters, &generated)
	stats.generatedTypes = len(generated)
	if stats.generatedTypes > 0 {
		schema.Children = insertChildren(schema.Children, typeInsertAt, generated)
	}

	output, err := marshalXML(schema)
	if err != nil {
		return stats, err
	}

	if bytes.Equal(data, output) {
		return stats, nil
	}

	if err := os.WriteFile(path, output, 0o644); err != nil {
		return stats, err
	}
	stats.rewritten = 1
	return stats, nil
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
	if err := encodeElement(encoder, &schema, true); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func encodeElement(enc *xml.Encoder, node *element, root bool) error {
	start := xml.StartElement{Name: xml.Name{Local: node.XMLName.Local}, Attr: encodeAttrs(node.Attrs, root)}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	if strings.TrimSpace(node.Content) != "" {
		if err := enc.EncodeToken(xml.CharData([]byte(node.Content))); err != nil {
			return err
		}
	}

	for _, child := range node.Children {
		if err := encodeElement(enc, child, false); err != nil {
			return err
		}
	}

	return enc.EncodeToken(start.End())
}

func encodeAttrs(attrs []xml.Attr, root bool) []xml.Attr {
	encoded := make([]xml.Attr, 0, len(attrs))
	namespaces := rootNamespaces(attrs)

	if root {
		encoded = append(encoded, namespaces...)
	}

	seen := map[string]bool{}
	for _, attr := range attrs {
		if isNamespaceAttr(attr) {
			continue
		}
		if attr.Name.Local == "" {
			continue
		}

		name := xml.Name{Local: attr.Name.Local}
		if seen[name.Local] {
			continue
		}
		seen[name.Local] = true
		encoded = append(encoded, xml.Attr{Name: name, Value: attr.Value})
	}

	return encoded
}

func rootNamespaces(attrs []xml.Attr) []xml.Attr {
	var namespaces []xml.Attr
	var defaultNamespace *xml.Attr
	seen := map[string]bool{}

	for _, attr := range attrs {
		name, ok := namespaceAttrName(attr)
		if !ok {
			continue
		}

		if name.Local == "xmlns" {
			candidate := xml.Attr{Name: name, Value: attr.Value}
			if defaultNamespace == nil || attr.Value == "http://www.w3.org/2001/XMLSchema" {
				defaultNamespace = &candidate
			}
			continue
		}

		if strings.HasPrefix(name.Local, "xmlns:_") || seen[name.Local] {
			continue
		}
		seen[name.Local] = true
		namespaces = append(namespaces, xml.Attr{Name: name, Value: attr.Value})
	}

	if defaultNamespace != nil {
		namespaces = append([]xml.Attr{*defaultNamespace}, namespaces...)
	}
	return namespaces
}

func isNamespaceAttr(attr xml.Attr) bool {
	_, ok := namespaceAttrName(attr)
	return ok
}

func namespaceAttrName(attr xml.Attr) (xml.Name, bool) {
	switch {
	case attr.Name.Space == "" && attr.Name.Local == "xmlns":
		return xml.Name{Local: "xmlns"}, true
	case attr.Name.Space == "xmlns" && attr.Name.Local != "":
		return xml.Name{Local: "xmlns:" + attr.Name.Local}, true
	case strings.HasPrefix(attr.Name.Local, "xmlns:"):
		return xml.Name{Local: attr.Name.Local}, true
	default:
		return xml.Name{}, false
	}
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

func insertChildren(children []*element, idx int, insert []element) []*element {
	if len(insert) == 0 {
		return children
	}

	converted := make([]*element, 0, len(insert))
	for i := range insert {
		child := insert[i]
		converted = append(converted, &child)
	}

	updated := make([]*element, 0, len(children)+len(converted))
	updated = append(updated, children[:idx]...)
	updated = append(updated, converted...)
	updated = append(updated, children[idx:]...)
	return updated
}

func removeChild(children []*element, idx int) []*element {
	updated := make([]*element, 0, len(children)-1)
	updated = append(updated, children[:idx]...)
	updated = append(updated, children[idx+1:]...)
	return updated
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
			if errors.Is(err, io.EOF) {
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
	return encodeElement(enc, e, true)
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
	copy := e.deepCopyValue()
	return &copy
}

func (e *element) deepCopyValue() element {
	if e == nil {
		return element{}
	}

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
