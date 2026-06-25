package nfse

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	schemaV101 "github.com/awafinance/fiscal/internal/nfse/gen/v1_1/core"
	"github.com/awafinance/fiscal/internal/xmlutil"
)

const (
	TomadorManifestacaoConfirmacao = "e203202"
	TomadorManifestacaoRejeicao    = "e203206"

	tomadorManifestacaoVersao         = "1.01"
	tomadorManifestacaoDateTimeLayout = "2006-01-02T15:04:05-07:00"
	tomadorConfirmacaoDesc            = "Manifestação de NFS-e - Confirmação do Tomador"
	tomadorRejeicaoDesc               = "Manifestação de NFS-e - Rejeição do Tomador"
)

var (
	ErrInvalidNFSeAccessKey             = errors.New("nfse manifestacao: chNFSe must be 50 digits")
	ErrInvalidNFSeAuthor                = errors.New("nfse manifestacao: exactly one author CPF/CNPJ is required")
	ErrInvalidNFSeEnvironment           = errors.New(`nfse manifestacao: tpAmb must be "1" (produção) or "2" (homologação)`)
	ErrInvalidNFSeEventTime             = errors.New("nfse manifestacao: dhEvento must match NFS-e TSDateTimeUTC")
	ErrInvalidNFSeApplicationVersion    = errors.New("nfse manifestacao: verAplic is required")
	ErrUnsupportedTomadorManifestacao   = errors.New("nfse manifestacao: unsupported tomador event")
	ErrInvalidTomadorRejeicaoMotivo     = errors.New("nfse manifestacao: invalid rejeicao motivo")
	ErrInvalidTomadorRejeicaoXMotivo    = errors.New("nfse manifestacao: invalid rejeicao justificativa")
	ErrInvalidTomadorManifestacaoEvento = errors.New("nfse manifestacao: invalid pedRegEvento")
)

// TomadorManifestacaoInput is one NFS-e Nacional tomador manifestation request.
// TpEvento must be TomadorManifestacaoConfirmacao or TomadorManifestacaoRejeicao.
// CNPJAutor and CPFAutor are mutually exclusive; Nyx uses CNPJAutor for company
// tomador events.
type TomadorManifestacaoInput struct {
	TpEvento      string
	ChNFSe        string
	CNPJAutor     string
	CPFAutor      string
	DhEvento      time.Time
	TpAmb         string
	VerAplic      string
	Motivo        string
	Justificativa string
}

// BuildTomadorManifestacao builds an unsigned NFS-e v1.01 pedRegEvento for
// SEFIN Eventos. Signing is intentionally left to the caller.
func BuildTomadorManifestacao(in TomadorManifestacaoInput) (*PedRegEventoV101, error) {
	if err := validateTomadorManifestacaoInput(in); err != nil {
		return nil, err
	}

	inf := &schemaV101.TCInfPedReg{
		IdAttr:   "PRE" + in.ChNFSe + eventCodeDigits(in.TpEvento),
		TpAmb:    in.TpAmb,
		VerAplic: strings.TrimSpace(in.VerAplic),
		DhEvento: in.DhEvento.Format(tomadorManifestacaoDateTimeLayout),
		ChNFSe:   in.ChNFSe,
	}
	if in.CNPJAutor != "" {
		cnpj := in.CNPJAutor
		inf.CNPJAutor = &cnpj
	} else {
		cpf := in.CPFAutor
		inf.CPFAutor = &cpf
	}

	switch in.TpEvento {
	case TomadorManifestacaoConfirmacao:
		inf.E203202 = &schemaV101.TE203202{XDesc: tomadorConfirmacaoDesc}
	case TomadorManifestacaoRejeicao:
		just := strings.TrimSpace(in.Justificativa)
		inf.E203206 = &schemaV101.TE203206{
			XDesc:   tomadorRejeicaoDesc,
			CMotivo: in.Motivo,
		}
		if just != "" {
			inf.E203206.XMotivo = &just
		}
	}

	return &schemaV101.TCPedRegEvt{
		VersaoAttr: tomadorManifestacaoVersao,
		InfPedReg:  inf,
	}, nil
}

// MarshalTomadorManifestacao serializes a v1.01 tomador manifestation
// pedRegEvento under the NFS-e namespace.
func MarshalTomadorManifestacao(evento *PedRegEventoV101) ([]byte, error) {
	if evento == nil || evento.InfPedReg == nil {
		return nil, ErrInvalidTomadorManifestacaoEvento
	}

	type root struct {
		XMLName    xml.Name                  `xml:"pedRegEvento"`
		XMLNS      string                    `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                    `xml:"versao,attr"`
		InfPedReg  *schemaV101.TCInfPedReg   `xml:"infPedReg"`
		Signature  *schemaV101.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}

	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := xmlutil.EncodeCanonical(enc, root{
		XMLName:    xml.Name{Local: "pedRegEvento"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(evento.VersaoAttr, tomadorManifestacaoVersao),
		InfPedReg:  evento.InfPedReg,
		Signature:  evento.DsSignature,
	}); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func validateTomadorManifestacaoInput(in TomadorManifestacaoInput) error {
	switch in.TpEvento {
	case TomadorManifestacaoConfirmacao, TomadorManifestacaoRejeicao:
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedTomadorManifestacao, in.TpEvento)
	}
	if !isNFSeDigits(in.ChNFSe, 50) {
		return ErrInvalidNFSeAccessKey
	}
	if err := validateTomadorManifestacaoAuthor(in); err != nil {
		return err
	}
	if in.TpAmb != "1" && in.TpAmb != "2" {
		return ErrInvalidNFSeEnvironment
	}
	if !validTomadorManifestacaoEventTime(in.DhEvento) {
		return ErrInvalidNFSeEventTime
	}
	if strings.TrimSpace(in.VerAplic) == "" {
		return ErrInvalidNFSeApplicationVersion
	}
	return validateTomadorManifestacaoMotivo(in)
}

func validateTomadorManifestacaoAuthor(in TomadorManifestacaoInput) error {
	hasCNPJ := in.CNPJAutor != ""
	hasCPF := in.CPFAutor != ""
	switch {
	case hasCNPJ == hasCPF:
		return ErrInvalidNFSeAuthor
	case hasCNPJ && !isNFSeDigits(in.CNPJAutor, 14):
		return ErrInvalidNFSeAuthor
	case hasCPF && !isNFSeDigits(in.CPFAutor, 11):
		return ErrInvalidNFSeAuthor
	default:
		return nil
	}
}

func validateTomadorManifestacaoMotivo(in TomadorManifestacaoInput) error {
	just := strings.TrimSpace(in.Justificativa)
	if in.TpEvento == TomadorManifestacaoConfirmacao {
		if in.Motivo != "" || just != "" {
			return fmt.Errorf("%w: confirmacao does not accept motivo", ErrInvalidTomadorRejeicaoMotivo)
		}
		return nil
	}

	switch in.Motivo {
	case "1", "2", "3", "4", "5", "9":
	default:
		return ErrInvalidTomadorRejeicaoMotivo
	}
	if in.Motivo == "9" && just == "" {
		return fmt.Errorf("%w: motivo 9 requires justificativa", ErrInvalidTomadorRejeicaoXMotivo)
	}
	if just == "" {
		return nil
	}
	if n := utf8.RuneCountInString(just); n < 15 || n > 255 {
		return fmt.Errorf("%w: xMotivo must be 15-255 chars, got %d", ErrInvalidTomadorRejeicaoXMotivo, n)
	}
	for _, r := range just {
		if r < ' ' || r > '\u00ff' {
			return fmt.Errorf("%w: xMotivo contains characters outside the schema range", ErrInvalidTomadorRejeicaoXMotivo)
		}
	}
	return nil
}

func validTomadorManifestacaoEventTime(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	year := t.Year()
	_, offset := t.Zone()
	offsetHours := offset / 3600
	return year >= 2000 && year <= 2099 && offset%3600 == 0 && offsetHours >= -11 && offsetHours <= 12
}

func eventCodeDigits(tpEvento string) string {
	return strings.TrimPrefix(tpEvento, "e")
}

func isNFSeDigits(s string, n int) bool {
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
