package gen

//go:generate go run ../tools/codegen normalize-schemas internal/cte/schemas/v4_0/cte internal/cte/schemas/v4_0/cte_os internal/cte/schemas/v4_0/consulta_situacao internal/cte/schemas/v4_0/status_servico internal/cte/schemas/v4_0/cte_simp internal/cte/schemas/v4_0/gtve internal/cte/schemas/v4_0/modal_aereo internal/cte/schemas/v4_0/modal_aquaviario internal/cte/schemas/v4_0/modal_dutoviario internal/cte/schemas/v4_0/modal_ferroviario internal/cte/schemas/v4_0/modal_rodoviario internal/cte/schemas/v4_0/modal_rodoviario_os internal/cte/schemas/v4_0/modal_multimodal internal/cte/schemas/v4_0/evento_cce internal/cte/schemas/v4_0/evento_cancel internal/cte/schemas/v4_0/evento_ce internal/cte/schemas/v4_0/evento_cancel_ce internal/cte/schemas/v4_0/evento_cancel_ie internal/cte/schemas/v4_0/evento_cancel_prest_desacordo internal/cte/schemas/v4_0/evento_epec internal/cte/schemas/v4_0/evento_gtv internal/cte/schemas/v4_0/evento_ie internal/cte/schemas/v4_0/evento_prest_desacordo internal/cte/schemas/v4_0/evento_reg_multimodal internal/cte/schemas/v1_0/dist_dfe
//go:generate xgen -i ../schemas/v1_0/dist_dfe -o ./v1_0/dist_dfe -l Go
//go:generate xgen -i ../schemas/v4_0/cte -o ./v4_0/cte -l Go
//go:generate xgen -i ../schemas/v4_0/cte_os -o ./v4_0/cte_os -l Go
//go:generate xgen -i ../schemas/v4_0/consulta_situacao -o ./v4_0/consulta_situacao -l Go
//go:generate xgen -i ../schemas/v4_0/status_servico -o ./v4_0/status_servico -l Go
//go:generate xgen -i ../schemas/v4_0/cte_simp -o ./v4_0/cte_simp -l Go
//go:generate xgen -i ../schemas/v4_0/gtve -o ./v4_0/gtve -l Go
//go:generate xgen -i ../schemas/v4_0/modal_aereo/cteModalAereo_v4.00.xsd -o ./v4_0/modal_aereo -l Go
//go:generate xgen -i ../schemas/v4_0/modal_aquaviario/cteModalAquaviario_v4.00.xsd -o ./v4_0/modal_aquaviario -l Go
//go:generate xgen -i ../schemas/v4_0/modal_dutoviario/cteModalDutoviario_v4.00.xsd -o ./v4_0/modal_dutoviario -l Go
//go:generate xgen -i ../schemas/v4_0/modal_ferroviario/cteModalFerroviario_v4.00.xsd -o ./v4_0/modal_ferroviario -l Go
//go:generate xgen -i ../schemas/v4_0/modal_rodoviario/cteModalRodoviario_v4.00.xsd -o ./v4_0/modal_rodoviario -l Go
//go:generate xgen -i ../schemas/v4_0/modal_rodoviario_os/cteModalRodoviarioOS_v4.00.xsd -o ./v4_0/modal_rodoviario_os -l Go
//go:generate xgen -i ../schemas/v4_0/modal_multimodal/cteMultiModal_v4.00.xsd -o ./v4_0/modal_multimodal -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cce -o ./v4_0/evento_cce -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cancel -o ./v4_0/evento_cancel -l Go
//go:generate xgen -i ../schemas/v4_0/evento_ce -o ./v4_0/evento_ce -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cancel_ce -o ./v4_0/evento_cancel_ce -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cancel_ie -o ./v4_0/evento_cancel_ie -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cancel_prest_desacordo -o ./v4_0/evento_cancel_prest_desacordo -l Go
//go:generate xgen -i ../schemas/v4_0/evento_epec -o ./v4_0/evento_epec -l Go
//go:generate xgen -i ../schemas/v4_0/evento_gtv -o ./v4_0/evento_gtv -l Go
//go:generate xgen -i ../schemas/v4_0/evento_ie -o ./v4_0/evento_ie -l Go
//go:generate xgen -i ../schemas/v4_0/evento_prest_desacordo -o ./v4_0/evento_prest_desacordo -l Go
//go:generate xgen -i ../schemas/v4_0/evento_reg_multimodal -o ./v4_0/evento_reg_multimodal -l Go
//go:generate go run ../tools/codegen postprocess-generated
