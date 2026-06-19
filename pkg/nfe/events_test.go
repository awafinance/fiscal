package nfe_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/nfe"
	"github.com/stretchr/testify/require"
)

func TestLifecycleEventAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfe_evento_cancel/v1_0/35180803102452000172550010000476051695511860-cancel.xml")
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "110111", doc.GetEventType())
	require.Equal(t, "1", doc.GetEventSequence())
	require.Equal(t, "35180803102452000172550010000476051695511860", doc.GetAccessKey())
}

func TestLifecycleEventAccessorsHandleResEvento(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfe_dist_dfe/v1_0/35180803102452000172550010000476051695511860-resEvento.xml")
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "110111", doc.GetEventType())
	require.Equal(t, "1", doc.GetEventSequence())
	require.Equal(t, "35180803102452000172550010000476051695511860", doc.GetAccessKey())
	require.Equal(t, "2018-08-31T09:09:09-03:00", doc.GetIssueDate())
	require.Equal(t, "135180000000001", doc.GetProtocolNumber())
	require.Empty(t, doc.GetStatusCode())
	require.Empty(t, doc.GetStatusReason())
	require.False(t, doc.IsAuthorized())
}
