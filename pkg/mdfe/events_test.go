package mdfe_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/mdfe"
	"github.com/stretchr/testify/require"
)

func TestLifecycleEventAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/mdfe/v3_0/cancelameto1101103511031029073900013955001000000001105112804101-ped-eve.xml")
	require.NoError(t, err)

	doc, err := mdfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "110111", doc.GetEventType())
	require.Equal(t, "1", doc.GetEventSequence())
	require.Equal(t, "35110310290739000139550010000000011051128041", doc.GetAccessKey())
}
