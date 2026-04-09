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
const detEventoStruct = "type DetEvento struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
const anonDetEventoStruct = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
const infModalStruct = "type InfModal struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
const anonInfModalStruct = "type TAnonComplexInfModal1 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
const anonInfModalStruct2 = "type TAnonComplexInfModal2 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
const anonInfModalStruct3 = "type TAnonComplexInfModal3 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
const xmlOnlyImportBlock = "import (\n\t\"encoding/xml\"\n)"
const typedInfModalImportBlock = "import (\n\t\"encoding/xml\"\n\n\tmodalaereo \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_aereo\"\n\tmodalaquaviario \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_aquaviario\"\n\tmodaldutoviario \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_dutoviario\"\n\tmodalferroviario \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_ferroviario\"\n\tmodalmultimodal \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_multimodal\"\n\tmodalrodoviario \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_rodoviario\"\n\tmodalrodoviarioos \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_rodoviario_os\"\n)"
const typedInfModalImportBlockOS = "import (\n\t\"encoding/xml\"\n\n\tmodalrodoviarioos \"github.com/awa/nota-fiscal/internal/cte/gen/v4_0/modal_rodoviario_os\"\n)"

var anonComplexXMLName = regexp.MustCompile("`xml:\"TAnonComplex_([^\"_]+(?:_[^\"_]+)*)_\\d+\"`")
var optionalFieldDhCont = regexp.MustCompile(`\n\tDhCont\s+string\s+` + "`xml:\"dhCont\"`")
var optionalFieldXJust = regexp.MustCompile(`\n\tXJust\s+string\s+` + "`xml:\"xJust\"`")
var optionalFieldCRT = regexp.MustCompile(`\n\tCRT\s+string\s+` + "`xml:\"CRT\"`")
var eventPayloads = map[string]string{
	"evento_cce":                    "evCCeCTe",
	"evento_cancel":                 "evCancCTe",
	"evento_ce":                     "evCECTe",
	"evento_cancel_ce":              "evCancCECTe",
	"evento_cancel_ie":              "evCancIECTe",
	"evento_cancel_prest_desacordo": "evCancPrestDesacordo",
	"evento_epec":                   "evEPECCTe",
	"evento_gtv":                    "evGTV",
	"evento_ie":                     "evIECTe",
	"evento_prest_desacordo":        "evPrestDesacordo",
	"evento_reg_multimodal":         "evRegMultimodal",
}

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: go run ./internal/cte/tools/codegen <normalize-schemas|postprocess-generated> [schema-dir ...]")
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

	root := filepath.Join(repoRoot, "internal", "cte", "gen")
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		if isDiscardedModalSupportFile(path) {
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Printf("removed modal support schema package %s\n", path)
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
		updated = strings.ReplaceAll(updated, "*interface{}", "*string")
		updated = strings.ReplaceAll(updated, "interface{}", "string")
		updated = optionalFieldDhCont.ReplaceAllString(updated, "\n\tDhCont         *string `xml:\"dhCont\"`")
		updated = optionalFieldXJust.ReplaceAllString(updated, "\n\tXJust          *string `xml:\"xJust\"`")
		updated = optionalFieldCRT.ReplaceAllString(updated, "\n\tCRT       *string   `xml:\"CRT\"`")
		if strings.HasSuffix(path, string(filepath.Separator)+"cteTiposBasico_v4.00.xsd.go") && usesTypedInfModal(path) {
			if strings.Contains(path, string(filepath.Separator)+"cte_os"+string(filepath.Separator)) {
				updated = strings.Replace(updated, xmlOnlyImportBlock, typedInfModalImportBlockOS, 1)
				updated = strings.Replace(updated, infModalStruct, typedInfModal("InfModal", true), 1)
				updated = strings.Replace(updated, anonInfModalStruct, typedInfModal("TAnonComplexInfModal1", true), 1)
				updated = strings.Replace(updated, anonInfModalStruct2, typedInfModal("TAnonComplexInfModal2", true), 1)
				updated = strings.Replace(updated, anonInfModalStruct3, typedInfModal("TAnonComplexInfModal3", true), 1)
			} else {
				updated = strings.Replace(updated, xmlOnlyImportBlock, typedInfModalImportBlock, 1)
				updated = strings.Replace(updated, infModalStruct, typedInfModal("InfModal", false), 1)
				updated = strings.Replace(updated, anonInfModalStruct, typedInfModal("TAnonComplexInfModal1", false), 1)
				updated = strings.Replace(updated, anonInfModalStruct2, typedInfModal("TAnonComplexInfModal2", false), 1)
				updated = strings.Replace(updated, anonInfModalStruct3, typedInfModal("TAnonComplexInfModal3", true), 1)
			}
		}
		if strings.HasSuffix(path, string(filepath.Separator)+"eventoCTeTiposBasico_v4.00.xsd.go") {
			for folder, element := range eventPayloads {
				if strings.Contains(path, string(filepath.Separator)+folder+string(filepath.Separator)) {
					updated = strings.Replace(updated, detEventoStruct, detEventoReplacement("DetEvento", element), 1)
					updated = strings.Replace(updated, anonDetEventoStruct, detEventoReplacement("TAnonComplexDetEvento1", element), 1)
					break
				}
			}
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

func isDiscardedModalSupportFile(path string) bool {
	if filepath.Base(filepath.Dir(path)) == "v4_0" && strings.HasPrefix(filepath.Base(path), "modal_") {
		return true
	}

	parent := filepath.Base(filepath.Dir(path))
	allowed := map[string]string{
		"modal_aereo":         "cteModalAereo_v4.00.xsd.go",
		"modal_aquaviario":    "cteModalAquaviario_v4.00.xsd.go",
		"modal_dutoviario":    "cteModalDutoviario_v4.00.xsd.go",
		"modal_ferroviario":   "cteModalFerroviario_v4.00.xsd.go",
		"modal_rodoviario":    "cteModalRodoviario_v4.00.xsd.go",
		"modal_rodoviario_os": "cteModalRodoviarioOS_v4.00.xsd.go",
		"modal_multimodal":    "cteMultiModal_v4.00.xsd.go",
	}

	rootFile, ok := allowed[parent]
	if !ok {
		return false
	}

	return filepath.Base(path) != rootFile
}

func isNestedImportedSchema(path string) bool {
	clean := filepath.Clean(path)
	patterns := []string{
		string(filepath.Separator) + filepath.Join("nfelib", "nfelib", "cte", "schemas") + string(filepath.Separator),
		string(filepath.Separator) + filepath.Join("internal", "cte", "schemas") + string(filepath.Separator),
		string(filepath.Separator) + filepath.Join("internal", "cte", "gen") + string(filepath.Separator) + "v4_0" + string(filepath.Separator) + "schemas" + string(filepath.Separator),
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
			filepath.Join("internal", "cte", "schemas", "v4_0", "cte"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "cte_os"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "consulta_situacao"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "status_servico"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "cte_simp"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "gtve"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_cce"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_cancel"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_ce"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_cancel_ce"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_cancel_ie"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_cancel_prest_desacordo"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_epec"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_gtv"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_ie"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_prest_desacordo"),
			filepath.Join("internal", "cte", "schemas", "v4_0", "evento_reg_multimodal"),
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

func detEventoReplacement(typeName, element string) string {
	typeSuffix := strings.ToUpper(element[:1]) + element[1:]
	return fmt.Sprintf(
		"type %s struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\t%s         *TAnonComplex%s1 `xml:\"%s\"`\n}",
		typeName,
		typeSuffix,
		typeSuffix,
		element,
	)
}

func usesTypedInfModal(path string) bool {
	parent := filepath.Base(filepath.Dir(path))
	for _, folder := range []string{"cte", "cte_os", "cte_simp", "gtve"} {
		if parent == folder {
			return true
		}
	}
	return false
}

func typedInfModal(typeName string, osOnly bool) string {
	if osOnly {
		return fmt.Sprintf(
			"type %s struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tRodoOS          modalrodoviarioos.RodoOS `xml:\"rodoOS,omitempty\"`\n}",
			typeName,
		)
	}
	return fmt.Sprintf(
		"type %s struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tRodo            modalrodoviario.Rodo `xml:\"rodo,omitempty\"`\n\tAereo           modalaereo.Aereo `xml:\"aereo,omitempty\"`\n\tAquav           modalaquaviario.Aquav `xml:\"aquav,omitempty\"`\n\tFerrov          modalferroviario.Ferrov `xml:\"ferrov,omitempty\"`\n\tDuto            modaldutoviario.Duto `xml:\"duto,omitempty\"`\n\tMultimodal      modalmultimodal.Multimodal `xml:\"multimodal,omitempty\"`\n}",
		typeName,
	)
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
