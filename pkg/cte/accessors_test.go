package cte_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/cte"
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cte/v4_0/43120178408960000182570010000000041000000047-cte.xml")
	require.NoError(t, err)

	doc, err := cte.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "43120178408960000182570010000000041000000047", doc.GetAccessKey())
	require.Equal(t, "3.00", doc.GetVersion())
	require.Equal(t, "2", doc.GetEnvironment())
	require.Equal(t, "4", doc.GetNumber())
	require.Equal(t, "1", doc.GetSeries())
	require.Equal(t, "57", doc.GetModel())
	require.Equal(t, "2012-01-06T17:25:56-02:00", doc.GetIssueDate())
	require.Equal(t, "2300.00", doc.GetAmount())
	require.Equal(t, "KERBER E CIA. LTDA.", doc.GetIssuer())
	require.Equal(t, "78408960000182", doc.GetIssuerDocument())
	require.Equal(t, "HOBI E CIA LTDA. - MATRIZ", doc.GetRecipient())
	require.Equal(t, "81639791000104", doc.GetRecipientDocument())
	require.Empty(t, doc.GetProtocolNumber())
	require.Empty(t, doc.GetStatusCode())
	require.False(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "service", Value: "2300.00"})
	require.Contains(t, doc.GetParties(), info.Party{Role: "sender", Name: "KERBER E CIA. LTDA.", Document: "78408960000182"})
	require.Equal(t, "01", doc.GetModal())
	require.Equal(t, info.Location{State: "SC", CityCode: "4213609", CityName: "PORTO UNIAO"}, doc.GetOrigin())
}

func TestDocumentGetDuplicatas(t *testing.T) {
	t.Parallel()

	t.Run("exposes duplicates and invoice from cobr block", func(t *testing.T) {
		t.Parallel()

		doc, err := cte.Parse([]byte(cteWithCobrXML))
		require.NoError(t, err)

		require.Equal(t, []info.Duplicata{
			{Number: "001", DueDate: "2024-12-01", Amount: "1500.00"},
			{Number: "002", DueDate: "2025-01-01", Amount: "500.00"},
		}, doc.GetDuplicatas())

		billing := doc.GetBilling()
		require.NotNil(t, billing)
		require.NotNil(t, billing.Invoice)
		require.Equal(t, "42", billing.Invoice.Number)
		require.Equal(t, "2000.00", billing.Invoice.NetAmount)
	})

	t.Run("returns nil when cobr is absent", func(t *testing.T) {
		t.Parallel()

		data, err := os.ReadFile("../../testdata/cte/v4_0/43120178408960000182570010000000041000000047-cte.xml")
		require.NoError(t, err)

		doc, err := cte.Parse(data)
		require.NoError(t, err)

		require.Nil(t, doc.GetBilling())
		require.Nil(t, doc.GetDuplicatas())
	})
}

func TestDocumentGetAdditionalInfo(t *testing.T) {
	t.Parallel()

	doc, err := cte.Parse([]byte(cteWithComplXML))
	require.NoError(t, err)

	require.Equal(t, "observation text\nissuer notes\nadditional chars", doc.GetAdditionalInfo())
}

const cteWithCobrXML = `<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe43120178408960000182570010000000041000000047" versao="4.00"><ide><cUF>43</cUF><cCT>00000004</cCT><CFOP>6353</CFOP><natOp>SERV</natOp><mod>57</mod><serie>1</serie><nCT>4</nCT><dhEmi>2012-01-06T17:25:56-02:00</dhEmi><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>7</cDV><tpAmb>2</tpAmb><tpCTe>0</tpCTe><procEmi>0</procEmi><verProc>104</verProc><cMunEnv>4213609</cMunEnv><xMunEnv>PORTO UNIAO</xMunEnv><UFEnv>SC</UFEnv><modal>01</modal><tpServ>0</tpServ><cMunIni>4213609</cMunIni><xMunIni>PORTO UNIAO</xMunIni><UFIni>SC</UFIni><cMunFim>4213609</cMunFim><xMunFim>PORTO UNIAO</xMunFim><UFFim>SC</UFFim><retira>0</retira><indIEToma>9</indIEToma></ide><emit><CNPJ>78408960000182</CNPJ><IE>ISENTO</IE><xNome>KERBER</xNome><enderEmit><xLgr>R</xLgr><nro>1</nro><xBairro>B</xBairro><cMun>4213609</cMun><xMun>PORTO UNIAO</xMun><CEP>89400000</CEP><UF>SC</UF></enderEmit></emit><infCTeNorm><cobr><fat><nFat>42</nFat><vOrig>2000.00</vOrig><vLiq>2000.00</vLiq></fat><dup><nDup>001</nDup><dVenc>2024-12-01</dVenc><vDup>1500.00</vDup></dup><dup><nDup>002</nDup><dVenc>2025-01-01</dVenc><vDup>500.00</vDup></dup></cobr></infCTeNorm></infCte></CTe>`

var cteWithComplXML = fmt.Sprintf(`<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe43120178408960000182570010000000041000000047" versao="4.00"><ide><cUF>43</cUF><cCT>00000004</cCT><CFOP>6353</CFOP><natOp>SERV</natOp><mod>57</mod><serie>1</serie><nCT>4</nCT><dhEmi>2012-01-06T17:25:56-02:00</dhEmi><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>7</cDV><tpAmb>2</tpAmb><tpCTe>0</tpCTe><procEmi>0</procEmi><verProc>104</verProc><cMunEnv>4213609</cMunEnv><xMunEnv>PORTO UNIAO</xMunEnv><UFEnv>SC</UFEnv><modal>01</modal><tpServ>0</tpServ><cMunIni>4213609</cMunIni><xMunIni>PORTO UNIAO</xMunIni><UFIni>SC</UFIni><cMunFim>4213609</cMunFim><xMunFim>PORTO UNIAO</xMunFim><UFFim>SC</UFFim><retira>0</retira><indIEToma>9</indIEToma></ide><compl><xObs>%s</xObs><xEmi>%s</xEmi><xCaracAd>%s</xCaracAd></compl><emit><CNPJ>78408960000182</CNPJ><IE>ISENTO</IE><xNome>KERBER</xNome><enderEmit><xLgr>R</xLgr><nro>1</nro><xBairro>B</xBairro><cMun>4213609</cMun><xMun>PORTO UNIAO</xMun><CEP>89400000</CEP><UF>SC</UF></enderEmit></emit></infCte></CTe>`, "observation text", "issuer notes", "additional chars")
