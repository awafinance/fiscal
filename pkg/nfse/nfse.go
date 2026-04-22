package nfse

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	schema "github.com/awafinance/fiscal/internal/nfse/gen/v1_0/core"
	"github.com/awafinance/fiscal/internal/xmlutil"
	"github.com/awafinance/fiscal/pkg/fiscalerr"
)

const namespace = "http://www.sped.fazenda.gov.br/nfse"

type Document struct {
	VersaoAttr   string              `json:"versao,omitempty"`
	DPS          *schema.TCDPS       `json:"DPS,omitempty"`
	NFSe         *schema.TCNFSe      `json:"NFSe,omitempty"`
	PedRegEvento *schema.TCPedRegEvt `json:"pedRegEvento,omitempty"`
	RootName     string              `json:"rootName,omitempty"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	encode := func(v any) error { return xmlutil.EncodeCanonical(e, v) }

	switch d.RootName {
	case "DPS", "":
		if d.DPS != nil && d.NFSe == nil && d.PedRegEvento == nil {
			type root struct {
				XMLName    xml.Name              `xml:"DPS"`
				XMLNS      string                `xml:"xmlns,attr,omitempty"`
				VersaoAttr string                `xml:"versao,attr"`
				InfDPS     *schema.TCInfDPS      `xml:"infDPS"`
				Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return encode(root{
				XMLName:    xml.Name{Local: "DPS"},
				XMLNS:      namespace,
				VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.DPS.VersaoAttr),
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

			return encode(root{
				XMLName:    xml.Name{Local: "NFSe"},
				XMLNS:      namespace,
				VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.NFSe.VersaoAttr),
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

			return encode(root{
				XMLName:    xml.Name{Local: "pedRegEvento"},
				XMLNS:      namespace,
				VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.PedRegEvento.VersaoAttr),
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
		return nil, fmt.Errorf("parse nfse: %w", fiscalerr.ErrEmptyDocument)
	}

	RootName, rootErr := xmlutil.ParseRootName(data)
	if rootErr != nil && RootName == "" {
		return nil, fmt.Errorf("parse nfse: read root: %w", rootErr)
	}

	switch RootName {
	case "DPS":
		var parsed schema.TCDPS
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfse: decode DPS: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DPS: &parsed, RootName: RootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "NFSe":
		var parsed schema.TCNFSe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfse: decode NFSe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, NFSe: &parsed, RootName: RootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "pedRegEvento":
		var parsed schema.TCPedRegEvt
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfse: decode pedRegEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, PedRegEvento: &parsed, RootName: RootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse nfse: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse nfse: %w", &fiscalerr.UnsupportedRootError{Family: fiscalerr.NFSe, Root: RootName})
	}
}

func ParseReader(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parse nfse: read xml: %w", err)
	}
	return Parse(data)
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
