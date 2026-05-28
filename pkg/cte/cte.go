package cte

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	distSchema "github.com/awafinance/fiscal/internal/cte/gen/v1_0/dist_dfe"
	consSitSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/consulta_situacao"
	cteSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte"
	cteOSSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte_os"
	cteSimpSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte_simp"
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
	gtveSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/gtve"
	statusSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/status_servico"
	"github.com/awafinance/fiscal/internal/xmlutil"
	"github.com/awafinance/fiscal/pkg/fiscalerr"
)

const namespace = "http://www.portalfiscal.inf.br/cte"

var errSingleRoot = errors.New("marshal cte: document must contain exactly one supported root")

var parsersByRoot = map[string]func([]byte, string) (*Document, error){
	"CTe":                parseCTe,
	"cteProc":            parseCTeProc,
	"retCTe":             parseRetCTe,
	"CTeOS":              parseCTeOS,
	"cteOSProc":          parseCTeOSProc,
	"retCTeOS":           parseRetCTeOS,
	"CTeSimp":            parseCTeSimp,
	"cteSimpProc":        parseCTeSimpProc,
	"retCTeSimp":         parseRetCTeSimp,
	"GTVe":               parseGTVe,
	"GTVeProc":           parseGTVeProc,
	"retGTVe":            parseRetGTVe,
	"consSitCTe":         parseConsSitCTe,
	"retConsSitCTe":      parseRetConsSitCTe,
	"consStatServCTe":    parseConsStatServCTe,
	"retConsStatServCTe": parseRetConsStatServCTe,
	"eventoCTe":          func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseEventDocument) },
	"retEventoCTe":       func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseRetEventDocument) },
	"procEventoCTe":      func(d []byte, rn string) (*Document, error) { return parseEventRoot(d, rn, parseProcEventDocument) },
	"distDFeInt":         parseDistDFeInt,
	"retDistDFeInt":      parseRetDistDFeInt,
}

type Document struct {
	VersaoAttr                   string                                       `json:"versao,omitempty"`
	CTe                          *cteSchema.TCTe                              `json:"CTe,omitempty"`
	CTeProc                      *cteSchema.TAnonComplexCteProc1              `json:"cteProc,omitempty"`
	RetCTe                       *cteSchema.TRetCTe                           `json:"retCTe,omitempty"`
	CTeOS                        *cteOSSchema.TCTeOS                          `json:"CTeOS,omitempty"`
	CTeOSProc                    *cteOSSchema.TAnonComplexCteOSProc1          `json:"cteOSProc,omitempty"`
	RetCTeOS                     *cteOSSchema.TRetCTeOS                       `json:"retCTeOS,omitempty"`
	CTeSimp                      *cteSimpSchema.TCTeSimp                      `json:"CTeSimp,omitempty"`
	CTeSimpProc                  *cteSimpSchema.TAnonComplexCteSimpProc1      `json:"cteSimpProc,omitempty"`
	RetCTeSimp                   *cteSimpSchema.TRetCTeSimp                   `json:"retCTeSimp,omitempty"`
	GTVe                         *gtveSchema.TGTVe                            `json:"GTVe,omitempty"`
	GTVeProc                     *gtveSchema.TAnonComplexGTVeProc1            `json:"GTVeProc,omitempty"`
	RetGTVe                      *gtveSchema.TRetGTVe                         `json:"retGTVe,omitempty"`
	ConsSitCTe                   *consSitSchema.TConsSitCTe                   `json:"consSitCTe,omitempty"`
	RetConsSitCTe                *consSitSchema.TRetConsSitCTe                `json:"retConsSitCTe,omitempty"`
	ConsStatServCTe              *statusSchema.TConsStatServ                  `json:"consStatServCTe,omitempty"`
	RetConsStatServCTe           *statusSchema.TRetConsStatServ               `json:"retConsStatServCTe,omitempty"`
	EventoCTe                    *eventSchema.TEvento                         `json:"eventoCTe,omitempty"`
	RetEventoCTe                 *eventSchema.TRetEvento                      `json:"retEventoCTe,omitempty"`
	ProcEventoCTe                *eventSchema.TProcEvento                     `json:"procEventoCTe,omitempty"`
	EventoCancCTe                *cancelEventSchema.TEvento                   `json:"eventoCancCTe,omitempty"`
	RetEventoCancCTe             *cancelEventSchema.TRetEvento                `json:"retEventoCancCTe,omitempty"`
	ProcEventoCancCTe            *cancelEventSchema.TProcEvento               `json:"procEventoCancCTe,omitempty"`
	EventoCECTe                  *ceEventSchema.TEvento                       `json:"eventoCECTe,omitempty"`
	RetEventoCECTe               *ceEventSchema.TRetEvento                    `json:"retEventoCECTe,omitempty"`
	ProcEventoCECTe              *ceEventSchema.TProcEvento                   `json:"procEventoCECTe,omitempty"`
	EventoCancCECTe              *cancelCEEventSchema.TEvento                 `json:"eventoCancCECTe,omitempty"`
	RetEventoCancCECTe           *cancelCEEventSchema.TRetEvento              `json:"retEventoCancCECTe,omitempty"`
	ProcEventoCancCECTe          *cancelCEEventSchema.TProcEvento             `json:"procEventoCancCECTe,omitempty"`
	EventoEPECCTe                *epecEventSchema.TEvento                     `json:"eventoEPECCTe,omitempty"`
	RetEventoEPECCTe             *epecEventSchema.TRetEvento                  `json:"retEventoEPECCTe,omitempty"`
	ProcEventoEPECCTe            *epecEventSchema.TProcEvento                 `json:"procEventoEPECCTe,omitempty"`
	EventoRegMultimodal          *regMultimodalEventSchema.TEvento            `json:"eventoRegMultimodal,omitempty"`
	RetEventoRegMultimodal       *regMultimodalEventSchema.TRetEvento         `json:"retEventoRegMultimodal,omitempty"`
	ProcEventoRegMultimodal      *regMultimodalEventSchema.TProcEvento        `json:"procEventoRegMultimodal,omitempty"`
	EventoGTV                    *gtvEventSchema.TEvento                      `json:"eventoGTV,omitempty"`
	RetEventoGTV                 *gtvEventSchema.TRetEvento                   `json:"retEventoGTV,omitempty"`
	ProcEventoGTV                *gtvEventSchema.TProcEvento                  `json:"procEventoGTV,omitempty"`
	EventoIECTe                  *ieEventSchema.TEvento                       `json:"eventoIECTe,omitempty"`
	RetEventoIECTe               *ieEventSchema.TRetEvento                    `json:"retEventoIECTe,omitempty"`
	ProcEventoIECTe              *ieEventSchema.TProcEvento                   `json:"procEventoIECTe,omitempty"`
	EventoCancIECTe              *cancelIEEventSchema.TEvento                 `json:"eventoCancIECTe,omitempty"`
	RetEventoCancIECTe           *cancelIEEventSchema.TRetEvento              `json:"retEventoCancIECTe,omitempty"`
	ProcEventoCancIECTe          *cancelIEEventSchema.TProcEvento             `json:"procEventoCancIECTe,omitempty"`
	EventoPrestDesacordo         *prestDesacordoEventSchema.TEvento           `json:"eventoPrestDesacordo,omitempty"`
	RetEventoPrestDesacordo      *prestDesacordoEventSchema.TRetEvento        `json:"retEventoPrestDesacordo,omitempty"`
	ProcEventoPrestDesacordo     *prestDesacordoEventSchema.TProcEvento       `json:"procEventoPrestDesacordo,omitempty"`
	EventoCancPrestDesacordo     *cancelPrestDesacordoEventSchema.TEvento     `json:"eventoCancPrestDesacordo,omitempty"`
	RetEventoCancPrestDesacordo  *cancelPrestDesacordoEventSchema.TRetEvento  `json:"retEventoCancPrestDesacordo,omitempty"`
	ProcEventoCancPrestDesacordo *cancelPrestDesacordoEventSchema.TProcEvento `json:"procEventoCancPrestDesacordo,omitempty"`
	EventoGenerico               *genericEventSchema.TEvento                  `json:"eventoGenerico,omitempty"`
	RetEventoGenerico            *genericEventSchema.TRetEvento               `json:"retEventoGenerico,omitempty"`
	ProcEventoGenerico           *genericEventSchema.TProcEvento              `json:"procEventoGenerico,omitempty"`
	DistDFeInt                   *distSchema.TAnonComplexDistDFeInt1          `json:"distDFeInt,omitempty"`
	RetDistDFeInt                *distSchema.TAnonComplexRetDistDFeInt1       `json:"retDistDFeInt,omitempty"`
	RootName                     string                                       `json:"rootName,omitempty"`
}

var marshalersByRoot = map[string]func(*xml.Encoder, *Document) error{
	"CTe":                marshalCTe,
	"":                   marshalCTe,
	"cteProc":            marshalCTeProc,
	"retCTe":             marshalRetCTe,
	"CTeOS":              marshalCTeOS,
	"cteOSProc":          marshalCTeOSProc,
	"retCTeOS":           marshalRetCTeOS,
	"CTeSimp":            marshalCTeSimp,
	"cteSimpProc":        marshalCTeSimpProc,
	"retCTeSimp":         marshalRetCTeSimp,
	"GTVe":               marshalGTVe,
	"GTVeProc":           marshalGTVeProc,
	"retGTVe":            marshalRetGTVe,
	"consSitCTe":         marshalConsSitCTe,
	"retConsSitCTe":      marshalRetConsSitCTe,
	"consStatServCTe":    marshalConsStatServCTe,
	"retConsStatServCTe": marshalRetConsStatServCTe,
	"eventoCTe":          marshalEventRoot,
	"retEventoCTe":       marshalRetEventRoot,
	"procEventoCTe":      marshalProcEventRoot,
	"distDFeInt":         marshalDistDFeInt,
	"retDistDFeInt":      marshalRetDistDFeInt,
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	if fn, ok := marshalersByRoot[d.RootName]; ok {
		return fn(e, d)
	}
	return errSingleRoot
}

func marshalCTe(e *xml.Encoder, d *Document) error {
	if d.CTe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                           `xml:"CTe"`
		XMLNS       string                             `xml:"xmlns,attr,omitempty"`
		InfCte      *cteSchema.TAnonComplexInfCte3     `xml:"infCte"`
		InfCTeSupl  *cteSchema.TAnonComplexInfCTeSupl3 `xml:"infCTeSupl,omitempty"`
		DsSignature *cteSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "CTe"},
		XMLNS:       namespace,
		InfCte:      d.CTe.InfCte,
		InfCTeSupl:  d.CTe.InfCTeSupl,
		DsSignature: d.CTe.DsSignature,
	})
}

func marshalCTeProc(e *xml.Encoder, d *Document) error {
	if d.CTeProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName           xml.Name            `xml:"cteProc"`
		XMLNS             string              `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string              `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string             `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string             `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string             `xml:"dhConexao,attr,omitempty"`
		CTe               *cteSchema.TCTe     `xml:"CTe"`
		ProtCTe           *cteSchema.TProtCTe `xml:"protCTe"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "cteProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.CTeProc.VersaoAttr),
		IpTransmissorAttr: d.CTeProc.IpTransmissorAttr,
		NPortaConAttr:     d.CTeProc.NPortaConAttr,
		DhConexaoAttr:     d.CTeProc.DhConexaoAttr,
		CTe:               d.CTeProc.CTe,
		ProtCTe:           d.CTeProc.ProtCTe,
	})
}

func marshalRetCTe(e *xml.Encoder, d *Document) error {
	if d.RetCTe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retCTe", namespace, d.RetCTe)
}

func marshalCTeOS(e *xml.Encoder, d *Document) error {
	if d.CTeOS == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                             `xml:"CTeOS"`
		XMLNS       string                               `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string                               `xml:"versao,attr,omitempty"`
		InfCte      *cteOSSchema.TAnonComplexInfCte4     `xml:"infCte"`
		InfCTeSupl  *cteOSSchema.TAnonComplexInfCTeSupl4 `xml:"infCTeSupl,omitempty"`
		DsSignature *cteOSSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "CTeOS"},
		XMLNS:       namespace,
		VersaoAttr:  xmlutil.FirstNonEmpty(d.VersaoAttr, d.CTeOS.VersaoAttr),
		InfCte:      d.CTeOS.InfCte,
		InfCTeSupl:  d.CTeOS.InfCTeSupl,
		DsSignature: d.CTeOS.DsSignature,
	})
}

func marshalCTeOSProc(e *xml.Encoder, d *Document) error {
	if d.CTeOSProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName           xml.Name                `xml:"cteOSProc"`
		XMLNS             string                  `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string                  `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string                 `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string                 `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string                 `xml:"dhConexao,attr,omitempty"`
		CTeOS             *cteOSSchema.TCTeOS     `xml:"CTeOS"`
		ProtCTe           *cteOSSchema.TProtCTeOS `xml:"protCTe"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "cteOSProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.CTeOSProc.VersaoAttr),
		IpTransmissorAttr: d.CTeOSProc.IpTransmissorAttr,
		NPortaConAttr:     d.CTeOSProc.NPortaConAttr,
		DhConexaoAttr:     d.CTeOSProc.DhConexaoAttr,
		CTeOS:             d.CTeOSProc.CTeOS,
		ProtCTe:           d.CTeOSProc.ProtCTe,
	})
}

func marshalRetCTeOS(e *xml.Encoder, d *Document) error {
	if d.RetCTeOS == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retCTeOS", namespace, d.RetCTeOS)
}

func marshalCTeSimp(e *xml.Encoder, d *Document) error {
	if d.CTeSimp == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                               `xml:"CTeSimp"`
		XMLNS       string                                 `xml:"xmlns,attr,omitempty"`
		InfCte      *cteSimpSchema.TAnonComplexInfCte2     `xml:"infCte"`
		InfCTeSupl  *cteSimpSchema.TAnonComplexInfCTeSupl2 `xml:"infCTeSupl,omitempty"`
		DsSignature *cteSimpSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "CTeSimp"},
		XMLNS:       namespace,
		InfCte:      d.CTeSimp.InfCte,
		InfCTeSupl:  d.CTeSimp.InfCTeSupl,
		DsSignature: d.CTeSimp.DsSignature,
	})
}

func marshalCTeSimpProc(e *xml.Encoder, d *Document) error {
	if d.CTeSimpProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName           xml.Name                `xml:"cteSimpProc"`
		XMLNS             string                  `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string                  `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string                 `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string                 `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string                 `xml:"dhConexao,attr,omitempty"`
		CTeSimp           *cteSimpSchema.TCTeSimp `xml:"CTeSimp"`
		ProtCTe           *cteSimpSchema.TProtCTe `xml:"protCTe"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "cteSimpProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.CTeSimpProc.VersaoAttr),
		IpTransmissorAttr: d.CTeSimpProc.IpTransmissorAttr,
		NPortaConAttr:     d.CTeSimpProc.NPortaConAttr,
		DhConexaoAttr:     d.CTeSimpProc.DhConexaoAttr,
		CTeSimp:           d.CTeSimpProc.CTeSimp,
		ProtCTe:           d.CTeSimpProc.ProtCTe,
	})
}

func marshalRetCTeSimp(e *xml.Encoder, d *Document) error {
	if d.RetCTeSimp == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retCTeSimp", namespace, d.RetCTeSimp)
}

func marshalGTVe(e *xml.Encoder, d *Document) error {
	if d.GTVe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName     xml.Name                            `xml:"GTVe"`
		XMLNS       string                              `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string                              `xml:"versao,attr,omitempty"`
		InfCte      *gtveSchema.TAnonComplexInfCte1     `xml:"infCte"`
		InfCTeSupl  *gtveSchema.TAnonComplexInfCTeSupl1 `xml:"infCTeSupl,omitempty"`
		DsSignature *gtveSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:     xml.Name{Local: "GTVe"},
		XMLNS:       namespace,
		VersaoAttr:  xmlutil.FirstNonEmpty(d.VersaoAttr, d.GTVe.VersaoAttr),
		InfCte:      d.GTVe.InfCte,
		InfCTeSupl:  d.GTVe.InfCTeSupl,
		DsSignature: d.GTVe.DsSignature,
	})
}

func marshalGTVeProc(e *xml.Encoder, d *Document) error {
	if d.GTVeProc == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName           xml.Name              `xml:"GTVeProc"`
		XMLNS             string                `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string                `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string               `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string               `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string               `xml:"dhConexao,attr,omitempty"`
		GTVe              *gtveSchema.TGTVe     `xml:"GTVe"`
		ProtCTe           *gtveSchema.TProtGTVe `xml:"protCTe"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:           xml.Name{Local: "GTVeProc"},
		XMLNS:             namespace,
		VersaoAttr:        xmlutil.FirstNonEmpty(d.VersaoAttr, d.GTVeProc.VersaoAttr),
		IpTransmissorAttr: d.GTVeProc.IpTransmissorAttr,
		NPortaConAttr:     d.GTVeProc.NPortaConAttr,
		DhConexaoAttr:     d.GTVeProc.DhConexaoAttr,
		GTVe:              d.GTVeProc.GTVe,
		ProtCTe:           d.GTVeProc.ProtCTe,
	})
}

func marshalRetGTVe(e *xml.Encoder, d *Document) error {
	if d.RetGTVe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retGTVe", namespace, d.RetGTVe)
}

func marshalConsSitCTe(e *xml.Encoder, d *Document) error {
	if d.ConsSitCTe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "consSitCTe", namespace, d.ConsSitCTe)
}

func marshalRetConsSitCTe(e *xml.Encoder, d *Document) error {
	if d.RetConsSitCTe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retConsSitCTe", namespace, d.RetConsSitCTe)
}

func marshalConsStatServCTe(e *xml.Encoder, d *Document) error {
	if d.ConsStatServCTe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "consStatServCTe", namespace, d.ConsStatServCTe)
}

func marshalRetConsStatServCTe(e *xml.Encoder, d *Document) error {
	if d.RetConsStatServCTe == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	return xmlutil.EncodeNamespacedRoot(e, "retConsStatServCTe", namespace, d.RetConsStatServCTe)
}

func marshalDistDFeInt(e *xml.Encoder, d *Document) error {
	if d.DistDFeInt == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName    xml.Name                         `xml:"distDFeInt"`
		XMLNS      string                           `xml:"xmlns,attr,omitempty"`
		VersaoAttr string                           `xml:"versao,attr,omitempty"`
		TpAmb      string                           `xml:"tpAmb"`
		CUFAutor   string                           `xml:"cUFAutor"`
		CNPJ       *string                          `xml:"CNPJ,omitempty"`
		CPF        *string                          `xml:"CPF,omitempty"`
		DistNSU    *distSchema.TAnonComplexDistNSU1 `xml:"distNSU,omitempty"`
		ConsNSU    *distSchema.TAnonComplexConsNSU1 `xml:"consNSU,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:    xml.Name{Local: "distDFeInt"},
		XMLNS:      namespace,
		VersaoAttr: xmlutil.FirstNonEmpty(d.VersaoAttr, d.DistDFeInt.VersaoAttr),
		TpAmb:      d.DistDFeInt.TpAmb,
		CUFAutor:   d.DistDFeInt.CUFAutor,
		CNPJ:       d.DistDFeInt.CNPJ,
		CPF:        d.DistDFeInt.CPF,
		DistNSU:    d.DistDFeInt.DistNSU,
		ConsNSU:    d.DistDFeInt.ConsNSU,
	})
}

func marshalRetDistDFeInt(e *xml.Encoder, d *Document) error {
	if d.RetDistDFeInt == nil || activeRootCount(d) != 1 {
		return errSingleRoot
	}
	type root struct {
		XMLName        xml.Name                                `xml:"retDistDFeInt"`
		XMLNS          string                                  `xml:"xmlns,attr,omitempty"`
		VersaoAttr     string                                  `xml:"versao,attr,omitempty"`
		TpAmb          string                                  `xml:"tpAmb"`
		VerAplic       *distSchema.TString                     `xml:"verAplic,omitempty"`
		CStat          string                                  `xml:"cStat"`
		XMotivo        *distSchema.TString                     `xml:"xMotivo,omitempty"`
		DhResp         string                                  `xml:"dhResp"`
		UltNSU         string                                  `xml:"ultNSU"`
		MaxNSU         string                                  `xml:"maxNSU"`
		LoteDistDFeInt *distSchema.TAnonComplexLoteDistDFeInt1 `xml:"loteDistDFeInt,omitempty"`
	}
	return xmlutil.EncodeCanonical(e, root{
		XMLName:        xml.Name{Local: "retDistDFeInt"},
		XMLNS:          namespace,
		VersaoAttr:     xmlutil.FirstNonEmpty(d.VersaoAttr, d.RetDistDFeInt.VersaoAttr),
		TpAmb:          d.RetDistDFeInt.TpAmb,
		VerAplic:       d.RetDistDFeInt.VerAplic,
		CStat:          d.RetDistDFeInt.CStat,
		XMotivo:        d.RetDistDFeInt.XMotivo,
		DhResp:         d.RetDistDFeInt.DhResp,
		UltNSU:         d.RetDistDFeInt.UltNSU,
		MaxNSU:         d.RetDistDFeInt.MaxNSU,
		LoteDistDFeInt: d.RetDistDFeInt.LoteDistDFeInt,
	})
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("parse cte: %w", fiscalerr.ErrEmptyDocument)
	}

	RootName, rootErr := xmlutil.ParseRootName(data)
	if rootErr != nil && RootName == "" {
		return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
	}

	if fn, ok := parsersByRoot[RootName]; ok {
		return fn(data, RootName)
	}
	if rootErr != nil {
		return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
	}
	return nil, fmt.Errorf("parse cte: %w", &fiscalerr.UnsupportedRootError{Family: fiscalerr.CTe, Root: RootName})
}

func ParseReader(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parse cte: read xml: %w", err)
	}
	return Parse(data)
}

func finalizeDoc(doc *Document) (*Document, error) {
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func parseCTe(data []byte, rootName string) (*Document, error) {
	var parsed cteSchema.TCTe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode CTe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: versionFromInfCte(parsed.InfCte), CTe: &parsed, RootName: rootName})
}

func parseCTeProc(data []byte, rootName string) (*Document, error) {
	var parsed cteSchema.TAnonComplexCteProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode cteProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, CTeProc: &parsed, RootName: rootName})
}

func parseRetCTe(data []byte, rootName string) (*Document, error) {
	var parsed cteSchema.TRetCTe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retCTe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetCTe: &parsed, RootName: rootName})
}

func parseCTeOS(data []byte, rootName string) (*Document, error) {
	var parsed struct {
		VersaoAttr string                               `xml:"versao,attr"`
		InfCte     *cteOSSchema.TAnonComplexInfCte4     `xml:"infCte"`
		InfCTeSupl *cteOSSchema.TAnonComplexInfCTeSupl4 `xml:"infCTeSupl"`
		Signature  *cteOSSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode CTeOS: %w", err)
	}
	return finalizeDoc(&Document{
		VersaoAttr: parsed.VersaoAttr,
		CTeOS: &cteOSSchema.TCTeOS{
			VersaoAttr:  parsed.VersaoAttr,
			InfCte:      parsed.InfCte,
			InfCTeSupl:  parsed.InfCTeSupl,
			DsSignature: parsed.Signature,
		},
		RootName: rootName,
	})
}

func parseCTeOSProc(data []byte, rootName string) (*Document, error) {
	var parsed cteOSSchema.TAnonComplexCteOSProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode cteOSProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, CTeOSProc: &parsed, RootName: rootName})
}

func parseRetCTeOS(data []byte, rootName string) (*Document, error) {
	var parsed cteOSSchema.TRetCTeOS
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retCTeOS: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetCTeOS: &parsed, RootName: rootName})
}

func parseCTeSimp(data []byte, rootName string) (*Document, error) {
	var parsed cteSimpSchema.TCTeSimp
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode CTeSimp: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: versionFromInfCteSimp(parsed.InfCte), CTeSimp: &parsed, RootName: rootName})
}

func parseCTeSimpProc(data []byte, rootName string) (*Document, error) {
	var parsed cteSimpSchema.TAnonComplexCteSimpProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode cteSimpProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, CTeSimpProc: &parsed, RootName: rootName})
}

func parseRetCTeSimp(data []byte, rootName string) (*Document, error) {
	var parsed cteSimpSchema.TRetCTeSimp
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retCTeSimp: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetCTeSimp: &parsed, RootName: rootName})
}

func parseGTVe(data []byte, rootName string) (*Document, error) {
	var parsed gtveSchema.TGTVe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode GTVe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, GTVe: &parsed, RootName: rootName})
}

func parseGTVeProc(data []byte, rootName string) (*Document, error) {
	var parsed gtveSchema.TAnonComplexGTVeProc1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode GTVeProc: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, GTVeProc: &parsed, RootName: rootName})
}

func parseRetGTVe(data []byte, rootName string) (*Document, error) {
	var parsed gtveSchema.TRetGTVe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retGTVe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetGTVe: &parsed, RootName: rootName})
}

func parseConsSitCTe(data []byte, rootName string) (*Document, error) {
	var parsed consSitSchema.TConsSitCTe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode consSitCTe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsSitCTe: &parsed, RootName: rootName})
}

func parseRetConsSitCTe(data []byte, rootName string) (*Document, error) {
	var parsed consSitSchema.TRetConsSitCTe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retConsSitCTe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsSitCTe: &parsed, RootName: rootName})
}

func parseConsStatServCTe(data []byte, rootName string) (*Document, error) {
	var parsed statusSchema.TConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode consStatServCTe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ConsStatServCTe: &parsed, RootName: rootName})
}

func parseRetConsStatServCTe(data []byte, rootName string) (*Document, error) {
	var parsed statusSchema.TRetConsStatServ
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retConsStatServCTe: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServCTe: &parsed, RootName: rootName})
}

func parseDistDFeInt(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexDistDFeInt1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode distDFeInt: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, RootName: rootName})
}

func parseRetDistDFeInt(data []byte, rootName string) (*Document, error) {
	var parsed distSchema.TAnonComplexRetDistDFeInt1
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse cte: decode retDistDFeInt: %w", err)
	}
	return finalizeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, RootName: rootName})
}

func versionFromInfCte(inf *cteSchema.TAnonComplexInfCte3) string {
	if inf == nil {
		return ""
	}
	return inf.VersaoAttr
}

func versionFromInfCteSimp(inf *cteSimpSchema.TAnonComplexInfCte2) string {
	if inf == nil {
		return ""
	}
	return inf.VersaoAttr
}
