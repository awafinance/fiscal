package cte

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	distSchema "github.com/awa/fiscal/internal/cte/gen/v1_0/dist_dfe"
	consSitSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/consulta_situacao"
	cteSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/cte"
	cteOSSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/cte_os"
	cteSimpSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/cte_simp"
	cancelEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_cancel"
	cancelCEEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_cancel_ce"
	cancelIEEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_cancel_ie"
	cancelPrestDesacordoEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_cancel_prest_desacordo"
	eventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_cce"
	ceEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_ce"
	epecEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_epec"
	gtvEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_gtv"
	ieEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_ie"
	prestDesacordoEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_prest_desacordo"
	regMultimodalEventSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/evento_reg_multimodal"
	gtveSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/gtve"
	statusSchema "github.com/awa/fiscal/internal/cte/gen/v4_0/status_servico"
)

const namespace = "http://www.portalfiscal.inf.br/cte"

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
	DistDFeInt                   *distSchema.TAnonComplexDistDFeInt1          `json:"distDFeInt,omitempty"`
	RetDistDFeInt                *distSchema.TAnonComplexRetDistDFeInt1       `json:"retDistDFeInt,omitempty"`
	rootName                     string                                       `json:"-"`
}

func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	switch d.rootName {
	case "CTe", "":
		if d.CTe != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName     xml.Name                           `xml:"CTe"`
				XMLNS       string                             `xml:"xmlns,attr,omitempty"`
				InfCte      *cteSchema.TAnonComplexInfCte3     `xml:"infCte"`
				InfCTeSupl  *cteSchema.TAnonComplexInfCTeSupl3 `xml:"infCTeSupl,omitempty"`
				DsSignature *cteSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "CTe"},
				XMLNS:       namespace,
				InfCte:      d.CTe.InfCte,
				InfCTeSupl:  d.CTe.InfCTeSupl,
				DsSignature: d.CTe.DsSignature,
			})
		}
	case "cteProc":
		if d.CTeProc != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:           xml.Name{Local: "cteProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.CTeProc.VersaoAttr),
				IpTransmissorAttr: d.CTeProc.IpTransmissorAttr,
				NPortaConAttr:     d.CTeProc.NPortaConAttr,
				DhConexaoAttr:     d.CTeProc.DhConexaoAttr,
				CTe:               d.CTeProc.CTe,
				ProtCTe:           d.CTeProc.ProtCTe,
			})
		}
	case "retCTe":
		if d.RetCTe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retCTe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*cteSchema.TRetCTe
			}{
				XMLName: xml.Name{Local: "retCTe"},
				XMLNS:   namespace,
				TRetCTe: d.RetCTe,
			})
		}
	case "CTeOS":
		if d.CTeOS != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName     xml.Name                             `xml:"CTeOS"`
				XMLNS       string                               `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                               `xml:"versao,attr,omitempty"`
				InfCte      *cteOSSchema.TAnonComplexInfCte4     `xml:"infCte"`
				InfCTeSupl  *cteOSSchema.TAnonComplexInfCTeSupl4 `xml:"infCTeSupl,omitempty"`
				DsSignature *cteOSSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}

			return e.Encode(root{
				XMLName:     xml.Name{Local: "CTeOS"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.CTeOS.VersaoAttr),
				InfCte:      d.CTeOS.InfCte,
				InfCTeSupl:  d.CTeOS.InfCTeSupl,
				DsSignature: d.CTeOS.DsSignature,
			})
		}
	case "cteOSProc":
		if d.CTeOSProc != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:           xml.Name{Local: "cteOSProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.CTeOSProc.VersaoAttr),
				IpTransmissorAttr: d.CTeOSProc.IpTransmissorAttr,
				NPortaConAttr:     d.CTeOSProc.NPortaConAttr,
				DhConexaoAttr:     d.CTeOSProc.DhConexaoAttr,
				CTeOS:             d.CTeOSProc.CTeOS,
				ProtCTe:           d.CTeOSProc.ProtCTe,
			})
		}
	case "retCTeOS":
		if d.RetCTeOS != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retCTeOS"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*cteOSSchema.TRetCTeOS
			}{
				XMLName:   xml.Name{Local: "retCTeOS"},
				XMLNS:     namespace,
				TRetCTeOS: d.RetCTeOS,
			})
		}
	case "CTeSimp":
		if d.CTeSimp != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName     xml.Name                               `xml:"CTeSimp"`
				XMLNS       string                                 `xml:"xmlns,attr,omitempty"`
				InfCte      *cteSimpSchema.TAnonComplexInfCte2     `xml:"infCte"`
				InfCTeSupl  *cteSimpSchema.TAnonComplexInfCTeSupl2 `xml:"infCTeSupl,omitempty"`
				DsSignature *cteSimpSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}
			return e.Encode(root{
				XMLName:     xml.Name{Local: "CTeSimp"},
				XMLNS:       namespace,
				InfCte:      d.CTeSimp.InfCte,
				InfCTeSupl:  d.CTeSimp.InfCTeSupl,
				DsSignature: d.CTeSimp.DsSignature,
			})
		}
	case "cteSimpProc":
		if d.CTeSimpProc != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:           xml.Name{Local: "cteSimpProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.CTeSimpProc.VersaoAttr),
				IpTransmissorAttr: d.CTeSimpProc.IpTransmissorAttr,
				NPortaConAttr:     d.CTeSimpProc.NPortaConAttr,
				DhConexaoAttr:     d.CTeSimpProc.DhConexaoAttr,
				CTeSimp:           d.CTeSimpProc.CTeSimp,
				ProtCTe:           d.CTeSimpProc.ProtCTe,
			})
		}
	case "retCTeSimp":
		if d.RetCTeSimp != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retCTeSimp"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*cteSimpSchema.TRetCTeSimp
			}{
				XMLName:     xml.Name{Local: "retCTeSimp"},
				XMLNS:       namespace,
				TRetCTeSimp: d.RetCTeSimp,
			})
		}
	case "GTVe":
		if d.GTVe != nil && activeRootCount(d) == 1 {
			type root struct {
				XMLName     xml.Name                            `xml:"GTVe"`
				XMLNS       string                              `xml:"xmlns,attr,omitempty"`
				VersaoAttr  string                              `xml:"versao,attr,omitempty"`
				InfCte      *gtveSchema.TAnonComplexInfCte1     `xml:"infCte"`
				InfCTeSupl  *gtveSchema.TAnonComplexInfCTeSupl1 `xml:"infCTeSupl,omitempty"`
				DsSignature *gtveSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
			}
			return e.Encode(root{
				XMLName:     xml.Name{Local: "GTVe"},
				XMLNS:       namespace,
				VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.GTVe.VersaoAttr),
				InfCte:      d.GTVe.InfCte,
				InfCTeSupl:  d.GTVe.InfCTeSupl,
				DsSignature: d.GTVe.DsSignature,
			})
		}
	case "GTVeProc":
		if d.GTVeProc != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:           xml.Name{Local: "GTVeProc"},
				XMLNS:             namespace,
				VersaoAttr:        firstNonEmpty(d.VersaoAttr, d.GTVeProc.VersaoAttr),
				IpTransmissorAttr: d.GTVeProc.IpTransmissorAttr,
				NPortaConAttr:     d.GTVeProc.NPortaConAttr,
				DhConexaoAttr:     d.GTVeProc.DhConexaoAttr,
				GTVe:              d.GTVeProc.GTVe,
				ProtCTe:           d.GTVeProc.ProtCTe,
			})
		}
	case "retGTVe":
		if d.RetGTVe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retGTVe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*gtveSchema.TRetGTVe
			}{
				XMLName:  xml.Name{Local: "retGTVe"},
				XMLNS:    namespace,
				TRetGTVe: d.RetGTVe,
			})
		}
	case "consSitCTe":
		if d.ConsSitCTe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"consSitCTe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consSitSchema.TConsSitCTe
			}{
				XMLName:     xml.Name{Local: "consSitCTe"},
				XMLNS:       namespace,
				TConsSitCTe: d.ConsSitCTe,
			})
		}
	case "retConsSitCTe":
		if d.RetConsSitCTe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsSitCTe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*consSitSchema.TRetConsSitCTe
			}{
				XMLName:        xml.Name{Local: "retConsSitCTe"},
				XMLNS:          namespace,
				TRetConsSitCTe: d.RetConsSitCTe,
			})
		}
	case "consStatServCTe":
		if d.ConsStatServCTe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"consStatServCTe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*statusSchema.TConsStatServ
			}{
				XMLName:       xml.Name{Local: "consStatServCTe"},
				XMLNS:         namespace,
				TConsStatServ: d.ConsStatServCTe,
			})
		}
	case "retConsStatServCTe":
		if d.RetConsStatServCTe != nil && activeRootCount(d) == 1 {
			return e.Encode(struct {
				XMLName xml.Name `xml:"retConsStatServCTe"`
				XMLNS   string   `xml:"xmlns,attr,omitempty"`
				*statusSchema.TRetConsStatServ
			}{
				XMLName:          xml.Name{Local: "retConsStatServCTe"},
				XMLNS:            namespace,
				TRetConsStatServ: d.RetConsStatServCTe,
			})
		}
	case "eventoCTe":
		return marshalEventRoot(e, d)
	case "retEventoCTe":
		return marshalRetEventRoot(e, d)
	case "procEventoCTe":
		return marshalProcEventRoot(e, d)
	case "distDFeInt":
		if d.DistDFeInt != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:    xml.Name{Local: "distDFeInt"},
				XMLNS:      namespace,
				VersaoAttr: firstNonEmpty(d.VersaoAttr, d.DistDFeInt.VersaoAttr),
				TpAmb:      d.DistDFeInt.TpAmb,
				CUFAutor:   d.DistDFeInt.CUFAutor,
				CNPJ:       d.DistDFeInt.CNPJ,
				CPF:        d.DistDFeInt.CPF,
				DistNSU:    d.DistDFeInt.DistNSU,
				ConsNSU:    d.DistDFeInt.ConsNSU,
			})
		}
	case "retDistDFeInt":
		if d.RetDistDFeInt != nil && activeRootCount(d) == 1 {
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
			return e.Encode(root{
				XMLName:        xml.Name{Local: "retDistDFeInt"},
				XMLNS:          namespace,
				VersaoAttr:     firstNonEmpty(d.VersaoAttr, d.RetDistDFeInt.VersaoAttr),
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
	}

	return errors.New("marshal cte: document must contain exactly one supported root")
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse cte: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
	}

	switch rootName {
	case "CTe":
		var parsed cteSchema.TCTe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode CTe: %w", err)
		}
		doc := &Document{
			VersaoAttr: versionFromInfCte(parsed.InfCte),
			CTe:        &parsed,
			rootName:   rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "cteProc":
		var parsed cteSchema.TAnonComplexCteProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode cteProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, CTeProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retCTe":
		var parsed cteSchema.TRetCTe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retCTe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetCTe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "CTeOS":
		var parsed struct {
			VersaoAttr string                               `xml:"versao,attr"`
			InfCte     *cteOSSchema.TAnonComplexInfCte4     `xml:"infCte"`
			InfCTeSupl *cteOSSchema.TAnonComplexInfCTeSupl4 `xml:"infCTeSupl"`
			Signature  *cteOSSchema.SignatureType           `xml:"http://www.w3.org/2000/09/xmldsig# Signature"`
		}
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode CTeOS: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			CTeOS: &cteOSSchema.TCTeOS{
				VersaoAttr:  parsed.VersaoAttr,
				InfCte:      parsed.InfCte,
				InfCTeSupl:  parsed.InfCTeSupl,
				DsSignature: parsed.Signature,
			},
			rootName: rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "cteOSProc":
		var parsed cteOSSchema.TAnonComplexCteOSProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode cteOSProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, CTeOSProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retCTeOS":
		var parsed cteOSSchema.TRetCTeOS
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retCTeOS: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetCTeOS: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "CTeSimp":
		var parsed cteSimpSchema.TCTeSimp
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode CTeSimp: %w", err)
		}
		doc := &Document{VersaoAttr: versionFromInfCteSimp(parsed.InfCte), CTeSimp: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "cteSimpProc":
		var parsed cteSimpSchema.TAnonComplexCteSimpProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode cteSimpProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, CTeSimpProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retCTeSimp":
		var parsed cteSimpSchema.TRetCTeSimp
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retCTeSimp: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetCTeSimp: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "GTVe":
		var parsed gtveSchema.TGTVe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode GTVe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, GTVe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "GTVeProc":
		var parsed gtveSchema.TAnonComplexGTVeProc1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode GTVeProc: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, GTVeProc: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retGTVe":
		var parsed gtveSchema.TRetGTVe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retGTVe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetGTVe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consSitCTe":
		var parsed consSitSchema.TConsSitCTe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode consSitCTe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsSitCTe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsSitCTe":
		var parsed consSitSchema.TRetConsSitCTe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retConsSitCTe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsSitCTe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consStatServCTe":
		var parsed statusSchema.TConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode consStatServCTe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsStatServCTe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsStatServCTe":
		var parsed statusSchema.TRetConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retConsStatServCTe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServCTe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "eventoCTe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse cte: missing infEvento")
		}
		return parseEventDocument(data, rootName, tpEvento)
	case "retEventoCTe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse cte: missing infEvento")
		}
		return parseRetEventDocument(data, rootName, tpEvento)
	case "procEventoCTe":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe head: %w", err)
		}
		if tpEvento == "" {
			return nil, errors.New("parse cte: missing infEvento")
		}
		return parseProcEventDocument(data, rootName, tpEvento)
	case "distDFeInt":
		var parsed distSchema.TAnonComplexDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode distDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retDistDFeInt":
		var parsed distSchema.TAnonComplexRetDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retDistDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse cte: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse cte: unsupported root element %q", rootName)
	}
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

func validateDocument(doc *Document) error {
	count := activeRootCount(doc)
	if doc.CTe != nil {
		if err := validateInfCte(doc.CTe.InfCte); err != nil {
			return err
		}
	}
	if doc.CTeProc != nil {
		if doc.CTeProc.CTe == nil {
			return errors.New("parse cte: missing CTe")
		}
		if doc.CTeProc.ProtCTe == nil {
			return errors.New("parse cte: missing protCTe")
		}
	}
	if doc.RetCTe != nil && doc.RetCTe.CStat == "" {
		return errors.New("parse cte: missing cStat")
	}
	if doc.CTeOS != nil {
		if err := validateInfCteOS(doc.CTeOS.InfCte); err != nil {
			return err
		}
	}
	if doc.CTeOSProc != nil {
		if doc.CTeOSProc.CTeOS == nil {
			return errors.New("parse cte: missing CTeOS")
		}
		if doc.CTeOSProc.ProtCTe == nil {
			return errors.New("parse cte: missing protCTe")
		}
	}
	if doc.RetCTeOS != nil && doc.RetCTeOS.CStat == "" {
		return errors.New("parse cte: missing cStat")
	}
	if doc.CTeSimp != nil && doc.CTeSimp.InfCte == nil {
		return errors.New("parse cte: missing infCte")
	}
	if doc.CTeSimpProc != nil {
		if doc.CTeSimpProc.CTeSimp == nil {
			return errors.New("parse cte: missing CTeSimp")
		}
		if doc.CTeSimpProc.ProtCTe == nil {
			return errors.New("parse cte: missing protCTe")
		}
	}
	if doc.RetCTeSimp != nil && doc.RetCTeSimp.CStat == "" {
		return errors.New("parse cte: missing cStat")
	}
	if doc.GTVe != nil && doc.GTVe.InfCte == nil {
		return errors.New("parse cte: missing infCte")
	}
	if doc.GTVeProc != nil {
		if doc.GTVeProc.GTVe == nil {
			return errors.New("parse cte: missing GTVe")
		}
		if doc.GTVeProc.ProtCTe == nil {
			return errors.New("parse cte: missing protCTe")
		}
	}
	if doc.RetGTVe != nil && doc.RetGTVe.CStat == "" {
		return errors.New("parse cte: missing cStat")
	}
	if doc.ConsSitCTe != nil && doc.ConsSitCTe.ChCTe == "" {
		return errors.New("parse cte: missing chCTe")
	}
	if doc.RetConsSitCTe != nil && doc.RetConsSitCTe.CStat == "" {
		return errors.New("parse cte: missing cStat")
	}
	if doc.ConsStatServCTe != nil && doc.ConsStatServCTe.XServ == "" {
		return errors.New("parse cte: missing xServ")
	}
	if doc.RetConsStatServCTe != nil && doc.RetConsStatServCTe.CStat == "" {
		return errors.New("parse cte: missing cStat")
	}
	if doc.EventoCTe != nil {
		if err := validateCTeEvent(doc.EventoCTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoCancCTe != nil {
		if err := validateCTeEvent(doc.EventoCancCTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoCECTe != nil {
		if err := validateCTeEvent(doc.EventoCECTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoCancCECTe != nil {
		if err := validateCTeEvent(doc.EventoCancCECTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoEPECCTe != nil {
		if err := validateCTeEvent(doc.EventoEPECCTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoRegMultimodal != nil {
		if err := validateCTeEvent(doc.EventoRegMultimodal.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoGTV != nil {
		if err := validateCTeEvent(doc.EventoGTV.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoIECTe != nil {
		if err := validateCTeEvent(doc.EventoIECTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoCancIECTe != nil {
		if err := validateCTeEvent(doc.EventoCancIECTe.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoPrestDesacordo != nil {
		if err := validateCTeEvent(doc.EventoPrestDesacordo.InfEvento); err != nil {
			return err
		}
	}
	if doc.EventoCancPrestDesacordo != nil {
		if err := validateCTeEvent(doc.EventoCancPrestDesacordo.InfEvento); err != nil {
			return err
		}
	}
	if doc.RetEventoCTe != nil && doc.RetEventoCTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoCTe != nil && doc.ProcEventoCTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoCancCTe != nil && doc.RetEventoCancCTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoCancCTe != nil && doc.ProcEventoCancCTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoCECTe != nil && doc.RetEventoCECTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoCECTe != nil && doc.ProcEventoCECTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoCancCECTe != nil && doc.RetEventoCancCECTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoCancCECTe != nil && doc.ProcEventoCancCECTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoEPECCTe != nil && doc.RetEventoEPECCTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoEPECCTe != nil && doc.ProcEventoEPECCTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoRegMultimodal != nil && doc.RetEventoRegMultimodal.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoRegMultimodal != nil && doc.ProcEventoRegMultimodal.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoGTV != nil && doc.RetEventoGTV.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoGTV != nil && doc.ProcEventoGTV.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoIECTe != nil && doc.RetEventoIECTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoIECTe != nil && doc.ProcEventoIECTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoCancIECTe != nil && doc.RetEventoCancIECTe.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoCancIECTe != nil && doc.ProcEventoCancIECTe.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoPrestDesacordo != nil && doc.RetEventoPrestDesacordo.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoPrestDesacordo != nil && doc.ProcEventoPrestDesacordo.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.RetEventoCancPrestDesacordo != nil && doc.RetEventoCancPrestDesacordo.InfEvento == nil {
		return errors.New("parse cte: missing infEvento")
	}
	if doc.ProcEventoCancPrestDesacordo != nil && doc.ProcEventoCancPrestDesacordo.EventoCTe == nil {
		return errors.New("parse cte: missing eventoCTe")
	}
	if doc.DistDFeInt != nil {
		if doc.DistDFeInt.TpAmb == "" {
			return errors.New("parse cte: missing tpAmb")
		}
		if doc.DistDFeInt.CUFAutor == "" {
			return errors.New("parse cte: missing cUFAutor")
		}
		if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
			return errors.New("parse cte: missing dist document")
		}
		if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil {
			return errors.New("parse cte: missing dist query")
		}
	}
	if doc.RetDistDFeInt != nil {
		if doc.RetDistDFeInt.TpAmb == "" {
			return errors.New("parse cte: missing tpAmb")
		}
		if doc.RetDistDFeInt.CStat == "" {
			return errors.New("parse cte: missing cStat")
		}
		if doc.RetDistDFeInt.UltNSU == "" {
			return errors.New("parse cte: missing ultNSU")
		}
		if doc.RetDistDFeInt.MaxNSU == "" {
			return errors.New("parse cte: missing maxNSU")
		}
	}
	if count != 1 {
		return errors.New("parse cte: document must contain exactly one supported root")
	}
	return nil
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
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *ceEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelCEEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *epecEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *regMultimodalEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *gtvEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *ieEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelIEEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *prestDesacordoEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelPrestDesacordoEventSchema.TAnonComplexInfEvento1:
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
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

func marshalEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal cte: document must contain exactly one supported root")
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
	default:
		return errors.New("marshal cte: document must contain exactly one supported root")
	}
}

func marshalRetEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal cte: document must contain exactly one supported root")
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
	default:
		return errors.New("marshal cte: document must contain exactly one supported root")
	}
}

func marshalProcEventRoot(e *xml.Encoder, d *Document) error {
	if activeRootCount(d) != 1 {
		return errors.New("marshal cte: document must contain exactly one supported root")
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
	default:
		return errors.New("marshal cte: document must contain exactly one supported root")
	}
}

func encodeCTeEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	type root struct {
		XMLName     xml.Name `xml:"eventoCTe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return e.Encode(root{
		XMLName:     xml.Name{Local: "eventoCTe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeCTeRetEvent(e *xml.Encoder, versao string, infEvento any, signature any) error {
	type root struct {
		XMLName     xml.Name `xml:"retEventoCTe"`
		XMLNS       string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr  string   `xml:"versao,attr,omitempty"`
		InfEvento   any      `xml:"infEvento"`
		DsSignature any      `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
	}
	return e.Encode(root{
		XMLName:     xml.Name{Local: "retEventoCTe"},
		XMLNS:       namespace,
		VersaoAttr:  versao,
		InfEvento:   infEvento,
		DsSignature: signature,
	})
}

func encodeCTeProcEvent(e *xml.Encoder, versao string, ipTransmissor, nPortaCon, dhConexao *string, evento any, retEvento any) error {
	type root struct {
		XMLName           xml.Name `xml:"procEventoCTe"`
		XMLNS             string   `xml:"xmlns,attr,omitempty"`
		VersaoAttr        string   `xml:"versao,attr,omitempty"`
		IpTransmissorAttr *string  `xml:"ipTransmissor,attr,omitempty"`
		NPortaConAttr     *string  `xml:"nPortaCon,attr,omitempty"`
		DhConexaoAttr     *string  `xml:"dhConexao,attr,omitempty"`
		EventoCTe         any      `xml:"eventoCTe"`
		RetEventoCTe      any      `xml:"retEventoCTe"`
	}
	return e.Encode(root{
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
	switch tpEvento {
	case "110110":
		var parsed eventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe cce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoCTe: &parsed, rootName: rootName})
	case "110111":
		var parsed cancelEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe cancel: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoCancCTe: &parsed, rootName: rootName})
	case "110113":
		var parsed epecEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe epec: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoEPECCTe: &parsed, rootName: rootName})
	case "110160":
		var parsed regMultimodalEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe reg multimodal: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoRegMultimodal: &parsed, rootName: rootName})
	case "110170":
		var parsed gtvEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe gtv: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoGTV: &parsed, rootName: rootName})
	case "110180":
		var parsed ceEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe ce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoCECTe: &parsed, rootName: rootName})
	case "110181":
		var parsed cancelCEEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe cancel ce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoCancCECTe: &parsed, rootName: rootName})
	case "110190":
		var parsed ieEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe ie: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoIECTe: &parsed, rootName: rootName})
	case "110191":
		var parsed cancelIEEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe cancel ie: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoCancIECTe: &parsed, rootName: rootName})
	case "610110":
		var parsed prestDesacordoEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe prest desacordo: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoPrestDesacordo: &parsed, rootName: rootName})
	case "610111":
		var parsed cancelPrestDesacordoEventSchema.TEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode eventoCTe cancel prest desacordo: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, EventoCancPrestDesacordo: &parsed, rootName: rootName})
	default:
		return nil, fmt.Errorf("parse cte: unsupported tpEvento %q", tpEvento)
	}
}

func parseRetEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110110":
		var parsed eventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe cce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoCTe: &parsed, rootName: rootName})
	case "110111":
		var parsed cancelEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe cancel: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoCancCTe: &parsed, rootName: rootName})
	case "110113":
		var parsed epecEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe epec: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoEPECCTe: &parsed, rootName: rootName})
	case "110160":
		var parsed regMultimodalEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe reg multimodal: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoRegMultimodal: &parsed, rootName: rootName})
	case "110170":
		var parsed gtvEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe gtv: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoGTV: &parsed, rootName: rootName})
	case "110180":
		var parsed ceEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe ce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoCECTe: &parsed, rootName: rootName})
	case "110181":
		var parsed cancelCEEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe cancel ce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoCancCECTe: &parsed, rootName: rootName})
	case "110190":
		var parsed ieEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe ie: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoIECTe: &parsed, rootName: rootName})
	case "110191":
		var parsed cancelIEEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe cancel ie: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoCancIECTe: &parsed, rootName: rootName})
	case "610110":
		var parsed prestDesacordoEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe prest desacordo: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoPrestDesacordo: &parsed, rootName: rootName})
	case "610111":
		var parsed cancelPrestDesacordoEventSchema.TRetEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode retEventoCTe cancel prest desacordo: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, RetEventoCancPrestDesacordo: &parsed, rootName: rootName})
	default:
		return nil, fmt.Errorf("parse cte: unsupported tpEvento %q", tpEvento)
	}
}

func parseProcEventDocument(data []byte, rootName, tpEvento string) (*Document, error) {
	switch tpEvento {
	case "110110":
		var parsed eventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe cce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCTe: &parsed, rootName: rootName})
	case "110111":
		var parsed cancelEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe cancel: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCancCTe: &parsed, rootName: rootName})
	case "110113":
		var parsed epecEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe epec: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoEPECCTe: &parsed, rootName: rootName})
	case "110160":
		var parsed regMultimodalEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe reg multimodal: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoRegMultimodal: &parsed, rootName: rootName})
	case "110170":
		var parsed gtvEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe gtv: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoGTV: &parsed, rootName: rootName})
	case "110180":
		var parsed ceEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe ce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCECTe: &parsed, rootName: rootName})
	case "110181":
		var parsed cancelCEEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe cancel ce: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCancCECTe: &parsed, rootName: rootName})
	case "110190":
		var parsed ieEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe ie: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoIECTe: &parsed, rootName: rootName})
	case "110191":
		var parsed cancelIEEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe cancel ie: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCancIECTe: &parsed, rootName: rootName})
	case "610110":
		var parsed prestDesacordoEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe prest desacordo: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoPrestDesacordo: &parsed, rootName: rootName})
	case "610111":
		var parsed cancelPrestDesacordoEventSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse cte: decode procEventoCTe cancel prest desacordo: %w", err)
		}
		return validatedCTeDoc(&Document{VersaoAttr: parsed.VersaoAttr, ProcEventoCancPrestDesacordo: &parsed, rootName: rootName})
	default:
		return nil, fmt.Errorf("parse cte: unsupported tpEvento %q", tpEvento)
	}
}

func validatedCTeDoc(doc *Document) (*Document, error) {
	if err := validateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
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
		doc.DistDFeInt != nil,
		doc.RetDistDFeInt != nil,
	} {
		if ok {
			count++
		}
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
