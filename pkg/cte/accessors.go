package cte

import (
	"encoding/xml"
	"strconv"
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
	if info := d.eventInfo(); info.AccessKey != "" {
		return info.AccessKey
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
	if info := d.eventInfo(); info.Environment != "" {
		return info.Environment
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
	if info := d.eventInfo(); info.IssueDate != "" {
		return info.IssueDate
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
	if name, _, ok := d.tomadorParty(); ok {
		return name
	}
	if inf := d.infCTeOS(); inf != nil && inf.Toma != nil {
		return inf.Toma.XNome
	}
	return ""
}

func (d *Document) GetRecipientDocument() string {
	if _, doc, ok := d.tomadorParty(); ok {
		return doc
	}
	if inf := d.infCTeOS(); inf != nil && inf.Toma != nil {
		return firstStringPtr(inf.Toma.CNPJ, inf.Toma.CPF)
	}
	return ""
}

// tomadorParty resolves the regular-CTe tomador (the party paying for the
// service). Ide.Toma3 carries an indicator (0=Rem, 1=Exped, 2=Receb, 3=Dest)
// pointing at one of the document's existing party blocks; Ide.Toma4 is the
// "outros" fallback block which carries its own identification.
func (d *Document) tomadorParty() (name, document string, ok bool) {
	inf := d.infCTe()
	if inf == nil || inf.Ide == nil {
		return "", "", false
	}
	if t := inf.Ide.Toma3; t != nil {
		switch t.Toma {
		case "0":
			if inf.Rem != nil {
				return inf.Rem.XNome, firstStringPtr(inf.Rem.CNPJ, inf.Rem.CPF), true
			}
		case "1":
			if inf.Exped != nil {
				return inf.Exped.XNome, firstStringPtr(inf.Exped.CNPJ, inf.Exped.CPF), true
			}
		case "2":
			if inf.Receb != nil {
				return inf.Receb.XNome, firstStringPtr(inf.Receb.CNPJ, inf.Receb.CPF), true
			}
		case "3":
			if inf.Dest != nil {
				return inf.Dest.XNome, firstStringPtr(inf.Dest.CNPJ, inf.Dest.CPF), true
			}
		}
	}
	if t := inf.Ide.Toma4; t != nil {
		return t.XNome, firstStringPtr(t.CNPJ, t.CPF), true
	}
	return "", "", false
}

func (d *Document) GetProtocolNumber() string {
	if prot := d.cteProt(); prot != nil {
		return stringPtrValue(prot.NProt)
	}
	if prot := d.cteOSProt(); prot != nil {
		return stringPtrValue(prot.NProt)
	}
	if info := d.eventInfo(); info.ProtocolNumber != "" {
		return info.ProtocolNumber
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
	if info := d.eventInfo(); info.StatusCode != "" {
		return info.StatusCode
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
	if info := d.eventInfo(); info.StatusReason != "" {
		return info.StatusReason
	}
	return ""
}

func (d *Document) IsAuthorized() bool {
	return d.GetStatusCode() == "100" || d.GetStatusCode() == "150"
}

func (d *Document) GetAmounts() []info.Amount {
	amounts := []info.Amount{{Type: "service", Value: d.GetAmount()}}
	amounts = append(amounts, d.taxAmounts()...)
	return compactAmounts(amounts...)
}

func (d *Document) taxAmounts() []info.Amount {
	if inf := d.infCTe(); inf != nil && inf.Imp != nil {
		imp := inf.Imp
		amounts := []info.Amount{{Type: "taxes", Value: stringPtrValue(imp.VTotTrib)}}
		amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_icms", Value: cteImpICMSValue(imp.ICMS)})...)
		if imp.ICMSUFFim != nil {
			amounts = append(amounts, nonZeroAmounts(
				info.Amount{Type: "tax_icms_uf_fim", Value: imp.ICMSUFFim.VICMSUFFim},
				info.Amount{Type: "tax_icms_uf_ini", Value: imp.ICMSUFFim.VICMSUFIni},
				info.Amount{Type: "tax_fcp_uf_fim", Value: imp.ICMSUFFim.VFCPUFFim},
			)...)
		}
		amounts = append(amounts, cteIBSCBSAmounts(imp.IBSCBS)...)
		return amounts
	}
	if inf := d.infCTeOS(); inf != nil && inf.Imp != nil {
		imp := inf.Imp
		amounts := []info.Amount{{Type: "taxes", Value: stringPtrValue(imp.VTotTrib)}}
		amounts = append(amounts, nonZeroAmounts(info.Amount{Type: "tax_icms", Value: cteOSImpICMSValue(imp.ICMS)})...)
		if imp.ICMSUFFim != nil {
			amounts = append(amounts, nonZeroAmounts(
				info.Amount{Type: "tax_icms_uf_fim", Value: imp.ICMSUFFim.VICMSUFFim},
				info.Amount{Type: "tax_icms_uf_ini", Value: imp.ICMSUFFim.VICMSUFIni},
				info.Amount{Type: "tax_fcp_uf_fim", Value: imp.ICMSUFFim.VFCPUFFim},
			)...)
		}
		if fed := imp.InfTribFed; fed != nil {
			amounts = append(amounts, nonZeroAmounts(
				info.Amount{Type: "tax_pis", Value: stringPtrValue(fed.VPIS)},
				info.Amount{Type: "tax_cofins", Value: stringPtrValue(fed.VCOFINS)},
				info.Amount{Type: "tax_ir", Value: stringPtrValue(fed.VIR)},
				info.Amount{Type: "tax_inss", Value: stringPtrValue(fed.VINSS)},
				info.Amount{Type: "tax_csll", Value: stringPtrValue(fed.VCSLL)},
			)...)
		}
		amounts = append(amounts, cteOSIBSCBSAmounts(imp.IBSCBS)...)
		return amounts
	}
	return nil
}

func cteImpICMSValue(t *CTeImp) string {
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
	case t.ICMS60 != nil:
		return t.ICMS60.VICMSSTRet
	}
	return ""
}

func cteOSImpICMSValue(t *CTeOSImpOS) string {
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

func cteIBSCBSAmounts(t *CTeTribCTe) []info.Amount {
	if t == nil || t.GIBSCBS == nil {
		return nil
	}
	amounts := []info.Amount{{Type: "tax_ibs", Value: t.GIBSCBS.VIBS}}
	if t.GIBSCBS.GCBS != nil {
		amounts = append(amounts, info.Amount{Type: "tax_cbs", Value: t.GIBSCBS.GCBS.VCBS})
	}
	return nonZeroAmounts(amounts...)
}

func cteOSIBSCBSAmounts(t *CTeOSTribCTe) []info.Amount {
	if t == nil || t.GIBSCBS == nil {
		return nil
	}
	amounts := []info.Amount{{Type: "tax_ibs", Value: t.GIBSCBS.VIBS}}
	if t.GIBSCBS.GCBS != nil {
		amounts = append(amounts, info.Amount{Type: "tax_cbs", Value: t.GIBSCBS.GCBS.VCBS})
	}
	return nonZeroAmounts(amounts...)
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
		if inf.Dest != nil {
			parties = append(parties, info.Party{Role: "addressee", Name: inf.Dest.XNome, Document: firstStringPtr(inf.Dest.CNPJ, inf.Dest.CPF)})
		}
	}
	return compactParties(parties...)
}

func (d *Document) GetBilling() *info.Billing {
	fat, dups := d.billingSource()
	if fat == nil && len(dups) == 0 {
		return nil
	}

	billing := &info.Billing{}
	if fat != nil {
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
	for _, dup := range dups {
		if dup.Number == "" && dup.DueDate == "" && dup.Amount == "" {
			continue
		}
		billing.Duplicates = append(billing.Duplicates, dup)
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
	if inf := d.infCTe(); inf != nil && inf.Compl != nil {
		return joinNonEmpty("\n",
			stringPtrValue(inf.Compl.XObs),
			stringPtrValue(inf.Compl.XEmi),
			stringPtrValue(inf.Compl.XCaracAd),
			stringPtrValue(inf.Compl.XCaracSer),
		)
	}
	if inf := d.infCTeOS(); inf != nil && inf.Compl != nil {
		return joinNonEmpty("\n",
			stringPtrValue(inf.Compl.XObs),
			stringPtrValue(inf.Compl.XEmi),
			stringPtrValue(inf.Compl.XCaracAd),
			stringPtrValue(inf.Compl.XCaracSer),
		)
	}
	return ""
}

// GetType returns the raw tpCTe (0 normal, 1 complemento, 3 substitution) for a
// CT-e or CT-e OS; "" for non-CT-e roots such as events.
func (d *Document) GetType() string {
	if inf := d.infCTe(); inf != nil && inf.Ide != nil {
		return inf.Ide.TpCTe
	}
	if inf := d.infCTeOS(); inf != nil && inf.Ide != nil {
		return inf.Ide.TpCTe
	}
	return ""
}

func (d *Document) GetEventType() string {
	return d.eventInfo().EventType
}

func (d *Document) GetEventSequence() string {
	return d.eventInfo().Sequence
}

func (d *Document) GetRelatedDocuments() []info.RelatedDocument {
	docs := d.cteRelatedDocuments()
	if related := d.eventInfo().RelatedDocument; !isEmptyRelatedDocument(related) {
		docs = append(docs, related)
	}
	if len(docs) == 0 {
		return nil
	}
	return docs
}

func (d *Document) cteRelatedDocuments() []info.RelatedDocument {
	var docs []info.RelatedDocument
	// Modelo 57. infCTeNorm and infCteComp are an XSD choice, so a complemento
	// carries infCteComp with no infCTeNorm — read infCteComp unconditionally.
	if inf := d.infCTe(); inf != nil {
		if norm := inf.InfCTeNorm; norm != nil {
			docs = append(docs, cteInfDocRelatedDocuments(norm.InfDoc)...)
			docs = append(docs, cteSubstitutionDocument(norm.InfCteSub)...)
		}
		docs = append(docs, cteComplementDocuments(inf.InfCteComp)...)
	}
	// Modelo 67 (CT-e OS) carries the same complement and substitute back-references.
	if inf := d.infCTeOS(); inf != nil {
		if norm := inf.InfCTeNorm; norm != nil && norm.InfCteSub != nil && norm.InfCteSub.ChCte != "" {
			docs = append(docs, info.RelatedDocument{Type: "cte", AccessKey: norm.InfCteSub.ChCte})
		}
		for _, comp := range inf.InfCteComp {
			if comp != nil && comp.ChCTe != "" {
				docs = append(docs, info.RelatedDocument{Type: "cte", AccessKey: comp.ChCTe})
			}
		}
	}
	return docs
}

func cteInfDocRelatedDocuments(infDoc *CTeAnonComplexInfDoc1) []info.RelatedDocument {
	if infDoc == nil {
		return nil
	}

	docs := make([]info.RelatedDocument, 0)
	for _, nfe := range infDoc.InfNFe {
		if nfe != nil && nfe.Chave != "" {
			docs = append(docs, info.RelatedDocument{Type: "nfe", AccessKey: nfe.Chave})
		}
	}
	for _, dce := range infDoc.InfDCe {
		if dce != nil && dce.Chave != "" {
			docs = append(docs, info.RelatedDocument{Type: "dce", AccessKey: dce.Chave})
		}
	}
	return docs
}

func cteComplementDocuments(complements []*CTeAnonComplexInfCteComp1) []info.RelatedDocument {
	var docs []info.RelatedDocument
	for _, comp := range complements {
		if comp != nil && comp.ChCTe != "" {
			docs = append(docs, info.RelatedDocument{Type: "cte", AccessKey: comp.ChCTe})
		}
	}
	return docs
}

func cteSubstitutionDocument(sub *CTeAnonComplexInfCteSub2) []info.RelatedDocument {
	if sub == nil || sub.ChCte == "" {
		return nil
	}
	return []info.RelatedDocument{{Type: "cte", AccessKey: sub.ChCte}}
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

type cteEventInfo struct {
	EventType       string
	Sequence        string
	AccessKey       string
	Environment     string
	IssueDate       string
	ProtocolNumber  string
	StatusCode      string
	StatusReason    string
	RelatedDocument info.RelatedDocument
}

func (d *Document) eventInfo() cteEventInfo {
	if d == nil {
		return cteEventInfo{}
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
	return cteEventInfo{}
}

func (d *Document) processedEventInfo() (cteEventInfo, bool) {
	if d.ProcEventoCTe != nil {
		return procEventInfo(d.ProcEventoCTe.EventoCTe, d.ProcEventoCTe.RetEventoCTe), true
	}
	if d.ProcEventoCancCTe != nil {
		return procEventInfo(d.ProcEventoCancCTe.EventoCTe, d.ProcEventoCancCTe.RetEventoCTe), true
	}
	if d.ProcEventoCECTe != nil {
		return procEventInfo(d.ProcEventoCECTe.EventoCTe, d.ProcEventoCECTe.RetEventoCTe), true
	}
	if d.ProcEventoCancCECTe != nil {
		return procEventInfo(d.ProcEventoCancCECTe.EventoCTe, d.ProcEventoCancCECTe.RetEventoCTe), true
	}
	if d.ProcEventoEPECCTe != nil {
		return procEventInfo(d.ProcEventoEPECCTe.EventoCTe, d.ProcEventoEPECCTe.RetEventoCTe), true
	}
	if d.ProcEventoRegMultimodal != nil {
		return procEventInfo(d.ProcEventoRegMultimodal.EventoCTe, d.ProcEventoRegMultimodal.RetEventoCTe), true
	}
	if d.ProcEventoGTV != nil {
		return procEventInfo(d.ProcEventoGTV.EventoCTe, d.ProcEventoGTV.RetEventoCTe), true
	}
	if d.ProcEventoIECTe != nil {
		return procEventInfo(d.ProcEventoIECTe.EventoCTe, d.ProcEventoIECTe.RetEventoCTe), true
	}
	if d.ProcEventoCancIECTe != nil {
		return procEventInfo(d.ProcEventoCancIECTe.EventoCTe, d.ProcEventoCancIECTe.RetEventoCTe), true
	}
	if d.ProcEventoPrestDesacordo != nil {
		return procEventInfo(d.ProcEventoPrestDesacordo.EventoCTe, d.ProcEventoPrestDesacordo.RetEventoCTe), true
	}
	if d.ProcEventoCancPrestDesacordo != nil {
		return procEventInfo(d.ProcEventoCancPrestDesacordo.EventoCTe, d.ProcEventoCancPrestDesacordo.RetEventoCTe), true
	}
	if d.ProcEventoGenerico != nil {
		return procEventInfo(d.ProcEventoGenerico.EventoCTe, d.ProcEventoGenerico.RetEventoCTe), true
	}
	return cteEventInfo{}, false
}

func (d *Document) standaloneSentEventInfo() (cteEventInfo, bool) {
	if d.EventoCTe != nil {
		return sentEventInfoFromRoot(d.EventoCTe), true
	}
	if d.EventoCancCTe != nil {
		return sentEventInfoFromRoot(d.EventoCancCTe), true
	}
	if d.EventoCECTe != nil {
		return sentEventInfoFromRoot(d.EventoCECTe), true
	}
	if d.EventoCancCECTe != nil {
		return sentEventInfoFromRoot(d.EventoCancCECTe), true
	}
	if d.EventoEPECCTe != nil {
		return sentEventInfoFromRoot(d.EventoEPECCTe), true
	}
	if d.EventoRegMultimodal != nil {
		return sentEventInfoFromRoot(d.EventoRegMultimodal), true
	}
	if d.EventoGTV != nil {
		return sentEventInfoFromRoot(d.EventoGTV), true
	}
	if d.EventoIECTe != nil {
		return sentEventInfoFromRoot(d.EventoIECTe), true
	}
	if d.EventoCancIECTe != nil {
		return sentEventInfoFromRoot(d.EventoCancIECTe), true
	}
	if d.EventoPrestDesacordo != nil {
		return sentEventInfoFromRoot(d.EventoPrestDesacordo), true
	}
	if d.EventoCancPrestDesacordo != nil {
		return sentEventInfoFromRoot(d.EventoCancPrestDesacordo), true
	}
	if d.EventoGenerico != nil {
		return sentEventInfoFromRoot(d.EventoGenerico), true
	}
	return cteEventInfo{}, false
}

func (d *Document) standaloneRetEventInfo() (cteEventInfo, bool) {
	if d.RetEventoCTe != nil {
		return retEventInfoFromRoot(d.RetEventoCTe), true
	}
	if d.RetEventoCancCTe != nil {
		return retEventInfoFromRoot(d.RetEventoCancCTe), true
	}
	if d.RetEventoCECTe != nil {
		return retEventInfoFromRoot(d.RetEventoCECTe), true
	}
	if d.RetEventoCancCECTe != nil {
		return retEventInfoFromRoot(d.RetEventoCancCECTe), true
	}
	if d.RetEventoEPECCTe != nil {
		return retEventInfoFromRoot(d.RetEventoEPECCTe), true
	}
	if d.RetEventoRegMultimodal != nil {
		return retEventInfoFromRoot(d.RetEventoRegMultimodal), true
	}
	if d.RetEventoGTV != nil {
		return retEventInfoFromRoot(d.RetEventoGTV), true
	}
	if d.RetEventoIECTe != nil {
		return retEventInfoFromRoot(d.RetEventoIECTe), true
	}
	if d.RetEventoCancIECTe != nil {
		return retEventInfoFromRoot(d.RetEventoCancIECTe), true
	}
	if d.RetEventoPrestDesacordo != nil {
		return retEventInfoFromRoot(d.RetEventoPrestDesacordo), true
	}
	if d.RetEventoCancPrestDesacordo != nil {
		return retEventInfoFromRoot(d.RetEventoCancPrestDesacordo), true
	}
	if d.RetEventoGenerico != nil {
		return retEventInfoFromRoot(d.RetEventoGenerico), true
	}
	return cteEventInfo{}, false
}

func procEventInfo(evento, retEvento any) cteEventInfo {
	return mergeCTeEventInfo(sentEventInfoFromRoot(evento), retEventInfoFromRoot(retEvento))
}

func sentEventInfoFromRoot(evento any) cteEventInfo {
	env, ok := readCTeSentEventEnvelope(evento)
	if !ok || !env.InfPresent {
		return cteEventInfo{}
	}
	eventInfo := sentEventInfo(env.AccessKey, env.Environment, env.IssueDate)
	eventInfo.EventType = env.EventType
	eventInfo.Sequence = env.Sequence
	if env.EventType == "310610" {
		eventInfo.RelatedDocument = mdfeDocumentFrom310610(env.DetailXML)
	}
	return eventInfo
}

func retEventInfoFromRoot(retEvento any) cteEventInfo {
	env, ok := readCTeRetEventEnvelope(retEvento)
	if !ok || !env.InfPresent {
		return cteEventInfo{}
	}
	return retEventInfo(env.EventType, env.Sequence, env.AccessKey, env.Environment, env.ProtocolNumber, env.StatusCode, env.StatusReason)
}

func sentEventInfo(chCTe, tpAmb, dhEvento string) cteEventInfo {
	return cteEventInfo{
		AccessKey:   chCTe,
		Environment: tpAmb,
		IssueDate:   dhEvento,
	}
}

func retEventInfo(tpEvento, nSeqEvento, chCTe, tpAmb, nProt, cStat, xMotivo string) cteEventInfo {
	return cteEventInfo{
		EventType:      tpEvento,
		Sequence:       nSeqEvento,
		AccessKey:      chCTe,
		Environment:    tpAmb,
		ProtocolNumber: nProt,
		StatusCode:     cStat,
		StatusReason:   xMotivo,
	}
}

func mdfeDocumentFrom310610(raw string) info.RelatedDocument {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return info.RelatedDocument{}
	}
	var payload struct {
		MDFe struct {
			ChMDFe string `xml:"chMDFe"`
		} `xml:"MDFe"`
	}
	if err := xml.Unmarshal([]byte(raw), &payload); err != nil || payload.MDFe.ChMDFe == "" {
		return info.RelatedDocument{}
	}
	return info.RelatedDocument{Type: "mdfe", AccessKey: payload.MDFe.ChMDFe}
}

func mergeCTeEventInfo(primary, fallback cteEventInfo) cteEventInfo {
	if primary.EventType == "" {
		primary.EventType = fallback.EventType
	}
	if primary.Sequence == "" {
		primary.Sequence = fallback.Sequence
	}
	if primary.AccessKey == "" {
		primary.AccessKey = fallback.AccessKey
	}
	if primary.Environment == "" {
		primary.Environment = fallback.Environment
	}
	if primary.IssueDate == "" {
		primary.IssueDate = fallback.IssueDate
	}
	if primary.ProtocolNumber == "" {
		primary.ProtocolNumber = fallback.ProtocolNumber
	}
	if primary.StatusCode == "" {
		primary.StatusCode = fallback.StatusCode
	}
	if primary.StatusReason == "" {
		primary.StatusReason = fallback.StatusReason
	}
	if isEmptyRelatedDocument(primary.RelatedDocument) {
		primary.RelatedDocument = fallback.RelatedDocument
	}
	return primary
}

func isEmptyRelatedDocument(doc info.RelatedDocument) bool {
	return doc.Type == "" && doc.AccessKey == "" && doc.Number == "" && doc.Series == ""
}

func typedStringPtrValue[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

func (d *Document) billingSource() (*billingFat, []info.Duplicata) {
	if inf := d.infCTe(); inf != nil && inf.InfCTeNorm != nil && inf.InfCTeNorm.Cobr != nil {
		cobr := inf.InfCTeNorm.Cobr
		var fat *billingFat
		if cobr.Fat != nil {
			fat = &billingFat{NFat: cobr.Fat.NFat, VOrig: cobr.Fat.VOrig, VDesc: cobr.Fat.VDesc, VLiq: cobr.Fat.VLiq}
		}
		dups := make([]info.Duplicata, 0, len(cobr.Dup))
		for _, dup := range cobr.Dup {
			if dup == nil {
				continue
			}
			dups = append(dups, info.Duplicata{Number: stringPtrValue(dup.NDup), DueDate: stringPtrValue(dup.DVenc), Amount: stringPtrValue(dup.VDup)})
		}
		return fat, dups
	}
	if inf := d.infCTeOS(); inf != nil && inf.InfCTeNorm != nil && inf.InfCTeNorm.Cobr != nil {
		cobr := inf.InfCTeNorm.Cobr
		var fat *billingFat
		if cobr.Fat != nil {
			fat = &billingFat{NFat: cobr.Fat.NFat, VOrig: cobr.Fat.VOrig, VDesc: cobr.Fat.VDesc, VLiq: cobr.Fat.VLiq}
		}
		dups := make([]info.Duplicata, 0, len(cobr.Dup))
		for _, dup := range cobr.Dup {
			if dup == nil {
				continue
			}
			dups = append(dups, info.Duplicata{Number: stringPtrValue(dup.NDup), DueDate: stringPtrValue(dup.DVenc), Amount: stringPtrValue(dup.VDup)})
		}
		return fat, dups
	}
	return nil, nil
}

type billingFat struct {
	NFat  *string
	VOrig *string
	VDesc *string
	VLiq  *string
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
