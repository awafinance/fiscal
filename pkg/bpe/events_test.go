package bpe_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/stretchr/testify/require"
)

func TestLifecycleEventAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/bpe/v1_0/1101114319081234567800019563001000000001100000001101-eventoBPe.xml")
	require.NoError(t, err)

	doc, err := bpe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "110111", doc.GetEventType())
	require.Equal(t, "1", doc.GetEventSequence())
	require.Equal(t, "43190812345678000195630010000000011000000011", doc.GetAccessKey())
}
