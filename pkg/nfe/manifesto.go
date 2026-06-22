package nfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	mdeSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_mde"
	"github.com/awafinance/fiscal/internal/xmlutil"
)

// Manifestação do destinatário event types (MOC NF-e — NFeRecepcaoEvento,
// NT 2020.001). These are the recipient/buyer's answers to an NF-e issued
// against their CNPJ.
const (
	ManifestoCiencia              = "210210" // Ciência da Operação — non-conclusive ack; unlocks the full XML
	ManifestoConfirmacao          = "210200" // Confirmação da Operação — conclusive positive
	ManifestoDesconhecimento      = "210220" // Desconhecimento da Operação — conclusive negative
	ManifestoOperacaoNaoRealizada = "210240" // Operação não Realizada — conclusive negative; needs justificativa
)

const (
	manifestoVersao         = "1.00"
	manifestoCOrgao         = "91" // Ambiente Nacional (RFB) — destination for every manifestação event
	manifestoDateTimeLayout = "2006-01-02T15:04:05-07:00"
)

// descEvento carries the fixed literal SEFAZ expects per tpEvento (ASCII, no
// accents, per the MOC). The map doubles as the set of supported events.
var manifestoDescEvento = map[string]string{
	ManifestoCiencia:              "Ciencia da Operacao",
	ManifestoConfirmacao:          "Confirmacao da Operacao",
	ManifestoDesconhecimento:      "Desconhecimento da Operacao",
	ManifestoOperacaoNaoRealizada: "Operacao nao Realizada",
}

// Validation sentinels. Callers (e.g. the Nyx domain layer) match them with
// errors.Is to surface predictable command errors.
var (
	ErrInvalidAccessKey     = errors.New("nfe manifesto: chNFe must be 44 digits")
	ErrInvalidCNPJ          = errors.New("nfe manifesto: CNPJ must be 14 digits")
	ErrInvalidSequence      = errors.New("nfe manifesto: nSeqEvento must be between 1 and 20")
	ErrInvalidEnvironment   = errors.New(`nfe manifesto: tpAmb must be "1" (produção) or "2" (homologação)`)
	ErrInvalidEventTime     = errors.New("nfe manifesto: dhEvento must be non-zero")
	ErrUnsupportedManifesto = errors.New("nfe manifesto: unsupported tpEvento")
	ErrJustificativa        = errors.New("nfe manifesto: invalid justificativa")
	ErrInvalidLoteID        = errors.New("nfe manifesto: idLote must be 1 to 15 digits")
	ErrInvalidEnvelope      = errors.New("nfe manifesto: envEvento must contain 1 to 20 non-nil eventos")
)

// ManifestoInput is one recipient manifestação command to build. TpEvento must
// be one of the Manifesto* constants. Justificativa is required for
// ManifestoOperacaoNaoRealizada (15–255 chars), optional for
// ManifestoDesconhecimento (15–255 chars), and must be empty otherwise.
type ManifestoInput struct {
	TpEvento      string
	ChNFe         string    // 44-digit NF-e access key
	CNPJ          string    // recipient/author CNPJ (14 digits)
	NSeqEvento    int       // event sequence, 1–20
	DhEvento      time.Time // event timestamp; carries the offset (SEFAZ wants -03:00, not Z)
	TpAmb         string    // "1" produção, "2" homologação
	Justificativa string    // for ManifestoDesconhecimento and ManifestoOperacaoNaoRealizada
}

// BuildManifesto builds a single, unsigned recipient manifestação evento, ready
// to be signed (enveloped XMLDSig over infEvento) and wrapped in an envEvento
// for NFeRecepcaoEvento4. Signing is intentionally left to the caller
// (fiscal-transporter), which owns the e-CNPJ certificate.
func BuildManifesto(in ManifestoInput) (*EventoMDETEvento, error) {
	desc, ok := manifestoDescEvento[in.TpEvento]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedManifesto, in.TpEvento)
	}
	if !isDigits(in.ChNFe, 44) {
		return nil, ErrInvalidAccessKey
	}
	if !isDigits(in.CNPJ, 14) {
		return nil, ErrInvalidCNPJ
	}
	if in.NSeqEvento < 1 || in.NSeqEvento > 20 {
		return nil, ErrInvalidSequence
	}
	if in.TpAmb != "1" && in.TpAmb != "2" {
		return nil, ErrInvalidEnvironment
	}
	if in.DhEvento.IsZero() {
		return nil, ErrInvalidEventTime
	}
	just, err := manifestoJustificativa(in.TpEvento, in.Justificativa)
	if err != nil {
		return nil, err
	}

	seq := strconv.Itoa(in.NSeqEvento)
	idSeq := fmt.Sprintf("%02d", in.NSeqEvento)
	cnpj := in.CNPJ
	return &mdeSchema.TEvento{
		VersaoAttr: manifestoVersao,
		InfEvento: &mdeSchema.TAnonComplexInfEvento1{
			IdAttr:     "ID" + in.TpEvento + in.ChNFe + idSeq,
			COrgao:     manifestoCOrgao,
			TpAmb:      in.TpAmb,
			CNPJ:       &cnpj,
			ChNFe:      in.ChNFe,
			DhEvento:   in.DhEvento.Format(manifestoDateTimeLayout),
			TpEvento:   in.TpEvento,
			NSeqEvento: seq,
			VerEvento:  manifestoVersao,
			DetEvento: &mdeSchema.TAnonComplexDetEvento1{
				VersaoAttr: manifestoVersao,
				DescEvento: desc,
				XJust:      just,
			},
		},
	}, nil
}

// MarshalManifestoEnvEvento serializes a manifestação envEvento under the NF-e
// namespace. It marshals the typed MDE eventos (so descEvento/xJust survive,
// unlike the generic envelope) and preserves any DsSignature already injected
// into an evento.
func MarshalManifestoEnvEvento(idLote string, eventos ...*EventoMDETEvento) ([]byte, error) {
	if !isDigitsBetween(idLote, 1, 15) {
		return nil, ErrInvalidLoteID
	}
	if len(eventos) < 1 || len(eventos) > 20 {
		return nil, fmt.Errorf("%w: got %d eventos", ErrInvalidEnvelope, len(eventos))
	}
	for i, ev := range eventos {
		if ev == nil {
			return nil, fmt.Errorf("%w: nil evento at index %d", ErrInvalidEnvelope, i)
		}
	}
	type root struct {
		XMLName    xml.Name             `xml:"envEvento"`
		XMLNS      string               `xml:"xmlns,attr,omitempty"`
		VersaoAttr string               `xml:"versao,attr,omitempty"`
		IdLote     string               `xml:"idLote"`
		Evento     []*mdeSchema.TEvento `xml:"evento"`
	}
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := xmlutil.EncodeCanonical(enc, root{
		XMLName:    xml.Name{Local: "envEvento"},
		XMLNS:      namespace,
		VersaoAttr: manifestoVersao,
		IdLote:     idLote,
		Evento:     eventos,
	}); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// manifestoJustificativa enforces the per-event xJust rule: required (15–255
// chars) for Operação não Realizada, optional for Desconhecimento, and
// forbidden for every other event.
func manifestoJustificativa(tpEvento, just string) (*string, error) {
	just = strings.TrimSpace(just)
	switch tpEvento {
	case ManifestoOperacaoNaoRealizada:
		if n := utf8.RuneCountInString(just); n < 15 || n > 255 {
			return nil, fmt.Errorf("%w: %s requires 15–255 chars, got %d", ErrJustificativa, ManifestoOperacaoNaoRealizada, n)
		}
		return &just, nil
	case ManifestoDesconhecimento:
		if just == "" {
			return nil, nil
		}
		if n := utf8.RuneCountInString(just); n < 15 || n > 255 {
			return nil, fmt.Errorf("%w: %s accepts 15–255 chars, got %d", ErrJustificativa, ManifestoDesconhecimento, n)
		}
		return &just, nil
	default:
		if just != "" {
			return nil, fmt.Errorf("%w: only %s and %s accept a justificativa", ErrJustificativa, ManifestoDesconhecimento, ManifestoOperacaoNaoRealizada)
		}
		return nil, nil
	}
}

func isDigits(s string, n int) bool {
	return isDigitsBetween(s, n, n)
}

func isDigitsBetween(s string, minLen, maxLen int) bool {
	if len(s) < minLen || len(s) > maxLen {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
