package mdfe_test

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

	distSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v1_0/dist_dfe"
	"github.com/awa/nota-fiscal/pkg/mdfe"
	"github.com/stretchr/testify/require"
)

const (
	mdfeNamespace = "http://www.portalfiscal.inf.br/mdfe"
	dsNamespace   = "http://www.w3.org/2000/09/xmldsig#"
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

			reparsed, err := mdfe.Parse(roundTripped)
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
		{name: "invalid mdfe", data: []byte(`<MDFe xmlns="http://www.portalfiscal.inf.br/mdfe"></MDFe>`), errContains: "missing infMDFe"},
		{name: "invalid consult nao encerrado", data: []byte(`<consMDFeNaoEnc xmlns="http://www.portalfiscal.inf.br/mdfe" versao="3.00"></consMDFeNaoEnc>`), errContains: "missing tpAmb"},
		{name: "invalid consult recibo", data: []byte(`<consReciMDFe xmlns="http://www.portalfiscal.inf.br/mdfe" versao="3.00"><tpAmb>1</tpAmb></consReciMDFe>`), errContains: "missing nRec"},
		{name: "invalid event", data: []byte(`<eventoMDFe xmlns="http://www.portalfiscal.inf.br/mdfe" versao="3.00"></eventoMDFe>`), errContains: "missing infEvento"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := mdfe.Parse(tt.data)
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
		assert func(t *testing.T, doc *mdfe.Document)
	}{
		{
			name: "distDFeInt",
			value: struct {
				XMLName xml.Name `xml:"distDFeInt"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distSchema.TAnonComplexDistDFeInt1
			}{
				XMLName: xml.Name{Local: "distDFeInt"},
				XMLNS:   mdfeNamespace,
				TAnonComplexDistDFeInt1: &distSchema.TAnonComplexDistDFeInt1{
					VersaoAttr: "1.00",
					TpAmb:      "1",
					CNPJ:       stringPtr("12345678000195"),
					DistNSU:    &distSchema.TAnonComplexDistNSU1{UltNSU: "000000000000001"},
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.DistDFeInt)
				require.Equal(t, "000000000000001", doc.DistDFeInt.DistNSU.UltNSU)
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
				XMLNS:   mdfeNamespace,
				TAnonComplexRetDistDFeInt1: &distSchema.TAnonComplexRetDistDFeInt1{
					VersaoAttr: "1.00",
					TpAmb:      "1",
					VerAplic:   mdfeTStringPtr("test"),
					CStat:      "138",
					XMotivo:    mdfeTStringPtr("Documento localizado"),
					DhResp:     "2024-01-02T03:04:05",
					UltNSU:     "000000000000010",
					MaxNSU:     "000000000000099",
					LoteDistDFeInt: &distSchema.TAnonComplexLoteDistDFeInt1{
						DocZip: []*distSchema.TAnonComplexDocZip1{{NSUAttr: "000000000000010", SchemaAttr: "procMDFe_v3.00.xsd", Value: "ZGF0YQ=="}},
					},
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
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

			doc, err := mdfe.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
		})
	}
}

func TestMarshalXML_NilReceiver(t *testing.T) {
	t.Parallel()

	var doc *mdfe.Document
	data, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Empty(t, data)
}

func assertFixtureShape(t *testing.T, fixture string, doc *mdfe.Document) {
	t.Helper()

	switch fixture {
	case "26999999999999999999999999999999999999999991-mdfe.xml":
		require.NotNil(t, doc.MDFe)
		require.Equal(t, "3.00", doc.MDFe.InfMDFe.VersaoAttr)
		require.Equal(t, "MDFe26200500000000000000222220000202631413000260", doc.MDFe.InfMDFe.IdAttr)
		require.Equal(t, "58", doc.MDFe.InfMDFe.Ide.Mod)
		require.Contains(t, doc.MDFe.InfMDFe.InfModal.InnerXML, "<rodo>")
	case "41190876676436000167580010000500001000437558-mdfe.xml":
		require.NotNil(t, doc.MDFe)
		require.Equal(t, "50000", doc.MDFe.InfMDFe.Ide.NMDF)
		require.Equal(t, "76676436000167", requirePtr(t, doc.MDFe.InfMDFe.Emit.CNPJ))
	case "ComPagtoPIX_41210780568835000181580010402005751006005791-procMDFe.xml":
		require.NotNil(t, doc.MDFe)
		require.Contains(t, doc.MDFe.InfMDFe.InfModal.InnerXML, "<rodo>")
		require.Contains(t, doc.MDFe.InfMDFe.InfModal.InnerXML, "<infPag>")
		require.Contains(t, doc.MDFe.InfMDFe.InfModal.InnerXML, "<PIX>")
	case "01010101010-ped-cons-mdfe-naoenc.xml":
		require.NotNil(t, doc.ConsNaoEnc)
		require.Equal(t, "2", doc.ConsNaoEnc.TpAmb)
		require.Equal(t, "55801377000131", requirePtr(t, doc.ConsNaoEnc.CNPJ))
	case "310000007934162-ped-rec.xml":
		require.NotNil(t, doc.ConsReciMDFe)
		require.Equal(t, "1", doc.ConsReciMDFe.TpAmb)
		require.Equal(t, "310000007934162", doc.ConsReciMDFe.NRec)
	case "cancelameto1101103511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110111", doc.EventoMDFe.InfEvento.TpEvento)
		require.Equal(t, "35110310290739000139550010000000011051128041", doc.EventoMDFe.InfEvento.ChMDFe)
		require.Contains(t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML, "<evCancMDFe>")
	case "encerramento1101123511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110112", doc.EventoMDFe.InfEvento.TpEvento)
		require.Contains(t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML, "<evEncMDFe>")
	case "inclusaocondutor31131223864838000129580000000000051003000003-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110114", doc.EventoMDFe.InfEvento.TpEvento)
		require.Contains(t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML, "<evIncCondutorMDFe>")
	case "inclusaoDFe1101154119060611747300015058001000000001111700344401-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110115", doc.EventoMDFe.InfEvento.TpEvento)
		require.Contains(t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML, "<evIncDFeMDFe>")
	case "PagamentoOperacaoMDFe_1101164120039999999999999958001000000999999999999901-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110116", doc.EventoMDFe.InfEvento.TpEvento)
		require.Contains(t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML, "<evPagtoOperMDFe>")
	case "pagamentoOperacao1101103511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110116", doc.EventoMDFe.InfEvento.TpEvento)
		require.Contains(t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML, "<evPagtoOperMDFe>")
	case "50170876063965000276580010000011311421039568-mdfe.xml":
		require.NotNil(t, doc.MDFe)
		require.Equal(t, "1131", doc.MDFe.InfMDFe.Ide.NMDF)
	default:
		t.Fatalf("unhandled fixture %s", fixture)
	}
}

func assertSameRoot(t *testing.T, expected, actual *mdfe.Document) {
	t.Helper()

	require.Equal(t, expected.MDFe != nil, actual.MDFe != nil)
	require.Equal(t, expected.ConsNaoEnc != nil, actual.ConsNaoEnc != nil)
	require.Equal(t, expected.ConsReciMDFe != nil, actual.ConsReciMDFe != nil)
	require.Equal(t, expected.EventoMDFe != nil, actual.EventoMDFe != nil)
}

func allFixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "..", "testdata", "mdfe", "v3_0"))
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

func parseFixture(t *testing.T, name string) *mdfe.Document {
	t.Helper()

	data := readFixture(t, name)
	doc, err := mdfe.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "mdfe", "v3_0", name))
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
	case "", mdfeNamespace:
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

func mdfeTStringPtr(v string) *distSchema.TString {
	value := distSchema.TString(v)
	return &value
}
