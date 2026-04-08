package gen

//go:generate go run ../tools/codegen normalize-schemas internal/cte/schemas/v4_0/cte internal/cte/schemas/v4_0/cte_os internal/cte/schemas/v4_0/evento_cce internal/cte/schemas/v4_0/evento_cancel internal/cte/schemas/v1_0/dist_dfe
//go:generate xgen -i ../schemas/v1_0/dist_dfe -o ./v1_0/dist_dfe -l Go
//go:generate xgen -i ../schemas/v4_0/cte -o ./v4_0/cte -l Go
//go:generate xgen -i ../schemas/v4_0/cte_os -o ./v4_0/cte_os -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cce -o ./v4_0/evento_cce -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cancel -o ./v4_0/evento_cancel -l Go
//go:generate go run ../tools/codegen postprocess-generated
