package gen

//go:generate xgen -i ../schemas/v4_0/nfe_proc -o ./v4_0/nfe_proc -l Go
//go:generate xgen -i ../schemas/v4_0/consulta_situacao -o ./v4_0/consulta_situacao -l Go
//go:generate xgen -i ../schemas/v4_0/status_servico -o ./v4_0/status_servico -l Go
//go:generate xgen -i ../schemas/v4_0/inutilizacao -o ./v4_0/inutilizacao -l Go
//go:generate go run ../tools/codegen postprocess-generated
