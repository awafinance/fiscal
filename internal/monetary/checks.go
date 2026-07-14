package monetary

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

type calculationTerm struct {
	fact      *interpretedFact
	operation string
	reason    string
	value     string
}

type calculationVariant struct {
	name  string
	terms []calculationTerm
}

func equalityCheck(
	rule string,
	description string,
	rejectionCode string,
	declared *interpretedFact,
	tolerance string,
	variants []calculationVariant,
) info.MonetaryCheck {
	check := info.MonetaryCheck{
		Rule:          rule,
		Description:   description,
		Tolerance:     tolerance,
		RejectionCode: rejectionCode,
	}
	if declared == nil {
		check.Status = info.MonetaryCheckInsufficientData
		check.Reason = "the declared comparison value is absent"
		return check
	}
	check.DeclaredFact = declared.fact.ID
	check.DeclaredValue = declared.fact.Value

	declaredValue, err := parseDecimal(declared.fact.Value)
	if err != nil {
		check.Status = info.MonetaryCheckInsufficientData
		check.Reason = "the declared comparison value is not an exact decimal"
		return check
	}
	toleranceValue, err := parseDecimal(tolerance)
	if err != nil {
		panic("invalid hard-coded monetary check tolerance: " + tolerance)
	}

	check.Status = info.MonetaryCheckFailed
	for _, variant := range variants {
		calculation := calculateEquality(variant.name, declaredValue, toleranceValue, variant.terms)
		check.Calculations = append(check.Calculations, calculation)
		if calculation.Status == info.MonetaryCheckPassed {
			check.Status = info.MonetaryCheckPassed
		}
	}
	return check
}

func calculateEquality(
	variant string,
	declared decimal,
	tolerance decimal,
	terms []calculationTerm,
) info.MonetaryCalculation {
	calculation := info.MonetaryCalculation{Variant: variant, Status: info.MonetaryCheckInsufficientData}
	total := zeroDecimal(2)
	for _, term := range terms {
		valueText := term.value
		factID := ""
		if term.fact != nil {
			valueText = term.fact.fact.Value
			factID = term.fact.fact.ID
		}
		value, err := parseDecimal(valueText)
		if err != nil {
			return calculation
		}
		if term.operation == "-" {
			total = total.subtract(value)
		} else {
			total = total.add(value)
		}
		calculation.Terms = append(calculation.Terms, info.MonetaryCheckTerm{
			FactID:    factID,
			Operation: term.operation,
			Value:     valueText,
			Reason:    term.reason,
		})
	}
	difference := declared.subtract(total).abs()
	calculation.Calculated = total.stringAtLeast(2)
	calculation.Difference = difference.stringAtLeast(2)
	calculation.Status = info.MonetaryCheckFailed
	if difference.cmp(tolerance) <= 0 {
		calculation.Status = info.MonetaryCheckPassed
	}
	return calculation
}

func factWithCodeSuffix(facts []interpretedFact, suffix string) *interpretedFact {
	for i := range facts {
		if strings.HasSuffix(facts[i].fact.Code, suffix) {
			return &facts[i]
		}
	}
	return nil
}

func factsMatching(facts []interpretedFact, match func(interpretedFact) bool) []*interpretedFact {
	var found []*interpretedFact
	for i := range facts {
		if match(facts[i]) {
			found = append(found, &facts[i])
		}
	}
	return found
}

func factForNode(facts []interpretedFact, wanted *node) *interpretedFact {
	for i := range facts {
		if facts[i].node == wanted {
			return &facts[i]
		}
	}
	return nil
}

func termsForFacts(operation, reason string, facts ...*interpretedFact) []calculationTerm {
	terms := make([]calculationTerm, 0, len(facts))
	for _, fact := range facts {
		if fact != nil {
			terms = append(terms, calculationTerm{fact: fact, operation: operation, reason: reason})
		}
	}
	return terms
}

func declaredOrZero(fact *interpretedFact, operation, reason string) calculationTerm {
	if fact == nil {
		return calculationTerm{operation: operation, value: "0.00", reason: "absent; contributes zero under the rule"}
	}
	return calculationTerm{fact: fact, operation: operation, reason: reason}
}
