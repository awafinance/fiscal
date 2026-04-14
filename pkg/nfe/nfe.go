package nfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	atorSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/ator_interessado"
	distSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/dist_dfe"
	epecSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/epec"
	cancelSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel"
	cancelEntregaSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel_entrega"
	insucessoCancelSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel_insucesso"
	cceSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cce"
	entregaSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_entrega"
	genericSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_generico"
	insucessoSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_insucesso"
	mdeSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_mde"
	consSchema "github.com/awafinance/fiscal/internal/nfe/gen/v2_0/cons"
	inutSchema "github.com/awafinance/fiscal/internal/nfe/gen/v4_0/inutilizacao"
	schema "github.com/awafinance/fiscal/internal/nfe/gen/v4_0/nfe_proc"
	statusSchema "github.com/awafinance/fiscal/internal/nfe/gen/v4_0/status_servico"
	"github.com/awafinance/fiscal/internal/xmlutil"
)

const (
	namespace = "http://www.portalfiscal.inf.br/nfe"
	nfeProc   = "nfeProc"
)

var errUnsupportedRoot = errors.New("marshal nfe: document must contain exactly one supported root")

type Document struct {
	VersaoAttr            string                                 `json:"versao,omitempty"`
	NFe                   *schema.TNFe                           `json:"NFe,omitempty"`
	ProtNFe               *schema.TProtNFe                       `json:"protNFe,omitempty"`
	EnviNFe               *schema.TEnviNFe                       `json:"enviNFe,omitempty"`
	RetEnviNFe            *schema.TRetEnviNFe                    `json:"retEnviNFe,omitempty"`
	ConsReciNFe           *schema.TConsReciNFe                   `json:"consReciNFe,omitempty"`
	RetConsReciNFe        *schema.TRetConsReciNFe                `json:"retConsReciNFe,omitempty"`
	EventoCancel          *cancelSchema.TEvento                  `json:"eventoCancel,omitempty"`
	EventoEntrega         *entregaSchema.TEvento                 `json:"eventoEntrega,omitempty"`
	EventoCancEntrega     *cancelEntregaSchema.TEvento           `json:"eventoCancEntrega,omitempty"`
	EventoCCe             *cceSchema.TEvento                     `json:"eventoCCe,omitempty"`
	EventoEPEC            *epecSchema.TEvento                    `json:"eventoEPEC,omitempty"`
	EventoAtorInteressado *atorSchema.TEvento                    `json:"eventoAtorInteressado,omitempty"`
	EventoMDE             *mdeSchema.TEvento                     `json:"eventoMDE,omitempty"`
	EventoInsucesso       *insucessoSchema.TEvento               `json:"eventoInsucesso,omitempty"`
	EventoCancInsucesso   *insucessoCancelSchema.TEvento         `json:"eventoCancInsucesso,omitempty"`
	EventoGenerico        *genericSchema.TEvento                 `json:"eventoGenerico,omitempty"`
	EnvEvento             *genericSchema.TEnvEvento              `json:"envEvento,omitempty"`
	RetEnvEvento          *genericSchema.TRetEnvEvento           `json:"retEnvEvento,omitempty"`
	ProcEventoNFe         *genericSchema.TProcEvento             `json:"procEventoNFe,omitempty"`
	ConsSitNFe            *consSchema.TConsSitNFe                `json:"consSitNFe,omitempty"`
	RetConsSitNFe         *consSchema.TRetConsSitNFe             `json:"retConsSitNFe,omitempty"`
	ConsStatServ          *statusSchema.TConsStatServ            `json:"consStatServ,omitempty"`
	RetConsStatServ       *statusSchema.TRetConsStatServ         `json:"retConsStatServ,omitempty"`
	InutNFe               *inutSchema.TInutNFe                   `json:"inutNFe,omitempty"`
	RetInutNFe            *inutSchema.TRetInutNFe                `json:"retInutNFe,omitempty"`
	ProcInutNFe           *inutSchema.TProcInutNFe               `json:"procInutNFe,omitempty"`
	DistDFeInt            *distSchema.TAnonComplexDistDFeInt1    `json:"distDFeInt,omitempty"`
	RetDistDFeInt         *distSchema.TAnonComplexRetDistDFeInt1 `json:"retDistDFeInt,omitempty"`
	ResNFe                *distSchema.TAnonComplexResNFe1        `json:"resNFe,omitempty"`
	ResEvento             *distSchema.TAnonComplexResEvento1     `json:"resEvento,omitempty"`
	RootName              string                                 `json:"rootName,omitempty"`
}

// MarshalXML preserves the parsed root when possible.
// If protocol data is present, the document is always encoded as nfeProc.
func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	if d.RootName != nfeProc && d.ProtNFe == nil {
		if d.NFe != nil || activeRootCount(d) == 0 {
			return encodeBareNFe(e, d.NFe)
		}
		return marshalSingleRoot(e, d)
	}

	return encodeNFeProc(e, d)
}

type rootEncoder struct {
	match  func(*Document) bool
	encode func(*xml.Encoder, *Document) error
}

var nfeRoots = []rootEncoder{
	{match: func(d *Document) bool { return d.EnviNFe != nil }, encode: encodeEnviNFe},
	{match: func(d *Document) bool { return d.RetEnviNFe != nil }, encode: encodeRetEnviNFe},
	{match: func(d *Document) bool { return d.ConsReciNFe != nil }, encode: encodeConsReciNFe},
	{match: func(d *Document) bool { return d.RetConsReciNFe != nil }, encode: encodeRetConsReciNFe},
	{match: func(d *Document) bool { return d.EnvEvento != nil }, encode: encodeEnvEvento},
	{match: func(d *Document) bool { return d.RetEnvEvento != nil }, encode: encodeRetEnvEvento},
	{match: func(d *Document) bool { return d.ProcEventoNFe != nil }, encode: encodeProcEventoNFe},
	{match: func(d *Document) bool { return d.ConsSitNFe != nil }, encode: encodeConsSitNFe},
	{match: func(d *Document) bool { return d.RetConsSitNFe != nil }, encode: encodeRetConsSitNFe},
	{match: func(d *Document) bool { return d.ConsStatServ != nil }, encode: encodeConsStatServ},
	{match: func(d *Document) bool { return d.RetConsStatServ != nil }, encode: encodeRetConsStatServ},
	{match: func(d *Document) bool { return d.InutNFe != nil }, encode: encodeInutNFe},
	{match: func(d *Document) bool { return d.RetInutNFe != nil }, encode: encodeRetInutNFe},
	{match: func(d *Document) bool { return d.ProcInutNFe != nil }, encode: encodeProcInutNFe},
	{match: func(d *Document) bool { return d.DistDFeInt != nil }, encode: encodeDistDFeInt},
	{match: func(d *Document) bool { return d.RetDistDFeInt != nil }, encode: encodeRetDistDFeInt},
	{match: func(d *Document) bool { return d.ResNFe != nil }, encode: encodeResNFe},
	{match: func(d *Document) bool { return d.ResEvento != nil }, encode: encodeResEvento},
}

func marshalSingleRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errUnsupportedRoot
	}
	if err := marshalEventRoot(e, d); err == nil {
		return nil
	}
	for _, root := range nfeRoots {
		if root.match(d) {
			return root.encode(e, d)
		}
	}
	return errUnsupportedRoot
}

func marshalEventRoot(e *xml.Encoder, d *Document) error {
	switch {
	case d.EventoCancel != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoCancel.VersaoAttr), d.EventoCancel.InfEvento, d.EventoCancel.DsSignature)
	case d.EventoEntrega != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoEntrega.VersaoAttr), d.EventoEntrega.InfEvento, d.EventoEntrega.DsSignature)
	case d.EventoCancEntrega != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoCancEntrega.VersaoAttr), d.EventoCancEntrega.InfEvento, d.EventoCancEntrega.DsSignature)
	case d.EventoCCe != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoCCe.VersaoAttr), d.EventoCCe.InfEvento, d.EventoCCe.DsSignature)
	case d.EventoEPEC != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoEPEC.VersaoAttr), d.EventoEPEC.InfEvento, d.EventoEPEC.DsSignature)
	case d.EventoAtorInteressado != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoAtorInteressado.VersaoAttr), d.EventoAtorInteressado.InfEvento, d.EventoAtorInteressado.DsSignature)
	case d.EventoMDE != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoMDE.VersaoAttr), d.EventoMDE.InfEvento, d.EventoMDE.DsSignature)
	case d.EventoInsucesso != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoInsucesso.VersaoAttr), d.EventoInsucesso.InfEvento, d.EventoInsucesso.DsSignature)
	case d.EventoCancInsucesso != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoCancInsucesso.VersaoAttr), d.EventoCancInsucesso.InfEvento, d.EventoCancInsucesso.DsSignature)
	case d.EventoGenerico != nil:
		return encodeNFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoGenerico.VersaoAttr), d.EventoGenerico.InfEvento, d.EventoGenerico.DsSignature)
	default:
		return errUnsupportedRoot
	}
}

func encodeNFeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	type root struct {
		XMLName     xml.Name `xml:"evento"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "evento"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeBareNFe(e *xml.Encoder, invoice *schema.TNFe) error {
	return xmlutil.EncodeNamespacedRoot(e, "NFe", namespace, invoice)
}

func encodeNFeProc(e *xml.Encoder, d *Document) error {
	type procNFe struct {
		XMLName    xml.Name         `xml:"nfeProc"`
		XMLNS      string           `xml:"xmlns,attr,omitempty"`
		VersaoAttr string           `xml:"versao,attr"`
		NFe        *schema.TNFe     `xml:"NFe"`
		ProtNFe    *schema.TProtNFe `xml:"protNFe"`
	}
	return xmlutil.EncodeCanonical(e, procNFe{
		XMLName:    xml.Name{Local: nfeProc},
		XMLNS:      namespace,
		VersaoAttr: d.VersaoAttr,
		NFe:        d.NFe,
		ProtNFe:    d.ProtNFe,
	})
}

func encodeEnviNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "enviNFe", namespace, d.EnviNFe)
}

func encodeRetEnviNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "retEnviNFe", namespace, d.RetEnviNFe)
}

func encodeConsReciNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "consReciNFe", namespace, d.ConsReciNFe)
}

func encodeRetConsReciNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "retConsReciNFe", namespace, d.RetConsReciNFe)
}

func encodeEnvEvento(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "envEvento", namespace, d.EnvEvento)
}

func encodeRetEnvEvento(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "retEnvEvento", namespace, d.RetEnvEvento)
}

func encodeProcEventoNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "procEventoNFe", namespace, d.ProcEventoNFe)
}

func encodeConsSitNFe(e *xml.Encoder, d *Document) error {
	type root struct {
		XMLName    xml.Name `xml:"consSitNFe"`
		XMLNS      string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr string   `xml:"versao,attr,omitempty"`
		TpAmb      string   `xml:"tpAmb"`
		XServ      string   `xml:"xServ"`
		ChNFe      string   `xml:"chNFe"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "consSitNFe"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.ConsSitNFe.VersaoAttr),
		TpAmb:      d.ConsSitNFe.TpAmb,
		XServ:      d.ConsSitNFe.XServ,
		ChNFe:      d.ConsSitNFe.ChNFe,
	})
}

func encodeRetConsSitNFe(e *xml.Encoder, d *Document) error {
	type root struct {
		XMLName       xml.Name                  `xml:"retConsSitNFe"`
		XMLNS         string                    `xml:"xmlns,attr,omitempty"`
		VersaoAttr    string                    `xml:"versao,attr,omitempty"`
		TpAmb         string                    `xml:"tpAmb"`
		VerAplic      *consSchema.TString       `xml:"verAplic,omitempty"`
		CStat         string                    `xml:"cStat"`
		XMotivo       *consSchema.TString       `xml:"xMotivo,omitempty"`
		CUF           string                    `xml:"cUF"`
		ChNFe         string                    `xml:"chNFe"`
		ProtNFe       *consSchema.TProtNFe      `xml:"protNFe,omitempty"`
		ProcEventoNFe []*consSchema.TProcEvento `xml:"procEventoNFe,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:       xml.Name{Local: "retConsSitNFe"},
		XMLNS:         namespace,
		VersaoAttr:    xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetConsSitNFe.VersaoAttr),
		TpAmb:         d.RetConsSitNFe.TpAmb,
		VerAplic:      d.RetConsSitNFe.VerAplic,
		CStat:         d.RetConsSitNFe.CStat,
		XMotivo:       d.RetConsSitNFe.XMotivo,
		CUF:           d.RetConsSitNFe.CUF,
		ChNFe:         d.RetConsSitNFe.ChNFe,
		ProtNFe:       d.RetConsSitNFe.ProtNFe,
		ProcEventoNFe: d.RetConsSitNFe.ProcEventoNFe,
	})
}

func encodeConsStatServ(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "consStatServ", namespace, d.ConsStatServ)
}

func encodeRetConsStatServ(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "retConsStatServ", namespace, d.RetConsStatServ)
}

func encodeInutNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "inutNFe", namespace, d.InutNFe)
}

func encodeRetInutNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "retInutNFe", namespace, d.RetInutNFe)
}

func encodeProcInutNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "procInutNFe", namespace, d.ProcInutNFe)
}

func encodeDistDFeInt(e *xml.Encoder, d *Document) error {
	type root struct {
		XMLName    xml.Name                           `xml:"distDFeInt"`
		XMLNS      string                             `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                             `xml:"versao,attr,omitempty"`
		TpAmb      string                             `xml:"tpAmb"`
		CUFAutor   *string                            `xml:"cUFAutor,omitempty"`
		CNPJ       *string                            `xml:"CNPJ,omitempty"`
		CPF        *string                            `xml:"CPF,omitempty"`
		DistNSU    *distSchema.TAnonComplexDistNSU1   `xml:"distNSU,omitempty"`
		ConsNSU    *distSchema.TAnonComplexConsNSU1   `xml:"consNSU,omitempty"`
		ConsChNFe  *distSchema.TAnonComplexConsChNFe1 `xml:"consChNFe,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "distDFeInt"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.DistDFeInt.VersaoAttr),
		TpAmb:      d.DistDFeInt.TpAmb,
		CUFAutor:   d.DistDFeInt.CUFAutor,
		CNPJ:       d.DistDFeInt.CNPJ,
		CPF:        d.DistDFeInt.CPF,
		DistNSU:    d.DistDFeInt.DistNSU,
		ConsNSU:    d.DistDFeInt.ConsNSU,
		ConsChNFe:  d.DistDFeInt.ConsChNFe,
	})
}

func encodeRetDistDFeInt(e *xml.Encoder, d *Document) error {
	type root struct {
		XMLName        xml.Name                                `xml:"retDistDFeInt"`
		XMLNS          string                                  `xml:"xmlns,attr,omitempty"`
		VersaoAttr     string                                  `xml:"versao,attr,omitempty"`
		TpAmb          string                                  `xml:"tpAmb"`
		VerAplic       *distSchema.TString                     `xml:"verAplic,omitempty"`
		CStat          string                                  `xml:"cStat"`
		XMotivo        *distSchema.TString                     `xml:"xMotivo,omitempty"`
		DhResp         string                                  `xml:"dhResp"`
		UltNSU         string                                  `xml:"ultNSU"`
		MaxNSU         string                                  `xml:"maxNSU"`
		LoteDistDFeInt *distSchema.TAnonComplexLoteDistDFeInt1 `xml:"loteDistDFeInt,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:        xml.Name{Local: "retDistDFeInt"},
		XMLNS:          namespace,
		VersaoAttr:     xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetDistDFeInt.VersaoAttr),
		TpAmb:          d.RetDistDFeInt.TpAmb,
		VerAplic:       d.RetDistDFeInt.VerAplic,
		CStat:          d.RetDistDFeInt.CStat,
		XMotivo:        d.RetDistDFeInt.XMotivo,
		DhResp:         d.RetDistDFeInt.DhResp,
		UltNSU:         d.RetDistDFeInt.UltNSU,
		MaxNSU:         d.RetDistDFeInt.MaxNSU,
		LoteDistDFeInt: d.RetDistDFeInt.LoteDistDFeInt,
	})
}

func encodeResNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "resNFe", namespace, d.ResNFe)
}

func encodeResEvento(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeNamespacedRoot(e, "resEvento", namespace, d.ResEvento)
}

var parsersByRoot = map[string]func([]byte, string) (*Document, error){
	nfeProc:           parseNFeProc,
	"NFe":             parseNFe,
	"enviNFe":         parseEnviNFe,
	"retEnviNFe":      parseRetEnviNFe,
	"consReciNFe":     parseConsReciNFe,
	"retConsReciNFe":  parseRetConsReciNFe,
	"evento":          parseEventRoot,
	"envEvento":       parseEnvEvento,
	"retEnvEvento":    parseRetEnvEvento,
	"procEventoNFe":   parseProcEventoNFe,
	"consSitNFe":      parseConsSitNFe,
	"retConsSitNFe":   parseRetConsSitNFe,
	"consStatServ":    parseConsStatServ,
	"retConsStatServ": parseRetConsStatServ,
	"inutNFe":         parseInutNFe,
	"retInutNFe":      parseRetInutNFe,
	"procInutNFe":     parseProcInutNFe,
	"distDFeInt":      parseDistDFeInt,
	"retDistDFeInt":   parseRetDistDFeInt,
	"resNFe":          parseResNFe,
	"resEvento":       parseResEvento,
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse nfe: empty xml document")
	}

	rootName, rootErr := xmlutil.ParseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse nfe: read root: %w", rootErr)
	}

	if fn, ok := parsersByRoot[rootName]; ok {
		return fn(data, rootName)
	}
	if rootErr != nil {
		return nil, fmt.Errorf("parse nfe: read root: %w", rootErr)
	}
	return nil, fmt.Errorf("parse nfe: unsupported root element %q", rootName)
}

func finalizeDoc(doc *Document) (*Document, error) {
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func parseNFeProc(data []byte, _ string) (*Document, error) {
	var parsed struct {
		VersaoAttr string           `xml:"versao,attr"`
		NFe        *schema.TNFe     `xml:"NFe"`
		ProtNFe    *schema.TProtNFe `xml:"protNFe"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode nfeProc: %w", err)
	}
	return finalizeDoc(&Document{
		VersaoAttr: parsed.VersaoAttr,
		NFe:        parsed.NFe,
		ProtNFe:    parsed.ProtNFe,
		RootName:   nfeProc,
	})
}

func parseNFe(data []byte, rootName string) (*Document, error) {
	var invoice schema.TNFe
	if err := xml.Unmarshal(data, &invoice); err != nil {
		return nil, fmt.Errorf("parse nfe: decode NFe: %w", err)
	}
	return finalizeDoc(&Document{
		VersaoAttr: versionFromNFe(&invoice),
		NFe:        &invoice,
		RootName:   rootName,
	})
}

func parseEnviNFe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TEnviNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode enviNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EnviNFe: &parsed, RootName: rootName})
}

func parseRetEnviNFe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TRetEnviNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retEnviNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEnviNFe: &parsed, RootName: rootName})
}

func parseConsReciNFe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TConsReciNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode consReciNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsReciNFe: &parsed, RootName: rootName})
}

func parseRetConsReciNFe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TRetConsReciNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retConsReciNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsReciNFe: &parsed, RootName: rootName})
}

func decodeEvent[T any](data []byte, context string, assign func(*T) *Document) (*Document, error) {
	var parsed T
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode %s: %w", context, err)
	}
	return finalizeDoc(assign(&parsed))
}

func parseEventRoot(data []byte, rootName string) (*Document, error) {
	tpEvento, err := eventTypeFromXML(data)
	if err != nil {
		return nil, fmt.Errorf("parse nfe: decode evento head: %w", err)
	}
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "cancel event", func(p *cancelSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancel: p, RootName: rootName}
		})
	case "110130":
		return decodeEvent(data, "entrega event", func(p *entregaSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoEntrega: p, RootName: rootName}
		})
	case "110131":
		return decodeEvent(data, "cancel entrega event", func(p *cancelEntregaSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancEntrega: p, RootName: rootName}
		})
	case "110110":
		return decodeEvent(data, "cce event", func(p *cceSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCCe: p, RootName: rootName}
		})
	case "110140":
		return decodeEvent(data, "epec event", func(p *epecSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoEPEC: p, RootName: rootName}
		})
	case "110150":
		return decodeEvent(data, "ator interessado event", func(p *atorSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoAtorInteressado: p, RootName: rootName}
		})
	case "110192":
		return decodeEvent(data, "insucesso event", func(p *insucessoSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoInsucesso: p, RootName: rootName}
		})
	case "110193":
		return decodeEvent(data, "cancel insucesso event", func(p *insucessoCancelSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancInsucesso: p, RootName: rootName}
		})
	case "210200", "210210", "210220", "210240":
		return decodeEvent(data, "mde event", func(p *mdeSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoMDE: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "generic event", func(p *genericSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoGenerico: p, RootName: rootName}
		})
	}
}

func parseEnvEvento(data []byte, rootName string) (*Document, error) {
	var parsed genericSchema.TEnvEvento
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode envEvento: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EnvEvento: &parsed, RootName: rootName})
}

func parseRetEnvEvento(data []byte, rootName string) (*Document, error) {
	var parsed genericSchema.TRetEnvEvento
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retEnvEvento: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEnvEvento: &parsed, RootName: rootName})
}

func parseProcEventoNFe(data []byte, rootName string) (*Document, error) {
	var parsed genericSchema.TProcEvento
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode procEventoNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoNFe: &parsed, RootName: rootName})
}

func parseConsSitNFe(data []byte, rootName string) (*Document, error) {
	var parsed consSchema.TConsSitNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode consSitNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsSitNFe: &parsed, RootName: rootName})
}

func parseRetConsSitNFe(data []byte, rootName string) (*Document, error) {
	var parsed struct {
		VersaoAttr    string                    `xml:"versao,attr"`
		TpAmb         string                    `xml:"tpAmb"`
		VerAplic      *consSchema.TString       `xml:"verAplic"`
		CStat         string                    `xml:"cStat"`
		XMotivo       *consSchema.TString       `xml:"xMotivo"`
		CUF           string                    `xml:"cUF"`
		ChNFe         string                    `xml:"chNFe"`
		ProtNFe       *consSchema.TProtNFe      `xml:"protNFe"`
		ProcEventoNFe []*consSchema.TProcEvento `xml:"procEventoNFe"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retConsSitNFe: %w", err)
	}
	return finalizeDoc(&Document{
		VersaoAttr: parsed.VersaoAttr,
		RetConsSitNFe: &consSchema.TRetConsSitNFe{
			VersaoAttr:    parsed.VersaoAttr,
			TpAmb:         parsed.TpAmb,
			VerAplic:      parsed.VerAplic,
			CStat:         parsed.CStat,
			XMotivo:       parsed.XMotivo,
			CUF:           parsed.CUF,
			ChNFe:         parsed.ChNFe,
			ProtNFe:       parsed.ProtNFe,
			ProcEventoNFe: parsed.ProcEventoNFe,
		},
		RootName: rootName,
	})
}

func parseConsStatServ(data []byte, rootName string) (*Document, error) {
	var parsed statusSchema.TConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode consStatServ: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsStatServ: &parsed, RootName: rootName})
}

func parseRetConsStatServ(data []byte, rootName string) (*Document, error) {
	var parsed statusSchema.TRetConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retConsStatServ: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServ: &parsed, RootName: rootName})
}

func parseInutNFe(data []byte, rootName string) (*Document, error) {
	var parsed inutSchema.TInutNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode inutNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, InutNFe: &parsed, RootName: rootName})
}

func parseRetInutNFe(data []byte, rootName string) (*Document, error) {
	var parsed inutSchema.TRetInutNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retInutNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetInutNFe: &parsed, RootName: rootName})
}

func parseProcInutNFe(data []byte, rootName string) (*Document, error) {
	var parsed inutSchema.TProcInutNFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode procInutNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcInutNFe: &parsed, RootName: rootName})
}

func parseDistDFeInt(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexDistDFeInt1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode distDFeInt: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, RootName: rootName})
}

func parseRetDistDFeInt(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexRetDistDFeInt1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode retDistDFeInt: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, RootName: rootName})
}

func parseResNFe(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexResNFe1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode resNFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ResNFe: &parsed, RootName: rootName})
}

func parseResEvento(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexResEvento1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfe: decode resEvento: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ResEvento: &parsed, RootName: rootName})
}

func ParseReader(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parse nfe: read xml: %w", err)
	}
	return Parse(data)
}

func eventTypeFromXML(data []byte) (string, error) {
	var head struct {
		InfEvento struct {
			TpEvento string `xml:"tpEvento"`
		} `xml:"infEvento"`
	}
	if err := xml.Unmarshal(data, &head); err != nil {
		return "", err
	}
	return head.InfEvento.TpEvento, nil
}

func versionFromNFe(invoice *schema.TNFe) string {
	if invoice.InfNFe == nil {
		return ""
	}

	return invoice.InfNFe.VersaoAttr
}

var rootValidators = []func(*Document) error{
	validateNFeRoot,
	validateEnviNFeRoot,
	validateRetEnviNFeRoot,
	validateConsReciNFeRoot,
	validateRetConsReciNFeRoot,
	validateEventoCancelRoot,
	validateEventoEntregaRoot,
	validateEventoCancEntregaRoot,
	validateEventoCCeRoot,
	validateEventoEPECRoot,
	validateEventoAtorInteressadoRoot,
	validateEventoMDERoot,
	validateEventoInsucessoRoot,
	validateEventoCancInsucessoRoot,
	validateEventoGenericoRoot,
	validateEnvEventoRoot,
	validateRetEnvEventoRoot,
	validateProcEventoNFeRoot,
	validateConsSitNFeRoot,
	validateRetConsSitNFeRoot,
	validateConsStatServRoot,
	validateRetConsStatServRoot,
	validateInutNFeRoot,
	validateRetInutNFeRoot,
	validateProcInutNFeRoot,
	validateDistDFeIntRoot,
	validateRetDistDFeIntRoot,
	validateResNFeRoot,
	validateResEventoRoot,
}

func validateDocument(doc *Document) error {
	if doc.RootName == "NFe" || doc.RootName == nfeProc || (doc.RootName == "" && doc.ProtNFe != nil) {
		if doc.NFe == nil {
			return errors.New("parse nfe: missing NFe")
		}
	}

	for _, v := range rootValidators {
		if err := v(doc); err != nil {
			return err
		}
	}

	if activeRootCount(doc) != 1 {
		return errors.New("parse nfe: document must contain exactly one supported root")
	}
	return nil
}

func missing(field, value string) error {
	if value == "" {
		return errors.New("parse nfe: missing " + field)
	}
	return nil
}

func firstMissing(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func validateNFeRoot(doc *Document) error {
	if doc.NFe == nil {
		return nil
	}
	if doc.NFe.InfNFe == nil {
		return errors.New("parse nfe: missing infNFe")
	}
	if doc.NFe.InfNFe.Ide == nil {
		return errors.New("parse nfe: missing ide")
	}
	if doc.NFe.InfNFe.Emit == nil {
		return errors.New("parse nfe: missing emit")
	}
	if doc.NFe.InfNFe.Emit.CNPJ == nil && doc.NFe.InfNFe.Emit.CPF == nil {
		return errors.New("parse nfe: missing emit document")
	}
	if len(doc.NFe.InfNFe.Det) == 0 {
		return errors.New("parse nfe: missing det")
	}
	return nil
}

func validateEnviNFeRoot(doc *Document) error {
	if doc.EnviNFe == nil {
		return nil
	}
	if doc.EnviNFe.IdLote == "" {
		return errors.New("parse nfe: missing idLote")
	}
	if len(doc.EnviNFe.NFe) == 0 {
		return errors.New("parse nfe: missing NFe")
	}
	return nil
}

func validateRetEnviNFeRoot(doc *Document) error {
	if doc.RetEnviNFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetEnviNFe.TpAmb),
		missing("cStat", doc.RetEnviNFe.CStat),
		missing("cUF", doc.RetEnviNFe.CUF),
	)
}

func validateConsReciNFeRoot(doc *Document) error {
	if doc.ConsReciNFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.ConsReciNFe.TpAmb),
		missing("nRec", doc.ConsReciNFe.NRec),
	)
}

func validateRetConsReciNFeRoot(doc *Document) error {
	if doc.RetConsReciNFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsReciNFe.TpAmb),
		missing("cStat", doc.RetConsReciNFe.CStat),
		missing("cUF", doc.RetConsReciNFe.CUF),
	)
}

func validateEventoFields(isNil bool, chNFe string, missingDetEvento bool) error {
	if isNil {
		return errors.New("parse nfe: missing infEvento")
	}
	if chNFe == "" {
		return errors.New("parse nfe: missing chNFe")
	}
	if missingDetEvento {
		return errors.New("parse nfe: missing detEvento")
	}
	return nil
}

func validateEventoCancelRoot(doc *Document) error {
	if doc.EventoCancel == nil {
		return nil
	}
	inf := doc.EventoCancel.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoEntregaRoot(doc *Document) error {
	if doc.EventoEntrega == nil {
		return nil
	}
	inf := doc.EventoEntrega.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoCancEntregaRoot(doc *Document) error {
	if doc.EventoCancEntrega == nil {
		return nil
	}
	inf := doc.EventoCancEntrega.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoCCeRoot(doc *Document) error {
	if doc.EventoCCe == nil {
		return nil
	}
	inf := doc.EventoCCe.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoEPECRoot(doc *Document) error {
	if doc.EventoEPEC == nil {
		return nil
	}
	inf := doc.EventoEPEC.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoAtorInteressadoRoot(doc *Document) error {
	if doc.EventoAtorInteressado == nil {
		return nil
	}
	inf := doc.EventoAtorInteressado.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoMDERoot(doc *Document) error {
	if doc.EventoMDE == nil {
		return nil
	}
	inf := doc.EventoMDE.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoInsucessoRoot(doc *Document) error {
	if doc.EventoInsucesso == nil {
		return nil
	}
	inf := doc.EventoInsucesso.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoCancInsucessoRoot(doc *Document) error {
	if doc.EventoCancInsucesso == nil {
		return nil
	}
	inf := doc.EventoCancInsucesso.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEventoGenericoRoot(doc *Document) error {
	if doc.EventoGenerico == nil {
		return nil
	}
	inf := doc.EventoGenerico.InfEvento
	if inf == nil {
		return validateEventoFields(true, "", true)
	}
	return validateEventoFields(false, inf.ChNFe, inf.DetEvento == nil)
}

func validateEnvEventoRoot(doc *Document) error {
	if doc.EnvEvento == nil {
		return nil
	}
	if doc.EnvEvento.IdLote == "" {
		return errors.New("parse nfe: missing idLote")
	}
	if len(doc.EnvEvento.Evento) == 0 {
		return errors.New("parse nfe: missing evento")
	}
	return nil
}

func validateRetEnvEventoRoot(doc *Document) error {
	if doc.RetEnvEvento == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetEnvEvento.TpAmb),
		missing("cStat", doc.RetEnvEvento.CStat),
	)
}

func validateProcEventoNFeRoot(doc *Document) error {
	if doc.ProcEventoNFe == nil {
		return nil
	}
	if doc.ProcEventoNFe.Evento == nil {
		return errors.New("parse nfe: missing evento")
	}
	if doc.ProcEventoNFe.RetEvento == nil {
		return errors.New("parse nfe: missing retEvento")
	}
	return nil
}

func validateConsSitNFeRoot(doc *Document) error {
	if doc.ConsSitNFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.ConsSitNFe.TpAmb),
		missing("xServ", doc.ConsSitNFe.XServ),
		missing("chNFe", doc.ConsSitNFe.ChNFe),
	)
}

func validateRetConsSitNFeRoot(doc *Document) error {
	if doc.RetConsSitNFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsSitNFe.TpAmb),
		missing("cStat", doc.RetConsSitNFe.CStat),
		missing("cUF", doc.RetConsSitNFe.CUF),
		missing("chNFe", doc.RetConsSitNFe.ChNFe),
	)
}

func validateConsStatServRoot(doc *Document) error {
	if doc.ConsStatServ == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.ConsStatServ.TpAmb),
		missing("cUF", doc.ConsStatServ.CUF),
		missing("xServ", doc.ConsStatServ.XServ),
	)
}

func validateRetConsStatServRoot(doc *Document) error {
	if doc.RetConsStatServ == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsStatServ.TpAmb),
		missing("cStat", doc.RetConsStatServ.CStat),
		missing("cUF", doc.RetConsStatServ.CUF),
		missing("dhRecbto", doc.RetConsStatServ.DhRecbto),
	)
}

func validateInutNFeRoot(doc *Document) error {
	if doc.InutNFe == nil {
		return nil
	}
	if doc.InutNFe.InfInut == nil {
		return errors.New("parse nfe: missing infInut")
	}
	if doc.InutNFe.InfInut.TpAmb == "" {
		return errors.New("parse nfe: missing tpAmb")
	}
	if doc.InutNFe.InfInut.CNPJ == "" {
		return errors.New("parse nfe: missing CNPJ")
	}
	return nil
}

func validateRetInutNFeRoot(doc *Document) error {
	if doc.RetInutNFe == nil {
		return nil
	}
	if doc.RetInutNFe.InfInut == nil {
		return errors.New("parse nfe: missing infInut")
	}
	if doc.RetInutNFe.InfInut.TpAmb == "" {
		return errors.New("parse nfe: missing tpAmb")
	}
	if doc.RetInutNFe.InfInut.CNPJ == nil || *doc.RetInutNFe.InfInut.CNPJ == "" {
		return errors.New("parse nfe: missing CNPJ")
	}
	return nil
}

func validateProcInutNFeRoot(doc *Document) error {
	if doc.ProcInutNFe == nil {
		return nil
	}
	if doc.ProcInutNFe.InutNFe == nil {
		return errors.New("parse nfe: missing inutNFe")
	}
	if doc.ProcInutNFe.RetInutNFe == nil {
		return errors.New("parse nfe: missing retInutNFe")
	}
	return nil
}

func validateDistDFeIntRoot(doc *Document) error {
	if doc.DistDFeInt == nil {
		return nil
	}
	if doc.DistDFeInt.TpAmb == "" {
		return errors.New("parse nfe: missing tpAmb")
	}
	if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
		return errors.New("parse nfe: missing dist document")
	}
	if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil && doc.DistDFeInt.ConsChNFe == nil {
		return errors.New("parse nfe: missing dist query")
	}
	return nil
}

func validateRetDistDFeIntRoot(doc *Document) error {
	if doc.RetDistDFeInt == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetDistDFeInt.TpAmb),
		missing("cStat", doc.RetDistDFeInt.CStat),
		missing("ultNSU", doc.RetDistDFeInt.UltNSU),
		missing("maxNSU", doc.RetDistDFeInt.MaxNSU),
	)
}

func validateResNFeRoot(doc *Document) error {
	if doc.ResNFe == nil {
		return nil
	}
	return firstMissing(
		missing("chNFe", doc.ResNFe.ChNFe),
		missing("xNome", doc.ResNFe.XNome),
	)
}

func validateResEventoRoot(doc *Document) error {
	if doc.ResEvento == nil {
		return nil
	}
	return firstMissing(
		missing("chNFe", doc.ResEvento.ChNFe),
		missing("tpEvento", doc.ResEvento.TpEvento),
	)
}

func activeRootCount(doc *Document) int {
	count := 0
	for _, ok := range []bool{
		doc.NFe != nil,
		doc.EnviNFe != nil,
		doc.RetEnviNFe != nil,
		doc.ConsReciNFe != nil,
		doc.RetConsReciNFe != nil,
		doc.EventoCancel != nil,
		doc.EventoEntrega != nil,
		doc.EventoCancEntrega != nil,
		doc.EventoCCe != nil,
		doc.EventoEPEC != nil,
		doc.EventoAtorInteressado != nil,
		doc.EventoMDE != nil,
		doc.EventoInsucesso != nil,
		doc.EventoCancInsucesso != nil,
		doc.EventoGenerico != nil,
		doc.EnvEvento != nil,
		doc.RetEnvEvento != nil,
		doc.ProcEventoNFe != nil,
		doc.ConsSitNFe != nil,
		doc.RetConsSitNFe != nil,
		doc.ConsStatServ != nil,
		doc.RetConsStatServ != nil,
		doc.InutNFe != nil,
		doc.RetInutNFe != nil,
		doc.ProcInutNFe != nil,
		doc.DistDFeInt != nil,
		doc.RetDistDFeInt != nil,
		doc.ResNFe != nil,
		doc.ResEvento != nil,
	} {
		if ok {
			count++
		}
	}
	return count
}
