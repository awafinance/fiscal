package nfse_test

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nfseManifestacaoChave = "35489062205687512000104000000000019226013986396200"
	nfseManifestacaoCNPJ  = "05687512000104"
)

func nfseManifestacaoTime() time.Time {
	return time.Date(2026, 2, 4, 10, 56, 12, 0, time.FixedZone("BRT", -3*3600))
}

func TestBuildTomadorManifestacaoConfirmacao(t *testing.T) {
	t.Parallel()

	evento, err := nfse.BuildTomadorManifestacao(nfse.TomadorManifestacaoInput{
		TpEvento:  nfse.TomadorManifestacaoConfirmacao,
		ChNFSe:    nfseManifestacaoChave,
		CNPJAutor: nfseManifestacaoCNPJ,
		DhEvento:  nfseManifestacaoTime(),
		TpAmb:     "2",
		VerAplic:  "nyx-test",
	})
	require.NoError(t, err)
	require.NotNil(t, evento.InfPedReg)

	inf := evento.InfPedReg
	assert.Equal(t, "1.01", evento.VersaoAttr)
	assert.Equal(t, "PRE"+nfseManifestacaoChave+"203202", inf.IdAttr)
	assert.Equal(t, "2", inf.TpAmb)
	assert.Equal(t, "nyx-test", inf.VerAplic)
	assert.Equal(t, "2026-02-04T10:56:12-03:00", inf.DhEvento)
	require.NotNil(t, inf.CNPJAutor)
	assert.Equal(t, nfseManifestacaoCNPJ, *inf.CNPJAutor)
	assert.Nil(t, inf.CPFAutor)
	assert.Equal(t, nfseManifestacaoChave, inf.ChNFSe)
	require.NotNil(t, inf.E203202)
	assert.Equal(t, "Manifestação de NFS-e - Confirmação do Tomador", inf.E203202.XDesc)

	data, err := nfse.MarshalTomadorManifestacao(evento)
	require.NoError(t, err)
	xmlText := string(data)
	assert.Contains(t, xmlText, `<pedRegEvento xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.01">`)
	assert.Contains(t, xmlText, `Id="PRE`+nfseManifestacaoChave+`203202"`)
	assert.Contains(t, xmlText, "<e203202>")
	assert.NotContains(t, xmlText, "nPedRegEvento")

	doc, err := nfse.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "pedRegEvento", doc.RootName)
	assert.Equal(t, "e203202", doc.GetEventType())
	assert.Equal(t, nfseManifestacaoChave, doc.GetAccessKey())
	assert.Equal(t, "PRE"+nfseManifestacaoChave+"203202", doc.GetRequestEventID())
	assert.Equal(t, &info.LifecycleEventFacts{
		RegistrationState: info.LifecycleEventRegistrationStateRequest,
		Type:              "e203202",
		IssueDate:         "2026-02-04T10:56:12-03:00",
	}, doc.GetLifecycleEventFacts())
}

func TestBuildTomadorManifestacaoRejeicao(t *testing.T) {
	t.Parallel()

	evento, err := nfse.BuildTomadorManifestacao(nfse.TomadorManifestacaoInput{
		TpEvento:      nfse.TomadorManifestacaoRejeicao,
		ChNFSe:        nfseManifestacaoChave,
		CNPJAutor:     nfseManifestacaoCNPJ,
		DhEvento:      nfseManifestacaoTime(),
		TpAmb:         "1",
		VerAplic:      "nyx-test",
		Motivo:        "9",
		Justificativa: "Servico nao reconhecido pelo tomador",
	})
	require.NoError(t, err)
	require.NotNil(t, evento.InfPedReg)
	require.NotNil(t, evento.InfPedReg.E203206)

	rej := evento.InfPedReg.E203206
	assert.Equal(t, "Manifestação de NFS-e - Rejeição do Tomador", rej.XDesc)
	assert.Equal(t, "9", rej.CMotivo)
	require.NotNil(t, rej.XMotivo)
	assert.Equal(t, "Servico nao reconhecido pelo tomador", *rej.XMotivo)
	assert.Equal(t, "PRE"+nfseManifestacaoChave+"203206", evento.InfPedReg.IdAttr)

	data, err := nfse.MarshalTomadorManifestacao(evento)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<cMotivo>9</cMotivo>")
	assert.Contains(t, string(data), "<xMotivo>Servico nao reconhecido pelo tomador</xMotivo>")
	assert.NotContains(t, string(data), "nPedRegEvento")

	doc, err := nfse.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "e203206", doc.GetEventType())
	assert.Equal(t, "PRE"+nfseManifestacaoChave+"203206", doc.GetRequestEventID())
}

func TestBuildTomadorManifestacaoValidation(t *testing.T) {
	t.Parallel()

	base := nfse.TomadorManifestacaoInput{
		TpEvento:  nfse.TomadorManifestacaoConfirmacao,
		ChNFSe:    nfseManifestacaoChave,
		CNPJAutor: nfseManifestacaoCNPJ,
		DhEvento:  nfseManifestacaoTime(),
		TpAmb:     "2",
		VerAplic:  "nyx-test",
	}

	tests := []struct {
		name    string
		mutate  func(*nfse.TomadorManifestacaoInput)
		wantErr error
	}{
		{"unsupported event", func(in *nfse.TomadorManifestacaoInput) { in.TpEvento = "e205208" }, nfse.ErrUnsupportedTomadorManifestacao},
		{"short key", func(in *nfse.TomadorManifestacaoInput) { in.ChNFSe = "123" }, nfse.ErrInvalidNFSeAccessKey},
		{"non numeric key", func(in *nfse.TomadorManifestacaoInput) { in.ChNFSe = strings.Repeat("x", 50) }, nfse.ErrInvalidNFSeAccessKey},
		{"missing author", func(in *nfse.TomadorManifestacaoInput) { in.CNPJAutor = "" }, nfse.ErrInvalidNFSeAuthor},
		{"two authors", func(in *nfse.TomadorManifestacaoInput) { in.CPFAutor = "12345678901" }, nfse.ErrInvalidNFSeAuthor},
		{"bad cnpj", func(in *nfse.TomadorManifestacaoInput) { in.CNPJAutor = "123" }, nfse.ErrInvalidNFSeAuthor},
		{"cpf author", func(in *nfse.TomadorManifestacaoInput) {
			in.CNPJAutor = ""
			in.CPFAutor = "12345678901"
		}, nil},
		{"bad environment", func(in *nfse.TomadorManifestacaoInput) { in.TpAmb = "3" }, nfse.ErrInvalidNFSeEnvironment},
		{"zero event time", func(in *nfse.TomadorManifestacaoInput) { in.DhEvento = time.Time{} }, nfse.ErrInvalidNFSeEventTime},
		{"missing app version", func(in *nfse.TomadorManifestacaoInput) { in.VerAplic = "" }, nfse.ErrInvalidNFSeApplicationVersion},
		{"motivo on confirmacao", func(in *nfse.TomadorManifestacaoInput) { in.Motivo = "1" }, nfse.ErrInvalidTomadorRejeicaoMotivo},
		{"bad rejeicao motivo", func(in *nfse.TomadorManifestacaoInput) {
			in.TpEvento = nfse.TomadorManifestacaoRejeicao
			in.Motivo = "8"
		}, nfse.ErrInvalidTomadorRejeicaoMotivo},
		{"motivo 9 missing justificativa", func(in *nfse.TomadorManifestacaoInput) {
			in.TpEvento = nfse.TomadorManifestacaoRejeicao
			in.Motivo = "9"
		}, nfse.ErrInvalidTomadorRejeicaoXMotivo},
		{"short justificativa", func(in *nfse.TomadorManifestacaoInput) {
			in.TpEvento = nfse.TomadorManifestacaoRejeicao
			in.Motivo = "1"
			in.Justificativa = "curta"
		}, nfse.ErrInvalidTomadorRejeicaoXMotivo},
		{"unsupported justificativa char", func(in *nfse.TomadorManifestacaoInput) {
			in.TpEvento = nfse.TomadorManifestacaoRejeicao
			in.Motivo = "1"
			in.Justificativa = "Servico nao reconhecido 😀"
		}, nfse.ErrInvalidTomadorRejeicaoXMotivo},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := base
			tc.mutate(&in)
			_, err := nfse.BuildTomadorManifestacao(in)
			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestParseV101RegisteredEvento(t *testing.T) {
	t.Parallel()

	xmlText := `<evento xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.01">` +
		`<infEvento Id="EVT` + nfseManifestacaoChave + `203206001">` +
		`<verAplic>SefinNacional_1.6.0</verAplic><ambGer>2</ambGer><nSeqEvento>001</nSeqEvento>` +
		`<dhProc>2026-02-04T10:57:12-03:00</dhProc><nDFSe>123</nDFSe>` +
		`<pedRegEvento versao="1.01"><infPedReg Id="PRE` + nfseManifestacaoChave + `203206">` +
		`<tpAmb>2</tpAmb><verAplic>nyx-test</verAplic><dhEvento>2026-02-04T10:56:12-03:00</dhEvento>` +
		`<CNPJAutor>` + nfseManifestacaoCNPJ + `</CNPJAutor><chNFSe>` + nfseManifestacaoChave + `</chNFSe>` +
		`<e203206><xDesc>Manifestação de NFS-e - Rejeição do Tomador</xDesc><cMotivo>1</cMotivo></e203206>` +
		`</infPedReg></pedRegEvento>` +
		`</infEvento><Signature xmlns="http://www.w3.org/2000/09/xmldsig#"></Signature></evento>`

	doc, err := nfse.Parse([]byte(xmlText))
	require.NoError(t, err)
	assert.Equal(t, "evento", doc.RootName)
	assert.Equal(t, "e203206", doc.GetEventType())
	assert.Equal(t, "001", doc.GetEventSequence())
	assert.Equal(t, "EVT"+nfseManifestacaoChave+"203206001", doc.GetEventID())
	assert.Equal(t, "PRE"+nfseManifestacaoChave+"203206", doc.GetRequestEventID())
	assert.Equal(t, &info.LifecycleEventFacts{
		RegistrationState: info.LifecycleEventRegistrationStateRegistered,
		Type:              "e203206",
		Sequence:          "001",
		IssueDate:         "2026-02-04T10:56:12-03:00",
		ProcessingTime:    "2026-02-04T10:57:12-03:00",
	}, doc.GetLifecycleEventFacts())

	out, err := xml.Marshal(doc)
	require.NoError(t, err)
	require.Contains(t, string(out), "<evento")
	reparsed, err := nfse.Parse(out)
	require.NoError(t, err)
	assert.Equal(t, "e203206", reparsed.GetEventType())
	assert.Equal(t, "001", reparsed.GetEventSequence())
}

func TestMarshalTomadorManifestacaoRejectsInvalidEvent(t *testing.T) {
	t.Parallel()

	_, err := nfse.MarshalTomadorManifestacao(nil)
	require.True(t, errors.Is(err, nfse.ErrInvalidTomadorManifestacaoEvento))
}
