package cte

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	distSchema "github.com/awa/nota-fiscal/internal/cte/gen/v1_0/dist_dfe"
	cteSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/cte"
	cteOSSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/cte_os"
	cancelEventSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/evento_cancel"
	eventSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/evento_cce"
)

const namespace = "http://www.portalfiscal.inf.br/cte"

type Document struct {
	VersaoAttr    string
	CTe           *cteSchema.TCTe
	CTeOS         *cteOSSchema.TCTeOS
	EventoCTe     *eventSchema.TEvento
	EventoCancCTe *cancelEventSchema.TEvento
	DistDFeInt    *distSchema.TAnonComplexDistDFeInt1
	RetDistDFeInt *distSchema.TAnonComplexRetDistDFeInt1
	rootName      string
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "CTe", "":
		if d.CTe != nil && d.CTeOS == nil && d.EventoCTe == nil {
			type root struct {
				XMLName     xml.Name                           `xml:"CTe"`
				XMLNS       string                             `xml:"xmlns,attr,omitempty"`
				InfCte      *cteSchema.TAnonComplexInfCte3     `xml:"infCte"`
				InfCTeSupl  *cteSchema.TAnonComplexInfCTeSupl3 `xml:"infCTeSupl,omitempty"`
				DsSignature *cteSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "CTe"},
				XMLNS:       namespace,
				InfCte:      d.CTe.InfCte,
				InfCTeSupl:  d.CTe.InfCTeSupl,
				DsSignature: d.CTe.DsSignature,
			})
		}
	case "CTeOS":
		if d.CTeOS != nil && d.CTe == nil && d.EventoCTe == nil {
			type root struct {
				XMLName     xml.Name                             `xml:"CTeOS"`
				XMLNS       string                               `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                               `xml:"versao,attr,omitempty"`
				InfCte      *cteOSSchema.TAnonComplexInfCte4     `xml:"infCte"`
				InfCTeSupl  *cteOSSchema.TAnonComplexInfCTeSupl4 `xml:"infCTeSupl,omitempty"`
				DsSignature *cteOSSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "CTeOS"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.CTeOS.VersaoAttr),
				InfCte:      d.CTeOS.InfCte,
				InfCTeSupl:  d.CTeOS.InfCTeSupl,
				DsSignature: d.CTeOS.DsSignature,
			})
		}
	case "eventoCTe":
		if d.EventoCTe != nil && d.CTe == nil && d.CTeOS == nil && d.EventoCancCTe == nil {
			type root struct {
				XMLName     xml.Name                            `xml:"eventoCTe"`
				XMLNS       string                              `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                              `xml:"versao,attr,omitempty"`
				InfEvento   *eventSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
				DsSignature *eventSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "eventoCTe"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoCTe.VersaoAttr),
				InfEvento:   d.EventoCTe.InfEvento,
				DsSignature: d.EventoCTe.DsSignature,
			})
		}
		if d.EventoCancCTe != nil && d.CTe == nil && d.CTeOS == nil && d.EventoCTe == nil {
			type root struct {
				XMLName     xml.Name                                  `xml:"eventoCTe"`
				XMLNS       string                                    `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                                    `xml:"versao,attr,omitempty"`
				InfEvento   *cancelEventSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
				DsSignature *cancelEventSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "eventoCTe"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoCancCTe.VersaoAttr),
				InfEvento:   d.EventoCancCTe.InfEvento,
				DsSignature: d.EventoCancCTe.DsSignature,
			})
		}
	case "distDFeInt":
		if d.DistDFeInt != nil && d.CTe == nil && d.CTeOS == nil && d.EventoCTe == nil && d.EventoCancCTe == nil && d.RetDistDFeInt == nil {
			type root struct {
				XMLName    xml.Name                         `xml:"distDFeInt"`
				XMLNS      string                           `xml:"xmlns,attr,omitempty"`
				VersaoAttr string                           `xml:"versao,attr,omitempty"`
				TpAmb      string                           `xml:"tpAmb"`
				CUFAutor   string                           `xml:"cUFAutor"`
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
				CUFAutor:   d.DistDFeInt.CUFAutor,
				CNPJ:       d.DistDFeInt.CNPJ,
				CPF:        d.DistDFeInt.CPF,
				DistNSU:    d.DistDFeInt.DistNSU,
				ConsNSU:    d.DistDFeInt.ConsNSU,
			})
		}
	case "retDistDFeInt":
		if d.RetDistDFeInt != nil && d.CTe == nil && d.CTeOS == nil && d.EventoCTe == nil && d.EventoCancCTe == nil && d.DistDFeInt == nil {
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

	return errors.New("marshal cte: document must contain exactly one supported root")
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse cte: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
	}

	switch rootName {
	case "CTe":
		var parsed cteSchema.TCTe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode CTe: %w", err)
		}
		doc := &Document{
			VersaoAttr: versionFromInfCte(parsed.InfCte),
			CTe:        &parsed,
			rootName:   rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "CTeOS":
		var parsed struct {
			VersaoAttr string                               `xml:"versao,attr"`
			InfCte     *cteOSSchema.TAnonComplexInfCte4     `xml:"infCte"`
			InfCTeSupl *cteOSSchema.TAnonComplexInfCTeSupl4 `xml:"infCTeSupl"`
			Signature  *cteOSSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
		}
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode CTeOS: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			CTeOS: &cteOSSchema.TCTeOS{
				VersaoAttr:  parsed.VersaoAttr,
				InfCte:      parsed.InfCte,
				InfCTeSupl:  parsed.InfCTeSupl,
				DsSignature: parsed.Signature,
			},
			rootName: rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "eventoCTe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse cte: missing infEvento")
		}
		switch tpEvento {
		case "110110":
			var parsed eventSchema.TEvento
			if err := xml.Unmarshal(data, &parsed); err != nil {
				return nil, fmt.Errorf("parse cte: decode eventoCTe cce: %w", err)
			}
			doc := &Document{
				VersaoAttr: parsed.VersaoAttr,
				EventoCTe:  &parsed,
				rootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110111":
			var parsed cancelEventSchema.TEvento
			if err := xml.Unmarshal(data, &parsed); err != nil {
				return nil, fmt.Errorf("parse cte: decode eventoCTe cancel: %w", err)
			}
			doc := &Document{
				VersaoAttr:    parsed.VersaoAttr,
				EventoCancCTe: &parsed,
				rootName:      rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		default:
			return nil, fmt.Errorf("parse cte: unsupported tpEvento %q", tpEvento)
		}
	case "distDFeInt":
		var parsed distSchema.TAnonComplexDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode distDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retDistDFeInt":
		var parsed distSchema.TAnonComplexRetDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retDistDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse cte: unsupported root element %q", rootName)
	}
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

func versionFromInfCte(inf *cteSchema.TAnonComplexInfCte3) string {
	if inf == nil {
		return ""
	}
	return inf.VersaoAttr
}

func validateDocument(doc *Document) error {
	count := 0
	if doc.CTe != nil {
		count++
		if err := validateInfCte(doc.CTe.InfCte); err != nil {
			return err
		}
	}
	if doc.CTeOS != nil {
		count++
		if err := validateInfCteOS(doc.CTeOS.InfCte); err != nil {
			return err
		}
	}
	if doc.EventoCTe != nil {
		count++
		if doc.EventoCTe.InfEvento == nil {
			return errors.New("parse cte: missing infEvento")
		}
		if doc.EventoCTe.InfEvento.ChCTe == "" {
			return errors.New("parse cte: missing chCTe")
		}
		if doc.EventoCTe.InfEvento.DetEvento == nil {
			return errors.New("parse cte: missing detEvento")
		}
	}
	if doc.EventoCancCTe != nil {
		count++
		if doc.EventoCancCTe.InfEvento == nil {
			return errors.New("parse cte: missing infEvento")
		}
		if doc.EventoCancCTe.InfEvento.ChCTe == "" {
			return errors.New("parse cte: missing chCTe")
		}
		if doc.EventoCancCTe.InfEvento.DetEvento == nil {
			return errors.New("parse cte: missing detEvento")
		}
	}
	if doc.DistDFeInt != nil {
		count++
		if doc.DistDFeInt.TpAmb == "" {
			return errors.New("parse cte: missing tpAmb")
		}
		if doc.DistDFeInt.CUFAutor == "" {
			return errors.New("parse cte: missing cUFAutor")
		}
		if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
			return errors.New("parse cte: missing dist document")
		}
		if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil {
			return errors.New("parse cte: missing dist query")
		}
	}
	if doc.RetDistDFeInt != nil {
		count++
		if doc.RetDistDFeInt.TpAmb == "" {
			return errors.New("parse cte: missing tpAmb")
		}
		if doc.RetDistDFeInt.CStat == "" {
			return errors.New("parse cte: missing cStat")
		}
		if doc.RetDistDFeInt.UltNSU == "" {
			return errors.New("parse cte: missing ultNSU")
		}
		if doc.RetDistDFeInt.MaxNSU == "" {
			return errors.New("parse cte: missing maxNSU")
		}
	}
	if count != 1 {
		return errors.New("parse cte: document must contain exactly one supported root")
	}
	return nil
}

func validateInfCte(inf *cteSchema.TAnonComplexInfCte3) error {
	if inf == nil {
		return errors.New("parse cte: missing infCte")
	}
	if inf.Ide == nil {
		return errors.New("parse cte: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse cte: missing emit")
	}
	if inf.Emit.CNPJ == nil && inf.Emit.CPF == nil {
		return errors.New("parse cte: missing emit document")
	}
	return nil
}

func validateInfCteOS(inf *cteOSSchema.TAnonComplexInfCte4) error {
	if inf == nil {
		return errors.New("parse cte: missing infCte")
	}
	if inf.Ide == nil {
		return errors.New("parse cte: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse cte: missing emit")
	}
	if inf.Emit.CNPJ == "" {
		return errors.New("parse cte: missing emit document")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
