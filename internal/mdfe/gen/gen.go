package gen

//go:generate go run ../tools/codegen normalize-schemas internal/mdfe/schemas/v3_0/mdfe internal/mdfe/schemas/v3_0/cons_nao_enc internal/mdfe/schemas/v3_0/cons_reci internal/mdfe/schemas/v3_0/consulta_situacao internal/mdfe/schemas/v3_0/status_servico internal/mdfe/schemas/v3_0/evento internal/mdfe/schemas/v3_0/evento_cancel internal/mdfe/schemas/v3_0/evento_enc internal/mdfe/schemas/v3_0/evento_inc_condutor internal/mdfe/schemas/v3_0/evento_inclusao_dfe internal/mdfe/schemas/v3_0/evento_pagto_oper internal/mdfe/schemas/v3_0/evento_alteracao_pagto_serv internal/mdfe/schemas/v3_0/evento_confirma_serv internal/mdfe/schemas/v3_0/dist_mdfe internal/mdfe/schemas/v3_0/consulta_dfe internal/mdfe/schemas/v1_0/dist_dfe
//go:generate xgen -i ../schemas/v1_0/dist_dfe -o ./v1_0/dist_dfe -l Go
//go:generate xgen -i ../schemas/v3_0/mdfe -o ./v3_0/mdfe -l Go
//go:generate xgen -i ../schemas/v3_0/cons_nao_enc -o ./v3_0/cons_nao_enc -l Go
//go:generate xgen -i ../schemas/v3_0/cons_reci -o ./v3_0/cons_reci -l Go
//go:generate xgen -i ../schemas/v3_0/consulta_situacao -o ./v3_0/consulta_situacao -l Go
//go:generate xgen -i ../schemas/v3_0/status_servico -o ./v3_0/status_servico -l Go
//go:generate xgen -i ../schemas/v3_0/evento -o ./v3_0/evento -l Go
//go:generate xgen -i ../schemas/v3_0/evento_cancel -o ./v3_0/evento_cancel -l Go
//go:generate xgen -i ../schemas/v3_0/evento_enc -o ./v3_0/evento_enc -l Go
//go:generate xgen -i ../schemas/v3_0/evento_inc_condutor -o ./v3_0/evento_inc_condutor -l Go
//go:generate xgen -i ../schemas/v3_0/evento_inclusao_dfe -o ./v3_0/evento_inclusao_dfe -l Go
//go:generate xgen -i ../schemas/v3_0/evento_pagto_oper -o ./v3_0/evento_pagto_oper -l Go
//go:generate xgen -i ../schemas/v3_0/evento_alteracao_pagto_serv -o ./v3_0/evento_alteracao_pagto_serv -l Go
//go:generate xgen -i ../schemas/v3_0/evento_confirma_serv -o ./v3_0/evento_confirma_serv -l Go
//go:generate xgen -i ../schemas/v3_0/dist_mdfe -o ./v3_0/dist_mdfe -l Go
//go:generate xgen -i ../schemas/v3_0/consulta_dfe -o ./v3_0/consulta_dfe -l Go
//go:generate go run ../tools/codegen postprocess-generated
