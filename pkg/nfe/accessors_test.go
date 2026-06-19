package nfe_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfe"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "42220575277525000178550030000292481295366801-procNFe.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "42220575277525000178550030000292481295366801", doc.GetAccessKey())
	require.Equal(t, "4.00", doc.GetVersion())
	require.Equal(t, "1", doc.GetEnvironment())
	require.Equal(t, "29248", doc.GetNumber())
	require.Equal(t, "3", doc.GetSeries())
	require.Equal(t, "55", doc.GetModel())
	require.Equal(t, "2022-05-27T08:52:04-03:00", doc.GetIssueDate())
	require.Equal(t, "64237.04", doc.GetAmount())
	require.Equal(t, "FORNOS LTDA", doc.GetIssuer())
	require.Equal(t, "75277525000178", doc.GetIssuerDocument())
	require.Equal(t, "Jung Usa Corporation", doc.GetRecipient())
	require.Equal(t, "371780142", doc.GetRecipientDocument())
	require.Equal(t, "342220106391922", doc.GetProtocolNumber())
	require.Equal(t, "100", doc.GetStatusCode())
	require.Equal(t, "Autorizado o uso da NF-e", doc.GetStatusReason())
	require.True(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "total", Value: "64237.04"})
	issuer := requireParty(t, doc.GetParties(), "issuer")
	require.Equal(t, "FORNOS LTDA", issuer.Name)
	require.Equal(t, "75277525000178", issuer.Document)
	require.Equal(t, "250745615", issuer.StateRegistration)
	require.Equal(t, "554733270000", issuer.Phone)
	require.Equal(t, "3", issuer.SimpleNationalOption)
	require.Equal(t, &info.Address{
		Street:       "Rua Bahia",
		Number:       "3465",
		Neighborhood: "Salto",
		PostalCode:   "89031002",
		CityCode:     "4202404",
		CityName:     "Blumenau",
		State:        "SC",
		CountryCode:  "1058",
	}, issuer.Address)
	recipient := requireParty(t, doc.GetParties(), "recipient")
	require.Equal(t, "Jung Usa Corporation", recipient.Name)
	require.Equal(t, "371780142", recipient.Document)
	require.Equal(t, "export@jung.com.br", recipient.Email)
	require.Equal(t, "00037741", recipient.Address.PostalCode)
	require.Equal(t, "2496", recipient.Address.CountryCode)
}

func TestDocumentGetItems(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "42220575277525000178550030000292481295366801-procNFe.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	items := doc.GetItems()
	require.Len(t, items, 2)
	require.Equal(t, nfe.Item{
		Number:      "1",
		Code:        "TB2001210",
		EAN:         "SEM GTIN",
		Description: "[TB2001210] FORNO LINHA INDUSTRIAL MODELO TB2001210",
		NCM:         "85141900",
		CFOP:        "7101",
		Unit:        "UN",
		Quantity:    "1.0000",
		UnitAmount:  "23942.1312000000",
		Amount:      "23942.13",
	}, items[0])
}

func TestDocumentGetPayments(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "42220575277525000178550030000292481295366801-procNFe.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, []info.Payment{
		{Method: "90", Amount: "0.00"},
	}, doc.GetPayments())
}

func TestDocumentGetAdditionalInfo(t *testing.T) {
	data := []byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe versao="4.00" Id="NFe35180834128745000152550010000476121675985748"><ide><cUF>35</cUF><cNF>1</cNF><natOp>Venda</natOp><mod>55</mod><serie>1</serie><nNF>1</nNF><dhEmi>2020-01-01T12:00:00-03:00</dhEmi><tpNF>1</tpNF><idDest>1</idDest><cMunFG>3550308</cMunFG><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>1</cDV><tpAmb>2</tpAmb><finNFe>1</finNFe><indFinal>1</indFinal><indPres>0</indPres><procEmi>0</procEmi><verProc>x</verProc></ide><emit><CNPJ>12345678000195</CNPJ><xNome>E</xNome><enderEmit><xLgr>R</xLgr><nro>1</nro><xBairro>B</xBairro><cMun>3550308</cMun><xMun>Sao Paulo</xMun><UF>SP</UF><CEP>01001000</CEP></enderEmit><IE>123</IE><CRT>1</CRT></emit><det nItem="1"><prod><cProd>1</cProd><cEAN>SEM GTIN</cEAN><xProd>P</xProd><NCM>00000000</NCM><CFOP>5102</CFOP><uCom>UN</uCom><qCom>1</qCom><vUnCom>1</vUnCom><vProd>1</vProd><cEANTrib>SEM GTIN</cEANTrib><uTrib>UN</uTrib><qTrib>1</qTrib><vUnTrib>1</vUnTrib><indTot>1</indTot></prod></det><infAdic><infAdFisco>fiscal observation</infAdFisco><infCpl>boleto linha digitavel 12345</infCpl></infAdic></infNFe></NFe>`)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "boleto linha digitavel 12345\nfiscal observation", doc.GetAdditionalInfo())
}

func TestDocumentGetAmountsIncludesTaxBreakdown(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "nfe_reforma_tributaria.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	amounts := doc.GetAmounts()
	require.Contains(t, amounts, info.Amount{Type: "tax_icms", Value: "18972.77"})
	require.Contains(t, amounts, info.Amount{Type: "tax_icms_st", Value: "6640.46"})
	require.Contains(t, amounts, info.Amount{Type: "tax_pis", Value: "2295.71"})
	require.Contains(t, amounts, info.Amount{Type: "tax_cofins", Value: "10574.15"})
	require.Contains(t, amounts, info.Amount{Type: "tax_ibs", Value: "50"})
	require.Contains(t, amounts, info.Amount{Type: "tax_cbs", Value: "87.29"})
	require.Contains(t, amounts, info.Amount{Type: "tax_is", Value: "1000"})
	for _, a := range amounts {
		require.NotEqual(t, "tax_ipi", a.Type, "zero-valued tax_ipi should be filtered out")
	}
}

func requireParty(t *testing.T, parties []info.Party, role string) info.Party {
	t.Helper()
	for _, party := range parties {
		if party.Role == role {
			return party
		}
	}
	require.Failf(t, "party not found", "role %q in %#v", role, parties)
	return info.Party{}
}

func TestDocumentGetAmountsIncludesRetentions(t *testing.T) {
	data := []byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe versao="4.00" Id="NFe35180834128745000152550010000476121675985748"><ide><cUF>35</cUF><cNF>1</cNF><natOp>Venda</natOp><mod>55</mod><serie>1</serie><nNF>1</nNF><dhEmi>2020-01-01T12:00:00-03:00</dhEmi><tpNF>1</tpNF><idDest>1</idDest><cMunFG>3550308</cMunFG><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>1</cDV><tpAmb>2</tpAmb><finNFe>1</finNFe><indFinal>1</indFinal><indPres>0</indPres><procEmi>0</procEmi><verProc>x</verProc></ide><emit><CNPJ>12345678000195</CNPJ><xNome>E</xNome><enderEmit><xLgr>R</xLgr><nro>1</nro><xBairro>B</xBairro><cMun>3550308</cMun><xMun>Sao Paulo</xMun><UF>SP</UF><CEP>01001000</CEP></enderEmit><IE>123</IE><CRT>1</CRT></emit><det nItem="1"><prod><cProd>1</cProd><cEAN>SEM GTIN</cEAN><xProd>P</xProd><NCM>00000000</NCM><CFOP>5102</CFOP><uCom>UN</uCom><qCom>1</qCom><vUnCom>1</vUnCom><vProd>1000.00</vProd><cEANTrib>SEM GTIN</cEANTrib><uTrib>UN</uTrib><qTrib>1</qTrib><vUnTrib>1</vUnTrib><indTot>1</indTot></prod></det><total><ICMSTot><vBC>0</vBC><vICMS>0</vICMS><vICMSDeson>0</vICMSDeson><vFCP>0</vFCP><vBCST>0</vBCST><vST>0</vST><vFCPST>0</vFCPST><vFCPSTRet>0</vFCPSTRet><vProd>1000.00</vProd><vFrete>0</vFrete><vSeg>0</vSeg><vDesc>0</vDesc><vII>0</vII><vIPI>0</vIPI><vIPIDevol>0</vIPIDevol><vPIS>0</vPIS><vCOFINS>0</vCOFINS><vOutro>0</vOutro><vNF>1000.00</vNF></ICMSTot><retTrib><vRetPIS>6.50</vRetPIS><vRetCOFINS>30.00</vRetCOFINS><vRetCSLL>10.00</vRetCSLL><vIRRF>15.00</vIRRF><vRetPrev>110.00</vRetPrev></retTrib></total></infNFe></NFe>`)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	amounts := doc.GetAmounts()
	require.Contains(t, amounts, info.Amount{Type: "retained_pis", Value: "6.50"})
	require.Contains(t, amounts, info.Amount{Type: "retained_cofins", Value: "30.00"})
	require.Contains(t, amounts, info.Amount{Type: "retained_csll", Value: "10.00"})
	require.Contains(t, amounts, info.Amount{Type: "retained_irrf", Value: "15.00"})
	require.Contains(t, amounts, info.Amount{Type: "retained_inss", Value: "110.00"})
}

func TestDocumentConvenienceAccessorsHandleResNFe(t *testing.T) {
	data := []byte(`<resNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.01">
		<chNFe>35180803102452000172550010000476051695511860</chNFe>
		<CNPJ>12345678000195</CNPJ>
		<xNome>Emitente Teste</xNome>
		<IE>ISENTO</IE>
		<dhEmi>2024-01-02T03:04:05-03:00</dhEmi>
		<tpNF>1</tpNF>
		<vNF>100.00</vNF>
		<dhRecbto>2024-01-02T03:05:06-03:00</dhRecbto>
		<nProt>123456789012345</nProt>
		<cSitNFe>1</cSitNFe>
	</resNFe>`)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "35180803102452000172550010000476051695511860", doc.GetAccessKey())
	require.Equal(t, "100.00", doc.GetAmount())
	require.Equal(t, "Emitente Teste", doc.GetIssuer())
	require.Equal(t, "12345678000195", doc.GetIssuerDocument())
	require.Equal(t, "2024-01-02T03:04:05-03:00", doc.GetIssueDate())
	require.Equal(t, "123456789012345", doc.GetProtocolNumber())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetStatusCode())
	require.False(t, doc.IsAuthorized())
}

func TestDocumentConvenienceAccessorsHandleNilDocument(t *testing.T) {
	var doc *nfe.Document

	require.Empty(t, doc.GetAccessKey())
	require.Empty(t, doc.GetNumber())
	require.Empty(t, doc.GetAmount())
	require.Empty(t, doc.GetIssuer())
	require.Empty(t, doc.GetIssuerDocument())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetRecipientDocument())
	require.Empty(t, doc.GetProtocolNumber())
	require.Empty(t, doc.GetStatusCode())
	require.Empty(t, doc.GetStatusReason())
	require.False(t, doc.IsAuthorized())
	require.Nil(t, doc.GetItems())
	require.Nil(t, doc.GetPayments())
}
