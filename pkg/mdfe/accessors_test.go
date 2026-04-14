package mdfe_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/mdfe"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/mdfe/v3_0/41190876676436000167580010000500001000437558-mdfe.xml")
	require.NoError(t, err)

	doc, err := mdfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "41190876676436000167580010000500001000437558", doc.GetAccessKey())
	require.Equal(t, "3.00", doc.GetVersion())
	require.Equal(t, "2", doc.GetEnvironment())
	require.Equal(t, "50000", doc.GetNumber())
	require.Equal(t, "1", doc.GetSeries())
	require.Equal(t, "58", doc.GetModel())
	require.Equal(t, "2019-08-16T15:00:00-02:00", doc.GetIssueDate())
	require.Equal(t, "33.19", doc.GetAmount())
	require.Equal(t, "SUL DEFENSIVOS AGRICOLAS LTDA", doc.GetIssuer())
	require.Equal(t, "76676436000167", doc.GetIssuerDocument())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetRecipientDocument())
	require.False(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "cargo", Value: "33.19"})
	require.Contains(t, doc.GetRelatedDocuments(), info.RelatedDocument{Type: "nfe", AccessKey: "41190806117473000150550010000586251016759484"})
	require.Equal(t, info.Location{State: "PR", CityCode: "4101804", CityName: "ARAUCARIA"}, doc.GetOrigin())
	require.Equal(t, info.Location{State: "SC"}, doc.GetDestination())
}
