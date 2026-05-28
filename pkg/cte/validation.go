package cte

import (
	"errors"

	cteSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte"
	cteOSSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/cte_os"
	cancelEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel"
	cancelCEEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel_ce"
	cancelIEEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel_ie"
	cancelPrestDesacordoEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cancel_prest_desacordo"
	eventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_cce"
	ceEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_ce"
	epecEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_epec"
	genericEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_generico"
	gtvEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_gtv"
	ieEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_ie"
	prestDesacordoEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_prest_desacordo"
	regMultimodalEventSchema "github.com/awafinance/fiscal/internal/cte/gen/v4_0/evento_reg_multimodal"
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
	validateEventoCTeRoot,
	validateEventoCancCTeRoot,
	validateEventoCECTeRoot,
	validateEventoCancCECTeRoot,
	validateEventoEPECCTeRoot,
	validateEventoRegMultimodalRoot,
	validateEventoGTVRoot,
	validateEventoIECTeRoot,
	validateEventoCancIECTeRoot,
	validateEventoPrestDesacordoRoot,
	validateEventoCancPrestDesacordoRoot,
	validateEventoGenericoRoot,
	validateRetEventoCTeRoot,
	validateRetEventoCancCTeRoot,
	validateRetEventoCECTeRoot,
	validateRetEventoCancCECTeRoot,
	validateRetEventoEPECCTeRoot,
	validateRetEventoRegMultimodalRoot,
	validateRetEventoGTVRoot,
	validateRetEventoIECTeRoot,
	validateRetEventoCancIECTeRoot,
	validateRetEventoPrestDesacordoRoot,
	validateRetEventoCancPrestDesacordoRoot,
	validateRetEventoGenericoRoot,
	validateProcEventoCTeRoot,
	validateProcEventoCancCTeRoot,
	validateProcEventoCECTeRoot,
	validateProcEventoCancCECTeRoot,
	validateProcEventoEPECCTeRoot,
	validateProcEventoRegMultimodalRoot,
	validateProcEventoGTVRoot,
	validateProcEventoIECTeRoot,
	validateProcEventoCancIECTeRoot,
	validateProcEventoPrestDesacordoRoot,
	validateProcEventoCancPrestDesacordoRoot,
	validateProcEventoGenericoRoot,
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

func validateEventoCTeRoot(doc *Document) error {
	if doc.EventoCTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCTe.InfEvento)
}

func validateEventoCancCTeRoot(doc *Document) error {
	if doc.EventoCancCTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancCTe.InfEvento)
}

func validateEventoCECTeRoot(doc *Document) error {
	if doc.EventoCECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCECTe.InfEvento)
}

func validateEventoCancCECTeRoot(doc *Document) error {
	if doc.EventoCancCECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancCECTe.InfEvento)
}

func validateEventoEPECCTeRoot(doc *Document) error {
	if doc.EventoEPECCTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoEPECCTe.InfEvento)
}

func validateEventoRegMultimodalRoot(doc *Document) error {
	if doc.EventoRegMultimodal == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoRegMultimodal.InfEvento)
}

func validateEventoGTVRoot(doc *Document) error {
	if doc.EventoGTV == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoGTV.InfEvento)
}

func validateEventoIECTeRoot(doc *Document) error {
	if doc.EventoIECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoIECTe.InfEvento)
}

func validateEventoCancIECTeRoot(doc *Document) error {
	if doc.EventoCancIECTe == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancIECTe.InfEvento)
}

func validateEventoPrestDesacordoRoot(doc *Document) error {
	if doc.EventoPrestDesacordo == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoPrestDesacordo.InfEvento)
}

func validateEventoCancPrestDesacordoRoot(doc *Document) error {
	if doc.EventoCancPrestDesacordo == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoCancPrestDesacordo.InfEvento)
}

func validateEventoGenericoRoot(doc *Document) error {
	if doc.EventoGenerico == nil {
		return nil
	}
	return validateCTeEvent(doc.EventoGenerico.InfEvento)
}

func missingEventoCTeIfNil(present, eventoNil bool) error {
	if present && eventoNil {
		return errors.New("parse cte: missing eventoCTe")
	}
	return nil
}

func missingRetEventoCTeIfNil(present, retEventoNil bool) error {
	if present && retEventoNil {
		return errors.New("parse cte: missing retEventoCTe")
	}
	return nil
}

func validateRetEventoCTeRoot(doc *Document) error {
	if doc.RetEventoCTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCTe)
}

func validateRetEventoCancCTeRoot(doc *Document) error {
	if doc.RetEventoCancCTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancCTe)
}

func validateRetEventoCECTeRoot(doc *Document) error {
	if doc.RetEventoCECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCECTe)
}

func validateRetEventoCancCECTeRoot(doc *Document) error {
	if doc.RetEventoCancCECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancCECTe)
}

func validateRetEventoEPECCTeRoot(doc *Document) error {
	if doc.RetEventoEPECCTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoEPECCTe)
}

func validateRetEventoRegMultimodalRoot(doc *Document) error {
	if doc.RetEventoRegMultimodal == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoRegMultimodal)
}

func validateRetEventoGTVRoot(doc *Document) error {
	if doc.RetEventoGTV == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoGTV)
}

func validateRetEventoIECTeRoot(doc *Document) error {
	if doc.RetEventoIECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoIECTe)
}

func validateRetEventoCancIECTeRoot(doc *Document) error {
	if doc.RetEventoCancIECTe == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancIECTe)
}

func validateRetEventoPrestDesacordoRoot(doc *Document) error {
	if doc.RetEventoPrestDesacordo == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoPrestDesacordo)
}

func validateRetEventoCancPrestDesacordoRoot(doc *Document) error {
	if doc.RetEventoCancPrestDesacordo == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoCancPrestDesacordo)
}

func validateRetEventoGenericoRoot(doc *Document) error {
	if doc.RetEventoGenerico == nil {
		return nil
	}
	return validateCTeRetEventRoot(doc.RetEventoGenerico)
}

func validateProcEventoCTeRoot(doc *Document) error {
	if doc.ProcEventoCTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCTe.EventoCTe, doc.ProcEventoCTe.RetEventoCTe)
}

func validateProcEventoCancCTeRoot(doc *Document) error {
	if doc.ProcEventoCancCTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancCTe.EventoCTe, doc.ProcEventoCancCTe.RetEventoCTe)
}

func validateProcEventoCECTeRoot(doc *Document) error {
	if doc.ProcEventoCECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCECTe.EventoCTe, doc.ProcEventoCECTe.RetEventoCTe)
}

func validateProcEventoCancCECTeRoot(doc *Document) error {
	if doc.ProcEventoCancCECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancCECTe.EventoCTe, doc.ProcEventoCancCECTe.RetEventoCTe)
}

func validateProcEventoEPECCTeRoot(doc *Document) error {
	if doc.ProcEventoEPECCTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoEPECCTe.EventoCTe, doc.ProcEventoEPECCTe.RetEventoCTe)
}

func validateProcEventoRegMultimodalRoot(doc *Document) error {
	if doc.ProcEventoRegMultimodal == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoRegMultimodal.EventoCTe, doc.ProcEventoRegMultimodal.RetEventoCTe)
}

func validateProcEventoGTVRoot(doc *Document) error {
	if doc.ProcEventoGTV == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoGTV.EventoCTe, doc.ProcEventoGTV.RetEventoCTe)
}

func validateProcEventoIECTeRoot(doc *Document) error {
	if doc.ProcEventoIECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoIECTe.EventoCTe, doc.ProcEventoIECTe.RetEventoCTe)
}

func validateProcEventoCancIECTeRoot(doc *Document) error {
	if doc.ProcEventoCancIECTe == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancIECTe.EventoCTe, doc.ProcEventoCancIECTe.RetEventoCTe)
}

func validateProcEventoPrestDesacordoRoot(doc *Document) error {
	if doc.ProcEventoPrestDesacordo == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoPrestDesacordo.EventoCTe, doc.ProcEventoPrestDesacordo.RetEventoCTe)
}

func validateProcEventoCancPrestDesacordoRoot(doc *Document) error {
	if doc.ProcEventoCancPrestDesacordo == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoCancPrestDesacordo.EventoCTe, doc.ProcEventoCancPrestDesacordo.RetEventoCTe)
}

func validateProcEventoGenericoRoot(doc *Document) error {
	if doc.ProcEventoGenerico == nil {
		return nil
	}
	return validateCTeProcessedEvent(doc.ProcEventoGenerico.EventoCTe, doc.ProcEventoGenerico.RetEventoCTe)
}

func validateCTeProcessedEvent(evento, retEvento any) error {
	if err := validateCTeSentEventRoot(evento); err != nil {
		return err
	}
	return validateCTeRetEventRoot(retEvento)
}

func validateCTeSentEventRoot(evento any) error {
	switch v := evento.(type) {
	case *eventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *ceEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelCEEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *epecEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *regMultimodalEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *gtvEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *ieEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelIEEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *prestDesacordoEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *cancelPrestDesacordoEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	case *genericEventSchema.TEvento:
		if v == nil {
			return missingEventoCTeIfNil(true, true)
		}
		return validateCTeEvent(v.InfEvento)
	default:
		return missingEventoCTeIfNil(true, true)
	}
}

func validateCTeRetEventRoot(retEvento any) error {
	switch v := retEvento.(type) {
	case *eventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *ceEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelCEEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *epecEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *regMultimodalEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *gtvEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *ieEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelIEEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *prestDesacordoEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *cancelPrestDesacordoEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	case *genericEventSchema.TRetEvento:
		if v == nil {
			return missingRetEventoCTeIfNil(true, true)
		}
		return validateCTeRetEvent(v.InfEvento)
	default:
		return missingRetEventoCTeIfNil(true, true)
	}
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

func validateCTeEvent(inf any) error {
	switch v := inf.(type) {
	case *eventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *ceEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelCEEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *epecEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *regMultimodalEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *gtvEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *ieEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelIEEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *prestDesacordoEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *cancelPrestDesacordoEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	case *genericEventSchema.TAnonComplexInfEvento1:
		if v == nil {
			return validateInfEvento(false, "", false)
		}
		return validateInfEvento(v != nil, v.ChCTe, v.DetEvento != nil)
	default:
		return errors.New("parse cte: missing infEvento")
	}
}

func validateCTeRetEvent(inf any) error {
	switch v := inf.(type) {
	case *eventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *ceEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelCEEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *epecEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *regMultimodalEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *gtvEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *ieEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelIEEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *prestDesacordoEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *cancelPrestDesacordoEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	case *genericEventSchema.TAnonComplexInfEvento2:
		if v == nil {
			return validateRetInfEvento(false, "", "")
		}
		return validateRetInfEvento(v != nil, v.TpAmb, v.CStat)
	default:
		return errors.New("parse cte: missing infEvento")
	}
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
		doc.EventoCTe != nil,
		doc.RetEventoCTe != nil,
		doc.ProcEventoCTe != nil,
		doc.EventoCancCTe != nil,
		doc.RetEventoCancCTe != nil,
		doc.ProcEventoCancCTe != nil,
		doc.EventoCECTe != nil,
		doc.RetEventoCECTe != nil,
		doc.ProcEventoCECTe != nil,
		doc.EventoCancCECTe != nil,
		doc.RetEventoCancCECTe != nil,
		doc.ProcEventoCancCECTe != nil,
		doc.EventoEPECCTe != nil,
		doc.RetEventoEPECCTe != nil,
		doc.ProcEventoEPECCTe != nil,
		doc.EventoRegMultimodal != nil,
		doc.RetEventoRegMultimodal != nil,
		doc.ProcEventoRegMultimodal != nil,
		doc.EventoGTV != nil,
		doc.RetEventoGTV != nil,
		doc.ProcEventoGTV != nil,
		doc.EventoIECTe != nil,
		doc.RetEventoIECTe != nil,
		doc.ProcEventoIECTe != nil,
		doc.EventoCancIECTe != nil,
		doc.RetEventoCancIECTe != nil,
		doc.ProcEventoCancIECTe != nil,
		doc.EventoPrestDesacordo != nil,
		doc.RetEventoPrestDesacordo != nil,
		doc.ProcEventoPrestDesacordo != nil,
		doc.EventoCancPrestDesacordo != nil,
		doc.RetEventoCancPrestDesacordo != nil,
		doc.ProcEventoCancPrestDesacordo != nil,
		doc.EventoGenerico != nil,
		doc.RetEventoGenerico != nil,
		doc.ProcEventoGenerico != nil,
		doc.DistDFeInt != nil,
		doc.RetDistDFeInt != nil,
	} {
		if ok {
			count++
		}
	}
	return count
}
