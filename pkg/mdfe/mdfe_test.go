package mdfe_test

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	distSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v1_0/dist_dfe"
	consNaoEncSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/cons_nao_enc"
	consReciSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/cons_reci"
	consultaDFESchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/consulta_dfe"
	consSitSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/consulta_situacao"
	distMDFeSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/dist_mdfe"
	mdfeSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/mdfe"
	statusSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/status_servico"
	"github.com/awafinance/fiscal/pkg/mdfe"
	"github.com/stretchr/testify/require"
)

const (
	mdfeNamespace   = "http://www.portalfiscal.inf.br/mdfe"
	dsNamespace     = "http://www.w3.org/2000/09/xmldsig#"
	mdfeDocumentKey = "41240112345678000195580010000000011000000011"
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

func TestParse_SupportedRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  any
		assert func(t *testing.T, doc *mdfe.Document)
	}{
		{
			name: "mdfeProc",
			value: struct {
				XMLName    xml.Name              `xml:"mdfeProc"`
				XMLNS      string                `xml:"xmlns,attr"`
				VersaoAttr string                `xml:"versao,attr"`
				MDFe       *mdfeSchema.TMDFe     `xml:"MDFe"`
				ProtMDFe   *mdfeSchema.TProtMDFe `xml:"protMDFe"`
			}{
				XMLName:    xml.Name{Local: "mdfeProc"},
				XMLNS:      mdfeNamespace,
				VersaoAttr: "3.00",
				MDFe:       minimalMDFe(),
				ProtMDFe:   minimalProtMDFe(),
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.MDFeProc)
				require.NotNil(t, doc.MDFeProc.MDFe)
				require.NotNil(t, doc.MDFeProc.ProtMDFe)
			},
		},
		{
			name: "enviMDFe",
			value: struct {
				XMLName xml.Name `xml:"enviMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*mdfeSchema.TEnviMDFe
			}{
				XMLName: xml.Name{Local: "enviMDFe"},
				XMLNS:   mdfeNamespace,
				TEnviMDFe: &mdfeSchema.TEnviMDFe{
					VersaoAttr: "3.00",
					IdLote:     "1",
					MDFe:       minimalMDFe(),
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.EnviMDFe)
				require.Equal(t, "1", doc.EnviMDFe.IdLote)
			},
		},
		{
			name: "retEnviMDFe",
			value: struct {
				XMLName xml.Name `xml:"retEnviMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*mdfeSchema.TRetEnviMDFe
			}{
				XMLName: xml.Name{Local: "retEnviMDFe"},
				XMLNS:   mdfeNamespace,
				TRetEnviMDFe: &mdfeSchema.TRetEnviMDFe{
					VersaoAttr: "3.00",
					TpAmb:      stringPtr("2"),
					CUF:        "41",
					CStat:      "103",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetEnviMDFe)
				require.Equal(t, "103", doc.RetEnviMDFe.CStat)
			},
		},
		{
			name: "retMDFe",
			value: struct {
				XMLName xml.Name `xml:"retMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*mdfeSchema.TRetMDFe
			}{
				XMLName: xml.Name{Local: "retMDFe"},
				XMLNS:   mdfeNamespace,
				TRetMDFe: &mdfeSchema.TRetMDFe{
					VersaoAttr: "3.00",
					TpAmb:      stringPtr("2"),
					CUF:        "41",
					CStat:      "100",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetMDFe)
				require.Equal(t, "100", doc.RetMDFe.CStat)
			},
		},
		{
			name: "retConsMDFeNaoEnc",
			value: struct {
				XMLName xml.Name `xml:"retConsMDFeNaoEnc"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consNaoEncSchema.TRetConsMDFeNaoEnc
			}{
				XMLName: xml.Name{Local: "retConsMDFeNaoEnc"},
				XMLNS:   mdfeNamespace,
				TRetConsMDFeNaoEnc: &consNaoEncSchema.TRetConsMDFeNaoEnc{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					CUF:        "41",
					CStat:      "111",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetConsNaoEnc)
				require.Equal(t, "111", doc.RetConsNaoEnc.CStat)
			},
		},
		{
			name: "retConsReciMDFe",
			value: struct {
				XMLName xml.Name `xml:"retConsReciMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consReciSchema.TRetConsReciMDFe
			}{
				XMLName: xml.Name{Local: "retConsReciMDFe"},
				XMLNS:   mdfeNamespace,
				TRetConsReciMDFe: &consReciSchema.TRetConsReciMDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					NRec:       "410000000000001",
					CUF:        "41",
					CStat:      "104",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetConsReciMDFe)
				require.Equal(t, "104", doc.RetConsReciMDFe.CStat)
			},
		},
		{
			name: "consSitMDFe",
			value: struct {
				XMLName xml.Name `xml:"consSitMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consSitSchema.TConsSitMDFe
			}{
				XMLName: xml.Name{Local: "consSitMDFe"},
				XMLNS:   mdfeNamespace,
				TConsSitMDFe: &consSitSchema.TConsSitMDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					ChMDFe:     mdfeDocumentKey,
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.ConsSitMDFe)
				require.Equal(t, mdfeDocumentKey, doc.ConsSitMDFe.ChMDFe)
			},
		},
		{
			name: "retConsSitMDFe",
			value: struct {
				XMLName xml.Name `xml:"retConsSitMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consSitSchema.TRetConsSitMDFe
			}{
				XMLName: xml.Name{Local: "retConsSitMDFe"},
				XMLNS:   mdfeNamespace,
				TRetConsSitMDFe: &consSitSchema.TRetConsSitMDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					CUF:        "41",
					CStat:      "100",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetConsSitMDFe)
				require.Equal(t, "100", doc.RetConsSitMDFe.CStat)
			},
		},
		{
			name: "consStatServMDFe",
			value: struct {
				XMLName xml.Name `xml:"consStatServMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*statusSchema.TConsStatServ
			}{
				XMLName: xml.Name{Local: "consStatServMDFe"},
				XMLNS:   mdfeNamespace,
				TConsStatServ: &statusSchema.TConsStatServ{
					VersaoAttr: "3.00",
					TpAmb:      "2",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.ConsStatServMDFe)
				require.Equal(t, "2", doc.ConsStatServMDFe.TpAmb)
			},
		},
		{
			name: "retConsStatServMDFe",
			value: struct {
				XMLName xml.Name `xml:"retConsStatServMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*statusSchema.TRetConsStatServ
			}{
				XMLName: xml.Name{Local: "retConsStatServMDFe"},
				XMLNS:   mdfeNamespace,
				TRetConsStatServ: &statusSchema.TRetConsStatServ{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					CUF:        "41",
					CStat:      "107",
					DhRecbto:   "2024-01-02T03:04:05-03:00",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetConsStatServMDFe)
				require.Equal(t, "107", doc.RetConsStatServMDFe.CStat)
			},
		},
		{
			name: "distMDFe",
			value: struct {
				XMLName xml.Name `xml:"distMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distMDFeSchema.TDistDFe
			}{
				XMLName: xml.Name{Local: "distMDFe"},
				XMLNS:   mdfeNamespace,
				TDistDFe: &distMDFeSchema.TDistDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					UltNSU:     "000000000000001",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.DistMDFe)
				require.Equal(t, "000000000000001", doc.DistMDFe.UltNSU)
			},
		},
		{
			name: "retDistMDFe",
			value: struct {
				XMLName xml.Name `xml:"retDistMDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distMDFeSchema.TRetDistDFe
			}{
				XMLName: xml.Name{Local: "retDistMDFe"},
				XMLNS:   mdfeNamespace,
				TRetDistDFe: &distMDFeSchema.TRetDistDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					CStat:      "138",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetDistMDFe)
				require.Equal(t, "138", doc.RetDistMDFe.CStat)
			},
		},
		{
			name: "mdfeConsultaDFe",
			value: struct {
				XMLName xml.Name `xml:"mdfeConsultaDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consultaDFESchema.TMDFeConsultaDFe
			}{
				XMLName: xml.Name{Local: "mdfeConsultaDFe"},
				XMLNS:   mdfeNamespace,
				TMDFeConsultaDFe: &consultaDFESchema.TMDFeConsultaDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					ChMDFe:     mdfeDocumentKey,
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.MDFeConsultaDFe)
				require.Equal(t, mdfeDocumentKey, doc.MDFeConsultaDFe.ChMDFe)
			},
		},
		{
			name: "retMDFeConsultaDFe",
			value: struct {
				XMLName xml.Name `xml:"retMDFeConsultaDFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consultaDFESchema.TRetMDFeConsultaDFe
			}{
				XMLName: xml.Name{Local: "retMDFeConsultaDFe"},
				XMLNS:   mdfeNamespace,
				TRetMDFeConsultaDFe: &consultaDFESchema.TRetMDFeConsultaDFe{
					VersaoAttr: "3.00",
					TpAmb:      "2",
					CStat:      "100",
				},
			},
			assert: func(t *testing.T, doc *mdfe.Document) {
				require.NotNil(t, doc.RetMDFeConsultaDFe)
				require.Equal(t, "100", doc.RetMDFeConsultaDFe.CStat)
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

func TestParse_EventReturnAndProcRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		assert func(t *testing.T, doc *mdfe.Document)
	}{
		{name: "ret generico", data: minimalMDFERetEventXML("990001"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoMDFe) }},
		{name: "ret cancelamento", data: minimalMDFERetEventXML("110111"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoCancMDFe) }},
		{name: "ret encerramento", data: minimalMDFERetEventXML("110112"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoEncMDFe) }},
		{name: "ret inclusao condutor", data: minimalMDFERetEventXML("110114"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoIncCondutorMDFe) }},
		{name: "ret inclusao dfe", data: minimalMDFERetEventXML("110115"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoInclusaoDFeMDFe) }},
		{name: "ret pagamento operacao", data: minimalMDFERetEventXML("110116"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoPagtoOperMDFe) }},
		{name: "ret confirma servico", data: minimalMDFERetEventXML("110117"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoConfirmaServMDFe) }},
		{name: "ret alteracao pagamento servico", data: minimalMDFERetEventXML("110118"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.RetEventoAlteracaoPagtoServMDFe) }},
		{name: "proc generico", data: minimalMDFEProcEventXML("990001"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoMDFe) }},
		{name: "proc cancelamento", data: minimalMDFEProcEventXML("110111"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoCancMDFe) }},
		{name: "proc encerramento", data: minimalMDFEProcEventXML("110112"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoEncMDFe) }},
		{name: "proc inclusao condutor", data: minimalMDFEProcEventXML("110114"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoIncCondutorMDFe) }},
		{name: "proc inclusao dfe", data: minimalMDFEProcEventXML("110115"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoInclusaoDFeMDFe) }},
		{name: "proc pagamento operacao", data: minimalMDFEProcEventXML("110116"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoPagtoOperMDFe) }},
		{name: "proc confirma servico", data: minimalMDFEProcEventXML("110117"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoConfirmaServMDFe) }},
		{name: "proc alteracao pagamento servico", data: minimalMDFEProcEventXML("110118"), assert: func(t *testing.T, doc *mdfe.Document) { require.NotNil(t, doc.ProcEventoAlteracaoPagtoServMDFe) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := mdfe.Parse([]byte(tt.data))
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)

			reparsed, err := mdfe.Parse(roundTripped)
			require.NoError(t, err)
			tt.assert(t, reparsed)
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
		require.NotNil(t, doc.EventoCancMDFe)
		require.Equal(t, "110111", doc.EventoCancMDFe.InfEvento.TpEvento)
		require.Equal(t, "35110310290739000139550010000000011051128041", doc.EventoCancMDFe.InfEvento.ChMDFe)
		evento := decodeInnerXML[mdfeCancEvento](t, doc.EventoCancMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Cancelamento", evento.EvCancMDFe.DescEvento)
		require.Equal(t, "010101010101010", evento.EvCancMDFe.NProt)
		require.Equal(t, "Justificativa do cancelamento", evento.EvCancMDFe.XJust)
	case "encerramento1101123511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoEncMDFe)
		require.Equal(t, "110112", doc.EventoEncMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfeEncEvento](t, doc.EventoEncMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Encerramento", evento.EvEncMDFe.DescEvento)
		require.Equal(t, "010101010101010", evento.EvEncMDFe.NProt)
		require.Equal(t, "2013-10-31", evento.EvEncMDFe.DtEnc)
		require.Equal(t, "35", evento.EvEncMDFe.CUF)
		require.Equal(t, "4118402", evento.EvEncMDFe.CMun)
	case "inclusaocondutor31131223864838000129580000000000051003000003-ped-eve.xml":
		require.NotNil(t, doc.EventoIncCondutorMDFe)
		require.Equal(t, "110114", doc.EventoIncCondutorMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfeIncCondutorEvento](t, doc.EventoIncCondutorMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Inclusao Condutor", evento.EvIncCondutorMDFe.DescEvento)
		require.Equal(t, "JOSE ALMEIDA", evento.EvIncCondutorMDFe.Condutor.XNome)
		require.Equal(t, "00000000191", evento.EvIncCondutorMDFe.Condutor.CPF)
	case "inclusaoDFe1101154119060611747300015058001000000001111700344401-ped-eve.xml":
		require.NotNil(t, doc.EventoInclusaoDFeMDFe)
		require.Equal(t, "110115", doc.EventoInclusaoDFeMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfeIncDFeEvento](t, doc.EventoInclusaoDFeMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Inclusao DF-e", evento.EvIncDFeMDFe.DescEvento)
		require.Equal(t, "941190000014312", evento.EvIncDFeMDFe.NProt)
		require.Equal(t, "4118402", evento.EvIncDFeMDFe.CMunCarrega)
		require.Equal(t, "PARANAVAI", evento.EvIncDFeMDFe.XMunCarrega)
		require.Equal(t, "41190606117473000150550020000025691118027981", evento.EvIncDFeMDFe.InfDoc.ChNFe)
	case "PagamentoOperacaoMDFe_1101164120039999999999999958001000000999999999999901-ped-eve.xml":
		require.NotNil(t, doc.EventoPagtoOperMDFe)
		require.Equal(t, "110116", doc.EventoPagtoOperMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfePagtoEvento](t, doc.EventoPagtoOperMDFe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Pagamento Operacao MDF-e", evento.EvPagtoOperMDFe.DescEvento)
		require.Equal(t, "999999999999999", evento.EvPagtoOperMDFe.NProt)
		require.Equal(t, "7184", evento.EvPagtoOperMDFe.InfViagens.NroViagem)
		require.Len(t, evento.EvPagtoOperMDFe.InfPag.Comp, 3)
		require.Equal(t, "3000.00", evento.EvPagtoOperMDFe.InfPag.VContrato)
		require.Equal(t, "500.00", evento.EvPagtoOperMDFe.InfPag.VAdiant)
		require.Equal(t, "+5544993333223", evento.EvPagtoOperMDFe.InfPag.InfBanc.PIX)
	case "pagamentoOperacao1101103511031029073900013955001000000001105112804101-ped-eve.xml":
		require.NotNil(t, doc.EventoPagtoOperMDFe)
		require.Equal(t, "110116", doc.EventoPagtoOperMDFe.InfEvento.TpEvento)
		evento := decodeInnerXML[mdfePagtoEvento](t, doc.EventoPagtoOperMDFe.InfEvento.DetEvento.InnerXML)
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
	require.Equal(t, expected.MDFeProc != nil, actual.MDFeProc != nil)
	require.Equal(t, expected.EnviMDFe != nil, actual.EnviMDFe != nil)
	require.Equal(t, expected.RetEnviMDFe != nil, actual.RetEnviMDFe != nil)
	require.Equal(t, expected.RetMDFe != nil, actual.RetMDFe != nil)
	require.Equal(t, expected.ConsNaoEnc != nil, actual.ConsNaoEnc != nil)
	require.Equal(t, expected.RetConsNaoEnc != nil, actual.RetConsNaoEnc != nil)
	require.Equal(t, expected.ConsReciMDFe != nil, actual.ConsReciMDFe != nil)
	require.Equal(t, expected.RetConsReciMDFe != nil, actual.RetConsReciMDFe != nil)
	require.Equal(t, expected.EventoMDFe != nil, actual.EventoMDFe != nil)
	require.Equal(t, expected.EventoCancMDFe != nil, actual.EventoCancMDFe != nil)
	require.Equal(t, expected.EventoEncMDFe != nil, actual.EventoEncMDFe != nil)
	require.Equal(t, expected.EventoIncCondutorMDFe != nil, actual.EventoIncCondutorMDFe != nil)
	require.Equal(t, expected.EventoInclusaoDFeMDFe != nil, actual.EventoInclusaoDFeMDFe != nil)
	require.Equal(t, expected.EventoPagtoOperMDFe != nil, actual.EventoPagtoOperMDFe != nil)
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

func minimalMDFERetEventXML(tpEvento string) string {
	return fmt.Sprintf(`<retEventoMDFe xmlns="%s" versao="3.00"><infEvento><tpAmb>2</tpAmb><cStat>135</cStat><tpEvento>%s</tpEvento></infEvento></retEventoMDFe>`, mdfeNamespace, tpEvento)
}

func minimalMDFEProcEventXML(tpEvento string) string {
	return fmt.Sprintf(`<procEventoMDFe xmlns="%s" versao="3.00"><eventoMDFe versao="3.00"><infEvento Id="ID%s4124011234567800019558001000000001100000001101"><cOrgao>41</cOrgao><tpAmb>2</tpAmb><CNPJ>12345678000195</CNPJ><chMDFe>%s</chMDFe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>%s</tpEvento><nSeqEvento>1</nSeqEvento><detEvento></detEvento></infEvento></eventoMDFe><retEventoMDFe versao="3.00"><infEvento><tpAmb>2</tpAmb><cStat>135</cStat><tpEvento>%s</tpEvento></infEvento></retEventoMDFe></procEventoMDFe>`, mdfeNamespace, tpEvento, mdfeDocumentKey, tpEvento, tpEvento)
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

func minimalMDFe() *mdfeSchema.TMDFe {
	return &mdfeSchema.TMDFe{
		InfMDFe: &mdfeSchema.TAnonComplexInfMDFe1{
			VersaoAttr: "3.00",
			IdAttr:     "MDFe" + mdfeDocumentKey,
			Ide: &mdfeSchema.TAnonComplexIde1{
				CUF:     "41",
				TpAmb:   "2",
				Mod:     "58",
				Serie:   "1",
				NMDF:    "1",
				CMDF:    "12345678",
				CDV:     "1",
				Modal:   "1",
				DhEmi:   "2024-01-02T03:04:05-03:00",
				TpEmis:  "1",
				ProcEmi: "0",
				VerProc: "test",
				UFIni:   "PR",
				UFFim:   "SP",
				InfMunCarrega: []*mdfeSchema.TAnonComplexInfMunCarrega1{{
					CMunCarrega: "4106902",
					XMunCarrega: "CURITIBA",
				}},
			},
			Emit: &mdfeSchema.TAnonComplexEmit1{
				CNPJ:  stringPtr("12345678000195"),
				XNome: "Emitente",
			},
			InfModal: &mdfeSchema.TAnonComplexInfModal1{VersaoModalAttr: "3.00", InnerXML: "<rodo></rodo>"},
		},
	}
}

func minimalProtMDFe() *mdfeSchema.TProtMDFe {
	return &mdfeSchema.TProtMDFe{
		InfProt: &mdfeSchema.TAnonComplexInfProt1{
			TpAmb:    "2",
			ChMDFe:   mdfeDocumentKey,
			DhRecbto: "2024-01-02T03:04:05-03:00",
			CStat:    "100",
		},
	}
}

func mdfeTStringPtr(v string) *distSchema.TString {
	value := distSchema.TString(v)
	return &value
}
