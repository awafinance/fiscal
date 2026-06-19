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
	// Fixture has <toma3><toma>0</toma></toma3>, so the tomador (recipient) is the remetente.
	require.Equal(t, "KERBER E CIA. LTDA.", doc.GetRecipient())
	require.Equal(t, "78408960000182", doc.GetRecipientDocument())
	require.Empty(t, doc.GetProtocolNumber())
	require.Empty(t, doc.GetStatusCode())
	require.False(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "service", Value: "2300.00"})
	sender := requireParty(t, doc.GetParties(), "sender")
	require.Equal(t, "KERBER E CIA. LTDA.", sender.Name)
	require.Equal(t, "78408960000182", sender.Document)
	require.Equal(t, "251079554", sender.StateRegistration)
	require.Equal(t, "4235224933", sender.Phone)
	require.Equal(t, "pedreira@kerberecia.com.br", sender.Email)
	require.Equal(t, &info.Address{
		Street:       "ESTRADA VELHA DE PALMAS",
		Number:       "S/N",
		Complement:   "CAIXA POSTAL 268",
		Neighborhood: "RIO DA AREIA",
		PostalCode:   "89400000",
		CityCode:     "4213609",
		CityName:     "PORTO UNIAO",
		State:        "SC",
		CountryCode:  "1058",
	}, sender.Address)
	addressee := requireParty(t, doc.GetParties(), "addressee")
	require.Equal(t, "HOBI E CIA LTDA. - MATRIZ", addressee.Name)
	require.Equal(t, "81639791000104", addressee.Document)
	require.Equal(t, "3010264714", addressee.StateRegistration)
	require.Equal(t, "4235211922", addressee.Phone)
	require.Equal(t, "84600000", addressee.Address.PostalCode)
	require.Equal(t, "01", doc.GetModal())
	require.Equal(t, info.Location{State: "SC", CityCode: "4213609", CityName: "PORTO UNIAO"}, doc.GetOrigin())
}

func TestDocumentGetAmountsIncludesTaxBreakdown(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cte/v4_0/35170799999999999999670000000000261309301440-cte-of.xml")
	require.NoError(t, err)

	doc, err := cte.Parse(data)
	require.NoError(t, err)

	amounts := doc.GetAmounts()
	require.Contains(t, amounts, info.Amount{Type: "service", Value: "2500.00"})
	require.Contains(t, amounts, info.Amount{Type: "tax_icms", Value: "450.00"})
	for _, a := range amounts {
		require.NotEqual(t, "tax_pis", a.Type, "zero-valued tax_pis should be filtered out")
	}

	issuer := requireParty(t, doc.GetParties(), "issuer")
	require.Equal(t, "99999999999999", issuer.Document)
	require.Equal(t, "999999999999", issuer.StateRegistration)
	require.Equal(t, "9999999999", issuer.Phone)
	require.Equal(t, &info.Address{
		Street:       "XXXXXXXXXXXXXXXX",
		Number:       "720",
		Neighborhood: "XXXXXXXXXXX",
		PostalCode:   "99999999",
		CityCode:     "3534401",
		CityName:     "SAO PAULO",
		State:        "SP",
	}, issuer.Address)

	recipient := requireParty(t, doc.GetParties(), "recipient")
	require.Equal(t, "XXXXXXXXXXXXXXXXXXXXXX", recipient.Name)
	require.Equal(t, "99999999999999", recipient.Document)
	require.Equal(t, "999999999999", recipient.StateRegistration)
	require.Equal(t, "9999999999", recipient.Phone)
	require.Equal(t, "xxxxxx@xxxxxxx.com.br", recipient.Email)
	require.Equal(t, &info.Address{
		Street:       "XXXXXXXXXXXXXXXXXXXX",
		Number:       "209",
		Neighborhood: "XXXXXXXXXXXXX",
		PostalCode:   "99999999",
		CityCode:     "3550308",
		CityName:     "SAO PAULO",
		State:        "SP",
		CountryCode:  "1058",
	}, recipient.Address)
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

func TestDocumentResolvesTomadorViaToma3(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		toma         string
		wantName     string
		wantDocument string
	}{
		{"remetente", "0", "REM LTDA", "11111111000111"},
		{"expedidor", "1", "EXPED LTDA", "22222222000122"},
		{"recebedor", "2", "RECEB LTDA", "33333333000133"},
		{"destinatario", "3", "DEST LTDA", "44444444000144"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := cte.Parse([]byte(cteWithTomadorXML(tc.toma)))
			require.NoError(t, err)

			require.Equal(t, tc.wantName, doc.GetRecipient())
			require.Equal(t, tc.wantDocument, doc.GetRecipientDocument())
		})
	}
}

func TestDocumentResolvesTomadorViaToma4(t *testing.T) {
	t.Parallel()

	doc, err := cte.Parse([]byte(cteWithToma4XML))
	require.NoError(t, err)

	require.Equal(t, "OUTRO PAGADOR LTDA", doc.GetRecipient())
	require.Equal(t, "55555555000155", doc.GetRecipientDocument())
}

func cteWithTomadorXML(toma string) string {
	return fmt.Sprintf(`<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe43120178408960000182570010000000041000000047" versao="4.00"><ide><cUF>43</cUF><cCT>00000004</cCT><CFOP>6353</CFOP><natOp>SERV</natOp><mod>57</mod><serie>1</serie><nCT>4</nCT><dhEmi>2012-01-06T17:25:56-02:00</dhEmi><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>7</cDV><tpAmb>2</tpAmb><tpCTe>0</tpCTe><procEmi>0</procEmi><verProc>104</verProc><cMunEnv>4213609</cMunEnv><xMunEnv>PORTO UNIAO</xMunEnv><UFEnv>SC</UFEnv><modal>01</modal><tpServ>0</tpServ><cMunIni>4213609</cMunIni><xMunIni>PORTO UNIAO</xMunIni><UFIni>SC</UFIni><cMunFim>4213609</cMunFim><xMunFim>PORTO UNIAO</xMunFim><UFFim>SC</UFFim><retira>0</retira><indIEToma>9</indIEToma><toma3><toma>%s</toma></toma3></ide><emit><CNPJ>78408960000182</CNPJ><IE>ISENTO</IE><xNome>EMIT LTDA</xNome></emit><rem><CNPJ>11111111000111</CNPJ><xNome>REM LTDA</xNome></rem><exped><CNPJ>22222222000122</CNPJ><xNome>EXPED LTDA</xNome></exped><receb><CNPJ>33333333000133</CNPJ><xNome>RECEB LTDA</xNome></receb><dest><CNPJ>44444444000144</CNPJ><xNome>DEST LTDA</xNome></dest></infCte></CTe>`, toma)
}

const cteWithToma4XML = `<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe43120178408960000182570010000000041000000047" versao="4.00"><ide><cUF>43</cUF><cCT>00000004</cCT><CFOP>6353</CFOP><natOp>SERV</natOp><mod>57</mod><serie>1</serie><nCT>4</nCT><dhEmi>2012-01-06T17:25:56-02:00</dhEmi><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>7</cDV><tpAmb>2</tpAmb><tpCTe>0</tpCTe><procEmi>0</procEmi><verProc>104</verProc><cMunEnv>4213609</cMunEnv><xMunEnv>PORTO UNIAO</xMunEnv><UFEnv>SC</UFEnv><modal>01</modal><tpServ>0</tpServ><cMunIni>4213609</cMunIni><xMunIni>PORTO UNIAO</xMunIni><UFIni>SC</UFIni><cMunFim>4213609</cMunFim><xMunFim>PORTO UNIAO</xMunFim><UFFim>SC</UFFim><retira>0</retira><indIEToma>9</indIEToma><toma4><toma>4</toma><CNPJ>55555555000155</CNPJ><xNome>OUTRO PAGADOR LTDA</xNome></toma4></ide><emit><CNPJ>78408960000182</CNPJ><IE>ISENTO</IE><xNome>EMIT LTDA</xNome></emit><rem><CNPJ>11111111000111</CNPJ><xNome>REM LTDA</xNome></rem><dest><CNPJ>44444444000144</CNPJ><xNome>DEST LTDA</xNome></dest></infCte></CTe>`

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
