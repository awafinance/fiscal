package gen

//go:generate go run ../tools/codegen normalize-schemas internal/bpe/schemas/v1_0/core internal/bpe/schemas/v1_0/evento_cancel internal/bpe/schemas/v1_0/evento_alteracao_poltrona internal/bpe/schemas/v1_0/evento_excesso_bagagem internal/bpe/schemas/v1_0/evento_nao_emb
//go:generate xgen -i ../schemas/v1_0/core -o ./v1_0/core -l Go
//go:generate xgen -i ../schemas/v1_0/evento_cancel -o ./v1_0/evento_cancel -l Go
//go:generate xgen -i ../schemas/v1_0/evento_alteracao_poltrona -o ./v1_0/evento_alteracao_poltrona -l Go
//go:generate xgen -i ../schemas/v1_0/evento_excesso_bagagem -o ./v1_0/evento_excesso_bagagem -l Go
//go:generate xgen -i ../schemas/v1_0/evento_nao_emb -o ./v1_0/evento_nao_emb -l Go
//go:generate go run ../tools/codegen postprocess-generated
