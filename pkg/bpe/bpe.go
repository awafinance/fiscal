package bpe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	schema "github.com/awa/nota-fiscal/internal/bpe/gen/v1_0/core"
)

const namespace = "http://www.portalfiscal.inf.br/bpe"

type Document struct {
	VersaoAttr         string
	BPe                *schema.TBPe
	BPeTM              *schema.TBPeTM
	BPeProc            *schema.TAnonComplexBpeProc1
	BPeTMProc          *schema.TAnonComplexBpeTMProc1
	RetBPe             *schema.TRetBPe
	ConsSitBPe         *schema.TConsSitBPe
	RetConsSitBPe      *schema.TRetConsSitBPe
	ConsStatServBPe    *schema.TConsStatServ
	RetConsStatServBPe *schema.TRetConsStatServ
	EventoBPe          *schema.TEvento
	RetEventoBPe       *schema.TRetEvento
	ProcEventoBPe      *schema.TProcEvento
	rootName           string
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "BPe", "":
		if d.BPe != nil && onlyBPeRoot(d) {
			type root struct {
				XMLName     xml.Name                        `xml:"BPe"`
				XMLNS       string                          `xml:"xmlns,attr,omitempty"`
				InfBPe      *schema.TAnonComplexInfBPe2     `xml:"infBPe"`
				InfBPeSupl  *schema.TAnonComplexInfBPeSupl1 `xml:"infBPeSupl,omitempty"`
				DsSignature *schema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}
			return e.Encode(root{
				XMLName:     xml.Name{Local: "BPe"},
				XMLNS:       namespace,
				InfBPe:      d.BPe.InfBPe,
				InfBPeSupl:  d.BPe.InfBPeSupl,
				DsSignature: d.BPe.DsSignature,
			})
		}
	case "BPeTM":
		if d.BPeTM != nil && onlyBPeTMRoot(d) {
			type root struct {
				XMLName     xml.Name                    `xml:"BPeTM"`
				XMLNS       string                      `xml:"xmlns,attr,omitempty"`
				InfBPe      *schema.TAnonComplexInfBPe1 `xml:"infBPe"`
				DsSignature *schema.SignatureType       `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}
			return e.Encode(root{
				XMLName:     xml.Name{Local: "BPeTM"},
				XMLNS:       namespace,
				InfBPe:      d.BPeTM.InfBPe,
				DsSignature: d.BPeTM.DsSignature,
			})
		}
	case "bpeProc":
		if d.BPeProc != nil && onlyBPeProcRoot(d) {
			type root struct {
				XMLName           xml.Name         `xml:"bpeProc"`
				XMLNS             string           `xml:"xmlns,attr,omitempty"`
				VersaoAttr        string           `xml:"versao,attr,omitempty"`
				IpTransmissorAttr *string          `xml:"ipTransmissor,attr,omitempty"`
				NPortaConAttr     *string          `xml:"nPortaCon,attr,omitempty"`
				DhConexaoAttr     *string          `xml:"dhConexao,attr,omitempty"`
				BPe               *schema.TBPe     `xml:"BPe"`
				ProtBPe           *schema.TProtBPe `xml:"protBPe"`
			}
			return e.Encode(root{
				XMLName:           xml.Name{Local: "bpeProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.BPeProc.VersaoAttr),
				IpTransmissorAttr: d.BPeProc.IpTransmissorAttr,
				NPortaConAttr:     d.BPeProc.NPortaConAttr,
				DhConexaoAttr:     d.BPeProc.DhConexaoAttr,
				BPe:               d.BPeProc.BPe,
				ProtBPe:           d.BPeProc.ProtBPe,
			})
		}
	case "bpeTMProc":
		if d.BPeTMProc != nil && onlyBPeTMProcRoot(d) {
			type root struct {
				XMLName           xml.Name         `xml:"bpeTMProc"`
				XMLNS             string           `xml:"xmlns,attr,omitempty"`
				VersaoAttr        string           `xml:"versao,attr,omitempty"`
				IpTransmissorAttr *string          `xml:"ipTransmissor,attr,omitempty"`
				NPortaConAttr     *string          `xml:"nPortaCon,attr,omitempty"`
				DhConexaoAttr     *string          `xml:"dhConexao,attr,omitempty"`
				BPeTM             *schema.TBPeTM   `xml:"BPeTM"`
				ProtBPe           *schema.TProtBPe `xml:"protBPe"`
			}
			return e.Encode(root{
				XMLName:           xml.Name{Local: "bpeTMProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.BPeTMProc.VersaoAttr),
				IpTransmissorAttr: d.BPeTMProc.IpTransmissorAttr,
				NPortaConAttr:     d.BPeTMProc.NPortaConAttr,
				DhConexaoAttr:     d.BPeTMProc.DhConexaoAttr,
				BPeTM:             d.BPeTMProc.BPeTM,
				ProtBPe:           d.BPeTMProc.ProtBPe,
			})
		}
	case "retBPe":
		if d.RetBPe != nil && onlyRetBPeRoot(d) {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retBPe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*schema.TRetBPe
			}{
				XMLName: xml.Name{Local: "retBPe"},
				XMLNS:   namespace,
				TRetBPe: d.RetBPe,
			})
		}
	case "consSitBPe":
		if d.ConsSitBPe != nil && onlyConsSitRoot(d) {
			return e.Encode(struct {
				XMLName xml.Name `xml:"consSitBPe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*schema.TConsSitBPe
			}{
				XMLName:     xml.Name{Local: "consSitBPe"},
				XMLNS:       namespace,
				TConsSitBPe: d.ConsSitBPe,
			})
		}
	case "retConsSitBPe":
		if d.RetConsSitBPe != nil && onlyRetConsSitRoot(d) {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsSitBPe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*schema.TRetConsSitBPe
			}{
				XMLName:        xml.Name{Local: "retConsSitBPe"},
				XMLNS:          namespace,
				TRetConsSitBPe: d.RetConsSitBPe,
			})
		}
	case "consStatServBPe":
		if d.ConsStatServBPe != nil && onlyConsStatRoot(d) {
			return e.Encode(struct {
				XMLName xml.Name `xml:"consStatServBPe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*schema.TConsStatServ
			}{
				XMLName:       xml.Name{Local: "consStatServBPe"},
				XMLNS:         namespace,
				TConsStatServ: d.ConsStatServBPe,
			})
		}
	case "retConsStatServBPe":
		if d.RetConsStatServBPe != nil && onlyRetConsStatRoot(d) {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsStatServBPe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*schema.TRetConsStatServ
			}{
				XMLName:          xml.Name{Local: "retConsStatServBPe"},
				XMLNS:            namespace,
				TRetConsStatServ: d.RetConsStatServBPe,
			})
		}
	case "eventoBPe":
		if d.EventoBPe != nil && onlyEventoRoot(d) {
			type root struct {
				XMLName     xml.Name                       `xml:"eventoBPe"`
				XMLNS       string                         `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                         `xml:"versao,attr,omitempty"`
				InfEvento   *schema.TAnonComplexInfEvento1 `xml:"infEvento"`
				DsSignature *schema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}
			return e.Encode(root{
				XMLName:     xml.Name{Local: "eventoBPe"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoBPe.VersaoAttr),
				InfEvento:   d.EventoBPe.InfEvento,
				DsSignature: d.EventoBPe.DsSignature,
			})
		}
	case "retEventoBPe":
		if d.RetEventoBPe != nil && onlyRetEventoRoot(d) {
			type root struct {
				XMLName     xml.Name                       `xml:"retEventoBPe"`
				XMLNS       string                         `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                         `xml:"versao,attr,omitempty"`
				InfEvento   *schema.TAnonComplexInfEvento2 `xml:"infEvento"`
				DsSignature *schema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}
			return e.Encode(root{
				XMLName:     xml.Name{Local: "retEventoBPe"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.RetEventoBPe.VersaoAttr),
				InfEvento:   d.RetEventoBPe.InfEvento,
				DsSignature: d.RetEventoBPe.DsSignature,
			})
		}
	case "procEventoBPe":
		if d.ProcEventoBPe != nil && onlyProcEventoRoot(d) {
			type root struct {
				XMLName           xml.Name           `xml:"procEventoBPe"`
				XMLNS             string             `xml:"xmlns,attr,omitempty"`
				VersaoAttr        string             `xml:"versao,attr,omitempty"`
				IpTransmissorAttr *string            `xml:"ipTransmissor,attr,omitempty"`
				NPortaConAttr     *string            `xml:"nPortaCon,attr,omitempty"`
				DhConexaoAttr     *string            `xml:"dhConexao,attr,omitempty"`
				EventoBPe         *schema.TEvento    `xml:"eventoBPe"`
				RetEventoBPe      *schema.TRetEvento `xml:"retEventoBPe"`
			}
			return e.Encode(root{
				XMLName:           xml.Name{Local: "procEventoBPe"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.ProcEventoBPe.VersaoAttr),
				IpTransmissorAttr: d.ProcEventoBPe.IpTransmissorAttr,
				NPortaConAttr:     d.ProcEventoBPe.NPortaConAttr,
				DhConexaoAttr:     d.ProcEventoBPe.DhConexaoAttr,
				EventoBPe:         d.ProcEventoBPe.EventoBPe,
				RetEventoBPe:      d.ProcEventoBPe.RetEventoBPe,
			})
		}
	}

	return errors.New("marshal bpe: document must contain exactly one supported root")
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse bpe: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse bpe: read root: %w", rootErr)
	}

	switch rootName {
	case "BPe":
		var parsed schema.TBPe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode BPe: %w", err)
		}
		doc := &Document{VersaoAttr: versionFromBPe(parsed.InfBPe), BPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "BPeTM":
		var parsed schema.TBPeTM
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode BPeTM: %w", err)
		}
		doc := &Document{VersaoAttr: versionFromBPeTM(parsed.InfBPe), BPeTM: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "bpeProc":
		var parsed schema.TAnonComplexBpeProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode bpeProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, BPeProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "bpeTMProc":
		var parsed schema.TAnonComplexBpeTMProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode bpeTMProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, BPeTMProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retBPe":
		var parsed schema.TRetBPe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consSitBPe":
		var parsed schema.TConsSitBPe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode consSitBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsSitBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsSitBPe":
		var parsed schema.TRetConsSitBPe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retConsSitBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsSitBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consStatServBPe":
		var parsed schema.TConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode consStatServBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsStatServBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsStatServBPe":
		var parsed schema.TRetConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retConsStatServBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "eventoBPe":
		var parsed schema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retEventoBPe":
		var parsed schema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "procEventoBPe":
		var parsed schema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse bpe: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse bpe: unsupported root element %q", rootName)
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

	if doc.BPe != nil {
		count++
		if err := validateInfBPe(doc.BPe.InfBPe); err != nil {
			return err
		}
	}
	if doc.BPeTM != nil {
		count++
		if err := validateInfBPeTM(doc.BPeTM.InfBPe); err != nil {
			return err
		}
	}
	if doc.BPeProc != nil {
		count++
		if doc.BPeProc.BPe == nil {
			return errors.New("parse bpe: missing BPe")
		}
		if doc.BPeProc.ProtBPe == nil {
			return errors.New("parse bpe: missing protBPe")
		}
	}
	if doc.BPeTMProc != nil {
		count++
		if doc.BPeTMProc.BPeTM == nil {
			return errors.New("parse bpe: missing BPeTM")
		}
		if doc.BPeTMProc.ProtBPe == nil {
			return errors.New("parse bpe: missing protBPe")
		}
	}
	if doc.RetBPe != nil {
		count++
		if doc.RetBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
		if doc.RetBPe.CUF == "" {
			return errors.New("parse bpe: missing cUF")
		}
		if doc.RetBPe.CStat == "" {
			return errors.New("parse bpe: missing cStat")
		}
	}
	if doc.ConsSitBPe != nil {
		count++
		if doc.ConsSitBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
		if doc.ConsSitBPe.ChBPe == "" {
			return errors.New("parse bpe: missing chBPe")
		}
	}
	if doc.RetConsSitBPe != nil {
		count++
		if doc.RetConsSitBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
		if doc.RetConsSitBPe.CUF == "" {
			return errors.New("parse bpe: missing cUF")
		}
		if doc.RetConsSitBPe.CStat == "" {
			return errors.New("parse bpe: missing cStat")
		}
	}
	if doc.ConsStatServBPe != nil {
		count++
		if doc.ConsStatServBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
	}
	if doc.RetConsStatServBPe != nil {
		count++
		if doc.RetConsStatServBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
		if doc.RetConsStatServBPe.CUF == "" {
			return errors.New("parse bpe: missing cUF")
		}
		if doc.RetConsStatServBPe.CStat == "" {
			return errors.New("parse bpe: missing cStat")
		}
		if doc.RetConsStatServBPe.DhRecbto == "" {
			return errors.New("parse bpe: missing dhRecbto")
		}
	}
	if doc.EventoBPe != nil {
		count++
		if err := validateEvento(doc.EventoBPe.InfEvento); err != nil {
			return err
		}
	}
	if doc.RetEventoBPe != nil {
		count++
		if doc.RetEventoBPe.InfEvento == nil {
			return errors.New("parse bpe: missing infEvento")
		}
		if doc.RetEventoBPe.InfEvento.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
		if doc.RetEventoBPe.InfEvento.CStat == "" {
			return errors.New("parse bpe: missing cStat")
		}
	}
	if doc.ProcEventoBPe != nil {
		count++
		if doc.ProcEventoBPe.EventoBPe == nil {
			return errors.New("parse bpe: missing eventoBPe")
		}
		if doc.ProcEventoBPe.RetEventoBPe == nil {
			return errors.New("parse bpe: missing retEventoBPe")
		}
	}

	if count != 1 {
		return errors.New("parse bpe: document must contain exactly one supported root")
	}
	return nil
}

func validateInfBPe(inf *schema.TAnonComplexInfBPe2) error {
	if inf == nil {
		return errors.New("parse bpe: missing infBPe")
	}
	if inf.Ide == nil {
		return errors.New("parse bpe: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse bpe: missing emit")
	}
	if inf.Emit.CNPJ == "" {
		return errors.New("parse bpe: missing emit document")
	}
	if len(inf.InfViagem) == 0 {
		return errors.New("parse bpe: missing infViagem")
	}
	return nil
}

func validateInfBPeTM(inf *schema.TAnonComplexInfBPe1) error {
	if inf == nil {
		return errors.New("parse bpe: missing infBPe")
	}
	if inf.Ide == nil {
		return errors.New("parse bpe: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse bpe: missing emit")
	}
	if inf.Emit.CNPJ == "" {
		return errors.New("parse bpe: missing emit document")
	}
	if len(inf.DetBPeTM) == 0 {
		return errors.New("parse bpe: missing detBPeTM")
	}
	return nil
}

func validateEvento(inf *schema.TAnonComplexInfEvento1) error {
	if inf == nil {
		return errors.New("parse bpe: missing infEvento")
	}
	if inf.ChBPe == "" {
		return errors.New("parse bpe: missing chBPe")
	}
	if inf.DetEvento == nil {
		return errors.New("parse bpe: missing detEvento")
	}
	return nil
}

func versionFromBPe(inf *schema.TAnonComplexInfBPe2) string {
	if inf == nil {
		return ""
	}
	return inf.VersaoAttr
}

func versionFromBPeTM(inf *schema.TAnonComplexInfBPe1) string {
	if inf == nil {
		return ""
	}
	return inf.VersaoAttr
}

func onlyBPeRoot(d *Document) bool {
	return d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyBPeTMRoot(d *Document) bool {
	return d.BPe == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyBPeProcRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyBPeTMProcRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyRetBPeRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyConsSitRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyRetConsSitRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyConsStatRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyRetConsStatRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyEventoRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.RetEventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyRetEventoRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.ProcEventoBPe == nil
}

func onlyProcEventoRoot(d *Document) bool {
	return d.BPe == nil && d.BPeTM == nil && d.BPeProc == nil && d.BPeTMProc == nil && d.RetBPe == nil && d.ConsSitBPe == nil && d.RetConsSitBPe == nil && d.ConsStatServBPe == nil && d.RetConsStatServBPe == nil && d.EventoBPe == nil && d.RetEventoBPe == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
