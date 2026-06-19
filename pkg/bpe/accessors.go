package bpe

import (
	"strconv"
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
	if ev := d.eventInfo().AccessKey; ev != "" {
		return ev
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
	if ev := d.eventInfo().IssueDate; ev != "" {
		return ev
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
	inf := d.infBPe()
	if inf == nil {
		return nil
	}
	amounts := make([]info.Amount, 0, 8)
	if inf.InfValorBPe != nil {
		amounts = append(amounts,
			info.Amount{Type: "ticket", Value: inf.InfValorBPe.VBP},
			info.Amount{Type: "discount", Value: inf.InfValorBPe.VDesconto},
			info.Amount{Type: "paid", Value: inf.InfValorBPe.VPgto},
			info.Amount{Type: "change", Value: inf.InfValorBPe.VTroco},
		)
	}
	if inf.Imp != nil {
		amounts = append(amounts, info.Amount{Type: "taxes", Value: stringPtrValue(inf.Imp.VTotTrib)})
		amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_icms", Value: bpeImpICMSValue(inf.Imp.ICMS)})...)
		if uf := inf.Imp.ICMSUFFim; uf != nil {
			amounts = append(amounts, nonZeroAmounts(
				info.Amount{Type: "tax_icms_uf_fim", Value: uf.VICMSUFFim},
				info.Amount{Type: "tax_icms_uf_ini", Value: uf.VICMSUFIni},
				info.Amount{Type: "tax_fcp_uf_fim", Value: uf.VFCPUFFim},
			)...)
		}
	}
	if len(amounts) == 0 {
		return nil
	}
	return compactAmounts(amounts...)
}

func bpeImpICMSValue(t *TImp) string {
	if t == nil {
		return ""
	}
	switch {
	case t.ICMS00 != nil:
		return t.ICMS00.VICMS
	case t.ICMS20 != nil:
		return t.ICMS20.VICMS
	case t.ICMS90 != nil:
		return t.ICMS90.VICMS
	}
	return ""
}

func (d *Document) GetParties() []info.Party {
	inf := d.infBPe()
	if inf == nil {
		return nil
	}
	return compactParties(
		bpeIssuerParty(inf.Emit),
		bpeBuyerParty(inf.Comp),
		bpePassengerParty(inf.InfPassagem),
	)
}

func bpeIssuerParty(emit *TAnonComplexEmit2) info.Party {
	if emit == nil {
		return info.Party{Role: "issuer"}
	}
	party := info.Party{
		Role:                  "issuer",
		Name:                  emit.XNome,
		Document:              emit.CNPJ,
		StateRegistration:     emit.IE,
		MunicipalRegistration: emit.IM,
		SimpleNationalOption:  emit.CRT,
	}
	if emit.EnderEmit != nil {
		party.Phone = stringPtrValue(emit.EnderEmit.Fone)
		party.Email = stringPtrValue(emit.EnderEmit.Email)
		party.Address = bpeIssuerAddress(emit.EnderEmit)
	}
	return party
}

func bpeBuyerParty(comp *TAnonComplexComp12) info.Party {
	if comp == nil {
		return info.Party{Role: "buyer"}
	}
	party := info.Party{
		Role:              "buyer",
		Name:              comp.XNome,
		Document:          firstStringPtr(comp.CNPJ, comp.CPF, comp.IdEstrangeiro),
		StateRegistration: stringPtrValue(comp.IE),
	}
	if comp.EnderComp != nil {
		party.Phone = stringPtrValue(comp.EnderComp.Fone)
		party.Email = stringPtrValue(comp.EnderComp.Email)
		party.Address = bpeAddress(comp.EnderComp)
	}
	return party
}

func bpePassengerParty(passagem *TAnonComplexInfPassagem1) info.Party {
	if passagem == nil || passagem.InfPassageiro == nil {
		return info.Party{Role: "passenger"}
	}
	passenger := passagem.InfPassageiro
	return info.Party{
		Role:     "passenger",
		Name:     passenger.XNome,
		Document: firstStringPtr(passenger.CPF),
		Phone:    stringPtrValue(passenger.Fone),
		Email:    stringPtrValue(passenger.Email),
	}
}

func bpeIssuerAddress(end *TEndeEmi) *info.Address {
	if end == nil {
		return nil
	}
	return compactAddress(&info.Address{
		Street:       end.XLgr,
		Number:       end.Nro,
		Complement:   stringPtrValue(end.XCpl),
		Neighborhood: end.XBairro,
		PostalCode:   stringPtrValue(end.CEP),
		CityCode:     end.CMun,
		CityName:     end.XMun,
		State:        end.UF,
	})
}

func bpeAddress(end *TEndereco) *info.Address {
	if end == nil {
		return nil
	}
	return compactAddress(&info.Address{
		Street:       end.XLgr,
		Number:       end.Nro,
		Complement:   stringPtrValue(end.XCpl),
		Neighborhood: end.XBairro,
		PostalCode:   stringPtrValue(end.CEP),
		CityCode:     end.CMun,
		CityName:     end.XMun,
		State:        end.UF,
		CountryCode:  stringPtrValue(end.CPais),
	})
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

func nonZeroAmounts(amounts ...info.Amount) []info.Amount {
	out := make([]info.Amount, 0, len(amounts))
	for _, amount := range amounts {
		if isZeroAmount(amount.Value) {
			continue
		}
		out = append(out, amount)
	}
	return out
}

func isZeroAmount(value string) bool {
	if value == "" {
		return true
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	return f == 0
}

func compactAddress(address *info.Address) *info.Address {
	if address == nil || *address == (info.Address{}) {
		return nil
	}
	return address
}

func compactParties(parties ...info.Party) []info.Party {
	out := make([]info.Party, 0, len(parties))
	for _, party := range parties {
		if partyHasData(party) {
			out = append(out, party)
		}
	}
	return out
}

func partyHasData(party info.Party) bool {
	return party.Name != "" ||
		party.Document != "" ||
		party.StateRegistration != "" ||
		party.MunicipalRegistration != "" ||
		party.Address != nil ||
		party.Phone != "" ||
		party.Email != "" ||
		party.SimpleNationalOption != "" ||
		party.SimpleNationalRegime != "" ||
		party.SpecialTaxRegime != ""
}
