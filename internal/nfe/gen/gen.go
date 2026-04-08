package gen

//go:generate go run ../tools/codegen normalize-schemas internal/nfe/schemas/v1_0/dist_dfe
//go:generate xgen -i ../schemas/v1_0/evento_entrega -o ./v1_0/evento_entrega -l Go
//go:generate xgen -i ../schemas/v1_0/evento_cancel_entrega -o ./v1_0/evento_cancel_entrega -l Go
//go:generate xgen -i ../schemas/v1_0/evento_cancel -o ./v1_0/evento_cancel -l Go
//go:generate xgen -i ../schemas/v1_0/evento_cce -o ./v1_0/evento_cce -l Go
//go:generate xgen -i ../schemas/v1_0/epec -o ./v1_0/epec -l Go
//go:generate xgen -i ../schemas/v1_0/dist_dfe -o ./v1_0/dist_dfe -l Go
//go:generate xgen -i ../schemas/v4_0/nfe_proc -o ./v4_0/nfe_proc -l Go
//go:generate xgen -i ../schemas/v4_0/consulta_situacao -o ./v4_0/consulta_situacao -l Go
//go:generate xgen -i ../schemas/v4_0/status_servico -o ./v4_0/status_servico -l Go
//go:generate xgen -i ../schemas/v4_0/inutilizacao -o ./v4_0/inutilizacao -l Go
//go:generate go run ../tools/codegen postprocess-generated
