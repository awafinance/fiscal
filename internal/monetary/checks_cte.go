package monetary

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

const cteTotalServiceElement = "vTPrest"

func cteChecks(root *node, facts []interpretedFact, provenance info.MonetaryProvenance) []info.MonetaryCheck {
	if provenance.Scope != info.FiscalArtifactFullDocument {
		return []info.MonetaryCheck{}
	}
	checks := []info.MonetaryCheck{cteReceivableCheck(facts)}
	if root.firstDescendant("IBSCBS") != nil || root.firstDescendant("vTotDFe") != nil {
		checks = append(checks, cteDocumentTotalCheck(root, facts))
	}
	if strings.Contains(provenance.Root, "CTeOS") || strings.Contains(provenance.Root, "cteOS") {
		if root.firstDescendant("infGTVe") != nil {
			checks = append(checks, cteOSGTVeComponentsCheck(facts))
		}
	}
	return checks
}

func cteReceivableCheck(facts []interpretedFact) info.MonetaryCheck {
	total := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == cteTotalServiceElement && fact.fact.Scope == info.MonetaryScopeDocument
	})
	receivable := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vRec" && fact.fact.Scope == info.MonetaryScopeDocument
	})
	check := info.MonetaryCheck{
		Rule:          "G048/H056",
		Description:   "CT-e amount receivable does not exceed total service value",
		RejectionCode: "531",
		Tolerance:     "0.00",
	}
	if len(total) == 0 || len(receivable) == 0 {
		check.Status = info.MonetaryCheckInsufficientData
		check.Reason = "vTPrest or vRec is absent"
		return check
	}
	check.DeclaredFact = receivable[0].fact.ID
	check.DeclaredValue = receivable[0].fact.Value
	totalValue, totalErr := parseDecimal(total[0].fact.Value)
	receivableValue, receivableErr := parseDecimal(receivable[0].fact.Value)
	if totalErr != nil || receivableErr != nil {
		check.Status = info.MonetaryCheckInsufficientData
		check.Reason = "vTPrest or vRec is not an exact decimal"
		return check
	}
	difference := receivableValue.subtract(totalValue)
	check.Status = info.MonetaryCheckPassed
	if difference.cmp(zeroDecimal(2)) > 0 {
		check.Status = info.MonetaryCheckFailed
	}
	check.Calculations = []info.MonetaryCalculation{
		{
			Variant:    "vRec_lte_vTPrest",
			Status:     check.Status,
			Calculated: totalValue.stringAtLeast(2),
			Difference: difference.stringAtLeast(2),
			Terms: []info.MonetaryCheckTerm{
				{FactID: receivable[0].fact.ID, Operation: "compare", Value: receivable[0].fact.Value},
				{FactID: total[0].fact.ID, Operation: "<=", Value: total[0].fact.Value},
			},
		},
	}
	return check
}

func cteDocumentTotalCheck(root *node, facts []interpretedFact) info.MonetaryCheck {
	declared := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vTotDFe" && fact.fact.Scope == info.MonetaryScopeDocument
	})
	if len(declared) == 0 {
		return info.MonetaryCheck{
			Rule:          "vTotDFe-rule-01",
			Description:   "vTotDFe is required when the CT-e IBSCBS group is present",
			Status:        info.MonetaryCheckFailed,
			RejectionCode: "360",
			Reason:        "IBSCBS is present and vTotDFe is absent",
		}
	}
	total := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == cteTotalServiceElement && fact.fact.Scope == info.MonetaryScopeDocument
	})
	terms := termsForFacts("+", "total service value", total...)
	variant := "2026_vTPrest"
	if yearOf(firstNonEmpty(firstValue(root, "dhEmi"), firstValue(root, "dEmi"))) >= 2027 {
		variant = "2027_vTPrest_plus_IBS_CBS"
		ibs := factsMatching(facts, func(fact interpretedFact) bool {
			return fact.fact.Element == "vIBS" && fact.fact.Scope == info.MonetaryScopeDocument
		})
		cbs := factsMatching(facts, func(fact interpretedFact) bool {
			return fact.fact.Element == "vCBS" && fact.fact.Scope == info.MonetaryScopeDocument
		})
		terms = append(terms, termsForFacts("+", "RTC tax included after the transition", ibs...)...)
		terms = append(terms, termsForFacts("+", "RTC tax included after the transition", cbs...)...)
	}
	return equalityCheck(
		"vTotDFe-rule-02",
		"declared CT-e RTC document total follows the effective transition equation",
		"365",
		declared[0],
		"0.01",
		[]calculationVariant{{name: variant, terms: terms}},
	)
}

func cteOSGTVeComponentsCheck(facts []interpretedFact) info.MonetaryCheck {
	total := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == cteTotalServiceElement && fact.fact.Scope == info.MonetaryScopeDocument
	})
	components := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vComp" && strings.HasPrefix(fact.fact.Code, "cte.gtve.")
	})
	var declared *interpretedFact
	if len(total) > 0 {
		declared = total[0]
	}
	return equalityCheck(
		"H053",
		"CT-e OS GTV-e component values sum to total service value when the GTV-e group applies",
		"899",
		declared,
		"0.10",
		[]calculationVariant{
			{name: "sum_infGTVe_components", terms: termsForFacts("+", "infGTVe component", components...)},
		},
	)
}

func yearOf(value string) int {
	if len(value) < 4 {
		return 0
	}
	year := 0
	for _, r := range value[:4] {
		if r < '0' || r > '9' {
			return 0
		}
		year = year*10 + int(r-'0')
	}
	return year
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
