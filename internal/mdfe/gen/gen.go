package gen

//go:generate go run ../tools/codegen normalize-schemas internal/mdfe/schemas/v3_0/mdfe internal/mdfe/schemas/v3_0/cons_nao_enc internal/mdfe/schemas/v3_0/cons_reci internal/mdfe/schemas/v3_0/evento internal/mdfe/schemas/v1_0/dist_dfe
//go:generate xgen -i ../schemas/v1_0/dist_dfe -o ./v1_0/dist_dfe -l Go
//go:generate xgen -i ../schemas/v3_0/mdfe -o ./v3_0/mdfe -l Go
//go:generate xgen -i ../schemas/v3_0/cons_nao_enc -o ./v3_0/cons_nao_enc -l Go
//go:generate xgen -i ../schemas/v3_0/cons_reci -o ./v3_0/cons_reci -l Go
//go:generate xgen -i ../schemas/v3_0/evento -o ./v3_0/evento -l Go
//go:generate go run ../tools/codegen postprocess-generated
