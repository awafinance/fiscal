package fiscal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRoutesByNamespace(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		family   Family
		rootName string
	}{
		{
			name:     "nfe",
			path:     "testdata/nfe/42220575277525000178550030000292481295366801-procNFe.xml",
			family:   NFe,
			rootName: "nfeProc",
		},
		{
			name:     "nfse",
			path:     "testdata/nfse/v1_0/dps-simples.xml",
			family:   NFSe,
			rootName: "DPS",
		},
		{
			name:     "cte",
			path:     "testdata/cte/v4_0/43120178408960000182570010000000041000000047-cte.xml",
			family:   CTe,
			rootName: "CTe",
		},
		{
			name:     "mdfe",
			path:     "testdata/mdfe/v3_0/41190876676436000167580010000500001000437558-mdfe.xml",
			family:   MDFe,
			rootName: "MDFe",
		},
		{
			name:     "bpe",
			path:     "testdata/bpe/v1_0/43190812345678000195630010000000011000000011-bpe.xml",
			family:   BPe,
			rootName: "BPe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.path)
			require.NoError(t, err)

			doc, err := Parse(data)
			require.NoError(t, err)
			require.Equal(t, tt.family, doc.Family)
			require.Equal(t, tt.rootName, doc.RootName)
			require.NotNil(t, doc.Info())
			require.Implements(t, (*AmountsInfo)(nil), doc.Info())
			require.Implements(t, (*PartiesInfo)(nil), doc.Info())
		})
	}
}

func TestDocumentNFeConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile("testdata/nfe/42220575277525000178550030000292481295366801-procNFe.xml")
	require.NoError(t, err)

	doc, err := Parse(data)
	require.NoError(t, err)

	info := doc.Info()
	require.NotNil(t, info)
	require.Equal(t, "42220575277525000178550030000292481295366801", info.GetAccessKey())
	require.Equal(t, "4.00", info.GetVersion())
	require.Equal(t, "1", info.GetEnvironment())
	require.Equal(t, "64237.04", info.GetAmount())
	require.Equal(t, "FORNOS LTDA", info.GetIssuer())
	require.Equal(t, "75277525000178", info.GetIssuerDocument())
	require.Equal(t, "Jung Usa Corporation", info.GetRecipient())
	require.Equal(t, "342220106391922", info.GetProtocolNumber())
	require.True(t, info.IsAuthorized())
}

func TestParseDetectsPrefixedRootNamespace(t *testing.T) {
	data := []byte(`<x:NFe xmlns:x="http://www.portalfiscal.inf.br/nfe"></x:NFe>`)

	_, err := Parse(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse nfe:")
	require.NotErrorIs(t, err, ErrUnsupportedNamespace)
}

func TestParseRejectsUnsupportedNamespace(t *testing.T) {
	_, err := Parse([]byte(`<doc xmlns="urn:example"></doc>`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedNamespace)

	var nsErr *UnsupportedNamespaceError
	require.ErrorAs(t, err, &nsErr)
	require.Equal(t, "urn:example", nsErr.Namespace)
	require.Equal(t, "doc", nsErr.Root)
}

func TestParseRejectsUnsupportedRoot(t *testing.T) {
	_, err := Parse([]byte(`<foo xmlns="http://www.portalfiscal.inf.br/nfe"></foo>`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedRoot)

	var rootErr *UnsupportedRootError
	require.ErrorAs(t, err, &rootErr)
	require.Equal(t, NFe, rootErr.Family)
	require.Equal(t, "foo", rootErr.Root)
}

func TestParseRejectsEmptyDocument(t *testing.T) {
	_, err := Parse(nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrEmptyDocument)
}

func TestLifecycleEventInfoViaInfo(t *testing.T) {
	data, err := os.ReadFile("testdata/cte/v4_0/cancel35150107565416000104570000000012301000012300-ped-eve.xml")
	require.NoError(t, err)

	doc, err := Parse(data)
	require.NoError(t, err)

	event, ok := doc.Info().(LifecycleEventInfo)
	require.True(t, ok)
	require.Equal(t, "110111", event.GetEventType())
	require.Equal(t, "1", event.GetEventSequence())
}

func TestNFSeLifecycleEventFactsViaInfo(t *testing.T) {
	data, err := os.ReadFile("testdata/nfse/v1_0/CancelarNFSe-ped-cannfse.xml")
	require.NoError(t, err)

	doc, err := Parse(data)
	require.NoError(t, err)

	facts, ok := doc.Info().(LifecycleEventFactsInfo)
	require.True(t, ok)
	require.Equal(t, &LifecycleEventFacts{
		RegistrationState: LifecycleEventRegistrationStateRequest,
		Type:              "e101101",
		RequestNumber:     "001",
		IssueDate:         "2022-09-28T13:50:29-03:00",
	}, facts.GetLifecycleEventFacts())
}

func TestNFSeSubstituicaoViaRelatedDocuments(t *testing.T) {
	data, err := os.ReadFile("testdata/nfse/v1_0/nfse-prod-substituicao.xml")
	require.NoError(t, err)

	doc, err := Parse(data)
	require.NoError(t, err)

	related, ok := doc.Info().(RelatedDocumentsInfo)
	require.True(t, ok)
	require.Contains(t, related.GetRelatedDocuments(), RelatedDocument{
		Type:      "nfse",
		AccessKey: "33045572214043710000103000000000001726012751532718",
	})
}
