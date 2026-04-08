package gen

//go:generate go run ../tools/codegen normalize-schemas internal/bpe/schemas/v1_0/core
//go:generate xgen -i ../schemas/v1_0/core -o ./v1_0/core -l Go
//go:generate go run ../tools/codegen postprocess-generated
