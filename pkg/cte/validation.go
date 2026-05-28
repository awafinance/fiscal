package cte

import (
	"errors"

	cteSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte"
	cteOSSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte_os"
)

var rootValidators = []func(*Document) error{
	validateCTeRoot,
	validateCTeProcRoot,
	validateRetCTeRoot,
	validateCTeOSRoot,
	validateCTeOSProcRoot,
	validateRetCTeOSRoot,
	validateCTeSimpRoot,
	validateCTeSimpProcRoot,
	validateRetCTeSimpRoot,
	validateGTVeRoot,
	validateGTVeProcRoot,
	validateRetGTVeRoot,
	validateConsSitCTeRoot,
	validateRetConsSitCTeRoot,
	validateConsStatServCTeRoot,
	validateRetConsStatServCTeRoot,
	validateCTeEventRoots,
	validateDistDFeIntRoot,
	validateRetDistDFeIntRoot,
}

func validateDocument(doc *Document) error {
	for _, v := range rootValidators {
		if err := v(doc); err != nil {
			return err
		}
	}
	if activeRootCount(doc) != 1 {
		return errors.New("parse cte: document must contain exactly one supported root")
	}
	return nil
}

func missing(field, value string) error {
	if value == "" {
		return errors.New("parse cte: missing " + field)
	}
	return nil
}

func firstMissing(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCTeRoot(doc *Document) error {
	if doc.CTe == nil {
		return nil
	}
	return validateInfCte(doc.CTe.InfCte)
}

func validateCTeProcRoot(doc *Document) error {
	if doc.CTeProc == nil {
		return nil
	}
	if doc.CTeProc.CTe == nil {
		return errors.New("parse cte: missing CTe")
	}
	if doc.CTeProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetCTeRoot(doc *Document) error {
	if doc.RetCTe == nil {
		return nil
	}
	return missing("cStat", doc.RetCTe.CStat)
}

func validateCTeOSRoot(doc *Document) error {
	if doc.CTeOS == nil {
		return nil
	}
	return validateInfCteOS(doc.CTeOS.InfCte)
}

func validateCTeOSProcRoot(doc *Document) error {
	if doc.CTeOSProc == nil {
		return nil
	}
	if doc.CTeOSProc.CTeOS == nil {
		return errors.New("parse cte: missing CTeOS")
	}
	if doc.CTeOSProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetCTeOSRoot(doc *Document) error {
	if doc.RetCTeOS == nil {
		return nil
	}
	return missing("cStat", doc.RetCTeOS.CStat)
}

func validateCTeSimpRoot(doc *Document) error {
	if doc.CTeSimp == nil {
		return nil
	}
	if doc.CTeSimp.InfCte == nil {
		return errors.New("parse cte: missing infCte")
	}
	return nil
}

func validateCTeSimpProcRoot(doc *Document) error {
	if doc.CTeSimpProc == nil {
		return nil
	}
	if doc.CTeSimpProc.CTeSimp == nil {
		return errors.New("parse cte: missing CTeSimp")
	}
	if doc.CTeSimpProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetCTeSimpRoot(doc *Document) error {
	if doc.RetCTeSimp == nil {
		return nil
	}
	return missing("cStat", doc.RetCTeSimp.CStat)
}

func validateGTVeRoot(doc *Document) error {
	if doc.GTVe == nil {
		return nil
	}
	if doc.GTVe.InfCte == nil {
		return errors.New("parse cte: missing infCte")
	}
	return nil
}

func validateGTVeProcRoot(doc *Document) error {
	if doc.GTVeProc == nil {
		return nil
	}
	if doc.GTVeProc.GTVe == nil {
		return errors.New("parse cte: missing GTVe")
	}
	if doc.GTVeProc.ProtCTe == nil {
		return errors.New("parse cte: missing protCTe")
	}
	return nil
}

func validateRetGTVeRoot(doc *Document) error {
	if doc.RetGTVe == nil {
		return nil
	}
	return missing("cStat", doc.RetGTVe.CStat)
}

func validateConsSitCTeRoot(doc *Document) error {
	if doc.ConsSitCTe == nil {
		return nil
	}
	return missing("chCTe", doc.ConsSitCTe.ChCTe)
}

func validateRetConsSitCTeRoot(doc *Document) error {
	if doc.RetConsSitCTe == nil {
		return nil
	}
	return missing("cStat", doc.RetConsSitCTe.CStat)
}

func validateConsStatServCTeRoot(doc *Document) error {
	if doc.ConsStatServCTe == nil {
		return nil
	}
	return missing("xServ", doc.ConsStatServCTe.XServ)
}

func validateRetConsStatServCTeRoot(doc *Document) error {
	if doc.RetConsStatServCTe == nil {
		return nil
	}
	return missing("cStat", doc.RetConsStatServCTe.CStat)
}

func validateDistDFeIntRoot(doc *Document) error {
	if doc.DistDFeInt == nil {
		return nil
	}
	if err := firstMissing(
		missing("tpAmb", doc.DistDFeInt.TpAmb),
		missing("cUFAutor", doc.DistDFeInt.CUFAutor),
	); err != nil {
		return err
	}
	if doc.DistDFeInt.CNPJ == nil && doc.DistDFeInt.CPF == nil {
		return errors.New("parse cte: missing dist document")
	}
	if doc.DistDFeInt.DistNSU == nil && doc.DistDFeInt.ConsNSU == nil {
		return errors.New("parse cte: missing dist query")
	}
	return nil
}

func validateRetDistDFeIntRoot(doc *Document) error {
	if doc.RetDistDFeInt == nil {
		return nil
	}
	return firstMissing(
		missing("tpAmb", doc.RetDistDFeInt.TpAmb),
		missing("cStat", doc.RetDistDFeInt.CStat),
		missing("ultNSU", doc.RetDistDFeInt.UltNSU),
		missing("maxNSU", doc.RetDistDFeInt.MaxNSU),
	)
}

func validateCTeEventRoots(doc *Document) error {
	for i := range cteEventSpecs {
		spec := &cteEventSpecs[i]
		if err := validateCTeEventField(doc, spec, cteSentEventRoot, validateCTeSentEventRoot); err != nil {
			return err
		}
		if err := validateCTeEventField(doc, spec, cteRetEventRoot, validateCTeRetEventRoot); err != nil {
			return err
		}
		if err := validateCTeProcEventField(doc, spec); err != nil {
			return err
		}
	}
	return nil
}

func validateCTeEventField(doc *Document, spec *cteEventSpec, kind cteEventRootKind, validate func(any) error) error {
	root := cteDocumentEventField(doc, spec, kind)
	if !cteHasValue(root) {
		return nil
	}
	return validate(root.Interface())
}

func validateCTeProcEventField(doc *Document, spec *cteEventSpec) error {
	root := cteDocumentEventField(doc, spec, cteProcEventRoot)
	if !cteHasValue(root) {
		return nil
	}
	return validateCTeProcessedEvent(cteAnyField(root, "EventoCTe"), cteAnyField(root, "RetEventoCTe"))
}

func validateCTeProcessedEvent(evento, retEvento any) error {
	if err := validateCTeSentEventRoot(evento); err != nil {
		return err
	}
	return validateCTeRetEventRoot(retEvento)
}

func validateCTeSentEventRoot(evento any) error {
	env, ok := readCTeSentEventEnvelope(evento)
	if !ok || !env.RootPresent {
		return errors.New("parse cte: missing eventoCTe")
	}
	return validateInfEvento(env.InfPresent, env.AccessKey, env.HasDetail)
}

func validateCTeRetEventRoot(retEvento any) error {
	env, ok := readCTeRetEventEnvelope(retEvento)
	if !ok || !env.RootPresent {
		return errors.New("parse cte: missing retEventoCTe")
	}
	return validateRetInfEvento(env.InfPresent, env.Environment, env.StatusCode)
}

func validateInfCte(inf *cteSchema.TAnonComplexInfCte3) error {
	if inf == nil {
		return errors.New("parse cte: missing infCte")
	}
	if inf.Ide == nil {
		return errors.New("parse cte: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse cte: missing emit")
	}
	if inf.Emit.CNPJ == nil && inf.Emit.CPF == nil {
		return errors.New("parse cte: missing emit document")
	}
	return nil
}

func validateInfCteOS(inf *cteOSSchema.TAnonComplexInfCte4) error {
	if inf == nil {
		return errors.New("parse cte: missing infCte")
	}
	if inf.Ide == nil {
		return errors.New("parse cte: missing ide")
	}
	if inf.Emit == nil {
		return errors.New("parse cte: missing emit")
	}
	if inf.Emit.CNPJ == "" {
		return errors.New("parse cte: missing emit document")
	}
	return nil
}

func validateInfEvento(ok bool, chCTe string, hasDet bool) error {
	if !ok {
		return errors.New("parse cte: missing infEvento")
	}
	if chCTe == "" {
		return errors.New("parse cte: missing chCTe")
	}
	if !hasDet {
		return errors.New("parse cte: missing detEvento")
	}
	return nil
}

func validateRetInfEvento(ok bool, tpAmb, cStat string) error {
	if !ok {
		return errors.New("parse cte: missing infEvento")
	}
	return firstMissing(
		missing("tpAmb", tpAmb),
		missing("cStat", cStat),
	)
}

func activeRootCount(doc *Document) int {
	count := 0
	for _, ok := range []bool{
		doc.CTe != nil,
		doc.CTeProc != nil,
		doc.RetCTe != nil,
		doc.CTeOS != nil,
		doc.CTeOSProc != nil,
		doc.RetCTeOS != nil,
		doc.CTeSimp != nil,
		doc.CTeSimpProc != nil,
		doc.RetCTeSimp != nil,
		doc.GTVe != nil,
		doc.GTVeProc != nil,
		doc.RetGTVe != nil,
		doc.ConsSitCTe != nil,
		doc.RetConsSitCTe != nil,
		doc.ConsStatServCTe != nil,
		doc.RetConsStatServCTe != nil,
		doc.DistDFeInt != nil,
		doc.RetDistDFeInt != nil,
	} {
		if ok {
			count++
		}
	}
	count += activeCTeEventRootCount(doc)
	return count
}
