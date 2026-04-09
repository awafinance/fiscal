package nfse

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	schema "github.com/awafinance/fiscal/internal/nfse/gen/v1_0/core"
)

const namespace = "http://www.sped.fazenda.gov.br/nfse"

type Document struct {
	VersaoAttr   string              `json:"versao,omitempty"`
	DPS          *schema.TCDPS       `json:"DPS,omitempty"`
	NFSe         *schema.TCNFSe      `json:"NFSe,omitempty"`
	PedRegEvento *schema.TCPedRegEvt `json:"pedRegEvento,omitempty"`
	rootName     string              `json:"-"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "DPS", "":
		if d.DPS != nil && d.NFSe == nil && d.PedRegEvento == nil {
			type root struct {
				XMLName    xml.Name              `xml:"DPS"`
				XMLNS      string                `xml:"xmlns,attr,omitempty"`
				VersaoAttr string                `xml:"versao,attr"`
				InfDPS     *schema.TCInfDPS      `xml:"infDPS"`
				Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:    xml.Name{Local: "DPS"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.DPS.VersaoAttr),
				InfDPS:     d.DPS.InfDPS,
				Signature:  d.DPS.DsSignature,
			})
		}
	case "NFSe":
		if d.NFSe != nil && d.DPS == nil && d.PedRegEvento == nil {
			type root struct {
				XMLName    xml.Name              `xml:"NFSe"`
				XMLNS      string                `xml:"xmlns,attr,omitempty"`
				VersaoAttr string                `xml:"versao,attr"`
				InfNFSe    *schema.TCInfNFSe     `xml:"infNFSe"`
				Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:    xml.Name{Local: "NFSe"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.NFSe.VersaoAttr),
				InfNFSe:    d.NFSe.InfNFSe,
				Signature:  d.NFSe.DsSignature,
			})
		}
	case "pedRegEvento":
		if d.PedRegEvento != nil && d.DPS == nil && d.NFSe == nil {
			type root struct {
				XMLName    xml.Name              `xml:"pedRegEvento"`
				XMLNS      string                `xml:"xmlns,attr,omitempty"`
				VersaoAttr string                `xml:"versao,attr"`
				InfPedReg  *schema.TCInfPedReg   `xml:"infPedReg"`
				Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:    xml.Name{Local: "pedRegEvento"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.PedRegEvento.VersaoAttr),
				InfPedReg:  d.PedRegEvento.InfPedReg,
				Signature:  d.PedRegEvento.DsSignature,
			})
		}
	}

	return errors.New("marshal nfse: document must contain exactly one supported root")
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse nfse: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse nfse: read root: %w", rootErr)
	}

	switch rootName {
	case "DPS":
		var parsed schema.TCDPS
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfse: decode DPS: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DPS: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "NFSe":
		var parsed schema.TCNFSe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfse: decode NFSe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, NFSe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "pedRegEvento":
		var parsed schema.TCPedRegEvt
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfse: decode pedRegEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, PedRegEvento: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse nfse: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse nfse: unsupported root element %q", rootName)
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
	if doc.DPS != nil {
		count++
		if doc.DPS.InfDPS == nil {
			return errors.New("parse nfse: missing infDPS")
		}
		if doc.DPS.InfDPS.Prest == nil {
			return errors.New("parse nfse: missing prest")
		}
		if doc.DPS.InfDPS.Serv == nil {
			return errors.New("parse nfse: missing serv")
		}
	}
	if doc.NFSe != nil {
		count++
		if doc.NFSe.InfNFSe == nil {
			return errors.New("parse nfse: missing infNFSe")
		}
		if doc.NFSe.InfNFSe.Emit == nil {
			return errors.New("parse nfse: missing emit")
		}
		if doc.NFSe.InfNFSe.DPS == nil {
			return errors.New("parse nfse: missing DPS")
		}
	}
	if doc.PedRegEvento != nil {
		count++
		if doc.PedRegEvento.InfPedReg == nil {
			return errors.New("parse nfse: missing infPedReg")
		}
		if doc.PedRegEvento.InfPedReg.ChNFSe == "" {
			return errors.New("parse nfse: missing chNFSe")
		}
	}
	if count != 1 {
		return errors.New("parse nfse: document must contain exactly one supported root")
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
