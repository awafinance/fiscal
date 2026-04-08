package gen

//go:generate xgen -i ../schemas/v1_0/core -o ./v1_0/core -l Go
//go:generate go run ../tools/codegen postprocess-generated
