package nfe

import (
	"strconv"
	"strings"

	schema "github.com/awafinance/fiscal/internal/nfe/gen/v4_0/nfe_proc"
	"github.com/awafinance/fiscal/pkg/info"
)

type Item struct {
	Number      string `json:"number,omitempty"`
	Code        string `json:"code,omitempty"`
	EAN         string `json:"ean,omitempty"`
	Description string `json:"description,omitempty"`
	NCM         string `json:"ncm,omitempty"`
	CFOP        string `json:"cfop,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Quantity    string `json:"quantity,omitempty"`
	UnitAmount  string `json:"unitAmount,omitempty"`
	Amount      string `json:"amount,omitempty"`
}

func (d *Document) GetAccessKey() string {
	if d == nil {
		return ""
	}
	if d.ProtNFe != nil && d.ProtNFe.InfProt != nil && d.ProtNFe.InfProt.ChNFe != "" {
		return d.ProtNFe.InfProt.ChNFe
	}
	if d.ResNFe != nil && d.ResNFe.ChNFe != "" {
		return d.ResNFe.ChNFe
	}
	if inf := d.infNFe(); inf != nil {
		return strings.TrimPrefix(inf.IdAttr, "NFe")
	}
	if ev := d.eventInfo().AccessKey; ev != "" {
		return ev
	}
	return ""
}

func (d *Document) GetVersion() string {
	if inf := d.infNFe(); inf != nil {
		return inf.VersaoAttr
	}
	if d != nil {
		return d.VersaoAttr
	}
	return ""
}

func (d *Document) GetEnvironment() string {
	if ide := d.ide(); ide != nil {
		return ide.TpAmb
	}
	if d != nil && d.ProtNFe != nil && d.ProtNFe.InfProt != nil {
		return d.ProtNFe.InfProt.TpAmb
	}
	return ""
}

func (d *Document) GetNumber() string {
	if ide := d.ide(); ide != nil {
		return ide.NNF
	}
	return ""
}

func (d *Document) GetSeries() string {
	if ide := d.ide(); ide != nil {
		return ide.Serie
	}
	return ""
}

func (d *Document) GetModel() string {
	if ide := d.ide(); ide != nil {
		return ide.Mod
	}
	return ""
}

func (d *Document) GetIssueDate() string {
	if ide := d.ide(); ide != nil {
		return ide.DhEmi
	}
	if d != nil && d.ResNFe != nil {
		return d.ResNFe.DhEmi
	}
	if ev := d.eventInfo().IssueDate; ev != "" {
		return ev
	}
	return ""
}

func (d *Document) GetAmount() string {
	if inf := d.infNFe(); inf != nil && inf.Total != nil && inf.Total.ICMSTot != nil {
		return inf.Total.ICMSTot.VNF
	}
	if d != nil && d.ResNFe != nil {
		return d.ResNFe.VNF
	}
	return ""
}

func (d *Document) GetIssuer() string {
	if inf := d.infNFe(); inf != nil && inf.Emit != nil {
		return inf.Emit.XNome
	}
	if d != nil && d.ResNFe != nil {
		return d.ResNFe.XNome
	}
	return ""
}

func (d *Document) GetIssuerDocument() string {
	if inf := d.infNFe(); inf != nil && inf.Emit != nil {
		return firstString(inf.Emit.CNPJ, inf.Emit.CPF)
	}
	if d != nil && d.ResNFe != nil {
		return firstString(d.ResNFe.CNPJ, d.ResNFe.CPF)
	}
	return ""
}

func (d *Document) GetRecipient() string {
	if inf := d.infNFe(); inf != nil && inf.Dest != nil {
		return stringPtrValue(inf.Dest.XNome)
	}
	return ""
}

func (d *Document) GetRecipientDocument() string {
	if inf := d.infNFe(); inf != nil && inf.Dest != nil {
		return firstString(inf.Dest.CNPJ, inf.Dest.CPF, inf.Dest.IdEstrangeiro)
	}
	return ""
}

func (d *Document) GetProtocolNumber() string {
	if d != nil && d.ProtNFe != nil && d.ProtNFe.InfProt != nil {
		return stringPtrValue(d.ProtNFe.InfProt.NProt)
	}
	if d != nil && d.ResNFe != nil {
		return d.ResNFe.NProt
	}
	if d != nil && d.ResEvento != nil {
		return d.ResEvento.NProt
	}
	return d.eventInfo().ProtocolNumber
}

func (d *Document) GetStatusCode() string {
	if d != nil && d.ProtNFe != nil && d.ProtNFe.InfProt != nil {
		return d.ProtNFe.InfProt.CStat
	}
	return d.eventInfo().StatusCode
}

func (d *Document) GetStatusReason() string {
	if d != nil && d.ProtNFe != nil && d.ProtNFe.InfProt != nil && d.ProtNFe.InfProt.XMotivo != nil {
		return string(*d.ProtNFe.InfProt.XMotivo)
	}
	return d.eventInfo().StatusReason
}

func (d *Document) IsAuthorized() bool {
	switch d.GetStatusCode() {
	case "100", "150":
		return true
	default:
		return false
	}
}

func (d *Document) GetItems() []Item {
	inf := d.infNFe()
	if inf == nil || len(inf.Det) == 0 {
		return nil
	}

	items := make([]Item, 0, len(inf.Det))
	for _, det := range inf.Det {
		if det == nil || det.Prod == nil {
			continue
		}
		items = append(items, Item{
			Number:      det.NItemAttr,
			Code:        det.Prod.CProd,
			EAN:         det.Prod.CEAN,
			Description: det.Prod.XProd,
			NCM:         det.Prod.NCM,
			CFOP:        det.Prod.CFOP,
			Unit:        det.Prod.UCom,
			Quantity:    det.Prod.QCom,
			UnitAmount:  det.Prod.VUnCom,
			Amount:      det.Prod.VProd,
		})
	}
	return items
}

func (d *Document) GetPayments() []info.Payment {
	inf := d.infNFe()
	if inf == nil || inf.Pag == nil || len(inf.Pag.DetPag) == 0 {
		return nil
	}

	payments := make([]info.Payment, 0, len(inf.Pag.DetPag))
	for _, detPag := range inf.Pag.DetPag {
		if detPag == nil {
			continue
		}
		payment := info.Payment{
			Method:        detPag.TPag,
			Amount:        detPag.VPag,
			Date:          stringPtrValue(detPag.DPag),
			PayerDocument: stringPtrValue(detPag.CNPJPag),
		}
		if detPag.Card != nil {
			payment.ReceiverDocument = stringPtrValue(detPag.Card.CNPJReceb)
		}
		payments = append(payments, payment)
	}
	return payments
}

func (d *Document) GetBilling() *info.Billing {
	inf := d.infNFe()
	if inf == nil || inf.Cobr == nil {
		return nil
	}

	billing := &info.Billing{}
	if fat := inf.Cobr.Fat; fat != nil {
		invoice := info.Invoice{
			Number:     stringPtrValue(fat.NFat),
			OrigAmount: stringPtrValue(fat.VOrig),
			Discount:   stringPtrValue(fat.VDesc),
			NetAmount:  stringPtrValue(fat.VLiq),
		}
		if invoice != (info.Invoice{}) {
			billing.Invoice = &invoice
		}
	}

	for _, dup := range inf.Cobr.Dup {
		if dup == nil {
			continue
		}
		entry := info.Duplicata{
			Number:  stringPtrValue(dup.NDup),
			DueDate: stringPtrValue(dup.DVenc),
			Amount:  dup.VDup,
		}
		if entry == (info.Duplicata{}) {
			continue
		}
		billing.Duplicates = append(billing.Duplicates, entry)
	}

	return billing
}

func (d *Document) GetDuplicatas() []info.Duplicata {
	billing := d.GetBilling()
	if billing == nil {
		return nil
	}
	return billing.Duplicates
}

func (d *Document) GetAdditionalInfo() string {
	inf := d.infNFe()
	if inf == nil || inf.InfAdic == nil {
		return ""
	}
	return joinNonEmpty("\n", stringPtrValue(inf.InfAdic.InfCpl), stringPtrValue(inf.InfAdic.InfAdFisco))
}

func (d *Document) GetAmounts() []info.Amount {
	if inf := d.infNFe(); inf != nil && inf.Total != nil {
		amounts := make([]info.Amount, 0, 24)
		amounts = append(amounts, headlineAmounts(inf.Total.ICMSTot)...)
		amounts = append(amounts, taxAmounts(inf.Total)...)
		amounts = append(amounts, retentionAmounts(inf.Total)...)
		return compactAmounts(amounts...)
	}
	if d != nil && d.ResNFe != nil {
		return compactAmounts(info.Amount{Type: "total", Value: d.ResNFe.VNF})
	}
	return nil
}

func headlineAmounts(t *schema.TAnonComplexICMSTot1) []info.Amount {
	if t == nil {
		return nil
	}
	return []info.Amount{
		{Type: "total", Value: t.VNF},
		{Type: "products", Value: t.VProd},
		{Type: "freight", Value: t.VFrete},
		{Type: "discount", Value: t.VDesc},
		{Type: "taxes", Value: stringPtrValue(t.VTotTrib)},
	}
}

func taxAmounts(total *schema.TAnonComplexTotal1) []info.Amount {
	var amounts []info.Amount
	if t := total.ICMSTot; t != nil {
		amounts = append(amounts, nonZeroAmounts(
			info.Amount{Type: "tax_icms", Value: t.VICMS},
			info.Amount{Type: "tax_icms_st", Value: t.VST},
			info.Amount{Type: "tax_icms_deson", Value: t.VICMSDeson},
			info.Amount{Type: "tax_fcp", Value: t.VFCP},
			info.Amount{Type: "tax_fcp_st", Value: t.VFCPST},
			info.Amount{Type: "tax_fcp_st_ret", Value: t.VFCPSTRet},
			info.Amount{Type: "tax_ipi", Value: t.VIPI},
			info.Amount{Type: "tax_ipi_devol", Value: t.VIPIDevol},
			info.Amount{Type: "tax_ii", Value: t.VII},
			info.Amount{Type: "tax_pis", Value: t.VPIS},
			info.Amount{Type: "tax_cofins", Value: t.VCOFINS},
		)...)
	}
	if t := total.ISSQNtot; t != nil {
		amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_iss", Value: stringPtrValue(t.VISS)})...)
	}
	if t := total.IBSCBSTot; t != nil {
		if t.GIBS != nil {
			amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_ibs", Value: t.GIBS.VIBS})...)
		}
		if t.GCBS != nil {
			amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_cbs", Value: t.GCBS.VCBS})...)
		}
	}
	if t := total.ISTot; t != nil {
		amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_is", Value: t.VIS})...)
	}
	return amounts
}

func retentionAmounts(total *schema.TAnonComplexTotal1) []info.Amount {
	var amounts []info.Amount
	if r := total.RetTrib; r != nil {
		amounts = append(amounts,
			info.Amount{Type: "retained_pis", Value: stringPtrValue(r.VRetPIS)},
			info.Amount{Type: "retained_cofins", Value: stringPtrValue(r.VRetCOFINS)},
			info.Amount{Type: "retained_csll", Value: stringPtrValue(r.VRetCSLL)},
			info.Amount{Type: "retained_irrf", Value: stringPtrValue(r.VIRRF)},
			info.Amount{Type: "retained_inss", Value: stringPtrValue(r.VRetPrev)},
		)
	}
	if t := total.ISSQNtot; t != nil {
		amounts = append(amounts, info.Amount{Type: "retained_iss", Value: stringPtrValue(t.VISSRet)})
	}
	return amounts
}

func (d *Document) GetParties() []info.Party {
	inf := d.infNFe()
	if inf == nil {
		return nil
	}
	return compactParties(
		nfeIssuerParty(inf.Emit),
		nfeRecipientParty(inf.Dest),
	)
}

func nfeIssuerParty(emit *schema.TAnonComplexEmit1) info.Party {
	if emit == nil {
		return info.Party{Role: "issuer"}
	}
	party := info.Party{
		Role:                  "issuer",
		Name:                  emit.XNome,
		Document:              firstString(emit.CNPJ, emit.CPF),
		StateRegistration:     emit.IE,
		MunicipalRegistration: stringPtrValue(emit.IM),
		SimpleNationalOption:  emit.CRT,
	}
	if emit.EnderEmit != nil {
		party.Phone = stringPtrValue(emit.EnderEmit.Fone)
		party.Address = nfeIssuerAddress(emit.EnderEmit)
	}
	return party
}

func nfeRecipientParty(dest *schema.TAnonComplexDest1) info.Party {
	if dest == nil {
		return info.Party{Role: "recipient"}
	}
	party := info.Party{
		Role:                  "recipient",
		Name:                  stringPtrValue(dest.XNome),
		Document:              firstString(dest.CNPJ, dest.CPF, dest.IdEstrangeiro),
		StateRegistration:     stringPtrValue(dest.IE),
		MunicipalRegistration: stringPtrValue(dest.IM),
		Email:                 stringPtrValue(dest.Email),
	}
	if dest.EnderDest != nil {
		party.Phone = stringPtrValue(dest.EnderDest.Fone)
		party.Address = nfeAddress(dest.EnderDest)
	}
	return party
}

func nfeIssuerAddress(end *schema.TEnderEmi) *info.Address {
	if end == nil {
		return nil
	}
	return compactAddress(&info.Address{
		Street:       end.XLgr,
		Number:       end.Nro,
		Complement:   stringPtrValue(end.XCpl),
		Neighborhood: stringPtrValue(end.XBairro),
		PostalCode:   end.CEP,
		CityCode:     end.CMun,
		CityName:     end.XMun,
		State:        end.UF,
		CountryCode:  stringPtrValue(end.CPais),
	})
}

func nfeAddress(end *schema.TEndereco) *info.Address {
	if end == nil {
		return nil
	}
	return compactAddress(&info.Address{
		Street:       end.XLgr,
		Number:       end.Nro,
		Complement:   stringPtrValue(end.XCpl),
		Neighborhood: stringPtrValue(end.XBairro),
		PostalCode:   stringPtrValue(end.CEP),
		CityCode:     end.CMun,
		CityName:     end.XMun,
		State:        end.UF,
		CountryCode:  stringPtrValue(end.CPais),
	})
}

func (d *Document) GetRelatedDocuments() []info.RelatedDocument {
	if ide := d.ide(); ide != nil && len(ide.NFref) > 0 {
		docs := make([]info.RelatedDocument, 0, len(ide.NFref))
		for _, ref := range ide.NFref {
			if ref == nil {
				continue
			}
			switch {
			case ref.RefNFe != nil:
				docs = append(docs, info.RelatedDocument{Type: "nfe", AccessKey: *ref.RefNFe})
			case ref.RefNFeSig != nil:
				docs = append(docs, info.RelatedDocument{Type: "nfe", AccessKey: *ref.RefNFeSig})
			case ref.RefCTe != nil:
				docs = append(docs, info.RelatedDocument{Type: "cte", AccessKey: *ref.RefCTe})
			}
		}
		return docs
	}
	return nil
}

func (d *Document) infNFe() *schema.TAnonComplexInfNFe1 {
	if d == nil || d.NFe == nil {
		return nil
	}
	return d.NFe.InfNFe
}

func (d *Document) ide() *schema.TAnonComplexIde1 {
	if inf := d.infNFe(); inf != nil {
		return inf.Ide
	}
	return nil
}

func firstString(values ...*string) string {
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

func compactAddress(address *info.Address) *info.Address {
	if address == nil || *address == (info.Address{}) {
		return nil
	}
	return address
}

func joinNonEmpty(sep string, values ...string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	return strings.Join(out, sep)
}
