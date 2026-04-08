package cte_test

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/awa/nota-fiscal/pkg/cte"
	"github.com/stretchr/testify/require"
)

const (
	cteNamespace = "http://www.portalfiscal.inf.br/cte"
	dsNamespace  = "http://www.w3.org/2000/09/xmldsig#"
)

func TestParse_Fixtures(t *testing.T) {
	t.Parallel()

	for _, fixture := range allFixtureNames(t) {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			original := readFixture(t, fixture)
			doc := parseFixture(t, fixture)

			assertFixtureShape(t, fixture, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, original), normalizeXML(t, roundTripped))

			reparsed, err := cte.Parse(roundTripped)
			require.NoError(t, err)
			assertSameRoot(t, doc, reparsed)
			assertFixtureShape(t, fixture, reparsed)
		})
	}
}

func TestParse_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		data        []byte
		errContains string
	}{
		{name: "empty", data: nil, errContains: "empty xml document"},
		{name: "unsupported root", data: []byte(`<foo></foo>`), errContains: `unsupported root element "foo"`},
		{name: "invalid cte", data: []byte(`<CTe xmlns="http://www.portalfiscal.inf.br/cte"></CTe>`), errContains: "missing infCte"},
		{name: "invalid cteos", data: []byte(`<CTeOS xmlns="http://www.portalfiscal.inf.br/cte" versao="4.00"></CTeOS>`), errContains: "missing infCte"},
		{name: "invalid event", data: []byte(`<eventoCTe xmlns="http://www.portalfiscal.inf.br/cte" versao="4.00"></eventoCTe>`), errContains: "missing infEvento"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := cte.Parse(tt.data)
			require.Error(t, err)
			require.Nil(t, doc)
			require.ErrorContains(t, err, tt.errContains)
		})
	}
}

func TestMarshalXML_NilReceiver(t *testing.T) {
	t.Parallel()

	var doc *cte.Document
	data, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Empty(t, data)
}

func assertFixtureShape(t *testing.T, fixture string, doc *cte.Document) {
	t.Helper()

	switch fixture {
	case "43120178408960000182570010000000041000000047-cte.xml":
		require.NotNil(t, doc.CTe)
		require.Equal(t, "3.00", doc.CTe.InfCte.VersaoAttr)
		require.Equal(t, "4", doc.CTe.InfCte.Ide.NCT)
		require.Equal(t, "78408960000182", requirePtr(t, doc.CTe.InfCte.Emit.CNPJ))
		require.Equal(t, "MASTER", requirePtr(t, doc.CTe.InfCte.Compl.XEmi))
	case "35190602427026001207570040000522031000522035-cte-multimodal.xml":
		require.NotNil(t, doc.CTe)
		require.Equal(t, "22222", doc.CTe.InfCte.Ide.NCT)
		require.Equal(t, "06", doc.CTe.InfCte.Ide.Modal)
		require.Equal(t, "Transporte Multimodal", doc.CTe.InfCte.Ide.NatOp)
	case "35170799999999999999670000000000261309301440-cte-of.xml":
		require.NotNil(t, doc.CTeOS)
		require.Equal(t, "3.00", doc.CTeOS.VersaoAttr)
		require.Equal(t, "26", doc.CTeOS.InfCte.Ide.NCT)
		require.Equal(t, "67", doc.CTeOS.InfCte.Ide.Mod)
	case "cce35150107565416000104570000000012301000012300-ped-eve.xml":
		require.NotNil(t, doc.EventoCTe)
		require.Equal(t, "3.00", doc.EventoCTe.VersaoAttr)
		require.Equal(t, "35150107565416000104570000000012301000012300", doc.EventoCTe.InfEvento.ChCTe)
		require.Equal(t, "110110", doc.EventoCTe.InfEvento.TpEvento)
		require.Equal(t, "1", doc.EventoCTe.InfEvento.NSeqEvento)
	default:
		t.Fatalf("unhandled fixture %s", fixture)
	}
}

func assertSameRoot(t *testing.T, expected, actual *cte.Document) {
	t.Helper()

	require.Equal(t, expected.CTe != nil, actual.CTe != nil)
	require.Equal(t, expected.CTeOS != nil, actual.CTeOS != nil)
	require.Equal(t, expected.EventoCTe != nil, actual.EventoCTe != nil)
}

func allFixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "..", "testdata", "cte", "v4_0"))
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xml" {
			continue
		}
		names = append(names, entry.Name())
	}

	slices.Sort(names)
	return names
}

func parseFixture(t *testing.T, name string) *cte.Document {
	t.Helper()

	data := readFixture(t, name)
	doc, err := cte.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "cte", "v4_0", name))
	require.NoError(t, err)
	return data
}

func normalizeXML(t *testing.T, data []byte) string {
	t.Helper()

	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	decoder := xml.NewDecoder(bytes.NewReader(data))
	var b strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			b.WriteByte('<')
			b.WriteString(qualifiedName(tok.Name))
			attrs := make([]xml.Attr, 0, len(tok.Attr))
			for _, attr := range tok.Attr {
				if isNamespaceDecl(attr) {
					continue
				}
				attrs = append(attrs, attr)
			}
			slices.SortFunc(attrs, func(a, b xml.Attr) int {
				return cmp.Or(
					strings.Compare(a.Name.Space, b.Name.Space),
					strings.Compare(a.Name.Local, b.Name.Local),
				)
			})
			for _, attr := range attrs {
				b.WriteByte(' ')
				b.WriteString(qualifiedName(attr.Name))
				b.WriteString(`="`)
				b.WriteString(strings.TrimSpace(attr.Value))
				b.WriteByte('"')
			}
			b.WriteByte('>')
		case xml.EndElement:
			b.WriteString("</")
			b.WriteString(qualifiedName(tok.Name))
			b.WriteByte('>')
		case xml.CharData:
			text := strings.TrimSpace(string(tok))
			if text != "" {
				b.WriteString(text)
			}
		}
	}

	return b.String()
}

func qualifiedName(name xml.Name) string {
	switch name.Space {
	case "", cteNamespace:
		return name.Local
	case dsNamespace:
		return "ds:" + name.Local
	case "xmlns":
		if name.Local == "" {
			return "xmlns"
		}
		return "xmlns:" + name.Local
	}

	return name.Space + ":" + name.Local
}

func isNamespaceDecl(attr xml.Attr) bool {
	return attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns"
}

func requirePtr[T any](t *testing.T, v *T) T {
	t.Helper()
	require.NotNil(t, v)
	return *v
}
