package bpe

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

func (d *Document) GetAccessKey() string {
	if info := d.eventInfo(); info.AccessKey != "" {
		return info.AccessKey
	}
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
	if info := d.eventInfo(); info.Environment != "" {
		return info.Environment
	}
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
	if info := d.eventInfo(); info.IssueDate != "" {
		return info.IssueDate
	}
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
	if info := d.eventInfo(); info.ProtocolNumber != "" {
		return info.ProtocolNumber
	}
	return ""
}

func (d *Document) GetStatusCode() string {
	if prot := d.prot(); prot != nil {
		return prot.CStat
	}
	if info := d.eventInfo(); info.StatusCode != "" {
		return info.StatusCode
	}
	return ""
}

func (d *Document) GetStatusReason() string {
	if prot := d.prot(); prot != nil && prot.XMotivo != nil {
		return string(*prot.XMotivo)
	}
	if info := d.eventInfo(); info.StatusReason != "" {
		return info.StatusReason
	}
	return ""
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

type bpeEventInfo struct {
	AccessKey      string
	Environment    string
	IssueDate      string
	ProtocolNumber string
	StatusCode     string
	StatusReason   string
}

func (d *Document) eventInfo() bpeEventInfo {
	if d == nil {
		return bpeEventInfo{}
	}
	if info, ok := d.processedEventInfo(); ok {
		return info
	}
	if info, ok := d.standaloneSentEventInfo(); ok {
		return info
	}
	if info, ok := d.standaloneRetEventInfo(); ok {
		return info
	}
	return bpeEventInfo{}
}

func (d *Document) processedEventInfo() (bpeEventInfo, bool) {
	_, root, ok := bpeEventSpecForDocument(d, bpeProcEventRoot)
	if !ok {
		return bpeEventInfo{}, false
	}
	return mergeBPeEventInfo(
		sentEventInfoFromRoot(bpeAnyField(root, "EventoBPe")),
		retEventInfoFromRoot(bpeAnyField(root, "RetEventoBPe")),
	), true
}

func (d *Document) standaloneSentEventInfo() (bpeEventInfo, bool) {
	_, root, ok := bpeEventSpecForDocument(d, bpeSentEventRoot)
	if !ok {
		return bpeEventInfo{}, false
	}
	return sentEventInfoFromRoot(root.Interface()), true
}

func (d *Document) standaloneRetEventInfo() (bpeEventInfo, bool) {
	_, root, ok := bpeEventSpecForDocument(d, bpeRetEventRoot)
	if !ok {
		return bpeEventInfo{}, false
	}
	return retEventInfoFromRoot(root.Interface()), true
}

func sentEventInfoFromRoot(evento any) bpeEventInfo {
	inf := bpeField(reflect.ValueOf(evento), "InfEvento")
	if !bpeHasValue(inf) {
		return bpeEventInfo{}
	}
	return bpeEventInfo{
		AccessKey:   bpeStringField(inf, "ChBPe"),
		Environment: bpeStringField(inf, "TpAmb"),
		IssueDate:   bpeStringField(inf, "DhEvento"),
	}
}

func retEventInfoFromRoot(retEvento any) bpeEventInfo {
	inf := bpeField(reflect.ValueOf(retEvento), "InfEvento")
	if !bpeHasValue(inf) {
		return bpeEventInfo{}
	}
	return bpeEventInfo{
		AccessKey:      bpeStringField(inf, "ChBPe"),
		Environment:    bpeStringField(inf, "TpAmb"),
		ProtocolNumber: bpeStringField(inf, "NProt"),
		StatusCode:     bpeStringField(inf, "CStat"),
		StatusReason:   bpeStringField(inf, "XMotivo"),
	}
}

func mergeBPeEventInfo(primary, fallback bpeEventInfo) bpeEventInfo {
	return bpeEventInfo{
		AccessKey:      firstNonEmpty(primary.AccessKey, fallback.AccessKey),
		Environment:    firstNonEmpty(primary.Environment, fallback.Environment),
		IssueDate:      firstNonEmpty(primary.IssueDate, fallback.IssueDate),
		ProtocolNumber: firstNonEmpty(primary.ProtocolNumber, fallback.ProtocolNumber),
		StatusCode:     firstNonEmpty(primary.StatusCode, fallback.StatusCode),
		StatusReason:   firstNonEmpty(primary.StatusReason, fallback.StatusReason),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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

func compactParties(parties ...info.Party) []info.Party {
	out := make([]info.Party, 0, len(parties))
	for _, party := range parties {
		if party.Name != "" || party.Document != "" {
			out = append(out, party)
		}
	}
	return out
}
