package nfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	schema "github.com/awa/nota-fiscal/internal/nfe/gen/v4_0/nfe_proc"
)

type Document struct {
	VersaoAttr string
	NFe        *schema.TNFe
	ProtNFe    *schema.TProtNFe
	rootName   string
}

// MarshalXML preserves the parsed root when possible.
// If protocol data is present, the document is always encoded as nfeProc.
func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	if d.rootName != "nfeProc" && d.ProtNFe == nil {
		type bareNFe struct {
			XMLName xml.Name `xml:"NFe"`
			XMLNS   string   `xml:"xmlns,attr,omitempty"`
			*schema.TNFe
		}

		return e.Encode(bareNFe{
			XMLName: xml.Name{Local: "NFe"},
			XMLNS:   "http://www.portalfiscal.inf.br/nfe",
			TNFe:    d.NFe,
		})
	}

	type procNFe struct {
		XMLName    xml.Name         `xml:"nfeProc"`
		XMLNS      string           `xml:"xmlns,attr,omitempty"`
		VersaoAttr string           `xml:"versao,attr"`
		NFe        *schema.TNFe     `xml:"NFe"`
		ProtNFe    *schema.TProtNFe `xml:"protNFe"`
	}

	return e.Encode(procNFe{
		XMLName:    xml.Name{Local: "nfeProc"},
		XMLNS:      "http://www.portalfiscal.inf.br/nfe",
		VersaoAttr: d.VersaoAttr,
		NFe:        d.NFe,
		ProtNFe:    d.ProtNFe,
	})
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse nfe: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse nfe: read root: %w", rootErr)
	}

	switch rootName {
	case "nfeProc":
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
			rootName:   "nfeProc",
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
			rootName:   "NFe",
		}
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

func versionFromNFe(invoice *schema.TNFe) string {
	if invoice.InfNFe == nil {
		return ""
	}

	return invoice.InfNFe.VersaoAttr
}

func validateDocument(doc *Document) error {
	if doc.NFe == nil {
		return errors.New("parse nfe: missing NFe")
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
