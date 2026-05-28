package cte

import (
	"encoding/xml"
	"errors"
	"fmt"

	cancelEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel"
	cancelCEEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel_ce"
	cancelIEEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel_ie"
	cancelPrestDesacordoEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel_prest_desacordo"
	eventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cce"
	ceEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_ce"
	epecEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_epec"
	genericEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_generico"
	gtvEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_gtv"
	ieEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_ie"
	prestDesacordoEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_prest_desacordo"
	regMultimodalEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_reg_multimodal"
	"github.com/awafinance/fiscal/internal/xmlutil"
)

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
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}
	switch {
	case d.EventoCTe != nil:
		return encodeCTeEvent(e, d.EventoCTe.VersaoAttr, d.EventoCTe.InfEvento, d.EventoCTe.DsSignature)
	case d.EventoCancCTe != nil:
		return encodeCTeEvent(e, d.EventoCancCTe.VersaoAttr, d.EventoCancCTe.InfEvento, d.EventoCancCTe.DsSignature)
	case d.EventoCECTe != nil:
		return encodeCTeEvent(e, d.EventoCECTe.VersaoAttr, d.EventoCECTe.InfEvento, d.EventoCECTe.DsSignature)
	case d.EventoCancCECTe != nil:
		return encodeCTeEvent(e, d.EventoCancCECTe.VersaoAttr, d.EventoCancCECTe.InfEvento, d.EventoCancCECTe.DsSignature)
	case d.EventoEPECCTe != nil:
		return encodeCTeEvent(e, d.EventoEPECCTe.VersaoAttr, d.EventoEPECCTe.InfEvento, d.EventoEPECCTe.DsSignature)
	case d.EventoRegMultimodal != nil:
		return encodeCTeEvent(e, d.EventoRegMultimodal.VersaoAttr, d.EventoRegMultimodal.InfEvento, d.EventoRegMultimodal.DsSignature)
	case d.EventoGTV != nil:
		return encodeCTeEvent(e, d.EventoGTV.VersaoAttr, d.EventoGTV.InfEvento, d.EventoGTV.DsSignature)
	case d.EventoIECTe != nil:
		return encodeCTeEvent(e, d.EventoIECTe.VersaoAttr, d.EventoIECTe.InfEvento, d.EventoIECTe.DsSignature)
	case d.EventoCancIECTe != nil:
		return encodeCTeEvent(e, d.EventoCancIECTe.VersaoAttr, d.EventoCancIECTe.InfEvento, d.EventoCancIECTe.DsSignature)
	case d.EventoPrestDesacordo != nil:
		return encodeCTeEvent(e, d.EventoPrestDesacordo.VersaoAttr, d.EventoPrestDesacordo.InfEvento, d.EventoPrestDesacordo.DsSignature)
	case d.EventoCancPrestDesacordo != nil:
		return encodeCTeEvent(e, d.EventoCancPrestDesacordo.VersaoAttr, d.EventoCancPrestDesacordo.InfEvento, d.EventoCancPrestDesacordo.DsSignature)
	case d.EventoGenerico != nil:
		return encodeCTeEvent(e, d.EventoGenerico.VersaoAttr, d.EventoGenerico.InfEvento, d.EventoGenerico.DsSignature)
	default:
		return errSingleRoot
	}
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}
	switch {
	case d.RetEventoCTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoCTe.VersaoAttr, d.RetEventoCTe.InfEvento, d.RetEventoCTe.DsSignature)
	case d.RetEventoCancCTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoCancCTe.VersaoAttr, d.RetEventoCancCTe.InfEvento, d.RetEventoCancCTe.DsSignature)
	case d.RetEventoCECTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoCECTe.VersaoAttr, d.RetEventoCECTe.InfEvento, d.RetEventoCECTe.DsSignature)
	case d.RetEventoCancCECTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoCancCECTe.VersaoAttr, d.RetEventoCancCECTe.InfEvento, d.RetEventoCancCECTe.DsSignature)
	case d.RetEventoEPECCTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoEPECCTe.VersaoAttr, d.RetEventoEPECCTe.InfEvento, d.RetEventoEPECCTe.DsSignature)
	case d.RetEventoRegMultimodal != nil:
		return encodeCTeRetEvent(e, d.RetEventoRegMultimodal.VersaoAttr, d.RetEventoRegMultimodal.InfEvento, d.RetEventoRegMultimodal.DsSignature)
	case d.RetEventoGTV != nil:
		return encodeCTeRetEvent(e, d.RetEventoGTV.VersaoAttr, d.RetEventoGTV.InfEvento, d.RetEventoGTV.DsSignature)
	case d.RetEventoIECTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoIECTe.VersaoAttr, d.RetEventoIECTe.InfEvento, d.RetEventoIECTe.DsSignature)
	case d.RetEventoCancIECTe != nil:
		return encodeCTeRetEvent(e, d.RetEventoCancIECTe.VersaoAttr, d.RetEventoCancIECTe.InfEvento, d.RetEventoCancIECTe.DsSignature)
	case d.RetEventoPrestDesacordo != nil:
		return encodeCTeRetEvent(e, d.RetEventoPrestDesacordo.VersaoAttr, d.RetEventoPrestDesacordo.InfEvento, d.RetEventoPrestDesacordo.DsSignature)
	case d.RetEventoCancPrestDesacordo != nil:
		return encodeCTeRetEvent(e, d.RetEventoCancPrestDesacordo.VersaoAttr, d.RetEventoCancPrestDesacordo.InfEvento, d.RetEventoCancPrestDesacordo.DsSignature)
	case d.RetEventoGenerico != nil:
		return encodeCTeRetEvent(e, d.RetEventoGenerico.VersaoAttr, d.RetEventoGenerico.InfEvento, d.RetEventoGenerico.DsSignature)
	default:
		return errSingleRoot
	}
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errSingleRoot
	}
	switch {
	case d.ProcEventoCTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoCTe.VersaoAttr, d.ProcEventoCTe.IpTransmissorAttr, d.ProcEventoCTe.NPortaConAttr, d.ProcEventoCTe.DhConexaoAttr, d.ProcEventoCTe.EventoCTe, d.ProcEventoCTe.RetEventoCTe)
	case d.ProcEventoCancCTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoCancCTe.VersaoAttr, d.ProcEventoCancCTe.IpTransmissorAttr, d.ProcEventoCancCTe.NPortaConAttr, d.ProcEventoCancCTe.DhConexaoAttr, d.ProcEventoCancCTe.EventoCTe, d.ProcEventoCancCTe.RetEventoCTe)
	case d.ProcEventoCECTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoCECTe.VersaoAttr, d.ProcEventoCECTe.IpTransmissorAttr, d.ProcEventoCECTe.NPortaConAttr, d.ProcEventoCECTe.DhConexaoAttr, d.ProcEventoCECTe.EventoCTe, d.ProcEventoCECTe.RetEventoCTe)
	case d.ProcEventoCancCECTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoCancCECTe.VersaoAttr, d.ProcEventoCancCECTe.IpTransmissorAttr, d.ProcEventoCancCECTe.NPortaConAttr, d.ProcEventoCancCECTe.DhConexaoAttr, d.ProcEventoCancCECTe.EventoCTe, d.ProcEventoCancCECTe.RetEventoCTe)
	case d.ProcEventoEPECCTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoEPECCTe.VersaoAttr, d.ProcEventoEPECCTe.IpTransmissorAttr, d.ProcEventoEPECCTe.NPortaConAttr, d.ProcEventoEPECCTe.DhConexaoAttr, d.ProcEventoEPECCTe.EventoCTe, d.ProcEventoEPECCTe.RetEventoCTe)
	case d.ProcEventoRegMultimodal != nil:
		return encodeCTeProcEvent(e, d.ProcEventoRegMultimodal.VersaoAttr, d.ProcEventoRegMultimodal.IpTransmissorAttr, d.ProcEventoRegMultimodal.NPortaConAttr, d.ProcEventoRegMultimodal.DhConexaoAttr, d.ProcEventoRegMultimodal.EventoCTe, d.ProcEventoRegMultimodal.RetEventoCTe)
	case d.ProcEventoGTV != nil:
		return encodeCTeProcEvent(e, d.ProcEventoGTV.VersaoAttr, d.ProcEventoGTV.IpTransmissorAttr, d.ProcEventoGTV.NPortaConAttr, d.ProcEventoGTV.DhConexaoAttr, d.ProcEventoGTV.EventoCTe, d.ProcEventoGTV.RetEventoCTe)
	case d.ProcEventoIECTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoIECTe.VersaoAttr, d.ProcEventoIECTe.IpTransmissorAttr, d.ProcEventoIECTe.NPortaConAttr, d.ProcEventoIECTe.DhConexaoAttr, d.ProcEventoIECTe.EventoCTe, d.ProcEventoIECTe.RetEventoCTe)
	case d.ProcEventoCancIECTe != nil:
		return encodeCTeProcEvent(e, d.ProcEventoCancIECTe.VersaoAttr, d.ProcEventoCancIECTe.IpTransmissorAttr, d.ProcEventoCancIECTe.NPortaConAttr, d.ProcEventoCancIECTe.DhConexaoAttr, d.ProcEventoCancIECTe.EventoCTe, d.ProcEventoCancIECTe.RetEventoCTe)
	case d.ProcEventoPrestDesacordo != nil:
		return encodeCTeProcEvent(e, d.ProcEventoPrestDesacordo.VersaoAttr, d.ProcEventoPrestDesacordo.IpTransmissorAttr, d.ProcEventoPrestDesacordo.NPortaConAttr, d.ProcEventoPrestDesacordo.DhConexaoAttr, d.ProcEventoPrestDesacordo.EventoCTe, d.ProcEventoPrestDesacordo.RetEventoCTe)
	case d.ProcEventoCancPrestDesacordo != nil:
		return encodeCTeProcEvent(e, d.ProcEventoCancPrestDesacordo.VersaoAttr, d.ProcEventoCancPrestDesacordo.IpTransmissorAttr, d.ProcEventoCancPrestDesacordo.NPortaConAttr, d.ProcEventoCancPrestDesacordo.DhConexaoAttr, d.ProcEventoCancPrestDesacordo.EventoCTe, d.ProcEventoCancPrestDesacordo.RetEventoCTe)
	case d.ProcEventoGenerico != nil:
		return encodeCTeProcEvent(e, d.ProcEventoGenerico.VersaoAttr, d.ProcEventoGenerico.IpTransmissorAttr, d.ProcEventoGenerico.NPortaConAttr, d.ProcEventoGenerico.DhConexaoAttr, d.ProcEventoGenerico.EventoCTe, d.ProcEventoGenerico.RetEventoCTe)
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

func decodeEvent[T any](data []byte, context string, assign func(*T) *Document) (*Document, error) {
	var parsed T
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode %s: %w", context, err)
	}
	return finalizeDoc(assign(&parsed))
}

func parseEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110110":
		return decodeEvent(data, "eventoCTe cce", func(p *eventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCTe: p, RootName: rootName}
		})
	case "110111":
		return decodeEvent(data, "eventoCTe cancel", func(p *cancelEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancCTe: p, RootName: rootName}
		})
	case "110113":
		return decodeEvent(data, "eventoCTe epec", func(p *epecEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoEPECCTe: p, RootName: rootName}
		})
	case "110160":
		return decodeEvent(data, "eventoCTe reg multimodal", func(p *regMultimodalEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoRegMultimodal: p, RootName: rootName}
		})
	case "110170":
		return decodeEvent(data, "eventoCTe gtv", func(p *gtvEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoGTV: p, RootName: rootName}
		})
	case "110180":
		return decodeEvent(data, "eventoCTe ce", func(p *ceEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCECTe: p, RootName: rootName}
		})
	case "110181":
		return decodeEvent(data, "eventoCTe cancel ce", func(p *cancelCEEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancCECTe: p, RootName: rootName}
		})
	case "110190":
		return decodeEvent(data, "eventoCTe ie", func(p *ieEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoIECTe: p, RootName: rootName}
		})
	case "110191":
		return decodeEvent(data, "eventoCTe cancel ie", func(p *cancelIEEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancIECTe: p, RootName: rootName}
		})
	case "610110":
		return decodeEvent(data, "eventoCTe prest desacordo", func(p *prestDesacordoEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoPrestDesacordo: p, RootName: rootName}
		})
	case "610111":
		return decodeEvent(data, "eventoCTe cancel prest desacordo", func(p *cancelPrestDesacordoEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoCancPrestDesacordo: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "eventoCTe generic", func(p *genericEventSchema.TEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, EventoGenerico: p, RootName: rootName}
		})
	}
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110110":
		return decodeEvent(data, "retEventoCTe cce", func(p *eventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCTe: p, RootName: rootName}
		})
	case "110111":
		return decodeEvent(data, "retEventoCTe cancel", func(p *cancelEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCancCTe: p, RootName: rootName}
		})
	case "110113":
		return decodeEvent(data, "retEventoCTe epec", func(p *epecEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoEPECCTe: p, RootName: rootName}
		})
	case "110160":
		return decodeEvent(data, "retEventoCTe reg multimodal", func(p *regMultimodalEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoRegMultimodal: p, RootName: rootName}
		})
	case "110170":
		return decodeEvent(data, "retEventoCTe gtv", func(p *gtvEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoGTV: p, RootName: rootName}
		})
	case "110180":
		return decodeEvent(data, "retEventoCTe ce", func(p *ceEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCECTe: p, RootName: rootName}
		})
	case "110181":
		return decodeEvent(data, "retEventoCTe cancel ce", func(p *cancelCEEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCancCECTe: p, RootName: rootName}
		})
	case "110190":
		return decodeEvent(data, "retEventoCTe ie", func(p *ieEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoIECTe: p, RootName: rootName}
		})
	case "110191":
		return decodeEvent(data, "retEventoCTe cancel ie", func(p *cancelIEEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCancIECTe: p, RootName: rootName}
		})
	case "610110":
		return decodeEvent(data, "retEventoCTe prest desacordo", func(p *prestDesacordoEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoPrestDesacordo: p, RootName: rootName}
		})
	case "610111":
		return decodeEvent(data, "retEventoCTe cancel prest desacordo", func(p *cancelPrestDesacordoEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoCancPrestDesacordo: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "retEventoCTe generic", func(p *genericEventSchema.TRetEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, RetEventoGenerico: p, RootName: rootName}
		})
	}
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110110":
		return decodeEvent(data, "procEventoCTe cce", func(p *eventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCTe: p, RootName: rootName}
		})
	case "110111":
		return decodeEvent(data, "procEventoCTe cancel", func(p *cancelEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCancCTe: p, RootName: rootName}
		})
	case "110113":
		return decodeEvent(data, "procEventoCTe epec", func(p *epecEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoEPECCTe: p, RootName: rootName}
		})
	case "110160":
		return decodeEvent(data, "procEventoCTe reg multimodal", func(p *regMultimodalEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoRegMultimodal: p, RootName: rootName}
		})
	case "110170":
		return decodeEvent(data, "procEventoCTe gtv", func(p *gtvEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoGTV: p, RootName: rootName}
		})
	case "110180":
		return decodeEvent(data, "procEventoCTe ce", func(p *ceEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCECTe: p, RootName: rootName}
		})
	case "110181":
		return decodeEvent(data, "procEventoCTe cancel ce", func(p *cancelCEEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCancCECTe: p, RootName: rootName}
		})
	case "110190":
		return decodeEvent(data, "procEventoCTe ie", func(p *ieEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoIECTe: p, RootName: rootName}
		})
	case "110191":
		return decodeEvent(data, "procEventoCTe cancel ie", func(p *cancelIEEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCancIECTe: p, RootName: rootName}
		})
	case "610110":
		return decodeEvent(data, "procEventoCTe prest desacordo", func(p *prestDesacordoEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoPrestDesacordo: p, RootName: rootName}
		})
	case "610111":
		return decodeEvent(data, "procEventoCTe cancel prest desacordo", func(p *cancelPrestDesacordoEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoCancPrestDesacordo: p, RootName: rootName}
		})
	default:
		return decodeEvent(data, "procEventoCTe generic", func(p *genericEventSchema.TProcEvento) *Document {
			return &Document{VersaoAttr: p.VersaoAttr, ProcEventoGenerico: p, RootName: rootName}
		})
	}
}
