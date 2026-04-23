package nfse_test

import (
	"os"
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
}

func TestDocumentGetAmountsIncludesRetentions(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-regime-normal.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "retained_inss", Value: "0.40"})
}

func TestDocumentGetAmountsRetainedISS(t *testing.T) {
	t.Run("real production cooperativa with ISS retention", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/nfse-prod-iss-retido-cooperativa.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		amounts := doc.GetAmounts()
		require.Contains(t, amounts, info.Amount{Type: "net", Value: "9878.06"})
		require.Contains(t, amounts, info.Amount{Type: "iss", Value: "519.90", Aliquot: "5.00"})
		require.Contains(t, amounts, info.Amount{Type: "retained", Value: "519.90"})
		require.Contains(t, amounts, info.Amount{Type: "retained_iss", Value: "519.90", Aliquot: "5.00"})
	})

	t.Run("ISS retention with IRRF and CSLL", func(t *testing.T) {
		data, err := os.ReadFile("../../testdata/nfse/v1_0/third_party/nfse-iss-retido-irrf-csll.xml")
		require.NoError(t, err)

		doc, err := nfse.Parse(data)
		require.NoError(t, err)

		amounts := doc.GetAmounts()
		require.Contains(t, amounts, info.Amount{Type: "retained_iss", Value: "500.00", Aliquot: "5.00"})
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
