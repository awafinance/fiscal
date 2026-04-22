package mdfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

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
	return nil, &fiscalerr.UnsupportedRootError{Family: "mdfe", Root: RootName}
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
	if err := validateEventos(doc); err != nil {
		return err
	}
	if err := validateRetEventos(doc); err != nil {
		return err
	}
	return validateProcEventos(doc)
}

func validateEventos(doc *Document) error {
	switch {
	case doc.EventoMDFe != nil:
		return validateMDFeEvento(doc.EventoMDFe.InfEvento)
	case doc.EventoCancMDFe != nil:
		return validateMDFeEvento(doc.EventoCancMDFe.InfEvento)
	case doc.EventoEncMDFe != nil:
		return validateMDFeEvento(doc.EventoEncMDFe.InfEvento)
	case doc.EventoIncCondutorMDFe != nil:
		return validateMDFeEvento(doc.EventoIncCondutorMDFe.InfEvento)
	case doc.EventoInclusaoDFeMDFe != nil:
		return validateMDFeEvento(doc.EventoInclusaoDFeMDFe.InfEvento)
	case doc.EventoPagtoOperMDFe != nil:
		return validateMDFeEvento(doc.EventoPagtoOperMDFe.InfEvento)
	case doc.EventoAlteracaoPagtoServMDFe != nil:
		return validateMDFeEvento(doc.EventoAlteracaoPagtoServMDFe.InfEvento)
	case doc.EventoConfirmaServMDFe != nil:
		return validateMDFeEvento(doc.EventoConfirmaServMDFe.InfEvento)
	}
	return nil
}

func validateRetEventos(doc *Document) error {
	switch {
	case doc.RetEventoMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoMDFe.InfEvento)
	case doc.RetEventoCancMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoCancMDFe.InfEvento)
	case doc.RetEventoEncMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoEncMDFe.InfEvento)
	case doc.RetEventoIncCondutorMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoIncCondutorMDFe.InfEvento)
	case doc.RetEventoInclusaoDFeMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoInclusaoDFeMDFe.InfEvento)
	case doc.RetEventoPagtoOperMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoPagtoOperMDFe.InfEvento)
	case doc.RetEventoAlteracaoPagtoServMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoAlteracaoPagtoServMDFe.InfEvento)
	case doc.RetEventoConfirmaServMDFe != nil:
		return validateMDFeRetEvento(doc.RetEventoConfirmaServMDFe.InfEvento)
	}
	return nil
}

func validateProcEventos(doc *Document) error {
	switch {
	case doc.ProcEventoMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoMDFe.EventoMDFe, doc.ProcEventoMDFe.RetEventoMDFe)
	case doc.ProcEventoCancMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoCancMDFe.EventoMDFe, doc.ProcEventoCancMDFe.RetEventoMDFe)
	case doc.ProcEventoEncMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoEncMDFe.EventoMDFe, doc.ProcEventoEncMDFe.RetEventoMDFe)
	case doc.ProcEventoIncCondutorMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoIncCondutorMDFe.EventoMDFe, doc.ProcEventoIncCondutorMDFe.RetEventoMDFe)
	case doc.ProcEventoInclusaoDFeMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoInclusaoDFeMDFe.EventoMDFe, doc.ProcEventoInclusaoDFeMDFe.RetEventoMDFe)
	case doc.ProcEventoPagtoOperMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoPagtoOperMDFe.EventoMDFe, doc.ProcEventoPagtoOperMDFe.RetEventoMDFe)
	case doc.ProcEventoAlteracaoPagtoServMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoAlteracaoPagtoServMDFe.EventoMDFe, doc.ProcEventoAlteracaoPagtoServMDFe.RetEventoMDFe)
	case doc.ProcEventoConfirmaServMDFe != nil:
		return validateMDFeProcEvento(doc.ProcEventoConfirmaServMDFe.EventoMDFe, doc.ProcEventoConfirmaServMDFe.RetEventoMDFe)
	}
	return nil
}

func validateMDFeEvento(inf any) error {
	switch inf := inf.(type) {
	case *eventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *cancelEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *encEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *incCondutorEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *inclusaoDFeEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *pagtoOperEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *alteracaoPagtoServEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	case *confirmaServEventSchema.TAnonComplexInfEvento1:
		return validateMDFeEventoFields(inf == nil, inf.ChMDFe, inf.DetEvento == nil)
	default:
		return errors.New("parse mdfe: missing infEvento")
	}
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

func validateMDFeRetEvento(inf any) error {
	switch inf := inf.(type) {
	case *eventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *cancelEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *encEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *incCondutorEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *inclusaoDFeEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *pagtoOperEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *alteracaoPagtoServEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *confirmaServEventSchema.TAnonComplexInfEvento2:
		return validateMDFeRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	default:
		return errors.New("parse mdfe: missing infEvento")
	}
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
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}

	switch {
	case d.EventoMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoMDFe.VersaoAttr), d.EventoMDFe.InfEvento, d.EventoMDFe.DsSignature)
	case d.EventoCancMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoCancMDFe.VersaoAttr), d.EventoCancMDFe.InfEvento, d.EventoCancMDFe.DsSignature)
	case d.EventoEncMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoEncMDFe.VersaoAttr), d.EventoEncMDFe.InfEvento, d.EventoEncMDFe.DsSignature)
	case d.EventoIncCondutorMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoIncCondutorMDFe.VersaoAttr), d.EventoIncCondutorMDFe.InfEvento, d.EventoIncCondutorMDFe.DsSignature)
	case d.EventoInclusaoDFeMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoInclusaoDFeMDFe.VersaoAttr), d.EventoInclusaoDFeMDFe.InfEvento, d.EventoInclusaoDFeMDFe.DsSignature)
	case d.EventoPagtoOperMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoPagtoOperMDFe.VersaoAttr), d.EventoPagtoOperMDFe.InfEvento, d.EventoPagtoOperMDFe.DsSignature)
	case d.EventoAlteracaoPagtoServMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoAlteracaoPagtoServMDFe.VersaoAttr), d.EventoAlteracaoPagtoServMDFe.InfEvento, d.EventoAlteracaoPagtoServMDFe.DsSignature)
	case d.EventoConfirmaServMDFe != nil:
		return encodeMDFeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoConfirmaServMDFe.VersaoAttr), d.EventoConfirmaServMDFe.InfEvento, d.EventoConfirmaServMDFe.DsSignature)
	default:
		return errSingleRoot
	}
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}

	switch {
	case d.RetEventoMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoMDFe.VersaoAttr), d.RetEventoMDFe.InfEvento)
	case d.RetEventoCancMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoCancMDFe.VersaoAttr), d.RetEventoCancMDFe.InfEvento)
	case d.RetEventoEncMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoEncMDFe.VersaoAttr), d.RetEventoEncMDFe.InfEvento)
	case d.RetEventoIncCondutorMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoIncCondutorMDFe.VersaoAttr), d.RetEventoIncCondutorMDFe.InfEvento)
	case d.RetEventoInclusaoDFeMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoInclusaoDFeMDFe.VersaoAttr), d.RetEventoInclusaoDFeMDFe.InfEvento)
	case d.RetEventoPagtoOperMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoPagtoOperMDFe.VersaoAttr), d.RetEventoPagtoOperMDFe.InfEvento)
	case d.RetEventoAlteracaoPagtoServMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoAlteracaoPagtoServMDFe.VersaoAttr), d.RetEventoAlteracaoPagtoServMDFe.InfEvento)
	case d.RetEventoConfirmaServMDFe != nil:
		return encodeMDFeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoConfirmaServMDFe.VersaoAttr), d.RetEventoConfirmaServMDFe.InfEvento)
	default:
		return errSingleRoot
	}
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}

	switch {
	case d.ProcEventoMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoMDFe.VersaoAttr), d.ProcEventoMDFe.IpTransmissorAttr, d.ProcEventoMDFe.NPortaConAttr, d.ProcEventoMDFe.DhConexaoAttr, d.ProcEventoMDFe.EventoMDFe, d.ProcEventoMDFe.RetEventoMDFe)
	case d.ProcEventoCancMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoCancMDFe.VersaoAttr), d.ProcEventoCancMDFe.IpTransmissorAttr, d.ProcEventoCancMDFe.NPortaConAttr, d.ProcEventoCancMDFe.DhConexaoAttr, d.ProcEventoCancMDFe.EventoMDFe, d.ProcEventoCancMDFe.RetEventoMDFe)
	case d.ProcEventoEncMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoEncMDFe.VersaoAttr), d.ProcEventoEncMDFe.IpTransmissorAttr, d.ProcEventoEncMDFe.NPortaConAttr, d.ProcEventoEncMDFe.DhConexaoAttr, d.ProcEventoEncMDFe.EventoMDFe, d.ProcEventoEncMDFe.RetEventoMDFe)
	case d.ProcEventoIncCondutorMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoIncCondutorMDFe.VersaoAttr), d.ProcEventoIncCondutorMDFe.IpTransmissorAttr, d.ProcEventoIncCondutorMDFe.NPortaConAttr, d.ProcEventoIncCondutorMDFe.DhConexaoAttr, d.ProcEventoIncCondutorMDFe.EventoMDFe, d.ProcEventoIncCondutorMDFe.RetEventoMDFe)
	case d.ProcEventoInclusaoDFeMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoInclusaoDFeMDFe.VersaoAttr), d.ProcEventoInclusaoDFeMDFe.IpTransmissorAttr, d.ProcEventoInclusaoDFeMDFe.NPortaConAttr, d.ProcEventoInclusaoDFeMDFe.DhConexaoAttr, d.ProcEventoInclusaoDFeMDFe.EventoMDFe, d.ProcEventoInclusaoDFeMDFe.RetEventoMDFe)
	case d.ProcEventoPagtoOperMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoPagtoOperMDFe.VersaoAttr), d.ProcEventoPagtoOperMDFe.IpTransmissorAttr, d.ProcEventoPagtoOperMDFe.NPortaConAttr, d.ProcEventoPagtoOperMDFe.DhConexaoAttr, d.ProcEventoPagtoOperMDFe.EventoMDFe, d.ProcEventoPagtoOperMDFe.RetEventoMDFe)
	case d.ProcEventoAlteracaoPagtoServMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoAlteracaoPagtoServMDFe.VersaoAttr), d.ProcEventoAlteracaoPagtoServMDFe.IpTransmissorAttr, d.ProcEventoAlteracaoPagtoServMDFe.NPortaConAttr, d.ProcEventoAlteracaoPagtoServMDFe.DhConexaoAttr, d.ProcEventoAlteracaoPagtoServMDFe.EventoMDFe, d.ProcEventoAlteracaoPagtoServMDFe.RetEventoMDFe)
	case d.ProcEventoConfirmaServMDFe != nil:
		return encodeMDFeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoConfirmaServMDFe.VersaoAttr), d.ProcEventoConfirmaServMDFe.IpTransmissorAttr, d.ProcEventoConfirmaServMDFe.NPortaConAttr, d.ProcEventoConfirmaServMDFe.DhConexaoAttr, d.ProcEventoConfirmaServMDFe.EventoMDFe, d.ProcEventoConfirmaServMDFe.RetEventoMDFe)
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

func decodeEvent[T any](data []byte, context string, assign func(*T) *Document) (*Document, error) {
	var parsed T
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse mdfe: decode %s: %w", context, err)
	}
	return finalizeDoc(assign(&parsed))
}

func parseEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "eventoMDFe cancelamento", func(p *cancelEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancMDFe: p, RootName: rootName}
		})
	case "110112":
		return decodeEvent(data, "eventoMDFe encerramento", func(p *encEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoEncMDFe: p, RootName: rootName}
		})
	case "110114":
		return decodeEvent(data, "eventoMDFe inclusao condutor", func(p *incCondutorEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoIncCondutorMDFe: p, RootName: rootName}
		})
	case "110115":
		return decodeEvent(data, "eventoMDFe inclusao dfe", func(p *inclusaoDFeEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoInclusaoDFeMDFe: p, RootName: rootName}
		})
	case "110116":
		return decodeEvent(data, "eventoMDFe pagamento operacao", func(p *pagtoOperEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoPagtoOperMDFe: p, RootName: rootName}
		})
	case "110117":
		return decodeEvent(data, "eventoMDFe confirma servico", func(p *confirmaServEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoConfirmaServMDFe: p, RootName: rootName}
		})
	case "110118":
		return decodeEvent(data, "eventoMDFe alteracao pagamento servico", func(p *alteracaoPagtoServEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoAlteracaoPagtoServMDFe: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "eventoMDFe", func(p *eventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoMDFe: p, RootName: rootName}
		})
	}
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "retEventoMDFe cancelamento", func(p *cancelEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCancMDFe: p, RootName: rootName}
		})
	case "110112":
		return decodeEvent(data, "retEventoMDFe encerramento", func(p *encEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoEncMDFe: p, RootName: rootName}
		})
	case "110114":
		return decodeEvent(data, "retEventoMDFe inclusao condutor", func(p *incCondutorEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoIncCondutorMDFe: p, RootName: rootName}
		})
	case "110115":
		return decodeEvent(data, "retEventoMDFe inclusao dfe", func(p *inclusaoDFeEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoInclusaoDFeMDFe: p, RootName: rootName}
		})
	case "110116":
		return decodeEvent(data, "retEventoMDFe pagamento operacao", func(p *pagtoOperEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoPagtoOperMDFe: p, RootName: rootName}
		})
	case "110117":
		return decodeEvent(data, "retEventoMDFe confirma servico", func(p *confirmaServEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoConfirmaServMDFe: p, RootName: rootName}
		})
	case "110118":
		return decodeEvent(data, "retEventoMDFe alteracao pagamento servico", func(p *alteracaoPagtoServEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoAlteracaoPagtoServMDFe: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "retEventoMDFe", func(p *eventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoMDFe: p, RootName: rootName}
		})
	}
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "procEventoMDFe cancelamento", func(p *cancelEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCancMDFe: p, RootName: rootName}
		})
	case "110112":
		return decodeEvent(data, "procEventoMDFe encerramento", func(p *encEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoEncMDFe: p, RootName: rootName}
		})
	case "110114":
		return decodeEvent(data, "procEventoMDFe inclusao condutor", func(p *incCondutorEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoIncCondutorMDFe: p, RootName: rootName}
		})
	case "110115":
		return decodeEvent(data, "procEventoMDFe inclusao dfe", func(p *inclusaoDFeEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoInclusaoDFeMDFe: p, RootName: rootName}
		})
	case "110116":
		return decodeEvent(data, "procEventoMDFe pagamento operacao", func(p *pagtoOperEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoPagtoOperMDFe: p, RootName: rootName}
		})
	case "110117":
		return decodeEvent(data, "procEventoMDFe confirma servico", func(p *confirmaServEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoConfirmaServMDFe: p, RootName: rootName}
		})
	case "110118":
		return decodeEvent(data, "procEventoMDFe alteracao pagamento servico", func(p *alteracaoPagtoServEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoAlteracaoPagtoServMDFe: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "procEventoMDFe", func(p *eventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoMDFe: p, RootName: rootName}
		})
	}
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
		doc.EventoMDFe != nil,
		doc.RetEventoMDFe != nil,
		doc.ProcEventoMDFe != nil,
		doc.EventoCancMDFe != nil,
		doc.RetEventoCancMDFe != nil,
		doc.ProcEventoCancMDFe != nil,
		doc.EventoEncMDFe != nil,
		doc.RetEventoEncMDFe != nil,
		doc.ProcEventoEncMDFe != nil,
		doc.EventoIncCondutorMDFe != nil,
		doc.RetEventoIncCondutorMDFe != nil,
		doc.ProcEventoIncCondutorMDFe != nil,
		doc.EventoInclusaoDFeMDFe != nil,
		doc.RetEventoInclusaoDFeMDFe != nil,
		doc.ProcEventoInclusaoDFeMDFe != nil,
		doc.EventoPagtoOperMDFe != nil,
		doc.RetEventoPagtoOperMDFe != nil,
		doc.ProcEventoPagtoOperMDFe != nil,
		doc.EventoAlteracaoPagtoServMDFe != nil,
		doc.RetEventoAlteracaoPagtoServMDFe != nil,
		doc.ProcEventoAlteracaoPagtoServMDFe != nil,
		doc.EventoConfirmaServMDFe != nil,
		doc.RetEventoConfirmaServMDFe != nil,
		doc.ProcEventoConfirmaServMDFe != nil,
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
