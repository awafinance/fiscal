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
	eventSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/evento"
	mdfeSchema "github.com/awa/nota-fiscal/internal/mdfe/gen/v3_0/mdfe"
)

const namespace = "http://www.portalfiscal.inf.br/mdfe"

type Document struct {
	VersaoAttr    string
	MDFe          *mdfeSchema.TMDFe
	ConsNaoEnc    *consNaoEncSchema.TConsMDFeNaoEnc
	ConsReciMDFe  *consReciSchema.TConsReciMDFe
	EventoMDFe    *eventSchema.TEvento
	DistDFeInt    *distSchema.TAnonComplexDistDFeInt1
	RetDistDFeInt *distSchema.TAnonComplexRetDistDFeInt1
	rootName      string
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "MDFe", "":
		if d.MDFe != nil && d.ConsNaoEnc == nil && d.ConsReciMDFe == nil && d.EventoMDFe == nil {
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
	case "consMDFeNaoEnc":
		if d.ConsNaoEnc != nil && d.MDFe == nil && d.ConsReciMDFe == nil && d.EventoMDFe == nil {
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
	case "consReciMDFe":
		if d.ConsReciMDFe != nil && d.MDFe == nil && d.ConsNaoEnc == nil && d.EventoMDFe == nil {
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
	case "eventoMDFe":
		if d.EventoMDFe != nil && d.MDFe == nil && d.ConsNaoEnc == nil && d.ConsReciMDFe == nil {
			type root struct {
				XMLName     xml.Name                            `xml:"eventoMDFe"`
				XMLNS       string                              `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                              `xml:"versao,attr,omitempty"`
				InfEvento   *eventSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
				DsSignature *eventSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "eventoMDFe"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoMDFe.VersaoAttr),
				InfEvento:   d.EventoMDFe.InfEvento,
				DsSignature: d.EventoMDFe.DsSignature,
			})
		}
	case "distDFeInt":
		if d.DistDFeInt != nil && d.MDFe == nil && d.ConsNaoEnc == nil && d.ConsReciMDFe == nil && d.EventoMDFe == nil && d.RetDistDFeInt == nil {
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
		if d.RetDistDFeInt != nil && d.MDFe == nil && d.ConsNaoEnc == nil && d.ConsReciMDFe == nil && d.EventoMDFe == nil && d.DistDFeInt == nil {
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
		doc := &Document{
			VersaoAttr: versionFromMDFe(&parsed),
			MDFe:       &parsed,
			rootName:   rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consMDFeNaoEnc":
		var parsed consNaoEncSchema.TConsMDFeNaoEnc
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode consMDFeNaoEnc: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			ConsNaoEnc: &parsed,
			rootName:   rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consReciMDFe":
		var parsed consReciSchema.TConsReciMDFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode consReciMDFe: %w", err)
		}
		doc := &Document{
			VersaoAttr:   parsed.VersaoAttr,
			ConsReciMDFe: &parsed,
			rootName:     rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "eventoMDFe":
		var parsed eventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse mdfe: decode eventoMDFe: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			EventoMDFe: &parsed,
			rootName:   rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
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
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse mdfe: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse mdfe: unsupported root element %q", rootName)
	}
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
	count := 0

	if doc.MDFe != nil {
		count++
		if doc.MDFe.InfMDFe == nil {
			return errors.New("parse mdfe: missing infMDFe")
		}
		if doc.MDFe.InfMDFe.Ide == nil {
			return errors.New("parse mdfe: missing ide")
		}
		if doc.MDFe.InfMDFe.Emit == nil {
			return errors.New("parse mdfe: missing emit")
		}
		if doc.MDFe.InfMDFe.Emit.CNPJ == nil && doc.MDFe.InfMDFe.Emit.CPF == nil {
			return errors.New("parse mdfe: missing emit document")
		}
		if doc.MDFe.InfMDFe.InfModal == nil {
			return errors.New("parse mdfe: missing infModal")
		}
	}

	if doc.ConsNaoEnc != nil {
		count++
		if doc.ConsNaoEnc.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.ConsNaoEnc.CNPJ == nil && doc.ConsNaoEnc.CPF == nil {
			return errors.New("parse mdfe: missing consult document")
		}
	}

	if doc.ConsReciMDFe != nil {
		count++
		if doc.ConsReciMDFe.TpAmb == "" {
			return errors.New("parse mdfe: missing tpAmb")
		}
		if doc.ConsReciMDFe.NRec == "" {
			return errors.New("parse mdfe: missing nRec")
		}
	}

	if doc.EventoMDFe != nil {
		count++
		if doc.EventoMDFe.InfEvento == nil {
			return errors.New("parse mdfe: missing infEvento")
		}
		if doc.EventoMDFe.InfEvento.ChMDFe == "" {
			return errors.New("parse mdfe: missing chMDFe")
		}
		if doc.EventoMDFe.InfEvento.DetEvento == nil {
			return errors.New("parse mdfe: missing detEvento")
		}
	}

	if doc.DistDFeInt != nil {
		count++
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
		count++
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

	if count != 1 {
		return errors.New("parse mdfe: document must contain exactly one supported root")
	}

	return nil
}

func versionFromMDFe(doc *mdfeSchema.TMDFe) string {
	if doc == nil || doc.InfMDFe == nil {
		return ""
	}
	return doc.InfMDFe.VersaoAttr
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
