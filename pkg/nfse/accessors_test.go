package nfse_test

import (
	"os"
	"strings"
	"testing"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfse"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-simples.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "140015920176113500013200900000000000000006", doc.GetAccessKey())
	require.Equal(t, "1.00", doc.GetVersion())
	require.Equal(t, "2", doc.GetEnvironment())
	require.Equal(t, "6", doc.GetNumber())
	require.Equal(t, "900", doc.GetSeries())
	require.Empty(t, doc.GetModel())
	require.Equal(t, "2022-09-28T13:50:29-03:00", doc.GetIssueDate())
	require.Equal(t, "999999999.99", doc.GetAmount())
	require.Equal(t, "01761135000132", doc.GetIssuerDocument())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetStatusCode())
	require.False(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "service", Value: "999999999.99"})
	require.Contains(t, doc.GetParties(), info.Party{Role: "provider", Document: "01761135000132"})
}

func TestDocumentConvenienceAccessorsHandleIssuedNFSe(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/ConsultarNFSeEnvio-ped-sitnfse.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "14001591201761135000132000000000000022097781063609", doc.GetAccessKey())
	require.Equal(t, "2", doc.GetNumber())
	require.Equal(t, "989999961.04", doc.GetAmount())
	require.Equal(t, "LW SOFTWARES LTDA", doc.GetIssuer())
	require.Equal(t, "01761135000132", doc.GetIssuerDocument())
	require.Equal(t, "100", doc.GetStatusCode())
	require.True(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "service", Value: "999999999.99"})
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "net", Value: "989999961.04"})
}

func TestDocumentIsAuthorizedAcceptsGeneratedNFSeStatuses(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/ConsultarNFSeEnvio-ped-sitnfse.xml")
	require.NoError(t, err)

	for _, status := range []string{"100", "101", "102", "103", "107"} {
		t.Run(status, func(t *testing.T) {
			doc, err := nfse.Parse([]byte(strings.Replace(string(data), "<cStat>100</cStat>", "<cStat>"+status+"</cStat>", 1)))
			require.NoError(t, err)
			require.True(t, doc.IsAuthorized())
		})
	}

	doc, err := nfse.Parse([]byte(strings.Replace(string(data), "<cStat>100</cStat>", "<cStat>999</cStat>", 1)))
	require.NoError(t, err)
	require.False(t, doc.IsAuthorized())
}

func TestDocumentGetAmountsIncludesRetentions(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-regime-normal.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "retained_inss", Value: "0.40"})
}

func TestDocumentGetAmountsIncludesTaxBreakdown(t *testing.T) {
	data := []byte(`<NFSe xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.00"><infNFSe Id="NFS00000000000000000000000000000000000000000001"><nNFSe>1</nNFSe><verAplic>1.0</verAplic><ambGer>2</ambGer><tpEmis>1</tpEmis><procEmi>1</procEmi><cStat>100</cStat><dhProc>2024-01-02T03:04:05-03:00</dhProc><nDFSe>1</nDFSe><emit><CNPJ>12345678000195</CNPJ><xNome>PROVIDER LTDA</xNome></emit><valores><vBC>1000.00</vBC><pAliqAplic>5.00</pAliqAplic><vISSQN>50.00</vISSQN><vTotalRet>23.50</vTotalRet><vLiq>950.00</vLiq></valores><DPS versao="1.00"><infDPS Id="DPS00000000000000000000000000000000000000000001"><tpAmb>2</tpAmb><dhEmi>2024-01-02T03:04:05-03:00</dhEmi><verAplic>1.0</verAplic><serie>1</serie><nDPS>1</nDPS><dCompet>2024-01-02</dCompet><tpEmit>1</tpEmit><cLocEmi>3550308</cLocEmi><subst><chSubstda>00000000000000000000000000000000000000000000</chSubstda><cMotivo>1</cMotivo></subst><prest><CNPJ>12345678000195</CNPJ><xNome>PROVIDER LTDA</xNome><regTrib><opSimpNac>1</opSimpNac><regApTribSN>1</regApTribSN></regTrib></prest><serv><locPrest><cLocPrestacao>3550308</cLocPrestacao></locPrest><cServ><cTribNac>0101010101</cTribNac><cTribMun>0101010101</cTribMun><xDescServ>desc</xDescServ><cNBS>110000000</cNBS></cServ></serv><valores><vServPrest><vServ>1000.00</vServ></vServPrest><trib><tribMun><tribISSQN>1</tribISSQN><tpRetISSQN>1</tpRetISSQN></tribMun><tribFed><piscofins><CST>01</CST><vBCPisCofins>1000.00</vBCPisCofins><pAliqPis>0.65</pAliqPis><pAliqCofins>3.00</pAliqCofins><vPis>6.50</vPis><vCofins>30.00</vCofins><tpRetPisCofins>2</tpRetPisCofins></piscofins><vRetCP>0.00</vRetCP><vRetIRRF>0.00</vRetIRRF><vRetCSLL>0.00</vRetCSLL></tribFed><totTrib><vTotTrib><vTotTribFed>36.50</vTotTribFed><vTotTribEst>0.00</vTotTribEst><vTotTribMun>50.00</vTotTribMun></vTotTrib><indTotTrib>1</indTotTrib></totTrib></trib></valores></infDPS></DPS></infNFSe></NFSe>`)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	amounts := doc.GetAmounts()
	require.Contains(t, amounts, info.Amount{Type: "service", Value: "1000.00"})
	require.Contains(t, amounts, info.Amount{Type: "net", Value: "950.00"})
	require.Contains(t, amounts, info.Amount{Type: "retained", Value: "23.50"})
	require.Contains(t, amounts, info.Amount{Type: "tax_iss", Value: "50.00"})
	require.Contains(t, amounts, info.Amount{Type: "tax_pis", Value: "6.50"})
	require.Contains(t, amounts, info.Amount{Type: "tax_cofins", Value: "30.00"})
	require.Contains(t, amounts, info.Amount{Type: "taxes_fed", Value: "36.50"})
	require.Contains(t, amounts, info.Amount{Type: "taxes_mun", Value: "50.00"})
	for _, a := range amounts {
		require.NotEqual(t, "retained_pis", a.Type, "non-retained pis should not appear as retention")
		require.NotEqual(t, "taxes_est", a.Type, "zero-valued taxes_est should be filtered out")
	}
}

func TestDocumentGetAmountsRetainedISS(t *testing.T) {
	t.Run("real production cooperativa with ISS retention", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/nfse-prod-iss-retido-cooperativa.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		amounts := doc.GetAmounts()
		require.Contains(t, amounts, info.Amount{Type: "net", Value: "9878.06"})
		require.Contains(t, amounts, info.Amount{Type: "tax_iss", Value: "519.90"})
		require.Contains(t, amounts, info.Amount{Type: "retained", Value: "519.90"})
		require.Contains(t, amounts, info.Amount{Type: "retained_iss", Value: "519.90"})
	})

	t.Run("ISS retention with IRRF and CSLL", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/third_party/nfse-iss-retido-irrf-csll.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		amounts := doc.GetAmounts()
		require.Contains(t, amounts, info.Amount{Type: "retained_iss", Value: "500.00"})
		require.Contains(t, amounts, info.Amount{Type: "retained_irrf", Value: "150.00"})
		require.Contains(t, amounts, info.Amount{Type: "retained_csll", Value: "465.00"})
	})

	t.Run("ISS retention with INSS on transport service", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/third_party/nfse-iss-retido-transporte.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		amounts := doc.GetAmounts()
		require.Contains(t, amounts, info.Amount{Type: "retained_iss", Value: "150.00"})
		require.Contains(t, amounts, info.Amount{Type: "retained_inss", Value: "550.00"})
	})

	t.Run("no retained_iss when tpRetISSQN=1", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/ConsultarNFSeEnvio-ped-sitnfse.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		for _, a := range doc.GetAmounts() {
			require.NotEqual(t, "retained_iss", a.Type)
		}
	})

	t.Run("substituted NFSe has no ISS retention", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/nfse-prod-substituicao.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		for _, a := range doc.GetAmounts() {
			require.NotEqual(t, "retained_iss", a.Type)
		}
	})
}

func TestDocumentGetCompetenceDate(t *testing.T) {
	t.Run("standalone DPS", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-simples.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)
		require.Equal(t, "2022-09-28", doc.GetCompetenceDate())
	})

	t.Run("authorized NFSe with nested DPS", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/nfse-prod-iss-retido-cooperativa.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)
		require.Equal(t, "2026-02-03", doc.GetCompetenceDate())
	})

	t.Run("CPF taker DPS", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-cpf-taker-piscofins.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)
		require.Equal(t, "2025-12-04", doc.GetCompetenceDate())
	})
}
