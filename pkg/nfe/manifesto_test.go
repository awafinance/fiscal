package nfe_test

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/awafinance/fiscal/pkg/nfe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	manifestoChave = "35180803102452000172550010000476051695511860"
	manifestoCNPJ  = "12345678000195"
)

func manifestoTime() time.Time {
	return time.Date(2024, 1, 2, 3, 4, 5, 0, time.FixedZone("BRT", -3*3600))
}

func TestBuildManifesto_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tpEvento string
		just     string
		wantDesc string
		wantJust string // "" means xJust must be nil
	}{
		{"ciencia", nfe.ManifestoCiencia, "", "Ciencia da Operacao", ""},
		{"confirmacao", nfe.ManifestoConfirmacao, "", "Confirmacao da Operacao", ""},
		{"desconhecimento", nfe.ManifestoDesconhecimento, "", "Desconhecimento da Operacao", ""},
		{"desconhecimento with justificativa", nfe.ManifestoDesconhecimento, "Nao reconheco esta operacao", "Desconhecimento da Operacao", "Nao reconheco esta operacao"},
		{"nao realizada", nfe.ManifestoOperacaoNaoRealizada, "Mercadoria nao recebida", "Operacao nao Realizada", "Mercadoria nao recebida"},
	}
	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ev, err := nfe.BuildManifesto(nfe.ManifestoInput{
				TpEvento:      tc.tpEvento,
				ChNFe:         manifestoChave,
				CNPJ:          manifestoCNPJ,
				NSeqEvento:    1,
				DhEvento:      manifestoTime(),
				TpAmb:         "1",
				Justificativa: tc.just,
			})
			require.NoError(t, err)
			require.NotNil(t, ev.InfEvento)

			inf := ev.InfEvento
			assert.Equal(t, "1.00", ev.VersaoAttr)
			assert.Equal(t, "ID"+tc.tpEvento+manifestoChave+"01", inf.IdAttr)
			assert.Equal(t, "91", inf.COrgao)
			assert.Equal(t, "1", inf.TpAmb)
			require.NotNil(t, inf.CNPJ)
			assert.Equal(t, manifestoCNPJ, *inf.CNPJ)
			assert.Equal(t, manifestoChave, inf.ChNFe)
			assert.Equal(t, "2024-01-02T03:04:05-03:00", inf.DhEvento)
			assert.Equal(t, tc.tpEvento, inf.TpEvento)
			assert.Equal(t, "1", inf.NSeqEvento)
			assert.Equal(t, "1.00", inf.VerEvento)
			require.NotNil(t, inf.DetEvento)
			assert.Equal(t, tc.wantDesc, inf.DetEvento.DescEvento)
			if tc.wantJust == "" {
				assert.Nil(t, inf.DetEvento.XJust)
			} else {
				require.NotNil(t, inf.DetEvento.XJust)
				assert.Equal(t, tc.wantJust, *inf.DetEvento.XJust)
			}
		})
	}
}

func TestBuildManifesto_Validation(t *testing.T) {
	t.Parallel()

	base := nfe.ManifestoInput{
		TpEvento:   nfe.ManifestoCiencia,
		ChNFe:      manifestoChave,
		CNPJ:       manifestoCNPJ,
		NSeqEvento: 1,
		DhEvento:   manifestoTime(),
		TpAmb:      "1",
	}

	tests := []struct {
		name    string
		mutate  func(*nfe.ManifestoInput)
		wantErr error
	}{
		{"unsupported event", func(in *nfe.ManifestoInput) { in.TpEvento = "110111" }, nfe.ErrUnsupportedManifesto},
		{"short access key", func(in *nfe.ManifestoInput) { in.ChNFe = "123" }, nfe.ErrInvalidAccessKey},
		{"non-numeric access key", func(in *nfe.ManifestoInput) { in.ChNFe = strings.Repeat("x", 44) }, nfe.ErrInvalidAccessKey},
		{"bad cnpj", func(in *nfe.ManifestoInput) { in.CNPJ = "123" }, nfe.ErrInvalidCNPJ},
		{"sequence zero", func(in *nfe.ManifestoInput) { in.NSeqEvento = 0 }, nfe.ErrInvalidSequence},
		{"sequence too high", func(in *nfe.ManifestoInput) { in.NSeqEvento = 21 }, nfe.ErrInvalidSequence},
		{"bad environment", func(in *nfe.ManifestoInput) { in.TpAmb = "3" }, nfe.ErrInvalidEnvironment},
		{"zero event time", func(in *nfe.ManifestoInput) { in.DhEvento = time.Time{} }, nfe.ErrInvalidEventTime},
		{"justificativa on event without xJust", func(in *nfe.ManifestoInput) { in.Justificativa = "should not be here" }, nfe.ErrJustificativa},
		{"justificativa missing on 240", func(in *nfe.ManifestoInput) { in.TpEvento = nfe.ManifestoOperacaoNaoRealizada }, nfe.ErrJustificativa},
		{"justificativa too short on 240", func(in *nfe.ManifestoInput) {
			in.TpEvento = nfe.ManifestoOperacaoNaoRealizada
			in.Justificativa = "curta"
		}, nfe.ErrJustificativa},
		{"justificativa too short on 220", func(in *nfe.ManifestoInput) {
			in.TpEvento = nfe.ManifestoDesconhecimento
			in.Justificativa = "curta"
		}, nfe.ErrJustificativa},
	}
	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := base
			tc.mutate(&in)
			_, err := nfe.BuildManifesto(in)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestBuildManifesto_FormatsUTCWithNumericOffset(t *testing.T) {
	t.Parallel()

	ev, err := nfe.BuildManifesto(nfe.ManifestoInput{
		TpEvento:   nfe.ManifestoCiencia,
		ChNFe:      manifestoChave,
		CNPJ:       manifestoCNPJ,
		NSeqEvento: 1,
		DhEvento:   time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		TpAmb:      "1",
	})
	require.NoError(t, err)
	require.NotNil(t, ev.InfEvento)
	assert.Equal(t, "2024-01-02T03:04:05+00:00", ev.InfEvento.DhEvento)
}

func TestMarshalManifestoEnvEvento(t *testing.T) {
	t.Parallel()

	just := "Mercadoria recusada na portaria"
	ev, err := nfe.BuildManifesto(nfe.ManifestoInput{
		TpEvento:      nfe.ManifestoOperacaoNaoRealizada,
		ChNFe:         manifestoChave,
		CNPJ:          manifestoCNPJ,
		NSeqEvento:    1,
		DhEvento:      manifestoTime(),
		TpAmb:         "1",
		Justificativa: just,
	})
	require.NoError(t, err)

	data, err := nfe.MarshalManifestoEnvEvento("1", ev)
	require.NoError(t, err)

	got := string(data)
	assert.Contains(t, got, `<envEvento xmlns="http://www.portalfiscal.inf.br/nfe" versao="1.00">`)
	assert.Contains(t, got, "<idLote>1</idLote>")
	assert.Contains(t, got, "<tpEvento>210240</tpEvento>")
	assert.Contains(t, got, "<descEvento>Operacao nao Realizada</descEvento>")
	assert.Contains(t, got, "<xJust>"+just+"</xJust>")
	assert.Contains(t, got, `Id="ID210240`+manifestoChave+`01"`)

	// The serialized lote parses back through the generic envelope path.
	doc, err := nfe.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc.EnvEvento)
	assert.Equal(t, "1", doc.EnvEvento.IdLote)
	require.Len(t, doc.EnvEvento.Evento, 1)
	assert.Equal(t, "210240", doc.GetEventType())
}

func TestMarshalManifestoEnvEvento_Validation(t *testing.T) {
	t.Parallel()

	ev, err := nfe.BuildManifesto(nfe.ManifestoInput{
		TpEvento:   nfe.ManifestoCiencia,
		ChNFe:      manifestoChave,
		CNPJ:       manifestoCNPJ,
		NSeqEvento: 1,
		DhEvento:   manifestoTime(),
		TpAmb:      "1",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		idLote  string
		eventos []*nfe.EventoMDETEvento
		wantErr error
	}{
		{"empty idLote", "", []*nfe.EventoMDETEvento{ev}, nfe.ErrInvalidLoteID},
		{"non-numeric idLote", "abc", []*nfe.EventoMDETEvento{ev}, nfe.ErrInvalidLoteID},
		{"too long idLote", strings.Repeat("1", 16), []*nfe.EventoMDETEvento{ev}, nfe.ErrInvalidLoteID},
		{"zero events", "1", nil, nfe.ErrInvalidEnvelope},
		{"nil event", "1", []*nfe.EventoMDETEvento{nil}, nfe.ErrInvalidEnvelope},
		{"too many events", "1", manifestoEvents(ev, 21), nfe.ErrInvalidEnvelope},
	}
	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := nfe.MarshalManifestoEnvEvento(tc.idLote, tc.eventos...)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// A single evento round-trips through the typed MDE root, proving detEvento
// (descEvento + xJust) survives marshal — unlike the generic envelope.
func TestBuildManifesto_RoundTripDetEvento(t *testing.T) {
	t.Parallel()

	just := "Operacao nao concretizada pelo destinatario"
	ev, err := nfe.BuildManifesto(nfe.ManifestoInput{
		TpEvento:      nfe.ManifestoOperacaoNaoRealizada,
		ChNFe:         manifestoChave,
		CNPJ:          manifestoCNPJ,
		NSeqEvento:    2,
		DhEvento:      manifestoTime(),
		TpAmb:         "2",
		Justificativa: just,
	})
	require.NoError(t, err)

	data, err := xml.Marshal(&nfe.Document{EventoMDE: ev})
	require.NoError(t, err)

	doc, err := nfe.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, doc.EventoMDE)
	require.NotNil(t, doc.EventoMDE.InfEvento.DetEvento)
	assert.Equal(t, "Operacao nao Realizada", doc.EventoMDE.InfEvento.DetEvento.DescEvento)
	require.NotNil(t, doc.EventoMDE.InfEvento.DetEvento.XJust)
	assert.Equal(t, just, *doc.EventoMDE.InfEvento.DetEvento.XJust)
	assert.Equal(t, "210240", doc.GetEventType())
	assert.Equal(t, "2", doc.GetEventSequence())
}

func manifestoEvents(ev *nfe.EventoMDETEvento, n int) []*nfe.EventoMDETEvento {
	events := make([]*nfe.EventoMDETEvento, n)
	for i := range events {
		events[i] = ev
	}
	return events
}
