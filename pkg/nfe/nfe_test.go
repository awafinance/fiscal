package nfe_test

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	atorSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/ator_interessado"
	distSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/dist_dfe"
	insucessoCancelSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel_insucesso"
	genericSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_generico"
	insucessoSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_insucesso"
	mdeSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_mde"
	consSchema "github.com/awafinance/fiscal/internal/nfe/gen/v2_0/cons"
	schema "github.com/awafinance/fiscal/internal/nfe/gen/v4_0/nfe_proc"
	"github.com/awafinance/fiscal/pkg/nfe"
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

func TestParse_SignedFixture(t *testing.T) {
	t.Parallel()

	doc := parseFixture(t, "35180834128745000152550010000476121675985748-nfe.xml")

	require.NotNil(t, doc.NFe)
	require.NotNil(t, doc.NFe.DsSignature)
	require.NotNil(t, doc.NFe.DsSignature.SignedInfo)
	require.Equal(t, "http://www.w3.org/TR/2001/REC-xml-c14n-20010315", doc.NFe.DsSignature.SignedInfo.CanonicalizationMethod.AlgorithmAttr)
	require.Equal(t, "http://www.w3.org/2000/09/xmldsig#rsa-sha1", doc.NFe.DsSignature.SignedInfo.SignatureMethod.AlgorithmAttr)
	require.Equal(t, "#NFe35180834128745000152550010000476121675985748", doc.NFe.DsSignature.SignedInfo.Reference.URIAttr)
	require.Equal(t, "http://www.w3.org/2000/09/xmldsig#sha1", doc.NFe.DsSignature.SignedInfo.Reference.DigestMethod.AlgorithmAttr)
	require.NotEmpty(t, doc.NFe.DsSignature.SignedInfo.Reference.DigestValue)
	require.NotNil(t, doc.NFe.DsSignature.SignatureValue)
	require.NotEmpty(t, doc.NFe.DsSignature.SignatureValue.Value)
	require.NotNil(t, doc.NFe.DsSignature.KeyInfo)
	require.NotNil(t, doc.NFe.DsSignature.KeyInfo.X509Data)
	require.NotEmpty(t, doc.NFe.DsSignature.KeyInfo.X509Data.X509Certificate)
}

func TestParse_EventFixtures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fixture string
		assert  func(t *testing.T, doc *nfe.Document)
	}{
		{
			fixture: filepath.Join("..", "..", "testdata", "nfe_evento_cce", "v1_0", "35180803102452000172550010000476051695511860-01-cce.xml"),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoCCe)
				require.Equal(t, "110110", doc.EventoCCe.InfEvento.TpEvento)
				require.Equal(t, "VOLUME: 4 VOLUMES", doc.EventoCCe.InfEvento.DetEvento.XCorrecao)
			},
		},
		{
			fixture: filepath.Join("..", "..", "testdata", "nfe_evento_cancel", "v1_0", "35180803102452000172550010000476051695511860-cancel.xml"),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoCancel)
				require.Equal(t, "110111", doc.EventoCancel.InfEvento.TpEvento)
				require.Equal(t, "135180000000001", doc.EventoCancel.InfEvento.DetEvento.NProt)
			},
		},
		{
			fixture: filepath.Join("..", "..", "testdata", "nfe_epec", "v1_0", "35180803102452000172550010000476051695511860-epec.xml"),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoEPEC)
				require.Equal(t, "110140", doc.EventoEPEC.InfEvento.TpEvento)
				require.Equal(t, "EPEC", doc.EventoEPEC.InfEvento.DetEvento.DescEvento)
				require.Equal(t, "100.00", doc.EventoEPEC.InfEvento.DetEvento.Dest.VNF)
			},
		},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.fixture), func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(tt.fixture)
			require.NoError(t, err)

			doc, err := nfe.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
		})
	}
}

func TestParse_InutilizacaoFixture(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "testdata", "nfe_inutilizacao", "v4_0", "41080676472349000430550010000001041671821888-ped-inu.xml")

	data, err := os.ReadFile(fixture)
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotNil(t, doc.InutNFe)
	require.Equal(t, "76472349000430", doc.InutNFe.InfInut.CNPJ)

	roundTripped, err := xml.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)
	require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
}

func TestMarshalXML_MinimizesNamespaceDeclarations(t *testing.T) {
	t.Parallel()

	doc := parseFixture(t, "35180834128745000152550010000476121675985748-nfe.xml")

	roundTripped, err := xml.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)

	require.Equal(t, 1, strings.Count(string(roundTripped), `xmlns="http://www.portalfiscal.inf.br/nfe"`))
	require.Equal(t, 1, strings.Count(string(roundTripped), `xmlns:ds="http://www.w3.org/2000/09/xmldsig#"`))
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

func TestParse_NewEventRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  any
		assert func(t *testing.T, doc *nfe.Document)
	}{
		{
			name: "ator interessado",
			value: struct {
				XMLName xml.Name `xml:"evento"`
				XMLNS   string   `xml:"xmlns,attr"`
				*atorSchema.TEvento
			}{
				XMLName: xml.Name{Local: "evento"},
				XMLNS:   nfeNamespace,
				TEvento: &atorSchema.TEvento{
					VersaoAttr: "1.00",
					InfEvento: &atorSchema.TAnonComplexInfEvento1{
						IdAttr:     "ID1101503518080310245200017255001000047605169551186001",
						COrgao:     "91",
						TpAmb:      "1",
						CNPJ:       stringPtr("12345678000195"),
						ChNFe:      "35180803102452000172550010000476051695511860",
						DhEvento:   "2024-01-02T03:04:05-03:00",
						TpEvento:   "110150",
						NSeqEvento: "1",
						VerEvento:  "1.00",
						DetEvento: &atorSchema.TAnonComplexDetEvento1{
							VersaoAttr:  "1.00",
							DescEvento:  "Ator interessado na NF-e",
							COrgaoAutor: "91",
							TpAutor:     "1",
							AutXML:      &atorSchema.TAnonComplexAutXML1{CNPJ: stringPtr("12345678000195")},
						},
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoAtorInteressado)
				require.Equal(t, "110150", doc.EventoAtorInteressado.InfEvento.TpEvento)
			},
		},
		{
			name: "mde",
			value: struct {
				XMLName xml.Name `xml:"evento"`
				XMLNS   string   `xml:"xmlns,attr"`
				*mdeSchema.TEvento
			}{
				XMLName: xml.Name{Local: "evento"},
				XMLNS:   nfeNamespace,
				TEvento: &mdeSchema.TEvento{
					VersaoAttr: "1.00",
					InfEvento: &mdeSchema.TAnonComplexInfEvento1{
						IdAttr:     "ID2102403518080310245200017255001000047605169551186001",
						COrgao:     "91",
						TpAmb:      "1",
						CNPJ:       stringPtr("12345678000195"),
						ChNFe:      "35180803102452000172550010000476051695511860",
						DhEvento:   "2024-01-02T03:04:05-03:00",
						TpEvento:   "210240",
						NSeqEvento: "1",
						VerEvento:  "1.00",
						DetEvento: &mdeSchema.TAnonComplexDetEvento1{
							VersaoAttr: "1.00",
							DescEvento: "Operacao nao Realizada",
							XJust:      stringPtr("Mercadoria recusada"),
						},
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoMDE)
				require.Equal(t, "210240", doc.EventoMDE.InfEvento.TpEvento)
				require.Equal(t, "Mercadoria recusada", requirePtr(t, doc.EventoMDE.InfEvento.DetEvento.XJust))
			},
		},
		{
			name: "insucesso",
			value: struct {
				XMLName xml.Name `xml:"evento"`
				XMLNS   string   `xml:"xmlns,attr"`
				*insucessoSchema.TEvento
			}{
				XMLName: xml.Name{Local: "evento"},
				XMLNS:   nfeNamespace,
				TEvento: &insucessoSchema.TEvento{
					VersaoAttr: "1.00",
					InfEvento: &insucessoSchema.TAnonComplexInfEvento1{
						IdAttr:     "ID1101923518080310245200017255001000047605169551186001",
						COrgao:     "91",
						TpAmb:      "1",
						CNPJ:       stringPtr("12345678000195"),
						ChNFe:      "35180803102452000172550010000476051695511860",
						DhEvento:   "2024-01-02T03:04:05-03:00",
						TpEvento:   "110192",
						NSeqEvento: "1",
						VerEvento:  "1.00",
						DetEvento: &insucessoSchema.TAnonComplexDetEvento1{
							VersaoAttr:           "1.00",
							DescEvento:           "Insucesso na Entrega da NF-e",
							COrgaoAutor:          "91",
							DhTentativaEntrega:   "2024-01-02T03:04:05-03:00",
							NTentativa:           stringPtr("1"),
							TpMotivo:             "4",
							XJustMotivo:          stringPtr("Endereco fechado"),
							HashTentativaEntrega: "ABCDEF0123456789",
						},
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoInsucesso)
				require.Equal(t, "110192", doc.EventoInsucesso.InfEvento.TpEvento)
				require.Equal(t, "4", doc.EventoInsucesso.InfEvento.DetEvento.TpMotivo)
			},
		},
		{
			name: "cancel insucesso",
			value: struct {
				XMLName xml.Name `xml:"evento"`
				XMLNS   string   `xml:"xmlns,attr"`
				*insucessoCancelSchema.TEvento
			}{
				XMLName: xml.Name{Local: "evento"},
				XMLNS:   nfeNamespace,
				TEvento: &insucessoCancelSchema.TEvento{
					VersaoAttr: "1.00",
					InfEvento: &insucessoCancelSchema.TAnonComplexInfEvento1{
						IdAttr:     "ID1101933518080310245200017255001000047605169551186001",
						COrgao:     "91",
						TpAmb:      "1",
						CNPJ:       stringPtr("12345678000195"),
						ChNFe:      "35180803102452000172550010000476051695511860",
						DhEvento:   "2024-01-02T03:04:05-03:00",
						TpEvento:   "110193",
						NSeqEvento: "1",
						VerEvento:  "1.00",
						DetEvento: &insucessoCancelSchema.TAnonComplexDetEvento1{
							VersaoAttr:  "1.00",
							DescEvento:  "Cancelamento do Insucesso na Entrega da NF-e",
							COrgaoAutor: "91",
							NProtEvento: "135240000000001",
						},
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoCancInsucesso)
				require.Equal(t, "110193", doc.EventoCancInsucesso.InfEvento.TpEvento)
				require.Equal(t, "135240000000001", doc.EventoCancInsucesso.InfEvento.DetEvento.NProtEvento)
			},
		},
		{
			name: "generico fallback",
			value: struct {
				XMLName xml.Name `xml:"evento"`
				XMLNS   string   `xml:"xmlns,attr"`
				*genericSchema.TEvento
			}{
				XMLName: xml.Name{Local: "evento"},
				XMLNS:   nfeNamespace,
				TEvento: &genericSchema.TEvento{
					VersaoAttr: "1.00",
					InfEvento: &genericSchema.TAnonComplexInfEvento1{
						IdAttr:     "ID9999993518080310245200017255001000047605169551186001",
						COrgao:     "91",
						TpAmb:      "1",
						CNPJ:       stringPtr("12345678000195"),
						ChNFe:      "35180803102452000172550010000476051695511860",
						DhEvento:   "2024-01-02T03:04:05-03:00",
						TpEvento:   "999999",
						NSeqEvento: "1",
						VerEvento:  "1.00",
						DetEvento:  &genericSchema.TAnonComplexDetEvento1{},
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoGenerico)
				require.Equal(t, "999999", doc.EventoGenerico.InfEvento.TpEvento)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := xml.MarshalIndent(tt.value, "", "  ")
			require.NoError(t, err)

			doc, err := nfe.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
		})
	}
}

func TestParse_ConsRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  any
		assert func(t *testing.T, doc *nfe.Document)
	}{
		{
			name: "consSitNFe",
			value: struct {
				XMLName xml.Name `xml:"consSitNFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*consSchema.TConsSitNFe
			}{
				XMLName: xml.Name{Local: "consSitNFe"},
				XMLNS:   nfeNamespace,
				TConsSitNFe: &consSchema.TConsSitNFe{
					VersaoAttr: "2.01",
					TpAmb:      "1",
					XServ:      "CONSULTAR",
					ChNFe:      "35180803102452000172550010000476051695511860",
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ConsSitNFe)
				require.Equal(t, "CONSULTAR", doc.ConsSitNFe.XServ)
			},
		},
		{
			name: "retConsSitNFe",
			value: struct {
				XMLName    xml.Name             `xml:"retConsSitNFe"`
				XMLNS      string               `xml:"xmlns,attr"`
				VersaoAttr string               `xml:"versao,attr"`
				TpAmb      string               `xml:"tpAmb"`
				VerAplic   *consSchema.TString  `xml:"verAplic,omitempty"`
				CStat      string               `xml:"cStat"`
				XMotivo    *consSchema.TString  `xml:"xMotivo,omitempty"`
				CUF        string               `xml:"cUF"`
				ChNFe      string               `xml:"chNFe"`
				ProtNFe    *consSchema.TProtNFe `xml:"protNFe,omitempty"`
			}{
				XMLName:    xml.Name{Local: "retConsSitNFe"},
				XMLNS:      nfeNamespace,
				VersaoAttr: "2.01",
				TpAmb:      "1",
				VerAplic:   nfeConsTStringPtr("SVRS202401"),
				CStat:      "100",
				XMotivo:    nfeConsTStringPtr("Autorizado o uso da NF-e"),
				CUF:        "35",
				ChNFe:      "35180803102452000172550010000476051695511860",
				ProtNFe: &consSchema.TProtNFe{
					VersaoAttr: "2.01",
					InfProt: &consSchema.TAnonComplexInfProt1{
						TpAmb:    "1",
						ChNFe:    "35180803102452000172550010000476051695511860",
						DhRecbto: "2024-01-02T03:04:05-03:00",
						CStat:    "100",
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetConsSitNFe)
				require.Equal(t, "100", doc.RetConsSitNFe.CStat)
				require.NotNil(t, doc.RetConsSitNFe.ProtNFe)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := xml.MarshalIndent(tt.value, "", "  ")
			require.NoError(t, err)

			doc, err := nfe.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
		})
	}
}

func TestParse_DistDFeRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  any
		assert func(t *testing.T, doc *nfe.Document)
	}{
		{
			name: "distDFeInt",
			value: struct {
				XMLName xml.Name `xml:"distDFeInt"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distSchema.TAnonComplexDistDFeInt1
			}{
				XMLName: xml.Name{Local: "distDFeInt"},
				XMLNS:   nfeNamespace,
				TAnonComplexDistDFeInt1: &distSchema.TAnonComplexDistDFeInt1{
					VersaoAttr: "1.01",
					TpAmb:      "1",
					CNPJ:       stringPtr("12345678000195"),
					DistNSU:    &distSchema.TAnonComplexDistNSU1{UltNSU: "000000000000001"},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
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
				XMLNS:   nfeNamespace,
				TAnonComplexRetDistDFeInt1: &distSchema.TAnonComplexRetDistDFeInt1{
					VersaoAttr: "1.01",
					TpAmb:      "1",
					VerAplic:   nfeTStringPtr("test"),
					CStat:      "138",
					XMotivo:    nfeTStringPtr("Documento localizado"),
					DhResp:     "2024-01-02T03:04:05-03:00",
					UltNSU:     "000000000000010",
					MaxNSU:     "000000000000099",
					LoteDistDFeInt: &distSchema.TAnonComplexLoteDistDFeInt1{
						DocZip: []*distSchema.TAnonComplexDocZip1{{NSUAttr: "000000000000010", SchemaAttr: "resNFe_v1.01.xsd", Value: "ZGF0YQ=="}},
					},
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetDistDFeInt)
				require.Len(t, doc.RetDistDFeInt.LoteDistDFeInt.DocZip, 1)
			},
		},
		{
			name: "resNFe",
			value: struct {
				XMLName xml.Name `xml:"resNFe"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distSchema.TAnonComplexResNFe1
			}{
				XMLName: xml.Name{Local: "resNFe"},
				XMLNS:   nfeNamespace,
				TAnonComplexResNFe1: &distSchema.TAnonComplexResNFe1{
					VersaoAttr: "1.01",
					ChNFe:      "35180803102452000172550010000476051695511860",
					CNPJ:       stringPtr("12345678000195"),
					XNome:      "Emitente Teste",
					IE:         "ISENTO",
					DhEmi:      "2024-01-02T03:04:05-03:00",
					TpNF:       "1",
					VNF:        "100.00",
					DhRecbto:   "2024-01-02T03:05:06-03:00",
					NProt:      "123456789012345",
					CSitNFe:    "1",
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ResNFe)
				require.Equal(t, "Emitente Teste", doc.ResNFe.XNome)
			},
		},
		{
			name: "resEvento",
			value: struct {
				XMLName xml.Name `xml:"resEvento"`
				XMLNS   string   `xml:"xmlns,attr"`
				*distSchema.TAnonComplexResEvento1
			}{
				XMLName: xml.Name{Local: "resEvento"},
				XMLNS:   nfeNamespace,
				TAnonComplexResEvento1: &distSchema.TAnonComplexResEvento1{
					VersaoAttr: "1.01",
					COrgao:     "91",
					CNPJ:       stringPtr("12345678000195"),
					ChNFe:      "35180803102452000172550010000476051695511860",
					DhEvento:   "2024-01-02T03:04:05-03:00",
					TpEvento:   "110111",
					NSeqEvento: "1",
					XEvento:    "Cancelamento",
					DhRecbto:   "2024-01-02T03:05:06-03:00",
					NProt:      "123456789012345",
				},
			},
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ResEvento)
				require.Equal(t, "110111", doc.ResEvento.TpEvento)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := xml.MarshalIndent(tt.value, "", "  ")
			require.NoError(t, err)

			doc, err := nfe.Parse(data)
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, data), normalizeXML(t, roundTripped))
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

func TestMarshalXML_ManualDocumentEdgeStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		doc    *nfe.Document
		assert func(t *testing.T, data string)
	}{
		{
			name: "empty version attribute with protocol emits empty versao on nfeProc",
			doc: &nfe.Document{
				NFe:     minimalNFe(),
				ProtNFe: &schema.TProtNFe{VersaoAttr: "4.00", InfProt: minimalInfProt()},
			},
			assert: func(t *testing.T, data string) {
				t.Helper()
				require.Contains(t, data, `<nfeProc xmlns="`+nfeNamespace+`" versao="">`)
				require.Contains(t, data, "<protNFe")
			},
		},
		{
			name: "nil NFe without protocol still marshals as bare NFe",
			doc:  &nfe.Document{VersaoAttr: "4.00"},
			assert: func(t *testing.T, data string) {
				t.Helper()
				require.Equal(t, `<NFe xmlns="`+nfeNamespace+`"></NFe>`, data)
			},
		},
		{
			name: "protocol with nil infProt still marshals as processed nfe",
			doc: &nfe.Document{
				VersaoAttr: "4.00",
				NFe:        minimalNFe(),
				ProtNFe:    &schema.TProtNFe{VersaoAttr: "4.00"},
			},
			assert: func(t *testing.T, data string) {
				t.Helper()
				require.Contains(t, data, `<nfeProc xmlns="`+nfeNamespace+`" versao="4.00">`)
				require.Contains(t, data, `<protNFe versao="4.00"></protNFe>`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			encoded, err := xml.Marshal(tt.doc)
			require.NoError(t, err)

			tt.assert(t, string(encoded))
		})
	}
}

func TestParse_RootScanningEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   []byte
		assert func(t *testing.T, doc *nfe.Document)
	}{
		{
			name: "bom prefix before root",
			data: append([]byte{0xEF, 0xBB, 0xBF}, []byte(minimalNFEXML("NFe", "", ""))...),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.Nil(t, doc.ProtNFe)
			},
		},
		{
			name: "processing instructions before root",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?><?xml-stylesheet type="text/xsl" href="nfe.xsl"?>` + minimalNFEXML("NFe", "", "")),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.Nil(t, doc.ProtNFe)
			},
		},
		{
			name: "unusual namespace prefix on root",
			data: []byte(minimalNFEXML("xsd123:NFe", ` xmlns:xsd123="`+nfeNamespace+`"`, "xsd123:")),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.Nil(t, doc.ProtNFe)
			},
		},
		{
			name: "mixed prefixed descendants under default namespace root",
			data: []byte(minimalNFEXML(
				"NFe",
				` xmlns:nf="`+nfeNamespace+`"`,
				"nf:",
			)),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.Nil(t, doc.ProtNFe)
				require.Equal(t, "Emitente", doc.NFe.InfNFe.Emit.XNome)
				require.Equal(t, "Produto", doc.NFe.InfNFe.Det[0].Prod.XProd)
			},
		},
		{
			name: "prefixed nfeProc root and children",
			data: []byte(minimalProcNFEXML(
				"proc:nfeProc",
				` xmlns:proc="`+nfeNamespace+`"`,
				"proc:",
			)),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.NotNil(t, doc.ProtNFe)
				require.NotNil(t, doc.ProtNFe.InfProt)
				require.Equal(t, "35180834128745000152550010000476121675985748", doc.ProtNFe.InfProt.ChNFe)
			},
		},
		{
			name: "prefixed nfeProc with different prefixes for sibling elements",
			data: []byte(mixedPrefixProcNFEXML()),
			assert: func(t *testing.T, doc *nfe.Document) {
				t.Helper()
				require.NotNil(t, doc.ProtNFe)
				require.Equal(t, "100", doc.ProtNFe.InfProt.CStat)
				require.Equal(t, "Emitente", doc.NFe.InfNFe.Emit.XNome)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := nfe.Parse(tt.data)
			require.NoError(t, err)
			require.NotNil(t, doc)
			require.Equal(t, "4.00", doc.VersaoAttr)
			require.NotNil(t, doc.NFe)
			require.NotNil(t, doc.NFe.InfNFe)
			require.Equal(t, "1", doc.NFe.InfNFe.Ide.NNF)
			tt.assert(t, doc)
		})
	}
}

func TestParse_UnicodeFields(t *testing.T) {
	t.Parallel()

	doc, err := nfe.Parse([]byte(unicodeNFEXML()))
	require.NoError(t, err)
	require.Equal(t, "José & Filhos Comércio de Açúcar Ltda.", doc.NFe.InfNFe.Emit.XNome)
	require.Equal(t, "Rua São João, nº 123", doc.NFe.InfNFe.Emit.EnderEmit.XLgr)
	require.Equal(t, "São Paulo", doc.NFe.InfNFe.Emit.EnderEmit.XMun)
	require.Equal(t, "Café torrado e moído 500g", doc.NFe.InfNFe.Det[0].Prod.XProd)

	encoded, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "José &amp; Filhos Comércio de Açúcar Ltda.")
	require.Contains(t, string(encoded), "Rua São João, nº 123")
	require.Contains(t, string(encoded), "Café torrado e moído 500g")
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

	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
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

func TestParse_ExpandedNFESurfaceRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		assert func(t *testing.T, doc *nfe.Document)
	}{
		{
			name: "evento entrega",
			data: `<evento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00"><infEvento Id="ID1101303518080310245200017255001000047605169551186001"><cOrgao>91</cOrgao><tpAmb>1</tpAmb><CNPJ>12345678000195</CNPJ><chNFe>35180803102452000172550010000476051695511860</chNFe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>110130</tpEvento><nSeqEvento>1</nSeqEvento><verEvento>1.00</verEvento><detEvento versao="1.00"></detEvento></infEvento></evento>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoEntrega)
				require.Equal(t, "110130", doc.EventoEntrega.InfEvento.TpEvento)
			},
		},
		{
			name: "evento cancel entrega",
			data: `<evento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00"><infEvento Id="ID1101313518080310245200017255001000047605169551186001"><cOrgao>91</cOrgao><tpAmb>1</tpAmb><CNPJ>12345678000195</CNPJ><chNFe>35180803102452000172550010000476051695511860</chNFe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>110131</tpEvento><nSeqEvento>1</nSeqEvento><verEvento>1.00</verEvento><detEvento versao="1.00"></detEvento></infEvento></evento>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EventoCancEntrega)
				require.Equal(t, "110131", doc.EventoCancEntrega.InfEvento.TpEvento)
			},
		},
		{
			name: "envEvento",
			data: `<envEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00"><idLote>1</idLote><evento versao="1.00"><infEvento Id="ID9999993518080310245200017255001000047605169551186001"><cOrgao>91</cOrgao><tpAmb>1</tpAmb><CNPJ>12345678000195</CNPJ><chNFe>35180803102452000172550010000476051695511860</chNFe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>999999</tpEvento><nSeqEvento>1</nSeqEvento><verEvento>1.00</verEvento><detEvento></detEvento></infEvento></evento></envEvento>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EnvEvento)
				require.Equal(t, "1", doc.EnvEvento.IdLote)
				require.Len(t, doc.EnvEvento.Evento, 1)
			},
		},
		{
			name: "retEnvEvento",
			data: `<retEnvEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00"><idLote>1</idLote><tpAmb>1</tpAmb><verAplic>SVRS202401</verAplic><cOrgao>91</cOrgao><cStat>128</cStat><xMotivo>Lote de Evento Processado</xMotivo><retEvento versao="1.00"><infEvento Id="ID9999993518080310245200017255001000047605169551186001"><tpAmb>1</tpAmb><verAplic>SVRS202401</verAplic><cOrgao>91</cOrgao><cStat>135</cStat><xMotivo>Evento registrado</xMotivo><chNFe>35180803102452000172550010000476051695511860</chNFe><tpEvento>999999</tpEvento><nSeqEvento>1</nSeqEvento></infEvento></retEvento></retEnvEvento>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetEnvEvento)
				require.Equal(t, "128", doc.RetEnvEvento.CStat)
			},
		},
		{
			name: "procEventoNFe",
			data: `<procEventoNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00"><evento versao="1.00"><infEvento Id="ID9999993518080310245200017255001000047605169551186001"><cOrgao>91</cOrgao><tpAmb>1</tpAmb><CNPJ>12345678000195</CNPJ><chNFe>35180803102452000172550010000476051695511860</chNFe><dhEvento>2024-01-02T03:04:05-03:00</dhEvento><tpEvento>999999</tpEvento><nSeqEvento>1</nSeqEvento><verEvento>1.00</verEvento><detEvento></detEvento></infEvento></evento><retEvento versao="1.00"><infEvento Id="ID9999993518080310245200017255001000047605169551186001"><tpAmb>1</tpAmb><verAplic>SVRS202401</verAplic><cOrgao>91</cOrgao><cStat>135</cStat><xMotivo>Evento registrado</xMotivo><chNFe>35180803102452000172550010000476051695511860</chNFe><tpEvento>999999</tpEvento><nSeqEvento>1</nSeqEvento></infEvento></retEvento></procEventoNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ProcEventoNFe)
				require.NotNil(t, doc.ProcEventoNFe.Evento)
				require.NotNil(t, doc.ProcEventoNFe.RetEvento)
			},
		},
		{
			name: "consStatServ",
			data: `<consStatServ xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>2</tpAmb><cUF>35</cUF><xServ>STATUS</xServ></consStatServ>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ConsStatServ)
				require.Equal(t, "STATUS", doc.ConsStatServ.XServ)
			},
		},
		{
			name: "retConsStatServ",
			data: `<retConsStatServ xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>2</tpAmb><verAplic>SVRS202401</verAplic><cStat>107</cStat><xMotivo>Servico em Operacao</xMotivo><cUF>35</cUF><dhRecbto>2024-01-02T03:04:05-03:00</dhRecbto></retConsStatServ>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetConsStatServ)
				require.Equal(t, "107", doc.RetConsStatServ.CStat)
			},
		},
		{
			name: "enviNFe",
			data: `<enviNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><idLote>1</idLote><indSinc>1</indSinc><NFe></NFe></enviNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.EnviNFe)
				require.Equal(t, "1", doc.EnviNFe.IdLote)
				require.Len(t, doc.EnviNFe.NFe, 1)
			},
		},
		{
			name: "retEnviNFe",
			data: `<retEnviNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>2</tpAmb><verAplic>SVRS202401</verAplic><cStat>103</cStat><xMotivo>Lote recebido com sucesso</xMotivo><cUF>35</cUF><dhRecbto>2024-01-02T03:04:05-03:00</dhRecbto></retEnviNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetEnviNFe)
				require.Equal(t, "103", doc.RetEnviNFe.CStat)
			},
		},
		{
			name: "consReciNFe",
			data: `<consReciNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>2</tpAmb><nRec>351000000000001</nRec></consReciNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ConsReciNFe)
				require.Equal(t, "351000000000001", doc.ConsReciNFe.NRec)
			},
		},
		{
			name: "retConsReciNFe",
			data: `<retConsReciNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><tpAmb>2</tpAmb><verAplic>SVRS202401</verAplic><nRec>351000000000001</nRec><cStat>104</cStat><xMotivo>Lote processado</xMotivo><cUF>35</cUF><dhRecbto>2024-01-02T03:04:05-03:00</dhRecbto></retConsReciNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetConsReciNFe)
				require.Equal(t, "104", doc.RetConsReciNFe.CStat)
			},
		},
		{
			name: "inutNFe",
			data: `<inutNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><infInut Id="ID352401123456780001955500100000010000000010"><tpAmb>2</tpAmb><xServ>INUTILIZAR</xServ><cUF>35</cUF><ano>24</ano><CNPJ>12345678000195</CNPJ><mod>55</mod><serie>1</serie><nNFIni>100</nNFIni><nNFFin>100</nNFFin><xJust>Faixa nao utilizada</xJust></infInut></inutNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.InutNFe)
				require.Equal(t, "12345678000195", doc.InutNFe.InfInut.CNPJ)
			},
		},
		{
			name: "retInutNFe",
			data: `<retInutNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><infInut Id="ID352401123456780001955500100000010000000010"><tpAmb>2</tpAmb><xServ>INUTILIZAR</xServ><cUF>35</cUF><ano>24</ano><CNPJ>12345678000195</CNPJ><mod>55</mod><serie>1</serie><nNFIni>100</nNFIni><nNFFin>100</nNFFin><xJust>Faixa inutilizada</xJust></infInut></retInutNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.RetInutNFe)
				require.Equal(t, "12345678000195", requirePtr(t, doc.RetInutNFe.InfInut.CNPJ))
			},
		},
		{
			name: "procInutNFe",
			data: `<procInutNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00"><inutNFe versao="4.00"><infInut Id="ID352401123456780001955500100000010000000010"><tpAmb>2</tpAmb><xServ>INUTILIZAR</xServ><cUF>35</cUF><ano>24</ano><CNPJ>12345678000195</CNPJ><mod>55</mod><serie>1</serie><nNFIni>100</nNFIni><nNFFin>100</nNFFin><xJust>Faixa nao utilizada</xJust></infInut></inutNFe><retInutNFe versao="4.00"><infInut Id="ID352401123456780001955500100000010000000010"><tpAmb>2</tpAmb><xServ>INUTILIZAR</xServ><cUF>35</cUF><ano>24</ano><CNPJ>12345678000195</CNPJ><mod>55</mod><serie>1</serie><nNFIni>100</nNFIni><nNFFin>100</nNFFin><xJust>Faixa inutilizada</xJust></infInut></retInutNFe></procInutNFe>`,
			assert: func(t *testing.T, doc *nfe.Document) {
				require.NotNil(t, doc.ProcInutNFe)
				require.NotNil(t, doc.ProcInutNFe.InutNFe)
				require.NotNil(t, doc.ProcInutNFe.RetInutNFe)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := nfe.Parse([]byte(tt.data))
			require.NoError(t, err)
			tt.assert(t, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)

			reparsed, err := nfe.Parse(roundTripped)
			require.NoError(t, err)
			tt.assert(t, reparsed)
		})
	}
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

func stringPtr(v string) *string {
	return &v
}

func nfeConsTStringPtr(v string) *consSchema.TString {
	value := consSchema.TString(v)
	return &value
}

func nfeTStringPtr(v string) *distSchema.TString {
	value := distSchema.TString(v)
	return &value
}

func minimalNFe() *schema.TNFe {
	return &schema.TNFe{
		InfNFe: &schema.TAnonComplexInfNFe1{
			VersaoAttr: "4.00",
			IdAttr:     "NFe35180834128745000152550010000476121675985748",
			Ide: &schema.TAnonComplexIde1{
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
			Emit: &schema.TAnonComplexEmit1{
				CNPJ:      ptr("12345678000195"),
				XNome:     "Emitente",
				EnderEmit: &schema.TEnderEmi{XLgr: "Rua A", Nro: "1", CMun: "3550308", XMun: "Sao Paulo", UF: "SP", CEP: "01001000"},
				IE:        "123",
				CRT:       "1",
			},
			Det: []*schema.TAnonComplexDet1{
				{
					NItemAttr: "1",
					Prod: &schema.TAnonComplexProd1{
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

func minimalInfProt() *schema.TAnonComplexInfProt1 {
	return &schema.TAnonComplexInfProt1{
		TpAmb:    "2",
		ChNFe:    "35180834128745000152550010000476121675985748",
		DhRecbto: "2020-01-01T12:00:00-03:00",
		CStat:    "100",
	}
}

func minimalNFEXML(rootName, rootAttrs, prefix string) string {
	return `<` + rootName + rootAttrs + ` xmlns="` + nfeNamespace + `">
  <` + prefix + `infNFe versao="4.00" Id="NFe35180834128745000152550010000476121675985748">
    <` + prefix + `ide>
      <` + prefix + `cUF>35</` + prefix + `cUF>
      <` + prefix + `cNF>12345678</` + prefix + `cNF>
      <` + prefix + `natOp>Venda</` + prefix + `natOp>
      <` + prefix + `mod>55</` + prefix + `mod>
      <` + prefix + `serie>1</` + prefix + `serie>
      <` + prefix + `nNF>1</` + prefix + `nNF>
      <` + prefix + `dhEmi>2020-01-01T12:00:00-03:00</` + prefix + `dhEmi>
      <` + prefix + `tpNF>1</` + prefix + `tpNF>
      <` + prefix + `idDest>1</` + prefix + `idDest>
      <` + prefix + `cMunFG>3550308</` + prefix + `cMunFG>
      <` + prefix + `tpImp>1</` + prefix + `tpImp>
      <` + prefix + `tpEmis>1</` + prefix + `tpEmis>
      <` + prefix + `cDV>1</` + prefix + `cDV>
      <` + prefix + `tpAmb>2</` + prefix + `tpAmb>
      <` + prefix + `finNFe>1</` + prefix + `finNFe>
      <` + prefix + `indFinal>1</` + prefix + `indFinal>
      <` + prefix + `indPres>0</` + prefix + `indPres>
      <` + prefix + `procEmi>0</` + prefix + `procEmi>
      <` + prefix + `verProc>test</` + prefix + `verProc>
    </` + prefix + `ide>
    <` + prefix + `emit>
      <` + prefix + `CNPJ>12345678000195</` + prefix + `CNPJ>
      <` + prefix + `xNome>Emitente</` + prefix + `xNome>
      <` + prefix + `enderEmit>
        <` + prefix + `xLgr>Rua A</` + prefix + `xLgr>
        <` + prefix + `nro>1</` + prefix + `nro>
        <` + prefix + `xBairro>Centro</` + prefix + `xBairro>
        <` + prefix + `cMun>3550308</` + prefix + `cMun>
        <` + prefix + `xMun>Sao Paulo</` + prefix + `xMun>
        <` + prefix + `UF>SP</` + prefix + `UF>
        <` + prefix + `CEP>01001000</` + prefix + `CEP>
      </` + prefix + `enderEmit>
      <` + prefix + `IE>123</` + prefix + `IE>
      <` + prefix + `CRT>1</` + prefix + `CRT>
    </` + prefix + `emit>
    <` + prefix + `det nItem="1">
      <` + prefix + `prod>
        <` + prefix + `cProd>1</` + prefix + `cProd>
        <` + prefix + `cEAN>SEM GTIN</` + prefix + `cEAN>
        <` + prefix + `xProd>Produto</` + prefix + `xProd>
        <` + prefix + `NCM>00000000</` + prefix + `NCM>
        <` + prefix + `CFOP>5102</` + prefix + `CFOP>
        <` + prefix + `uCom>UN</` + prefix + `uCom>
        <` + prefix + `qCom>1.0000</` + prefix + `qCom>
        <` + prefix + `vUnCom>1.0000000000</` + prefix + `vUnCom>
        <` + prefix + `vProd>1.00</` + prefix + `vProd>
        <` + prefix + `cEANTrib>SEM GTIN</` + prefix + `cEANTrib>
        <` + prefix + `uTrib>UN</` + prefix + `uTrib>
        <` + prefix + `qTrib>1.0000</` + prefix + `qTrib>
        <` + prefix + `vUnTrib>1.0000000000</` + prefix + `vUnTrib>
        <` + prefix + `indTot>1</` + prefix + `indTot>
      </` + prefix + `prod>
    </` + prefix + `det>
  </` + prefix + `infNFe>
</` + rootName + `>`
}

func unicodeNFEXML() string {
	return `<NFe xmlns="` + nfeNamespace + `">
  <infNFe versao="4.00" Id="NFe35180834128745000152550010000476121675985748">
    <ide>
      <cUF>35</cUF>
      <cNF>12345678</cNF>
      <natOp>Venda</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>1</nNF>
      <dhEmi>2020-01-01T12:00:00-03:00</dhEmi>
      <tpNF>1</tpNF>
      <idDest>1</idDest>
      <cMunFG>3550308</cMunFG>
      <tpImp>1</tpImp>
      <tpEmis>1</tpEmis>
      <cDV>1</cDV>
      <tpAmb>2</tpAmb>
      <finNFe>1</finNFe>
      <indFinal>1</indFinal>
      <indPres>0</indPres>
      <procEmi>0</procEmi>
      <verProc>teste-ação</verProc>
    </ide>
    <emit>
      <CNPJ>12345678000195</CNPJ>
      <xNome>José &amp; Filhos Comércio de Açúcar Ltda.</xNome>
      <enderEmit>
        <xLgr>Rua São João, nº 123</xLgr>
        <nro>123</nro>
        <xBairro>Centro Histórico</xBairro>
        <cMun>3550308</cMun>
        <xMun>São Paulo</xMun>
        <UF>SP</UF>
        <CEP>01001000</CEP>
      </enderEmit>
      <IE>123</IE>
      <CRT>1</CRT>
    </emit>
    <det nItem="1">
      <prod>
        <cProd>1</cProd>
        <cEAN>SEM GTIN</cEAN>
        <xProd>Café torrado e moído 500g</xProd>
        <NCM>00000000</NCM>
        <CFOP>5102</CFOP>
        <uCom>UN</uCom>
        <qCom>1.0000</qCom>
        <vUnCom>1.0000000000</vUnCom>
        <vProd>1.00</vProd>
        <cEANTrib>SEM GTIN</cEANTrib>
        <uTrib>UN</uTrib>
        <qTrib>1.0000</qTrib>
        <vUnTrib>1.0000000000</vUnTrib>
        <indTot>1</indTot>
      </prod>
    </det>
  </infNFe>
</NFe>`
}

func minimalProcNFEXML(rootName, rootAttrs, prefix string) string {
	return `<` + rootName + rootAttrs + ` xmlns="` + nfeNamespace + `" versao="4.00">
  ` + minimalNFEXML(prefix+`NFe`, "", prefix) + `
  <` + prefix + `protNFe versao="4.00">
    <` + prefix + `infProt>
      <` + prefix + `tpAmb>2</` + prefix + `tpAmb>
      <` + prefix + `chNFe>35180834128745000152550010000476121675985748</` + prefix + `chNFe>
      <` + prefix + `dhRecbto>2020-01-01T12:00:00-03:00</` + prefix + `dhRecbto>
      <` + prefix + `cStat>100</` + prefix + `cStat>
    </` + prefix + `infProt>
  </` + prefix + `protNFe>
</` + rootName + `>`
}

func mixedPrefixProcNFEXML() string {
	return `<a:nfeProc xmlns:a="` + nfeNamespace + `" xmlns:b="` + nfeNamespace + `" xmlns:c="` + nfeNamespace + `" versao="4.00">
  ` + minimalNFEXML("b:NFe", "", "b:") + `
  <c:protNFe versao="4.00">
    <c:infProt>
      <c:tpAmb>2</c:tpAmb>
      <c:chNFe>35180834128745000152550010000476121675985748</c:chNFe>
      <c:dhRecbto>2020-01-01T12:00:00-03:00</c:dhRecbto>
      <c:cStat>100</c:cStat>
    </c:infProt>
  </c:protNFe>
</a:nfeProc>`
}
