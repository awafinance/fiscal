package bpe

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

func (d *Document) GetAccessKey() string {
	if prot := d.prot(); prot != nil && prot.ChBPe != "" {
		return prot.ChBPe
	}
	if inf := d.infBPe(); inf != nil {
		return strings.TrimPrefix(inf.IdAttr, "BPe")
	}
	return ""
}

func (d *Document) GetVersion() string {
	if inf := d.infBPe(); inf != nil {
		return inf.VersaoAttr
	}
	if d != nil {
		return d.VersaoAttr
	}
	return ""
}

func (d *Document) GetEnvironment() string {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return inf.Ide.TpAmb
	}
	if prot := d.prot(); prot != nil {
		return prot.TpAmb
	}
	return ""
}

func (d *Document) GetNumber() string {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return inf.Ide.NBP
	}
	return ""
}

func (d *Document) GetSeries() string {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Serie
	}
	return ""
}

func (d *Document) GetModel() string {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Mod
	}
	return ""
}

func (d *Document) GetIssueDate() string {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return inf.Ide.DhEmi
	}
	return ""
}

func (d *Document) GetAmount() string {
	if inf := d.infBPe(); inf != nil && inf.InfValorBPe != nil {
		return inf.InfValorBPe.VBP
	}
	return ""
}

func (d *Document) GetIssuer() string {
	if inf := d.infBPe(); inf != nil && inf.Emit != nil {
		return inf.Emit.XNome
	}
	return ""
}

func (d *Document) GetIssuerDocument() string {
	if inf := d.infBPe(); inf != nil && inf.Emit != nil {
		return inf.Emit.CNPJ
	}
	return ""
}

func (d *Document) GetRecipient() string {
	if inf := d.infBPe(); inf != nil && inf.Comp != nil {
		return inf.Comp.XNome
	}
	return ""
}

func (d *Document) GetRecipientDocument() string {
	if inf := d.infBPe(); inf != nil && inf.Comp != nil {
		return firstStringPtr(inf.Comp.CNPJ, inf.Comp.CPF, inf.Comp.IdEstrangeiro)
	}
	return ""
}

func (d *Document) GetProtocolNumber() string {
	if prot := d.prot(); prot != nil {
		return stringPtrValue(prot.NProt)
	}
	return ""
}

func (d *Document) GetStatusCode() string {
	if prot := d.prot(); prot != nil {
		return prot.CStat
	}
	return ""
}

func (d *Document) GetStatusReason() string {
	if prot := d.prot(); prot != nil && prot.XMotivo != nil {
		return string(*prot.XMotivo)
	}
	return ""
}

func (d *Document) IsAuthorized() bool {
	return d.GetStatusCode() == "100" || d.GetStatusCode() == "150"
}

func (d *Document) GetAmounts() []info.Amount {
	if inf := d.infBPe(); inf != nil && inf.InfValorBPe != nil {
		return compactAmounts(
			info.Amount{Type: "ticket", Value: inf.InfValorBPe.VBP},
			info.Amount{Type: "discount", Value: inf.InfValorBPe.VDesconto},
			info.Amount{Type: "paid", Value: inf.InfValorBPe.VPgto},
			info.Amount{Type: "change", Value: inf.InfValorBPe.VTroco},
		)
	}
	return nil
}

func (d *Document) GetParties() []info.Party {
	return compactParties(
		info.Party{Role: "issuer", Name: d.GetIssuer(), Document: d.GetIssuerDocument()},
		info.Party{Role: "buyer", Name: d.GetRecipient(), Document: d.GetRecipientDocument()},
	)
}

func (d *Document) GetModal() string {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return inf.Ide.Modal
	}
	return ""
}

func (d *Document) GetOrigin() info.Location {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return info.Location{State: inf.Ide.UFIni, CityCode: inf.Ide.CMunIni}
	}
	return info.Location{}
}

func (d *Document) GetDestination() info.Location {
	if inf := d.infBPe(); inf != nil && inf.Ide != nil {
		return info.Location{State: inf.Ide.UFFim, CityCode: inf.Ide.CMunFim}
	}
	return info.Location{}
}

func (d *Document) infBPe() *TAnonComplexInfBPe2 {
	switch {
	case d == nil:
		return nil
	case d.BPeProc != nil && d.BPeProc.BPe != nil:
		return d.BPeProc.BPe.InfBPe
	case d.BPe != nil:
		return d.BPe.InfBPe
	default:
		return nil
	}
}

func (d *Document) prot() *TAnonComplexInfProt1 {
	switch {
	case d == nil:
		return nil
	case d.BPeProc != nil && d.BPeProc.ProtBPe != nil:
		return d.BPeProc.ProtBPe.InfProt
	case d.BPeTMProc != nil && d.BPeTMProc.ProtBPe != nil:
		return d.BPeTMProc.ProtBPe.InfProt
	default:
		return nil
	}
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
