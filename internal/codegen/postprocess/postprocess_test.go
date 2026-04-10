package postprocess

import (
	"os"
	"path/filepath"
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

func TestGeneratedFixturePipeline(t *testing.T) {
	genDir := t.TempDir()

	keepPath := filepath.Join(genDir, "keep.go")
	dupPath := filepath.Join(genDir, "nested", "schemas", "dup.go")
	removePath := filepath.Join(genDir, "remove.go")

	if err := os.MkdirAll(filepath.Dir(dupPath), 0o755); err != nil {
		t.Fatalf("mkdir dup dir: %v", err)
	}

	keepSrc := `package schema

type Doc struct {
	Value interface{} ` + "`xml:\"value\"`" + `
}
`
	if err := os.WriteFile(keepPath, []byte(keepSrc), 0o644); err != nil {
		t.Fatalf("write keep file: %v", err)
	}

	if err := os.WriteFile(dupPath, []byte("package schema\n"), 0o644); err != nil {
		t.Fatalf("write dup file: %v", err)
	}
	if err := os.WriteFile(removePath, []byte("package schema\n"), 0o644); err != nil {
		t.Fatalf("write remove file: %v", err)
	}

	err := Generated(Options{
		GenDir: genDir,
		NestedImportPatterns: []string{
			string(filepath.Separator) + filepath.Join("nested", "schemas") + string(filepath.Separator),
		},
		RemoveFile: func(path string) (bool, string) {
			return filepath.Base(path) == "remove.go", "removed fixture file"
		},
		Replacements: []Replacement{
			ReplaceFieldType(FieldTypeEquals("interface{}"), "string"),
		},
		AddJSONTags: true,
	})
	if err != nil {
		t.Fatalf("Generated returned error: %v", err)
	}

	updated, err := os.ReadFile(keepPath)
	if err != nil {
		t.Fatalf("read keep file: %v", err)
	}
	got := string(updated)
	wantSnippets := []string{
		"Value string `xml:\"value\" json:\"value,omitempty\"`",
	}
	for _, snippet := range wantSnippets {
		if !strings.Contains(got, snippet) {
			t.Fatalf("updated keep.go missing snippet %q:\n%s", snippet, got)
		}
	}

	if _, err := os.Stat(dupPath); !os.IsNotExist(err) {
		t.Fatalf("expected nested imported schema file to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(removePath); !os.IsNotExist(err) {
		t.Fatalf("expected remove.go to be removed, stat err=%v", err)
	}
}

func TestAddStructField(t *testing.T) {
	src := `package schema

type DetEvento struct {
	XMLName string ` + "`xml:\"detEvento\"`" + `
}
`

	replacer := AddStructField(TypeNamed("DetEvento"), "InnerXML", "string", `xml:",innerxml"`)
	updated := replacer("doc.go", src)

	if !strings.Contains(updated, "InnerXML\tstring\t`xml:\",innerxml\"`") {
		t.Fatalf("updated source missing InnerXML field:\n%s", updated)
	}

	updatedAgain := replacer("doc.go", updated)
	if strings.Count(updatedAgain, "InnerXML") != 1 {
		t.Fatalf("expected InnerXML field to be inserted once, got:\n%s", updatedAgain)
	}
}
