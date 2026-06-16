package mdfe

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

func (d *Document) GetAccessKey() string {
	if prot := d.prot(); prot != nil && prot.ChMDFe != "" {
		return prot.ChMDFe
	}
	if inf := d.infMDFe(); inf != nil {
		return strings.TrimPrefix(inf.IdAttr, "MDFe")
	}
	if ev := d.eventInfo().AccessKey; ev != "" {
		return ev
	}
	return ""
}

func (d *Document) GetVersion() string {
	if inf := d.infMDFe(); inf != nil {
		return inf.VersaoAttr
	}
	if d != nil {
		return d.VersaoAttr
	}
	return ""
}

func (d *Document) GetEnvironment() string {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return inf.Ide.TpAmb
	}
	if prot := d.prot(); prot != nil {
		return prot.TpAmb
	}
	return ""
}

func (d *Document) GetNumber() string {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return inf.Ide.NMDF
	}
	return ""
}

func (d *Document) GetSeries() string {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Serie
	}
	return ""
}

func (d *Document) GetModel() string {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Mod
	}
	return ""
}

func (d *Document) GetIssueDate() string {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return inf.Ide.DhEmi
	}
	if ev := d.eventInfo().IssueDate; ev != "" {
		return ev
	}
	return ""
}

func (d *Document) GetAmount() string {
	if inf := d.infMDFe(); inf != nil && inf.Tot != nil {
		return inf.Tot.VCarga
	}
	return ""
}

func (d *Document) GetIssuer() string {
	if inf := d.infMDFe(); inf != nil && inf.Emit != nil {
		return inf.Emit.XNome
	}
	return ""
}

func (d *Document) GetIssuerDocument() string {
	if inf := d.infMDFe(); inf != nil && inf.Emit != nil {
		return firstStringPtr(inf.Emit.CNPJ, inf.Emit.CPF)
	}
	return ""
}

func (d *Document) GetRecipient() string {
	return ""
}

func (d *Document) GetRecipientDocument() string {
	return ""
}

func (d *Document) GetProtocolNumber() string {
	if prot := d.prot(); prot != nil {
		return stringPtrValue(prot.NProt)
	}
	return d.eventInfo().ProtocolNumber
}

func (d *Document) GetStatusCode() string {
	if prot := d.prot(); prot != nil {
		return prot.CStat
	}
	return d.eventInfo().StatusCode
}

func (d *Document) GetStatusReason() string {
	if prot := d.prot(); prot != nil && prot.XMotivo != nil {
		return string(*prot.XMotivo)
	}
	return d.eventInfo().StatusReason
}

func (d *Document) IsAuthorized() bool {
	return d.GetStatusCode() == "100" || d.GetStatusCode() == "150"
}

func (d *Document) GetAmounts() []info.Amount {
	if inf := d.infMDFe(); inf != nil && inf.Tot != nil {
		return compactAmounts(
			info.Amount{Type: "cargo", Value: inf.Tot.VCarga},
			info.Amount{Type: "quantity", Value: inf.Tot.QCarga},
		)
	}
	return nil
}

func (d *Document) GetParties() []info.Party {
	return compactParties(info.Party{Role: "issuer", Name: d.GetIssuer(), Document: d.GetIssuerDocument()})
}

func (d *Document) GetRelatedDocuments() []info.RelatedDocument {
	inf := d.infMDFe()
	if inf == nil || inf.InfDoc == nil {
		return nil
	}

	var docs []info.RelatedDocument
	for _, city := range inf.InfDoc.InfMunDescarga {
		docs = append(docs, relatedFromUnloadCity(city)...)
	}
	return docs
}

func (d *Document) GetModal() string {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Modal
	}
	return ""
}

func (d *Document) GetOrigin() info.Location {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		location := info.Location{State: inf.Ide.UFIni}
		if len(inf.Ide.InfMunCarrega) > 0 && inf.Ide.InfMunCarrega[0] != nil {
			location.CityCode = inf.Ide.InfMunCarrega[0].CMunCarrega
			location.CityName = inf.Ide.InfMunCarrega[0].XMunCarrega
		}
		return location
	}
	return info.Location{}
}

func (d *Document) GetDestination() info.Location {
	if inf := d.infMDFe(); inf != nil && inf.Ide != nil {
		return info.Location{State: inf.Ide.UFFim}
	}
	return info.Location{}
}

func (d *Document) infMDFe() *MDFeTAnonComplexInfMDFe1 {
	switch {
	case d == nil:
		return nil
	case d.MDFeProc != nil && d.MDFeProc.MDFe != nil:
		return d.MDFeProc.MDFe.InfMDFe
	case d.MDFe != nil:
		return d.MDFe.InfMDFe
	default:
		return nil
	}
}

func (d *Document) prot() *MDFeTAnonComplexInfProt1 {
	if d != nil && d.MDFeProc != nil && d.MDFeProc.ProtMDFe != nil {
		return d.MDFeProc.ProtMDFe.InfProt
	}
	return nil
}

func firstStringPtr(values ...*string) string {
	for _, value := range values {
		if value != nil && *value != "" {
			return *value
		}
	}
	return ""
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func relatedFromUnloadCity(city *MDFeTAnonComplexInfMunDescarga1) []info.RelatedDocument {
	if city == nil {
		return nil
	}

	docs := make([]info.RelatedDocument, 0, len(city.InfNFe)+len(city.InfCTe)+len(city.InfMDFeTransp))
	docs = append(docs, relatedNFe(city.InfNFe)...)
	docs = append(docs, relatedCTe(city.InfCTe)...)
	docs = append(docs, relatedMDFe(city.InfMDFeTransp)...)
	return docs
}

func relatedNFe(values []*MDFeTAnonComplexInfNFe1) []info.RelatedDocument {
	docs := make([]info.RelatedDocument, 0, len(values))
	for _, value := range values {
		if value != nil && value.ChNFe != "" {
			docs = append(docs, info.RelatedDocument{Type: "nfe", AccessKey: value.ChNFe})
		}
	}
	return docs
}

func relatedCTe(values []*MDFeTAnonComplexInfCTe1) []info.RelatedDocument {
	docs := make([]info.RelatedDocument, 0, len(values))
	for _, value := range values {
		if value != nil && value.ChCTe != "" {
			docs = append(docs, info.RelatedDocument{Type: "cte", AccessKey: value.ChCTe})
		}
	}
	return docs
}

func relatedMDFe(values []*MDFeTAnonComplexInfMDFeTransp1) []info.RelatedDocument {
	docs := make([]info.RelatedDocument, 0, len(values))
	for _, value := range values {
		if value != nil && value.ChMDFe != "" {
			docs = append(docs, info.RelatedDocument{Type: "mdfe", AccessKey: value.ChMDFe})
		}
	}
	return docs
}

func compactAmounts(amounts ...info.Amount) []info.Amount {
	out := make([]info.Amount, 0, len(amounts))
	for _, amount := range amounts {
		if amount.Value != "" {
			out = append(out, amount)
		}
	}
	return out
}

func compactParties(parties ...info.Party) []info.Party {
	out := make([]info.Party, 0, len(parties))
	for _, party := range parties {
		if party.Name != "" || party.Document != "" {
			out = append(out, party)
		}
	}
	return out
}
