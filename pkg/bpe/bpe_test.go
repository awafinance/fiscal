package bpe_test

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"io"
	"slices"
	"strings"
	"testing"

	schema "github.com/awa/nota-fiscal/internal/bpe/gen/v1_0/core"
	"github.com/awa/nota-fiscal/pkg/bpe"
	"github.com/stretchr/testify/require"
)

const (
	bpeNamespace = "http://www.portalfiscal.inf.br/bpe"
	dsNamespace  = "http://www.w3.org/2000/09/xmldsig#"
)

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
				require.NotNil(t, doc.EventoBPe)
				require.Contains(t, doc.EventoBPe.InfEvento.DetEvento.InnerXML, "<evCancBPe")
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
				require.NotNil(t, doc.ProcEventoBPe)
				require.NotNil(t, doc.ProcEventoBPe.EventoBPe)
				require.NotNil(t, doc.ProcEventoBPe.RetEventoBPe)
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
	}{
		{name: "empty", data: nil, errContains: "empty xml document"},
		{name: "unsupported root", data: []byte(`<foo></foo>`), errContains: `unsupported root element "foo"`},
		{name: "invalid bpe", data: []byte(`<BPe xmlns="http://www.portalfiscal.inf.br/bpe"></BPe>`), errContains: "missing infBPe"},
		{name: "invalid event", data: []byte(`<eventoBPe xmlns="http://www.portalfiscal.inf.br/bpe" versao="1.00"></eventoBPe>`), errContains: "missing infEvento"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := bpe.Parse(tt.data)
			require.Error(t, err)
			require.Nil(t, doc)
			require.ErrorContains(t, err, tt.errContains)
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

const documentKey = "43190812345678000195630010000000011000000011"
