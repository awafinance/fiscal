package mdfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"

	distSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v1_0/dist_dfe"
	consNaoEncSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/cons_nao_enc"
	consReciSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/cons_reci"
	consultaDFESchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/consulta_dfe"
	consSitSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/consulta_situacao"
	distMDFeSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/dist_mdfe"
	eventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento"
	alteracaoPagtoServEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_alteracao_pagto_serv"
	cancelEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_cancel"
	confirmaServEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_confirma_serv"
	encEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_enc"
	incCondutorEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_inc_condutor"
	inclusaoDFeEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_inclusao_dfe"
	pagtoOperEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_pagto_oper"
	mdfeSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/mdfe"
	statusSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/status_servico"
	"github.com/awafinance/fiscal/internal/xmlutil"
	"github.com/awafinance/fiscal/pkg/fiscalerr"
)

const namespace = "http://www.portalfiscal.inf.br/mdfe"

var errSingleRoot = errors.New("marshal mdfe: document must contain exactly one supported root")

var parsersByRoot = map[string]func([]byte, string) (*Document, error){
	"MDFe":                parseMDFe,
	"mdfeProc":            parseMDFeProc,
	"enviMDFe":            parseEnviMDFe,
	"retEnviMDFe":         parseRetEnviMDFe,
	"retMDFe":             parseRetMDFe,
	"consMDFeNaoEnc":      parseConsNaoEnc,
	"retConsMDFeNaoEnc":   parseRetConsNaoEnc,
	"consReciMDFe":        parseConsReciMDFe,
	"retConsReciMDFe":     parseRetConsReciMDFe,
	"consSitMDFe":         parseConsSitMDFe,
	"retConsSitMDFe":      parseRetConsSitMDFe,
	"consStatServMDFe":    parseConsStatServMDFe,
	"retConsStatServMDFe": parseRetConsStatServMDFe,
	"distDFeInt":          parseDistDFeInt,
	"retDistDFeInt":       parseRetDistDFeInt,
	"distMDFe":            parseDistMDFe,
	"retDistMDFe":         parseRetDistMDFe,
	"mdfeConsultaDFe":     parseMDFeConsultaDFe,
	"retMDFeConsultaDFe":  parseRetMDFeConsultaDFe,
	"eventoMDFe":          func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseEventDocument) },
	"retEventoMDFe":       func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseRetEventDocument) },
	"procEventoMDFe":      func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseProcEventDocument) },
}

type Document struct {
	VersaoAttr                       string                                     `json:"versao,omitempty"`
	MDFe                             *mdfeSchema.TMDFe                          `json:"MDFe,omitempty"`
	MDFeProc                         *mdfeSchema.TAnonComplexMdfeProc1          `json:"mdfeProc,omitempty"`
	EnviMDFe                         *mdfeSchema.TEnviMDFe                      `json:"enviMDFe,omitempty"`
	RetEnviMDFe                      *mdfeSchema.TRetEnviMDFe                   `json:"retEnviMDFe,omitempty"`
	RetMDFe                          *mdfeSchema.TRetMDFe                       `json:"retMDFe,omitempty"`
	ConsNaoEnc                       *consNaoEncSchema.TConsMDFeNaoEnc          `json:"consMDFeNaoEnc,omitempty"`
	RetConsNaoEnc                    *consNaoEncSchema.TRetConsMDFeNaoEnc       `json:"retConsMDFeNaoEnc,omitempty"`
	ConsReciMDFe                     *consReciSchema.TConsReciMDFe              `json:"consReciMDFe,omitempty"`
	RetConsReciMDFe                  *consReciSchema.TRetConsReciMDFe           `json:"retConsReciMDFe,omitempty"`
	ConsSitMDFe                      *consSitSchema.TConsSitMDFe                `json:"consSitMDFe,omitempty"`
	RetConsSitMDFe                   *consSitSchema.TRetConsSitMDFe             `json:"retConsSitMDFe,omitempty"`
	ConsStatServMDFe                 *statusSchema.TConsStatServ                `json:"consStatServMDFe,omitempty"`
	RetConsStatServMDFe              *statusSchema.TRetConsStatServ             `json:"retConsStatServMDFe,omitempty"`
	EventoMDFe                       *eventSchema.TEvento                       `json:"eventoMDFe,omitempty"`
	RetEventoMDFe                    *eventSchema.TRetEvento                    `json:"retEventoMDFe,omitempty"`
	ProcEventoMDFe                   *eventSchema.TProcEvento                   `json:"procEventoMDFe,omitempty"`
	EventoCancMDFe                   *cancelEventSchema.TEvento                 `json:"eventoCancMDFe,omitempty"`
	RetEventoCancMDFe                *cancelEventSchema.TRetEvento              `json:"retEventoCancMDFe,omitempty"`
	ProcEventoCancMDFe               *cancelEventSchema.TProcEvento             `json:"procEventoCancMDFe,omitempty"`
	EventoEncMDFe                    *encEventSchema.TEvento                    `json:"eventoEncMDFe,omitempty"`
	RetEventoEncMDFe                 *encEventSchema.TRetEvento                 `json:"retEventoEncMDFe,omitempty"`
	ProcEventoEncMDFe                *encEventSchema.TProcEvento                `json:"procEventoEncMDFe,omitempty"`
	EventoIncCondutorMDFe            *incCondutorEventSchema.TEvento            `json:"eventoIncCondutorMDFe,omitempty"`
	RetEventoIncCondutorMDFe         *incCondutorEventSchema.TRetEvento         `json:"retEventoIncCondutorMDFe,omitempty"`
	ProcEventoIncCondutorMDFe        *incCondutorEventSchema.TProcEvento        `json:"procEventoIncCondutorMDFe,omitempty"`
	EventoInclusaoDFeMDFe            *inclusaoDFeEventSchema.TEvento            `json:"eventoInclusaoDFeMDFe,omitempty"`
	RetEventoInclusaoDFeMDFe         *inclusaoDFeEventSchema.TRetEvento         `json:"retEventoInclusaoDFeMDFe,omitempty"`
	ProcEventoInclusaoDFeMDFe        *inclusaoDFeEventSchema.TProcEvento        `json:"procEventoInclusaoDFeMDFe,omitempty"`
	EventoPagtoOperMDFe              *pagtoOperEventSchema.TEvento              `json:"eventoPagtoOperMDFe,omitempty"`
	RetEventoPagtoOperMDFe           *pagtoOperEventSchema.TRetEvento           `json:"retEventoPagtoOperMDFe,omitempty"`
	ProcEventoPagtoOperMDFe          *pagtoOperEventSchema.TProcEvento          `json:"procEventoPagtoOperMDFe,omitempty"`
	EventoAlteracaoPagtoServMDFe     *alteracaoPagtoServEventSchema.TEvento     `json:"eventoAlteracaoPagtoServMDFe,omitempty"`
	RetEventoAlteracaoPagtoServMDFe  *alteracaoPagtoServEventSchema.TRetEvento  `json:"retEventoAlteracaoPagtoServMDFe,omitempty"`
	ProcEventoAlteracaoPagtoServMDFe *alteracaoPagtoServEventSchema.TProcEvento `json:"procEventoAlteracaoPagtoServMDFe,omitempty"`
	EventoConfirmaServMDFe           *confirmaServEventSchema.TEvento           `json:"eventoConfirmaServMDFe,omitempty"`
	RetEventoConfirmaServMDFe        *confirmaServEventSchema.TRetEvento        `json:"retEventoConfirmaServMDFe,omitempty"`
	ProcEventoConfirmaServMDFe       *confirmaServEventSchema.TProcEvento       `json:"procEventoConfirmaServMDFe,omitempty"`
	DistDFeInt                       *distSchema.TAnonComplexDistDFeInt1        `json:"distDFeInt,omitempty"`
	RetDistDFeInt                    *distSchema.TAnonComplexRetDistDFeInt1     `json:"retDistDFeInt,omitempty"`
	DistMDFe                         *distMDFeSchema.TDistDFe                   `json:"distMDFe,omitempty"`
	RetDistMDFe                      *distMDFeSchema.TRetDistDFe                `json:"retDistMDFe,omitempty"`
	MDFeConsultaDFe                  *consultaDFESchema.TMDFeConsultaDFe        `json:"mdfeConsultaDFe,omitempty"`
	RetMDFeConsultaDFe               *consultaDFESchema.TRetMDFeConsultaDFe     `json:"retMDFeConsultaDFe,omitempty"`
	RootName                         string                                     `json:"rootName,omitempty"`
}

var marshalersByRoot = map[string]func(*xml.Encoder, *Document) error{
	"MDFe":                marshalMDFe,
	"":                    marshalMDFe,
	"mdfeProc":            marshalMDFeProc,
	"enviMDFe":            marshalEnviMDFe,
	"retEnviMDFe":         marshalRetEnviMDFe,
	"retMDFe":             marshalRetMDFe,
	"consMDFeNaoEnc":      marshalConsNaoEnc,
	"retConsMDFeNaoEnc":   marshalRetConsNaoEnc,
	"consReciMDFe":        marshalConsReciMDFe,
	"retConsReciMDFe":     marshalRetConsReciMDFe,
	"consSitMDFe":         marshalConsSitMDFe,
	"retConsSitMDFe":      marshalRetConsSitMDFe,
	"consStatServMDFe":    marshalConsStatServMDFe,
	"retConsStatServMDFe": marshalRetConsStatServMDFe,
	"eventoMDFe":          marshalEventRoot,
	"retEventoMDFe":       marshalRetEventRoot,
	"procEventoMDFe":      marshalProcEventRoot,
	"distDFeInt":          marshalDistDFeInt,
	"retDistDFeInt":       marshalRetDistDFeInt,
	"distMDFe":            marshalDistMDFe,
	"retDistMDFe":         marshalRetDistMDFe,
	"mdfeConsultaDFe":     marshalMDFeConsultaDFe,
	"retMDFeConsultaDFe":  marshalRetMDFeConsultaDFe,
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	if fn, ok := marshalersByRoot[d.RootName]; ok {
		return fn(e, d)
	}
	return errSingleRoot
}

func marshalMDFe(e *xml.Encoder, d *Document) error {
	if d.MDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                             `xml:"MDFe"`
		XMLNS       string                               `xml:"xmlns,attr,omitempty"`
		InfMDFe     *mdfeSchema.TAnonComplexInfMDFe1     `xml:"infMDFe"`
		InfMDFeSupl *mdfeSchema.TAnonComplexInfMDFeSupl1 `xml:"infMDFeSupl,omitempty"`
		DsSignature *mdfeSchema.SignatureType            `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "MDFe"},
		XMLNS:       namespace,
		InfMDFe:     d.MDFe.InfMDFe,
		InfMDFeSupl: d.MDFe.InfMDFeSupl,
		DsSignature: d.MDFe.DsSignature,
	})
}

func marshalMDFeProc(e *xml.Encoder, d *Document) error {
	if d.MDFeProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName           xml.Name              `xml:"mdfeProc"`
		XMLNS             string                `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string                `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string               `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string               `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string               `xml:"dhConexao,attr,omitempty"`
		MDFe              *mdfeSchema.TMDFe     `xml:"MDFe"`
		ProtMDFe          *mdfeSchema.TProtMDFe `xml:"protMDFe"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "mdfeProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.MDFeProc.VersaoAttr),
		IpTransmissorAttr: d.MDFeProc.IpTransmissorAttr,
		NPortaConAttr:     d.MDFeProc.NPortaConAttr,
		DhConexaoAttr:     d.MDFeProc.DhConexaoAttr,
		MDFe:              d.MDFeProc.MDFe,
		ProtMDFe:          d.MDFeProc.ProtMDFe,
	})
}

func marshalEnviMDFe(e *xml.Encoder, d *Document) error {
	if d.EnviMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "enviMDFe", namespace, d.EnviMDFe)
}

func marshalRetEnviMDFe(e *xml.Encoder, d *Document) error {
	if d.RetEnviMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retEnviMDFe", namespace, d.RetEnviMDFe)
}

func marshalRetMDFe(e *xml.Encoder, d *Document) error {
	if d.RetMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retMDFe", namespace, d.RetMDFe)
}

func marshalConsNaoEnc(e *xml.Encoder, d *Document) error {
	if d.ConsNaoEnc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName    xml.Name                  `xml:"consMDFeNaoEnc"`
		XMLNS      string                    `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                    `xml:"versao,attr,omitempty"`
		TpAmb      string                    `xml:"tpAmb"`
		XServ      *consNaoEncSchema.TString `xml:"xServ,omitempty"`
		CNPJ       *string                   `xml:"CNPJ,omitempty"`
		CPF        *string                   `xml:"CPF,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "consMDFeNaoEnc"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.ConsNaoEnc.VersaoAttr),
		TpAmb:      d.ConsNaoEnc.TpAmb,
		XServ:      d.ConsNaoEnc.XServ,
		CNPJ:       d.ConsNaoEnc.CNPJ,
		CPF:        d.ConsNaoEnc.CPF,
	})
}

func marshalRetConsNaoEnc(e *xml.Encoder, d *Document) error {
	if d.RetConsNaoEnc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retConsMDFeNaoEnc", namespace, d.RetConsNaoEnc)
}

func marshalConsReciMDFe(e *xml.Encoder, d *Document) error {
	if d.ConsReciMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName    xml.Name `xml:"consReciMDFe"`
		XMLNS      string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr string   `xml:"versao,attr,omitempty"`
		TpAmb      string   `xml:"tpAmb"`
		NRec       string   `xml:"nRec"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "consReciMDFe"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.ConsReciMDFe.VersaoAttr),
		TpAmb:      d.ConsReciMDFe.TpAmb,
		NRec:       d.ConsReciMDFe.NRec,
	})
}

func marshalRetConsReciMDFe(e *xml.Encoder, d *Document) error {
	if d.RetConsReciMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retConsReciMDFe", namespace, d.RetConsReciMDFe)
}

func marshalConsSitMDFe(e *xml.Encoder, d *Document) error {
	if d.ConsSitMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "consSitMDFe", namespace, d.ConsSitMDFe)
}

func marshalRetConsSitMDFe(e *xml.Encoder, d *Document) error {
	if d.RetConsSitMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retConsSitMDFe", namespace, d.RetConsSitMDFe)
}

func marshalConsStatServMDFe(e *xml.Encoder, d *Document) error {
	if d.ConsStatServMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "consStatServMDFe", namespace, d.ConsStatServMDFe)
}

func marshalRetConsStatServMDFe(e *xml.Encoder, d *Document) error {
	if d.RetConsStatServMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retConsStatServMDFe", namespace, d.RetConsStatServMDFe)
}

func marshalDistDFeInt(e *xml.Encoder, d *Document) error {
	if d.DistDFeInt == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName    xml.Name                         `xml:"distDFeInt"`
		XMLNS      string                           `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                           `xml:"versao,attr,omitempty"`
		TpAmb      string                           `xml:"tpAmb"`
		CNPJ       *string                          `xml:"CNPJ,omitempty"`
		CPF        *string                          `xml:"CPF,omitempty"`
		DistNSU    *distSchema.TAnonComplexDistNSU1 `xml:"distNSU,omitempty"`
		ConsNSU    *distSchema.TAnonComplexConsNSU1 `xml:"consNSU,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "distDFeInt"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.DistDFeInt.VersaoAttr),
		TpAmb:      d.DistDFeInt.TpAmb,
		CNPJ:       d.DistDFeInt.CNPJ,
		CPF:        d.DistDFeInt.CPF,
		DistNSU:    d.DistDFeInt.DistNSU,
		ConsNSU:    d.DistDFeInt.ConsNSU,
	})
}

func marshalRetDistDFeInt(e *xml.Encoder, d *Document) error {
	if d.RetDistDFeInt == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
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

func marshalDistMDFe(e *xml.Encoder, d *Document) error {
	if d.DistMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "distMDFe", namespace, d.DistMDFe)
}

func marshalRetDistMDFe(e *xml.Encoder, d *Document) error {
	if d.RetDistMDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retDistMDFe", namespace, d.RetDistMDFe)
}

func marshalMDFeConsultaDFe(e *xml.Encoder, d *Document) error {
	if d.MDFeConsultaDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "mdfeConsultaDFe", namespace, d.MDFeConsultaDFe)
}

func marshalRetMDFeConsultaDFe(e *xml.Encoder, d *Document) error {
	if d.RetMDFeConsultaDFe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retMDFeConsultaDFe", namespace, d.RetMDFeConsultaDFe)
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("parse mdfe: %w", fiscalerr.ErrEmptyDocument)
	}

	RootName, rootErr := xmlutil.ParseRootName(data)
	if rootErr != nil && RootName == "" {
		return nil, fmt.Errorf("parse mdfe: read root: %w", rootErr)
	}

	if fn, ok := parsersByRoot[RootName]; ok {
		return fn(data, RootName)
	}
	if rootErr != nil {
		return nil, fmt.Errorf("parse mdfe: read root: %w", rootErr)
	}
	return nil, fmt.Errorf("parse mdfe: %w", &fiscalerr.UnsupportedRootError{Family: fiscalerr.MDFe, Root: RootName})
}

func ParseReader(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parse mdfe: read xml: %w", err)
	}
	return Parse(data)
}

func finalizeDoc(doc *Document) (*Document, error) {
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func parseMDFe(data []byte, rootName string) (*Document, error) {
	var parsed mdfeSchema.TMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode MDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: versionFromMDFe(&parsed), MDFe: &parsed, RootName: rootName})
}

func parseMDFeProc(data []byte, rootName string) (*Document, error) {
	var parsed mdfeSchema.TAnonComplexMdfeProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode mdfeProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, MDFeProc: &parsed, RootName: rootName})
}

func parseEnviMDFe(data []byte, rootName string) (*Document, error) {
	var parsed mdfeSchema.TEnviMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode enviMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EnviMDFe: &parsed, RootName: rootName})
}

func parseRetEnviMDFe(data []byte, rootName string) (*Document, error) {
	var parsed mdfeSchema.TRetEnviMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retEnviMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEnviMDFe: &parsed, RootName: rootName})
}

func parseRetMDFe(data []byte, rootName string) (*Document, error) {
	var parsed mdfeSchema.TRetMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetMDFe: &parsed, RootName: rootName})
}

func parseConsNaoEnc(data []byte, rootName string) (*Document, error) {
	var parsed consNaoEncSchema.TConsMDFeNaoEnc
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode consMDFeNaoEnc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsNaoEnc: &parsed, RootName: rootName})
}

func parseRetConsNaoEnc(data []byte, rootName string) (*Document, error) {
	var parsed consNaoEncSchema.TRetConsMDFeNaoEnc
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retConsMDFeNaoEnc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsNaoEnc: &parsed, RootName: rootName})
}

func parseConsReciMDFe(data []byte, rootName string) (*Document, error) {
	var parsed consReciSchema.TConsReciMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode consReciMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsReciMDFe: &parsed, RootName: rootName})
}

func parseRetConsReciMDFe(data []byte, rootName string) (*Document, error) {
	var parsed consReciSchema.TRetConsReciMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retConsReciMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsReciMDFe: &parsed, RootName: rootName})
}

func parseConsSitMDFe(data []byte, rootName string) (*Document, error) {
	var parsed consSitSchema.TConsSitMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode consSitMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsSitMDFe: &parsed, RootName: rootName})
}

func parseRetConsSitMDFe(data []byte, rootName string) (*Document, error) {
	var parsed consSitSchema.TRetConsSitMDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retConsSitMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsSitMDFe: &parsed, RootName: rootName})
}

func parseConsStatServMDFe(data []byte, rootName string) (*Document, error) {
	var parsed statusSchema.TConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode consStatServMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsStatServMDFe: &parsed, RootName: rootName})
}

func parseRetConsStatServMDFe(data []byte, rootName string) (*Document, error) {
	var parsed statusSchema.TRetConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retConsStatServMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServMDFe: &parsed, RootName: rootName})
}

func parseDistDFeInt(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexDistDFeInt1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode distDFeInt: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, RootName: rootName})
}

func parseRetDistDFeInt(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexRetDistDFeInt1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retDistDFeInt: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, RootName: rootName})
}

func parseDistMDFe(data []byte, rootName string) (*Document, error) {
	var parsed distMDFeSchema.TDistDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode distMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, DistMDFe: &parsed, RootName: rootName})
}

func parseRetDistMDFe(data []byte, rootName string) (*Document, error) {
	var parsed distMDFeSchema.TRetDistDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retDistMDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetDistMDFe: &parsed, RootName: rootName})
}

func parseMDFeConsultaDFe(data []byte, rootName string) (*Document, error) {
	var parsed consultaDFESchema.TMDFeConsultaDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode mdfeConsultaDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, MDFeConsultaDFe: &parsed, RootName: rootName})
}

func parseRetMDFeConsultaDFe(data []byte, rootName string) (*Document, error) {
	var parsed consultaDFESchema.TRetMDFeConsultaDFe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode retMDFeConsultaDFe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetMDFeConsultaDFe: &parsed, RootName: rootName})
}

func parseEventRoot(data []byte, rootName string, fn func([]byte, string, string) (*Document, error)) (*Document, error) {
	tpEvento, err := eventTypeFromXML(data)
	if err != nil {
		return nil, fmt.Errorf("parse mdfe: decode %s head: %w", rootName, err)
	}
	if tpEvento == "" {
		return nil, errors.New("parse mdfe: missing infEvento")
	}
	return fn(data, rootName, tpEvento)
}

func eventTypeFromXML(data []byte) (string, error) {
	var head struct {
		InfEvento struct {
			TpEvento string `xml:"tpEvento"`
		} `xml:"infEvento"`
		EventoMDFe struct {
			InfEvento struct {
				TpEvento string `xml:"tpEvento"`
			} `xml:"infEvento"`
		} `xml:"eventoMDFe"`
	}
	if err := xml.Unmarshal(data, &head); err != nil {
		return "", err
	}
	if head.InfEvento.TpEvento != "" {
		return head.InfEvento.TpEvento, nil
	}
	return head.EventoMDFe.InfEvento.TpEvento, nil
}

type mdfeEventRootKind uint8

const (
	mdfeSentEventRoot mdfeEventRootKind = iota
	mdfeRetEventRoot
	mdfeProcEventRoot
)

type mdfeEventSpec struct {
	eventTypes []string
	context    string
	eventField string

	eventTypeOf     reflect.Type
	retEventTypeOf  reflect.Type
	procEventTypeOf reflect.Type
}

var mdfeEventSpecs = []mdfeEventSpec{
	mdfeEvent[eventSchema.TEvento, eventSchema.TRetEvento, eventSchema.TProcEvento](nil, "generic", "EventoMDFe"),
	mdfeEvent[cancelEventSchema.TEvento, cancelEventSchema.TRetEvento, cancelEventSchema.TProcEvento]([]string{"110111"}, "cancelamento", "EventoCancMDFe"),
	mdfeEvent[encEventSchema.TEvento, encEventSchema.TRetEvento, encEventSchema.TProcEvento]([]string{"110112"}, "encerramento", "EventoEncMDFe"),
	mdfeEvent[incCondutorEventSchema.TEvento, incCondutorEventSchema.TRetEvento, incCondutorEventSchema.TProcEvento]([]string{"110114"}, "inclusao condutor", "EventoIncCondutorMDFe"),
	mdfeEvent[inclusaoDFeEventSchema.TEvento, inclusaoDFeEventSchema.TRetEvento, inclusaoDFeEventSchema.TProcEvento]([]string{"110115"}, "inclusao dfe", "EventoInclusaoDFeMDFe"),
	mdfeEvent[pagtoOperEventSchema.TEvento, pagtoOperEventSchema.TRetEvento, pagtoOperEventSchema.TProcEvento]([]string{"110116"}, "pagamento operacao", "EventoPagtoOperMDFe"),
	mdfeEvent[confirmaServEventSchema.TEvento, confirmaServEventSchema.TRetEvento, confirmaServEventSchema.TProcEvento]([]string{"110117"}, "confirma servico", "EventoConfirmaServMDFe"),
	mdfeEvent[alteracaoPagtoServEventSchema.TEvento, alteracaoPagtoServEventSchema.TRetEvento, alteracaoPagtoServEventSchema.TProcEvento]([]string{"110118"}, "alteracao pagamento servico", "EventoAlteracaoPagtoServMDFe"),
}

func mdfeEvent[E, R, P any](eventTypes []string, context, eventField string) mdfeEventSpec {
	return mdfeEventSpec{
		eventTypes:      eventTypes,
		context:         context,
		eventField:      eventField,
		eventTypeOf:     reflect.TypeFor[E](),
		retEventTypeOf:  reflect.TypeFor[R](),
		procEventTypeOf: reflect.TypeFor[P](),
	}
}

var rootValidators = []func(*Document) error{
	validateMDFeRootDoc,
	validateMDFeProcRoot,
	validateEnviMDFeRoot,
	validateRetEnviMDFeRoot,
	validateRetMDFeRoot,
	validateConsNaoEncRoot,
	validateRetConsNaoEncRoot,
	validateConsReciMDFeRoot,
	validateRetConsReciMDFeRoot,
	validateConsSitMDFeRoot,
	validateRetConsSitMDFeRoot,
	validateConsStatServMDFeRoot,
	validateRetConsStatServMDFeRoot,
	validateEventRoots,
	validateDistDFeIntRoot,
	validateRetDistDFeIntRoot,
	validateDistMDFeRoot,
	validateRetDistMDFeRoot,
	validateMDFeConsultaDFeRoot,
	validateRetMDFeConsultaDFeRoot,
}

func validateDocument(doc *Document) error {
	if activeRootCount(doc) != 1 {
		return errors.New("parse mdfe: document must contain exactly one supported root")
	}
	for _, v := range rootValidators {
		if err := v(doc); err != nil {
			return err
		}
	}
	return nil
}

func validateMDFeRootDoc(doc *Document) error {
	if doc.MDFe == nil {
		return nil
	}
	return validateMDFeRoot(doc.MDFe)
}

func validateMDFeProcRoot(doc *Document) error {
	if doc.MDFeProc == nil {
		return nil
	}
	if doc.MDFeProc.MDFe == nil {
		return errors.New("parse mdfe: missing MDFe")
	}
	if doc.MDFeProc.ProtMDFe == nil {
		return errors.New("parse mdfe: missing protMDFe")
	}
	return nil
}

func validateEnviMDFeRoot(doc *Document) error {
	if doc.EnviMDFe == nil {
		return nil
	}
	if doc.EnviMDFe.IdLote == "" {
		return errors.New("parse mdfe: missing idLote")
	}
	if doc.EnviMDFe.MDFe == nil {
		return errors.New("parse mdfe: missing MDFe")
	}
	return nil
}

func validateRetEnviMDFeRoot(doc *Document) error {
	if doc.RetEnviMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", valueOrEmpty(doc.RetEnviMDFe.TpAmb)),
		missing("cUF", doc.RetEnviMDFe.CUF),
		missing("cStat", doc.RetEnviMDFe.CStat),
	)
}

func validateRetMDFeRoot(doc *Document) error {
	if doc.RetMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", valueOrEmpty(doc.RetMDFe.TpAmb)),
		missing("cUF", doc.RetMDFe.CUF),
		missing("cStat", doc.RetMDFe.CStat),
	)
}

func validateConsNaoEncRoot(doc *Document) error {
	if doc.ConsNaoEnc == nil {
		return nil
	}
	if doc.ConsNaoEnc.TpAmb == "" {
		return errors.New("parse mdfe: missing tpAmb")
	}
	if doc.ConsNaoEnc.CNPJ == nil && doc.ConsNaoEnc.CPF == nil {
		return errors.New("parse mdfe: missing consult document")
	}
	return nil
}

func validateRetConsNaoEncRoot(doc *Document) error {
	if doc.RetConsNaoEnc == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsNaoEnc.TpAmb),
		missing("cUF", doc.RetConsNaoEnc.CUF),
		missing("cStat", doc.RetConsNaoEnc.CStat),
	)
}

func validateConsReciMDFeRoot(doc *Document) error {
	if doc.ConsReciMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.ConsReciMDFe.TpAmb),
		missing("nRec", doc.ConsReciMDFe.NRec),
	)
}

func validateRetConsReciMDFeRoot(doc *Document) error {
	if doc.RetConsReciMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsReciMDFe.TpAmb),
		missing("nRec", doc.RetConsReciMDFe.NRec),
		missing("cUF", doc.RetConsReciMDFe.CUF),
		missing("cStat", doc.RetConsReciMDFe.CStat),
	)
}

func validateConsSitMDFeRoot(doc *Document) error {
	if doc.ConsSitMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.ConsSitMDFe.TpAmb),
		missing("chMDFe", doc.ConsSitMDFe.ChMDFe),
	)
}

func validateRetConsSitMDFeRoot(doc *Document) error {
	if doc.RetConsSitMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsSitMDFe.TpAmb),
		missing("cUF", doc.RetConsSitMDFe.CUF),
		missing("cStat", doc.RetConsSitMDFe.CStat),
	)
}

func validateConsStatServMDFeRoot(doc *Document) error {
	if doc.ConsStatServMDFe == nil {
		return nil
	}
	return firstMissing(missing("tpAmb", doc.ConsStatServMDFe.TpAmb))
}

func validateRetConsStatServMDFeRoot(doc *Document) error {
	if doc.RetConsStatServMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsStatServMDFe.TpAmb),
		missing("cUF", doc.RetConsStatServMDFe.CUF),
		missing("cStat", doc.RetConsStatServMDFe.CStat),
		missing("dhRecbto", doc.RetConsStatServMDFe.DhRecbto),
	)
}

func validateDistDFeIntRoot(doc *Document) error {
	if doc.DistDFeInt == nil {
		return nil
	}
	if doc.DistDFeInt.TpAmb == "" {
		return errors.New("parse mdfe: missing tpAmb")
	}
	if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
		return errors.New("parse mdfe: missing dist document")
	}
	if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil {
		return errors.New("parse mdfe: missing dist query")
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

func validateDistMDFeRoot(doc *Document) error {
	if doc.DistMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.DistMDFe.TpAmb),
		missing("ultNSU", doc.DistMDFe.UltNSU),
	)
}

func validateRetDistMDFeRoot(doc *Document) error {
	if doc.RetDistMDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetDistMDFe.TpAmb),
		missing("cStat", doc.RetDistMDFe.CStat),
	)
}

func validateMDFeConsultaDFeRoot(doc *Document) error {
	if doc.MDFeConsultaDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.MDFeConsultaDFe.TpAmb),
		missing("chMDFe", doc.MDFeConsultaDFe.ChMDFe),
	)
}

func validateRetMDFeConsultaDFeRoot(doc *Document) error {
	if doc.RetMDFeConsultaDFe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetMDFeConsultaDFe.TpAmb),
		missing("cStat", doc.RetMDFeConsultaDFe.CStat),
	)
}

func missing(field, value string) error {
	if value == "" {
		return errors.New("parse mdfe: missing " + field)
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

func validateMDFeRoot(doc *mdfeSchema.TMDFe) error {
	if doc == nil || doc.InfMDFe == nil {
		return errors.New("parse mdfe: missing infMDFe")
	}
	if doc.InfMDFe.Ide == nil {
		return errors.New("parse mdfe: missing ide")
	}
	if doc.InfMDFe.Emit == nil {
		return errors.New("parse mdfe: missing emit")
	}
	if doc.InfMDFe.Emit.CNPJ == nil && doc.InfMDFe.Emit.CPF == nil {
		return errors.New("parse mdfe: missing emit document")
	}
	if doc.InfMDFe.InfModal == nil {
		return errors.New("parse mdfe: missing infModal")
	}
	return nil
}

func validateEventRoots(doc *Document) error {
	for i := range mdfeEventSpecs {
		spec := &mdfeEventSpecs[i]
		if err := validateMDFeEventField(doc, spec, mdfeSentEventRoot, validateMDFeSentEventRoot); err != nil {
			return err
		}
		if err := validateMDFeEventField(doc, spec, mdfeRetEventRoot, validateMDFeRetEventRoot); err != nil {
			return err
		}
		if err := validateMDFeProcEventField(doc, spec); err != nil {
			return err
		}
	}
	return nil
}

func validateMDFeEventoFields(isNil bool, chMDFe string, missingDetEvento bool) error {
	if isNil {
		return errors.New("parse mdfe: missing infEvento")
	}
	if chMDFe == "" {
		return errors.New("parse mdfe: missing chMDFe")
	}
	if missingDetEvento {
		return errors.New("parse mdfe: missing detEvento")
	}
	return nil
}

func validateMDFeRetEventoFields(isNil bool, tpAmb, cStat string) error {
	if isNil {
		return errors.New("parse mdfe: missing infEvento")
	}
	if tpAmb == "" {
		return errors.New("parse mdfe: missing tpAmb")
	}
	if cStat == "" {
		return errors.New("parse mdfe: missing cStat")
	}
	return nil
}

func validateMDFeEventField(doc *Document, spec *mdfeEventSpec, kind mdfeEventRootKind, validate func(any) error) error {
	root := mdfeDocumentEventField(doc, spec, kind)
	if !mdfeHasValue(root) {
		return nil
	}
	return validate(root.Interface())
}

func validateMDFeProcEventField(doc *Document, spec *mdfeEventSpec) error {
	root := mdfeDocumentEventField(doc, spec, mdfeProcEventRoot)
	if !mdfeHasValue(root) {
		return nil
	}
	return validateMDFeProcEvento(mdfeAnyField(root, "EventoMDFe"), mdfeAnyField(root, "RetEventoMDFe"))
}

func validateMDFeSentEventRoot(evento any) error {
	inf := mdfeField(reflect.ValueOf(evento), "InfEvento")
	return validateMDFeEventoFields(!mdfeHasValue(inf), mdfeStringField(inf, "ChMDFe"), !mdfeHasValue(mdfeField(inf, "DetEvento")))
}

func validateMDFeRetEventRoot(retEvento any) error {
	inf := mdfeField(reflect.ValueOf(retEvento), "InfEvento")
	return validateMDFeRetEventoFields(!mdfeHasValue(inf), mdfeStringField(inf, "TpAmb"), mdfeStringField(inf, "CStat"))
}

func validateMDFeProcEvento(evento any, retEvento any) error {
	if evento == nil {
		return errors.New("parse mdfe: missing eventoMDFe")
	}
	if retEvento == nil {
		return errors.New("parse mdfe: missing retEventoMDFe")
	}
	return nil
}

func marshalEventRoot(e *xml.Encoder, d *Document) error {
	return marshalMDFeEventRoot(e, d, mdfeSentEventRoot)
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	return marshalMDFeEventRoot(e, d, mdfeRetEventRoot)
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	return marshalMDFeEventRoot(e, d, mdfeProcEventRoot)
}

func marshalMDFeEventRoot(e *xml.Encoder, d *Document, kind mdfeEventRootKind) error {
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}
	_, root, ok := mdfeEventSpecForDocument(d, kind)
	if !ok {
		return errSingleRoot
	}
	versao := xmlutil.FirstNonEmpty(d.VersaoAttr, mdfeStringField(root, "VersaoAttr"))
	switch kind {
	case mdfeSentEventRoot:
		return encodeMDFeEvent(e, versao, mdfeAnyField(root, "InfEvento"), mdfeAnyField(root, "DsSignature"))
	case mdfeRetEventRoot:
		return encodeMDFeRetEvent(e, versao, mdfeAnyField(root, "InfEvento"))
	case mdfeProcEventRoot:
		return encodeMDFeProcEvent(
			e,
			versao,
			mdfeStringPtrField(root, "IpTransmissorAttr"),
			mdfeStringPtrField(root, "NPortaConAttr"),
			mdfeStringPtrField(root, "DhConexaoAttr"),
			mdfeAnyField(root, "EventoMDFe"),
			mdfeAnyField(root, "RetEventoMDFe"),
		)
	default:
		return errSingleRoot
	}
}

func encodeMDFeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName     xml.Name `xml:"eventoMDFe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}{
		XMLName:     xml.Name{Local: "eventoMDFe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeMDFeRetEvent(e *xml.Encoder, versao string, infEvento any) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName    xml.Name `xml:"retEventoMDFe"`
		XMLNS      string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr string   `xml:"versao,attr,omitempty"`
		InfEvento  any      `xml:"infEvento"`
	}{
		XMLName:    xml.Name{Local: "retEventoMDFe"},
		XMLNS:      namespace,
		VersaoAttr: versao,
		InfEvento:  infEvento,
	})
}

func encodeMDFeProcEvent(e *xml.Encoder, versao string, ipTransmissor, nPortaCon, dhConexao *string, evento any, retEvento any) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName           xml.Name `xml:"procEventoMDFe"`
		XMLNS             string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string   `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string  `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string  `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string  `xml:"dhConexao,attr,omitempty"`
		EventoMDFe        any      `xml:"eventoMDFe"`
		RetEventoMDFe     any      `xml:"retEventoMDFe"`
	}{
		XMLName:           xml.Name{Local: "procEventoMDFe"},
		XMLNS:             namespace,
		VersaoAttr:        versao,
		IpTransmissorAttr: ipTransmissor,
		NPortaConAttr:     nPortaCon,
		DhConexaoAttr:     dhConexao,
		EventoMDFe:        evento,
		RetEventoMDFe:     retEvento,
	})
}

func parseEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	return parseMDFeEventDocument(data, rootName, tpEvento, mdfeSentEventRoot)
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	return parseMDFeEventDocument(data, rootName, tpEvento, mdfeRetEventRoot)
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	return parseMDFeEventDocument(data, rootName, tpEvento, mdfeProcEventRoot)
}

func parseMDFeEventDocument(data []byte, rootName, tpEvento string, kind mdfeEventRootKind) (*Document, error) {
	spec := mdfeEventSpecForType(tpEvento)
	if spec == nil {
		return nil, fmt.Errorf("parse mdfe: unsupported eventoMDFe type %q", tpEvento)
	}

	parsed := reflect.New(spec.rootType(kind))
	if err := xml.Unmarshal(data, parsed.Interface()); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode %s %s: %w", kind.rootName(), spec.context, err)
	}

	doc := &Document{
		VersaoAttr: mdfeStringField(parsed, "VersaoAttr"),
		RootName:   rootName,
	}
	mdfeDocumentEventField(doc, spec, kind).Set(parsed)
	return finalizeDoc(doc)
}

func mdfeEventSpecForType(tpEvento string) *mdfeEventSpec {
	var generic *mdfeEventSpec
	for i := range mdfeEventSpecs {
		spec := &mdfeEventSpecs[i]
		if len(spec.eventTypes) == 0 {
			generic = spec
			continue
		}
		for _, eventType := range spec.eventTypes {
			if eventType == tpEvento {
				return spec
			}
		}
	}
	return generic
}

func mdfeEventSpecForDocument(d *Document, kind mdfeEventRootKind) (*mdfeEventSpec, reflect.Value, bool) {
	for i := range mdfeEventSpecs {
		spec := &mdfeEventSpecs[i]
		root := mdfeDocumentEventField(d, spec, kind)
		if root.IsValid() && !root.IsNil() {
			return spec, root, true
		}
	}
	return nil, reflect.Value{}, false
}

func mdfeDocumentEventField(d *Document, spec *mdfeEventSpec, kind mdfeEventRootKind) reflect.Value {
	if d == nil || spec == nil {
		return reflect.Value{}
	}
	return reflect.ValueOf(d).Elem().FieldByName(spec.docField(kind))
}

func activeMDFeEventRootCount(d *Document) int {
	count := 0
	for i := range mdfeEventSpecs {
		spec := &mdfeEventSpecs[i]
		for _, kind := range []mdfeEventRootKind{mdfeSentEventRoot, mdfeRetEventRoot, mdfeProcEventRoot} {
			root := mdfeDocumentEventField(d, spec, kind)
			if root.IsValid() && !root.IsNil() {
				count++
			}
		}
	}
	return count
}

func (s *mdfeEventSpec) rootType(kind mdfeEventRootKind) reflect.Type {
	switch kind {
	case mdfeSentEventRoot:
		return s.eventTypeOf
	case mdfeRetEventRoot:
		return s.retEventTypeOf
	case mdfeProcEventRoot:
		return s.procEventTypeOf
	default:
		return nil
	}
}

func (s *mdfeEventSpec) docField(kind mdfeEventRootKind) string {
	switch kind {
	case mdfeSentEventRoot:
		return s.eventField
	case mdfeRetEventRoot:
		return "Ret" + s.eventField
	case mdfeProcEventRoot:
		return "Proc" + s.eventField
	default:
		return ""
	}
}

func (kind mdfeEventRootKind) rootName() string {
	switch kind {
	case mdfeSentEventRoot:
		return "eventoMDFe"
	case mdfeRetEventRoot:
		return "retEventoMDFe"
	case mdfeProcEventRoot:
		return "procEventoMDFe"
	default:
		return "eventoMDFe"
	}
}

func mdfeField(value reflect.Value, name string) reflect.Value {
	if !value.IsValid() {
		return reflect.Value{}
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return value.FieldByName(name)
}

func mdfeAnyField(value reflect.Value, name string) any {
	field := mdfeField(value, name)
	if !field.IsValid() {
		return nil
	}
	if field.Kind() == reflect.Pointer && field.IsNil() {
		return nil
	}
	return field.Interface()
}

func mdfeStringField(value reflect.Value, name string) string {
	return mdfeStringValue(mdfeField(value, name))
}

func mdfeStringPtrField(value reflect.Value, name string) *string {
	field := mdfeField(value, name)
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return nil
	}
	if ptr, ok := field.Interface().(*string); ok {
		return ptr
	}
	if field.Type().Elem().Kind() == reflect.String {
		value := field.Elem().String()
		return &value
	}
	return nil
}

func mdfeStringValue(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.String {
		return ""
	}
	return value.String()
}

func mdfeHasValue(value reflect.Value) bool {
	return value.IsValid() && (value.Kind() != reflect.Pointer || !value.IsNil())
}

func activeRootCount(doc *Document) int {
	count := 0
	for _, ok := range []bool{
		doc.MDFe != nil,
		doc.MDFeProc != nil,
		doc.EnviMDFe != nil,
		doc.RetEnviMDFe != nil,
		doc.RetMDFe != nil,
		doc.ConsNaoEnc != nil,
		doc.RetConsNaoEnc != nil,
		doc.ConsReciMDFe != nil,
		doc.RetConsReciMDFe != nil,
		doc.ConsSitMDFe != nil,
		doc.RetConsSitMDFe != nil,
		doc.ConsStatServMDFe != nil,
		doc.RetConsStatServMDFe != nil,
		doc.DistDFeInt != nil,
		doc.RetDistDFeInt != nil,
		doc.DistMDFe != nil,
		doc.RetDistMDFe != nil,
		doc.MDFeConsultaDFe != nil,
		doc.RetMDFeConsultaDFe != nil,
	} {
		if ok {
			count++
		}
	}
	count += activeMDFeEventRootCount(doc)
	return count
}

func versionFromMDFe(doc *mdfeSchema.TMDFe) string {
	if doc == nil || doc.InfMDFe == nil {
		return ""
	}
	return doc.InfMDFe.VersaoAttr
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
