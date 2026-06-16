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
	doc, err := nfse.Parse([]byte(nfseEventoXML))
	require.NoError(t, err)

	require.Equal(t, "e101101", doc.GetEventType())
	require.Equal(t, "14001591201761135000132000000000000022096100197260", doc.GetAccessKey())
	require.Equal(t, "1", doc.GetEnvironment())
	require.Equal(t, "2022-09-28T13:50:29-03:00", doc.GetIssueDate())
	// A registered evento carries nSeqEvento; that is how it differs from a request.
	require.Equal(t, "2", doc.GetEventSequence())

	// Round-trip: marshal back to XML and re-parse; the event must survive intact.
	out, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Contains(t, string(out), "<evento")

	reparsed, err := nfse.Parse(out)
	require.NoError(t, err)
	require.Equal(t, "evento", reparsed.RootName)
	require.Equal(t, "e101101", reparsed.GetEventType())
	require.Equal(t, "2", reparsed.GetEventSequence())
	require.Equal(t, doc.GetAccessKey(), reparsed.GetAccessKey())
}

const nfseEventoXML = `<evento xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.00"><infEvento Id="EVT1"><ambGer>1</ambGer><nSeqEvento>2</nSeqEvento><dhProc>2022-09-28T13:50:30-03:00</dhProc><pedRegEvento versao="1.00"><infPedReg Id="PRE1"><tpAmb>2</tpAmb><dhEvento>2022-09-28T13:50:29-03:00</dhEvento><chNFSe>14001591201761135000132000000000000022096100197260</chNFSe><nPedRegEvento>1</nPedRegEvento><e101101><xDesc>Cancelamento</xDesc><cMotivo>1</cMotivo><xMotivo>erro na emissao</xMotivo></e101101></infPedReg></pedRegEvento></infEvento></evento>`
