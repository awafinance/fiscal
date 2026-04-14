package cte

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

func (d *Document) GetAccessKey() string {
	if prot := d.cteProt(); prot != nil && prot.ChCTe != "" {
		return prot.ChCTe
	}
	if prot := d.cteOSProt(); prot != nil && prot.ChCTe != "" {
		return prot.ChCTe
	}
	if inf := d.infCTe(); inf != nil {
		return strings.TrimPrefix(inf.IdAttr, "CTe")
	}
	if inf := d.infCTeOS(); inf != nil {
		return strings.TrimPrefix(inf.IdAttr, "CTe")
	}
	return ""
}

func (d *Document) GetVersion() string {
	if inf := d.infCTe(); inf != nil {
		return inf.VersaoAttr
	}
	if inf := d.infCTeOS(); inf != nil {
		return inf.VersaoAttr
	}
	if d != nil {
		return d.VersaoAttr
	}
	return ""
}

func (d *Document) GetEnvironment() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.TpAmb
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.TpAmb
	}
	if prot := d.cteProt(); prot != nil {
		return prot.TpAmb
	}
	if prot := d.cteOSProt(); prot != nil {
		return prot.TpAmb
	}
	return ""
}

func (d *Document) GetNumber() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.NCT
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.NCT
	}
	return ""
}

func (d *Document) GetSeries() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Serie
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.Serie
	}
	return ""
}

func (d *Document) GetModel() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Mod
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.Mod
	}
	return ""
}

func (d *Document) GetIssueDate() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.DhEmi
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.DhEmi
	}
	return ""
}

func (d *Document) GetAmount() string {
	if inf := d.infCTe(); inf != nil && inf.VPrest != nil {
		return inf.VPrest.VTPrest
	}
	if inf := d.infCTeOS(); inf != nil && inf.VPrest != nil {
		return inf.VPrest.VTPrest
	}
	return ""
}

func (d *Document) GetIssuer() string {
	if inf := d.infCTe(); inf != nil && inf.Emit != nil {
		return inf.Emit.XNome
	}
	if inf := d.infCTeOS(); inf != nil && inf.Emit != nil {
		return inf.Emit.XNome
	}
	return ""
}

func (d *Document) GetIssuerDocument() string {
	if inf := d.infCTe(); inf != nil && inf.Emit != nil {
		return firstStringPtr(inf.Emit.CNPJ, inf.Emit.CPF)
	}
	if inf := d.infCTeOS(); inf != nil && inf.Emit != nil {
		return inf.Emit.CNPJ
	}
	return ""
}

func (d *Document) GetRecipient() string {
	if inf := d.infCTe(); inf != nil && inf.Dest != nil {
		return inf.Dest.XNome
	}
	if inf := d.infCTeOS(); inf != nil && inf.Toma != nil {
		return inf.Toma.XNome
	}
	return ""
}

func (d *Document) GetRecipientDocument() string {
	if inf := d.infCTe(); inf != nil && inf.Dest != nil {
		return firstStringPtr(inf.Dest.CNPJ, inf.Dest.CPF)
	}
	if inf := d.infCTeOS(); inf != nil && inf.Toma != nil {
		return firstStringPtr(inf.Toma.CNPJ, inf.Toma.CPF)
	}
	return ""
}

func (d *Document) GetProtocolNumber() string {
	if prot := d.cteProt(); prot != nil {
		return stringPtrValue(prot.NProt)
	}
	if prot := d.cteOSProt(); prot != nil {
		return stringPtrValue(prot.NProt)
	}
	return ""
}

func (d *Document) GetStatusCode() string {
	if prot := d.cteProt(); prot != nil {
		return prot.CStat
	}
	if prot := d.cteOSProt(); prot != nil {
		return prot.CStat
	}
	return ""
}

func (d *Document) GetStatusReason() string {
	if prot := d.cteProt(); prot != nil && prot.XMotivo != nil {
		return string(*prot.XMotivo)
	}
	if prot := d.cteOSProt(); prot != nil && prot.XMotivo != nil {
		return string(*prot.XMotivo)
	}
	return ""
}

func (d *Document) IsAuthorized() bool {
	return d.GetStatusCode() == "100" || d.GetStatusCode() == "150"
}

func (d *Document) GetAmounts() []info.Amount {
	return compactAmounts(info.Amount{Type: "service", Value: d.GetAmount()})
}

func (d *Document) GetParties() []info.Party {
	parties := compactParties(
		info.Party{Role: "issuer", Name: d.GetIssuer(), Document: d.GetIssuerDocument()},
		info.Party{Role: "recipient", Name: d.GetRecipient(), Document: d.GetRecipientDocument()},
	)
	if inf := d.infCTe(); inf != nil {
		if inf.Rem != nil {
			parties = append(parties, info.Party{Role: "sender", Name: inf.Rem.XNome, Document: firstStringPtr(inf.Rem.CNPJ, inf.Rem.CPF)})
		}
		if inf.Exped != nil {
			parties = append(parties, info.Party{Role: "dispatcher", Name: inf.Exped.XNome, Document: firstStringPtr(inf.Exped.CNPJ, inf.Exped.CPF)})
		}
		if inf.Receb != nil {
			parties = append(parties, info.Party{Role: "receiver", Name: inf.Receb.XNome, Document: firstStringPtr(inf.Receb.CNPJ, inf.Receb.CPF)})
		}
	}
	return compactParties(parties...)
}

func (d *Document) GetRelatedDocuments() []info.RelatedDocument {
	inf := d.infCTe()
	if inf == nil || inf.InfCTeNorm == nil {
		return nil
	}
	docs := make([]info.RelatedDocument, 0)
	if inf.InfCTeNorm.InfDoc != nil {
		for _, nfe := range inf.InfCTeNorm.InfDoc.InfNFe {
			if nfe != nil && nfe.Chave != "" {
				docs = append(docs, info.RelatedDocument{Type: "nfe", AccessKey: nfe.Chave})
			}
		}
		for _, dce := range inf.InfCTeNorm.InfDoc.InfDCe {
			if dce != nil && dce.Chave != "" {
				docs = append(docs, info.RelatedDocument{Type: "dce", AccessKey: dce.Chave})
			}
		}
	}
	for _, comp := range inf.InfCteComp {
		if comp != nil && comp.ChCTe != "" {
			docs = append(docs, info.RelatedDocument{Type: "cte", AccessKey: comp.ChCTe})
		}
	}
	return docs
}

func (d *Document) GetModal() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Modal
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.Modal
	}
	return ""
}

func (d *Document) GetOrigin() info.Location {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return info.Location{State: inf.Ide.UFIni, CityCode: inf.Ide.CMunIni, CityName: inf.Ide.XMunIni}
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return info.Location{State: stringPtrValue(inf.Ide.UFIni), CityCode: stringPtrValue(inf.Ide.CMunIni), CityName: stringPtrValue(inf.Ide.XMunIni)}
	}
	return info.Location{}
}

func (d *Document) GetDestination() info.Location {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return info.Location{State: inf.Ide.UFFim, CityCode: inf.Ide.CMunFim, CityName: inf.Ide.XMunFim}
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return info.Location{State: stringPtrValue(inf.Ide.UFFim), CityCode: stringPtrValue(inf.Ide.CMunFim), CityName: stringPtrValue(inf.Ide.XMunFim)}
	}
	return info.Location{}
}

func (d *Document) infCTe() *CTeAnonComplexInfCte3 {
	switch {
	case d == nil:
		return nil
	case d.CTeProc != nil && d.CTeProc.CTe != nil:
		return d.CTeProc.CTe.InfCte
	case d.CTe != nil:
		return d.CTe.InfCte
	default:
		return nil
	}
}

func (d *Document) infCTeOS() *CTeOSAnonComplexInfCte4 {
	switch {
	case d == nil:
		return nil
	case d.CTeOSProc != nil && d.CTeOSProc.CTeOS != nil:
		return d.CTeOSProc.CTeOS.InfCte
	case d.CTeOS != nil:
		return d.CTeOS.InfCte
	default:
		return nil
	}
}

func (d *Document) cteProt() *CTeAnonComplexInfProt1 {
	if d != nil && d.CTeProc != nil && d.CTeProc.ProtCTe != nil {
		return d.CTeProc.ProtCTe.InfProt
	}
	return nil
}

func (d *Document) cteOSProt() *CTeOSAnonComplexInfProt2 {
	if d != nil && d.CTeOSProc != nil && d.CTeOSProc.ProtCTe != nil {
		return d.CTeOSProc.ProtCTe.InfProt
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
