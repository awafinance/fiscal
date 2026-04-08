package cte

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	cteSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/cte"
	cteOSSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/cte_os"
	eventSchema "github.com/awa/nota-fiscal/internal/cte/gen/v4_0/evento_cce"
)

const namespace = "http://www.portalfiscal.inf.br/cte"

type Document struct {
	VersaoAttr string
	CTe        *cteSchema.TCTe
	CTeOS      *cteOSSchema.TCTeOS
	EventoCTe  *eventSchema.TEvento
	rootName   string
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
		if d.EventoCTe != nil && d.CTe == nil && d.CTeOS == nil {
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
		var parsed eventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe: %w", err)
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
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse cte: unsupported root element %q", rootName)
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
