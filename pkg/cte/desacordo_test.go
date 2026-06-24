package cte_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/awafinance/fiscal/pkg/cte"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	desacordoCNPJ          = "12345678000195"
	desacordoCOrgao        = "43"
	desacordoJustificativa = "Tomador discorda da prestacao informada"
	desacordoTargetProt    = "135260000000009"
)

func desacordoTime() time.Time {
	return time.Date(2024, 1, 2, 3, 4, 5, 0, time.FixedZone("BRT", -3*3600))
}

func TestBuildDesacordo_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    cte.DesacordoInput
		assertFn func(t *testing.T, doc *cte.Document)
	}{
		{
			name: "prestacao em desacordo",
			input: cte.DesacordoInput{
				TpEvento:      cte.DesacordoPrestacao,
				ChCTe:         testCTeEventAccessKey,
				CNPJ:          desacordoCNPJ,
				COrgao:        desacordoCOrgao,
				NSeqEvento:    1,
				DhEvento:      desacordoTime(),
				TpAmb:         "2",
				Justificativa: desacordoJustificativa,
			},
			assertFn: func(t *testing.T, doc *cte.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoPrestDesacordo)
				require.NotNil(t, doc.EventoPrestDesacordo.InfEvento)
				inf := doc.EventoPrestDesacordo.InfEvento
				assert.Equal(t, "4.00", doc.EventoPrestDesacordo.VersaoAttr)
				assert.Equal(t, "ID610110"+testCTeEventAccessKey+"01", inf.IdAttr)
				assert.Equal(t, desacordoCOrgao, inf.COrgao)
				assert.Equal(t, "2", inf.TpAmb)
				require.NotNil(t, inf.CNPJ)
				assert.Equal(t, desacordoCNPJ, *inf.CNPJ)
				assert.Equal(t, testCTeEventAccessKey, inf.ChCTe)
				assert.Equal(t, testCTeEventIssueDate, inf.DhEvento)
				assert.Equal(t, cte.DesacordoPrestacao, inf.TpEvento)
				assert.Equal(t, "1", inf.NSeqEvento)
				require.NotNil(t, inf.DetEvento)
				assert.Equal(t, "4.00", inf.DetEvento.VersaoEventoAttr)
				require.NotNil(t, inf.DetEvento.EvPrestDesacordo)
				assert.Equal(t, "Prestacao do Servico em Desacordo", inf.DetEvento.EvPrestDesacordo.DescEvento)
				assert.Equal(t, "1", inf.DetEvento.EvPrestDesacordo.IndDesacordoOper)
				assert.Equal(t, desacordoJustificativa, inf.DetEvento.EvPrestDesacordo.XObs)
			},
		},
		{
			name: "cancelamento desacordo",
			input: cte.DesacordoInput{
				TpEvento:        cte.DesacordoCancelamento,
				ChCTe:           testCTeEventAccessKey,
				CNPJ:            desacordoCNPJ,
				COrgao:          desacordoCOrgao,
				NSeqEvento:      2,
				DhEvento:        desacordoTime(),
				TpAmb:           "2",
				ProtocoloEvento: desacordoTargetProt,
			},
			assertFn: func(t *testing.T, doc *cte.Document) {
				t.Helper()
				require.NotNil(t, doc.EventoCancPrestDesacordo)
				require.NotNil(t, doc.EventoCancPrestDesacordo.InfEvento)
				inf := doc.EventoCancPrestDesacordo.InfEvento
				assert.Equal(t, "4.00", doc.EventoCancPrestDesacordo.VersaoAttr)
				assert.Equal(t, "ID610111"+testCTeEventAccessKey+"02", inf.IdAttr)
				assert.Equal(t, cte.DesacordoCancelamento, inf.TpEvento)
				assert.Equal(t, "2", inf.NSeqEvento)
				require.NotNil(t, inf.DetEvento)
				assert.Equal(t, "4.00", inf.DetEvento.VersaoEventoAttr)
				require.NotNil(t, inf.DetEvento.EvCancPrestDesacordo)
				assert.Equal(t, "Cancelamento Prestacao do Servico em Desacordo", inf.DetEvento.EvCancPrestDesacordo.DescEvento)
				assert.Equal(t, desacordoTargetProt, inf.DetEvento.EvCancPrestDesacordo.NProtEvPrestDes)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc, err := cte.BuildDesacordo(tt.input)
			require.NoError(t, err)
			assert.Equal(t, "4.00", doc.VersaoAttr)
			assert.Equal(t, "eventoCTe", doc.RootName)
			assert.Equal(t, tt.input.TpEvento, doc.GetEventType())
			assert.Equal(t, strconv.Itoa(tt.input.NSeqEvento), doc.GetEventSequence())
			assert.Equal(t, testCTeEventAccessKey, doc.GetAccessKey())
			assert.Equal(t, "2", doc.GetEnvironment())
			assert.Equal(t, testCTeEventIssueDate, doc.GetIssueDate())
			tt.assertFn(t, doc)
		})
	}
}

func TestBuildDesacordo_Validation(t *testing.T) {
	t.Parallel()

	base := cte.DesacordoInput{
		TpEvento:      cte.DesacordoPrestacao,
		ChCTe:         testCTeEventAccessKey,
		CNPJ:          desacordoCNPJ,
		COrgao:        desacordoCOrgao,
		NSeqEvento:    1,
		DhEvento:      desacordoTime(),
		TpAmb:         "2",
		Justificativa: desacordoJustificativa,
	}

	tests := []struct {
		name    string
		mutate  func(*cte.DesacordoInput)
		wantErr error
	}{
		{"unsupported event", func(in *cte.DesacordoInput) { in.TpEvento = "999999" }, cte.ErrUnsupportedDesacordo},
		{"short access key", func(in *cte.DesacordoInput) { in.ChCTe = "123" }, cte.ErrInvalidDesacordoAccessKey},
		{"non-numeric access key", func(in *cte.DesacordoInput) { in.ChCTe = strings.Repeat("x", 44) }, cte.ErrInvalidDesacordoAccessKey},
		{"bad cnpj", func(in *cte.DesacordoInput) { in.CNPJ = "123" }, cte.ErrInvalidDesacordoCNPJ},
		{"bad org", func(in *cte.DesacordoInput) { in.COrgao = "4x" }, cte.ErrInvalidDesacordoAuthoringOrg},
		{"sequence zero", func(in *cte.DesacordoInput) { in.NSeqEvento = 0 }, cte.ErrInvalidDesacordoSequence},
		{"sequence too high", func(in *cte.DesacordoInput) { in.NSeqEvento = 1000 }, cte.ErrInvalidDesacordoSequence},
		{"bad environment", func(in *cte.DesacordoInput) { in.TpAmb = "3" }, cte.ErrInvalidDesacordoEnvironment},
		{"zero event time", func(in *cte.DesacordoInput) { in.DhEvento = time.Time{} }, cte.ErrInvalidDesacordoEventTime},
		{"event time with minute offset", func(in *cte.DesacordoInput) {
			in.DhEvento = time.Date(2024, 1, 2, 3, 4, 5, 0, time.FixedZone("IST", 5*3600+30*60))
		}, cte.ErrInvalidDesacordoEventTime},
		{"missing justificativa", func(in *cte.DesacordoInput) { in.Justificativa = "" }, cte.ErrInvalidDesacordoJustificativa},
		{"short justificativa", func(in *cte.DesacordoInput) { in.Justificativa = "curta" }, cte.ErrInvalidDesacordoJustificativa},
		{"justificativa with newline", func(in *cte.DesacordoInput) { in.Justificativa = "tomador discorda\nprestacao" }, cte.ErrInvalidDesacordoJustificativa},
		{"justificativa with unsupported character", func(in *cte.DesacordoInput) { in.Justificativa = "Tomador discorda da prestacao 😀" }, cte.ErrInvalidDesacordoJustificativa},
		{"protocol on prestacao", func(in *cte.DesacordoInput) { in.ProtocoloEvento = desacordoTargetProt }, cte.ErrInvalidDesacordoTargetProtocol},
		{"missing protocol on cancelamento", func(in *cte.DesacordoInput) {
			in.TpEvento = cte.DesacordoCancelamento
			in.Justificativa = ""
		}, cte.ErrInvalidDesacordoTargetProtocol},
		{"bad protocol on cancelamento", func(in *cte.DesacordoInput) {
			in.TpEvento = cte.DesacordoCancelamento
			in.Justificativa = ""
			in.ProtocoloEvento = "abc"
		}, cte.ErrInvalidDesacordoTargetProtocol},
		{"justificativa on cancelamento", func(in *cte.DesacordoInput) {
			in.TpEvento = cte.DesacordoCancelamento
			in.ProtocoloEvento = desacordoTargetProt
		}, cte.ErrInvalidDesacordoJustificativa},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := base
			tt.mutate(&in)
			_, err := cte.BuildDesacordo(in)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestMarshalDesacordoEvento_RoundTrip(t *testing.T) {
	t.Parallel()

	doc, err := cte.BuildDesacordo(cte.DesacordoInput{
		TpEvento:      cte.DesacordoPrestacao,
		ChCTe:         testCTeEventAccessKey,
		CNPJ:          desacordoCNPJ,
		COrgao:        desacordoCOrgao,
		NSeqEvento:    1,
		DhEvento:      desacordoTime(),
		TpAmb:         "2",
		Justificativa: desacordoJustificativa,
	})
	require.NoError(t, err)

	data, err := cte.MarshalDesacordoEvento(doc)
	require.NoError(t, err)
	assert.Contains(t, string(data), `<eventoCTe xmlns="http://www.portalfiscal.inf.br/cte" versao="4.00">`)
	assert.Contains(t, string(data), "<tpEvento>610110</tpEvento>")
	assert.Contains(t, string(data), "<descEvento>Prestacao do Servico em Desacordo</descEvento>")
	assert.Contains(t, string(data), "<indDesacordoOper>1</indDesacordoOper>")
	assert.Contains(t, string(data), "<xObs>"+desacordoJustificativa+"</xObs>")

	parsed, err := cte.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, parsed.EventoPrestDesacordo)
	assert.Equal(t, cte.DesacordoPrestacao, parsed.GetEventType())
	assert.Equal(t, "1", parsed.GetEventSequence())
}

func TestBuildDesacordoProcEvento_RoundTrip(t *testing.T) {
	t.Parallel()

	evento, err := cte.BuildDesacordo(cte.DesacordoInput{
		TpEvento:        cte.DesacordoCancelamento,
		ChCTe:           testCTeEventAccessKey,
		CNPJ:            desacordoCNPJ,
		COrgao:          desacordoCOrgao,
		NSeqEvento:      1,
		DhEvento:        desacordoTime(),
		TpAmb:           "2",
		ProtocoloEvento: desacordoTargetProt,
	})
	require.NoError(t, err)

	retEvento, err := cte.Parse([]byte(minimalCTeRetEventXML(cte.DesacordoCancelamento)))
	require.NoError(t, err)

	procEvento, err := cte.BuildDesacordoProcEvento(evento, retEvento)
	require.NoError(t, err)
	require.NotNil(t, procEvento.ProcEventoCancPrestDesacordo)

	data, err := cte.MarshalDesacordoProcEvento(procEvento)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<nProtEvPrestDes>"+desacordoTargetProt+"</nProtEvPrestDes>")

	parsed, err := cte.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, parsed.ProcEventoCancPrestDesacordo)
	assert.Equal(t, cte.DesacordoCancelamento, parsed.GetEventType())
	assert.Equal(t, testCTeEventProtocol, parsed.GetProtocolNumber())
}

func TestBuildDesacordoProcEvento_Validation(t *testing.T) {
	t.Parallel()

	evento, err := cte.BuildDesacordo(cte.DesacordoInput{
		TpEvento:      cte.DesacordoPrestacao,
		ChCTe:         testCTeEventAccessKey,
		CNPJ:          desacordoCNPJ,
		COrgao:        desacordoCOrgao,
		NSeqEvento:    1,
		DhEvento:      desacordoTime(),
		TpAmb:         "2",
		Justificativa: desacordoJustificativa,
	})
	require.NoError(t, err)
	retEvento, err := cte.Parse([]byte(minimalCTeRetEventXML(cte.DesacordoCancelamento)))
	require.NoError(t, err)

	_, err = cte.BuildDesacordoProcEvento(evento, retEvento)
	require.ErrorIs(t, err, cte.ErrInvalidDesacordoProcessedEvent)
}
