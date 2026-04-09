package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/awa/fiscal/internal/codegen/postprocess"
)

const (
	bpeDetEventoStruct         = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
	bpeDetEventoStructInnerXML = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\tInnerXML         string   `xml:\",innerxml\"`\n}"
	infBPeCompField            = "\tComp        *TAnonComplexComp1        `xml:\"comp\"`"
	infBPeCompFieldFixed       = "\tComp        *TAnonComplexComp12       `xml:\"comp\"`"

	mdfeInfModalStruct         = "type InfModal struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
	mdfeInfModalStructInnerXML = "type InfModal struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tInnerXML        string   `xml:\",innerxml\"`\n}"
	mdfeAnonInfModalStruct     = "type TAnonComplexInfModal1 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
	mdfeAnonInfModalInnerXML   = "type TAnonComplexInfModal1 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tInnerXML        string   `xml:\",innerxml\"`\n}"
	mdfeDetEventoStruct        = "type DetEvento struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
	mdfeDetEventoInnerXML      = "type DetEvento struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\tInnerXML         string   `xml:\",innerxml\"`\n}"
	mdfeAnonDetEventoStruct    = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
	mdfeAnonDetEventoInnerXML  = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\tInnerXML         string   `xml:\",innerxml\"`\n}"

	cteDetEventoStruct         = "type DetEvento struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
	cteAnonDetEventoStruct     = "type TAnonComplexDetEvento1 struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n}"
	cteInfModalStruct          = "type InfModal struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
	cteAnonInfModalStruct      = "type TAnonComplexInfModal1 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
	cteAnonInfModalStruct2     = "type TAnonComplexInfModal2 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
	cteAnonInfModalStruct3     = "type TAnonComplexInfModal3 struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n}"
	xmlOnlyImportBlock         = "import (\n\t\"encoding/xml\"\n)"
	typedInfModalImportBlock   = "import (\n\t\"encoding/xml\"\n\n\tmodalaereo \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_aereo\"\n\tmodalaquaviario \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_aquaviario\"\n\tmodaldutoviario \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_dutoviario\"\n\tmodalferroviario \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_ferroviario\"\n\tmodalmultimodal \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_multimodal\"\n\tmodalrodoviario \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_rodoviario\"\n\tmodalrodoviarioos \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_rodoviario_os\"\n)"
	typedInfModalImportBlockOS = "import (\n\t\"encoding/xml\"\n\n\tmodalrodoviarioos \"github.com/awa/fiscal/internal/cte/gen/v4_0/modal_rodoviario_os\"\n)"

	xDescInterfaceField = "interface{} `xml:\"xDesc\"`"
	xDescStringField    = "string `xml:\"xDesc\"`"
)

var (
	optionalFieldDhCont = regexp.MustCompile(`\n\tDhCont\s+string\s+` + "`xml:\"dhCont\"`")
	optionalFieldXJust  = regexp.MustCompile(`\n\tXJust\s+string\s+` + "`xml:\"xJust\"`")
	optionalFieldCRT    = regexp.MustCompile(`\n\tCRT\s+string\s+` + "`xml:\"CRT\"`")
	nProtTProtField     = regexp.MustCompile(`\n\tNProt\s+\*TProt\s+` + "`xml:\"nProt\"`")
	nProtPrestDesField  = regexp.MustCompile(`\n\tNProtEvPrestDes\s+\*TProt\s+` + "`xml:\"nProtEvPrestDes\"`")

	cteEventPayloads = map[string]string{
		"evento_cce":                    "evCCeCTe",
		"evento_cancel":                 "evCancCTe",
		"evento_ce":                     "evCECTe",
		"evento_cancel_ce":              "evCancCECTe",
		"evento_cancel_ie":              "evCancIECTe",
		"evento_cancel_prest_desacordo": "evCancPrestDesacordo",
		"evento_epec":                   "evEPECCTe",
		"evento_gtv":                    "evGTV",
		"evento_ie":                     "evIECTe",
		"evento_prest_desacordo":        "evPrestDesacordo",
		"evento_reg_multimodal":         "evRegMultimodal",
	}
)

func codegenFamilies() []family {
	return []family{
		bpeFamily(),
		cteFamily(),
		mdfeFamily(),
		nfeFamily(),
		nfseFamily(),
	}
}

func bpeFamily() family {
	return family{
		name: "bpe",
		normalizeDirs: []string{
			"internal/bpe/schemas/v1_0/core",
			"internal/bpe/schemas/v1_0/evento_cancel",
			"internal/bpe/schemas/v1_0/evento_alteracao_poltrona",
			"internal/bpe/schemas/v1_0/evento_excesso_bagagem",
			"internal/bpe/schemas/v1_0/evento_nao_emb",
		},
		xgenJobs: []xgenJob{
			{"v1_0/core", "v1_0/core"},
			{"v1_0/evento_cancel", "v1_0/evento_cancel"},
			{"v1_0/evento_alteracao_poltrona", "v1_0/evento_alteracao_poltrona"},
			{"v1_0/evento_excesso_bagagem", "v1_0/evento_excesso_bagagem"},
			{"v1_0/evento_nao_emb", "v1_0/evento_nao_emb"},
		},
		postprocess: postprocessBPe,
	}
}

func cteFamily() family {
	return family{
		name: "cte",
		normalizeDirs: []string{
			"internal/cte/schemas/v4_0/cte",
			"internal/cte/schemas/v4_0/cte_os",
			"internal/cte/schemas/v4_0/consulta_situacao",
			"internal/cte/schemas/v4_0/status_servico",
			"internal/cte/schemas/v4_0/cte_simp",
			"internal/cte/schemas/v4_0/gtve",
			"internal/cte/schemas/v4_0/modal_aereo",
			"internal/cte/schemas/v4_0/modal_aquaviario",
			"internal/cte/schemas/v4_0/modal_dutoviario",
			"internal/cte/schemas/v4_0/modal_ferroviario",
			"internal/cte/schemas/v4_0/modal_rodoviario",
			"internal/cte/schemas/v4_0/modal_rodoviario_os",
			"internal/cte/schemas/v4_0/modal_multimodal",
			"internal/cte/schemas/v4_0/evento_cce",
			"internal/cte/schemas/v4_0/evento_cancel",
			"internal/cte/schemas/v4_0/evento_ce",
			"internal/cte/schemas/v4_0/evento_cancel_ce",
			"internal/cte/schemas/v4_0/evento_cancel_ie",
			"internal/cte/schemas/v4_0/evento_cancel_prest_desacordo",
			"internal/cte/schemas/v4_0/evento_epec",
			"internal/cte/schemas/v4_0/evento_gtv",
			"internal/cte/schemas/v4_0/evento_ie",
			"internal/cte/schemas/v4_0/evento_prest_desacordo",
			"internal/cte/schemas/v4_0/evento_reg_multimodal",
			"internal/cte/schemas/v1_0/dist_dfe",
		},
		xgenJobs: []xgenJob{
			{"v1_0/dist_dfe", "v1_0/dist_dfe"},
			{"v4_0/cte", "v4_0/cte"},
			{"v4_0/cte_os", "v4_0/cte_os"},
			{"v4_0/consulta_situacao", "v4_0/consulta_situacao"},
			{"v4_0/status_servico", "v4_0/status_servico"},
			{"v4_0/cte_simp", "v4_0/cte_simp"},
			{"v4_0/gtve", "v4_0/gtve"},
			{"v4_0/modal_aereo/cteModalAereo_v4.00.xsd", "v4_0/modal_aereo/cteModalAereo_v4.00.xsd.go"},
			{"v4_0/modal_aquaviario/cteModalAquaviario_v4.00.xsd", "v4_0/modal_aquaviario/cteModalAquaviario_v4.00.xsd.go"},
			{"v4_0/modal_dutoviario/cteModalDutoviario_v4.00.xsd", "v4_0/modal_dutoviario/cteModalDutoviario_v4.00.xsd.go"},
			{"v4_0/modal_ferroviario/cteModalFerroviario_v4.00.xsd", "v4_0/modal_ferroviario/cteModalFerroviario_v4.00.xsd.go"},
			{"v4_0/modal_rodoviario/cteModalRodoviario_v4.00.xsd", "v4_0/modal_rodoviario/cteModalRodoviario_v4.00.xsd.go"},
			{"v4_0/modal_rodoviario_os/cteModalRodoviarioOS_v4.00.xsd", "v4_0/modal_rodoviario_os/cteModalRodoviarioOS_v4.00.xsd.go"},
			{"v4_0/modal_multimodal/cteMultiModal_v4.00.xsd", "v4_0/modal_multimodal/cteMultiModal_v4.00.xsd.go"},
			{"v4_0/evento_cce", "v4_0/evento_cce"},
			{"v4_0/evento_cancel", "v4_0/evento_cancel"},
			{"v4_0/evento_ce", "v4_0/evento_ce"},
			{"v4_0/evento_cancel_ce", "v4_0/evento_cancel_ce"},
			{"v4_0/evento_cancel_ie", "v4_0/evento_cancel_ie"},
			{"v4_0/evento_cancel_prest_desacordo", "v4_0/evento_cancel_prest_desacordo"},
			{"v4_0/evento_epec", "v4_0/evento_epec"},
			{"v4_0/evento_gtv", "v4_0/evento_gtv"},
			{"v4_0/evento_ie", "v4_0/evento_ie"},
			{"v4_0/evento_prest_desacordo", "v4_0/evento_prest_desacordo"},
			{"v4_0/evento_reg_multimodal", "v4_0/evento_reg_multimodal"},
		},
		postprocess: postprocessCTe,
	}
}

func mdfeFamily() family {
	return family{
		name: "mdfe",
		normalizeDirs: []string{
			"internal/mdfe/schemas/v3_0/mdfe",
			"internal/mdfe/schemas/v3_0/cons_nao_enc",
			"internal/mdfe/schemas/v3_0/cons_reci",
			"internal/mdfe/schemas/v3_0/consulta_situacao",
			"internal/mdfe/schemas/v3_0/status_servico",
			"internal/mdfe/schemas/v3_0/evento",
			"internal/mdfe/schemas/v3_0/evento_cancel",
			"internal/mdfe/schemas/v3_0/evento_enc",
			"internal/mdfe/schemas/v3_0/evento_inc_condutor",
			"internal/mdfe/schemas/v3_0/evento_inclusao_dfe",
			"internal/mdfe/schemas/v3_0/evento_pagto_oper",
			"internal/mdfe/schemas/v3_0/evento_alteracao_pagto_serv",
			"internal/mdfe/schemas/v3_0/evento_confirma_serv",
			"internal/mdfe/schemas/v3_0/dist_mdfe",
			"internal/mdfe/schemas/v3_0/consulta_dfe",
			"internal/mdfe/schemas/v1_0/dist_dfe",
		},
		xgenJobs: []xgenJob{
			{"v1_0/dist_dfe", "v1_0/dist_dfe"},
			{"v3_0/mdfe", "v3_0/mdfe"},
			{"v3_0/cons_nao_enc", "v3_0/cons_nao_enc"},
			{"v3_0/cons_reci", "v3_0/cons_reci"},
			{"v3_0/consulta_situacao", "v3_0/consulta_situacao"},
			{"v3_0/status_servico", "v3_0/status_servico"},
			{"v3_0/evento", "v3_0/evento"},
			{"v3_0/evento_cancel", "v3_0/evento_cancel"},
			{"v3_0/evento_enc", "v3_0/evento_enc"},
			{"v3_0/evento_inc_condutor", "v3_0/evento_inc_condutor"},
			{"v3_0/evento_inclusao_dfe", "v3_0/evento_inclusao_dfe"},
			{"v3_0/evento_pagto_oper", "v3_0/evento_pagto_oper"},
			{"v3_0/evento_alteracao_pagto_serv", "v3_0/evento_alteracao_pagto_serv"},
			{"v3_0/evento_confirma_serv", "v3_0/evento_confirma_serv"},
			{"v3_0/dist_mdfe", "v3_0/dist_mdfe"},
			{"v3_0/consulta_dfe", "v3_0/consulta_dfe"},
		},
		postprocess: postprocessMDFe,
	}
}

func nfeFamily() family {
	return family{
		name: "nfe",
		normalizeDirs: []string{
			"internal/nfe/schemas/v1_0/dist_dfe",
			"internal/nfe/schemas/v1_0/ator_interessado",
			"internal/nfe/schemas/v1_0/evento_entrega",
			"internal/nfe/schemas/v1_0/evento_cancel_entrega",
			"internal/nfe/schemas/v1_0/evento_cancel",
			"internal/nfe/schemas/v1_0/evento_cce",
			"internal/nfe/schemas/v1_0/evento_generico",
			"internal/nfe/schemas/v1_0/evento_mde",
			"internal/nfe/schemas/v1_0/evento_insucesso",
			"internal/nfe/schemas/v1_0/evento_cancel_insucesso",
			"internal/nfe/schemas/v1_0/epec",
			"internal/nfe/schemas/v2_0/cons",
			"internal/nfe/schemas/v4_0/nfe_proc",
			"internal/nfe/schemas/v4_0/consulta_situacao",
			"internal/nfe/schemas/v4_0/status_servico",
			"internal/nfe/schemas/v4_0/inutilizacao",
		},
		xgenJobs: []xgenJob{
			{"v1_0/ator_interessado", "v1_0/ator_interessado"},
			{"v1_0/evento_entrega", "v1_0/evento_entrega"},
			{"v1_0/evento_cancel_entrega", "v1_0/evento_cancel_entrega"},
			{"v1_0/evento_cancel", "v1_0/evento_cancel"},
			{"v1_0/evento_cce", "v1_0/evento_cce"},
			{"v1_0/evento_generico", "v1_0/evento_generico"},
			{"v1_0/evento_mde", "v1_0/evento_mde"},
			{"v1_0/evento_insucesso", "v1_0/evento_insucesso"},
			{"v1_0/evento_cancel_insucesso", "v1_0/evento_cancel_insucesso"},
			{"v1_0/epec", "v1_0/epec"},
			{"v1_0/dist_dfe", "v1_0/dist_dfe"},
			{"v2_0/cons", "v2_0/cons"},
			{"v4_0/nfe_proc", "v4_0/nfe_proc"},
			{"v4_0/consulta_situacao", "v4_0/consulta_situacao"},
			{"v4_0/status_servico", "v4_0/status_servico"},
			{"v4_0/inutilizacao", "v4_0/inutilizacao"},
		},
		postprocess: postprocessNFe,
	}
}

func nfseFamily() family {
	return family{
		name: "nfse",
		xgenJobs: []xgenJob{
			{"v1_0/core", "v1_0/core"},
		},
		postprocess: postprocessNFSe,
	}
}

func postprocessBPe(verbose bool) error {
	return postprocess.Generated(postprocess.Options{
		GenDir: "internal/bpe/gen",
		NestedImportPatterns: []string{
			pathPattern("internal", "bpe", "schemas"),
			pathPattern("internal", "bpe", "gen", "v1_0", "schemas"),
		},
		Replacements: []postprocess.Replacement{
			postprocess.ReplaceAll("*interface{}", "*string"),
			postprocess.ReplaceAll("interface{}", "string"),
			postprocess.ReplaceAll(infBPeCompField, infBPeCompFieldFixed),
			postprocess.IfPath(
				func(path string) bool {
					return strings.HasSuffix(path, string(filepath.Separator)+"eventoBPeTiposBasico_v1.00.xsd.go")
				},
				postprocess.Replace(bpeDetEventoStruct, bpeDetEventoStructInnerXML, 1),
			),
			postprocess.IfPath(
				func(path string) bool {
					return strings.HasSuffix(path, string(filepath.Separator)+"evNaoEmbBPe_v1.00.xsd.go")
				},
				postprocess.RegexReplaceAll(nProtTProtField, "\n\tNProt string `xml:\"nProt\"`"),
			),
		},
		AddJSONTags: true,
		Verbose:     verbose,
	})
}

func postprocessCTe(verbose bool) error {
	return postprocess.Generated(postprocess.Options{
		GenDir: "internal/cte/gen",
		NestedImportPatterns: []string{
			pathPattern("nfelib", "nfelib", "cte", "schemas"),
			pathPattern("internal", "cte", "schemas"),
			pathPattern("internal", "cte", "gen", "v4_0", "schemas"),
		},
		RemoveFile: func(path string) (bool, string) {
			return isDiscardedCTeModalSupportFile(path), "removed modal support schema package"
		},
		Replacements: []postprocess.Replacement{
			postprocess.ReplaceAll("*interface{}", "*string"),
			postprocess.ReplaceAll("interface{}", "string"),
			postprocess.RegexReplaceAll(optionalFieldDhCont, "\n\tDhCont         *string `xml:\"dhCont\"`"),
			postprocess.RegexReplaceAll(optionalFieldXJust, "\n\tXJust          *string `xml:\"xJust\"`"),
			postprocess.RegexReplaceAll(optionalFieldCRT, "\n\tCRT       *string   `xml:\"CRT\"`"),
			postprocess.IfPath(
				func(path string) bool {
					return hasAnySuffix(path,
						string(filepath.Separator)+"evCancCTe_v4.00.xsd.go",
						string(filepath.Separator)+"evCancCECTe_v4.00.xsd.go",
						string(filepath.Separator)+"evCancIECTe_v4.00.xsd.go",
					)
				},
				postprocess.RegexReplaceAll(nProtTProtField, "\n\tNProt string `xml:\"nProt\"`"),
			),
			postprocess.IfPath(
				func(path string) bool {
					return strings.HasSuffix(path, string(filepath.Separator)+"evCancPrestDesacordo_v4.00.xsd.go")
				},
				postprocess.RegexReplaceAll(nProtPrestDesField, "\n\tNProtEvPrestDes string `xml:\"nProtEvPrestDes\"`"),
			),
			replaceTypedCTeEventPayloads,
		},
		AddJSONTags: true,
		Verbose:     verbose,
	})
}

func postprocessMDFe(verbose bool) error {
	return postprocess.Generated(postprocess.Options{
		GenDir: "internal/mdfe/gen",
		NestedImportPatterns: []string{
			pathPattern("nfelib", "nfelib", "mdfe", "schemas"),
			pathPattern("internal", "mdfe", "schemas"),
			pathPattern("internal", "mdfe", "gen", "v3_0", "schemas"),
		},
		Replacements: []postprocess.Replacement{
			postprocess.ReplaceAll("*TpAmb", "*string"),
			postprocess.Replace(mdfeInfModalStruct, mdfeInfModalStructInnerXML, 1),
			postprocess.Replace(mdfeAnonInfModalStruct, mdfeAnonInfModalInnerXML, 1),
			postprocess.IfPath(
				func(path string) bool {
					return hasAnySuffix(path,
						string(filepath.Separator)+"evCancMDFe_v3.00.xsd.go",
						string(filepath.Separator)+"evConfirmaServMDFe_v3.00.xsd.go",
						string(filepath.Separator)+"evEncMDFe_v3.00.xsd.go",
					)
				},
				postprocess.RegexReplaceAll(nProtTProtField, "\n\tNProt string `xml:\"nProt\"`"),
			),
			postprocess.IfPath(
				func(path string) bool {
					return strings.HasSuffix(path, string(filepath.Separator)+"eventoMDFeTiposBasico_v3.00.xsd.go")
				},
				postprocess.Replace(mdfeDetEventoStruct, mdfeDetEventoInnerXML, 1),
				postprocess.Replace(mdfeAnonDetEventoStruct, mdfeAnonDetEventoInnerXML, 1),
			),
		},
		AddJSONTags: true,
		Verbose:     verbose,
	})
}

func postprocessNFe(verbose bool) error {
	return postprocess.Generated(postprocess.Options{
		GenDir: "internal/nfe/gen",
		NestedImportPatterns: []string{
			pathPattern("internal", "nfe", "schemas"),
			pathPattern("internal", "nfe", "gen", "v1_0", "schemas"),
			pathPattern("internal", "nfe", "gen", "v4_0", "schemas"),
		},
		RemoveFile: func(path string) (bool, string) {
			return isDuplicateGeneratedNFeFragment(path), "removed duplicated imported schema package"
		},
		Replacements: []postprocess.Replacement{
			postprocess.ReplaceAll("*interface{}", "*string"),
			postprocess.ReplaceAll("interface{}", "string"),
			postprocess.IfPath(
				postprocess.PathContains("internal", "nfe", "gen", "v1_0", "evento_cce"),
				postprocess.ReplaceAll("*TCOrgaoIBGE", "*string"),
				postprocess.ReplaceAll("*TVerEvento", "*string"),
			),
		},
		AddJSONTags: true,
		Verbose:     verbose,
	})
}

func postprocessNFSe(verbose bool) error {
	return postprocess.Generated(postprocess.Options{
		GenDir: "internal/nfse/gen",
		NestedImportPatterns: []string{
			pathPattern("nfelib", "nfelib", "nfse", "schemas"),
			pathPattern("internal", "nfse", "schemas"),
			pathPattern("internal", "nfse", "gen", "v1_0", "schemas"),
		},
		Replacements: []postprocess.Replacement{
			postprocess.ReplaceAll(xDescInterfaceField, xDescStringField),
		},
		AddJSONTags: true,
		Verbose:     verbose,
	})
}

func replaceTypedCTeEventPayloads(path, updated string) string {
	if strings.HasSuffix(path, string(filepath.Separator)+"cteTiposBasico_v4.00.xsd.go") && usesTypedCTeInfModal(path) {
		if strings.Contains(path, string(filepath.Separator)+"cte_os"+string(filepath.Separator)) {
			updated = strings.Replace(updated, xmlOnlyImportBlock, typedInfModalImportBlockOS, 1)
			updated = strings.Replace(updated, cteInfModalStruct, typedCTeInfModal("InfModal", true), 1)
			updated = strings.Replace(updated, cteAnonInfModalStruct, typedCTeInfModal("TAnonComplexInfModal1", true), 1)
			updated = strings.Replace(updated, cteAnonInfModalStruct2, typedCTeInfModal("TAnonComplexInfModal2", true), 1)
			updated = strings.Replace(updated, cteAnonInfModalStruct3, typedCTeInfModal("TAnonComplexInfModal3", true), 1)
		} else {
			updated = strings.Replace(updated, xmlOnlyImportBlock, typedInfModalImportBlock, 1)
			updated = strings.Replace(updated, cteInfModalStruct, typedCTeInfModal("InfModal", false), 1)
			updated = strings.Replace(updated, cteAnonInfModalStruct, typedCTeInfModal("TAnonComplexInfModal1", false), 1)
			updated = strings.Replace(updated, cteAnonInfModalStruct2, typedCTeInfModal("TAnonComplexInfModal2", false), 1)
			updated = strings.Replace(updated, cteAnonInfModalStruct3, typedCTeInfModal("TAnonComplexInfModal3", true), 1)
		}
	}
	if strings.HasSuffix(path, string(filepath.Separator)+"eventoCTeTiposBasico_v4.00.xsd.go") {
		for folder, element := range cteEventPayloads {
			if strings.Contains(path, string(filepath.Separator)+folder+string(filepath.Separator)) {
				updated = strings.Replace(updated, cteDetEventoStruct, cteDetEventoReplacement("DetEvento", element), 1)
				updated = strings.Replace(updated, cteAnonDetEventoStruct, cteDetEventoReplacement("TAnonComplexDetEvento1", element), 1)
				break
			}
		}
	}
	return updated
}

func isDiscardedCTeModalSupportFile(path string) bool {
	if filepath.Base(filepath.Dir(path)) == "v4_0" && strings.HasPrefix(filepath.Base(path), "modal_") {
		return true
	}

	parent := filepath.Base(filepath.Dir(path))
	allowed := map[string]string{
		"modal_aereo":         "cteModalAereo_v4.00.xsd.go",
		"modal_aquaviario":    "cteModalAquaviario_v4.00.xsd.go",
		"modal_dutoviario":    "cteModalDutoviario_v4.00.xsd.go",
		"modal_ferroviario":   "cteModalFerroviario_v4.00.xsd.go",
		"modal_rodoviario":    "cteModalRodoviario_v4.00.xsd.go",
		"modal_rodoviario_os": "cteModalRodoviarioOS_v4.00.xsd.go",
		"modal_multimodal":    "cteMultiModal_v4.00.xsd.go",
	}

	rootFile, ok := allowed[parent]
	if !ok {
		return false
	}
	return filepath.Base(path) != rootFile
}

func usesTypedCTeInfModal(path string) bool {
	parent := filepath.Base(filepath.Dir(path))
	for _, folder := range []string{"cte", "cte_os", "cte_simp", "gtve"} {
		if parent == folder {
			return true
		}
	}
	return false
}

func typedCTeInfModal(typeName string, osOnly bool) string {
	if osOnly {
		return fmt.Sprintf(
			"type %s struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tRodoOS          modalrodoviarioos.RodoOS `xml:\"rodoOS,omitempty\"`\n}",
			typeName,
		)
	}
	return fmt.Sprintf(
		"type %s struct {\n\tXMLName         xml.Name `xml:\"infModal\"`\n\tVersaoModalAttr string   `xml:\"versaoModal,attr\"`\n\tRodo            modalrodoviario.Rodo `xml:\"rodo,omitempty\"`\n\tAereo           modalaereo.Aereo `xml:\"aereo,omitempty\"`\n\tAquav           modalaquaviario.Aquav `xml:\"aquav,omitempty\"`\n\tFerrov          modalferroviario.Ferrov `xml:\"ferrov,omitempty\"`\n\tDuto            modaldutoviario.Duto `xml:\"duto,omitempty\"`\n\tMultimodal      modalmultimodal.Multimodal `xml:\"multimodal,omitempty\"`\n}",
		typeName,
	)
}

func cteDetEventoReplacement(typeName, element string) string {
	typeSuffix := strings.ToUpper(element[:1]) + element[1:]
	return fmt.Sprintf(
		"type %s struct {\n\tXMLName          xml.Name `xml:\"detEvento\"`\n\tVersaoEventoAttr string   `xml:\"versaoEvento,attr\"`\n\t%s         *TAnonComplex%s1 `xml:\"%s\"`\n}",
		typeName,
		typeSuffix,
		typeSuffix,
		element,
	)
}

func isDuplicateGeneratedNFeFragment(path string) bool {
	clean := filepath.Clean(path)
	base := filepath.Base(clean)

	switch {
	case strings.Contains(clean, pathPattern("internal", "nfe", "gen", "v1_0", "ator_interessado")):
		return base == "110150_v1.00.xsd.go"
	case strings.Contains(clean, pathPattern("internal", "nfe", "gen", "v1_0", "evento_mde")):
		return base == "e210200_v1.00.xsd.go" ||
			base == "e210210_v1.00.xsd.go" ||
			base == "e210220_v1.00.xsd.go" ||
			base == "e210240_v1.00.xsd.go"
	case strings.Contains(clean, pathPattern("internal", "nfe", "gen", "v1_0", "evento_insucesso")):
		return base == "tmp0000.xsd.go"
	}

	return false
}

func pathPattern(elem ...string) string {
	return string(filepath.Separator) + filepath.Join(elem...) + string(filepath.Separator)
}

func hasAnySuffix(path string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}
