package nfse_test

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfse"
	"github.com/stretchr/testify/require"
)

func TestPedRegEventoIsRequest(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/CancelarNFSe-ped-cannfse.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "e101101", doc.GetEventType())
	require.Equal(t, "14001591201761135000132000000000000022096100197260", doc.GetAccessKey())
	// A bare request carries no registration sequence or protocol.
	require.Empty(t, doc.GetEventSequence())
	require.Empty(t, doc.GetProtocolNumber())
	require.Equal(t, &info.LifecycleEventFacts{
		RegistrationState: info.LifecycleEventRegistrationStateRequest,
		Type:              "e101101",
		RequestNumber:     "001",
		IssueDate:         "2022-09-28T13:50:29-03:00",
	}, doc.GetLifecycleEventFacts())
}

func TestSubstituicaoBackRef(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/nfse-prod-substituicao.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Contains(t, doc.GetRelatedDocuments(), info.RelatedDocument{
		Type:      "nfse",
		AccessKey: "33045572214043710000103000000000001726012751532718",
	})
}

func TestRegisteredEvento(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/EventoNFSe-cancelamento-registrado.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "e101101", doc.GetEventType())
	require.Equal(t, "14001591201761135000132000000000000022096100197260", doc.GetAccessKey())
	require.Equal(t, "1", doc.GetEnvironment())
	require.Equal(t, "2022-09-28T13:50:29-03:00", doc.GetIssueDate())
	// A registered evento carries nSeqEvento; that is how it differs from a request.
	require.Equal(t, "002", doc.GetEventSequence())
	require.Equal(t, &info.LifecycleEventFacts{
		RegistrationState: info.LifecycleEventRegistrationStateRegistered,
		Type:              "e101101",
		Sequence:          "002",
		RequestNumber:     "001",
		IssueDate:         "2022-09-28T13:50:29-03:00",
		ProcessingTime:    "2022-09-28T13:50:30-03:00",
	}, doc.GetLifecycleEventFacts())

	// Round-trip: marshal back to XML and re-parse; the event must survive intact.
	out, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Contains(t, string(out), "<evento")

	reparsed, err := nfse.Parse(out)
	require.NoError(t, err)
	require.Equal(t, "evento", reparsed.RootName)
	require.Equal(t, "e101101", reparsed.GetEventType())
	require.Equal(t, "002", reparsed.GetEventSequence())
	require.Equal(t, doc.GetAccessKey(), reparsed.GetAccessKey())
}

func TestRegisteredSubstituicaoEventBackRef(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/EventoNFSe-substituicao-registrado.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "e105102", doc.GetEventType())
	require.Equal(t, "001", doc.GetEventSequence())
	require.Contains(t, doc.GetRelatedDocuments(), info.RelatedDocument{
		Type:      "nfse",
		AccessKey: "33045572214043710000103000000000001826012751532719",
	})
}
