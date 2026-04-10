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
	type bareNFe struct {
		XMLName xml.Name `xml:"NFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TNFe
	}
	return xmlutil.EncodeCanonical(e, bareNFe{
		XMLName: xml.Name{Local: "NFe"},
		XMLNS:   namespace,
		TNFe:    invoice,
	})
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
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"enviNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TEnviNFe
	}{
		XMLName:  xml.Name{Local: "enviNFe"},
		XMLNS:    namespace,
		TEnviNFe: d.EnviNFe,
	})
}

func encodeRetEnviNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retEnviNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TRetEnviNFe
	}{
		XMLName:     xml.Name{Local: "retEnviNFe"},
		XMLNS:       namespace,
		TRetEnviNFe: d.RetEnviNFe,
	})
}

func encodeConsReciNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"consReciNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TConsReciNFe
	}{
		XMLName:      xml.Name{Local: "consReciNFe"},
		XMLNS:        namespace,
		TConsReciNFe: d.ConsReciNFe,
	})
}

func encodeRetConsReciNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retConsReciNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TRetConsReciNFe
	}{
		XMLName:         xml.Name{Local: "retConsReciNFe"},
		XMLNS:           namespace,
		TRetConsReciNFe: d.RetConsReciNFe,
	})
}

func encodeEnvEvento(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"envEvento"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*genericSchema.TEnvEvento
	}{
		XMLName:    xml.Name{Local: "envEvento"},
		XMLNS:      namespace,
		TEnvEvento: d.EnvEvento,
	})
}

func encodeRetEnvEvento(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retEnvEvento"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*genericSchema.TRetEnvEvento
	}{
		XMLName:       xml.Name{Local: "retEnvEvento"},
		XMLNS:         namespace,
		TRetEnvEvento: d.RetEnvEvento,
	})
}

func encodeProcEventoNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"procEventoNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*genericSchema.TProcEvento
	}{
		XMLName:     xml.Name{Local: "procEventoNFe"},
		XMLNS:       namespace,
		TProcEvento: d.ProcEventoNFe,
	})
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
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"consStatServ"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*statusSchema.TConsStatServ
	}{
		XMLName:       xml.Name{Local: "consStatServ"},
		XMLNS:         namespace,
		TConsStatServ: d.ConsStatServ,
	})
}

func encodeRetConsStatServ(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retConsStatServ"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*statusSchema.TRetConsStatServ
	}{
		XMLName:          xml.Name{Local: "retConsStatServ"},
		XMLNS:            namespace,
		TRetConsStatServ: d.RetConsStatServ,
	})
}

func encodeInutNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"inutNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*inutSchema.TInutNFe
	}{
		XMLName:  xml.Name{Local: "inutNFe"},
		XMLNS:    namespace,
		TInutNFe: d.InutNFe,
	})
}

func encodeRetInutNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retInutNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*inutSchema.TRetInutNFe
	}{
		XMLName:     xml.Name{Local: "retInutNFe"},
		XMLNS:       namespace,
		TRetInutNFe: d.RetInutNFe,
	})
}

func encodeProcInutNFe(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"procInutNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*inutSchema.TProcInutNFe
	}{
		XMLName:      xml.Name{Local: "procInutNFe"},
		XMLNS:        namespace,
		TProcInutNFe: d.ProcInutNFe,
	})
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
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"resNFe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*distSchema.TAnonComplexResNFe1
	}{
		XMLName:             xml.Name{Local: "resNFe"},
		XMLNS:               namespace,
		TAnonComplexResNFe1: d.ResNFe,
	})
}

func encodeResEvento(e *xml.Encoder, d *Document) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"resEvento"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*distSchema.TAnonComplexResEvento1
	}{
		XMLName:                xml.Name{Local: "resEvento"},
		XMLNS:                  namespace,
		TAnonComplexResEvento1: d.ResEvento,
	})
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

	switch rootName {
	case nfeProc:
		var parsed struct {
			VersaoAttr string           `xml:"versao,attr"`
			NFe        *schema.TNFe     `xml:"NFe"`
			ProtNFe    *schema.TProtNFe `xml:"protNFe"`
		}
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode nfeProc: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			NFe:        parsed.NFe,
			ProtNFe:    parsed.ProtNFe,
			RootName:   nfeProc,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "NFe":
		var invoice schema.TNFe
		if err := xml.Unmarshal(data, &invoice); err != nil {
			return nil, fmt.Errorf("parse nfe: decode NFe: %w", err)
		}
		doc := &Document{
			VersaoAttr: versionFromNFe(&invoice),
			NFe:        &invoice,
			RootName:   "NFe",
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "enviNFe":
		var parsed schema.TEnviNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode enviNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EnviNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "retEnviNFe":
		var parsed schema.TRetEnviNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retEnviNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEnviNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "consReciNFe":
		var parsed schema.TConsReciNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode consReciNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsReciNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "retConsReciNFe":
		var parsed schema.TRetConsReciNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retConsReciNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsReciNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "evento":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse nfe: decode evento head: %w", err)
		}
		switch tpEvento {
		case "110111":
			var event cancelSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cancel event: %w", err)
			}
			doc := &Document{
				VersaoAttr:   event.VersaoAttr,
				EventoCancel: &event,
				RootName:     rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110130":
			var event entregaSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode entrega event: %w", err)
			}
			doc := &Document{
				VersaoAttr:    event.VersaoAttr,
				EventoEntrega: &event,
				RootName:      rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110131":
			var event cancelEntregaSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cancel entrega event: %w", err)
			}
			doc := &Document{
				VersaoAttr:        event.VersaoAttr,
				EventoCancEntrega: &event,
				RootName:          rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110110":
			var event cceSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cce event: %w", err)
			}
			doc := &Document{
				VersaoAttr: event.VersaoAttr,
				EventoCCe:  &event,
				RootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110140":
			var event epecSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode epec event: %w", err)
			}
			doc := &Document{
				VersaoAttr: event.VersaoAttr,
				EventoEPEC: &event,
				RootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110150":
			var event atorSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode ator interessado event: %w", err)
			}
			doc := &Document{
				VersaoAttr:            event.VersaoAttr,
				EventoAtorInteressado: &event,
				RootName:              rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110192":
			var event insucessoSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode insucesso event: %w", err)
			}
			doc := &Document{
				VersaoAttr:      event.VersaoAttr,
				EventoInsucesso: &event,
				RootName:        rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110193":
			var event insucessoCancelSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cancel insucesso event: %w", err)
			}
			doc := &Document{
				VersaoAttr:          event.VersaoAttr,
				EventoCancInsucesso: &event,
				RootName:            rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "210200", "210210", "210220", "210240":
			var event mdeSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode mde event: %w", err)
			}
			doc := &Document{
				VersaoAttr: event.VersaoAttr,
				EventoMDE:  &event,
				RootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		default:
			var event genericSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode generic event: %w", err)
			}
			doc := &Document{
				VersaoAttr:     event.VersaoAttr,
				EventoGenerico: &event,
				RootName:       rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		}
	case "envEvento":
		var parsed genericSchema.TEnvEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode envEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EnvEvento: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retEnvEvento":
		var parsed genericSchema.TRetEnvEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retEnvEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEnvEvento: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "procEventoNFe":
		var parsed genericSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode procEventoNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consSitNFe":
		var parsed consSchema.TConsSitNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode consSitNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsSitNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsSitNFe":
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
		doc := &Document{
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
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consStatServ":
		var parsed statusSchema.TConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode consStatServ: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsStatServ: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsStatServ":
		var parsed statusSchema.TRetConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retConsStatServ: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServ: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "inutNFe":
		var parsed inutSchema.TInutNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode inutNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, InutNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retInutNFe":
		var parsed inutSchema.TRetInutNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retInutNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetInutNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "procInutNFe":
		var parsed inutSchema.TProcInutNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode procInutNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcInutNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "distDFeInt":
		var parsed distSchema.TAnonComplexDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode distDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retDistDFeInt":
		var parsed distSchema.TAnonComplexRetDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retDistDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "resNFe":
		var parsed distSchema.TAnonComplexResNFe1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode resNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ResNFe: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "resEvento":
		var parsed distSchema.TAnonComplexResEvento1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode resEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ResEvento: &parsed, RootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse nfe: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse nfe: unsupported root element %q", rootName)
	}
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

func validateDocument(doc *Document) error {
	count := 0

	if doc.RootName == "NFe" || doc.RootName == nfeProc || (doc.RootName == "" && doc.ProtNFe != nil) {
		if doc.NFe == nil {
			return errors.New("parse nfe: missing NFe")
		}
	}

	if doc.NFe != nil {
		count++
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
	}

	if doc.EnviNFe != nil {
		count++
		if doc.EnviNFe.IdLote == "" {
			return errors.New("parse nfe: missing idLote")
		}
		if len(doc.EnviNFe.NFe) == 0 {
			return errors.New("parse nfe: missing NFe")
		}
	}

	if doc.RetEnviNFe != nil {
		count++
		if doc.RetEnviNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetEnviNFe.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetEnviNFe.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
	}

	if doc.ConsReciNFe != nil {
		count++
		if doc.ConsReciNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.ConsReciNFe.NRec == "" {
			return errors.New("parse nfe: missing nRec")
		}
	}

	if doc.RetConsReciNFe != nil {
		count++
		if doc.RetConsReciNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetConsReciNFe.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetConsReciNFe.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
	}

	if doc.EventoCancel != nil {
		count++
		if doc.EventoCancel.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCancel.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCancel.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoEntrega != nil {
		count++
		if doc.EventoEntrega.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoEntrega.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoEntrega.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoCancEntrega != nil {
		count++
		if doc.EventoCancEntrega.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCancEntrega.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCancEntrega.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoCCe != nil {
		count++
		if doc.EventoCCe.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCCe.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCCe.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoEPEC != nil {
		count++
		if doc.EventoEPEC.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoEPEC.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoEPEC.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoAtorInteressado != nil {
		count++
		if doc.EventoAtorInteressado.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoAtorInteressado.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoAtorInteressado.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoMDE != nil {
		count++
		if doc.EventoMDE.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoMDE.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoMDE.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoInsucesso != nil {
		count++
		if doc.EventoInsucesso.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoInsucesso.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoInsucesso.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoCancInsucesso != nil {
		count++
		if doc.EventoCancInsucesso.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCancInsucesso.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCancInsucesso.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoGenerico != nil {
		count++
		if doc.EventoGenerico.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoGenerico.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoGenerico.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EnvEvento != nil {
		count++
		if doc.EnvEvento.IdLote == "" {
			return errors.New("parse nfe: missing idLote")
		}
		if len(doc.EnvEvento.Evento) == 0 {
			return errors.New("parse nfe: missing evento")
		}
	}

	if doc.RetEnvEvento != nil {
		count++
		if doc.RetEnvEvento.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetEnvEvento.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
	}

	if doc.ProcEventoNFe != nil {
		count++
		if doc.ProcEventoNFe.Evento == nil {
			return errors.New("parse nfe: missing evento")
		}
		if doc.ProcEventoNFe.RetEvento == nil {
			return errors.New("parse nfe: missing retEvento")
		}
	}

	if doc.ConsSitNFe != nil {
		count++
		if doc.ConsSitNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.ConsSitNFe.XServ == "" {
			return errors.New("parse nfe: missing xServ")
		}
		if doc.ConsSitNFe.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
	}

	if doc.RetConsSitNFe != nil {
		count++
		if doc.RetConsSitNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetConsSitNFe.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetConsSitNFe.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
		if doc.RetConsSitNFe.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
	}

	if doc.ConsStatServ != nil {
		count++
		if doc.ConsStatServ.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.ConsStatServ.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
		if doc.ConsStatServ.XServ == "" {
			return errors.New("parse nfe: missing xServ")
		}
	}

	if doc.RetConsStatServ != nil {
		count++
		if doc.RetConsStatServ.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetConsStatServ.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetConsStatServ.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
		if doc.RetConsStatServ.DhRecbto == "" {
			return errors.New("parse nfe: missing dhRecbto")
		}
	}

	if doc.InutNFe != nil {
		count++
		if doc.InutNFe.InfInut == nil {
			return errors.New("parse nfe: missing infInut")
		}
		if doc.InutNFe.InfInut.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.InutNFe.InfInut.CNPJ == "" {
			return errors.New("parse nfe: missing CNPJ")
		}
	}

	if doc.RetInutNFe != nil {
		count++
		if doc.RetInutNFe.InfInut == nil {
			return errors.New("parse nfe: missing infInut")
		}
		if doc.RetInutNFe.InfInut.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetInutNFe.InfInut.CNPJ == nil || *doc.RetInutNFe.InfInut.CNPJ == "" {
			return errors.New("parse nfe: missing CNPJ")
		}
	}

	if doc.ProcInutNFe != nil {
		count++
		if doc.ProcInutNFe.InutNFe == nil {
			return errors.New("parse nfe: missing inutNFe")
		}
		if doc.ProcInutNFe.RetInutNFe == nil {
			return errors.New("parse nfe: missing retInutNFe")
		}
	}

	if doc.DistDFeInt != nil {
		count++
		if doc.DistDFeInt.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
			return errors.New("parse nfe: missing dist document")
		}
		if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil && doc.DistDFeInt.ConsChNFe == nil {
			return errors.New("parse nfe: missing dist query")
		}
	}

	if doc.RetDistDFeInt != nil {
		count++
		if doc.RetDistDFeInt.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetDistDFeInt.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetDistDFeInt.UltNSU == "" {
			return errors.New("parse nfe: missing ultNSU")
		}
		if doc.RetDistDFeInt.MaxNSU == "" {
			return errors.New("parse nfe: missing maxNSU")
		}
	}

	if doc.ResNFe != nil {
		count++
		if doc.ResNFe.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.ResNFe.XNome == "" {
			return errors.New("parse nfe: missing xNome")
		}
	}

	if doc.ResEvento != nil {
		count++
		if doc.ResEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.ResEvento.TpEvento == "" {
			return errors.New("parse nfe: missing tpEvento")
		}
	}

	if count != 1 {
		return errors.New("parse nfe: document must contain exactly one supported root")
	}
	return nil
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
