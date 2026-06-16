package bpe_test

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

	schema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/core"
	alteracaoPoltronaSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_alteracao_poltrona"
	excessoBagagemSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_excesso_bagagem"
	naoEmbSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_nao_emb"
	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/awafinance/fiscal/pkg/fiscalerr"
	"github.com/stretchr/testify/require"
)

const (
	bpeNamespace = "http://www.portalfiscal.inf.br/bpe"
	dsNamespace  = "http://www.w3.org/2000/09/xmldsig#"
)

type bpeCancEvento struct {
	EvCancBPe struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		XJust      string `xml:"xJust"`
	} `xml:"evCancBPe"`
}

type bpeAlteracaoPoltronaEvento struct {
	EvAlteracaoPoltrona struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		Poltrona   string `xml:"poltrona"`
	} `xml:"evAlteracaoPoltrona"`
}

type bpeExcessoBagagemEvento struct {
	EvExcessoBagagem struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		QBagagem   string `xml:"qBagagem"`
		VTotBag    string `xml:"vTotBag"`
	} `xml:"evExcessoBagagem"`
}

type bpeNaoEmbEvento struct {
	EvNaoEmbBPe struct {
		DescEvento string `xml:"descEvento"`
		NProt      string `xml:"nProt"`
		XJust      string `xml:"xJust"`
	} `xml:"evNaoEmbBPe"`
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
		})
	}
}

func TestParse_SupportedRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  any
		assert func(t *testing.T, doc *bpe.Document)
	}{
		{
			name: "BPe",
			value: struct {
				XMLName    xml.Name                        `xml:"BPe"`
				XMLNS      string                          `xml:"xmlns,attr"`
				InfBPe     *schema.TAnonComplexInfBPe2     `xml:"infBPe"`
				InfBPeSupl *schema.TAnonComplexInfBPeSupl1 `xml:"infBPeSupl,omitempty"`
			}{
				XMLName:    xml.Name{Local: "BPe"},
				XMLNS:      bpeNamespace,
				InfBPe:     minimalInfBPe(),
				InfBPeSupl: &schema.TAnonComplexInfBPeSupl1{QrCodBPe: "qr"},
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.BPe)
				require.Equal(t, "12345678000195", doc.BPe.InfBPe.Emit.CNPJ)
			},
		},
		{
			name: "BPeTM",
			value: struct {
				XMLName xml.Name                    `xml:"BPeTM"`
				XMLNS   string                      `xml:"xmlns,attr"`
				InfBPe  *schema.TAnonComplexInfBPe1 `xml:"infBPe"`
			}{
				XMLName: xml.Name{Local: "BPeTM"},
				XMLNS:   bpeNamespace,
				InfBPe:  minimalInfBPeTM(),
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.BPeTM)
				require.Len(t, doc.BPeTM.InfBPe.DetBPeTM, 1)
			},
		},
		{
			name: "bpeProc",
			value: struct {
				XMLName    xml.Name         `xml:"bpeProc"`
				XMLNS      string           `xml:"xmlns,attr"`
				VersaoAttr string           `xml:"versao,attr"`
				BPe        *schema.TBPe     `xml:"BPe"`
				ProtBPe    *schema.TProtBPe `xml:"protBPe"`
			}{
				XMLName:    xml.Name{Local: "bpeProc"},
				XMLNS:      bpeNamespace,
				VersaoAttr: "1.00",
				BPe:        &schema.TBPe{InfBPe: minimalInfBPe()},
				ProtBPe:    minimalProtBPe(),
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.BPeProc)
				require.NotNil(t, doc.BPeProc.ProtBPe)
			},
		},
		{
			name: "retBPe",
			value: struct {
				XMLName xml.Name `xml:"retBPe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*schema.TRetBPe
			}{
				XMLName: xml.Name{Local: "retBPe"},
				XMLNS:   bpeNamespace,
				TRetBPe: &schema.TRetBPe{VersaoAttr: "1.00", TpAmb: "2", CUF: "43", CStat: "100"},
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.RetBPe)
				require.Equal(t, "100", doc.RetBPe.CStat)
			},
		},
		{
			name: "consSitBPe",
			value: struct {
				XMLName xml.Name `xml:"consSitBPe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*schema.TConsSitBPe
			}{
				XMLName:     xml.Name{Local: "consSitBPe"},
				XMLNS:       bpeNamespace,
				TConsSitBPe: &schema.TConsSitBPe{VersaoAttr: "1.00", TpAmb: "2", ChBPe: documentKey},
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.ConsSitBPe)
				require.Equal(t, documentKey, doc.ConsSitBPe.ChBPe)
			},
		},
		{
			name: "retConsStatServBPe",
			value: struct {
				XMLName xml.Name `xml:"retConsStatServBPe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*schema.TRetConsStatServ
			}{
				XMLName: xml.Name{Local: "retConsStatServBPe"},
				XMLNS:   bpeNamespace,
				TRetConsStatServ: &schema.TRetConsStatServ{
					VersaoAttr: "1.00",
					TpAmb:      "2",
					CUF:        "43",
					CStat:      "107",
					DhRecbto:   "2024-01-02T03:04:05-03:00",
					VerAplic:   "app",
				},
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.RetConsStatServBPe)
				require.Equal(t, "107", doc.RetConsStatServBPe.CStat)
			},
		},
		{
			name: "eventoBPe",
			value: struct {
				XMLName    xml.Name                       `xml:"eventoBPe"`
				XMLNS      string                         `xml:"xmlns,attr"`
				VersaoAttr string                         `xml:"versao,attr"`
				InfEvento  *schema.TAnonComplexInfEvento1 `xml:"infEvento"`
			}{
				XMLName:    xml.Name{Local: "eventoBPe"},
				XMLNS:      bpeNamespace,
				VersaoAttr: "1.00",
				InfEvento:  minimalEventoInf(),
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoCancBPe)
				evento := decodeInnerXML[bpeCancEvento](t, doc.EventoCancBPe.InfEvento.DetEvento.InnerXML)
				require.Equal(t, "Cancelamento", evento.EvCancBPe.DescEvento)
				require.Equal(t, "123456789012345", evento.EvCancBPe.NProt)
				require.Equal(t, "Justificativa de cancelamento", evento.EvCancBPe.XJust)
			},
		},
		{
			name: "procEventoBPe",
			value: struct {
				XMLName      xml.Name           `xml:"procEventoBPe"`
				XMLNS        string             `xml:"xmlns,attr"`
				VersaoAttr   string             `xml:"versao,attr"`
				EventoBPe    *schema.TEvento    `xml:"eventoBPe"`
				RetEventoBPe *schema.TRetEvento `xml:"retEventoBPe"`
			}{
				XMLName:      xml.Name{Local: "procEventoBPe"},
				XMLNS:        bpeNamespace,
				VersaoAttr:   "1.00",
				EventoBPe:    &schema.TEvento{VersaoAttr: "1.00", InfEvento: minimalEventoInf()},
				RetEventoBPe: &schema.TRetEvento{VersaoAttr: "1.00", InfEvento: &schema.TAnonComplexInfEvento2{TpAmb: "2", CStat: "135"}},
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.ProcEventoCancBPe)
				require.NotNil(t, doc.ProcEventoCancBPe.EventoBPe)
				require.NotNil(t, doc.ProcEventoCancBPe.RetEventoBPe)
			},
		},
		{
			name: "eventoBPe alteracao poltrona",
			value: struct {
				XMLName    xml.Name                                        `xml:"eventoBPe"`
				XMLNS      string                                          `xml:"xmlns,attr"`
				VersaoAttr string                                          `xml:"versao,attr"`
				InfEvento  *alteracaoPoltronaSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
			}{
				XMLName:    xml.Name{Local: "eventoBPe"},
				XMLNS:      bpeNamespace,
				VersaoAttr: "1.00",
				InfEvento:  minimalAlteracaoPoltronaEventoInf(),
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoAlteracaoPoltrona)
				evento := decodeInnerXML[bpeAlteracaoPoltronaEvento](t, doc.EventoAlteracaoPoltrona.InfEvento.DetEvento.InnerXML)
				require.Equal(t, "110116", doc.EventoAlteracaoPoltrona.InfEvento.TpEvento)
				require.Equal(t, "42A", evento.EvAlteracaoPoltrona.Poltrona)
			},
		},
		{
			name: "eventoBPe excesso bagagem",
			value: struct {
				XMLName    xml.Name                                     `xml:"eventoBPe"`
				XMLNS      string                                       `xml:"xmlns,attr"`
				VersaoAttr string                                       `xml:"versao,attr"`
				InfEvento  *excessoBagagemSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
			}{
				XMLName:    xml.Name{Local: "eventoBPe"},
				XMLNS:      bpeNamespace,
				VersaoAttr: "1.00",
				InfEvento:  minimalExcessoBagagemEventoInf(),
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoExcessoBagagem)
				evento := decodeInnerXML[bpeExcessoBagagemEvento](t, doc.EventoExcessoBagagem.InfEvento.DetEvento.InnerXML)
				require.Equal(t, "2", evento.EvExcessoBagagem.QBagagem)
				require.Equal(t, "120.00", evento.EvExcessoBagagem.VTotBag)
			},
		},
		{
			name: "eventoBPe nao embarque",
			value: struct {
				XMLName    xml.Name                             `xml:"eventoBPe"`
				XMLNS      string                               `xml:"xmlns,attr"`
				VersaoAttr string                               `xml:"versao,attr"`
				InfEvento  *naoEmbSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
			}{
				XMLName:    xml.Name{Local: "eventoBPe"},
				XMLNS:      bpeNamespace,
				VersaoAttr: "1.00",
				InfEvento:  minimalNaoEmbEventoInf(),
			},
			assert: func(t *testing.T, doc *bpe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoNaoEmbBPe)
				evento := decodeInnerXML[bpeNaoEmbEvento](t, doc.EventoNaoEmbBPe.InfEvento.DetEvento.InnerXML)
				require.Equal(t, "Passageiro ausente", evento.EvNaoEmbBPe.XJust)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := xml.MarshalIndent(tt.value, "", "  ")
			require.NoError(t, err)

			doc, err := bpe.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
		})
	}
}

func TestParse_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		data        []byte
		errContains string
		errIs       error
	}{
		{name: "empty", data: nil, errIs: fiscalerr.ErrEmptyDocument},
		{name: "unsupported root", data: []byte(`<foo></foo>`), errIs: fiscalerr.ErrUnsupportedRoot},
		{name: "invalid bpe", data: []byte(`<BPe xmlns="http://www.portalfiscal.inf.br/bpe"></BPe>`), errContains: "missing infBPe"},
		{name: "invalid event", data: []byte(`<eventoBPe xmlns="http://www.portalfiscal.inf.br/bpe" versao="1.00"></eventoBPe>`), errContains: "missing infEvento"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := bpe.Parse(tt.data)
			require.Error(t, err)
			require.Nil(t, doc)
			if tt.errIs != nil {
				require.ErrorIs(t, err, tt.errIs)
			} else {
				require.ErrorContains(t, err, tt.errContains)
			}
		})
	}
}

func TestParse_EventReturnAndProcRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		assert func(t *testing.T, doc *bpe.Document)
	}{
		{name: "ret generico", data: minimalBPERetEventXML("990001"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.RetEventoBPe)
			assertBPEEventReturnInfo(t, doc, "990001")
		}},
		{name: "ret cancelamento", data: minimalBPERetEventXML("110111"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.RetEventoCancBPe)
			assertBPEEventReturnInfo(t, doc, "110111")
		}},
		{name: "ret nao embarque", data: minimalBPERetEventXML("110115"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.RetEventoNaoEmbBPe)
		}},
		{name: "ret alteracao poltrona", data: minimalBPERetEventXML("110116"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.RetEventoAlteracaoPoltrona)
		}},
		{name: "ret excesso bagagem", data: minimalBPERetEventXML("110117"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.RetEventoExcessoBagagem)
		}},
		{name: "proc generico", data: minimalBPEProcEventXML("990001"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.ProcEventoBPe)
			assertBPEProcessedEventInfo(t, doc, "990001")
		}},
		{name: "proc cancelamento", data: minimalBPEProcEventXML("110111"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.ProcEventoCancBPe)
			assertBPEProcessedEventInfo(t, doc, "110111")
		}},
		{name: "proc nao embarque", data: minimalBPEProcEventXML("110115"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.ProcEventoNaoEmbBPe)
		}},
		{name: "proc alteracao poltrona", data: minimalBPEProcEventXML("110116"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.ProcEventoAlteracaoPoltrona)
		}},
		{name: "proc excesso bagagem", data: minimalBPEProcEventXML("110117"), assert: func(t *testing.T, doc *bpe.Document) {
			t.Helper()
			require.NotNil(t, doc.ProcEventoExcessoBagagem)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := bpe.Parse([]byte(tt.data))
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)

			reparsed, err := bpe.Parse(roundTripped)
			require.NoError(t, err)
			tt.assert(t, reparsed)
		})
	}
}

func TestMarshalXML_NilReceiver(t *testing.T) {
	t.Parallel()

	var doc *bpe.Document
	data, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Empty(t, data)
}

func assertFixtureShape(t *testing.T, fixture string, doc *bpe.Document) {
	t.Helper()

	switch fixture {
	case "43190812345678000195630010000000011000000011-bpe.xml":
		require.NotNil(t, doc.BPe)
		require.Equal(t, "1.00", doc.BPe.InfBPe.VersaoAttr)
		require.Equal(t, "BPe43190812345678000195630010000000011000000011", doc.BPe.InfBPe.IdAttr)
		require.Equal(t, "43", doc.BPe.InfBPe.Ide.CUF)
		require.Equal(t, "12345678000195", doc.BPe.InfBPe.Emit.CNPJ)
		require.Len(t, doc.BPe.InfBPe.InfViagem, 1)
	case "43190812345678000195630010000000011000000011-bpeProc.xml":
		require.NotNil(t, doc.BPeProc)
		require.NotNil(t, doc.BPeProc.BPe)
		require.NotNil(t, doc.BPeProc.ProtBPe)
		require.Equal(t, "43190812345678000195630010000000011000000011", doc.BPeProc.ProtBPe.InfProt.ChBPe)
		require.Equal(t, "100", doc.BPeProc.ProtBPe.InfProt.CStat)
	case "1101114319081234567800019563001000000001100000001101-eventoBPe.xml":
		require.NotNil(t, doc.EventoCancBPe)
		require.Equal(t, "110111", doc.EventoCancBPe.InfEvento.TpEvento)
		require.Equal(t, "43190812345678000195630010000000011000000011", doc.EventoCancBPe.InfEvento.ChBPe)
		evento := decodeInnerXML[bpeCancEvento](t, doc.EventoCancBPe.InfEvento.DetEvento.InnerXML)
		require.Equal(t, "Cancelamento", evento.EvCancBPe.DescEvento)
		require.Equal(t, "123456789012345", evento.EvCancBPe.NProt)
		require.Equal(t, "Justificativa de cancelamento", evento.EvCancBPe.XJust)
	default:
		t.Fatalf("unhandled fixture %s", fixture)
	}
}

func allFixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "..", "testdata", "bpe", "v1_0"))
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

func parseFixture(t *testing.T, name string) *bpe.Document {
	t.Helper()

	data := readFixture(t, name)
	doc, err := bpe.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "bpe", "v1_0", name))
	require.NoError(t, err)
	return data
}

func minimalInfBPe() *schema.TAnonComplexInfBPe2 {
	return &schema.TAnonComplexInfBPe2{
		VersaoAttr: "1.00",
		IdAttr:     "BPe" + documentKey,
		Ide:        &schema.TAnonComplexIde2{CUF: "43", TpAmb: "2"},
		Emit:       &schema.TAnonComplexEmit2{CNPJ: "12345678000195"},
		InfViagem:  []*schema.TAnonComplexInfViagem1{{}},
	}
}

func minimalInfBPeTM() *schema.TAnonComplexInfBPe1 {
	return &schema.TAnonComplexInfBPe1{
		VersaoAttr: "1.00",
		IdAttr:     "BPeTM" + documentKey,
		Ide:        &schema.TAnonComplexIde1{CUF: "43", TpAmb: "2"},
		Emit:       &schema.TAnonComplexEmit1{CNPJ: "12345678000195"},
		DetBPeTM:   []*schema.TAnonComplexDetBPeTM1{{}},
	}
}

func minimalProtBPe() *schema.TProtBPe {
	return &schema.TProtBPe{
		VersaoAttr: "1.00",
		InfProt: &schema.TAnonComplexInfProt1{
			TpAmb:    "2",
			ChBPe:    documentKey,
			DhRecbto: "2024-01-02T03:04:05-03:00",
			CStat:    "100",
		},
	}
}

func minimalEventoInf() *schema.TAnonComplexInfEvento1 {
	return &schema.TAnonComplexInfEvento1{
		IdAttr:     "ID110111" + documentKey + "01",
		COrgao:     "43",
		TpAmb:      "2",
		CNPJ:       "12345678000195",
		ChBPe:      documentKey,
		DhEvento:   "2024-01-02T03:04:05-03:00",
		TpEvento:   "110111",
		NSeqEvento: "1",
		DetEvento: &schema.TAnonComplexDetEvento1{
			VersaoEventoAttr: "1.00",
			InnerXML:         `<evCancBPe xmlns="http://www.portalfiscal.inf.br/bpe"><descEvento>Cancelamento</descEvento><nProt>123456789012345</nProt><xJust>Justificativa de cancelamento</xJust></evCancBPe>`,
		},
	}
}

func minimalAlteracaoPoltronaEventoInf() *alteracaoPoltronaSchema.TAnonComplexInfEvento1 {
	return &alteracaoPoltronaSchema.TAnonComplexInfEvento1{
		IdAttr:     "ID110116" + documentKey + "01",
		COrgao:     "43",
		TpAmb:      "2",
		CNPJ:       "12345678000195",
		ChBPe:      documentKey,
		DhEvento:   "2024-01-02T03:04:05-03:00",
		TpEvento:   "110116",
		NSeqEvento: "1",
		DetEvento: &alteracaoPoltronaSchema.TAnonComplexDetEvento1{
			VersaoEventoAttr: "1.00",
			InnerXML:         `<evAlteracaoPoltrona xmlns="http://www.portalfiscal.inf.br/bpe"><descEvento>Alteracao Poltrona</descEvento><nProt>123456789012345</nProt><poltrona>42A</poltrona></evAlteracaoPoltrona>`,
		},
	}
}

func minimalExcessoBagagemEventoInf() *excessoBagagemSchema.TAnonComplexInfEvento1 {
	return &excessoBagagemSchema.TAnonComplexInfEvento1{
		IdAttr:     "ID110117" + documentKey + "01",
		COrgao:     "43",
		TpAmb:      "2",
		CNPJ:       "12345678000195",
		ChBPe:      documentKey,
		DhEvento:   "2024-01-02T03:04:05-03:00",
		TpEvento:   "110117",
		NSeqEvento: "1",
		DetEvento: &excessoBagagemSchema.TAnonComplexDetEvento1{
			VersaoEventoAttr: "1.00",
			InnerXML:         `<evExcessoBagagem xmlns="http://www.portalfiscal.inf.br/bpe"><descEvento>Excesso Bagagem</descEvento><nProt>123456789012345</nProt><qBagagem>2</qBagagem><vTotBag>120.00</vTotBag></evExcessoBagagem>`,
		},
	}
}

func minimalNaoEmbEventoInf() *naoEmbSchema.TAnonComplexInfEvento1 {
	return &naoEmbSchema.TAnonComplexInfEvento1{
		IdAttr:     "ID110115" + documentKey + "01",
		COrgao:     "43",
		TpAmb:      "2",
		CNPJ:       "12345678000195",
		ChBPe:      documentKey,
		DhEvento:   "2024-01-02T03:04:05-03:00",
		TpEvento:   "110115",
		NSeqEvento: "1",
		DetEvento: &naoEmbSchema.TAnonComplexDetEvento1{
			VersaoEventoAttr: "1.00",
			InnerXML:         `<evNaoEmbBPe xmlns="http://www.portalfiscal.inf.br/bpe"><descEvento>Nao Embarque</descEvento><nProt>123456789012345</nProt><xJust>Passageiro ausente</xJust></evNaoEmbBPe>`,
		},
	}
}

func minimalBPERetEventXML(tpEvento string) string {
	return fmt.Sprintf(`<retEventoBPe xmlns="%s" versao="1.00"><infEvento><tpAmb>2</tpAmb><cStat>135</cStat><tpEvento>%s</tpEvento></infEvento></retEventoBPe>`, bpeNamespace, tpEvento)
}

func minimalBPEProcEventXML(tpEvento string) string {
	return fmt.Sprintf(`<procEventoBPe xmlns="%s" versao="1.00"><eventoBPe versao="1.00"><infEvento Id="ID%s%s01"><cOrgao>43</cOrgao><tpAmb>2</tpAmb><CNPJ>12345678000195</CNPJ><chBPe>%s</chBPe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>%s</tpEvento><nSeqEvento>1</nSeqEvento><detEvento></detEvento></infEvento></eventoBPe><retEventoBPe versao="1.00"><infEvento><tpAmb>2</tpAmb><cStat>135</cStat><tpEvento>%s</tpEvento></infEvento></retEventoBPe></procEventoBPe>`, bpeNamespace, tpEvento, documentKey, documentKey, tpEvento, tpEvento)
}

func assertBPEEventReturnInfo(t *testing.T, doc *bpe.Document, expectedEventType string) {
	t.Helper()

	require.Equal(t, expectedEventType, doc.GetEventType())
	require.Empty(t, doc.GetEventSequence())
	require.Equal(t, "135", doc.GetStatusCode())
}

func assertBPEProcessedEventInfo(t *testing.T, doc *bpe.Document, expectedEventType string) {
	t.Helper()

	require.Equal(t, expectedEventType, doc.GetEventType())
	require.Equal(t, "1", doc.GetEventSequence())
	require.Equal(t, documentKey, doc.GetAccessKey())
	require.Equal(t, "135", doc.GetStatusCode())
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
	case "", bpeNamespace:
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

func decodeInnerXML[T any](t *testing.T, fragment string) T {
	t.Helper()

	var value T
	err := xml.Unmarshal([]byte("<root>"+fragment+"</root>"), &value)
	require.NoError(t, err)
	return value
}

const documentKey = "43190812345678000195630010000000011000000011"
