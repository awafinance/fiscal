package mdfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	distSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v1_0/dist_dfe"
	consNaoEncSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/cons_nao_enc"
	consReciSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/cons_reci"
	consultaDFESchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/consulta_dfe"
	consSitSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/consulta_situacao"
	distMDFeSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/dist_mdfe"
	eventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento"
	alteracaoPagtoServEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_alteracao_pagto_serv"
	cancelEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_cancel"
	confirmaServEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_confirma_serv"
	encEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_enc"
	incCondutorEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_inc_condutor"
	inclusaoDFeEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_inclusao_dfe"
	pagtoOperEventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento_pagto_oper"
	mdfeSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/mdfe"
	statusSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/status_servico"
)

const namespace = "http://www.portalfiscal.inf.br/mdfe"

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
	rootName                         string                                     `json:"-"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "MDFe", "":
		if d.MDFe != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName     xml.Name                             `xml:"MDFe"`
				XMLNS       string                               `xml:"xmlns,attr,omitempty"`
				InfMDFe     *mdfeSchema.TAnonComplexInfMDFe1     `xml:"infMDFe"`
				InfMDFeSupl *mdfeSchema.TAnonComplexInfMDFeSupl1 `xml:"infMDFeSupl,omitempty"`
				DsSignature *mdfeSchema.SignatureType            `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "MDFe"},
				XMLNS:       namespace,
				InfMDFe:     d.MDFe.InfMDFe,
				InfMDFeSupl: d.MDFe.InfMDFeSupl,
				DsSignature: d.MDFe.DsSignature,
			})
		}
	case "mdfeProc":
		if d.MDFeProc != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:           xml.Name{Local: "mdfeProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.MDFeProc.VersaoAttr),
				IpTransmissorAttr: d.MDFeProc.IpTransmissorAttr,
				NPortaConAttr:     d.MDFeProc.NPortaConAttr,
				DhConexaoAttr:     d.MDFeProc.DhConexaoAttr,
				MDFe:              d.MDFeProc.MDFe,
				ProtMDFe:          d.MDFeProc.ProtMDFe,
			})
		}
	case "enviMDFe":
		if d.EnviMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"enviMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*mdfeSchema.TEnviMDFe
			}{
				XMLName:   xml.Name{Local: "enviMDFe"},
				XMLNS:     namespace,
				TEnviMDFe: d.EnviMDFe,
			})
		}
	case "retEnviMDFe":
		if d.RetEnviMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retEnviMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*mdfeSchema.TRetEnviMDFe
			}{
				XMLName:      xml.Name{Local: "retEnviMDFe"},
				XMLNS:        namespace,
				TRetEnviMDFe: d.RetEnviMDFe,
			})
		}
	case "retMDFe":
		if d.RetMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*mdfeSchema.TRetMDFe
			}{
				XMLName:  xml.Name{Local: "retMDFe"},
				XMLNS:    namespace,
				TRetMDFe: d.RetMDFe,
			})
		}
	case "consMDFeNaoEnc":
		if d.ConsNaoEnc != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName    xml.Name                  `xml:"consMDFeNaoEnc"`
				XMLNS      string                    `xml:"xmlns,attr,omitempty"`
				VersaoAttr string                    `xml:"versao,attr,omitempty"`
				TpAmb      string                    `xml:"tpAmb"`
				XServ      *consNaoEncSchema.TString `xml:"xServ,omitempty"`
				CNPJ       *string                   `xml:"CNPJ,omitempty"`
				CPF        *string                   `xml:"CPF,omitempty"`
			}

			return e.Encode(root{
				XMLName:    xml.Name{Local: "consMDFeNaoEnc"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.ConsNaoEnc.VersaoAttr),
				TpAmb:      d.ConsNaoEnc.TpAmb,
				XServ:      d.ConsNaoEnc.XServ,
				CNPJ:       d.ConsNaoEnc.CNPJ,
				CPF:        d.ConsNaoEnc.CPF,
			})
		}
	case "retConsMDFeNaoEnc":
		if d.RetConsNaoEnc != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsMDFeNaoEnc"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consNaoEncSchema.TRetConsMDFeNaoEnc
			}{
				XMLName:            xml.Name{Local: "retConsMDFeNaoEnc"},
				XMLNS:              namespace,
				TRetConsMDFeNaoEnc: d.RetConsNaoEnc,
			})
		}
	case "consReciMDFe":
		if d.ConsReciMDFe != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName    xml.Name `xml:"consReciMDFe"`
				XMLNS      string   `xml:"xmlns,attr,omitempty"`
				VersaoAttr string   `xml:"versao,attr,omitempty"`
				TpAmb      string   `xml:"tpAmb"`
				NRec       string   `xml:"nRec"`
			}

			return e.Encode(root{
				XMLName:    xml.Name{Local: "consReciMDFe"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.ConsReciMDFe.VersaoAttr),
				TpAmb:      d.ConsReciMDFe.TpAmb,
				NRec:       d.ConsReciMDFe.NRec,
			})
		}
	case "retConsReciMDFe":
		if d.RetConsReciMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsReciMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consReciSchema.TRetConsReciMDFe
			}{
				XMLName:          xml.Name{Local: "retConsReciMDFe"},
				XMLNS:            namespace,
				TRetConsReciMDFe: d.RetConsReciMDFe,
			})
		}
	case "consSitMDFe":
		if d.ConsSitMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"consSitMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consSitSchema.TConsSitMDFe
			}{
				XMLName:      xml.Name{Local: "consSitMDFe"},
				XMLNS:        namespace,
				TConsSitMDFe: d.ConsSitMDFe,
			})
		}
	case "retConsSitMDFe":
		if d.RetConsSitMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsSitMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consSitSchema.TRetConsSitMDFe
			}{
				XMLName:         xml.Name{Local: "retConsSitMDFe"},
				XMLNS:           namespace,
				TRetConsSitMDFe: d.RetConsSitMDFe,
			})
		}
	case "consStatServMDFe":
		if d.ConsStatServMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"consStatServMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*statusSchema.TConsStatServ
			}{
				XMLName:       xml.Name{Local: "consStatServMDFe"},
				XMLNS:         namespace,
				TConsStatServ: d.ConsStatServMDFe,
			})
		}
	case "retConsStatServMDFe":
		if d.RetConsStatServMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsStatServMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*statusSchema.TRetConsStatServ
			}{
				XMLName:          xml.Name{Local: "retConsStatServMDFe"},
				XMLNS:            namespace,
				TRetConsStatServ: d.RetConsStatServMDFe,
			})
		}
	case "eventoMDFe":
		return marshalEventRoot(e, d)
	case "retEventoMDFe":
		return marshalRetEventRoot(e, d)
	case "procEventoMDFe":
		return marshalProcEventRoot(e, d)
	case "distDFeInt":
		if d.DistDFeInt != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:    xml.Name{Local: "distDFeInt"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.DistDFeInt.VersaoAttr),
				TpAmb:      d.DistDFeInt.TpAmb,
				CNPJ:       d.DistDFeInt.CNPJ,
				CPF:        d.DistDFeInt.CPF,
				DistNSU:    d.DistDFeInt.DistNSU,
				ConsNSU:    d.DistDFeInt.ConsNSU,
			})
		}
	case "retDistDFeInt":
		if d.RetDistDFeInt != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:        xml.Name{Local: "retDistDFeInt"},
				XMLNS:          namespace,
				VersaoAttr:     firstNonEmpty(d.VersaoAttr, d.RetDistDFeInt.VersaoAttr),
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
	case "distMDFe":
		if d.DistMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"distMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*distMDFeSchema.TDistDFe
			}{
				XMLName:  xml.Name{Local: "distMDFe"},
				XMLNS:    namespace,
				TDistDFe: d.DistMDFe,
			})
		}
	case "retDistMDFe":
		if d.RetDistMDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retDistMDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*distMDFeSchema.TRetDistDFe
			}{
				XMLName:     xml.Name{Local: "retDistMDFe"},
				XMLNS:       namespace,
				TRetDistDFe: d.RetDistMDFe,
			})
		}
	case "mdfeConsultaDFe":
		if d.MDFeConsultaDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"mdfeConsultaDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consultaDFESchema.TMDFeConsultaDFe
			}{
				XMLName:          xml.Name{Local: "mdfeConsultaDFe"},
				XMLNS:            namespace,
				TMDFeConsultaDFe: d.MDFeConsultaDFe,
			})
		}
	case "retMDFeConsultaDFe":
		if d.RetMDFeConsultaDFe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retMDFeConsultaDFe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consultaDFESchema.TRetMDFeConsultaDFe
			}{
				XMLName:             xml.Name{Local: "retMDFeConsultaDFe"},
				XMLNS:               namespace,
				TRetMDFeConsultaDFe: d.RetMDFeConsultaDFe,
			})
		}
	}

	return errors.New("marshal mdfe: document must contain exactly one supported root")
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse mdfe: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse mdfe: read root: %w", rootErr)
	}

	switch rootName {
	case "MDFe":
		var parsed mdfeSchema.TMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode MDFe: %w", err)
		}
		doc := &Document{VersaoAttr: versionFromMDFe(&parsed), MDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "mdfeProc":
		var parsed mdfeSchema.TAnonComplexMdfeProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode mdfeProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, MDFeProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "enviMDFe":
		var parsed mdfeSchema.TEnviMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode enviMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EnviMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retEnviMDFe":
		var parsed mdfeSchema.TRetEnviMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEnviMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEnviMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retMDFe":
		var parsed mdfeSchema.TRetMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consMDFeNaoEnc":
		var parsed consNaoEncSchema.TConsMDFeNaoEnc
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode consMDFeNaoEnc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsNaoEnc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsMDFeNaoEnc":
		var parsed consNaoEncSchema.TRetConsMDFeNaoEnc
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retConsMDFeNaoEnc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsNaoEnc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consReciMDFe":
		var parsed consReciSchema.TConsReciMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode consReciMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsReciMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsReciMDFe":
		var parsed consReciSchema.TRetConsReciMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retConsReciMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsReciMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consSitMDFe":
		var parsed consSitSchema.TConsSitMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode consSitMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsSitMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsSitMDFe":
		var parsed consSitSchema.TRetConsSitMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retConsSitMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsSitMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consStatServMDFe":
		var parsed statusSchema.TConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode consStatServMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsStatServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsStatServMDFe":
		var parsed statusSchema.TRetConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retConsStatServMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "eventoMDFe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse mdfe: missing infEvento")
		}
		return parseEventDocument(data, rootName, tpEvento)
	case "retEventoMDFe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse mdfe: missing infEvento")
		}
		return parseRetEventDocument(data, rootName, tpEvento)
	case "procEventoMDFe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse mdfe: missing infEvento")
		}
		return parseProcEventDocument(data, rootName, tpEvento)
	case "distDFeInt":
		var parsed distSchema.TAnonComplexDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode distDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retDistDFeInt":
		var parsed distSchema.TAnonComplexRetDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retDistDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "distMDFe":
		var parsed distMDFeSchema.TDistDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode distMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DistMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retDistMDFe":
		var parsed distMDFeSchema.TRetDistDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retDistMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetDistMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "mdfeConsultaDFe":
		var parsed consultaDFESchema.TMDFeConsultaDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode mdfeConsultaDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, MDFeConsultaDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retMDFeConsultaDFe":
		var parsed consultaDFESchema.TRetMDFeConsultaDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retMDFeConsultaDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetMDFeConsultaDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse mdfe: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse mdfe: unsupported root element %q", rootName)
	}
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

func parseRootName(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var rootName string

	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if rootName == "" {
					return "", err
				}
				return rootName, nil
			}
			return rootName, err
		}

		if start, ok := tok.(xml.StartElement); ok && rootName == "" {
			rootName = start.Name.Local
		}
	}
}

func validateDocument(doc *Document) error {
	if activeRootCount(doc) != 1 {
		return errors.New("parse mdfe: document must contain exactly one supported root")
	}

	if doc.MDFe != nil {
		if err := validateMDFeRoot(doc.MDFe); err != nil {
			return err
		}
	}
	if doc.MDFeProc != nil {
		if doc.MDFeProc.MDFe == nil {
			return errors.New("parse mdfe: missing MDFe")
		}
		if doc.MDFeProc.ProtMDFe == nil {
			return errors.New("parse mdfe: missing protMDFe")
		}
	}
	if doc.EnviMDFe != nil {
		if doc.EnviMDFe.IdLote == "" {
			return errors.New("parse mdfe: missing idLote")
		}
		if doc.EnviMDFe.MDFe == nil {
			return errors.New("parse mdfe: missing MDFe")
		}
	}
	if doc.RetEnviMDFe != nil {
		if valueOrEmpty(doc.RetEnviMDFe.TpAmb) == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetEnviMDFe.CUF == "" {
			return errors.New("parse mdfe: missing cUF")
		}
		if doc.RetEnviMDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
	}
	if doc.RetMDFe != nil {
		if valueOrEmpty(doc.RetMDFe.TpAmb) == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetMDFe.CUF == "" {
			return errors.New("parse mdfe: missing cUF")
		}
		if doc.RetMDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
	}
	if doc.ConsNaoEnc != nil {
		if doc.ConsNaoEnc.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.ConsNaoEnc.CNPJ == nil && doc.ConsNaoEnc.CPF == nil {
			return errors.New("parse mdfe: missing consult document")
		}
	}
	if doc.RetConsNaoEnc != nil {
		if doc.RetConsNaoEnc.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetConsNaoEnc.CUF == "" {
			return errors.New("parse mdfe: missing cUF")
		}
		if doc.RetConsNaoEnc.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
	}
	if doc.ConsReciMDFe != nil {
		if doc.ConsReciMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.ConsReciMDFe.NRec == "" {
			return errors.New("parse mdfe: missing nRec")
		}
	}
	if doc.RetConsReciMDFe != nil {
		if doc.RetConsReciMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetConsReciMDFe.NRec == "" {
			return errors.New("parse mdfe: missing nRec")
		}
		if doc.RetConsReciMDFe.CUF == "" {
			return errors.New("parse mdfe: missing cUF")
		}
		if doc.RetConsReciMDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
	}
	if doc.ConsSitMDFe != nil {
		if doc.ConsSitMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.ConsSitMDFe.ChMDFe == "" {
			return errors.New("parse mdfe: missing chMDFe")
		}
	}
	if doc.RetConsSitMDFe != nil {
		if doc.RetConsSitMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetConsSitMDFe.CUF == "" {
			return errors.New("parse mdfe: missing cUF")
		}
		if doc.RetConsSitMDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
	}
	if doc.ConsStatServMDFe != nil {
		if doc.ConsStatServMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
	}
	if doc.RetConsStatServMDFe != nil {
		if doc.RetConsStatServMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetConsStatServMDFe.CUF == "" {
			return errors.New("parse mdfe: missing cUF")
		}
		if doc.RetConsStatServMDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
		if doc.RetConsStatServMDFe.DhRecbto == "" {
			return errors.New("parse mdfe: missing dhRecbto")
		}
	}
	if err := validateEventRoots(doc); err != nil {
		return err
	}
	if doc.DistDFeInt != nil {
		if doc.DistDFeInt.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
			return errors.New("parse mdfe: missing dist document")
		}
		if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil {
			return errors.New("parse mdfe: missing dist query")
		}
	}
	if doc.RetDistDFeInt != nil {
		if doc.RetDistDFeInt.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetDistDFeInt.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
		if doc.RetDistDFeInt.UltNSU == "" {
			return errors.New("parse mdfe: missing ultNSU")
		}
		if doc.RetDistDFeInt.MaxNSU == "" {
			return errors.New("parse mdfe: missing maxNSU")
		}
	}
	if doc.DistMDFe != nil {
		if doc.DistMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.DistMDFe.UltNSU == "" {
			return errors.New("parse mdfe: missing ultNSU")
		}
	}
	if doc.RetDistMDFe != nil {
		if doc.RetDistMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetDistMDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
		}
	}
	if doc.MDFeConsultaDFe != nil {
		if doc.MDFeConsultaDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.MDFeConsultaDFe.ChMDFe == "" {
			return errors.New("parse mdfe: missing chMDFe")
		}
	}
	if doc.RetMDFeConsultaDFe != nil {
		if doc.RetMDFeConsultaDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.RetMDFeConsultaDFe.CStat == "" {
			return errors.New("parse mdfe: missing cStat")
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
	default:
		return nil
	}
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
		return errors.New("marshal mdfe: document must contain exactly one supported root")
	}

	switch {
	case d.EventoMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoMDFe.VersaoAttr), d.EventoMDFe.InfEvento, d.EventoMDFe.DsSignature)
	case d.EventoCancMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoCancMDFe.VersaoAttr), d.EventoCancMDFe.InfEvento, d.EventoCancMDFe.DsSignature)
	case d.EventoEncMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoEncMDFe.VersaoAttr), d.EventoEncMDFe.InfEvento, d.EventoEncMDFe.DsSignature)
	case d.EventoIncCondutorMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoIncCondutorMDFe.VersaoAttr), d.EventoIncCondutorMDFe.InfEvento, d.EventoIncCondutorMDFe.DsSignature)
	case d.EventoInclusaoDFeMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoInclusaoDFeMDFe.VersaoAttr), d.EventoInclusaoDFeMDFe.InfEvento, d.EventoInclusaoDFeMDFe.DsSignature)
	case d.EventoPagtoOperMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoPagtoOperMDFe.VersaoAttr), d.EventoPagtoOperMDFe.InfEvento, d.EventoPagtoOperMDFe.DsSignature)
	case d.EventoAlteracaoPagtoServMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoAlteracaoPagtoServMDFe.VersaoAttr), d.EventoAlteracaoPagtoServMDFe.InfEvento, d.EventoAlteracaoPagtoServMDFe.DsSignature)
	case d.EventoConfirmaServMDFe != nil:
		return encodeMDFeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoConfirmaServMDFe.VersaoAttr), d.EventoConfirmaServMDFe.InfEvento, d.EventoConfirmaServMDFe.DsSignature)
	default:
		return errors.New("marshal mdfe: document must contain exactly one supported root")
	}
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal mdfe: document must contain exactly one supported root")
	}

	switch {
	case d.RetEventoMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoMDFe.VersaoAttr), d.RetEventoMDFe.InfEvento)
	case d.RetEventoCancMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoCancMDFe.VersaoAttr), d.RetEventoCancMDFe.InfEvento)
	case d.RetEventoEncMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoEncMDFe.VersaoAttr), d.RetEventoEncMDFe.InfEvento)
	case d.RetEventoIncCondutorMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoIncCondutorMDFe.VersaoAttr), d.RetEventoIncCondutorMDFe.InfEvento)
	case d.RetEventoInclusaoDFeMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoInclusaoDFeMDFe.VersaoAttr), d.RetEventoInclusaoDFeMDFe.InfEvento)
	case d.RetEventoPagtoOperMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoPagtoOperMDFe.VersaoAttr), d.RetEventoPagtoOperMDFe.InfEvento)
	case d.RetEventoAlteracaoPagtoServMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoAlteracaoPagtoServMDFe.VersaoAttr), d.RetEventoAlteracaoPagtoServMDFe.InfEvento)
	case d.RetEventoConfirmaServMDFe != nil:
		return encodeMDFeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoConfirmaServMDFe.VersaoAttr), d.RetEventoConfirmaServMDFe.InfEvento)
	default:
		return errors.New("marshal mdfe: document must contain exactly one supported root")
	}
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal mdfe: document must contain exactly one supported root")
	}

	switch {
	case d.ProcEventoMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoMDFe.VersaoAttr), d.ProcEventoMDFe.IpTransmissorAttr, d.ProcEventoMDFe.NPortaConAttr, d.ProcEventoMDFe.DhConexaoAttr, d.ProcEventoMDFe.EventoMDFe, d.ProcEventoMDFe.RetEventoMDFe)
	case d.ProcEventoCancMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoCancMDFe.VersaoAttr), d.ProcEventoCancMDFe.IpTransmissorAttr, d.ProcEventoCancMDFe.NPortaConAttr, d.ProcEventoCancMDFe.DhConexaoAttr, d.ProcEventoCancMDFe.EventoMDFe, d.ProcEventoCancMDFe.RetEventoMDFe)
	case d.ProcEventoEncMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoEncMDFe.VersaoAttr), d.ProcEventoEncMDFe.IpTransmissorAttr, d.ProcEventoEncMDFe.NPortaConAttr, d.ProcEventoEncMDFe.DhConexaoAttr, d.ProcEventoEncMDFe.EventoMDFe, d.ProcEventoEncMDFe.RetEventoMDFe)
	case d.ProcEventoIncCondutorMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoIncCondutorMDFe.VersaoAttr), d.ProcEventoIncCondutorMDFe.IpTransmissorAttr, d.ProcEventoIncCondutorMDFe.NPortaConAttr, d.ProcEventoIncCondutorMDFe.DhConexaoAttr, d.ProcEventoIncCondutorMDFe.EventoMDFe, d.ProcEventoIncCondutorMDFe.RetEventoMDFe)
	case d.ProcEventoInclusaoDFeMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoInclusaoDFeMDFe.VersaoAttr), d.ProcEventoInclusaoDFeMDFe.IpTransmissorAttr, d.ProcEventoInclusaoDFeMDFe.NPortaConAttr, d.ProcEventoInclusaoDFeMDFe.DhConexaoAttr, d.ProcEventoInclusaoDFeMDFe.EventoMDFe, d.ProcEventoInclusaoDFeMDFe.RetEventoMDFe)
	case d.ProcEventoPagtoOperMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoPagtoOperMDFe.VersaoAttr), d.ProcEventoPagtoOperMDFe.IpTransmissorAttr, d.ProcEventoPagtoOperMDFe.NPortaConAttr, d.ProcEventoPagtoOperMDFe.DhConexaoAttr, d.ProcEventoPagtoOperMDFe.EventoMDFe, d.ProcEventoPagtoOperMDFe.RetEventoMDFe)
	case d.ProcEventoAlteracaoPagtoServMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoAlteracaoPagtoServMDFe.VersaoAttr), d.ProcEventoAlteracaoPagtoServMDFe.IpTransmissorAttr, d.ProcEventoAlteracaoPagtoServMDFe.NPortaConAttr, d.ProcEventoAlteracaoPagtoServMDFe.DhConexaoAttr, d.ProcEventoAlteracaoPagtoServMDFe.EventoMDFe, d.ProcEventoAlteracaoPagtoServMDFe.RetEventoMDFe)
	case d.ProcEventoConfirmaServMDFe != nil:
		return encodeMDFeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoConfirmaServMDFe.VersaoAttr), d.ProcEventoConfirmaServMDFe.IpTransmissorAttr, d.ProcEventoConfirmaServMDFe.NPortaConAttr, d.ProcEventoConfirmaServMDFe.DhConexaoAttr, d.ProcEventoConfirmaServMDFe.EventoMDFe, d.ProcEventoConfirmaServMDFe.RetEventoMDFe)
	default:
		return errors.New("marshal mdfe: document must contain exactly one supported root")
	}
}

func encodeMDFeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return e.Encode(struct {
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
	return e.Encode(struct {
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
	return e.Encode(struct {
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
	switch tpEvento {
	case "110111":
		var parsed cancelEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe cancelamento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoCancMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110112":
		var parsed encEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe encerramento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoEncMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110114":
		var parsed incCondutorEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe inclusao condutor: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoIncCondutorMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110115":
		var parsed inclusaoDFeEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe inclusao dfe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoInclusaoDFeMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110116":
		var parsed pagtoOperEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe pagamento operacao: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoPagtoOperMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110117":
		var parsed confirmaServEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe confirma servico: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoConfirmaServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110118":
		var parsed alteracaoPagtoServEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe alteracao pagamento servico: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoAlteracaoPagtoServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		var parsed eventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	}
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		var parsed cancelEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe cancelamento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoCancMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110112":
		var parsed encEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe encerramento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoEncMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110114":
		var parsed incCondutorEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe inclusao condutor: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoIncCondutorMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110115":
		var parsed inclusaoDFeEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe inclusao dfe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoInclusaoDFeMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110116":
		var parsed pagtoOperEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe pagamento operacao: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoPagtoOperMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110117":
		var parsed confirmaServEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe confirma servico: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoConfirmaServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110118":
		var parsed alteracaoPagtoServEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe alteracao pagamento servico: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoAlteracaoPagtoServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		var parsed eventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode retEventoMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	}
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		var parsed cancelEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe cancelamento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCancMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110112":
		var parsed encEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe encerramento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoEncMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110114":
		var parsed incCondutorEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe inclusao condutor: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoIncCondutorMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110115":
		var parsed inclusaoDFeEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe inclusao dfe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoInclusaoDFeMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110116":
		var parsed pagtoOperEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe pagamento operacao: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoPagtoOperMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110117":
		var parsed confirmaServEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe confirma servico: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoConfirmaServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110118":
		var parsed alteracaoPagtoServEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe alteracao pagamento servico: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoAlteracaoPagtoServMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		var parsed eventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode procEventoMDFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoMDFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	}
}

func activeRootCount(doc *Document) int {
	count := 0
	if doc.MDFe != nil {
		count++
	}
	if doc.MDFeProc != nil {
		count++
	}
	if doc.EnviMDFe != nil {
		count++
	}
	if doc.RetEnviMDFe != nil {
		count++
	}
	if doc.RetMDFe != nil {
		count++
	}
	if doc.ConsNaoEnc != nil {
		count++
	}
	if doc.RetConsNaoEnc != nil {
		count++
	}
	if doc.ConsReciMDFe != nil {
		count++
	}
	if doc.RetConsReciMDFe != nil {
		count++
	}
	if doc.ConsSitMDFe != nil {
		count++
	}
	if doc.RetConsSitMDFe != nil {
		count++
	}
	if doc.ConsStatServMDFe != nil {
		count++
	}
	if doc.RetConsStatServMDFe != nil {
		count++
	}
	if doc.EventoMDFe != nil {
		count++
	}
	if doc.RetEventoMDFe != nil {
		count++
	}
	if doc.ProcEventoMDFe != nil {
		count++
	}
	if doc.EventoCancMDFe != nil {
		count++
	}
	if doc.RetEventoCancMDFe != nil {
		count++
	}
	if doc.ProcEventoCancMDFe != nil {
		count++
	}
	if doc.EventoEncMDFe != nil {
		count++
	}
	if doc.RetEventoEncMDFe != nil {
		count++
	}
	if doc.ProcEventoEncMDFe != nil {
		count++
	}
	if doc.EventoIncCondutorMDFe != nil {
		count++
	}
	if doc.RetEventoIncCondutorMDFe != nil {
		count++
	}
	if doc.ProcEventoIncCondutorMDFe != nil {
		count++
	}
	if doc.EventoInclusaoDFeMDFe != nil {
		count++
	}
	if doc.RetEventoInclusaoDFeMDFe != nil {
		count++
	}
	if doc.ProcEventoInclusaoDFeMDFe != nil {
		count++
	}
	if doc.EventoPagtoOperMDFe != nil {
		count++
	}
	if doc.RetEventoPagtoOperMDFe != nil {
		count++
	}
	if doc.ProcEventoPagtoOperMDFe != nil {
		count++
	}
	if doc.EventoAlteracaoPagtoServMDFe != nil {
		count++
	}
	if doc.RetEventoAlteracaoPagtoServMDFe != nil {
		count++
	}
	if doc.ProcEventoAlteracaoPagtoServMDFe != nil {
		count++
	}
	if doc.EventoConfirmaServMDFe != nil {
		count++
	}
	if doc.RetEventoConfirmaServMDFe != nil {
		count++
	}
	if doc.ProcEventoConfirmaServMDFe != nil {
		count++
	}
	if doc.DistDFeInt != nil {
		count++
	}
	if doc.RetDistDFeInt != nil {
		count++
	}
	if doc.DistMDFe != nil {
		count++
	}
	if doc.RetDistMDFe != nil {
		count++
	}
	if doc.MDFeConsultaDFe != nil {
		count++
	}
	if doc.RetMDFeConsultaDFe != nil {
		count++
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
