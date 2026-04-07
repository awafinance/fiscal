package nfe_test

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

	schema "github.com/awa/nota-fiscal/internal/nfe/gen/v4_0/nfe_proc"
	"github.com/awa/nota-fiscal/pkg/nfe"
	"github.com/stretchr/testify/require"
)

const (
	nfeNamespace = "http://www.portalfiscal.inf.br/nfe"
	dsNamespace  = "http://www.w3.org/2000/09/xmldsig#"
)

func TestParse_Fixtures(t *testing.T) {
	t.Parallel()

	for _, fixture := range allFixtureNames(t) {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			data := readFixture(t, fixture)
			doc := parseFixture(t, fixture)

			assertDocumentContract(t, data, doc)
			assertRichFixtureShape(t, doc)
		})
	}
}

func TestParse_SpecialFixtures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		assert  func(t *testing.T, doc *nfe.Document)
	}{
		{
			name:    "processed nfe preserves protocol data",
			fixture: "42220575277525000178550030000292481295366801-procNFe.xml",
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()

				require.NotNil(t, doc.ProtNFe)
				require.NotNil(t, doc.ProtNFe.InfProt)
				require.Equal(t, "NFe42220575277525000178550030000292481295366801", doc.NFe.InfNFe.IdAttr)
				require.Equal(t, "29248", doc.NFe.InfNFe.Ide.NNF)
				require.Equal(t, "75277525000178", requirePtr(t, doc.NFe.InfNFe.Emit.CNPJ))
				require.Equal(t, "42220575277525000178550030000292481295366801", doc.ProtNFe.InfProt.ChNFe)
			},
		},
		{
			name:    "bare nfe normalizes into document shape",
			fixture: "NFe35200159594315000157550010000000012062777161.xml",
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()

				require.Nil(t, doc.ProtNFe)
				require.Equal(t, "35200159594315000157550010000000012062777161", doc.NFe.InfNFe.IdAttr)
				require.Equal(t, "1", doc.NFe.InfNFe.Ide.NNF)
				require.Equal(t, "59594315000157", requirePtr(t, doc.NFe.InfNFe.Emit.CNPJ))
				require.Len(t, doc.NFe.InfNFe.Det, 1)
			},
		},
		{
			name:    "reforma tributaria fields are available",
			fixture: "nfe_reforma_tributaria.xml",
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()

				require.NotNil(t, doc.NFe.InfNFe.Total)
				require.Equal(t, "NFe15231207933914000154550010006643501427015088", doc.NFe.InfNFe.IdAttr)
				require.NotNil(t, doc.NFe.InfNFe.Ide.GCompraGov)
				require.Equal(t, "1", doc.NFe.InfNFe.Ide.GCompraGov.TpEnteGov)
				require.NotNil(t, doc.NFe.InfNFe.Det[0].Imposto)
				require.NotNil(t, doc.NFe.InfNFe.Det[0].Imposto.IBSCBS)
				require.NotNil(t, doc.NFe.InfNFe.Total.IBSCBSTot)
				require.Equal(t, "2000.50", doc.NFe.InfNFe.Total.IBSCBSTot.VBCIBSCBS)
			},
		},
		{
			name:    "consumer invoice keeps cpf recipient totals and payment",
			fixture: "35180834128745000152550010000476121675985748-nfe.xml",
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()

				require.Equal(t, "68834846982", requirePtr(t, doc.NFe.InfNFe.Dest.CPF))
				require.NotNil(t, doc.NFe.InfNFe.Dest.XNome)
				require.Equal(t, "9.06", doc.NFe.InfNFe.Total.ICMSTot.VProd)
				require.Equal(t, "9.06", doc.NFe.InfNFe.Total.ICMSTot.VNF)
				require.Len(t, doc.NFe.InfNFe.Pag.DetPag, 1)
				require.Equal(t, "9.06", doc.NFe.InfNFe.Pag.DetPag[0].VPag)
				require.NotNil(t, doc.NFe.InfNFe.Transp.Transporta)
				require.Equal(t, "25663791000160", requirePtr(t, doc.NFe.InfNFe.Transp.Transporta.CNPJ))
				require.Equal(t, "9.06", doc.NFe.InfNFe.Det[0].Prod.VProd)
			},
		},
		{
			name:    "domestic multi item invoice keeps totals freight and payment",
			fixture: "26180875335849000115550010000016871192213331-nfe.xml",
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()

				require.Equal(t, "37148260000119", requirePtr(t, doc.NFe.InfNFe.Dest.CNPJ))
				require.Equal(t, "5780.00", doc.NFe.InfNFe.Total.ICMSTot.VProd)
				require.Equal(t, "5780.00", doc.NFe.InfNFe.Total.ICMSTot.VNF)
				require.Len(t, doc.NFe.InfNFe.Det, 3)
				require.Equal(t, "2490.00", doc.NFe.InfNFe.Det[0].Prod.VProd)
				require.Equal(t, "2490.00", doc.NFe.InfNFe.Det[1].Prod.VProd)
				require.Equal(t, "800.00", doc.NFe.InfNFe.Det[2].Prod.VProd)
				require.Len(t, doc.NFe.InfNFe.Pag.DetPag, 1)
				require.Equal(t, "5780.00", doc.NFe.InfNFe.Pag.DetPag[0].VPag)
				require.NotNil(t, doc.NFe.InfNFe.Transp.Transporta)
				require.Equal(t, "02012862002707", requirePtr(t, doc.NFe.InfNFe.Transp.Transporta.CNPJ))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := parseFixture(t, tt.fixture)
			tt.assert(t, doc)
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
		{name: "nil", data: nil, errContains: "empty xml document"},
		{name: "empty slice", data: []byte{}, errContains: "empty xml document"},
		{name: "whitespace only", data: []byte(" \n\t "), errContains: "empty xml document"},
		{name: "malformed xml", data: []byte(`<nfeProc>`), errContains: "decode nfeProc"},
		{name: "malformed xml with unclosed nested tag", data: []byte(`<nfeProc><NFe></nfeProc>`), errContains: "decode nfeProc"},
		{name: "unsupported root", data: []byte(`<not-nfe></not-nfe>`), errContains: `unsupported root element "not-nfe"`},
		{name: "unsupported root with malformed xml", data: []byte(`<not-nfe><oops>`), errContains: "read root"},
		{name: "nfeProc missing NFe", data: []byte(`<nfeProc xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"></nfeProc>`), errContains: "missing NFe"},
		{name: "nfeProc missing infNFe", data: []byte(`<nfeProc xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><NFe></NFe></nfeProc>`), errContains: "missing infNFe"},
		{name: "bare nfe missing infNFe", data: []byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe"></NFe>`), errContains: "missing infNFe"},
		{name: "missing emit", data: []byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe versao="4.00" Id="x"><ide><cUF>35</cUF><cNF>1</cNF><natOp>Venda</natOp><mod>55</mod><serie>1</serie><nNF>1</nNF><dhEmi>2020-01-01T12:00:00-03:00</dhEmi><tpNF>1</tpNF><idDest>1</idDest><cMunFG>3550308</cMunFG><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>1</cDV><tpAmb>2</tpAmb><finNFe>1</finNFe><indFinal>1</indFinal><indPres>0</indPres><procEmi>0</procEmi><verProc>test</verProc></ide><det nItem="1"><prod><cProd>1</cProd><cEAN>SEM GTIN</cEAN><xProd>Produto</xProd><NCM>00000000</NCM><CFOP>5102</CFOP><uCom>UN</uCom><qCom>1.0000</qCom><vUnCom>1.0000000000</vUnCom><vProd>1.00</vProd><cEANTrib>SEM GTIN</cEANTrib><uTrib>UN</uTrib><qTrib>1.0000</qTrib><vUnTrib>1.0000000000</vUnTrib><indTot>1</indTot></prod></det></infNFe></NFe>`), errContains: "missing emit"},
		{name: "missing issuer document", data: []byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe versao="4.00" Id="x"><ide><cUF>35</cUF><cNF>1</cNF><natOp>Venda</natOp><mod>55</mod><serie>1</serie><nNF>1</nNF><dhEmi>2020-01-01T12:00:00-03:00</dhEmi><tpNF>1</tpNF><idDest>1</idDest><cMunFG>3550308</cMunFG><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>1</cDV><tpAmb>2</tpAmb><finNFe>1</finNFe><indFinal>1</indFinal><indPres>0</indPres><procEmi>0</procEmi><verProc>test</verProc></ide><emit><xNome>Emitente</xNome><enderEmit><xLgr>Rua A</xLgr><nro>1</nro><xBairro>Centro</xBairro><cMun>3550308</cMun><xMun>Sao Paulo</xMun><UF>SP</UF><CEP>01001000</CEP></enderEmit><IE>123</IE><CRT>1</CRT></emit><det nItem="1"><prod><cProd>1</cProd><cEAN>SEM GTIN</cEAN><xProd>Produto</xProd><NCM>00000000</NCM><CFOP>5102</CFOP><uCom>UN</uCom><qCom>1.0000</qCom><vUnCom>1.0000000000</vUnCom><vProd>1.00</vProd><cEANTrib>SEM GTIN</cEANTrib><uTrib>UN</uTrib><qTrib>1.0000</qTrib><vUnTrib>1.0000000000</vUnTrib><indTot>1</indTot></prod></det></infNFe></NFe>`), errContains: "missing emit document"},
		{name: "missing det", data: []byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe versao="4.00" Id="x"><ide><cUF>35</cUF><cNF>1</cNF><natOp>Venda</natOp><mod>55</mod><serie>1</serie><nNF>1</nNF><dhEmi>2020-01-01T12:00:00-03:00</dhEmi><tpNF>1</tpNF><idDest>1</idDest><cMunFG>3550308</cMunFG><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>1</cDV><tpAmb>2</tpAmb><finNFe>1</finNFe><indFinal>1</indFinal><indPres>0</indPres><procEmi>0</procEmi><verProc>test</verProc></ide><emit><CNPJ>12345678000195</CNPJ><xNome>Emitente</xNome><enderEmit><xLgr>Rua A</xLgr><nro>1</nro><xBairro>Centro</xBairro><cMun>3550308</cMun><xMun>Sao Paulo</xMun><UF>SP</UF><CEP>01001000</CEP></enderEmit><IE>123</IE><CRT>1</CRT></emit></infNFe></NFe>`), errContains: "missing det"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := nfe.Parse(tt.data)

			require.Error(t, err)
			require.ErrorContains(t, err, tt.errContains)
			require.Nil(t, doc)
		})
	}
}

func TestRoundTrip_PreservesXML(t *testing.T) {
	t.Parallel()

	for _, fixture := range allFixtureNames(t) {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			original := readFixture(t, fixture)
			doc := parseFixture(t, fixture)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)

			require.Equal(t, normalizeXML(t, original), normalizeXML(t, roundTripped))
		})
	}
}

func TestMarshalXML_NilReceiver(t *testing.T) {
	t.Parallel()

	var doc *nfe.Document

	data, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Empty(t, data)
}

func TestMarshalXML_ManualDocumentWithoutRootName(t *testing.T) {
	t.Parallel()

	t.Run("without protocol marshals as bare nfe", func(t *testing.T) {
		t.Parallel()

		doc := &nfe.Document{VersaoAttr: "4.00", NFe: minimalNFe()}

		data, err := xml.Marshal(doc)
		require.NoError(t, err)
		require.Contains(t, string(data), `<NFe xmlns="`+nfeNamespace+`">`)
		require.NotContains(t, string(data), "<nfeProc")
	})

	t.Run("with protocol promotes to processed nfe", func(t *testing.T) {
		t.Parallel()

		doc := &nfe.Document{
			VersaoAttr: "4.00",
			NFe:        minimalNFe(),
			ProtNFe: &schema.TProtNFe{
				VersaoAttr: "4.00",
				InfProt: &schema.TAnonComplexInfProt1{
					TpAmb:    "2",
					ChNFe:    "35180834128745000152550010000476121675985748",
					DhRecbto: "2020-01-01T12:00:00-03:00",
					CStat:    "100",
				},
			},
		}

		data, err := xml.Marshal(doc)
		require.NoError(t, err)
		require.Contains(t, string(data), `<nfeProc xmlns="`+nfeNamespace+`" versao="4.00">`)
		require.Contains(t, string(data), "<protNFe")
	})
}

func allFixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "..", "testdata", "nfe"))
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

func parseFixture(t *testing.T, name string) *nfe.Document {
	t.Helper()

	data := readFixture(t, name)
	doc, err := nfe.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc)

	return doc
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", name))
	require.NoError(t, err)

	return data
}

func assertDocumentContract(t *testing.T, data []byte, doc *nfe.Document) {
	t.Helper()

	require.Equal(t, "4.00", doc.VersaoAttr)
	require.NotNil(t, doc.NFe)
	require.NotNil(t, doc.NFe.InfNFe)
	require.NotNil(t, doc.NFe.InfNFe.Ide)
	require.NotNil(t, doc.NFe.InfNFe.Emit)
	require.NotEmpty(t, doc.NFe.InfNFe.IdAttr)
	require.NotEmpty(t, doc.NFe.InfNFe.Ide.NNF)
	require.NotEmpty(t, doc.NFe.InfNFe.Det)
	require.NotEmpty(t, issuerDocument(t, doc))

	for i, det := range doc.NFe.InfNFe.Det {
		require.NotNilf(t, det.Prod, "det[%d] missing prod", i)
		require.NotEmptyf(t, det.Prod.QCom, "det[%d] missing qCom", i)
		require.NotEmptyf(t, det.Prod.VUnCom, "det[%d] missing vUnCom", i)
		require.NotEmptyf(t, det.Prod.VProd, "det[%d] missing vProd", i)
	}

	if fixtureHasProtNFe(data) {
		require.NotNil(t, doc.ProtNFe)
	} else {
		require.Nil(t, doc.ProtNFe)
	}
}

func assertRichFixtureShape(t *testing.T, doc *nfe.Document) {
	t.Helper()

	require.NotNil(t, doc.NFe.InfNFe.Total)
	require.NotNil(t, doc.NFe.InfNFe.Total.ICMSTot)
	require.NotNil(t, doc.NFe.InfNFe.Transp)
	require.NotNil(t, doc.NFe.InfNFe.Pag)
	require.NotEmpty(t, doc.NFe.InfNFe.Total.ICMSTot.VProd)
	require.NotEmpty(t, doc.NFe.InfNFe.Total.ICMSTot.VNF)
	require.NotEmpty(t, doc.NFe.InfNFe.Pag.DetPag)

	for i, detPag := range doc.NFe.InfNFe.Pag.DetPag {
		require.NotEmptyf(t, detPag.VPag, "detPag[%d] missing vPag", i)
	}
}

func fixtureHasProtNFe(data []byte) bool {
	return bytes.Contains(data, []byte("<protNFe"))
}

func issuerDocument(t *testing.T, doc *nfe.Document) string {
	t.Helper()

	emit := doc.NFe.InfNFe.Emit
	if emit.CNPJ != nil {
		return strings.TrimSpace(*emit.CNPJ)
	}
	if emit.CPF != nil {
		return strings.TrimSpace(*emit.CPF)
	}

	return ""
}

func normalizeXML(t *testing.T, data []byte) string {
	t.Helper()

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
	case "", nfeNamespace:
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

func minimalNFe() *schema.TNFe {
	return &schema.TNFe{
		InfNFe: &schema.TAnonComplexInfNFe1{
			VersaoAttr: "4.00",
			IdAttr:     "NFe35180834128745000152550010000476121675985748",
			Ide: &schema.Ide{
				CUF:      "35",
				CNF:      "12345678",
				NatOp:    "Venda",
				Mod:      "55",
				Serie:    "1",
				NNF:      "1",
				DhEmi:    "2020-01-01T12:00:00-03:00",
				TpNF:     "1",
				IdDest:   "1",
				CMunFG:   "3550308",
				TpImp:    "1",
				TpEmis:   "1",
				CDV:      "1",
				TpAmb:    "2",
				FinNFe:   "1",
				IndFinal: "1",
				IndPres:  "0",
				ProcEmi:  "0",
				VerProc:  "test",
			},
			Emit: &schema.Emit{
				CNPJ:      ptr("12345678000195"),
				XNome:     "Emitente",
				EnderEmit: &schema.TEnderEmi{XLgr: "Rua A", Nro: "1", CMun: "3550308", XMun: "Sao Paulo", UF: "SP", CEP: "01001000"},
				IE:        "123",
				CRT:       "1",
			},
			Det: []*schema.Det{
				{
					NItemAttr: "1",
					Prod: &schema.Prod{
						CProd:    "1",
						CEAN:     "SEM GTIN",
						XProd:    "Produto",
						NCM:      "00000000",
						CFOP:     "5102",
						UCom:     "UN",
						QCom:     "1.0000",
						VUnCom:   "1.0000000000",
						VProd:    "1.00",
						CEANTrib: "SEM GTIN",
						UTrib:    "UN",
						QTrib:    "1.0000",
						VUnTrib:  "1.0000000000",
						IndTot:   "1",
					},
				},
			},
		},
	}
}

func ptr[T any](v T) *T {
	return &v
}
