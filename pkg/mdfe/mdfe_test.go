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

type mdfeRodoModal struct {
	Rodo struct {
		InfANTT struct {
			RNTRC  string `xml:"RNTRC"`
			InfPag []struct {
				XNome     string `xml:"xNome"`
				CNPJ      string `xml:"CNPJ"`
				CPF       string `xml:"CPF"`
				VContrato string `xml:"vContrato"`
				IndPag    string `xml:"indPag"`
				Comp      []struct {
					TpComp string `xml:"tpComp"`
					VComp  string `xml:"vComp"`
					XComp  string `xml:"xComp"`
				} `xml:"Comp"`
				InfPrazo []struct {
					NParcela string `xml:"nParcela"`
					DVenc    string `xml:"dVenc"`
					VParcela string `xml:"vParcela"`
				} `xml:"infPrazo"`
				InfBanc struct {
					PIX        string `xml:"PIX"`
					CNPJIPEF   string `xml:"CNPJIPEF"`
					CodBanco   string `xml:"codBanco"`
					CodAgencia string `xml:"codAgencia"`
				} `xml:"infBanc"`
			} `xml:"infPag"`
		} `xml:"infANTT"`
		VeicTracao struct {
			Placa string `xml:"placa"`
			UF    string `xml:"UF"`
		} `xml:"veicTracao"`
	} `xml:"rodo"`
}

type mdfeCancEvento struct {
	EvCancMDFe struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		XJust      string `xml:"xJust"`
	} `xml:"evCancMDFe"`
}

type mdfeEncEvento struct {
	EvEncMDFe struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		DtEnc      string `xml:"dtEnc"`
		CUF        string `xml:"cUF"`
		CMun       string `xml:"cMun"`
	} `xml:"evEncMDFe"`
}

type mdfeIncCondutorEvento struct {
	EvIncCondutorMDFe struct {
		DescEvento string `xml:"descEvento"`
		Condutor   struct {
			XNome string `xml:"xNome"`
			CPF   string `xml:"CPF"`
		} `xml:"condutor"`
	} `xml:"evIncCondutorMDFe"`
}

type mdfeIncDFeEvento struct {
	EvIncDFeMDFe struct {
		DescEvento  string `xml:"descEvento"`
		NProt       string `xml:"nProt"`
		CMunCarrega string `xml:"cMunCarrega"`
		XMunCarrega string `xml:"xMunCarrega"`
		InfDoc      struct {
			CMunDescarga string `xml:"cMunDescarga"`
			XMunDescarga string `xml:"xMunDescarga"`
			ChNFe        string `xml:"chNFe"`
		} `xml:"infDoc"`
	} `xml:"evIncDFeMDFe"`
}

type mdfePagtoEvento struct {
	EvPagtoOperMDFe struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		InfViagens struct {
			QtdViagens string `xml:"qtdViagens"`
			NroViagem  string `xml:"nroViagem"`
		} `xml:"infViagens"`
		InfPag struct {
			XNome     string `xml:"xNome"`
			CNPJ      string `xml:"CNPJ"`
			VContrato string `xml:"vContrato"`
			IndPag    string `xml:"indPag"`
			VAdiant   string `xml:"vAdiant"`
			Comp      []struct {
				TpComp string `xml:"tpComp"`
				VComp  string `xml:"vComp"`
				XComp  string `xml:"xComp"`
			} `xml:"Comp"`
			InfPrazo []struct {
				NParcela string `xml:"nParcela"`
				DVenc    string `xml:"dVenc"`
				VParcela string `xml:"vParcela"`
			} `xml:"infPrazo"`
			InfBanc struct {
				PIX      string `xml:"PIX"`
				CNPJIPEF string `xml:"CNPJIPEF"`
			} `xml:"infBanc"`
		} `xml:"infPag"`
	} `xml:"evPagtoOperMDFe"`
}

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
		modal := decodeInnerXML[mdfeRodoModal](t, doc.MDFe.InfMDFe.InfModal.InnerXML)
		require.Equal(t, "12345678", modal.Rodo.InfANTT.RNTRC)
		require.Len(t, modal.Rodo.InfANTT.InfPag, 1)
		require.Equal(t, "99999999999", modal.Rodo.InfANTT.InfPag[0].CPF)
		require.Equal(t, "1000.00", modal.Rodo.InfANTT.InfPag[0].VContrato)
		require.Equal(t, "999", modal.Rodo.InfANTT.InfPag[0].InfBanc.CodBanco)
		require.Equal(t, "XXX9999", modal.Rodo.VeicTracao.Placa)
	case "41190876676436000167580010000500001000437558-mdfe.xml":
		require.NotNil(t, doc.MDFe)
		require.Equal(t, "50000", doc.MDFe.InfMDFe.Ide.NMDF)
		require.Equal(t, "76676436000167", requirePtr(t, doc.MDFe.InfMDFe.Emit.CNPJ))
	case "ComPagtoPIX_41210780568835000181580010402005751006005791-procMDFe.xml":
		require.NotNil(t, doc.MDFe)
		modal := decodeInnerXML[mdfeRodoModal](t, doc.MDFe.InfMDFe.InfModal.InnerXML)
		require.Equal(t, "00000000", modal.Rodo.InfANTT.RNTRC)
		require.Len(t, modal.Rodo.InfANTT.InfPag, 1)
		require.Equal(t, "XXXXXXXXXXXXXXXXXXXX", modal.Rodo.InfANTT.InfPag[0].XNome)
		require.Equal(t, "500.00", modal.Rodo.InfANTT.InfPag[0].VContrato)
		require.Equal(t, "0", modal.Rodo.InfANTT.InfPag[0].IndPag)
		require.Equal(t, "00000000000", modal.Rodo.InfANTT.InfPag[0].InfBanc.PIX)
		require.Equal(t, "XXX999", modal.Rodo.VeicTracao.Placa)
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
		evento := decodeInnerXML[mdfeCancEvento](t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Cancelamento", evento.EvCancMDFe.DescEvento)
		require.Equal(t, "010101010101010", evento.EvCancMDFe.NProt)
		require.Equal(t, "Justificativa do cancelamento", evento.EvCancMDFe.XJust)
	case "encerramento1101123511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110112", doc.EventoMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfeEncEvento](t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Encerramento", evento.EvEncMDFe.DescEvento)
		require.Equal(t, "010101010101010", evento.EvEncMDFe.NProt)
		require.Equal(t, "2013-10-31", evento.EvEncMDFe.DtEnc)
		require.Equal(t, "35", evento.EvEncMDFe.CUF)
		require.Equal(t, "4118402", evento.EvEncMDFe.CMun)
	case "inclusaocondutor31131223864838000129580000000000051003000003-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110114", doc.EventoMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfeIncCondutorEvento](t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Inclusao Condutor", evento.EvIncCondutorMDFe.DescEvento)
		require.Equal(t, "JOSE ALMEIDA", evento.EvIncCondutorMDFe.Condutor.XNome)
		require.Equal(t, "00000000191", evento.EvIncCondutorMDFe.Condutor.CPF)
	case "inclusaoDFe1101154119060611747300015058001000000001111700344401-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110115", doc.EventoMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfeIncDFeEvento](t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Inclusao DF-e", evento.EvIncDFeMDFe.DescEvento)
		require.Equal(t, "941190000014312", evento.EvIncDFeMDFe.NProt)
		require.Equal(t, "4118402", evento.EvIncDFeMDFe.CMunCarrega)
		require.Equal(t, "PARANAVAI", evento.EvIncDFeMDFe.XMunCarrega)
		require.Equal(t, "41190606117473000150550020000025691118027981", evento.EvIncDFeMDFe.InfDoc.ChNFe)
	case "PagamentoOperacaoMDFe_1101164120039999999999999958001000000999999999999901-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110116", doc.EventoMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfePagtoEvento](t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Pagamento Operacao MDF-e", evento.EvPagtoOperMDFe.DescEvento)
		require.Equal(t, "999999999999999", evento.EvPagtoOperMDFe.NProt)
		require.Equal(t, "7184", evento.EvPagtoOperMDFe.InfViagens.NroViagem)
		require.Len(t, evento.EvPagtoOperMDFe.InfPag.Comp, 3)
		require.Equal(t, "3000.00", evento.EvPagtoOperMDFe.InfPag.VContrato)
		require.Equal(t, "500.00", evento.EvPagtoOperMDFe.InfPag.VAdiant)
		require.Equal(t, "+5544993333223", evento.EvPagtoOperMDFe.InfPag.InfBanc.PIX)
	case "pagamentoOperacao1101103511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoMDFe)
		require.Equal(t, "110116", doc.EventoMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfePagtoEvento](t, doc.EventoMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Pagamento Operação MDF-e", evento.EvPagtoOperMDFe.DescEvento)
		require.Equal(t, "935200000016234", evento.EvPagtoOperMDFe.NProt)
		require.Equal(t, "1795", evento.EvPagtoOperMDFe.InfViagens.NroViagem)
		require.Len(t, evento.EvPagtoOperMDFe.InfPag.Comp, 1)
		require.Equal(t, "3003.51", evento.EvPagtoOperMDFe.InfPag.VContrato)
		require.Equal(t, "10290739000139", evento.EvPagtoOperMDFe.InfPag.InfBanc.CNPJIPEF)
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
	nsStack := []map[string]string{{}}

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
					if prefix == "" {
						attr = xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: value}
					} else {
						attr = xml.Attr{Name: xml.Name{Space: "xmlns", Local: prefix}, Value: value}
					}
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

func decodeInnerXML[T any](t *testing.T, fragment string) T {
	t.Helper()

	var value T
	data := []byte("<root>" + fragment + "</root>")
	err := xml.Unmarshal(data, &value)
	require.NoError(t, err)
	return value
}

func stringPtr(v string) *string {
	return &v
}

func mdfeTStringPtr(v string) *distSchema.TString {
	value := distSchema.TString(v)
	return &value
}
