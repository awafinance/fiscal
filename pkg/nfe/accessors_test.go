package nfe_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfe"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "42220575277525000178550030000292481295366801-procNFe.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "42220575277525000178550030000292481295366801", doc.GetAccessKey())
	require.Equal(t, "4.00", doc.GetVersion())
	require.Equal(t, "1", doc.GetEnvironment())
	require.Equal(t, "29248", doc.GetNumber())
	require.Equal(t, "3", doc.GetSeries())
	require.Equal(t, "55", doc.GetModel())
	require.Equal(t, "2022-05-27T08:52:04-03:00", doc.GetIssueDate())
	require.Equal(t, "64237.04", doc.GetAmount())
	require.Equal(t, "FORNOS LTDA", doc.GetIssuer())
	require.Equal(t, "75277525000178", doc.GetIssuerDocument())
	require.Equal(t, "Jung Usa Corporation", doc.GetRecipient())
	require.Equal(t, "371780142", doc.GetRecipientDocument())
	require.Equal(t, "342220106391922", doc.GetProtocolNumber())
	require.Equal(t, "100", doc.GetStatusCode())
	require.Equal(t, "Autorizado o uso da NF-e", doc.GetStatusReason())
	require.True(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "total", Value: "64237.04"})
	require.Contains(t, doc.GetParties(), info.Party{Role: "issuer", Name: "FORNOS LTDA", Document: "75277525000178"})
}

func TestDocumentGetItems(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "42220575277525000178550030000292481295366801-procNFe.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	items := doc.GetItems()
	require.Len(t, items, 2)
	require.Equal(t, nfe.Item{
		Number:      "1",
		Code:        "TB2001210",
		EAN:         "SEM GTIN",
		Description: "[TB2001210] FORNO LINHA INDUSTRIAL MODELO TB2001210",
		NCM:         "85141900",
		CFOP:        "7101",
		Unit:        "UN",
		Quantity:    "1.0000",
		UnitAmount:  "23942.1312000000",
		Amount:      "23942.13",
	}, items[0])
}

func TestDocumentGetPayments(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfe", "42220575277525000178550030000292481295366801-procNFe.xml"))
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, []nfe.Payment{
		{Method: "90", Amount: "0.00"},
	}, doc.GetPayments())
}

func TestDocumentConvenienceAccessorsHandleResNFe(t *testing.T) {
	data := []byte(`<resNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.01">
		<chNFe>35180803102452000172550010000476051695511860</chNFe>
		<CNPJ>12345678000195</CNPJ>
		<xNome>Emitente Teste</xNome>
		<IE>ISENTO</IE>
		<dhEmi>2024-01-02T03:04:05-03:00</dhEmi>
		<tpNF>1</tpNF>
		<vNF>100.00</vNF>
		<dhRecbto>2024-01-02T03:05:06-03:00</dhRecbto>
		<nProt>123456789012345</nProt>
		<cSitNFe>1</cSitNFe>
	</resNFe>`)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "35180803102452000172550010000476051695511860", doc.GetAccessKey())
	require.Equal(t, "100.00", doc.GetAmount())
	require.Equal(t, "Emitente Teste", doc.GetIssuer())
	require.Equal(t, "12345678000195", doc.GetIssuerDocument())
	require.Equal(t, "2024-01-02T03:04:05-03:00", doc.GetIssueDate())
	require.Equal(t, "123456789012345", doc.GetProtocolNumber())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetStatusCode())
	require.False(t, doc.IsAuthorized())
}

func TestDocumentConvenienceAccessorsHandleNilDocument(t *testing.T) {
	var doc *nfe.Document

	require.Empty(t, doc.GetAccessKey())
	require.Empty(t, doc.GetNumber())
	require.Empty(t, doc.GetAmount())
	require.Empty(t, doc.GetIssuer())
	require.Empty(t, doc.GetIssuerDocument())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetRecipientDocument())
	require.Empty(t, doc.GetProtocolNumber())
	require.Empty(t, doc.GetStatusCode())
	require.Empty(t, doc.GetStatusReason())
	require.False(t, doc.IsAuthorized())
	require.Nil(t, doc.GetItems())
	require.Nil(t, doc.GetPayments())
}
