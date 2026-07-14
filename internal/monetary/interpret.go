package monetary

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awafinance/fiscal/pkg/fiscalerr"
	"github.com/awafinance/fiscal/pkg/info"
)

const nfeSummaryRoot = "resNFe"

type interpretedFact struct {
	fact info.MonetaryFact
	node *node
}

func Interpret(data []byte, family fiscalerr.Family, rootName string) (*info.MonetaryInterpretation, error) {
	root, err := parseTree(data)
	if err != nil {
		return nil, fmt.Errorf("interpret monetary facts: %w", err)
	}

	provenance, status, reason, err := releaseFor(root, family, rootName)
	if err != nil {
		return nil, fmt.Errorf("interpret monetary facts: %w", err)
	}
	result := &info.MonetaryInterpretation{
		ContractVersion: info.MonetaryContractVersion,
		Status:          status,
		Reason:          reason,
		Provenance:      provenance,
		Facts:           []info.MonetaryFact{},
		Checks:          []info.MonetaryCheck{},
	}
	if status == info.MonetaryInterpretationUnsupported || provenance.Scope == info.FiscalArtifactLifecycle {
		return result, nil
	}

	facts, err := collectFacts(root, family)
	if err != nil {
		return nil, fmt.Errorf("interpret monetary facts: %w", err)
	}
	enrichRetentions(root, family, facts)
	for _, fact := range facts {
		result.Facts = append(result.Facts, fact.fact)
	}

	switch family {
	case fiscalerr.NFe:
		result.Checks = nfeChecks(root, facts, provenance)
	case fiscalerr.CTe:
		result.Checks = cteChecks(root, facts, provenance)
	case fiscalerr.NFSe:
		result.Checks = nfseChecks(root, facts, provenance)
	case fiscalerr.MDFe, fiscalerr.BPe:
	}
	return result, nil
}

//nolint:gocognit,gocyclo // Keeping the pinned release matrix together makes it auditable.
func releaseFor(
	root *node,
	family fiscalerr.Family,
	rootName string,
) (info.MonetaryProvenance, info.MonetaryInterpretationStatus, string, error) {
	provenance := info.MonetaryProvenance{
		Family:    string(family),
		Root:      rootName,
		Namespace: root.name.Space,
		Scope:     artifactScope(family, rootName),
	}
	if family == fiscalerr.MDFe || family == fiscalerr.BPe {
		return provenance, info.MonetaryInterpretationUnsupported,
			"this family has no monetary interpretation contract", nil
	}

	version, model := documentRelease(root, family, rootName)
	provenance.Layout = version
	provenance.Model = model

	switch family {
	case fiscalerr.NFe:
		switch provenance.Scope {
		case info.FiscalArtifactFullDocument:
			if version != "4.00" || (model != "55" && model != "65") {
				return provenance, "", "", unsupportedRelease(family, rootName, model, version)
			}
			provenance.Schema = "NF-e PL 010e v1.02"
			provenance.SchemaFiles = []info.MonetarySchemaSource{
				{
					Name:   "PL_010e_v1.02",
					URL:    "https://www.nfe.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=akib2DRpJN4%3D",
					SHA256: "d44ae5aa6a0d1cabf6235d2d2d47b75be5dd87bc6b90a7ec3dcec99c3d41bda1",
				},
			}
		case info.FiscalArtifactSummary:
			if rootName != nfeSummaryRoot || version != "1.01" {
				return provenance, "", "", unsupportedRelease(family, rootName, model, version)
			}
			provenance.Schema = "NF-e Distribuicao DF-e v1.04"
			provenance.SchemaFiles = []info.MonetarySchemaSource{
				{
					Name:   "PL_NFeDistDFe_104",
					URL:    "https://www.nfe.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=IzuP2y0G6hk%3D",
					SHA256: "9bd6d478dd04770016783a914111f3e2cc794c6a2f2b62ac7243fb101c138740",
				},
			}
		case info.FiscalArtifactDeclaration, info.FiscalArtifactLifecycle:
		}
	case fiscalerr.CTe:
		if provenance.Scope == info.FiscalArtifactFullDocument {
			if version != "4.00" || !cteModelMatchesRoot(rootName, model) {
				return provenance, "", "", unsupportedRelease(family, rootName, model, version)
			}
			provenance.Schema = "CT-e 4.00 through NT 2026.002 v1.00"
			provenance.SchemaFiles = []info.MonetarySchemaSource{
				{
					Name:   "PL_CTe_400_NT2026.001_1.01c_corr",
					URL:    "https://www.cte.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=2oAI3pZFSWk%3D",
					SHA256: "6d61b6a586f8afc7312761cd7d16450770d194b5eae1a84805431975bd1a3ee4",
				},
				{
					Name:   "PL_CTe_400_NT2025.001_RTC_1.14a_corr",
					URL:    "https://www.cte.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=DeEWRezKyLo%3D",
					SHA256: "ef33c97392c4de2e87bc43b7151202bfe836d999feeb310f3e27a4a7b2c8554b",
				},
				{
					Name:   "PL_CTe_400_NT2026.002_RTC_1.00",
					URL:    "https://www.cte.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=EGfyMjCQNeE%3D",
					SHA256: "b420eb4643056f29191f5c25da8aef96432dc5ad46952be475c9b771891b817d",
				},
			}
		}
	case fiscalerr.NFSe:
		if provenance.Scope == info.FiscalArtifactFullDocument || provenance.Scope == info.FiscalArtifactDeclaration {
			if version != "1.00" && version != "1.01" {
				return provenance, "", "", unsupportedRelease(family, rootName, model, version)
			}
			provenance.Schema = "NFS-e Nacional v1.01 2026-02-09"
			provenance.SchemaFiles = []info.MonetarySchemaSource{
				{
					Name:   "NFSe-ESQUEMAS_XSD-v1.01-20260209",
					URL:    "https://www.gov.br/nfse/pt-br/biblioteca/documentacao-tecnica/documentacao-atual/nfse-esquemas_xsd-v1-01-20260209.zip",
					SHA256: "e7935cbd9470527c6cc32984c1b2263e614183bf0139ce2733eaaed2de9a8072",
				},
			}
		}
	case fiscalerr.MDFe, fiscalerr.BPe:
	}

	return provenance, info.MonetaryInterpretationComplete, "", nil
}

func unsupportedRelease(family fiscalerr.Family, rootName, model, version string) error {
	return &fiscalerr.UnsupportedReleaseError{
		Family:  family,
		Root:    rootName,
		Model:   model,
		Version: version,
	}
}

func artifactScope(family fiscalerr.Family, rootName string) info.FiscalArtifactScope {
	switch family {
	case fiscalerr.NFe:
		switch rootName {
		case "NFe", "nfeProc":
			return info.FiscalArtifactFullDocument
		case nfeSummaryRoot:
			return info.FiscalArtifactSummary
		default:
			return info.FiscalArtifactLifecycle
		}
	case fiscalerr.CTe:
		switch rootName {
		case "CTe", "cteProc", "CTeOS", "cteOSProc", "CTeSimp", "cteSimpProc", "GTVe", "GTVeProc":
			return info.FiscalArtifactFullDocument
		default:
			return info.FiscalArtifactLifecycle
		}
	case fiscalerr.NFSe:
		switch rootName {
		case "NFSe":
			return info.FiscalArtifactFullDocument
		case "DPS":
			return info.FiscalArtifactDeclaration
		default:
			return info.FiscalArtifactLifecycle
		}
	default:
		return info.FiscalArtifactLifecycle
	}
}

func documentRelease(root *node, family fiscalerr.Family, rootName string) (string, string) {
	switch family {
	case fiscalerr.NFe:
		if rootName == nfeSummaryRoot {
			return root.attr("versao"), "55"
		}
		inf := root.firstDescendant("infNFe")
		if inf == nil {
			return root.attr("versao"), ""
		}
		return inf.attr("versao"), inf.firstDescendant("mod").value()
	case fiscalerr.CTe:
		inf := root.firstDescendant("infCte", "infCteOS", "infCteSimp", "infGTVe")
		if inf == nil {
			return root.attr("versao"), ""
		}
		return inf.attr("versao"), inf.firstDescendant("mod").value()
	case fiscalerr.NFSe:
		version := root.attr("versao")
		if version == "" {
			if inf := root.firstDescendant("infNFSe", "infDPS"); inf != nil {
				version = inf.attr("versao")
			}
		}
		return version, ""
	default:
		return root.attr("versao"), ""
	}
}

func cteModelMatchesRoot(rootName, model string) bool {
	switch rootName {
	case "CTe", "cteProc", "CTeSimp", "cteSimpProc":
		return model == "57"
	case "CTeOS", "cteOSProc":
		return model == "67"
	case "GTVe", "GTVeProc":
		return model == "64"
	default:
		return false
	}
}

func collectFacts(root *node, family fiscalerr.Family) ([]interpretedFact, error) {
	counts := make(map[string]int)
	var facts []interpretedFact
	var visit func(*node) error
	visit = func(current *node) error {
		if current.name.Space == "http://www.w3.org/2000/09/xmldsig#" {
			return nil
		}
		if len(current.children) == 0 && isMonetaryElement(current.name.Local) {
			value := current.value()
			if _, err := parseDecimal(value); err != nil {
				return fmt.Errorf("%s at %s is not an exact decimal: %w", current.name.Local, exactPath(current), err)
			}
			code := factCode(current, family)
			counts[code]++
			occurrence := counts[code]
			fact := info.MonetaryFact{
				ID:         code + "#" + strconv.Itoa(occurrence),
				Code:       code,
				Path:       exactPath(current),
				Element:    current.name.Local,
				Value:      value,
				Unit:       factUnit(current),
				Currency:   factCurrency(current),
				Kind:       factKind(current),
				Scope:      factScope(current),
				Occurrence: occurrence,
				Label:      factLabel(current),
				TaxTypes:   factTaxTypes(current),
				Qualifiers: factQualifiers(current),
			}
			facts = append(facts, interpretedFact{fact: fact, node: current})
		}
		for _, child := range current.children {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	return facts, nil
}

func isMonetaryElement(name string) bool {
	if strings.HasPrefix(name, "pTotTrib") {
		return true
	}
	return len(name) > 1 && name[0] == 'v' && name[1] >= 'A' && name[1] <= 'Z'
}

func exactPath(n *node) string {
	var path strings.Builder
	for _, part := range n.path() {
		path.WriteString("/Q{")
		path.WriteString(part.name.Space)
		path.WriteString("}")
		path.WriteString(part.name.Local)
		path.WriteString("[")
		path.WriteString(strconv.Itoa(part.occurrence))
		path.WriteString("]")
	}
	return path.String()
}

func factCode(n *node, family fiscalerr.Family) string {
	path := n.path()
	prefix := string(family)
	start := 0
	for i, part := range path {
		switch part.name.Local {
		case "infNFe":
			prefix, start = "nfe.document", i+1
		case nfeSummaryRoot:
			prefix, start = "nfe.summary", i+1
		case "infCte":
			prefix, start = "cte.document", i+1
		case "infCteOS":
			prefix, start = "cte.os", i+1
		case "infCteSimp":
			prefix, start = "cte.simplified", i+1
		case "infGTVe":
			prefix, start = "cte.gtve", i+1
		case "infNFSe":
			prefix, start = "nfse.document", i+1
		case "infDPS":
			prefix, start = "nfse.dps", i+1
		}
	}
	parts := []string{prefix}
	for _, part := range path[start:] {
		parts = append(parts, part.name.Local)
	}
	return strings.Join(parts, ".")
}
