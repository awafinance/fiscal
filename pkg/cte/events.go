package cte

import (
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"

	"github.com/awafinance/fiscal/internal/xmlutil"
)

type cteEventRootKind uint8

const (
	cteSentEventRoot cteEventRootKind = iota
	cteRetEventRoot
	cteProcEventRoot
)

type cteEventSpec struct {
	eventType string
	context   string
	generic   bool

	eventField string

	eventTypeOf     reflect.Type
	retEventTypeOf  reflect.Type
	procEventTypeOf reflect.Type

	captureDetailXML bool
}

var cteEventSpecs = []cteEventSpec{
	cteEvent[EventoCTeEvento, EventoCTeRetEvento, EventoCTeProcEvento]("110110", "cce", "EventoCTe"),
	cteEvent[EventoCancCTeEvento, EventoCancCTeRetEvento, EventoCancCTeProcEvento]("110111", "cancel", "EventoCancCTe"),
	cteEvent[EventoEPECCTeEvento, EventoEPECCTeRetEvento, EventoEPECCTeProcEvento]("110113", "epec", "EventoEPECCTe"),
	cteEvent[EventoRegMultimodalCTeEvento, EventoRegMultimodalCTeRetEvento, EventoRegMultimodalCTeProcEvento]("110160", "reg multimodal", "EventoRegMultimodal"),
	cteEvent[EventoGTVCTeEvento, EventoGTVCTeRetEvento, EventoGTVCTeProcEvento]("110170", "gtv", "EventoGTV"),
	cteEvent[EventoCECTeEvento, EventoCECTeRetEvento, EventoCECTeProcEvento]("110180", "ce", "EventoCECTe"),
	cteEvent[EventoCancCECTeEvento, EventoCancCECTeRetEvento, EventoCancCECTeProcEvento]("110181", "cancel ce", "EventoCancCECTe"),
	cteEvent[EventoIECTeEvento, EventoIECTeRetEvento, EventoIECTeProcEvento]("110190", "ie", "EventoIECTe"),
	cteEvent[EventoCancIECTeEvento, EventoCancIECTeRetEvento, EventoCancIECTeProcEvento]("110191", "cancel ie", "EventoCancIECTe"),
	cteEvent[EventoPrestDesacordoCTeEvento, EventoPrestDesacordoCTeRetEvento, EventoPrestDesacordoCTeProcEvento]("610110", "prest desacordo", "EventoPrestDesacordo"),
	cteEvent[EventoCancPrestDesacordoCTeEvento, EventoCancPrestDesacordoCTeRetEvento, EventoCancPrestDesacordoCTeProcEvento]("610111", "cancel prest desacordo", "EventoCancPrestDesacordo"),
	cteGenericEvent[EventoGenericoCTeEvento, EventoGenericoCTeRetEvento, EventoGenericoCTeProcEvento]("generic", "EventoGenerico"),
}

func cteEvent[E, R, P any](eventType, context, eventField string) cteEventSpec {
	return cteEventSpec{
		eventType:       eventType,
		context:         context,
		eventField:      eventField,
		eventTypeOf:     reflect.TypeFor[E](),
		retEventTypeOf:  reflect.TypeFor[R](),
		procEventTypeOf: reflect.TypeFor[P](),
	}
}

func cteGenericEvent[E, R, P any](context, eventField string) cteEventSpec {
	spec := cteEvent[E, R, P]("", context, eventField)
	spec.generic = true
	spec.captureDetailXML = true
	return spec
}

func parseEventRoot(data []byte, rootName string, fn func([]byte, string, string) (*Document, error)) (*Document, error) {
	tpEvento, err := eventTypeFromXML(data)
	if err != nil {
		return nil, fmt.Errorf("parse cte: decode %s head: %w", rootName, err)
	}
	if tpEvento == "" {
		if rootName == "retEventoCTe" {
			return fn(data, rootName, tpEvento)
		}
		return nil, errors.New("parse cte: missing infEvento")
	}
	return fn(data, rootName, tpEvento)
}

func eventTypeFromXML(data []byte) (string, error) {
	var head struct {
		InfEvento struct {
			TpEvento string `xml:"tpEvento"`
		} `xml:"infEvento"`
		EventoCTe struct {
			InfEvento struct {
				TpEvento string `xml:"tpEvento"`
			} `xml:"infEvento"`
		} `xml:"eventoCTe"`
	}
	if err := xml.Unmarshal(data, &head); err != nil {
		return "", err
	}
	if head.InfEvento.TpEvento != "" {
		return head.InfEvento.TpEvento, nil
	}
	return head.EventoCTe.InfEvento.TpEvento, nil
}

func marshalEventRoot(e *xml.Encoder, d *Document) error {
	return marshalCTeEventRoot(e, d, cteSentEventRoot)
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	return marshalCTeEventRoot(e, d, cteRetEventRoot)
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	return marshalCTeEventRoot(e, d, cteProcEventRoot)
}

func marshalCTeEventRoot(e *xml.Encoder, d *Document, kind cteEventRootKind) error {
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}
	_, root, ok := cteEventSpecForDocument(d, kind)
	if !ok {
		return errSingleRoot
	}
	switch kind {
	case cteSentEventRoot:
		return encodeCTeEvent(e, cteStringField(root, "VersaoAttr"), cteAnyField(root, "InfEvento"), cteAnyField(root, "DsSignature"))
	case cteRetEventRoot:
		return encodeCTeRetEvent(e, cteStringField(root, "VersaoAttr"), cteAnyField(root, "InfEvento"), cteAnyField(root, "DsSignature"))
	case cteProcEventRoot:
		return encodeCTeProcEvent(
			e,
			cteStringField(root, "VersaoAttr"),
			cteStringPtrField(root, "IpTransmissorAttr"),
			cteStringPtrField(root, "NPortaConAttr"),
			cteStringPtrField(root, "DhConexaoAttr"),
			cteAnyField(root, "EventoCTe"),
			cteAnyField(root, "RetEventoCTe"),
		)
	default:
		return errSingleRoot
	}
}

func encodeCTeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName     xml.Name `xml:"eventoCTe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}{
		XMLName:     xml.Name{Local: "eventoCTe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeCTeRetEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName     xml.Name `xml:"retEventoCTe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}{
		XMLName:     xml.Name{Local: "retEventoCTe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeCTeProcEvent(e *xml.Encoder, versao string, ipTransmissor, nPortaCon, dhConexao *string, evento any, retEvento any) error {
	return xmlutil.EncodeCanonical(e, struct {
		XMLName           xml.Name `xml:"procEventoCTe"`
		XMLNS             string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string   `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string  `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string  `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string  `xml:"dhConexao,attr,omitempty"`
		EventoCTe         any      `xml:"eventoCTe"`
		RetEventoCTe      any      `xml:"retEventoCTe"`
	}{
		XMLName:           xml.Name{Local: "procEventoCTe"},
		XMLNS:             namespace,
		VersaoAttr:        versao,
		IpTransmissorAttr: ipTransmissor,
		NPortaConAttr:     nPortaCon,
		DhConexaoAttr:     dhConexao,
		EventoCTe:         evento,
		RetEventoCTe:      retEvento,
	})
}

func parseEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	return parseCTeEventDocument(data, rootName, tpEvento, cteSentEventRoot)
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	return parseCTeEventDocument(data, rootName, tpEvento, cteRetEventRoot)
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	return parseCTeEventDocument(data, rootName, tpEvento, cteProcEventRoot)
}

func parseCTeEventDocument(data []byte, rootName, tpEvento string, kind cteEventRootKind) (*Document, error) {
	spec := cteEventSpecForType(tpEvento)
	if spec == nil {
		return nil, fmt.Errorf("parse cte: unsupported eventoCTe type %q", tpEvento)
	}

	parsed := reflect.New(spec.rootType(kind))
	if err := xml.Unmarshal(data, parsed.Interface()); err != nil {
		return nil, fmt.Errorf("parse cte: decode %s %s: %w", kind.rootName(), spec.context, err)
	}

	doc := &Document{
		VersaoAttr: cteStringField(parsed, "VersaoAttr"),
		RootName:   rootName,
	}
	cteDocumentEventField(doc, spec, kind).Set(parsed)
	return finalizeDoc(doc)
}

func cteEventSpecForType(tpEvento string) *cteEventSpec {
	var generic *cteEventSpec
	for i := range cteEventSpecs {
		spec := &cteEventSpecs[i]
		if spec.generic {
			generic = spec
			continue
		}
		if spec.eventType == tpEvento {
			return spec
		}
	}
	return generic
}

func cteEventSpecForDocument(d *Document, kind cteEventRootKind) (*cteEventSpec, reflect.Value, bool) {
	for i := range cteEventSpecs {
		spec := &cteEventSpecs[i]
		root := cteDocumentEventField(d, spec, kind)
		if root.IsValid() && !root.IsNil() {
			return spec, root, true
		}
	}
	return nil, reflect.Value{}, false
}

func cteEventSpecForRoot(root any, kind cteEventRootKind) (*cteEventSpec, reflect.Value, bool) {
	value := reflect.ValueOf(root)
	if !value.IsValid() {
		return nil, reflect.Value{}, false
	}
	for i := range cteEventSpecs {
		spec := &cteEventSpecs[i]
		if value.Type() == reflect.PointerTo(spec.rootType(kind)) {
			return spec, value, true
		}
	}
	return nil, reflect.Value{}, false
}

func cteDocumentEventField(d *Document, spec *cteEventSpec, kind cteEventRootKind) reflect.Value {
	if d == nil || spec == nil {
		return reflect.Value{}
	}
	return reflect.ValueOf(d).Elem().FieldByName(spec.docField(kind))
}

func activeCTeEventRootCount(d *Document) int {
	count := 0
	for i := range cteEventSpecs {
		spec := &cteEventSpecs[i]
		for _, kind := range []cteEventRootKind{cteSentEventRoot, cteRetEventRoot, cteProcEventRoot} {
			root := cteDocumentEventField(d, spec, kind)
			if root.IsValid() && !root.IsNil() {
				count++
			}
		}
	}
	return count
}

func (s *cteEventSpec) rootType(kind cteEventRootKind) reflect.Type {
	switch kind {
	case cteSentEventRoot:
		return s.eventTypeOf
	case cteRetEventRoot:
		return s.retEventTypeOf
	case cteProcEventRoot:
		return s.procEventTypeOf
	default:
		return nil
	}
}

func (s *cteEventSpec) docField(kind cteEventRootKind) string {
	switch kind {
	case cteSentEventRoot:
		return s.eventField
	case cteRetEventRoot:
		return "Ret" + s.eventField
	case cteProcEventRoot:
		return "Proc" + s.eventField
	default:
		return ""
	}
}

func (kind cteEventRootKind) rootName() string {
	switch kind {
	case cteSentEventRoot:
		return "eventoCTe"
	case cteRetEventRoot:
		return "retEventoCTe"
	case cteProcEventRoot:
		return "procEventoCTe"
	default:
		return "eventoCTe"
	}
}

func cteField(value reflect.Value, name string) reflect.Value {
	if !value.IsValid() {
		return reflect.Value{}
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return value.FieldByName(name)
}

func cteAnyField(value reflect.Value, name string) any {
	field := cteField(value, name)
	if !field.IsValid() {
		return nil
	}
	if field.Kind() == reflect.Pointer && field.IsNil() {
		return nil
	}
	return field.Interface()
}

func cteStringField(value reflect.Value, name string) string {
	return cteStringValue(cteField(value, name))
}

func cteStringPtrField(value reflect.Value, name string) *string {
	field := cteField(value, name)
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return nil
	}
	if ptr, ok := field.Interface().(*string); ok {
		return ptr
	}
	if field.Type().Elem().Kind() == reflect.String {
		value := field.Elem().String()
		return &value
	}
	return nil
}

func cteStringValue(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.String {
		return ""
	}
	return value.String()
}

func cteHasValue(value reflect.Value) bool {
	return value.IsValid() && (value.Kind() != reflect.Pointer || !value.IsNil())
}
