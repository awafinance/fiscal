package bpe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	schema "github.com/awa/fiscal/internal/bpe/gen/v1_0/core"
	alteracaoPoltronaEventSchema "github.com/awa/fiscal/internal/bpe/gen/v1_0/evento_alteracao_poltrona"
	cancelEventSchema "github.com/awa/fiscal/internal/bpe/gen/v1_0/evento_cancel"
	excessoBagagemEventSchema "github.com/awa/fiscal/internal/bpe/gen/v1_0/evento_excesso_bagagem"
	naoEmbEventSchema "github.com/awa/fiscal/internal/bpe/gen/v1_0/evento_nao_emb"
)

const namespace = "http://www.portalfiscal.inf.br/bpe"

type Document struct {
	VersaoAttr                  string                                    `json:"versao,omitempty"`
	BPe                         *schema.TBPe                              `json:"BPe,omitempty"`
	BPeTM                       *schema.TBPeTM                            `json:"BPeTM,omitempty"`
	BPeProc                     *schema.TAnonComplexBpeProc1              `json:"bpeProc,omitempty"`
	BPeTMProc                   *schema.TAnonComplexBpeTMProc1            `json:"bpeTMProc,omitempty"`
	RetBPe                      *schema.TRetBPe                           `json:"retBPe,omitempty"`
	ConsSitBPe                  *schema.TConsSitBPe                       `json:"consSitBPe,omitempty"`
	RetConsSitBPe               *schema.TRetConsSitBPe                    `json:"retConsSitBPe,omitempty"`
	ConsStatServBPe             *schema.TConsStatServ                     `json:"consStatServBPe,omitempty"`
	RetConsStatServBPe          *schema.TRetConsStatServ                  `json:"retConsStatServBPe,omitempty"`
	EventoBPe                   *schema.TEvento                           `json:"eventoBPe,omitempty"`
	RetEventoBPe                *schema.TRetEvento                        `json:"retEventoBPe,omitempty"`
	ProcEventoBPe               *schema.TProcEvento                       `json:"procEventoBPe,omitempty"`
	EventoCancBPe               *cancelEventSchema.TEvento                `json:"eventoCancBPe,omitempty"`
	RetEventoCancBPe            *cancelEventSchema.TRetEvento             `json:"retEventoCancBPe,omitempty"`
	ProcEventoCancBPe           *cancelEventSchema.TProcEvento            `json:"procEventoCancBPe,omitempty"`
	EventoAlteracaoPoltrona     *alteracaoPoltronaEventSchema.TEvento     `json:"eventoAlteracaoPoltrona,omitempty"`
	RetEventoAlteracaoPoltrona  *alteracaoPoltronaEventSchema.TRetEvento  `json:"retEventoAlteracaoPoltrona,omitempty"`
	ProcEventoAlteracaoPoltrona *alteracaoPoltronaEventSchema.TProcEvento `json:"procEventoAlteracaoPoltrona,omitempty"`
	EventoExcessoBagagem        *excessoBagagemEventSchema.TEvento        `json:"eventoExcessoBagagem,omitempty"`
	RetEventoExcessoBagagem     *excessoBagagemEventSchema.TRetEvento     `json:"retEventoExcessoBagagem,omitempty"`
	ProcEventoExcessoBagagem    *excessoBagagemEventSchema.TProcEvento    `json:"procEventoExcessoBagagem,omitempty"`
	EventoNaoEmbBPe             *naoEmbEventSchema.TEvento                `json:"eventoNaoEmbBPe,omitempty"`
	RetEventoNaoEmbBPe          *naoEmbEventSchema.TRetEvento             `json:"retEventoNaoEmbBPe,omitempty"`
	ProcEventoNaoEmbBPe         *naoEmbEventSchema.TProcEvento            `json:"procEventoNaoEmbBPe,omitempty"`
	rootName                    string                                    `json:"-"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "BPe", "":
		if d.BPe != nil && activeRootCount(d) == 1 {
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
		if d.BPeTM != nil && activeRootCount(d) == 1 {
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
		if d.BPeProc != nil && activeRootCount(d) == 1 {
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
		if d.BPeTMProc != nil && activeRootCount(d) == 1 {
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
		if d.RetBPe != nil && activeRootCount(d) == 1 {
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
		if d.ConsSitBPe != nil && activeRootCount(d) == 1 {
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
		if d.RetConsSitBPe != nil && activeRootCount(d) == 1 {
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
		if d.ConsStatServBPe != nil && activeRootCount(d) == 1 {
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
		if d.RetConsStatServBPe != nil && activeRootCount(d) == 1 {
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
		return marshalEventRoot(e, d)
	case "retEventoBPe":
		return marshalRetEventRoot(e, d)
	case "procEventoBPe":
		return marshalProcEventRoot(e, d)
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
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse bpe: missing infEvento")
		}
		return parseEventDocument(data, rootName, tpEvento)
	case "retEventoBPe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse bpe: missing infEvento")
		}
		return parseRetEventDocument(data, rootName, tpEvento)
	case "procEventoBPe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse bpe: missing infEvento")
		}
		return parseProcEventDocument(data, rootName, tpEvento)
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse bpe: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse bpe: unsupported root element %q", rootName)
	}
}

func eventTypeFromXML(data []byte) (string, error) {
	var head struct {
		InfEvento struct {
			TpEvento string `xml:"tpEvento"`
		} `xml:"infEvento"`
		EventoBPe struct {
			InfEvento struct {
				TpEvento string `xml:"tpEvento"`
			} `xml:"infEvento"`
		} `xml:"eventoBPe"`
	}
	if err := xml.Unmarshal(data, &head); err != nil {
		return "", err
	}
	if head.InfEvento.TpEvento != "" {
		return head.InfEvento.TpEvento, nil
	}
	return head.EventoBPe.InfEvento.TpEvento, nil
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
		return errors.New("parse bpe: document must contain exactly one supported root")
	}

	if doc.BPe != nil {
		if err := validateInfBPe(doc.BPe.InfBPe); err != nil {
			return err
		}
	}
	if doc.BPeTM != nil {
		if err := validateInfBPeTM(doc.BPeTM.InfBPe); err != nil {
			return err
		}
	}
	if doc.BPeProc != nil {
		if doc.BPeProc.BPe == nil {
			return errors.New("parse bpe: missing BPe")
		}
		if doc.BPeProc.ProtBPe == nil {
			return errors.New("parse bpe: missing protBPe")
		}
	}
	if doc.BPeTMProc != nil {
		if doc.BPeTMProc.BPeTM == nil {
			return errors.New("parse bpe: missing BPeTM")
		}
		if doc.BPeTMProc.ProtBPe == nil {
			return errors.New("parse bpe: missing protBPe")
		}
	}
	if doc.RetBPe != nil {
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
		if doc.ConsSitBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
		if doc.ConsSitBPe.ChBPe == "" {
			return errors.New("parse bpe: missing chBPe")
		}
	}
	if doc.RetConsSitBPe != nil {
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
		if doc.ConsStatServBPe.TpAmb == "" {
			return errors.New("parse bpe: missing tpAmb")
		}
	}
	if doc.RetConsStatServBPe != nil {
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
	if err := validateEventRoots(doc); err != nil {
		return err
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

func validateEventRoots(doc *Document) error {
	switch {
	case doc.EventoBPe != nil:
		return validateEvento(doc.EventoBPe.InfEvento)
	case doc.EventoCancBPe != nil:
		return validateEvento(doc.EventoCancBPe.InfEvento)
	case doc.EventoAlteracaoPoltrona != nil:
		return validateEvento(doc.EventoAlteracaoPoltrona.InfEvento)
	case doc.EventoExcessoBagagem != nil:
		return validateEvento(doc.EventoExcessoBagagem.InfEvento)
	case doc.EventoNaoEmbBPe != nil:
		return validateEvento(doc.EventoNaoEmbBPe.InfEvento)
	case doc.RetEventoBPe != nil:
		return validateRetEvento(doc.RetEventoBPe.InfEvento)
	case doc.RetEventoCancBPe != nil:
		return validateRetEvento(doc.RetEventoCancBPe.InfEvento)
	case doc.RetEventoAlteracaoPoltrona != nil:
		return validateRetEvento(doc.RetEventoAlteracaoPoltrona.InfEvento)
	case doc.RetEventoExcessoBagagem != nil:
		return validateRetEvento(doc.RetEventoExcessoBagagem.InfEvento)
	case doc.RetEventoNaoEmbBPe != nil:
		return validateRetEvento(doc.RetEventoNaoEmbBPe.InfEvento)
	case doc.ProcEventoBPe != nil:
		return validateProcEvento(doc.ProcEventoBPe.EventoBPe, doc.ProcEventoBPe.RetEventoBPe)
	case doc.ProcEventoCancBPe != nil:
		return validateProcEvento(doc.ProcEventoCancBPe.EventoBPe, doc.ProcEventoCancBPe.RetEventoBPe)
	case doc.ProcEventoAlteracaoPoltrona != nil:
		return validateProcEvento(doc.ProcEventoAlteracaoPoltrona.EventoBPe, doc.ProcEventoAlteracaoPoltrona.RetEventoBPe)
	case doc.ProcEventoExcessoBagagem != nil:
		return validateProcEvento(doc.ProcEventoExcessoBagagem.EventoBPe, doc.ProcEventoExcessoBagagem.RetEventoBPe)
	case doc.ProcEventoNaoEmbBPe != nil:
		return validateProcEvento(doc.ProcEventoNaoEmbBPe.EventoBPe, doc.ProcEventoNaoEmbBPe.RetEventoBPe)
	default:
		return nil
	}
}

func validateEvento(inf any) error {
	switch inf := inf.(type) {
	case *schema.TAnonComplexInfEvento1:
		return validateEventoFields(inf == nil, inf.ChBPe, inf.DetEvento == nil)
	case *cancelEventSchema.TAnonComplexInfEvento1:
		return validateEventoFields(inf == nil, inf.ChBPe, inf.DetEvento == nil)
	case *alteracaoPoltronaEventSchema.TAnonComplexInfEvento1:
		return validateEventoFields(inf == nil, inf.ChBPe, inf.DetEvento == nil)
	case *excessoBagagemEventSchema.TAnonComplexInfEvento1:
		return validateEventoFields(inf == nil, inf.ChBPe, inf.DetEvento == nil)
	case *naoEmbEventSchema.TAnonComplexInfEvento1:
		return validateEventoFields(inf == nil, inf.ChBPe, inf.DetEvento == nil)
	default:
		return errors.New("parse bpe: missing infEvento")
	}
}

func validateEventoFields(isNil bool, chBPe string, missingDetEvento bool) error {
	if isNil {
		return errors.New("parse bpe: missing infEvento")
	}
	if chBPe == "" {
		return errors.New("parse bpe: missing chBPe")
	}
	if missingDetEvento {
		return errors.New("parse bpe: missing detEvento")
	}
	return nil
}

func validateRetEvento(inf any) error {
	switch inf := inf.(type) {
	case *schema.TAnonComplexInfEvento2:
		return validateRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *cancelEventSchema.TAnonComplexInfEvento2:
		return validateRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *alteracaoPoltronaEventSchema.TAnonComplexInfEvento2:
		return validateRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *excessoBagagemEventSchema.TAnonComplexInfEvento2:
		return validateRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	case *naoEmbEventSchema.TAnonComplexInfEvento2:
		return validateRetEventoFields(inf == nil, inf.TpAmb, inf.CStat)
	default:
		return errors.New("parse bpe: missing infEvento")
	}
}

func validateRetEventoFields(isNil bool, tpAmb, cStat string) error {
	if isNil {
		return errors.New("parse bpe: missing infEvento")
	}
	if tpAmb == "" {
		return errors.New("parse bpe: missing tpAmb")
	}
	if cStat == "" {
		return errors.New("parse bpe: missing cStat")
	}
	return nil
}

func validateProcEvento(evento any, retEvento any) error {
	if evento == nil {
		return errors.New("parse bpe: missing eventoBPe")
	}
	if retEvento == nil {
		return errors.New("parse bpe: missing retEventoBPe")
	}
	return nil
}

func marshalEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}

	switch {
	case d.EventoBPe != nil:
		return encodeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoBPe.VersaoAttr), d.EventoBPe.InfEvento, d.EventoBPe.DsSignature)
	case d.EventoCancBPe != nil:
		return encodeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoCancBPe.VersaoAttr), d.EventoCancBPe.InfEvento, d.EventoCancBPe.DsSignature)
	case d.EventoAlteracaoPoltrona != nil:
		return encodeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoAlteracaoPoltrona.VersaoAttr), d.EventoAlteracaoPoltrona.InfEvento, d.EventoAlteracaoPoltrona.DsSignature)
	case d.EventoExcessoBagagem != nil:
		return encodeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoExcessoBagagem.VersaoAttr), d.EventoExcessoBagagem.InfEvento, d.EventoExcessoBagagem.DsSignature)
	case d.EventoNaoEmbBPe != nil:
		return encodeEvent(e, firstNonEmpty(d.VersaoAttr, d.EventoNaoEmbBPe.VersaoAttr), d.EventoNaoEmbBPe.InfEvento, d.EventoNaoEmbBPe.DsSignature)
	default:
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}

	switch {
	case d.RetEventoBPe != nil:
		return encodeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoBPe.VersaoAttr), d.RetEventoBPe.InfEvento, d.RetEventoBPe.DsSignature)
	case d.RetEventoCancBPe != nil:
		return encodeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoCancBPe.VersaoAttr), d.RetEventoCancBPe.InfEvento, d.RetEventoCancBPe.DsSignature)
	case d.RetEventoAlteracaoPoltrona != nil:
		return encodeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoAlteracaoPoltrona.VersaoAttr), d.RetEventoAlteracaoPoltrona.InfEvento, d.RetEventoAlteracaoPoltrona.DsSignature)
	case d.RetEventoExcessoBagagem != nil:
		return encodeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoExcessoBagagem.VersaoAttr), d.RetEventoExcessoBagagem.InfEvento, d.RetEventoExcessoBagagem.DsSignature)
	case d.RetEventoNaoEmbBPe != nil:
		return encodeRetEvent(e, firstNonEmpty(d.VersaoAttr, d.RetEventoNaoEmbBPe.VersaoAttr), d.RetEventoNaoEmbBPe.InfEvento, d.RetEventoNaoEmbBPe.DsSignature)
	default:
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}

	switch {
	case d.ProcEventoBPe != nil:
		return encodeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoBPe.VersaoAttr), d.ProcEventoBPe.IpTransmissorAttr, d.ProcEventoBPe.NPortaConAttr, d.ProcEventoBPe.DhConexaoAttr, d.ProcEventoBPe.EventoBPe, d.ProcEventoBPe.RetEventoBPe)
	case d.ProcEventoCancBPe != nil:
		return encodeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoCancBPe.VersaoAttr), d.ProcEventoCancBPe.IpTransmissorAttr, d.ProcEventoCancBPe.NPortaConAttr, d.ProcEventoCancBPe.DhConexaoAttr, d.ProcEventoCancBPe.EventoBPe, d.ProcEventoCancBPe.RetEventoBPe)
	case d.ProcEventoAlteracaoPoltrona != nil:
		return encodeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoAlteracaoPoltrona.VersaoAttr), d.ProcEventoAlteracaoPoltrona.IpTransmissorAttr, d.ProcEventoAlteracaoPoltrona.NPortaConAttr, d.ProcEventoAlteracaoPoltrona.DhConexaoAttr, d.ProcEventoAlteracaoPoltrona.EventoBPe, d.ProcEventoAlteracaoPoltrona.RetEventoBPe)
	case d.ProcEventoExcessoBagagem != nil:
		return encodeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoExcessoBagagem.VersaoAttr), d.ProcEventoExcessoBagagem.IpTransmissorAttr, d.ProcEventoExcessoBagagem.NPortaConAttr, d.ProcEventoExcessoBagagem.DhConexaoAttr, d.ProcEventoExcessoBagagem.EventoBPe, d.ProcEventoExcessoBagagem.RetEventoBPe)
	case d.ProcEventoNaoEmbBPe != nil:
		return encodeProcEvent(e, firstNonEmpty(d.VersaoAttr, d.ProcEventoNaoEmbBPe.VersaoAttr), d.ProcEventoNaoEmbBPe.IpTransmissorAttr, d.ProcEventoNaoEmbBPe.NPortaConAttr, d.ProcEventoNaoEmbBPe.DhConexaoAttr, d.ProcEventoNaoEmbBPe.EventoBPe, d.ProcEventoNaoEmbBPe.RetEventoBPe)
	default:
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}
}

func encodeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return e.Encode(struct {
		XMLName     xml.Name `xml:"eventoBPe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}{
		XMLName:     xml.Name{Local: "eventoBPe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeRetEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return e.Encode(struct {
		XMLName     xml.Name `xml:"retEventoBPe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}{
		XMLName:     xml.Name{Local: "retEventoBPe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeProcEvent(e *xml.Encoder, versao string, ipTransmissor, nPortaCon, dhConexao *string, evento any, retEvento any) error {
	return e.Encode(struct {
		XMLName           xml.Name `xml:"procEventoBPe"`
		XMLNS             string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string   `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string  `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string  `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string  `xml:"dhConexao,attr,omitempty"`
		EventoBPe         any      `xml:"eventoBPe"`
		RetEventoBPe      any      `xml:"retEventoBPe"`
	}{
		XMLName:           xml.Name{Local: "procEventoBPe"},
		XMLNS:             namespace,
		VersaoAttr:        versao,
		IpTransmissorAttr: ipTransmissor,
		NPortaConAttr:     nPortaCon,
		DhConexaoAttr:     dhConexao,
		EventoBPe:         evento,
		RetEventoBPe:      retEvento,
	})
}

func parseEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		var parsed cancelEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe cancelamento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoCancBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110115":
		var parsed naoEmbEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe nao embarque: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoNaoEmbBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110116":
		var parsed alteracaoPoltronaEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe alteracao poltrona: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoAlteracaoPoltrona: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110117":
		var parsed excessoBagagemEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe excesso bagagem: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoExcessoBagagem: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		var parsed schema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode eventoBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EventoBPe: &parsed, rootName: rootName}
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
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe cancelamento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoCancBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110115":
		var parsed naoEmbEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe nao embarque: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoNaoEmbBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110116":
		var parsed alteracaoPoltronaEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe alteracao poltrona: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoAlteracaoPoltrona: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110117":
		var parsed excessoBagagemEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe excesso bagagem: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoExcessoBagagem: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		var parsed schema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode retEventoBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEventoBPe: &parsed, rootName: rootName}
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
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe cancelamento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCancBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110115":
		var parsed naoEmbEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe nao embarque: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoNaoEmbBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110116":
		var parsed alteracaoPoltronaEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe alteracao poltrona: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoAlteracaoPoltrona: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "110117":
		var parsed excessoBagagemEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe excesso bagagem: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoExcessoBagagem: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		var parsed schema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse bpe: decode procEventoBPe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoBPe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	}
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

func activeRootCount(doc *Document) int {
	count := 0
	if doc.BPe != nil {
		count++
	}
	if doc.BPeTM != nil {
		count++
	}
	if doc.BPeProc != nil {
		count++
	}
	if doc.BPeTMProc != nil {
		count++
	}
	if doc.RetBPe != nil {
		count++
	}
	if doc.ConsSitBPe != nil {
		count++
	}
	if doc.RetConsSitBPe != nil {
		count++
	}
	if doc.ConsStatServBPe != nil {
		count++
	}
	if doc.RetConsStatServBPe != nil {
		count++
	}
	if doc.EventoBPe != nil {
		count++
	}
	if doc.RetEventoBPe != nil {
		count++
	}
	if doc.ProcEventoBPe != nil {
		count++
	}
	if doc.EventoCancBPe != nil {
		count++
	}
	if doc.RetEventoCancBPe != nil {
		count++
	}
	if doc.ProcEventoCancBPe != nil {
		count++
	}
	if doc.EventoAlteracaoPoltrona != nil {
		count++
	}
	if doc.RetEventoAlteracaoPoltrona != nil {
		count++
	}
	if doc.ProcEventoAlteracaoPoltrona != nil {
		count++
	}
	if doc.EventoExcessoBagagem != nil {
		count++
	}
	if doc.RetEventoExcessoBagagem != nil {
		count++
	}
	if doc.ProcEventoExcessoBagagem != nil {
		count++
	}
	if doc.EventoNaoEmbBPe != nil {
		count++
	}
	if doc.RetEventoNaoEmbBPe != nil {
		count++
	}
	if doc.ProcEventoNaoEmbBPe != nil {
		count++
	}
	return count
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
