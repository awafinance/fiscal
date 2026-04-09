package nfe

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	atorSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/ator_interessado"
	distSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/dist_dfe"
	epecSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/epec"
	cancelSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_cancel"
	cancelEntregaSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_cancel_entrega"
	insucessoCancelSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_cancel_insucesso"
	cceSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_cce"
	entregaSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_entrega"
	genericSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_generico"
	insucessoSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_insucesso"
	mdeSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v1_0/evento_mde"
	consSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v2_0/cons"
	inutSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v4_0/inutilizacao"
	schema "github.com/awa/nota-fiscal/internal/nfe/gen/v4_0/nfe_proc"
	statusSchema "github.com/awa/nota-fiscal/internal/nfe/gen/v4_0/status_servico"
)

type Document struct {
	VersaoAttr            string                                 `json:"versao,omitempty"`
	NFe                   *schema.TNFe                           `json:"NFe,omitempty"`
	ProtNFe               *schema.TProtNFe                       `json:"protNFe,omitempty"`
	EnviNFe               *schema.TEnviNFe                       `json:"enviNFe,omitempty"`
	RetEnviNFe            *schema.TRetEnviNFe                    `json:"retEnviNFe,omitempty"`
	ConsReciNFe           *schema.TConsReciNFe                   `json:"consReciNFe,omitempty"`
	RetConsReciNFe        *schema.TRetConsReciNFe                `json:"retConsReciNFe,omitempty"`
	EventoCancel          *cancelSchema.TEvento                  `json:"eventoCancel,omitempty"`
	EventoEntrega         *entregaSchema.TEvento                 `json:"eventoEntrega,omitempty"`
	EventoCancEntrega     *cancelEntregaSchema.TEvento           `json:"eventoCancEntrega,omitempty"`
	EventoCCe             *cceSchema.TEvento                     `json:"eventoCCe,omitempty"`
	EventoEPEC            *epecSchema.TEvento                    `json:"eventoEPEC,omitempty"`
	EventoAtorInteressado *atorSchema.TEvento                    `json:"eventoAtorInteressado,omitempty"`
	EventoMDE             *mdeSchema.TEvento                     `json:"eventoMDE,omitempty"`
	EventoInsucesso       *insucessoSchema.TEvento               `json:"eventoInsucesso,omitempty"`
	EventoCancInsucesso   *insucessoCancelSchema.TEvento         `json:"eventoCancInsucesso,omitempty"`
	EventoGenerico        *genericSchema.TEvento                 `json:"eventoGenerico,omitempty"`
	EnvEvento             *genericSchema.TEnvEvento              `json:"envEvento,omitempty"`
	RetEnvEvento          *genericSchema.TRetEnvEvento           `json:"retEnvEvento,omitempty"`
	ProcEventoNFe         *genericSchema.TProcEvento             `json:"procEventoNFe,omitempty"`
	ConsSitNFe            *consSchema.TConsSitNFe                `json:"consSitNFe,omitempty"`
	RetConsSitNFe         *consSchema.TRetConsSitNFe             `json:"retConsSitNFe,omitempty"`
	ConsStatServ          *statusSchema.TConsStatServ            `json:"consStatServ,omitempty"`
	RetConsStatServ       *statusSchema.TRetConsStatServ         `json:"retConsStatServ,omitempty"`
	InutNFe               *inutSchema.TInutNFe                   `json:"inutNFe,omitempty"`
	RetInutNFe            *inutSchema.TRetInutNFe                `json:"retInutNFe,omitempty"`
	ProcInutNFe           *inutSchema.TProcInutNFe               `json:"procInutNFe,omitempty"`
	DistDFeInt            *distSchema.TAnonComplexDistDFeInt1    `json:"distDFeInt,omitempty"`
	RetDistDFeInt         *distSchema.TAnonComplexRetDistDFeInt1 `json:"retDistDFeInt,omitempty"`
	ResNFe                *distSchema.TAnonComplexResNFe1        `json:"resNFe,omitempty"`
	ResEvento             *distSchema.TAnonComplexResEvento1     `json:"resEvento,omitempty"`
	rootName              string                                 `json:"-"`
}

// MarshalXML preserves the parsed root when possible.
// If protocol data is present, the document is always encoded as nfeProc.
func (d *Document) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}

	if d.rootName != "nfeProc" && d.ProtNFe == nil {
		if d.NFe == nil {
			if activeRootCount(d) > 1 {
				return errors.New("marshal nfe: document must contain exactly one supported root")
			}

			switch {
			case d.EnviNFe != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"enviNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*schema.TEnviNFe
				}{
					XMLName:  xml.Name{Local: "enviNFe"},
					XMLNS:    "http://www.portalfiscal.inf.br/nfe",
					TEnviNFe: d.EnviNFe,
				})
			case d.RetEnviNFe != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"retEnviNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*schema.TRetEnviNFe
				}{
					XMLName:     xml.Name{Local: "retEnviNFe"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					TRetEnviNFe: d.RetEnviNFe,
				})
			case d.ConsReciNFe != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"consReciNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*schema.TConsReciNFe
				}{
					XMLName:      xml.Name{Local: "consReciNFe"},
					XMLNS:        "http://www.portalfiscal.inf.br/nfe",
					TConsReciNFe: d.ConsReciNFe,
				})
			case d.RetConsReciNFe != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"retConsReciNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*schema.TRetConsReciNFe
				}{
					XMLName:         xml.Name{Local: "retConsReciNFe"},
					XMLNS:           "http://www.portalfiscal.inf.br/nfe",
					TRetConsReciNFe: d.RetConsReciNFe,
				})
			case d.EnvEvento != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"envEvento"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*genericSchema.TEnvEvento
				}{
					XMLName:    xml.Name{Local: "envEvento"},
					XMLNS:      "http://www.portalfiscal.inf.br/nfe",
					TEnvEvento: d.EnvEvento,
				})
			case d.RetEnvEvento != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"retEnvEvento"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*genericSchema.TRetEnvEvento
				}{
					XMLName:       xml.Name{Local: "retEnvEvento"},
					XMLNS:         "http://www.portalfiscal.inf.br/nfe",
					TRetEnvEvento: d.RetEnvEvento,
				})
			case d.ProcEventoNFe != nil && activeRootCount(d) == 1:
				return e.Encode(struct {
					XMLName xml.Name `xml:"procEventoNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*genericSchema.TProcEvento
				}{
					XMLName:     xml.Name{Local: "procEventoNFe"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					TProcEvento: d.ProcEventoNFe,
				})
			case d.EventoCancel != nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                             `xml:"evento"`
					XMLNS       string                               `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                               `xml:"versao,attr,omitempty"`
					InfEvento   *cancelSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *cancelSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoCancel.VersaoAttr),
					InfEvento:   d.EventoCancel.InfEvento,
					DsSignature: d.EventoCancel.DsSignature,
				})
			case d.EventoEntrega != nil && d.EventoCancel == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                              `xml:"evento"`
					XMLNS       string                                `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                                `xml:"versao,attr,omitempty"`
					InfEvento   *entregaSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *entregaSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoEntrega.VersaoAttr),
					InfEvento:   d.EventoEntrega.InfEvento,
					DsSignature: d.EventoEntrega.DsSignature,
				})
			case d.EventoCancEntrega != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                                    `xml:"evento"`
					XMLNS       string                                      `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                                      `xml:"versao,attr,omitempty"`
					InfEvento   *cancelEntregaSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *cancelEntregaSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoCancEntrega.VersaoAttr),
					InfEvento:   d.EventoCancEntrega.InfEvento,
					DsSignature: d.EventoCancEntrega.DsSignature,
				})
			case d.EventoCCe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                          `xml:"evento"`
					XMLNS       string                            `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                            `xml:"versao,attr,omitempty"`
					InfEvento   *cceSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *cceSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoCCe.VersaoAttr),
					InfEvento:   d.EventoCCe.InfEvento,
					DsSignature: d.EventoCCe.DsSignature,
				})
			case d.EventoEPEC != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                           `xml:"evento"`
					XMLNS       string                             `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                             `xml:"versao,attr,omitempty"`
					InfEvento   *epecSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *epecSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoEPEC.VersaoAttr),
					InfEvento:   d.EventoEPEC.InfEvento,
					DsSignature: d.EventoEPEC.DsSignature,
				})
			case d.EventoAtorInteressado != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                           `xml:"evento"`
					XMLNS       string                             `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                             `xml:"versao,attr,omitempty"`
					InfEvento   *atorSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *atorSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoAtorInteressado.VersaoAttr),
					InfEvento:   d.EventoAtorInteressado.InfEvento,
					DsSignature: d.EventoAtorInteressado.DsSignature,
				})
			case d.EventoMDE != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                          `xml:"evento"`
					XMLNS       string                            `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                            `xml:"versao,attr,omitempty"`
					InfEvento   *mdeSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *mdeSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoMDE.VersaoAttr),
					InfEvento:   d.EventoMDE.InfEvento,
					DsSignature: d.EventoMDE.DsSignature,
				})
			case d.EventoInsucesso != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                                `xml:"evento"`
					XMLNS       string                                  `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                                  `xml:"versao,attr,omitempty"`
					InfEvento   *insucessoSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *insucessoSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoInsucesso.VersaoAttr),
					InfEvento:   d.EventoInsucesso.InfEvento,
					DsSignature: d.EventoInsucesso.DsSignature,
				})
			case d.EventoCancInsucesso != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                                      `xml:"evento"`
					XMLNS       string                                        `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                                        `xml:"versao,attr,omitempty"`
					InfEvento   *insucessoCancelSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *insucessoCancelSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoCancInsucesso.VersaoAttr),
					InfEvento:   d.EventoCancInsucesso.InfEvento,
					DsSignature: d.EventoCancInsucesso.DsSignature,
				})
			case d.EventoGenerico != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName     xml.Name                              `xml:"evento"`
					XMLNS       string                                `xml:"xmlns,attr,omitempty"`
					VersaoAttr  string                                `xml:"versao,attr,omitempty"`
					InfEvento   *genericSchema.TAnonComplexInfEvento1 `xml:"infEvento"`
					DsSignature *genericSchema.SignatureType          `xml:"http://www.w3.org/2000/09/xmldsig# Signature,omitempty"`
				}
				return e.Encode(root{
					XMLName:     xml.Name{Local: "evento"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:  firstNonEmpty(d.VersaoAttr, d.EventoGenerico.VersaoAttr),
					InfEvento:   d.EventoGenerico.InfEvento,
					DsSignature: d.EventoGenerico.DsSignature,
				})
			case d.ConsSitNFe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName    xml.Name `xml:"consSitNFe"`
					XMLNS      string   `xml:"xmlns,attr,omitempty"`
					VersaoAttr string   `xml:"versao,attr,omitempty"`
					TpAmb      string   `xml:"tpAmb"`
					XServ      string   `xml:"xServ"`
					ChNFe      string   `xml:"chNFe"`
				}
				return e.Encode(root{
					XMLName:    xml.Name{Local: "consSitNFe"},
					XMLNS:      "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr: firstNonEmpty(d.VersaoAttr, d.ConsSitNFe.VersaoAttr),
					TpAmb:      d.ConsSitNFe.TpAmb,
					XServ:      d.ConsSitNFe.XServ,
					ChNFe:      d.ConsSitNFe.ChNFe,
				})
			case d.RetConsSitNFe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				type root struct {
					XMLName       xml.Name                  `xml:"retConsSitNFe"`
					XMLNS         string                    `xml:"xmlns,attr,omitempty"`
					VersaoAttr    string                    `xml:"versao,attr,omitempty"`
					TpAmb         string                    `xml:"tpAmb"`
					VerAplic      *consSchema.TString       `xml:"verAplic,omitempty"`
					CStat         string                    `xml:"cStat"`
					XMotivo       *consSchema.TString       `xml:"xMotivo,omitempty"`
					CUF           string                    `xml:"cUF"`
					ChNFe         string                    `xml:"chNFe"`
					ProtNFe       *consSchema.TProtNFe      `xml:"protNFe,omitempty"`
					ProcEventoNFe []*consSchema.TProcEvento `xml:"procEventoNFe,omitempty"`
				}
				return e.Encode(root{
					XMLName:       xml.Name{Local: "retConsSitNFe"},
					XMLNS:         "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr:    firstNonEmpty(d.VersaoAttr, d.RetConsSitNFe.VersaoAttr),
					TpAmb:         d.RetConsSitNFe.TpAmb,
					VerAplic:      d.RetConsSitNFe.VerAplic,
					CStat:         d.RetConsSitNFe.CStat,
					XMotivo:       d.RetConsSitNFe.XMotivo,
					CUF:           d.RetConsSitNFe.CUF,
					ChNFe:         d.RetConsSitNFe.ChNFe,
					ProtNFe:       d.RetConsSitNFe.ProtNFe,
					ProcEventoNFe: d.RetConsSitNFe.ProcEventoNFe,
				})
			case d.ConsStatServ != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"consStatServ"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*statusSchema.TConsStatServ
				}{
					XMLName:       xml.Name{Local: "consStatServ"},
					XMLNS:         "http://www.portalfiscal.inf.br/nfe",
					TConsStatServ: d.ConsStatServ,
				})
			case d.RetConsStatServ != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"retConsStatServ"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*statusSchema.TRetConsStatServ
				}{
					XMLName:          xml.Name{Local: "retConsStatServ"},
					XMLNS:            "http://www.portalfiscal.inf.br/nfe",
					TRetConsStatServ: d.RetConsStatServ,
				})
			case d.InutNFe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"inutNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*inutSchema.TInutNFe
				}{
					XMLName:  xml.Name{Local: "inutNFe"},
					XMLNS:    "http://www.portalfiscal.inf.br/nfe",
					TInutNFe: d.InutNFe,
				})
			case d.RetInutNFe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.ProcInutNFe == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"retInutNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*inutSchema.TRetInutNFe
				}{
					XMLName:     xml.Name{Local: "retInutNFe"},
					XMLNS:       "http://www.portalfiscal.inf.br/nfe",
					TRetInutNFe: d.RetInutNFe,
				})
			case d.ProcInutNFe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"procInutNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*inutSchema.TProcInutNFe
				}{
					XMLName:      xml.Name{Local: "procInutNFe"},
					XMLNS:        "http://www.portalfiscal.inf.br/nfe",
					TProcInutNFe: d.ProcInutNFe,
				})
			case d.DistDFeInt != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil && d.RetDistDFeInt == nil && d.ResNFe == nil && d.ResEvento == nil:
				type root struct {
					XMLName    xml.Name                           `xml:"distDFeInt"`
					XMLNS      string                             `xml:"xmlns,attr,omitempty"`
					VersaoAttr string                             `xml:"versao,attr,omitempty"`
					TpAmb      string                             `xml:"tpAmb"`
					CUFAutor   *string                            `xml:"cUFAutor,omitempty"`
					CNPJ       *string                            `xml:"CNPJ,omitempty"`
					CPF        *string                            `xml:"CPF,omitempty"`
					DistNSU    *distSchema.TAnonComplexDistNSU1   `xml:"distNSU,omitempty"`
					ConsNSU    *distSchema.TAnonComplexConsNSU1   `xml:"consNSU,omitempty"`
					ConsChNFe  *distSchema.TAnonComplexConsChNFe1 `xml:"consChNFe,omitempty"`
				}
				return e.Encode(root{
					XMLName:    xml.Name{Local: "distDFeInt"},
					XMLNS:      "http://www.portalfiscal.inf.br/nfe",
					VersaoAttr: firstNonEmpty(d.VersaoAttr, d.DistDFeInt.VersaoAttr),
					TpAmb:      d.DistDFeInt.TpAmb,
					CUFAutor:   d.DistDFeInt.CUFAutor,
					CNPJ:       d.DistDFeInt.CNPJ,
					CPF:        d.DistDFeInt.CPF,
					DistNSU:    d.DistDFeInt.DistNSU,
					ConsNSU:    d.DistDFeInt.ConsNSU,
					ConsChNFe:  d.DistDFeInt.ConsChNFe,
				})
			case d.RetDistDFeInt != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil && d.DistDFeInt == nil && d.ResNFe == nil && d.ResEvento == nil:
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
					XMLNS:          "http://www.portalfiscal.inf.br/nfe",
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
			case d.ResNFe != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil && d.DistDFeInt == nil && d.RetDistDFeInt == nil && d.ResEvento == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"resNFe"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*distSchema.TAnonComplexResNFe1
				}{
					XMLName:             xml.Name{Local: "resNFe"},
					XMLNS:               "http://www.portalfiscal.inf.br/nfe",
					TAnonComplexResNFe1: d.ResNFe,
				})
			case d.ResEvento != nil && d.EventoCancel == nil && d.EventoEntrega == nil && d.EventoCancEntrega == nil && d.EventoCCe == nil && d.EventoEPEC == nil && d.EventoAtorInteressado == nil && d.EventoMDE == nil && d.EventoInsucesso == nil && d.EventoCancInsucesso == nil && d.EventoGenerico == nil && d.ConsSitNFe == nil && d.RetConsSitNFe == nil && d.ConsStatServ == nil && d.RetConsStatServ == nil && d.InutNFe == nil && d.RetInutNFe == nil && d.ProcInutNFe == nil && d.DistDFeInt == nil && d.RetDistDFeInt == nil && d.ResNFe == nil:
				return e.Encode(struct {
					XMLName xml.Name `xml:"resEvento"`
					XMLNS   string   `xml:"xmlns,attr,omitempty"`
					*distSchema.TAnonComplexResEvento1
				}{
					XMLName:                xml.Name{Local: "resEvento"},
					XMLNS:                  "http://www.portalfiscal.inf.br/nfe",
					TAnonComplexResEvento1: d.ResEvento,
				})
			}
		}

		type bareNFe struct {
			XMLName xml.Name `xml:"NFe"`
			XMLNS   string   `xml:"xmlns,attr,omitempty"`
			*schema.TNFe
		}

		return e.Encode(bareNFe{
			XMLName: xml.Name{Local: "NFe"},
			XMLNS:   "http://www.portalfiscal.inf.br/nfe",
			TNFe:    d.NFe,
		})
	}

	type procNFe struct {
		XMLName    xml.Name         `xml:"nfeProc"`
		XMLNS      string           `xml:"xmlns,attr,omitempty"`
		VersaoAttr string           `xml:"versao,attr"`
		NFe        *schema.TNFe     `xml:"NFe"`
		ProtNFe    *schema.TProtNFe `xml:"protNFe"`
	}

	return e.Encode(procNFe{
		XMLName:    xml.Name{Local: "nfeProc"},
		XMLNS:      "http://www.portalfiscal.inf.br/nfe",
		VersaoAttr: d.VersaoAttr,
		NFe:        d.NFe,
		ProtNFe:    d.ProtNFe,
	})
}

func Parse(data []byte) (*Document, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("parse nfe: empty xml document")
	}

	rootName, rootErr := parseRootName(data)
	if rootErr != nil && rootName == "" {
		return nil, fmt.Errorf("parse nfe: read root: %w", rootErr)
	}

	switch rootName {
	case "nfeProc":
		var parsed struct {
			VersaoAttr string           `xml:"versao,attr"`
			NFe        *schema.TNFe     `xml:"NFe"`
			ProtNFe    *schema.TProtNFe `xml:"protNFe"`
		}
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode nfeProc: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			NFe:        parsed.NFe,
			ProtNFe:    parsed.ProtNFe,
			rootName:   "nfeProc",
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "NFe":
		var invoice schema.TNFe
		if err := xml.Unmarshal(data, &invoice); err != nil {
			return nil, fmt.Errorf("parse nfe: decode NFe: %w", err)
		}
		doc := &Document{
			VersaoAttr: versionFromNFe(&invoice),
			NFe:        &invoice,
			rootName:   "NFe",
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "enviNFe":
		var parsed schema.TEnviNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode enviNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EnviNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "retEnviNFe":
		var parsed schema.TRetEnviNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retEnviNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEnviNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "consReciNFe":
		var parsed schema.TConsReciNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode consReciNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsReciNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "retConsReciNFe":
		var parsed schema.TRetConsReciNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retConsReciNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsReciNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil

	case "evento":
		tpEvento, err := eventTypeFromXML(data)
		if err != nil {
			return nil, fmt.Errorf("parse nfe: decode evento head: %w", err)
		}
		switch tpEvento {
		case "110111":
			var event cancelSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cancel event: %w", err)
			}
			doc := &Document{
				VersaoAttr:   event.VersaoAttr,
				EventoCancel: &event,
				rootName:     rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110130":
			var event entregaSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode entrega event: %w", err)
			}
			doc := &Document{
				VersaoAttr:    event.VersaoAttr,
				EventoEntrega: &event,
				rootName:      rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110131":
			var event cancelEntregaSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cancel entrega event: %w", err)
			}
			doc := &Document{
				VersaoAttr:        event.VersaoAttr,
				EventoCancEntrega: &event,
				rootName:          rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110110":
			var event cceSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cce event: %w", err)
			}
			doc := &Document{
				VersaoAttr: event.VersaoAttr,
				EventoCCe:  &event,
				rootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110140":
			var event epecSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode epec event: %w", err)
			}
			doc := &Document{
				VersaoAttr: event.VersaoAttr,
				EventoEPEC: &event,
				rootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110150":
			var event atorSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode ator interessado event: %w", err)
			}
			doc := &Document{
				VersaoAttr:            event.VersaoAttr,
				EventoAtorInteressado: &event,
				rootName:              rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110192":
			var event insucessoSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode insucesso event: %w", err)
			}
			doc := &Document{
				VersaoAttr:      event.VersaoAttr,
				EventoInsucesso: &event,
				rootName:        rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "110193":
			var event insucessoCancelSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode cancel insucesso event: %w", err)
			}
			doc := &Document{
				VersaoAttr:          event.VersaoAttr,
				EventoCancInsucesso: &event,
				rootName:            rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		case "210200", "210210", "210220", "210240":
			var event mdeSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode mde event: %w", err)
			}
			doc := &Document{
				VersaoAttr: event.VersaoAttr,
				EventoMDE:  &event,
				rootName:   rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		default:
			var event genericSchema.TEvento
			if err := xml.Unmarshal(data, &event); err != nil {
				return nil, fmt.Errorf("parse nfe: decode generic event: %w", err)
			}
			doc := &Document{
				VersaoAttr:     event.VersaoAttr,
				EventoGenerico: &event,
				rootName:       rootName,
			}
			if err := validateDocument(doc); err != nil {
				return nil, err
			}
			return doc, nil
		}
	case "envEvento":
		var parsed genericSchema.TEnvEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode envEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, EnvEvento: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retEnvEvento":
		var parsed genericSchema.TRetEnvEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retEnvEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetEnvEvento: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "procEventoNFe":
		var parsed genericSchema.TProcEvento
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode procEventoNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcEventoNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consSitNFe":
		var parsed consSchema.TConsSitNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode consSitNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsSitNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsSitNFe":
		var parsed struct {
			VersaoAttr    string                    `xml:"versao,attr"`
			TpAmb         string                    `xml:"tpAmb"`
			VerAplic      *consSchema.TString       `xml:"verAplic"`
			CStat         string                    `xml:"cStat"`
			XMotivo       *consSchema.TString       `xml:"xMotivo"`
			CUF           string                    `xml:"cUF"`
			ChNFe         string                    `xml:"chNFe"`
			ProtNFe       *consSchema.TProtNFe      `xml:"protNFe"`
			ProcEventoNFe []*consSchema.TProcEvento `xml:"procEventoNFe"`
		}
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retConsSitNFe: %w", err)
		}
		doc := &Document{
			VersaoAttr: parsed.VersaoAttr,
			RetConsSitNFe: &consSchema.TRetConsSitNFe{
				VersaoAttr:    parsed.VersaoAttr,
				TpAmb:         parsed.TpAmb,
				VerAplic:      parsed.VerAplic,
				CStat:         parsed.CStat,
				XMotivo:       parsed.XMotivo,
				CUF:           parsed.CUF,
				ChNFe:         parsed.ChNFe,
				ProtNFe:       parsed.ProtNFe,
				ProcEventoNFe: parsed.ProcEventoNFe,
			},
			rootName: rootName,
		}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "consStatServ":
		var parsed statusSchema.TConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode consStatServ: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ConsStatServ: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retConsStatServ":
		var parsed statusSchema.TRetConsStatServ
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retConsStatServ: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetConsStatServ: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "inutNFe":
		var parsed inutSchema.TInutNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode inutNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, InutNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retInutNFe":
		var parsed inutSchema.TRetInutNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retInutNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetInutNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "procInutNFe":
		var parsed inutSchema.TProcInutNFe
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode procInutNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ProcInutNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "distDFeInt":
		var parsed distSchema.TAnonComplexDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode distDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, DistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "retDistDFeInt":
		var parsed distSchema.TAnonComplexRetDistDFeInt1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode retDistDFeInt: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, RetDistDFeInt: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "resNFe":
		var parsed distSchema.TAnonComplexResNFe1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode resNFe: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ResNFe: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	case "resEvento":
		var parsed distSchema.TAnonComplexResEvento1
		if err := xml.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("parse nfe: decode resEvento: %w", err)
		}
		doc := &Document{VersaoAttr: parsed.VersaoAttr, ResEvento: &parsed, rootName: rootName}
		if err := validateDocument(doc); err != nil {
			return nil, err
		}
		return doc, nil
	default:
		if rootErr != nil {
			return nil, fmt.Errorf("parse nfe: read root: %w", rootErr)
		}
		return nil, fmt.Errorf("parse nfe: unsupported root element %q", rootName)
	}
}

func eventTypeFromXML(data []byte) (string, error) {
	var head struct {
		InfEvento struct {
			TpEvento string `xml:"tpEvento"`
		} `xml:"infEvento"`
	}
	if err := xml.Unmarshal(data, &head); err != nil {
		return "", err
	}
	return head.InfEvento.TpEvento, nil
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

func versionFromNFe(invoice *schema.TNFe) string {
	if invoice.InfNFe == nil {
		return ""
	}

	return invoice.InfNFe.VersaoAttr
}

func validateDocument(doc *Document) error {
	count := 0

	if doc.rootName == "NFe" || doc.rootName == "nfeProc" || (doc.rootName == "" && doc.ProtNFe != nil) {
		if doc.NFe == nil {
			return errors.New("parse nfe: missing NFe")
		}
	}

	if doc.NFe != nil {
		count++
		if doc.NFe.InfNFe == nil {
			return errors.New("parse nfe: missing infNFe")
		}
		if doc.NFe.InfNFe.Ide == nil {
			return errors.New("parse nfe: missing ide")
		}
		if doc.NFe.InfNFe.Emit == nil {
			return errors.New("parse nfe: missing emit")
		}
		if doc.NFe.InfNFe.Emit.CNPJ == nil && doc.NFe.InfNFe.Emit.CPF == nil {
			return errors.New("parse nfe: missing emit document")
		}
		if len(doc.NFe.InfNFe.Det) == 0 {
			return errors.New("parse nfe: missing det")
		}
	}

	if doc.EnviNFe != nil {
		count++
		if doc.EnviNFe.IdLote == "" {
			return errors.New("parse nfe: missing idLote")
		}
		if len(doc.EnviNFe.NFe) == 0 {
			return errors.New("parse nfe: missing NFe")
		}
	}

	if doc.RetEnviNFe != nil {
		count++
		if doc.RetEnviNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetEnviNFe.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetEnviNFe.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
	}

	if doc.ConsReciNFe != nil {
		count++
		if doc.ConsReciNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.ConsReciNFe.NRec == "" {
			return errors.New("parse nfe: missing nRec")
		}
	}

	if doc.RetConsReciNFe != nil {
		count++
		if doc.RetConsReciNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetConsReciNFe.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetConsReciNFe.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
	}

	if doc.EventoCancel != nil {
		count++
		if doc.EventoCancel.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCancel.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCancel.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoEntrega != nil {
		count++
		if doc.EventoEntrega.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoEntrega.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoEntrega.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoCancEntrega != nil {
		count++
		if doc.EventoCancEntrega.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCancEntrega.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCancEntrega.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoCCe != nil {
		count++
		if doc.EventoCCe.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCCe.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCCe.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoEPEC != nil {
		count++
		if doc.EventoEPEC.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoEPEC.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoEPEC.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoAtorInteressado != nil {
		count++
		if doc.EventoAtorInteressado.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoAtorInteressado.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoAtorInteressado.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoMDE != nil {
		count++
		if doc.EventoMDE.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoMDE.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoMDE.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoInsucesso != nil {
		count++
		if doc.EventoInsucesso.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoInsucesso.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoInsucesso.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoCancInsucesso != nil {
		count++
		if doc.EventoCancInsucesso.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoCancInsucesso.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoCancInsucesso.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EventoGenerico != nil {
		count++
		if doc.EventoGenerico.InfEvento == nil {
			return errors.New("parse nfe: missing infEvento")
		}
		if doc.EventoGenerico.InfEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.EventoGenerico.InfEvento.DetEvento == nil {
			return errors.New("parse nfe: missing detEvento")
		}
	}

	if doc.EnvEvento != nil {
		count++
		if doc.EnvEvento.IdLote == "" {
			return errors.New("parse nfe: missing idLote")
		}
		if len(doc.EnvEvento.Evento) == 0 {
			return errors.New("parse nfe: missing evento")
		}
	}

	if doc.RetEnvEvento != nil {
		count++
		if doc.RetEnvEvento.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetEnvEvento.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
	}

	if doc.ProcEventoNFe != nil {
		count++
		if doc.ProcEventoNFe.Evento == nil {
			return errors.New("parse nfe: missing evento")
		}
		if doc.ProcEventoNFe.RetEvento == nil {
			return errors.New("parse nfe: missing retEvento")
		}
	}

	if doc.ConsSitNFe != nil {
		count++
		if doc.ConsSitNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.ConsSitNFe.XServ == "" {
			return errors.New("parse nfe: missing xServ")
		}
		if doc.ConsSitNFe.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
	}

	if doc.RetConsSitNFe != nil {
		count++
		if doc.RetConsSitNFe.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetConsSitNFe.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetConsSitNFe.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
		if doc.RetConsSitNFe.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
	}

	if doc.ConsStatServ != nil {
		count++
		if doc.ConsStatServ.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.ConsStatServ.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
		if doc.ConsStatServ.XServ == "" {
			return errors.New("parse nfe: missing xServ")
		}
	}

	if doc.RetConsStatServ != nil {
		count++
		if doc.RetConsStatServ.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetConsStatServ.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetConsStatServ.CUF == "" {
			return errors.New("parse nfe: missing cUF")
		}
		if doc.RetConsStatServ.DhRecbto == "" {
			return errors.New("parse nfe: missing dhRecbto")
		}
	}

	if doc.InutNFe != nil {
		count++
		if doc.InutNFe.InfInut == nil {
			return errors.New("parse nfe: missing infInut")
		}
		if doc.InutNFe.InfInut.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.InutNFe.InfInut.CNPJ == "" {
			return errors.New("parse nfe: missing CNPJ")
		}
	}

	if doc.RetInutNFe != nil {
		count++
		if doc.RetInutNFe.InfInut == nil {
			return errors.New("parse nfe: missing infInut")
		}
		if doc.RetInutNFe.InfInut.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetInutNFe.InfInut.CNPJ == nil || *doc.RetInutNFe.InfInut.CNPJ == "" {
			return errors.New("parse nfe: missing CNPJ")
		}
	}

	if doc.ProcInutNFe != nil {
		count++
		if doc.ProcInutNFe.InutNFe == nil {
			return errors.New("parse nfe: missing inutNFe")
		}
		if doc.ProcInutNFe.RetInutNFe == nil {
			return errors.New("parse nfe: missing retInutNFe")
		}
	}

	if doc.DistDFeInt != nil {
		count++
		if doc.DistDFeInt.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
			return errors.New("parse nfe: missing dist document")
		}
		if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil && doc.DistDFeInt.ConsChNFe == nil {
			return errors.New("parse nfe: missing dist query")
		}
	}

	if doc.RetDistDFeInt != nil {
		count++
		if doc.RetDistDFeInt.TpAmb == "" {
			return errors.New("parse nfe: missing tpAmb")
		}
		if doc.RetDistDFeInt.CStat == "" {
			return errors.New("parse nfe: missing cStat")
		}
		if doc.RetDistDFeInt.UltNSU == "" {
			return errors.New("parse nfe: missing ultNSU")
		}
		if doc.RetDistDFeInt.MaxNSU == "" {
			return errors.New("parse nfe: missing maxNSU")
		}
	}

	if doc.ResNFe != nil {
		count++
		if doc.ResNFe.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.ResNFe.XNome == "" {
			return errors.New("parse nfe: missing xNome")
		}
	}

	if doc.ResEvento != nil {
		count++
		if doc.ResEvento.ChNFe == "" {
			return errors.New("parse nfe: missing chNFe")
		}
		if doc.ResEvento.TpEvento == "" {
			return errors.New("parse nfe: missing tpEvento")
		}
	}

	if count != 1 {
		return errors.New("parse nfe: document must contain exactly one supported root")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func activeRootCount(doc *Document) int {
	count := 0
	if doc.NFe != nil {
		count++
	}
	if doc.EnviNFe != nil {
		count++
	}
	if doc.RetEnviNFe != nil {
		count++
	}
	if doc.ConsReciNFe != nil {
		count++
	}
	if doc.RetConsReciNFe != nil {
		count++
	}
	if doc.EventoCancel != nil {
		count++
	}
	if doc.EventoEntrega != nil {
		count++
	}
	if doc.EventoCancEntrega != nil {
		count++
	}
	if doc.EventoCCe != nil {
		count++
	}
	if doc.EventoEPEC != nil {
		count++
	}
	if doc.EventoAtorInteressado != nil {
		count++
	}
	if doc.EventoMDE != nil {
		count++
	}
	if doc.EventoInsucesso != nil {
		count++
	}
	if doc.EventoCancInsucesso != nil {
		count++
	}
	if doc.EventoGenerico != nil {
		count++
	}
	if doc.EnvEvento != nil {
		count++
	}
	if doc.RetEnvEvento != nil {
		count++
	}
	if doc.ProcEventoNFe != nil {
		count++
	}
	if doc.ConsSitNFe != nil {
		count++
	}
	if doc.RetConsSitNFe != nil {
		count++
	}
	if doc.ConsStatServ != nil {
		count++
	}
	if doc.RetConsStatServ != nil {
		count++
	}
	if doc.InutNFe != nil {
		count++
	}
	if doc.RetInutNFe != nil {
		count++
	}
	if doc.ProcInutNFe != nil {
		count++
	}
	if doc.DistDFeInt != nil {
		count++
	}
	if doc.RetDistDFeInt != nil {
		count++
	}
	if doc.ResNFe != nil {
		count++
	}
	if doc.ResEvento != nil {
		count++
	}
	return count
}
