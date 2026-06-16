package nfse

import (
	schema "github.com/awafinance/fiscal/internal/nfse/gen/v1_0/core"
	"github.com/awafinance/fiscal/pkg/info"
)

// infPedReg returns the registered-event payload from either a bare
// pedRegEvento (request) or an evento (registered, wrapping the pedRegEvento).
func (d *Document) infPedReg() *schema.TCInfPedReg {
	switch {
	case d == nil:
		return nil
	case d.PedRegEvento != nil:
		return d.PedRegEvento.InfPedReg
	case d.EventoNFSe != nil && d.EventoNFSe.InfEvento != nil && d.EventoNFSe.InfEvento.PedRegEvento != nil:
		return d.EventoNFSe.InfEvento.PedRegEvento.InfPedReg
	default:
		return nil
	}
}

// GetEventType returns the raw NFS-e event code (e.g. "e101101"), never a
// friendly name.
//
//nolint:gocyclo // NFS-e event type is encoded by exactly one generated payload field.
func (d *Document) GetEventType() string {
	reg := d.infPedReg()
	if reg == nil {
		return ""
	}
	switch {
	case reg.E101101 != nil:
		return "e101101"
	case reg.E105102 != nil:
		return "e105102"
	case reg.E101103 != nil:
		return "e101103"
	case reg.E105104 != nil:
		return "e105104"
	case reg.E105105 != nil:
		return "e105105"
	case reg.E202201 != nil:
		return "e202201"
	case reg.E203202 != nil:
		return "e203202"
	case reg.E204203 != nil:
		return "e204203"
	case reg.E205204 != nil:
		return "e205204"
	case reg.E202205 != nil:
		return "e202205"
	case reg.E203206 != nil:
		return "e203206"
	case reg.E204207 != nil:
		return "e204207"
	case reg.E205208 != nil:
		return "e205208"
	case reg.E305101 != nil:
		return "e305101"
	case reg.E305102 != nil:
		return "e305102"
	case reg.E305103 != nil:
		return "e305103"
	default:
		return ""
	}
}

// GetEventSequence returns nSeqEvento, present only on a registered evento; a
// bare pedRegEvento (request) returns "".
func (d *Document) GetEventSequence() string {
	if d != nil && d.EventoNFSe != nil && d.EventoNFSe.InfEvento != nil {
		return d.EventoNFSe.InfEvento.NSeqEvento
	}
	return ""
}

// GetRelatedDocuments surfaces the substituição back-reference: the substitute
// note carries the superseded key in subst/chSubstda, and the e105102 event
// carries it in chSubstituta. Both are typed with the bare kind "nfse".
func (d *Document) GetRelatedDocuments() []info.RelatedDocument {
	var docs []info.RelatedDocument
	if inf := d.infDPS(); inf != nil && inf.Subst != nil && inf.Subst.ChSubstda != "" {
		docs = append(docs, info.RelatedDocument{Type: "nfse", AccessKey: inf.Subst.ChSubstda})
	}
	if reg := d.infPedReg(); reg != nil && reg.E105102 != nil && reg.E105102.ChSubstituta != "" {
		docs = append(docs, info.RelatedDocument{Type: "nfse", AccessKey: reg.E105102.ChSubstituta})
	}
	return docs
}
