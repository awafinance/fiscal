package main

import (
	"encoding/xml"
	"testing"
)

func TestFlattenOptionalDirectElementSequences(t *testing.T) {
	root := &element{
		XMLName: xml.Name{Local: "complexType"},
		Children: []*element{
			{
				XMLName: xml.Name{Local: "sequence"},
				Attrs:   []xml.Attr{{Name: xml.Name{Local: "minOccurs"}, Value: "0"}},
				Children: []*element{
					{XMLName: xml.Name{Local: "element"}, Attrs: []xml.Attr{{Name: xml.Name{Local: "name"}, Value: "first"}}},
					{XMLName: xml.Name{Local: "element"}, Attrs: []xml.Attr{{Name: xml.Name{Local: "name"}, Value: "second"}}},
				},
			},
		},
	}

	count := flattenOptionalDirectElementSequences(root)
	if count != 1 {
		t.Fatalf("expected one flattened sequence, got %d", count)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected sequence to be replaced with two elements, got %d children", len(root.Children))
	}
	for _, child := range root.Children {
		if child.XMLName.Local != "element" {
			t.Fatalf("expected replacement child to be an element, got %q", child.XMLName.Local)
		}
		if got := child.attr("minOccurs"); got != "0" {
			t.Fatalf("expected replacement child minOccurs=0, got %q", got)
		}
	}
}

func TestCollectInlineTypesExtractsNamedTypes(t *testing.T) {
	root := &element{
		XMLName: xml.Name{Local: "schema"},
		Children: []*element{
			{
				XMLName: xml.Name{Local: "element"},
				Attrs:   []xml.Attr{{Name: xml.Name{Local: "name"}, Value: "Choice"}},
				Children: []*element{
					{XMLName: xml.Name{Local: "simpleType"}},
					{XMLName: xml.Name{Local: "complexType"}},
				},
			},
		},
	}

	var generated []element
	collectInlineTypes(root, map[string]int{}, map[string]int{}, &generated)

	if len(generated) != 2 {
		t.Fatalf("expected two generated types, got %d", len(generated))
	}
	if got := generated[0].attr("name"); got != "TAnon_Choice_1" {
		t.Fatalf("unexpected generated simple type name: %q", got)
	}
	if got := generated[1].attr("name"); got != "TAnonComplex_Choice_1" {
		t.Fatalf("unexpected generated complex type name: %q", got)
	}
	if got := root.Children[0].attr("type"); got != "TAnonComplex_Choice_1" {
		t.Fatalf("expected final element type to point at extracted complex type, got %q", got)
	}
	if len(root.Children[0].Children) != 0 {
		t.Fatalf("expected inline types to be removed from element, got %d children", len(root.Children[0].Children))
	}
}
