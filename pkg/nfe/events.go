package nfe

import (
	atorSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/ator_interessado"
	distSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/dist_dfe"
	epecSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/epec"
	cancelSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel"
	cancelEntregaSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel_entrega"
	insucessoCancelSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cancel_insucesso"
	cceSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_cce"
	entregaSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_entrega"
	genericSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_generico"
	insucessoSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_insucesso"
	mdeSchema "github.com/awafinance/fiscal/internal/nfe/gen/v1_0/evento_mde"
)

func tstringValue[T ~string](v *T) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

type nfeEventInfo struct {
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

// eventInfo normalizes the populated event root into a shared shape. NF-e events
// are dispatched into concrete typed fields by tpEvento; ret/proc data arrives
// through the generic envelopes. A parsed Document holds exactly one event root.
func (d *Document) eventInfo() nfeEventInfo {
	if d == nil {
		return nfeEventInfo{}
	}
	sent, ret := d.eventValues()
	var e nfeEventInfo
	for _, ev := range sent {
		if tp, seq, ch, dh := nfeSentEvent(ev); tp != "" || ch != "" || dh != "" || seq != "" {
			e.EventType, e.Sequence, e.AccessKey, e.IssueDate = tp, seq, ch, dh
			break
		}
	}
	for _, ev := range ret {
		tp, seq, ch, nprot, cstat, xmot := nfeRetEvent(ev)
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
		d.EventoCancel, d.EventoEntrega, d.EventoCancEntrega, d.EventoCCe, d.EventoEPEC,
		d.EventoAtorInteressado, d.EventoMDE, d.EventoInsucesso, d.EventoCancInsucesso,
		d.EventoGenerico, d.ResEvento,
	}
	if d.EnvEvento != nil {
		for _, ev := range d.EnvEvento.Evento {
			sent = append(sent, ev)
		}
	}
	if d.RetEnvEvento != nil {
		for _, ev := range d.RetEnvEvento.RetEvento {
			ret = append(ret, ev)
		}
	}
	if p := d.ProcEventoNFe; p != nil {
		sent, ret = append(sent, p.Evento), append(ret, p.RetEvento)
	}
	return sent, ret
}

//nolint:gocognit,gocyclo // Generated NF-e event roots require a flat type dispatch.
func nfeSentEvent(ev any) (tpEvento, sequence, accessKey, issueDate string) {
	switch v := ev.(type) {
	case *cancelSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *entregaSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *cancelEntregaSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *cceSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *epecSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *atorSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *mdeSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *insucessoSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *insucessoCancelSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *genericSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChNFe, i.DhEvento
		}
	case *distSchema.TAnonComplexResEvento1:
		if v != nil {
			return v.TpEvento, v.NSeqEvento, v.ChNFe, v.DhEvento
		}
	}
	return "", "", "", ""
}

func nfeRetEvent(ev any) (tpEvento, sequence, accessKey, protocolNumber, statusCode, statusReason string) {
	if v, ok := ev.(*genericSchema.TRetEvento); ok && v != nil && v.InfEvento != nil {
		i := v.InfEvento
		return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChNFe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
	}
	return "", "", "", "", "", ""
}
