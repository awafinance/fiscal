package monetary

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

func nfeChecks(root *node, facts []interpretedFact, provenance info.MonetaryProvenance) []info.MonetaryCheck {
	if provenance.Scope != info.FiscalArtifactFullDocument {
		return []info.MonetaryCheck{}
	}
	checks := []info.MonetaryCheck{nfeW16Check(root, facts)}
	if factWithCodeSuffix(facts, ".total.IBSCBSTot.vNFTot") != nil ||
		factWithCodeSuffix(facts, ".total.vNFTot") != nil || root.firstDescendant("vNFTot") != nil {
		checks = append(checks, nfeW60Check(facts))
	}
	return checks
}

func nfeW16Check(root *node, facts []interpretedFact) info.MonetaryCheck {
	declared := factWithCodeSuffix(facts, ".total.ICMSTot.vNF")
	check := info.MonetaryCheck{
		Rule:          "W16-10",
		Description:   "declared NF-e total equals the applicable official component equation",
		RejectionCode: "610",
		Tolerance:     "0.01",
	}
	for _, cfop := range root.descendants("CFOP") {
		if strings.HasPrefix(cfop.value(), "3") {
			check.Status = info.MonetaryCheckNotApplicable
			check.Reason = "W16-10 does not apply to import operations whose CFOP starts with 3"
			if declared != nil {
				check.DeclaredFact = declared.fact.ID
				check.DeclaredValue = declared.fact.Value
			}
			return check
		}
	}

	terms := nfeW16BaseTerms(facts)
	if directVehicleBilling(root) {
		terms = nfeW16DirectVehicleTerms(facts)
	}
	terms = append(terms, nfeSubstitutionTaxTerms(root, facts)...)

	withoutDesonerado := make([]calculationTerm, 0, len(terms))
	for _, term := range terms {
		if term.fact != nil && term.fact.fact.Element == "vICMSDeson" {
			continue
		}
		if term.fact == nil && strings.Contains(term.reason, "vICMSDeson") {
			continue
		}
		withoutDesonerado = append(withoutDesonerado, term)
	}
	return equalityCheck(
		check.Rule,
		check.Description,
		check.RejectionCode,
		declared,
		check.Tolerance,
		[]calculationVariant{
			{name: "deduct_icms_desonerado", terms: terms},
			{name: "do_not_deduct_icms_desonerado", terms: withoutDesonerado},
		},
	)
}

func nfeW16BaseTerms(facts []interpretedFact) []calculationTerm {
	return []calculationTerm{
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vProd"), "+", "W07 vProd"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vDesc"), "-", "W10 vDesc"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vICMSDeson"), "-", "W04a vICMSDeson"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vST"), "+", "W06 vST"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vFCPST"), "+", "W06a vFCPST"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vICMSMonoReten"), "+", "W06d vICMSMonoReten"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vFrete"), "+", "W08 vFrete"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vSeg"), "+", "W09 vSeg"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vOutro"), "+", "W15 vOutro"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vII"), "+", "W11 vII"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vIPI"), "+", "W12 vIPI"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vIPIDevol"), "+", "W12a vIPIDevol"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ISSQNtot.vServ"), "+", "W18 vServ"),
	}
}

func nfeW16DirectVehicleTerms(facts []interpretedFact) []calculationTerm {
	return []calculationTerm{
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vProd"), "+", "W07 vProd"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vDesc"), "-", "W10 vDesc"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vICMSDeson"), "-", "W04a vICMSDeson"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vICMSMonoReten"), "+", "W06d vICMSMonoReten"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vFrete"), "+", "W08 vFrete"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vSeg"), "+", "W09 vSeg"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vOutro"), "+", "W15 vOutro"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vII"), "+", "W11 vII"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ICMSTot.vIPI"), "+", "W12 vIPI"),
		declaredOrZero(factWithCodeSuffix(facts, ".total.ISSQNtot.vServ"), "+", "W18 vServ"),
	}
}

func directVehicleBilling(root *node) bool {
	for _, vehicle := range root.descendants("veicProd") {
		if vehicle.childValue("tpOp") == "2" {
			return true
		}
	}
	return false
}

func nfeSubstitutionTaxTerms(root *node, facts []interpretedFact) []calculationTerm {
	var terms []calculationTerm
	for _, groupName := range []string{"PISST", "COFINSST"} {
		for _, group := range root.descendants(groupName) {
			indicator := "indSoma" + groupName
			if group.childValue(indicator) != "1" {
				continue
			}
			valueName := "vPIS"
			if groupName == "COFINSST" {
				valueName = "vCOFINS"
			}
			valueNode := group.child(valueName)
			if valueNode == nil {
				continue
			}
			if fact := factForNode(facts, valueNode); fact != nil {
				terms = append(terms, calculationTerm{
					fact:      fact,
					operation: "+",
					reason:    groupName + " included because " + indicator + "=1",
				})
			}
		}
	}
	return terms
}

func nfeW60Check(facts []interpretedFact) info.MonetaryCheck {
	declared := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vNFTot" && fact.fact.Scope == info.MonetaryScopeDocument
	})
	var declaredFact *interpretedFact
	if len(declared) > 0 {
		declaredFact = declared[0]
	}
	items := factsMatching(facts, func(fact interpretedFact) bool {
		return fact.fact.Element == "vItem" && fact.fact.Scope == info.MonetaryScopeItem
	})
	return equalityCheck(
		"W60-10",
		"declared RTC NF-e total equals the sum of item vItem values",
		"1094",
		declaredFact,
		"0.01",
		[]calculationVariant{
			{name: "sum_item_vItem", terms: termsForFacts("+", "item vItem", items...)},
		},
	)
}
