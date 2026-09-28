package fiscal

import (
	"os"
	"strings"
	"testing"

	"github.com/awafinance/fiscal/pkg/nfse"
	"github.com/stretchr/testify/require"
)

func TestParseSaoPauloNFSe(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/nfse/sao_paulo.xml")
	require.NoError(t, err)
	doc, err := Parse(data)
	require.NoError(t, err)
	require.Equal(t, NFSe, doc.Family)
	require.Equal(t, "NFe", doc.RootName)
	info := doc.Info()
	require.Equal(t, "35503081232223020000118000000565182326095180335000", info.GetAccessKey())
	require.Equal(t, "5651823", info.GetNumber())
	require.Empty(t, info.GetSeries(), "the RPS series is not the invoice series")
	require.Equal(t, "2026-09-24T14:00:16", info.GetIssueDate())
	require.Equal(t, "32223020000118", info.GetIssuerDocument())
	require.Equal(t, "29478269000321", info.GetRecipientDocument())
	require.Equal(t, "FLASH TECNOLOGIA E INSTITUICAO DE PAGAMENTO LTDA", info.GetIssuer())
	require.Equal(t, "INDUSTRIA DE BEBIDAS TRES RIOS LTDA", info.GetRecipient())
	require.Equal(t, "N", info.GetStatusCode())
	require.True(t, info.IsAuthorized())
	require.Equal(t, "260", info.GetAmount())
	amounts, ok := info.(AmountsInfo)
	require.True(t, ok)
	require.Equal(t, []Amount{{Type: "service", Value: "260"}, {Type: "tax_iss", Value: "0"}}, amounts.GetAmounts())
	require.Equal(t, nfse.DeclaredAmounts{Service: "260", DeductionOrReduction: "260", ISS: "0"}, doc.NFSe.GetDeclaredAmounts())

	for _, status := range []string{"C", "unknown", ""} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			changed := strings.Replace(string(data), "<StatusNFe>N</StatusNFe>", "<StatusNFe>"+status+"</StatusNFe>", 1)
			parsed, parseErr := Parse([]byte(changed))
			require.NoError(t, parseErr)
			require.False(t, parsed.Info().IsAuthorized())
		})
	}
}

func TestParseRejectsUnidentifiedSaoPauloNFSe(t *testing.T) {
	t.Parallel()
	for _, input := range []string{`<NFe/>`, `<NFe><infNFe/></NFe>`, `<NFe><ChaveNFe><NumeroNFe>1</NumeroNFe></ChaveNFe></NFe>`} {
		_, err := Parse([]byte(input))
		require.Error(t, err)
	}
}

func TestSaoPauloDeclaredRetentions(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/nfse/sao_paulo.xml")
	require.NoError(t, err)
	// Synthetic tax values follow the municipal Web Service manual v3.3.6.
	input := strings.Replace(string(data), "</NFe>", `<ValorPIS>1</ValorPIS><ValorCOFINS>2</ValorCOFINS><ValorINSS>3</ValorINSS><ValorIR>4</ValorIR><ValorCSLL>5</ValorCSLL><RetencaoPisCofins>3</RetencaoPisCofins></NFe>`, 1)
	input = strings.Replace(input, "<ISSRetido>false</ISSRetido>", "<ISSRetido>true</ISSRetido>", 1)
	doc, err := Parse([]byte(input))
	require.NoError(t, err)
	amounts := doc.NFSe.GetDeclaredAmounts()
	require.Equal(t, "1", amounts.PIS)
	require.Equal(t, "2", amounts.COFINS)
	require.Equal(t, "3", amounts.INSSRetention)
	require.Equal(t, "4", amounts.IRRFRetention)
	require.Equal(t, "5", amounts.SocialContributionsRetention)
	require.Equal(t, "3", amounts.SocialContributionsRetentionCode)
	require.Equal(t, "taker", amounts.ISSWithheldBy)
}
