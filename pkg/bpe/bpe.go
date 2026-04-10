package bpe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	schema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/core"
	alteracaoPoltronaEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_alteracao_poltrona"
	cancelEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_cancel"
	excessoBagagemEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_excesso_bagagem"
	naoEmbEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_nao_emb"
	"github.com/awafinance/fiscal/internal/xmlutil"
)

const namespace = "http://www.portalfiscal.inf.br/bpe"

var errSingleRoot = errors.New("marshal bpe: document must contain exactly one supported root")

var parsersByRoot = map[string]func([]byte, string) (*Document, error){
	"BPe":                parseBPe,
	"BPeTM":              parseBPeTM,
	"bpeProc":            parseBPeProc,
	"bpeTMProc":          parseBPeTMProc,
	"retBPe":             parseRetBPe,
	"consSitBPe":         parseConsSitBPe,
	"retConsSitBPe":      parseRetConsSitBPe,
	"consStatServBPe":    parseConsStatServBPe,
	"retConsStatServBPe": parseRetConsStatServBPe,
	"eventoBPe":          func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseEventDocument) },
	"retEventoBPe":       func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseRetEventDocument) },
	"procEventoBPe":      func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseProcEventDocument) },
}

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
	RootName                    string                                    `json:"rootName,omitempty"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.RootName {
	case "BPe", "":
		return marshalBPe(e, d)
	case "BPeTM":
		return marshalBPeTM(e, d)
	case "bpeProc":
		return marshalBPeProc(e, d)
	case "bpeTMProc":
		return marshalBPeTMProc(e, d)
	case "retBPe":
		return marshalRetBPe(e, d)
	case "consSitBPe":
		return marshalConsSitBPe(e, d)
	case "retConsSitBPe":
		return marshalRetConsSitBPe(e, d)
	case "consStatServBPe":
		return marshalConsStatServBPe(e, d)
	case "retConsStatServBPe":
		return marshalRetConsStatServBPe(e, d)
	case "eventoBPe":
		return marshalEventRoot(e, d)
	case "retEventoBPe":
		return marshalRetEventRoot(e, d)
	case "procEventoBPe":
		return marshalProcEventRoot(e, d)
	}

	return errSingleRoot
}

func marshalBPe(e *xml.Encoder, d *Document) error {
	if d.BPe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                        `xml:"BPe"`
		XMLNS       string                          `xml:"xmlns,attr,omitempty"`
		InfBPe      *schema.TAnonComplexInfBPe2     `xml:"infBPe"`
		InfBPeSupl  *schema.TAnonComplexInfBPeSupl1 `xml:"infBPeSupl,omitempty"`
		DsSignature *schema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "BPe"},
		XMLNS:       namespace,
		InfBPe:      d.BPe.InfBPe,
		InfBPeSupl:  d.BPe.InfBPeSupl,
		DsSignature: d.BPe.DsSignature,
	})
}

func marshalBPeTM(e *xml.Encoder, d *Document) error {
	if d.BPeTM == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                    `xml:"BPeTM"`
		XMLNS       string                      `xml:"xmlns,attr,omitempty"`
		InfBPe      *schema.TAnonComplexInfBPe1 `xml:"infBPe"`
		DsSignature *schema.SignatureType       `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "BPeTM"},
		XMLNS:       namespace,
		InfBPe:      d.BPeTM.InfBPe,
		DsSignature: d.BPeTM.DsSignature,
	})
}

func marshalBPeProc(e *xml.Encoder, d *Document) error {
	if d.BPeProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
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
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "bpeProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.BPeProc.VersaoAttr),
		IpTransmissorAttr: d.BPeProc.IpTransmissorAttr,
		NPortaConAttr:     d.BPeProc.NPortaConAttr,
		DhConexaoAttr:     d.BPeProc.DhConexaoAttr,
		BPe:               d.BPeProc.BPe,
		ProtBPe:           d.BPeProc.ProtBPe,
	})
}

func marshalBPeTMProc(e *xml.Encoder, d *Document) error {
	if d.BPeTMProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
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
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "bpeTMProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.BPeTMProc.VersaoAttr),
		IpTransmissorAttr: d.BPeTMProc.IpTransmissorAttr,
		NPortaConAttr:     d.BPeTMProc.NPortaConAttr,
		DhConexaoAttr:     d.BPeTMProc.DhConexaoAttr,
		BPeTM:             d.BPeTMProc.BPeTM,
		ProtBPe:           d.BPeTMProc.ProtBPe,
	})
}

func marshalRetBPe(e *xml.Encoder, d *Document) error {
	if d.RetBPe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeCanonical(e,
		struct {
			XMLName xml.Name `xml:"retBPe"`
			XMLNS   string   `xml:"xmlns,attr,omitempty"`
			*schema.TRetBPe
		}{
			XMLName: xml.Name{Local: "retBPe"},
			XMLNS:   namespace,
			TRetBPe: d.RetBPe,
		})
}

func marshalConsSitBPe(e *xml.Encoder, d *Document) error {
	if d.ConsSitBPe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"consSitBPe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TConsSitBPe
	}{
		XMLName:     xml.Name{Local: "consSitBPe"},
		XMLNS:       namespace,
		TConsSitBPe: d.ConsSitBPe,
	})
}

func marshalRetConsSitBPe(e *xml.Encoder, d *Document) error {
	if d.RetConsSitBPe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retConsSitBPe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TRetConsSitBPe
	}{
		XMLName:        xml.Name{Local: "retConsSitBPe"},
		XMLNS:          namespace,
		TRetConsSitBPe: d.RetConsSitBPe,
	})
}

func marshalConsStatServBPe(e *xml.Encoder, d *Document) error {
	if d.ConsStatServBPe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"consStatServBPe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TConsStatServ
	}{
		XMLName:       xml.Name{Local: "consStatServBPe"},
		XMLNS:         namespace,
		TConsStatServ: d.ConsStatServBPe,
	})
}

func marshalRetConsStatServBPe(e *xml.Encoder, d *Document) error {
	if d.RetConsStatServBPe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeCanonical(e, struct {
		XMLName xml.Name `xml:"retConsStatServBPe"`
		XMLNS   string   `xml:"xmlns,attr,omitempty"`
		*schema.TRetConsStatServ
	}{
		XMLName:          xml.Name{Local: "retConsStatServBPe"},
		XMLNS:            namespace,
		TRetConsStatServ: d.RetConsStatServBPe,
	})
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse bpe: empty xml document")
	}

	RootName, rootErr := xmlutil.ParseRootName(data)
	if rootErr != nil && RootName == "" {
		return nil, fmt.Errorf("parse bpe: read root: %w", rootErr)
	}

	if fn, ok := parsersByRoot[RootName]; ok {
		return fn(data, RootName)
	}
	if rootErr != nil {
		return nil, fmt.Errorf("parse bpe: read root: %w", rootErr)
	}
	return nil, fmt.Errorf("parse bpe: unsupported root element %q", RootName)
}

func ParseReader(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parse bpe: read xml: %w", err)
	}
	return Parse(data)
}

func finalizeDoc(doc *Document) (*Document, error) {
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func parseBPe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TBPe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode BPe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: versionFromBPe(parsed.InfBPe), BPe: &parsed, RootName: rootName})
}

func parseBPeTM(data []byte, rootName string) (*Document, error) {
	var parsed schema.TBPeTM
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode BPeTM: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: versionFromBPeTM(parsed.InfBPe), BPeTM: &parsed, RootName: rootName})
}

func parseBPeProc(data []byte, rootName string) (*Document, error) {
	var parsed schema.TAnonComplexBpeProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode bpeProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, BPeProc: &parsed, RootName: rootName})
}

func parseBPeTMProc(data []byte, rootName string) (*Document, error) {
	var parsed schema.TAnonComplexBpeTMProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode bpeTMProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, BPeTMProc: &parsed, RootName: rootName})
}

func parseRetBPe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TRetBPe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode retBPe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetBPe: &parsed, RootName: rootName})
}

func parseConsSitBPe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TConsSitBPe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode consSitBPe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsSitBPe: &parsed, RootName: rootName})
}

func parseRetConsSitBPe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TRetConsSitBPe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode retConsSitBPe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsSitBPe: &parsed, RootName: rootName})
}

func parseConsStatServBPe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode consStatServBPe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsStatServBPe: &parsed, RootName: rootName})
}

func parseRetConsStatServBPe(data []byte, rootName string) (*Document, error) {
	var parsed schema.TRetConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode retConsStatServBPe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServBPe: &parsed, RootName: rootName})
}

func parseEventRoot(data []byte, rootName string, fn func([]byte, string, string) (*Document, error)) (*Document, error) {
	tpEvento, err := eventTypeFromXML(data)
	if err != nil {
		return nil, fmt.Errorf("parse bpe: decode %s head: %w", rootName, err)
	}
	if tpEvento == "" {
		return nil, errors.New("parse bpe: missing infEvento")
	}
	return fn(data, rootName, tpEvento)
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

var rootValidators = []func(*Document) error{
	validateBPeRoot,
	validateBPeTMRoot,
	validateBPeProcRoot,
	validateBPeTMProcRoot,
	validateRetBPeRoot,
	validateConsSitBPeRoot,
	validateRetConsSitBPeRoot,
	validateConsStatServBPeRoot,
	validateRetConsStatServBPeRoot,
	validateEventRoots,
}

func validateDocument(doc *Document) error {
	if activeRootCount(doc) != 1 {
		return errors.New("parse bpe: document must contain exactly one supported root")
	}
	for _, v := range rootValidators {
		if err := v(doc); err != nil {
			return err
		}
	}
	return nil
}

func validateBPeRoot(doc *Document) error {
	if doc.BPe == nil {
		return nil
	}
	return validateInfBPe(doc.BPe.InfBPe)
}

func validateBPeTMRoot(doc *Document) error {
	if doc.BPeTM == nil {
		return nil
	}
	return validateInfBPeTM(doc.BPeTM.InfBPe)
}

func validateBPeProcRoot(doc *Document) error {
	if doc.BPeProc == nil {
		return nil
	}
	if doc.BPeProc.BPe == nil {
		return errors.New("parse bpe: missing BPe")
	}
	if doc.BPeProc.ProtBPe == nil {
		return errors.New("parse bpe: missing protBPe")
	}
	return nil
}

func validateBPeTMProcRoot(doc *Document) error {
	if doc.BPeTMProc == nil {
		return nil
	}
	if doc.BPeTMProc.BPeTM == nil {
		return errors.New("parse bpe: missing BPeTM")
	}
	if doc.BPeTMProc.ProtBPe == nil {
		return errors.New("parse bpe: missing protBPe")
	}
	return nil
}

func validateRetBPeRoot(doc *Document) error {
	if doc.RetBPe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetBPe.TpAmb),
		missing("cUF", doc.RetBPe.CUF),
		missing("cStat", doc.RetBPe.CStat),
	)
}

func validateConsSitBPeRoot(doc *Document) error {
	if doc.ConsSitBPe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.ConsSitBPe.TpAmb),
		missing("chBPe", doc.ConsSitBPe.ChBPe),
	)
}

func validateRetConsSitBPeRoot(doc *Document) error {
	if doc.RetConsSitBPe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsSitBPe.TpAmb),
		missing("cUF", doc.RetConsSitBPe.CUF),
		missing("cStat", doc.RetConsSitBPe.CStat),
	)
}

func validateConsStatServBPeRoot(doc *Document) error {
	if doc.ConsStatServBPe == nil {
		return nil
	}
	return firstMissing(missing("tpAmb", doc.ConsStatServBPe.TpAmb))
}

func validateRetConsStatServBPeRoot(doc *Document) error {
	if doc.RetConsStatServBPe == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetConsStatServBPe.TpAmb),
		missing("cUF", doc.RetConsStatServBPe.CUF),
		missing("cStat", doc.RetConsStatServBPe.CStat),
		missing("dhRecbto", doc.RetConsStatServBPe.DhRecbto),
	)
}

func missing(field, value string) error {
	if value == "" {
		return errors.New("parse bpe: missing " + field)
	}
	return nil
}

func firstMissing(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
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
	if err := validateEventos(doc); err != nil {
		return err
	}
	if err := validateRetEventos(doc); err != nil {
		return err
	}
	return validateProcEventos(doc)
}

func validateEventos(doc *Document) error {
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
	}
	return nil
}

func validateRetEventos(doc *Document) error {
	switch {
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
	}
	return nil
}

func validateProcEventos(doc *Document) error {
	switch {
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
	}
	return nil
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
		return encodeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoBPe.VersaoAttr), d.EventoBPe.InfEvento, d.EventoBPe.DsSignature)
	case d.EventoCancBPe != nil:
		return encodeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoCancBPe.VersaoAttr), d.EventoCancBPe.InfEvento, d.EventoCancBPe.DsSignature)
	case d.EventoAlteracaoPoltrona != nil:
		return encodeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoAlteracaoPoltrona.VersaoAttr), d.EventoAlteracaoPoltrona.InfEvento, d.EventoAlteracaoPoltrona.DsSignature)
	case d.EventoExcessoBagagem != nil:
		return encodeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoExcessoBagagem.VersaoAttr), d.EventoExcessoBagagem.InfEvento, d.EventoExcessoBagagem.DsSignature)
	case d.EventoNaoEmbBPe != nil:
		return encodeEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.EventoNaoEmbBPe.VersaoAttr), d.EventoNaoEmbBPe.InfEvento, d.EventoNaoEmbBPe.DsSignature)
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
		return encodeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoBPe.VersaoAttr), d.RetEventoBPe.InfEvento, d.RetEventoBPe.DsSignature)
	case d.RetEventoCancBPe != nil:
		return encodeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoCancBPe.VersaoAttr), d.RetEventoCancBPe.InfEvento, d.RetEventoCancBPe.DsSignature)
	case d.RetEventoAlteracaoPoltrona != nil:
		return encodeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoAlteracaoPoltrona.VersaoAttr), d.RetEventoAlteracaoPoltrona.InfEvento, d.RetEventoAlteracaoPoltrona.DsSignature)
	case d.RetEventoExcessoBagagem != nil:
		return encodeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoExcessoBagagem.VersaoAttr), d.RetEventoExcessoBagagem.InfEvento, d.RetEventoExcessoBagagem.DsSignature)
	case d.RetEventoNaoEmbBPe != nil:
		return encodeRetEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetEventoNaoEmbBPe.VersaoAttr), d.RetEventoNaoEmbBPe.InfEvento, d.RetEventoNaoEmbBPe.DsSignature)
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
		return encodeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoBPe.VersaoAttr), d.ProcEventoBPe.IpTransmissorAttr, d.ProcEventoBPe.NPortaConAttr, d.ProcEventoBPe.DhConexaoAttr, d.ProcEventoBPe.EventoBPe, d.ProcEventoBPe.RetEventoBPe)
	case d.ProcEventoCancBPe != nil:
		return encodeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoCancBPe.VersaoAttr), d.ProcEventoCancBPe.IpTransmissorAttr, d.ProcEventoCancBPe.NPortaConAttr, d.ProcEventoCancBPe.DhConexaoAttr, d.ProcEventoCancBPe.EventoBPe, d.ProcEventoCancBPe.RetEventoBPe)
	case d.ProcEventoAlteracaoPoltrona != nil:
		return encodeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoAlteracaoPoltrona.VersaoAttr), d.ProcEventoAlteracaoPoltrona.IpTransmissorAttr, d.ProcEventoAlteracaoPoltrona.NPortaConAttr, d.ProcEventoAlteracaoPoltrona.DhConexaoAttr, d.ProcEventoAlteracaoPoltrona.EventoBPe, d.ProcEventoAlteracaoPoltrona.RetEventoBPe)
	case d.ProcEventoExcessoBagagem != nil:
		return encodeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoExcessoBagagem.VersaoAttr), d.ProcEventoExcessoBagagem.IpTransmissorAttr, d.ProcEventoExcessoBagagem.NPortaConAttr, d.ProcEventoExcessoBagagem.DhConexaoAttr, d.ProcEventoExcessoBagagem.EventoBPe, d.ProcEventoExcessoBagagem.RetEventoBPe)
	case d.ProcEventoNaoEmbBPe != nil:
		return encodeProcEvent(e, xmlutil.FirstNonEmpty(d.VersaoAttr, d.ProcEventoNaoEmbBPe.VersaoAttr), d.ProcEventoNaoEmbBPe.IpTransmissorAttr, d.ProcEventoNaoEmbBPe.NPortaConAttr, d.ProcEventoNaoEmbBPe.DhConexaoAttr, d.ProcEventoNaoEmbBPe.EventoBPe, d.ProcEventoNaoEmbBPe.RetEventoBPe)
	default:
		return errors.New("marshal bpe: document must contain exactly one supported root")
	}
}

func encodeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return xmlutil.EncodeCanonical(e, struct {
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
	return xmlutil.EncodeCanonical(e, struct {
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
	return xmlutil.EncodeCanonical(e, struct {
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

func decodeEvent[T any](data []byte, context string, assign func(*T) *Document) (*Document, error) {
	var parsed T
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse bpe: decode %s: %w", context, err)
	}
	return finalizeDoc(assign(&parsed))
}

func parseEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "eventoBPe cancelamento", func(p *cancelEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancBPe: p, RootName: rootName}
		})
	case "110115":
		return decodeEvent(data, "eventoBPe nao embarque", func(p *naoEmbEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoNaoEmbBPe: p, RootName: rootName}
		})
	case "110116":
		return decodeEvent(data, "eventoBPe alteracao poltrona", func(p *alteracaoPoltronaEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoAlteracaoPoltrona: p, RootName: rootName}
		})
	case "110117":
		return decodeEvent(data, "eventoBPe excesso bagagem", func(p *excessoBagagemEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoExcessoBagagem: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "eventoBPe", func(p *schema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoBPe: p, RootName: rootName}
		})
	}
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "retEventoBPe cancelamento", func(p *cancelEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCancBPe: p, RootName: rootName}
		})
	case "110115":
		return decodeEvent(data, "retEventoBPe nao embarque", func(p *naoEmbEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoNaoEmbBPe: p, RootName: rootName}
		})
	case "110116":
		return decodeEvent(data, "retEventoBPe alteracao poltrona", func(p *alteracaoPoltronaEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoAlteracaoPoltrona: p, RootName: rootName}
		})
	case "110117":
		return decodeEvent(data, "retEventoBPe excesso bagagem", func(p *excessoBagagemEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoExcessoBagagem: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "retEventoBPe", func(p *schema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoBPe: p, RootName: rootName}
		})
	}
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110111":
		return decodeEvent(data, "procEventoBPe cancelamento", func(p *cancelEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCancBPe: p, RootName: rootName}
		})
	case "110115":
		return decodeEvent(data, "procEventoBPe nao embarque", func(p *naoEmbEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoNaoEmbBPe: p, RootName: rootName}
		})
	case "110116":
		return decodeEvent(data, "procEventoBPe alteracao poltrona", func(p *alteracaoPoltronaEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoAlteracaoPoltrona: p, RootName: rootName}
		})
	case "110117":
		return decodeEvent(data, "procEventoBPe excesso bagagem", func(p *excessoBagagemEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoExcessoBagagem: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "procEventoBPe", func(p *schema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoBPe: p, RootName: rootName}
		})
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
	for _, ok := range []bool{
		doc.BPe != nil,
		doc.BPeTM != nil,
		doc.BPeProc != nil,
		doc.BPeTMProc != nil,
		doc.RetBPe != nil,
		doc.ConsSitBPe != nil,
		doc.RetConsSitBPe != nil,
		doc.ConsStatServBPe != nil,
		doc.RetConsStatServBPe != nil,
		doc.EventoBPe != nil,
		doc.RetEventoBPe != nil,
		doc.ProcEventoBPe != nil,
		doc.EventoCancBPe != nil,
		doc.RetEventoCancBPe != nil,
		doc.ProcEventoCancBPe != nil,
		doc.EventoAlteracaoPoltrona != nil,
		doc.RetEventoAlteracaoPoltrona != nil,
		doc.ProcEventoAlteracaoPoltrona != nil,
		doc.EventoExcessoBagagem != nil,
		doc.RetEventoExcessoBagagem != nil,
		doc.ProcEventoExcessoBagagem != nil,
		doc.EventoNaoEmbBPe != nil,
		doc.RetEventoNaoEmbBPe != nil,
		doc.ProcEventoNaoEmbBPe != nil,
	} {
		if ok {
			count++
		}
	}
	return count
}
