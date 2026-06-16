package mdfe

import (
	eventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento"
	alteracaoPagtoServEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_alteracao_pagto_serv"
	cancelEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_cancel"
	confirmaServEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_confirma_serv"
	encEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_enc"
	incCondutorEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_inc_condutor"
	inclusaoDFeEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_inclusao_dfe"
	pagtoOperEventSchema "github.com/awafinance/fiscal/internal/mdfe/gen/v3_0/evento_pagto_oper"
)

func tstringValue[T ~string](v *T) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

type mdfeEventInfo struct {
	EventType      string
	Sequence       string
	AccessKey      string
	IssueDate      string
	ProtocolNumber string
	StatusCode     string
	StatusReason   string
}

func (d *Document) GetEventType() string {
	return d.eventInfo().EventType
}

func (d *Document) GetEventSequence() string {
	return d.eventInfo().Sequence
}

// eventInfo normalizes the single populated event/retEvento/procEvento root into
// a shared shape. The generated event types differ per sub-package but carry the
// same infEvento fields. A parsed Document holds exactly one event root.
func (d *Document) eventInfo() mdfeEventInfo {
	if d == nil {
		return mdfeEventInfo{}
	}
	sent, ret := d.eventValues()
	var e mdfeEventInfo
	for _, ev := range sent {
		if tp, seq, ch, dh := mdfeSentEvent(ev); tp != "" || ch != "" || dh != "" || seq != "" {
			e.EventType, e.Sequence, e.AccessKey, e.IssueDate = tp, seq, ch, dh
			break
		}
	}
	for _, ev := range ret {
		tp, seq, ch, nprot, cstat, xmot := mdfeRetEvent(ev)
		if cstat == "" && nprot == "" && ch == "" {
			continue
		}
		if e.EventType == "" {
			e.EventType = tp
		}
		if e.Sequence == "" {
			e.Sequence = seq
		}
		if e.AccessKey == "" {
			e.AccessKey = ch
		}
		e.ProtocolNumber, e.StatusCode, e.StatusReason = nprot, cstat, xmot
		break
	}
	return e
}

func (d *Document) eventValues() (sent []any, ret []any) {
	sent = []any{
		d.EventoMDFe, d.EventoCancMDFe, d.EventoEncMDFe, d.EventoIncCondutorMDFe,
		d.EventoInclusaoDFeMDFe, d.EventoPagtoOperMDFe, d.EventoAlteracaoPagtoServMDFe,
		d.EventoConfirmaServMDFe,
	}
	ret = []any{
		d.RetEventoMDFe, d.RetEventoCancMDFe, d.RetEventoEncMDFe, d.RetEventoIncCondutorMDFe,
		d.RetEventoInclusaoDFeMDFe, d.RetEventoPagtoOperMDFe, d.RetEventoAlteracaoPagtoServMDFe,
		d.RetEventoConfirmaServMDFe,
	}
	if p := d.ProcEventoMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoCancMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoEncMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoIncCondutorMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoInclusaoDFeMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoPagtoOperMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoAlteracaoPagtoServMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	if p := d.ProcEventoConfirmaServMDFe; p != nil {
		sent, ret = append(sent, p.EventoMDFe), append(ret, p.RetEventoMDFe)
	}
	return sent, ret
}

//nolint:gocognit,gocyclo // Generated MDF-e event roots require a flat type dispatch.
func mdfeSentEvent(ev any) (tpEvento, sequence, accessKey, issueDate string) {
	switch v := ev.(type) {
	case *eventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *cancelEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *encEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *incCondutorEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *inclusaoDFeEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *pagtoOperEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *alteracaoPagtoServEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	case *confirmaServEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChMDFe, i.DhEvento
		}
	}
	return "", "", "", ""
}

//nolint:gocognit,gocyclo // Generated MDF-e return-event roots require a flat type dispatch.
func mdfeRetEvent(ev any) (tpEvento, sequence, accessKey, protocolNumber, statusCode, statusReason string) {
	switch v := ev.(type) {
	case *eventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *cancelEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *encEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *incCondutorEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *inclusaoDFeEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *pagtoOperEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *alteracaoPagtoServEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *confirmaServEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChMDFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	}
	return "", "", "", "", "", ""
}
