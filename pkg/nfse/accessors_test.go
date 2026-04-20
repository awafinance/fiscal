package nfse_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfse"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-simples.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "140015920176113500013200900000000000000006", doc.GetAccessKey())
	require.Equal(t, "1.00", doc.GetVersion())
	require.Equal(t, "2", doc.GetEnvironment())
	require.Equal(t, "6", doc.GetNumber())
	require.Equal(t, "900", doc.GetSeries())
	require.Empty(t, doc.GetModel())
	require.Equal(t, "2022-09-28T13:50:29-03:00", doc.GetIssueDate())
	require.Equal(t, "999999999.99", doc.GetAmount())
	require.Equal(t, "01761135000132", doc.GetIssuerDocument())
	require.Empty(t, doc.GetRecipient())
	require.Empty(t, doc.GetStatusCode())
	require.False(t, doc.IsAuthorized())
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "service", Value: "999999999.99"})
	require.Contains(t, doc.GetParties(), info.Party{Role: "provider", Document: "01761135000132"})
}

func TestDocumentConvenienceAccessorsHandleIssuedNFSe(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/ConsultarNFSeEnvio-ped-sitnfse.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "14001591201761135000132000000000000022097781063609", doc.GetAccessKey())
	require.Equal(t, "2", doc.GetNumber())
	require.Equal(t, "989999961.04", doc.GetAmount())
	require.Equal(t, "LW SOFTWARES LTDA", doc.GetIssuer())
	require.Equal(t, "01761135000132", doc.GetIssuerDocument())
	require.Equal(t, "100", doc.GetStatusCode())
	require.True(t, doc.IsAuthorized())
}

func TestDocumentGetEmitterDetailIssuedNFSe(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/ConsultarNFSeEnvio-ped-sitnfse.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	detail := doc.GetEmitterDetail()
	require.NotNil(t, detail)
	require.Equal(t, "01761135000132", detail.IM)
	require.Empty(t, detail.IE)
	require.Empty(t, detail.CRT)
	require.Empty(t, detail.CNAE)

	require.NotNil(t, detail.Address)
	require.Equal(t, "RUA A", detail.Address.Street)
	require.Equal(t, "10", detail.Address.Number)
	require.Equal(t, "CENTRO", detail.Address.Neighborhood)
	require.Equal(t, "1400159", detail.Address.CityCode)
	require.Equal(t, "RR", detail.Address.State)
	require.Equal(t, "69380000", detail.Address.ZipCode)
}

func TestDocumentGetEmitterDetailHandlesNilDocument(t *testing.T) {
	var doc *nfse.Document
	require.Nil(t, doc.GetEmitterDetail())

	require.Nil(t, (&nfse.Document{}).GetEmitterDetail())
}

func TestDocumentGetEmitterDetailDPS(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-regime-normal.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	detail := doc.GetEmitterDetail()
	require.NotNil(t, detail)
	require.Equal(t, "152422", detail.IM)
	require.Empty(t, detail.IE)
	require.Nil(t, detail.Address)
}

func TestDocumentGetEmitterDetailDPSWithAddress(t *testing.T) {
	data := []byte(`<DPS versao="1.00" xmlns="http://www.sped.fazenda.gov.br/nfse">
		<infDPS Id="DPS420240420000000000000000007000000000000099">
			<tpAmb>1</tpAmb>
			<dhEmi>2023-09-09T09:42:06-03:00</dhEmi>
			<verAplic>20220719</verAplic>
			<serie>00007</serie>
			<nDPS>99</nDPS>
			<dCompet>2023-09-09</dCompet>
			<tpEmit>1</tpEmit>
			<cLocEmi>4202404</cLocEmi>
			<prest>
				<CNPJ>00000000000000</CNPJ>
				<IM>152422</IM>
				<end>
					<endNac><cMun>4202404</cMun><CEP>88000000</CEP></endNac>
					<xLgr>RUA CENTRAL</xLgr>
					<nro>42</nro>
					<xBairro>CENTRO</xBairro>
				</end>
				<fone>48999990000</fone>
				<email>prest@example.com</email>
				<regTrib><opSimpNac>2</opSimpNac><regEspTrib>0</regEspTrib></regTrib>
			</prest>
			<serv><locPrest><cLocPrestacao>4202404</cLocPrestacao></locPrest><cServ><cTribNac>010101</cTribNac><xDescServ>Servico</xDescServ></cServ></serv>
			<valores><vServPrest><vServ>100.00</vServ></vServPrest><trib><tribMun><tribISSQN>1</tribISSQN><tpRetISSQN>1</tpRetISSQN></tribMun><totTrib><indTotTrib>0</indTotTrib></totTrib></trib></valores>
		</infDPS>
	</DPS>`)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	detail := doc.GetEmitterDetail()
	require.NotNil(t, detail)
	require.Equal(t, "152422", detail.IM)
	require.Equal(t, "48999990000", detail.Phone)
	require.Equal(t, "prest@example.com", detail.Email)

	require.NotNil(t, detail.Address)
	require.Equal(t, "RUA CENTRAL", detail.Address.Street)
	require.Equal(t, "42", detail.Address.Number)
	require.Equal(t, "CENTRO", detail.Address.Neighborhood)
	require.Equal(t, "4202404", detail.Address.CityCode)
	require.Equal(t, "88000000", detail.Address.ZipCode)
}

func TestDocumentGetAmountsIncludesRetentions(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nfse/v1_0/dps-regime-normal.xml")
	require.NoError(t, err)

	doc, err := nfse.Parse(data)
	require.NoError(t, err)

	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "retained_inss", Value: "0.40"})
}
