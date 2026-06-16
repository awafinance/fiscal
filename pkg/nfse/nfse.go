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
	EventoNFSe   *schema.TCEvento    `json:"evento,omitempty"`
	RootName     string              `json:"rootName,omitempty"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.RootName {
	case "DPS", "":
		if d.hasOnlyDPS() {
			return d.marshalDPS(e)
		}
	case "NFSe":
		if d.hasOnlyNFSe() {
			return d.marshalNFSe(e)
		}
	case "pedRegEvento":
		if d.hasOnlyPedRegEvento() {
			return d.marshalPedRegEvento(e)
		}
	case "evento":
		if d.hasOnlyEventoNFSe() {
			return d.marshalEventoNFSe(e)
		}
	}

	return errors.New("marshal nfse: document must contain exactly one supported root")
}

func (d *Document) hasOnlyDPS() bool {
	return d.DPS != nil && d.NFSe == nil && d.PedRegEvento == nil && d.EventoNFSe == nil
}

func (d *Document) hasOnlyNFSe() bool {
	return d.NFSe != nil && d.DPS == nil && d.PedRegEvento == nil && d.EventoNFSe == nil
}

func (d *Document) hasOnlyPedRegEvento() bool {
	return d.PedRegEvento != nil && d.DPS == nil && d.NFSe == nil && d.EventoNFSe == nil
}

func (d *Document) hasOnlyEventoNFSe() bool {
	return d.EventoNFSe != nil && d.DPS == nil && d.NFSe == nil && d.PedRegEvento == nil
}

func (d *Document) marshalDPS(e *xml.Encoder) error {
	type root struct {
		XMLName    xml.Name              `xml:"DPS"`
		XMLNS      string                `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                `xml:"versao,attr"`
		InfDPS     *schema.TCInfDPS      `xml:"infDPS"`
		Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}

	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "DPS"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.DPS.VersaoAttr),
		InfDPS:     d.DPS.InfDPS,
		Signature:  d.DPS.DsSignature,
	})
}

func (d *Document) marshalNFSe(e *xml.Encoder) error {
	type root struct {
		XMLName    xml.Name              `xml:"NFSe"`
		XMLNS      string                `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                `xml:"versao,attr"`
		InfNFSe    *schema.TCInfNFSe     `xml:"infNFSe"`
		Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}

	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "NFSe"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.NFSe.VersaoAttr),
		InfNFSe:    d.NFSe.InfNFSe,
		Signature:  d.NFSe.DsSignature,
	})
}

func (d *Document) marshalPedRegEvento(e *xml.Encoder) error {
	type root struct {
		XMLName    xml.Name              `xml:"pedRegEvento"`
		XMLNS      string                `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                `xml:"versao,attr"`
		InfPedReg  *schema.TCInfPedReg   `xml:"infPedReg"`
		Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}

	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "pedRegEvento"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.PedRegEvento.VersaoAttr),
		InfPedReg:  d.PedRegEvento.InfPedReg,
		Signature:  d.PedRegEvento.DsSignature,
	})
}

func (d *Document) marshalEventoNFSe(e *xml.Encoder) error {
	type root struct {
		XMLName    xml.Name              `xml:"evento"`
		XMLNS      string                `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                `xml:"versao,attr"`
		InfEvento  *schema.TCInfEvento   `xml:"infEvento"`
		Signature  *schema.SignatureType `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}

	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "evento"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoNFSe.VersaoAttr),
		InfEvento:  d.EventoNFSe.InfEvento,
		Signature:  d.EventoNFSe.DsSignature,
	})
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
		return parseDPS(data, RootName)
	case "NFSe":
		return parseNFSe(data, RootName)
	case "pedRegEvento":
		return parsePedRegEvento(data, RootName)
	case "evento":
		return parseEventoNFSe(data, RootName)
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse nfse: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse nfse: %w", &fiscalerr.UnsupportedRootError{Family: fiscalerr.NFSe, Root: RootName})
	}
}

func parseDPS(data []byte, rootName string) (*Document, error) {
	var parsed schema.TCDPS
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfse: decode DPS: %w", err)
	}
	return validateParsedDocument(&Document{VersaoAttr: parsed.VersaoAttr, DPS: &parsed, RootName: rootName})
}

func parseNFSe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TCNFSe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfse: decode NFSe: %w", err)
	}
	return validateParsedDocument(&Document{VersaoAttr: parsed.VersaoAttr, NFSe: &parsed, RootName: rootName})
}

func parsePedRegEvento(data []byte, rootName string) (*Document, error) {
	var parsed schema.TCPedRegEvt
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfse: decode pedRegEvento: %w", err)
	}
	return validateParsedDocument(&Document{VersaoAttr: parsed.VersaoAttr, PedRegEvento: &parsed, RootName: rootName})
}

func parseEventoNFSe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TCEvento
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfse: decode evento: %w", err)
	}
	return validateParsedDocument(&Document{VersaoAttr: parsed.VersaoAttr, EventoNFSe: &parsed, RootName: rootName})
}

func validateParsedDocument(doc *Document) (*Document, error) {
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
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
		if err := validateDPS(doc.DPS); err != nil {
			return err
		}
	}
	if doc.NFSe != nil {
		count++
		if err := validateNFSe(doc.NFSe); err != nil {
			return err
		}
	}
	if doc.PedRegEvento != nil {
		count++
		if err := validatePedRegEvento(doc.PedRegEvento); err != nil {
			return err
		}
	}
	if doc.EventoNFSe != nil {
		count++
		if err := validateEventoNFSe(doc.EventoNFSe); err != nil {
			return err
		}
	}
	if count != 1 {
		return errors.New("parse nfse: document must contain exactly one supported root")
	}
	return nil
}

func validateDPS(dps *schema.TCDPS) error {
	if dps.InfDPS == nil {
		return errors.New("parse nfse: missing infDPS")
	}
	if dps.InfDPS.Prest == nil {
		return errors.New("parse nfse: missing prest")
	}
	if dps.InfDPS.Serv == nil {
		return errors.New("parse nfse: missing serv")
	}
	return nil
}

func validateNFSe(nfse *schema.TCNFSe) error {
	if nfse.InfNFSe == nil {
		return errors.New("parse nfse: missing infNFSe")
	}
	if nfse.InfNFSe.Emit == nil {
		return errors.New("parse nfse: missing emit")
	}
	if nfse.InfNFSe.DPS == nil {
		return errors.New("parse nfse: missing DPS")
	}
	return nil
}

func validatePedRegEvento(evento *schema.TCPedRegEvt) error {
	if evento.InfPedReg == nil {
		return errors.New("parse nfse: missing infPedReg")
	}
	if evento.InfPedReg.ChNFSe == "" {
		return errors.New("parse nfse: missing chNFSe")
	}
	return nil
}

func validateEventoNFSe(evento *schema.TCEvento) error {
	if evento.InfEvento == nil {
		return errors.New("parse nfse: missing infEvento")
	}
	if evento.InfEvento.PedRegEvento == nil {
		return errors.New("parse nfse: missing infPedReg")
	}
	return validatePedRegEvento(evento.InfEvento.PedRegEvento)
}
