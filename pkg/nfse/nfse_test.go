package nfse_test

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/awafinance/fiscal/pkg/fiscalerr"
	"github.com/awafinance/fiscal/pkg/nfse"
	"github.com/stretchr/testify/require"
)

const (
	nfseNamespace = "http://www.sped.fazenda.gov.br/nfse"
	dsNamespace   = "http://www.w3.org/2000/09/xmldsig#"
)

func TestParse_Fixtures(t *testing.T) {
	t.Parallel()

	for _, fixture := range allFixtureNames(t) {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			original := readFixture(t, fixture)
			doc := parseFixture(t, fixture)

			assertFixtureShape(t, fixture, doc)

			roundTripped, err := xml.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			require.Equal(t, normalizeXML(t, original), normalizeXML(t, roundTripped))

			reparsed, err := nfse.Parse(roundTripped)
			require.NoError(t, err)
			assertSameRoot(t, doc, reparsed)
			assertFixtureShape(t, fixture, reparsed)
		})
	}
}

func TestParse_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		data        []byte
		errContains string
		errIs       error
	}{
		{name: "empty", data: nil, errIs: fiscalerr.ErrEmptyDocument},
		{name: "unsupported root", data: []byte(`<foo></foo>`), errIs: fiscalerr.ErrUnsupportedRoot},
		{name: "invalid dps", data: []byte(`<DPS xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.00"></DPS>`), errContains: "missing infDPS"},
		{name: "invalid nfse", data: []byte(`<NFSe xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.00"></NFSe>`), errContains: "missing infNFSe"},
		{name: "invalid event", data: []byte(`<pedRegEvento xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.00"></pedRegEvento>`), errContains: "missing infPedReg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := nfse.Parse(tt.data)
			require.Error(t, err)
			require.Nil(t, doc)
			if tt.errIs != nil {
				require.ErrorIs(t, err, tt.errIs)
			} else {
				require.ErrorContains(t, err, tt.errContains)
			}
		})
	}
}

func TestMarshalXML_NilReceiver(t *testing.T) {
	t.Parallel()

	var doc *nfse.Document
	data, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Empty(t, data)
}

func TestParse_SignedFixture(t *testing.T) {
	t.Parallel()

	doc := parseFixture(t, "ConsultarNFSeEnvio-ped-sitnfse.xml")

	require.NotNil(t, doc.NFSe)
	require.NotNil(t, doc.NFSe.DsSignature)
	require.NotNil(t, doc.NFSe.DsSignature.SignedInfo)
	require.Equal(t, "#NFS14001591201761135000132000000000000022097781063609", doc.NFSe.DsSignature.SignedInfo.Reference.URIAttr)
	require.Equal(t, "http://www.w3.org/2000/09/xmldsig#sha1", doc.NFSe.DsSignature.SignedInfo.Reference.DigestMethod.AlgorithmAttr)
	require.NotEmpty(t, doc.NFSe.DsSignature.SignatureValue.Value)
	require.NotEmpty(t, doc.NFSe.DsSignature.KeyInfo.X509Data.X509Certificate)

	require.NotNil(t, doc.NFSe.InfNFSe)
	require.NotNil(t, doc.NFSe.InfNFSe.DPS)
	require.NotNil(t, doc.NFSe.InfNFSe.DPS.DsSignature)
	require.NotNil(t, doc.NFSe.InfNFSe.DPS.DsSignature.SignedInfo)
	require.Equal(t, "#DPS140015920176113500013200900000000000000003", doc.NFSe.InfNFSe.DPS.DsSignature.SignedInfo.Reference.URIAttr)
	require.Equal(t, "http://www.w3.org/2000/09/xmldsig#rsa-sha1", doc.NFSe.InfNFSe.DPS.DsSignature.SignedInfo.SignatureMethod.AlgorithmAttr)
	require.Len(t, doc.NFSe.InfNFSe.DPS.DsSignature.SignedInfo.Reference.Transforms.Transform, 2)
	require.NotEmpty(t, doc.NFSe.InfNFSe.DPS.DsSignature.SignatureValue.Value)
	require.NotEmpty(t, doc.NFSe.InfNFSe.DPS.DsSignature.KeyInfo.X509Data.X509Certificate)
}

func assertFixtureShape(t *testing.T, fixture string, doc *nfse.Document) {
	t.Helper()

	switch fixture {
	case "dps-regime-normal.xml":
		require.NotNil(t, doc.DPS)
		require.Equal(t, "1.00", doc.DPS.VersaoAttr)
		require.Equal(t, "00007", doc.DPS.InfDPS.Serie)
		require.Equal(t, "2", doc.DPS.InfDPS.NDPS)
		require.Equal(t, "4202404", doc.DPS.InfDPS.CLocEmi)
		require.Equal(t, "010101", doc.DPS.InfDPS.Serv.CServ.CTribNac)
	case "dps-simples.xml", "ConsultarNFSeRPS-ped-sitnfserps.xml":
		require.NotNil(t, doc.DPS)
		require.Equal(t, "900", doc.DPS.InfDPS.Serie)
		require.Equal(t, "6", doc.DPS.InfDPS.NDPS)
		require.NotNil(t, doc.DPS.InfDPS.Prest.RegTrib)
	case "dps-cpf-taker-piscofins.xml":
		require.NotNil(t, doc.DPS)
		require.Equal(t, "1.00", doc.DPS.VersaoAttr)
		require.Equal(t, "31", doc.DPS.InfDPS.NDPS)
		require.Equal(t, "2025-12-04", doc.DPS.InfDPS.DCompet)
		require.NotNil(t, doc.DPS.InfDPS.Toma)
		require.NotNil(t, doc.DPS.InfDPS.Toma.CPF)
		require.Equal(t, "98216457200", *doc.DPS.InfDPS.Toma.CPF)
	case "ConsultarNFSeEnvio-ped-sitnfse.xml":
		require.NotNil(t, doc.NFSe)
		require.Equal(t, "1.00", doc.NFSe.VersaoAttr)
		require.Equal(t, "2", doc.NFSe.InfNFSe.NNFSe)
		require.Equal(t, "100", doc.NFSe.InfNFSe.CStat)
		require.NotNil(t, doc.NFSe.InfNFSe.DPS)
	case "nfse-prod-iss-retido-cooperativa.xml":
		require.NotNil(t, doc.NFSe)
		require.Equal(t, "76", doc.NFSe.InfNFSe.NNFSe)
		require.Equal(t, "100", doc.NFSe.InfNFSe.CStat)
		require.Equal(t, "2026-02-03", doc.NFSe.InfNFSe.DPS.InfDPS.DCompet)
		require.Equal(t, "2", doc.NFSe.InfNFSe.DPS.InfDPS.Valores.Trib.TribMun.TpRetISSQN)
	case "nfse-prod-iss-retido-piscofins.xml":
		require.NotNil(t, doc.NFSe)
		require.Equal(t, "99", doc.NFSe.InfNFSe.NNFSe)
		require.Equal(t, "100", doc.NFSe.InfNFSe.CStat)
		require.Equal(t, "2", doc.NFSe.InfNFSe.DPS.InfDPS.Valores.Trib.TribMun.TpRetISSQN)
		require.NotNil(t, doc.NFSe.InfNFSe.DPS.InfDPS.Valores.Trib.TribFed.Piscofins)
	case "nfse-prod-substituicao.xml":
		require.NotNil(t, doc.NFSe)
		require.Equal(t, "18", doc.NFSe.InfNFSe.NNFSe)
		require.Equal(t, "101", doc.NFSe.InfNFSe.CStat)
		require.NotNil(t, doc.NFSe.InfNFSe.DPS.InfDPS.Subst)
		require.Equal(t, "99", doc.NFSe.InfNFSe.DPS.InfDPS.Subst.CMotivo)
	case "CancelarNFSe-ped-cannfse.xml":
		require.NotNil(t, doc.PedRegEvento)
		require.Equal(t, "1.00", doc.PedRegEvento.VersaoAttr)
		require.Equal(t, "001", doc.PedRegEvento.InfPedReg.NPedRegEvento)
		require.NotNil(t, doc.PedRegEvento.InfPedReg.E101101)
		require.Equal(t, "1", doc.PedRegEvento.InfPedReg.E101101.CMotivo)
	default:
		t.Fatalf("unhandled fixture %s", fixture)
	}
}

func assertSameRoot(t *testing.T, expected, actual *nfse.Document) {
	t.Helper()

	require.Equal(t, expected.DPS != nil, actual.DPS != nil)
	require.Equal(t, expected.NFSe != nil, actual.NFSe != nil)
	require.Equal(t, expected.PedRegEvento != nil, actual.PedRegEvento != nil)
}

func allFixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "..", "testdata", "nfse", "v1_0"))
	require.NoError(t, err)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xml" {
			continue
		}
		names = append(names, entry.Name())
	}

	slices.Sort(names)
	return names
}

func parseFixture(t *testing.T, name string) *nfse.Document {
	t.Helper()

	data := readFixture(t, name)
	doc, err := nfse.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "nfse", "v1_0", name))
	require.NoError(t, err)
	return data
}

func normalizeXML(t *testing.T, data []byte) string {
	t.Helper()

	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var b strings.Builder
	nsStack := []map[string]string{{}}

	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			b.WriteByte('<')
			b.WriteString(qualifiedName(tok.Name))
			currentNS := make(map[string]string, len(nsStack[len(nsStack)-1]))
			for prefix, value := range nsStack[len(nsStack)-1] {
				currentNS[prefix] = value
			}
			attrs := make([]xml.Attr, 0, len(tok.Attr))
			for _, attr := range tok.Attr {
				if isNamespaceDecl(attr) {
					prefix := attr.Name.Local
					if attr.Name.Local == "xmlns" {
						prefix = ""
					}
					value := strings.TrimSpace(attr.Value)
					if currentNS[prefix] == value {
						continue
					}
					currentNS[prefix] = value
					continue
				}
				attrs = append(attrs, attr)
			}
			slices.SortFunc(attrs, func(a, b xml.Attr) int {
				return cmp.Or(
					strings.Compare(a.Name.Space, b.Name.Space),
					strings.Compare(a.Name.Local, b.Name.Local),
				)
			})
			for _, attr := range attrs {
				b.WriteByte(' ')
				b.WriteString(qualifiedName(attr.Name))
				b.WriteString(`="`)
				b.WriteString(strings.TrimSpace(attr.Value))
				b.WriteByte('"')
			}
			b.WriteByte('>')
			nsStack = append(nsStack, currentNS)
		case xml.EndElement:
			b.WriteString("</")
			b.WriteString(qualifiedName(tok.Name))
			b.WriteByte('>')
			if len(nsStack) > 1 {
				nsStack = nsStack[:len(nsStack)-1]
			}
		case xml.CharData:
			text := strings.TrimSpace(string(tok))
			if text != "" {
				b.WriteString(text)
			}
		}
	}

	return b.String()
}

func qualifiedName(name xml.Name) string {
	switch name.Space {
	case "", nfseNamespace:
		return name.Local
	case dsNamespace:
		return "ds:" + name.Local
	case "xmlns":
		if name.Local == "" {
			return "xmlns"
		}
		return "xmlns:" + name.Local
	}

	return name.Space + ":" + name.Local
}

func isNamespaceDecl(attr xml.Attr) bool {
	return attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns"
}
