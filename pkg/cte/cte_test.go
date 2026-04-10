package cte_test

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	distSchema "github.com/awafinance/fiscal/internal/cte/gen/v1_0/dist_dfe"
	"github.com/awafinance/fiscal/pkg/cte"
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
		{name: "unsupported event type", data: []byte(`<eventoCTe xmlns="http://www.portalfiscal.inf.br/cte" versao="4.00"><infEvento><tpEvento>999999</tpEvento></infEvento></eventoCTe>`), errContains: `unsupported tpEvento "999999"`},
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

func TestParse_DistDFeRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  any
		assert func(t *testing.T, doc *cte.Document)
	}{
		{
			name: "distDFeInt",
			value: struct {
				XMLName xml.Name `xml:"distDFeInt"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distSchema.TAnonComplexDistDFeInt1
			}{
				XMLName: xml.Name{Local: "distDFeInt"},
				XMLNS:   cteNamespace,
				TAnonComplexDistDFeInt1: &distSchema.TAnonComplexDistDFeInt1{
					VersaoAttr: "1.00",
					TpAmb:      "1",
					CUFAutor:   "35",
					CNPJ:       stringPtr("12345678000195"),
					DistNSU:    &distSchema.TAnonComplexDistNSU1{UltNSU: "000000000000001"},
				},
			},
			assert: func(t *testing.T, doc *cte.Document) {
				t.Helper()
				require.NotNil(t, doc.DistDFeInt)
				require.Equal(t, "35", doc.DistDFeInt.CUFAutor)
			},
		},
		{
			name: "retDistDFeInt",
			value: struct {
				XMLName xml.Name `xml:"retDistDFeInt"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distSchema.TAnonComplexRetDistDFeInt1
			}{
				XMLName: xml.Name{Local: "retDistDFeInt"},
				XMLNS:   cteNamespace,
				TAnonComplexRetDistDFeInt1: &distSchema.TAnonComplexRetDistDFeInt1{
					VersaoAttr: "1.00",
					TpAmb:      "1",
					VerAplic:   cteTStringPtr("test"),
					CStat:      "138",
					XMotivo:    cteTStringPtr("Documento localizado"),
					DhResp:     "2024-01-02T03:04:05",
					UltNSU:     "000000000000010",
					MaxNSU:     "000000000000099",
					LoteDistDFeInt: &distSchema.TAnonComplexLoteDistDFeInt1{
						DocZip: []*distSchema.TAnonComplexDocZip1{{NSUAttr: "000000000000010", SchemaAttr: "procCTe_v4.00.xsd", Value: "ZGF0YQ=="}},
					},
				},
			},
			assert: func(t *testing.T, doc *cte.Document) {
				t.Helper()
				require.NotNil(t, doc.RetDistDFeInt)
				require.Len(t, doc.RetDistDFeInt.LoteDistDFeInt.DocZip, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := xml.MarshalIndent(tt.value, "", "  ")
			require.NoError(t, err)

			doc, err := cte.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
		})
	}
}

func TestParse_EventReturnAndProcRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		assert func(t *testing.T, doc *cte.Document)
	}{
		{name: "ret cce", data: minimalCTeRetEventXML("110110"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoCTe) }},
		{name: "ret cancel", data: minimalCTeRetEventXML("110111"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoCancCTe) }},
		{name: "ret epec", data: minimalCTeRetEventXML("110113"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoEPECCTe) }},
		{name: "ret reg multimodal", data: minimalCTeRetEventXML("110160"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoRegMultimodal) }},
		{name: "ret gtv", data: minimalCTeRetEventXML("110170"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoGTV) }},
		{name: "ret ce", data: minimalCTeRetEventXML("110180"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoCECTe) }},
		{name: "ret cancel ce", data: minimalCTeRetEventXML("110181"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoCancCECTe) }},
		{name: "ret ie", data: minimalCTeRetEventXML("110190"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoIECTe) }},
		{name: "ret cancel ie", data: minimalCTeRetEventXML("110191"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoCancIECTe) }},
		{name: "ret prest desacordo", data: minimalCTeRetEventXML("610110"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoPrestDesacordo) }},
		{name: "ret cancel prest desacordo", data: minimalCTeRetEventXML("610111"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.RetEventoCancPrestDesacordo) }},
		{name: "proc cce", data: minimalCTeProcEventXML("110110"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoCTe) }},
		{name: "proc cancel", data: minimalCTeProcEventXML("110111"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoCancCTe) }},
		{name: "proc epec", data: minimalCTeProcEventXML("110113"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoEPECCTe) }},
		{name: "proc reg multimodal", data: minimalCTeProcEventXML("110160"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoRegMultimodal) }},
		{name: "proc gtv", data: minimalCTeProcEventXML("110170"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoGTV) }},
		{name: "proc ce", data: minimalCTeProcEventXML("110180"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoCECTe) }},
		{name: "proc cancel ce", data: minimalCTeProcEventXML("110181"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoCancCECTe) }},
		{name: "proc ie", data: minimalCTeProcEventXML("110190"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoIECTe) }},
		{name: "proc cancel ie", data: minimalCTeProcEventXML("110191"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoCancIECTe) }},
		{name: "proc prest desacordo", data: minimalCTeProcEventXML("610110"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoPrestDesacordo) }},
		{name: "proc cancel prest desacordo", data: minimalCTeProcEventXML("610111"), assert: func(t *testing.T, doc *cte.Document) { t.Helper(); require.NotNil(t, doc.ProcEventoCancPrestDesacordo) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := cte.Parse([]byte(tt.data))
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)

			reparsed, err := cte.Parse(roundTripped)
			require.NoError(t, err)
			tt.assert(t, reparsed)
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
	case "cancel35150107565416000104570000000012301000012300-ped-eve.xml":
		require.NotNil(t, doc.EventoCancCTe)
		require.Equal(t, "4.00", doc.EventoCancCTe.VersaoAttr)
		require.Equal(t, "35150107565416000104570000000012301000012300", doc.EventoCancCTe.InfEvento.ChCTe)
		require.Equal(t, "110111", doc.EventoCancCTe.InfEvento.TpEvento)
		require.Equal(t, "135150000000001", doc.EventoCancCTe.InfEvento.DetEvento.EvCancCTe.NProt)
	default:
		t.Fatalf("unhandled fixture %s", fixture)
	}
}

func assertSameRoot(t *testing.T, expected, actual *cte.Document) {
	t.Helper()

	require.Equal(t, expected.CTe != nil, actual.CTe != nil)
	require.Equal(t, expected.CTeOS != nil, actual.CTeOS != nil)
	require.Equal(t, expected.EventoCTe != nil, actual.EventoCTe != nil)
	require.Equal(t, expected.EventoCancCTe != nil, actual.EventoCancCTe != nil)
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

func minimalCTeRetEventXML(tpEvento string) string {
	return fmt.Sprintf(`<retEventoCTe xmlns="%s" versao="4.00"><infEvento><tpAmb>2</tpAmb><cStat>135</cStat><tpEvento>%s</tpEvento></infEvento></retEventoCTe>`, cteNamespace, tpEvento)
}

func minimalCTeProcEventXML(tpEvento string) string {
	return fmt.Sprintf(`<procEventoCTe xmlns="%s" versao="4.00"><eventoCTe versao="4.00"><infEvento Id="ID%s4319081234567800019557001000000001100000001101"><cOrgao>43</cOrgao><tpAmb>2</tpAmb><CNPJ>12345678000195</CNPJ><chCTe>43190812345678000195570010000000011000000011</chCTe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>%s</tpEvento><nSeqEvento>1</nSeqEvento><detEvento></detEvento></infEvento></eventoCTe><retEventoCTe versao="4.00"><infEvento><tpAmb>2</tpAmb><cStat>135</cStat><tpEvento>%s</tpEvento></infEvento></retEventoCTe></procEventoCTe>`, cteNamespace, tpEvento, tpEvento, tpEvento)
}

func normalizeXML(t *testing.T, data []byte) string {
	t.Helper()

	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	decoder := xml.NewDecoder(bytes.NewReader(data))
	var b strings.Builder
	nsStack := []map[string]string{{}}

	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			b.WriteByte('<')
			b.WriteString(qualifiedName(tok.Name))
			currentNS := make(map[string]string, len(nsStack[len(nsStack)-1]))
			for prefix, value := range nsStack[len(nsStack)-1] {
				currentNS[prefix] = value
			}
			attrs := make([]xml.Attr, 0, len(tok.Attr))
			for _, attr := range tok.Attr {
				if isNamespaceDecl(attr) {
					prefix := attr.Name.Local
					if attr.Name.Local == "xmlns" {
						prefix = ""
					}
					value := strings.TrimSpace(attr.Value)
					if value == dsNamespace {
						prefix = "ds"
					}
					if currentNS[prefix] == value {
						continue
					}
					currentNS[prefix] = value
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
			nsStack = append(nsStack, currentNS)
		case xml.EndElement:
			b.WriteString("</")
			b.WriteString(qualifiedName(tok.Name))
			b.WriteByte('>')
			if len(nsStack) > 1 {
				nsStack = nsStack[:len(nsStack)-1]
			}
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

func stringPtr(v string) *string {
	return &v
}

func cteTStringPtr(v string) *distSchema.TString {
	value := distSchema.TString(v)
	return &value
}
