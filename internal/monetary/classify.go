package monetary

import (
	"slices"
	"strings"

	"github.com/awafinance/fiscal/pkg/fiscalerr"
	"github.com/awafinance/fiscal/pkg/info"
)

func factUnit(n *node) string {
	if strings.HasPrefix(n.name.Local, "pTotTrib") {
		return "percent"
	}
	return "currency"
}

func factCurrency(n *node) string {
	if factUnit(n) != "currency" {
		return ""
	}
	if n.name.Local == "vServMoeda" {
		if parent := n.parent; parent != nil {
			return parent.childValue("tpMoeda")
		}
	}
	return "BRL"
}

//nolint:gocognit,gocyclo // Order encodes mutually exclusive fiscal semantic precedence.
func factKind(n *node) info.MonetaryFactKind {
	name := n.name.Local
	path := localPath(n)
	if strings.HasPrefix(name, "pTotTrib") {
		return info.MonetaryFactApproximateTaxRate
	}
	if name == "vTotTrib" || strings.HasPrefix(name, "vTotTribFed") ||
		strings.HasPrefix(name, "vTotTribEst") || strings.HasPrefix(name, "vTotTribMun") {
		return info.MonetaryFactApproximateTax
	}
	if name == "vNF" || name == "vNFTot" || name == "vTotDFe" || name == "vTotNF" {
		return info.MonetaryFactDocumentTotal
	}
	if name == "vTPrest" {
		return info.MonetaryFactServiceTotal
	}
	if name == "vRec" || name == "vLiq" {
		return info.MonetaryFactSettlementAmount
	}
	if name == "vServ" || name == "vServMoeda" {
		return info.MonetaryFactServiceAmount
	}
	if strings.HasPrefix(name, "vRet") || name == "vTotalRet" || strings.Contains(path, ".retTrib.") {
		return info.MonetaryFactRetention
	}
	if strings.HasPrefix(name, "vBC") && len(factTaxTypes(n)) > 0 {
		return info.MonetaryFactTaxBase
	}
	if len(factTaxTypes(n)) > 0 {
		return info.MonetaryFactActualTax
	}
	if name == "vDesc" || strings.HasPrefix(name, "vDesc") {
		return info.MonetaryFactDiscount
	}
	if strings.Contains(path, ".fat.") {
		return info.MonetaryFactBilling
	}
	if name == "vDup" || strings.Contains(path, ".dup.") {
		return info.MonetaryFactInstallment
	}
	if name == "vPag" || name == "vTroco" || strings.Contains(path, ".detPag.") {
		return info.MonetaryFactPayment
	}
	if strings.HasPrefix(name, "vUn") {
		return info.MonetaryFactUnitPrice
	}
	if name == "vComp" || name == "vProd" || name == "vFrete" || name == "vSeg" ||
		name == "vOutro" || name == "vItem" || name == "vIPIDevol" {
		return info.MonetaryFactComponent
	}
	return info.MonetaryFactDeclaredAmount
}

func factScope(n *node) info.MonetaryFactScope {
	if n.nearestAncestor("detPag") != nil {
		return info.MonetaryScopePayment
	}
	if n.nearestAncestor("dup") != nil {
		return info.MonetaryScopeInstallment
	}
	if n.nearestAncestor("fat", "cobr") != nil {
		return info.MonetaryScopeBilling
	}
	if n.nearestAncestor("Comp") != nil {
		return info.MonetaryScopeComponent
	}
	if n.nearestAncestor("det") != nil {
		return info.MonetaryScopeItem
	}
	return info.MonetaryScopeDocument
}

func factLabel(n *node) string {
	if comp := n.nearestAncestor("Comp"); comp != nil {
		return comp.childValue("xNome")
	}
	return ""
}

func factQualifiers(n *node) []info.MonetaryQualifier {
	var qualifiers []info.MonetaryQualifier
	for _, part := range n.path() {
		for _, attrName := range []string{"Id", "nItem"} {
			if value := part.attr(attrName); value != "" {
				qualifiers = append(qualifiers, info.MonetaryQualifier{Name: attrName, Value: value})
			}
		}
	}
	if dup := n.nearestAncestor("dup"); dup != nil {
		if value := dup.childValue("nDup"); value != "" {
			qualifiers = append(qualifiers, info.MonetaryQualifier{Name: "nDup", Value: value})
		}
	}
	if payment := n.nearestAncestor("detPag"); payment != nil {
		if value := payment.childValue("tPag"); value != "" {
			qualifiers = append(qualifiers, info.MonetaryQualifier{Name: "tPag", Value: value})
		}
	}
	return qualifiers
}

func factTaxTypes(n *node) []string {
	name := strings.ToUpper(n.name.Local)
	path := strings.ToUpper(localPath(n))
	var types []string
	appendType := func(taxType string) {
		if !slices.Contains(types, taxType) {
			types = append(types, taxType)
		}
	}

	for _, candidate := range []struct {
		needle string
		name   string
	}{
		{"ICMS", "ICMS"},
		{"FCP", "FCP"},
		{"ISSQN", "ISS"},
		{"PIS", "PIS"},
		{"COFINS", "COFINS"},
		{"IPI", "IPI"},
		{"IBS", "IBS"},
		{"CBS", "CBS"},
		{"INSS", "INSS"},
		{"RETPREV", "INSS"},
		{"IRRF", "IR"},
		{"CSLL", "CSLL"},
	} {
		if strings.Contains(name, candidate.needle) || strings.Contains(path, "."+candidate.needle) {
			appendType(candidate.name)
		}
	}
	if name == "VII" || strings.Contains(path, ".II.") {
		appendType("II")
	}
	if name == "VIS" || strings.Contains(path, ".IS.") || strings.Contains(path, ".ISTOT.") {
		appendType("IS")
	}
	if name == "VIR" {
		appendType("IR")
	}
	return types
}

func localPath(n *node) string {
	parts := make([]string, 0, len(n.path()))
	for _, part := range n.path() {
		parts = append(parts, part.name.Local)
	}
	return "." + strings.Join(parts, ".") + "."
}

func enrichRetentions(root *node, family fiscalerr.Family, facts []interpretedFact) {
	for i := range facts {
		fact := &facts[i]
		if fact.fact.Kind == info.MonetaryFactRetention {
			fact.fact.Retention = &info.MonetaryRetention{
				Actor:      "unspecified_withholding_agent",
				TaxTypes:   slices.Clone(fact.fact.TaxTypes),
				Allocation: "dedicated",
			}
		}
	}
	if family != fiscalerr.NFSe {
		return
	}

	retISS := firstValue(root, "tpRetISSQN")
	for i := range facts {
		fact := &facts[i]
		switch fact.fact.Element {
		case "vISSQN":
			switch retISS {
			case "2":
				fact.fact.Retention = &info.MonetaryRetention{
					IndicatorCode: retISS,
					Actor:         "taker",
					TaxTypes:      []string{"ISS"},
					Allocation:    "referenced_actual_tax",
				}
			case "3":
				fact.fact.Retention = &info.MonetaryRetention{
					IndicatorCode: retISS,
					Actor:         "intermediary",
					TaxTypes:      []string{"ISS"},
					Allocation:    "referenced_actual_tax",
				}
			}
		case "vRetCSLL":
			code := firstValue(root, "tpRetPisCofins")
			fact.fact.TaxTypes = retainedSocialTaxes(code)
			fact.fact.Retention = &info.MonetaryRetention{
				IndicatorCode: code,
				Actor:         "unspecified_withholding_agent",
				TaxTypes:      slices.Clone(fact.fact.TaxTypes),
				Allocation:    "aggregate_unallocated",
			}
		}
	}
}

func retainedSocialTaxes(code string) []string {
	switch code {
	case "1", "4":
		return []string{"PIS", "COFINS"}
	case "3":
		return []string{"PIS", "COFINS", "CSLL"}
	case "5":
		return []string{"PIS"}
	case "6":
		return []string{"COFINS"}
	case "7":
		return []string{"COFINS", "CSLL"}
	case "8":
		return []string{"CSLL"}
	case "9":
		return []string{"PIS", "CSLL"}
	default:
		return []string{}
	}
}

func firstValue(root *node, name string) string {
	if found := root.firstDescendant(name); found != nil {
		return found.value()
	}
	return ""
}
