package postprocess

import (
	"strings"
	"testing"
)

func TestAddJSONTagsAddsTagsFromXML(t *testing.T) {
	src := []byte(`package schema

import "encoding/xml"

type Doc struct {
	XMLName xml.Name ` + "`xml:\"doc\"`" + `
	Value   string   ` + "`xml:\"value\"`" + `
	Skip    string   ` + "`xml:\"-\"`" + `
	Attr    string   ` + "`xml:\"id,attr\"`" + `
	Tagged  string   ` + "`xml:\"tagged\" json:\"tagged\"`" + `
}
`)

	updated, changed, err := addJSONTags("doc.go", src)
	if err != nil {
		t.Fatalf("addJSONTags returned error: %v", err)
	}
	if !changed {
		t.Fatal("expected addJSONTags to report changes")
	}

	got := string(updated)
	wantSnippets := []string{
		"`xml:\"doc\" json:\"-\"`",
		"`xml:\"value\" json:\"value,omitempty\"`",
		"`xml:\"-\" json:\"-\"`",
		"`xml:\"id,attr\" json:\"id,omitempty\"`",
		"`xml:\"tagged\" json:\"tagged\"`",
	}
	for _, snippet := range wantSnippets {
		if !strings.Contains(got, snippet) {
			t.Fatalf("updated source missing snippet %q:\n%s", snippet, got)
		}
	}
}

func TestAddJSONTagsNoOpWhenTagsAlreadyPresent(t *testing.T) {
	src := []byte("package schema\n\ntype Doc struct {\n\tValue string `xml:\"value\" json:\"value\"`\n}\n")

	updated, changed, err := addJSONTags("doc.go", src)
	if err != nil {
		t.Fatalf("addJSONTags returned error: %v", err)
	}
	if changed {
		t.Fatal("expected addJSONTags to report no changes")
	}
	if string(updated) != string(src) {
		t.Fatalf("expected source to remain unchanged:\n%s", updated)
	}
}

func TestReplaceFieldType(t *testing.T) {
	src := `package schema

type Doc struct {
	Value   interface{}  ` + "`xml:\"value\"`" + `
	Pointer *interface{} ` + "`xml:\"pointer\"`" + `
	Keep    int          ` + "`xml:\"keep\"`" + `
}
`

	replacer := IfPath(
		func(path string) bool { return strings.HasSuffix(path, "doc.go") },
		ReplaceFieldType(FieldTypeEquals("interface{}"), "string"),
		ReplaceFieldType(FieldTypeEquals("*interface{}"), "*string"),
	)

	updated := replacer("doc.go", src)
	wantSnippets := []string{
		"Value\tstring\t`xml:\"value\"`",
		"Pointer\t*string\t`xml:\"pointer\"`",
		"Keep\tint\t`xml:\"keep\"`",
	}
	for _, snippet := range wantSnippets {
		if !strings.Contains(updated, snippet) {
			t.Fatalf("updated source missing snippet %q:\n%s", snippet, updated)
		}
	}
}
