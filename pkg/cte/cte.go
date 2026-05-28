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

var rootValidators = []func(*Document) error{
	validateCTeRoot,
	validateCTeProcRoot,
	validateRetCTeRoot,
	validateCTeOSRoot,
	validateCTeOSProcRoot,
	validateRetCTeOSRoot,
	validateCTeSimpRoot,
	validateCTeSimpProcRoot,
	validateRetCTeSimpRoot,
	validateGTVeRoot,
	validateGTVeProcRoot,
	validateRetGTVeRoot,
	validateConsSitCTeRoot,
	validateRetConsSitCTeRoot,
	validateConsStatServCTeRoot,
	validateRetConsStatServCTeRoot,
	validateEventoCTeRoot,
	validateEventoCancCTeRoot,
	validateEventoCECTeRoot,
	validateEventoCancCECTeRoot,
	validateEventoEPECCTeRoot,
	validateEventoRegMultimodalRoot,
	validateEventoGTVRoot,
	validateEventoIECTeRoot,
	validateEventoCancIECTeRoot,
	validateEventoPrestDesacordoRoot,
	validateEventoCancPrestDesacordoRoot,
	validateEventoGenericoRoot,
	validateRetEventoCTeRoot,
	validateRetEventoCancCTeRoot,
	validateRetEventoCECTeRoot,
	validateRetEventoCancCECTeRoot,
	validateRetEventoEPECCTeRoot,
	validateRetEventoRegMultimodalRoot,
	validateRetEventoGTVRoot,
	validateRetEventoIECTeRoot,
	validateRetEventoCancIECTeRoot,
	validateRetEventoPrestDesacordoRoot,
	validateRetEventoCancPrestDesacordoRoot,
	validateRetEventoGenericoRoot,
	validateProcEventoCTeRoot,
	validateProcEventoCancCTeRoot,
	validateProcEventoCECTeRoot,
	validateProcEventoCancCECTeRoot,
	validateProcEventoEPECCTeRoot,
	validateProcEventoRegMultimodalRoot,
	validateProcEventoGTVRoot,
	validateProcEventoIECTeRoot,
	validateProcEventoCancIECTeRoot,
	validateProcEventoPrestDesacordoRoot,
	validateProcEventoCancPrestDesacordoRoot,
	validateProcEventoGenericoRoot,
	validateDistDFeIntRoot,
	validateRetDistDFeIntRoot,
}

func validateDocument(doc *Document) error {
	for _, v := range rootValidators {
		if err := v(doc); err != nil {
			return err
		}
	}
	if activeRootCount(doc) != 1 {
		return errors.New("parse cte: document must contain exactly one supported root")
	}
	return nil
}

func missing(field, value string) error {
	if value == "" {
		return errors.New("parse cte: missing " + field)
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

func validateCTeRoot(doc *Document) error {
	if doc.CTe == nil {
		return nil
	}
	return validateInfCte(doc.CTe.InfCte)
}

func validateCTeProcRoot(doc *Document) error {
	if doc.CTeProc == nil {
		return nil
	}
	if doc.CTeProc.CTe == nil {
		return errors.New("parse cte: missing CTe")
	}
	if doc.CTeProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetCTeRoot(doc *Document) error {
	if doc.RetCTe == nil {
		return nil
	}
	return missing("cStat", doc.RetCTe.CStat)
}

func validateCTeOSRoot(doc *Document) error {
	if doc.CTeOS == nil {
		return nil
	}
	return validateInfCteOS(doc.CTeOS.InfCte)
}

func validateCTeOSProcRoot(doc *Document) error {
	if doc.CTeOSProc == nil {
		return nil
	}
	if doc.CTeOSProc.CTeOS == nil {
		return errors.New("parse cte: missing CTeOS")
	}
	if doc.CTeOSProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetCTeOSRoot(doc *Document) error {
	if doc.RetCTeOS == nil {
		return nil
	}
	return missing("cStat", doc.RetCTeOS.CStat)
}

func validateCTeSimpRoot(doc *Document) error {
	if doc.CTeSimp == nil {
		return nil
	}
	if doc.CTeSimp.InfCte == nil {
		return errors.New("parse cte: missing infCte")
	}
	return nil
}

func validateCTeSimpProcRoot(doc *Document) error {
	if doc.CTeSimpProc == nil {
		return nil
	}
	if doc.CTeSimpProc.CTeSimp == nil {
		return errors.New("parse cte: missing CTeSimp")
	}
	if doc.CTeSimpProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetCTeSimpRoot(doc *Document) error {
	if doc.RetCTeSimp == nil {
		return nil
	}
	return missing("cStat", doc.RetCTeSimp.CStat)
}

func validateGTVeRoot(doc *Document) error {
	if doc.GTVe == nil {
		return nil
	}
	if doc.GTVe.InfCte == nil {
		return errors.New("parse cte: missing infCte")
	}
	return nil
}

func validateGTVeProcRoot(doc *Document) error {
	if doc.GTVeProc == nil {
		return nil
	}
	if doc.GTVeProc.GTVe == nil {
		return errors.New("parse cte: missing GTVe")
	}
	if doc.GTVeProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetGTVeRoot(doc *Document) error {
	if doc.RetGTVe == nil {
		return nil
	}
	return missing("cStat", doc.RetGTVe.CStat)
}

func validateConsSitCTeRoot(doc *Document) error {
	if doc.ConsSitCTe == nil {
		return nil
	}
	return missing("chCTe", doc.ConsSitCTe.ChCTe)
}

func validateRetConsSitCTeRoot(doc *Document) error {
	if doc.RetConsSitCTe == nil {
		return nil
	}
	return missing("cStat", doc.RetConsSitCTe.CStat)
}

func validateConsStatServCTeRoot(doc *Document) error {
	if doc.ConsStatServCTe == nil {
		return nil
	}
	return missing("xServ", doc.ConsStatServCTe.XServ)
}

func validateRetConsStatServCTeRoot(doc *Document) error {
	if doc.RetConsStatServCTe == nil {
		return nil
	}
	return missing("cStat", doc.RetConsStatServCTe.CStat)
}

func validateDistDFeIntRoot(doc *Document) error {
	if doc.DistDFeInt == nil {
		return nil
	}
	if err := firstMissing(
		missing("tpAmb", doc.DistDFeInt.TpAmb),
		missing("cUFAutor", doc.DistDFeInt.CUFAutor),
	); err != nil {
		return err
	}
	if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
		return errors.New("parse cte: missing dist document")
	}
	if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil {
		return errors.New("parse cte: missing dist query")
	}
	return nil
}

func validateRetDistDFeIntRoot(doc *Document) error {
	if doc.RetDistDFeInt == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetDistDFeInt.TpAmb),
		missing("cStat", doc.RetDistDFeInt.CStat),
		missing("ultNSU", doc.RetDistDFeInt.UltNSU),
		missing("maxNSU", doc.RetDistDFeInt.MaxNSU),
	)
}

func validateEventoCTeRoot(doc *Document) error {
	if doc.EventoCTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCTe.InfEvento)
}

func validateEventoCancCTeRoot(doc *Document) error {
	if doc.EventoCancCTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancCTe.InfEvento)
}

func validateEventoCECTeRoot(doc *Document) error {
	if doc.EventoCECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCECTe.InfEvento)
}

func validateEventoCancCECTeRoot(doc *Document) error {
	if doc.EventoCancCECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancCECTe.InfEvento)
}

func validateEventoEPECCTeRoot(doc *Document) error {
	if doc.EventoEPECCTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoEPECCTe.InfEvento)
}

func validateEventoRegMultimodalRoot(doc *Document) error {
	if doc.EventoRegMultimodal == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoRegMultimodal.InfEvento)
}

func validateEventoGTVRoot(doc *Document) error {
	if doc.EventoGTV == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoGTV.InfEvento)
}

func validateEventoIECTeRoot(doc *Document) error {
	if doc.EventoIECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoIECTe.InfEvento)
}

func validateEventoCancIECTeRoot(doc *Document) error {
	if doc.EventoCancIECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancIECTe.InfEvento)
}

func validateEventoPrestDesacordoRoot(doc *Document) error {
	if doc.EventoPrestDesacordo == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoPrestDesacordo.InfEvento)
}

func validateEventoCancPrestDesacordoRoot(doc *Document) error {
	if doc.EventoCancPrestDesacordo == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancPrestDesacordo.InfEvento)
}

func validateEventoGenericoRoot(doc *Document) error {
	if doc.EventoGenerico == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoGenerico.InfEvento)
}

func missingEventoCTeIfNil(present, eventoNil bool) error {
	if present && eventoNil {
		return errors.New("parse cte: missing eventoCTe")
	}
	return nil
}

func missingRetEventoCTeIfNil(present, retEventoNil bool) error {
	if present && retEventoNil {
		return errors.New("parse cte: missing retEventoCTe")
	}
	return nil
}

func validateRetEventoCTeRoot(doc *Document) error {
	if doc.RetEventoCTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCTe)
}

func validateRetEventoCancCTeRoot(doc *Document) error {
	if doc.RetEventoCancCTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancCTe)
}

func validateRetEventoCECTeRoot(doc *Document) error {
	if doc.RetEventoCECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCECTe)
}

func validateRetEventoCancCECTeRoot(doc *Document) error {
	if doc.RetEventoCancCECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancCECTe)
}

func validateRetEventoEPECCTeRoot(doc *Document) error {
	if doc.RetEventoEPECCTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoEPECCTe)
}

func validateRetEventoRegMultimodalRoot(doc *Document) error {
	if doc.RetEventoRegMultimodal == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoRegMultimodal)
}

func validateRetEventoGTVRoot(doc *Document) error {
	if doc.RetEventoGTV == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoGTV)
}

func validateRetEventoIECTeRoot(doc *Document) error {
	if doc.RetEventoIECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoIECTe)
}

func validateRetEventoCancIECTeRoot(doc *Document) error {
	if doc.RetEventoCancIECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancIECTe)
}

func validateRetEventoPrestDesacordoRoot(doc *Document) error {
	if doc.RetEventoPrestDesacordo == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoPrestDesacordo)
}

func validateRetEventoCancPrestDesacordoRoot(doc *Document) error {
	if doc.RetEventoCancPrestDesacordo == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancPrestDesacordo)
}

func validateRetEventoGenericoRoot(doc *Document) error {
	if doc.RetEventoGenerico == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoGenerico)
}

func validateProcEventoCTeRoot(doc *Document) error {
	if doc.ProcEventoCTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCTe.EventoCTe, doc.ProcEventoCTe.RetEventoCTe)
}

func validateProcEventoCancCTeRoot(doc *Document) error {
	if doc.ProcEventoCancCTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancCTe.EventoCTe, doc.ProcEventoCancCTe.RetEventoCTe)
}

func validateProcEventoCECTeRoot(doc *Document) error {
	if doc.ProcEventoCECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCECTe.EventoCTe, doc.ProcEventoCECTe.RetEventoCTe)
}

func validateProcEventoCancCECTeRoot(doc *Document) error {
	if doc.ProcEventoCancCECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancCECTe.EventoCTe, doc.ProcEventoCancCECTe.RetEventoCTe)
}

func validateProcEventoEPECCTeRoot(doc *Document) error {
	if doc.ProcEventoEPECCTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoEPECCTe.EventoCTe, doc.ProcEventoEPECCTe.RetEventoCTe)
}

func validateProcEventoRegMultimodalRoot(doc *Document) error {
	if doc.ProcEventoRegMultimodal == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoRegMultimodal.EventoCTe, doc.ProcEventoRegMultimodal.RetEventoCTe)
}

func validateProcEventoGTVRoot(doc *Document) error {
	if doc.ProcEventoGTV == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoGTV.EventoCTe, doc.ProcEventoGTV.RetEventoCTe)
}

func validateProcEventoIECTeRoot(doc *Document) error {
	if doc.ProcEventoIECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoIECTe.EventoCTe, doc.ProcEventoIECTe.RetEventoCTe)
}

func validateProcEventoCancIECTeRoot(doc *Document) error {
	if doc.ProcEventoCancIECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancIECTe.EventoCTe, doc.ProcEventoCancIECTe.RetEventoCTe)
}

func validateProcEventoPrestDesacordoRoot(doc *Document) error {
	if doc.ProcEventoPrestDesacordo == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoPrestDesacordo.EventoCTe, doc.ProcEventoPrestDesacordo.RetEventoCTe)
}

func validateProcEventoCancPrestDesacordoRoot(doc *Document) error {
	if doc.ProcEventoCancPrestDesacordo == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancPrestDesacordo.EventoCTe, doc.ProcEventoCancPrestDesacordo.RetEventoCTe)
}

func validateProcEventoGenericoRoot(doc *Document) error {
	if doc.ProcEventoGenerico == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoGenerico.EventoCTe, doc.ProcEventoGenerico.RetEventoCTe)
}

func validateCTeProcessedEvent(evento, retEvento any) error {
	if err := validateCTeSentEventRoot(evento); err != nil {
		return err
	}
	return validateCTeRetEventRoot(retEvento)
}

func validateCTeSentEventRoot(evento any) error {
	switch v := evento.(type) {
	case *eventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *ceEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelCEEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *epecEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *regMultimodalEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *gtvEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *ieEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelIEEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *prestDesacordoEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelPrestDesacordoEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *genericEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	default:
		return missingEventoCTeIfNil(true, true)
	}
}

func validateCTeRetEventRoot(retEvento any) error {
	switch v := retEvento.(type) {
	case *eventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *ceEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelCEEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *epecEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *regMultimodalEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *gtvEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *ieEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelIEEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *prestDesacordoEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelPrestDesacordoEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *genericEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	default:
		return missingRetEventoCTeIfNil(true, true)
	}
}

func validateInfCte(inf *cteSchema.TAnonComplexInfCte3) error {
	if inf == nil {
		return errors.New("parse cte: missing infCte")
	}
	if inf.Ide == nil {
		return errors.New("parse cte: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse cte: missing emit")
	}
	if inf.Emit.CNPJ == nil && inf.Emit.CPF == nil {
		return errors.New("parse cte: missing emit document")
	}
	return nil
}

func validateInfCteOS(inf *cteOSSchema.TAnonComplexInfCte4) error {
	if inf == nil {
		return errors.New("parse cte: missing infCte")
	}
	if inf.Ide == nil {
		return errors.New("parse cte: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse cte: missing emit")
	}
	if inf.Emit.CNPJ == "" {
		return errors.New("parse cte: missing emit document")
	}
	return nil
}

func validateCTeEvent(inf any) error {
	switch v := inf.(type) {
	case *eventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *ceEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelCEEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *epecEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *regMultimodalEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *gtvEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *ieEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelIEEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *prestDesacordoEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelPrestDesacordoEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *genericEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	default:
		return errors.New("parse cte: missing infEvento")
	}
}

func validateCTeRetEvent(inf any) error {
	switch v := inf.(type) {
	case *eventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *ceEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelCEEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *epecEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *regMultimodalEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *gtvEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *ieEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelIEEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *prestDesacordoEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelPrestDesacordoEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *genericEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	default:
		return errors.New("parse cte: missing infEvento")
	}
}

func validateInfEvento(ok bool, chCTe string, hasDet bool) error {
	if !ok {
		return errors.New("parse cte: missing infEvento")
	}
	if chCTe == "" {
		return errors.New("parse cte: missing chCTe")
	}
	if !hasDet {
		return errors.New("parse cte: missing detEvento")
	}
	return nil
}

func validateRetInfEvento(ok bool, tpAmb, cStat string) error {
	if !ok {
		return errors.New("parse cte: missing infEvento")
	}
	return firstMissing(
		missing("tpAmb", tpAmb),
		missing("cStat", cStat),
	)
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

func activeRootCount(doc *Document) int {
	count := 0
	for _, ok := range []bool{
		doc.CTe != nil,
		doc.CTeProc != nil,
		doc.RetCTe != nil,
		doc.CTeOS != nil,
		doc.CTeOSProc != nil,
		doc.RetCTeOS != nil,
		doc.CTeSimp != nil,
		doc.CTeSimpProc != nil,
		doc.RetCTeSimp != nil,
		doc.GTVe != nil,
		doc.GTVeProc != nil,
		doc.RetGTVe != nil,
		doc.ConsSitCTe != nil,
		doc.RetConsSitCTe != nil,
		doc.ConsStatServCTe != nil,
		doc.RetConsStatServCTe != nil,
		doc.EventoCTe != nil,
		doc.RetEventoCTe != nil,
		doc.ProcEventoCTe != nil,
		doc.EventoCancCTe != nil,
		doc.RetEventoCancCTe != nil,
		doc.ProcEventoCancCTe != nil,
		doc.EventoCECTe != nil,
		doc.RetEventoCECTe != nil,
		doc.ProcEventoCECTe != nil,
		doc.EventoCancCECTe != nil,
		doc.RetEventoCancCECTe != nil,
		doc.ProcEventoCancCECTe != nil,
		doc.EventoEPECCTe != nil,
		doc.RetEventoEPECCTe != nil,
		doc.ProcEventoEPECCTe != nil,
		doc.EventoRegMultimodal != nil,
		doc.RetEventoRegMultimodal != nil,
		doc.ProcEventoRegMultimodal != nil,
		doc.EventoGTV != nil,
		doc.RetEventoGTV != nil,
		doc.ProcEventoGTV != nil,
		doc.EventoIECTe != nil,
		doc.RetEventoIECTe != nil,
		doc.ProcEventoIECTe != nil,
		doc.EventoCancIECTe != nil,
		doc.RetEventoCancIECTe != nil,
		doc.ProcEventoCancIECTe != nil,
		doc.EventoPrestDesacordo != nil,
		doc.RetEventoPrestDesacordo != nil,
		doc.ProcEventoPrestDesacordo != nil,
		doc.EventoCancPrestDesacordo != nil,
		doc.RetEventoCancPrestDesacordo != nil,
		doc.ProcEventoCancPrestDesacordo != nil,
		doc.EventoGenerico != nil,
		doc.RetEventoGenerico != nil,
		doc.ProcEventoGenerico != nil,
		doc.DistDFeInt != nil,
		doc.RetDistDFeInt != nil,
	} {
		if ok {
			count++
		}
	}
	return count
}
