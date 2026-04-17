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
