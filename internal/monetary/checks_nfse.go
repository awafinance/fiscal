package monetary

import "github.com/awafinance/fiscal/pkg/info"

func nfseChecks(root *node, facts []interpretedFact, provenance info.MonetaryProvenance) []info.MonetaryCheck {
	if provenance.Scope != info.FiscalArtifactFullDocument {
		return []info.MonetaryCheck{}
	}
	var checks []info.MonetaryCheck
	if factWithCodeSuffix(facts, ".valores.vLiq") != nil {
		checks = append(checks, nfseLiquidValueCheck(facts))
	}
	if root.firstDescendant("vTotNF") != nil {
		checks = append(checks, nfseDocumentTotalCheck(root, facts))
	}
	return checks
}

func nfseLiquidValueCheck(facts []interpretedFact) info.MonetaryCheck {
	declared := factWithCodeSuffix(facts, ".valores.vLiq")
	service := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vServ" && fact.fact.Kind == info.MonetaryFactServiceAmount
	})
	discounts := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Kind == info.MonetaryFactDiscount && fact.fact.Scope == info.MonetaryScopeDocument
	})
	retentions := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vTotalRet" && fact.fact.Scope == info.MonetaryScopeDocument
	})
	terms := termsForFacts("+", "DPS service value", service...)
	terms = append(terms, termsForFacts("-", "declared DPS discount", discounts...)...)
	terms = append(terms, termsForFacts("-", "generated NFS-e total retentions", retentions...)...)
	return equalityCheck(
		"NFSE-vLiq",
		"generated NFS-e liquid value follows the schema-defined service, discount, and retention equation",
		"",
		declared,
		"0.01",
		[]calculationVariant{{name: "schema_vLiq_equation", terms: terms}},
	)
}

func nfseDocumentTotalCheck(root *node, facts []interpretedFact) info.MonetaryCheck {
	declared := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vTotNF" && fact.fact.Scope == info.MonetaryScopeDocument
	})
	liquid := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vLiq" && fact.fact.Scope == info.MonetaryScopeDocument
	})
	terms := termsForFacts("+", "generated NFS-e liquid value", liquid...)
	variant := "2026_vLiq"
	if yearOf(firstNonEmpty(firstValue(root, "dhProc"), firstValue(root, "dhEmi"))) >= 2027 {
		variant = "2027_vLiq_plus_IBS_CBS"
		ibs := factsMatching(facts, func(fact interpretedFact) bool {
			return fact.fact.Element == "vIBSTot" && fact.fact.Scope == info.MonetaryScopeDocument
		})
		cbs := factsMatching(facts, func(fact interpretedFact) bool {
			return fact.fact.Element == "vCBS" && fact.fact.Scope == info.MonetaryScopeDocument
		})
		terms = append(terms, termsForFacts("+", "RTC tax included after the transition", ibs...)...)
		terms = append(terms, termsForFacts("+", "RTC tax included after the transition", cbs...)...)
	}
	var declaredFact *interpretedFact
	if len(declared) > 0 {
		declaredFact = declared[0]
	}
	return equalityCheck(
		"NFSE-vTotNF",
		"declared NFS-e RTC document total follows the effective transition equation",
		"",
		declaredFact,
		"0.01",
		[]calculationVariant{{name: variant, terms: terms}},
	)
}
