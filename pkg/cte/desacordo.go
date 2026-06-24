package cte

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// CT-e prestação de serviço em desacordo event types.
const (
	DesacordoPrestacao    = "610110"
	DesacordoCancelamento = "610111"
)

const (
	desacordoVersao                  = "4.00"
	desacordoDateTimeLayout          = "2006-01-02T15:04:05-07:00"
	desacordoPrestacaoDescEvento     = "Prestacao do Servico em Desacordo"
	desacordoCancelamentoDescEvento  = "Cancelamento Prestacao do Servico em Desacordo"
	desacordoPrestacaoIndDesacordoOp = "1"
)

var (
	ErrInvalidDesacordoAccessKey      = errors.New("cte desacordo: chCTe must be 44 digits")
	ErrInvalidDesacordoCNPJ           = errors.New("cte desacordo: CNPJ must be 14 digits")
	ErrInvalidDesacordoAuthoringOrg   = errors.New("cte desacordo: cOrgao must be 2 digits")
	ErrInvalidDesacordoSequence       = errors.New("cte desacordo: nSeqEvento must be between 1 and 999")
	ErrInvalidDesacordoEnvironment    = errors.New(`cte desacordo: tpAmb must be "1" (produção) or "2" (homologação)`)
	ErrInvalidDesacordoEventTime      = errors.New("cte desacordo: dhEvento must match CT-e TDateTimeUTC")
	ErrUnsupportedDesacordo           = errors.New("cte desacordo: unsupported tpEvento")
	ErrInvalidDesacordoJustificativa  = errors.New("cte desacordo: invalid justificativa")
	ErrInvalidDesacordoTargetProtocol = errors.New("cte desacordo: nProtEvPrestDes must be 15 digits")
	ErrInvalidDesacordoProcessedEvent = errors.New("cte desacordo: invalid procEventoCTe pair")
)

// DesacordoInput is one CT-e prestação de serviço em desacordo command to
// build. TpEvento must be DesacordoPrestacao or DesacordoCancelamento.
// Justificativa is required only for DesacordoPrestacao. ProtocoloEvento is
// required only for DesacordoCancelamento and must be the accepted 610110 event
// protocol being cancelled.
type DesacordoInput struct {
	TpEvento        string
	ChCTe           string    // 44-digit CT-e access key
	CNPJ            string    // event author CNPJ (14 digits)
	COrgao          string    // authoring órgão code, usually the CT-e UF code
	NSeqEvento      int       // event sequence, 1-999
	DhEvento        time.Time // event timestamp; carries the numeric offset (SEFAZ wants +00:00, not Z)
	TpAmb           string    // "1" produção, "2" homologação
	Justificativa   string    // xObs for DesacordoPrestacao
	ProtocoloEvento string    // nProtEvPrestDes for DesacordoCancelamento
}

// BuildDesacordo builds a single, unsigned CT-e desacordo eventoCTe document,
// ready to be signed over infEvento and submitted to CTeRecepcaoEventoV4.
func BuildDesacordo(in DesacordoInput) (*Document, error) {
	if err := validateDesacordoInput(in); err != nil {
		return nil, err
	}

	switch in.TpEvento {
	case DesacordoPrestacao:
		return buildPrestacaoDesacordo(in)
	case DesacordoCancelamento:
		return buildCancelamentoDesacordo(in)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedDesacordo, in.TpEvento)
	}
}

// MarshalDesacordoEvento serializes a CT-e desacordo eventoCTe document under
// the CT-e namespace, preserving any DsSignature already injected into it.
func MarshalDesacordoEvento(doc *Document) ([]byte, error) {
	if doc == nil || doc.RootName != "eventoCTe" || (doc.EventoPrestDesacordo == nil && doc.EventoCancPrestDesacordo == nil) {
		return nil, ErrUnsupportedDesacordo
	}
	return marshalDesacordoDocument(doc)
}

// BuildDesacordoProcEvento composes the accepted procEventoCTe from the signed
// sent eventoCTe and SEFAZ retEventoCTe. The two documents must be the same
// desacordo event type.
func BuildDesacordoProcEvento(evento, retEvento *Document) (*Document, error) {
	switch {
	case evento != nil && retEvento != nil && evento.EventoPrestDesacordo != nil && retEvento.RetEventoPrestDesacordo != nil && matchingDesacordoProcPair(evento, retEvento):
		return &Document{
			VersaoAttr: desacordoVersao,
			RootName:   "procEventoCTe",
			ProcEventoPrestDesacordo: &EventoPrestDesacordoCTeProcEvento{
				VersaoAttr:   desacordoVersao,
				EventoCTe:    evento.EventoPrestDesacordo,
				RetEventoCTe: retEvento.RetEventoPrestDesacordo,
			},
		}, nil
	case evento != nil && retEvento != nil && evento.EventoCancPrestDesacordo != nil && retEvento.RetEventoCancPrestDesacordo != nil && matchingDesacordoProcPair(evento, retEvento):
		return &Document{
			VersaoAttr: desacordoVersao,
			RootName:   "procEventoCTe",
			ProcEventoCancPrestDesacordo: &EventoCancPrestDesacordoCTeProcEvento{
				VersaoAttr:   desacordoVersao,
				EventoCTe:    evento.EventoCancPrestDesacordo,
				RetEventoCTe: retEvento.RetEventoCancPrestDesacordo,
			},
		}, nil
	default:
		return nil, ErrInvalidDesacordoProcessedEvent
	}
}

// MarshalDesacordoProcEvento serializes a CT-e desacordo procEventoCTe under
// the CT-e namespace.
func MarshalDesacordoProcEvento(doc *Document) ([]byte, error) {
	if doc == nil || doc.RootName != "procEventoCTe" || (doc.ProcEventoPrestDesacordo == nil && doc.ProcEventoCancPrestDesacordo == nil) {
		return nil, ErrInvalidDesacordoProcessedEvent
	}
	return marshalDesacordoDocument(doc)
}

func buildPrestacaoDesacordo(in DesacordoInput) (*Document, error) {
	if strings.TrimSpace(in.ProtocoloEvento) != "" {
		return nil, fmt.Errorf("%w: %s does not accept nProtEvPrestDes", ErrInvalidDesacordoTargetProtocol, DesacordoPrestacao)
	}
	xObs, err := desacordoJustificativa(in.Justificativa)
	if err != nil {
		return nil, err
	}

	cnpj := in.CNPJ
	return &Document{
		VersaoAttr: desacordoVersao,
		RootName:   "eventoCTe",
		EventoPrestDesacordo: &EventoPrestDesacordoCTeEvento{
			VersaoAttr: desacordoVersao,
			InfEvento: &EventoPrestDesacordoCTeAnonComplexInfEvento1{
				IdAttr:     desacordoID(in.TpEvento, in.ChCTe, in.NSeqEvento),
				COrgao:     in.COrgao,
				TpAmb:      in.TpAmb,
				CNPJ:       &cnpj,
				ChCTe:      in.ChCTe,
				DhEvento:   in.DhEvento.Format(desacordoDateTimeLayout),
				TpEvento:   in.TpEvento,
				NSeqEvento: strconv.Itoa(in.NSeqEvento),
				DetEvento: &EventoPrestDesacordoCTeAnonComplexDetEvento1{
					VersaoEventoAttr: desacordoVersao,
					EvPrestDesacordo: &EventoPrestDesacordoCTeAnonComplexEvPrestDesacordo1{
						DescEvento:       desacordoPrestacaoDescEvento,
						IndDesacordoOper: desacordoPrestacaoIndDesacordoOp,
						XObs:             xObs,
					},
				},
			},
		},
	}, nil
}

func buildCancelamentoDesacordo(in DesacordoInput) (*Document, error) {
	protocol := strings.TrimSpace(in.ProtocoloEvento)
	if !isDesacordoDigits(protocol, 15) {
		return nil, ErrInvalidDesacordoTargetProtocol
	}
	if strings.TrimSpace(in.Justificativa) != "" {
		return nil, fmt.Errorf("%w: %s does not accept justificativa", ErrInvalidDesacordoJustificativa, DesacordoCancelamento)
	}

	cnpj := in.CNPJ
	return &Document{
		VersaoAttr: desacordoVersao,
		RootName:   "eventoCTe",
		EventoCancPrestDesacordo: &EventoCancPrestDesacordoCTeEvento{
			VersaoAttr: desacordoVersao,
			InfEvento: &EventoCancPrestDesacordoCTeAnonComplexInfEvento1{
				IdAttr:     desacordoID(in.TpEvento, in.ChCTe, in.NSeqEvento),
				COrgao:     in.COrgao,
				TpAmb:      in.TpAmb,
				CNPJ:       &cnpj,
				ChCTe:      in.ChCTe,
				DhEvento:   in.DhEvento.Format(desacordoDateTimeLayout),
				TpEvento:   in.TpEvento,
				NSeqEvento: strconv.Itoa(in.NSeqEvento),
				DetEvento: &EventoCancPrestDesacordoCTeAnonComplexDetEvento1{
					VersaoEventoAttr: desacordoVersao,
					EvCancPrestDesacordo: &EventoCancPrestDesacordoCTeAnonComplexEvCancPrestDesacordo1{
						DescEvento:      desacordoCancelamentoDescEvento,
						NProtEvPrestDes: protocol,
					},
				},
			},
		},
	}, nil
}

func validateDesacordoInput(in DesacordoInput) error {
	switch in.TpEvento {
	case DesacordoPrestacao, DesacordoCancelamento:
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedDesacordo, in.TpEvento)
	}
	if !isDesacordoDigits(in.ChCTe, 44) {
		return ErrInvalidDesacordoAccessKey
	}
	if !isDesacordoDigits(in.CNPJ, 14) {
		return ErrInvalidDesacordoCNPJ
	}
	if !isDesacordoDigits(in.COrgao, 2) {
		return ErrInvalidDesacordoAuthoringOrg
	}
	if in.NSeqEvento < 1 || in.NSeqEvento > 999 {
		return ErrInvalidDesacordoSequence
	}
	if in.TpAmb != "1" && in.TpAmb != "2" {
		return ErrInvalidDesacordoEnvironment
	}
	if !validDesacordoEventTime(in.DhEvento) {
		return ErrInvalidDesacordoEventTime
	}
	return nil
}

func desacordoJustificativa(just string) (string, error) {
	just = strings.TrimSpace(just)
	if n := utf8.RuneCountInString(just); n < 15 || n > 255 {
		return "", fmt.Errorf("%w: %s requires 15-255 chars, got %d", ErrInvalidDesacordoJustificativa, DesacordoPrestacao, n)
	}
	for _, r := range just {
		if r < ' ' || r > '\u00ff' {
			return "", fmt.Errorf("%w: %s contains characters outside the xObs schema range", ErrInvalidDesacordoJustificativa, DesacordoPrestacao)
		}
	}
	return just, nil
}

func validDesacordoEventTime(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	year := t.Year()
	_, offset := t.Zone()
	offsetHours := offset / 3600
	return year >= 2000 && year <= 2099 && offset%3600 == 0 && offsetHours >= -11 && offsetHours <= 12
}

func desacordoID(tpEvento, chCTe string, nSeqEvento int) string {
	return "ID" + tpEvento + chCTe + desacordoIDSequence(nSeqEvento)
}

func desacordoIDSequence(n int) string {
	if n < 100 {
		return fmt.Sprintf("%02d", n)
	}
	return strconv.Itoa(n)
}

func marshalDesacordoDocument(doc *Document) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := doc.MarshalXML(enc, xml.StartElement{}); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func matchingDesacordoProcPair(evento, retEvento *Document) bool {
	return matchingOptionalValue(evento.GetEventType(), retEvento.GetEventType()) &&
		matchingOptionalValue(evento.GetEventSequence(), retEvento.GetEventSequence()) &&
		matchingOptionalValue(evento.GetAccessKey(), retEvento.GetAccessKey()) &&
		matchingOptionalValue(evento.GetEnvironment(), retEvento.GetEnvironment())
}

func matchingOptionalValue(a, b string) bool {
	return a == "" || b == "" || a == b
}

func isDesacordoDigits(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
