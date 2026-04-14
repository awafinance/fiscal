package cte_test

import (
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
