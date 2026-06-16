package bpe

import (
	schema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/core"
	alteracaoPoltronaEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_alteracao_poltrona"
	cancelEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_cancel"
	excessoBagagemEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_excesso_bagagem"
	naoEmbEventSchema "github.com/awafinance/fiscal/internal/bpe/gen/v1_0/evento_nao_emb"
)

func tstringValue[T ~string](v *T) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

type bpeEventInfo struct {
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
func (d *Document) eventInfo() bpeEventInfo {
	if d == nil {
		return bpeEventInfo{}
	}
	sent, ret := d.eventValues()
	var e bpeEventInfo
	for _, ev := range sent {
		if tp, seq, ch, dh := bpeSentEvent(ev); tp != "" || ch != "" || dh != "" || seq != "" {
			e.EventType, e.Sequence, e.AccessKey, e.IssueDate = tp, seq, ch, dh
			break
		}
	}
	for _, ev := range ret {
		tp, seq, ch, nprot, cstat, xmot := bpeRetEvent(ev)
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
		d.EventoBPe, d.EventoCancBPe, d.EventoAlteracaoPoltrona,
		d.EventoExcessoBagagem, d.EventoNaoEmbBPe,
	}
	ret = []any{
		d.RetEventoBPe, d.RetEventoCancBPe, d.RetEventoAlteracaoPoltrona,
		d.RetEventoExcessoBagagem, d.RetEventoNaoEmbBPe,
	}
	if p := d.ProcEventoBPe; p != nil {
		sent, ret = append(sent, p.EventoBPe), append(ret, p.RetEventoBPe)
	}
	if p := d.ProcEventoCancBPe; p != nil {
		sent, ret = append(sent, p.EventoBPe), append(ret, p.RetEventoBPe)
	}
	if p := d.ProcEventoAlteracaoPoltrona; p != nil {
		sent, ret = append(sent, p.EventoBPe), append(ret, p.RetEventoBPe)
	}
	if p := d.ProcEventoExcessoBagagem; p != nil {
		sent, ret = append(sent, p.EventoBPe), append(ret, p.RetEventoBPe)
	}
	if p := d.ProcEventoNaoEmbBPe; p != nil {
		sent, ret = append(sent, p.EventoBPe), append(ret, p.RetEventoBPe)
	}
	return sent, ret
}

//nolint:gocyclo // Generated BP-e event roots require a flat type dispatch.
func bpeSentEvent(ev any) (tpEvento, sequence, accessKey, issueDate string) {
	switch v := ev.(type) {
	case *schema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChBPe, i.DhEvento
		}
	case *cancelEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChBPe, i.DhEvento
		}
	case *alteracaoPoltronaEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChBPe, i.DhEvento
		}
	case *excessoBagagemEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChBPe, i.DhEvento
		}
	case *naoEmbEventSchema.TEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return i.TpEvento, i.NSeqEvento, i.ChBPe, i.DhEvento
		}
	}
	return "", "", "", ""
}

//nolint:gocyclo // Generated BP-e return-event roots require a flat type dispatch.
func bpeRetEvent(ev any) (tpEvento, sequence, accessKey, protocolNumber, statusCode, statusReason string) {
	switch v := ev.(type) {
	case *schema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChBPe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *cancelEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChBPe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *alteracaoPoltronaEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChBPe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *excessoBagagemEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChBPe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	case *naoEmbEventSchema.TRetEvento:
		if v != nil && v.InfEvento != nil {
			i := v.InfEvento
			return stringPtrValue(i.TpEvento), stringPtrValue(i.NSeqEvento), stringPtrValue(i.ChBPe), stringPtrValue(i.NProt), i.CStat, tstringValue(i.XMotivo)
		}
	}
	return "", "", "", "", "", ""
}
