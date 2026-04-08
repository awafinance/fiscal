# Porting XSDs From `nfelib`

This repository does not consume upstream `nfelib` schema folders directly. We port schemas into internal, family-specific trees and generate Go bindings from those internal copies.

Current families in this repo:

- `internal/nfe` with public package `pkg/nfe`
- `internal/nfse` with public package `pkg/nfse`
- `internal/cte` with public package `pkg/cte`

This separation matters. NF-e, NFSe, CT-e, MDF-e, BP-e, and others are different document families. A new family must get its own `internal/<family>` and `pkg/<family>` tree. Do not add NFSe, CT-e, MDF-e, or BP-e schemas under `internal/nfe`.

## What “porting” means here

Porting is more than copying files:

1. Pick the correct document family and create or reuse `internal/<family>`.
2. Decide the schema package boundary for that family.
3. Copy every XSD dependency needed by that package into the internal schema folder.
4. Normalize copied XSDs when that family needs normalization before `xgen`.
5. Generate Go code with `xgen`.
6. Postprocess generated code to fix `xgen` output details we rely on.
7. Check in normalized XSDs and generated Go files together.
8. Add or update the public package in `pkg/<family>`.
9. Verify parsing and normalized round-trip behavior with fixtures.

## Family Boundaries

Use one internal tree per fiscal document family.

Examples:

- NF-e: `internal/nfe/...`, public package `pkg/nfe`
- NFSe: `internal/nfse/...`, public package `pkg/nfse`
- CT-e: `internal/cte/...`, public package `pkg/cte`

If we later port CT-e, MDF-e, or BP-e, they should follow the same pattern:

- `internal/cte` and `pkg/cte`
- `internal/mdfe` and `pkg/mdfe`
- `internal/bpe` and `pkg/bpe`

## Folder Layout Conventions

There are two patterns currently in use.

### NF-e style: operation-scoped folders

Use one folder per operation under `internal/nfe/schemas/<versao>`.

Examples already in the repo:

- `internal/nfe/schemas/v4_0/nfe_proc`
- `internal/nfe/schemas/v4_0/consulta_situacao`
- `internal/nfe/schemas/v4_0/status_servico`
- `internal/nfe/schemas/v4_0/inutilizacao`
- `internal/nfe/schemas/v1_0/evento_entrega`
- `internal/nfe/schemas/v1_0/evento_cancel_entrega`

For `nfe_entrega`, the practical split was:

- `evento_entrega`: `EventoEntregaNFe`, `envEventoEntregaNFe`, `retEventoEntregaNFe`, `procEventoEntregaNFe`, `leiauteEventoEntregaNFe`, `e110130`, plus shared dependencies
- `evento_cancel_entrega`: `EventoCancEntregaNFe`, `envEventoCancEntregaNFe`, `retEventoCancEntregaNFe`, `procEventoCancEntregaNFe`, `leiauteEventoCancEntregaNFe`, `e110131`, plus shared dependencies

### NFSe style: whole-set package folder

If the family schema set is already self-contained and we want a single generated package from the whole versioned set, use one folder such as:

- `internal/nfse/schemas/v1_0/core`

This is how `nfelib/nfelib/nfse/schemas/v1_0` is currently ported.

The important rule in both patterns is the same: every file referenced by `xs:include` or `xs:import` must exist inside the internal folder you generate from.

### CT-e style: operation-scoped lean folders

CT-e also uses operation-scoped folders, but the package boundary must stay lean.

Current CT-e packages in the repo:

- `internal/cte/schemas/v4_0/cte`
- `internal/cte/schemas/v4_0/cte_os`
- `internal/cte/schemas/v4_0/evento_cce`

Important CT-e rule:

- Do not copy modal XSDs such as rodoviário, multimodal, aquaviário, etc into the same generated package when `infModal` is already modeled as open content.
- For CT-e round-trip, preserve `infModal` inner XML in generated Go code instead of trying to generate all modal-specific types into the same package.

## Normalization Step

Normalization is family-specific.

### NF-e normalization

Run:

```bash
go run ./internal/nfe/tools/codegen normalize-schemas <schema-dir> [<schema-dir>...]
```

Today the NF-e normalizer performs the two transformations we needed for `xgen` to generate usable Go code:

1. Inline `xs:simpleType` and `xs:complexType` declarations inside `xs:element` are hoisted into named top-level types such as `TAnon_*` and `TAnonComplex_*`.
2. Optional `xs:sequence minOccurs="0"` blocks that contain only direct `xs:element` children are flattened by moving `minOccurs="0"` onto each child element.

Those transformations were exercised in the `nfe_entrega` port:

- `e110130_v1.00.xsd`
- `e110131_v1.00.xsd`
- `leiauteEventoEntregaNFe_v1.00.xsd`
- `leiauteEventoCancEntregaNFe_v1.00.xsd`
- `xmldsig-core-schema_v1.01.xsd`

### NFSe normalization

NFSe currently does not need a pre-generation schema normalization step. Its fixes are handled in postprocessing of generated Go code.

### CT-e normalization

Run:

```bash
go run ./internal/cte/tools/codegen normalize-schemas <schema-dir> [<schema-dir>...]
```

CT-e needs the same base normalization as NF-e:

1. Inline `xs:simpleType` and `xs:complexType` declarations inside `xs:element` are hoisted into named top-level `TAnon_*` and `TAnonComplex_*` types.
2. Optional `xs:sequence minOccurs="0"` blocks containing only direct `xs:element` children are flattened.

Without that step, `xgen` collapses repeated anonymous CT-e structures into the wrong Go types and valid fields disappear on round-trip.

## Generation Step

Generate from the internal folder, not from `nfelib`.

### NF-e

```bash
xgen -i internal/nfe/schemas/<versao>/<operacao> -o internal/nfe/gen/<versao>/<operacao> -l Go
go run ./internal/nfe/tools/codegen postprocess-generated
```

NF-e postprocessing currently fixes:

1. `ds:Signature` XML tags so generated structs marshal and unmarshal the XMLDSig namespace correctly.
2. `TAnonComplex_*` XML names so anonymous complex types keep the original element name in struct tags.
3. Duplicate nested packages that `xgen` may emit under paths like `internal/nfe/gen/.../internal/nfe/schemas/...`.

### NFSe

```bash
xgen -i internal/nfse/schemas/<versao>/core -o internal/nfse/gen/<versao>/core -l Go
go run ./internal/nfse/tools/codegen postprocess-generated
```

NFSe postprocessing currently fixes:

1. `ds:Signature` XML tags so generated structs marshal and unmarshal the XMLDSig namespace correctly.
2. `xDesc` fields that `xgen` emits as `interface{}` so event payloads round-trip correctly as strings.
3. Duplicate nested packages that `xgen` may emit under generated `.../schemas/...` paths.

### CT-e

```bash
go run ./internal/cte/tools/codegen normalize-schemas internal/cte/schemas/<versao>/<operacao> [internal/cte/schemas/<versao>/<operacao>...]
xgen -i internal/cte/schemas/<versao>/<operacao> -o internal/cte/gen/<versao>/<operacao> -l Go
go run ./internal/cte/tools/codegen postprocess-generated
```

CT-e postprocessing currently fixes:

1. `ds:Signature` XML tags so generated structs marshal and unmarshal the XMLDSig namespace correctly.
2. `TAnonComplex_*` XML names so hoisted anonymous complex types keep the original element name in struct tags.
3. `infModal` generated structs so modal payload XML is preserved via `,innerxml`.
4. CCe event `detEvento` payload binding so `evCCeCTe` decodes to the concrete generated type.
5. Optional CT-e scalar fields that `xgen` emits as non-pointer strings even though they are optional in practice, such as `dhCont`, `xJust`, and `CRT`.
6. Duplicate nested packages that `xgen` may emit under generated `.../schemas/...` paths.

## Checked-in Generator Entrypoints

Each family owns its own generator entrypoints.

Examples:

[`internal/nfe/gen/gen.go`](/Users/guarilha/Code/awa/nota-fiscal/internal/nfe/gen/gen.go)

```go
//go:generate xgen -i ../schemas/v1_0/evento_entrega -o ./v1_0/evento_entrega -l Go
//go:generate xgen -i ../schemas/v1_0/evento_cancel_entrega -o ./v1_0/evento_cancel_entrega -l Go
//go:generate go run ../tools/codegen postprocess-generated
```

[`internal/nfse/gen/gen.go`](/Users/guarilha/Code/awa/nota-fiscal/internal/nfse/gen/gen.go)

```go
//go:generate xgen -i ../schemas/v1_0/core -o ./v1_0/core -l Go
//go:generate go run ../tools/codegen postprocess-generated
```

[`internal/cte/gen/gen.go`](/Users/guarilha/Code/awa/nota-fiscal/internal/cte/gen/gen.go)

```go
//go:generate go run ../tools/codegen normalize-schemas internal/cte/schemas/v4_0/cte internal/cte/schemas/v4_0/cte_os internal/cte/schemas/v4_0/evento_cce
//go:generate xgen -i ../schemas/v4_0/cte -o ./v4_0/cte -l Go
//go:generate xgen -i ../schemas/v4_0/cte_os -o ./v4_0/cte_os -l Go
//go:generate xgen -i ../schemas/v4_0/evento_cce -o ./v4_0/evento_cce -l Go
//go:generate go run ../tools/codegen postprocess-generated
```

## Validated Examples

### NF-e: `nfe_entrega`

The exact sequence used for this port was:

```bash
mkdir -p internal/nfe/schemas/v1_0/evento_entrega internal/nfe/schemas/v1_0/evento_cancel_entrega

cp nfelib/nfelib/nfe_entrega/schemas/v1_0/{EventoEntregaNFe_v1.00.xsd,envEventoEntregaNFe_v1.00.xsd,retEventoEntregaNFe_v1.00.xsd,procEventoEntregaNFe_v1.00.xsd,leiauteEventoEntregaNFe_v1.00.xsd,e110130_v1.00.xsd,tiposBasico_v1.03.xsd,xmldsig-core-schema_v1.01.xsd} internal/nfe/schemas/v1_0/evento_entrega/

cp nfelib/nfelib/nfe_entrega/schemas/v1_0/{EventoCancEntregaNFe_v1.00.xsd,envEventoCancEntregaNFe_v1.00.xsd,retEventoCancEntregaNFe_v1.00.xsd,procEventoCancEntregaNFe_v1.00.xsd,leiauteEventoCancEntregaNFe_v1.00.xsd,e110131_v1.00.xsd,tiposBasico_v1.03.xsd,xmldsig-core-schema_v1.01.xsd} internal/nfe/schemas/v1_0/evento_cancel_entrega/

go run ./internal/nfe/tools/codegen normalize-schemas internal/nfe/schemas/v1_0/evento_entrega internal/nfe/schemas/v1_0/evento_cancel_entrega

xgen -i internal/nfe/schemas/v1_0/evento_entrega -o internal/nfe/gen/v1_0/evento_entrega -l Go
xgen -i internal/nfe/schemas/v1_0/evento_cancel_entrega -o internal/nfe/gen/v1_0/evento_cancel_entrega -l Go

go run ./internal/nfe/tools/codegen postprocess-generated
go test ./...
```

### NFSe: `v1_0/core`

The exact sequence used for this port was:

```bash
mkdir -p internal/nfse/schemas/v1_0/core

cp nfelib/nfelib/nfse/schemas/v1_0/*.xsd internal/nfse/schemas/v1_0/core/

xgen -i internal/nfse/schemas/v1_0/core -o internal/nfse/gen/v1_0/core -l Go

go run ./internal/nfse/tools/codegen postprocess-generated
go test ./pkg/nfse
```

### CT-e: `v4_0/{cte,cte_os,evento_cce}`

The current CT-e port uses lean schema bundles:

- `cte`: `cte_v4.00.xsd`, `cteTiposBasico_v4.00.xsd`, `tiposGeralCTe_v4.00.xsd`, `DFeTiposBasicos_v1.00.xsd`, `xmldsig-core-schema_v1.01.xsd`
- `cte_os`: `cteOS_v4.00.xsd`, `cteTiposBasico_v4.00.xsd`, `tiposGeralCTe_v4.00.xsd`, `DFeTiposBasicos_v1.00.xsd`, `xmldsig-core-schema_v1.01.xsd`
- `evento_cce`: `eventoCTe_v4.00.xsd`, `eventoCTeTiposBasico_v4.00.xsd`, `evCCeCTe_v4.00.xsd`, `tiposGeralCTe_v4.00.xsd`, `xmldsig-core-schema_v1.01.xsd`

Generation flow:

```bash
go generate ./internal/cte/gen
go test ./pkg/cte
```

CT-e-specific note:

- Do not add the modal XSD packages into these folders unless you are intentionally creating separate modal-specific generated packages. The current public CT-e package preserves `infModal` as inner XML and round-trips successfully without those modal schemas in the same package.

## Round-Trip Requirement

Porting is not complete when the structs merely parse.

For public packages, we require normalized structural round-trip behavior, matching the existing NF-e tests:

1. Parse fixture XML into the public `pkg/<family>` document type.
2. Marshal the parsed document with `encoding/xml`.
3. Normalize both original and marshaled XML tokens before comparing.
4. Require semantic equality after normalization.

Important:

- This is not “return the original bytes back to the caller”.
- Raw-byte replay is not a valid substitute for correct marshaling.
- If `xgen` emits lossy field types such as `interface{}` for schema-constrained text content, fix generation or postprocessing so the XML can round-trip structurally.

See [`pkg/nfe/nfe_test.go`](/Users/guarilha/Code/awa/nota-fiscal/pkg/nfe/nfe_test.go) and [`pkg/nfse/nfse_test.go`](/Users/guarilha/Code/awa/nota-fiscal/pkg/nfse/nfse_test.go).

## Review Checklist For Future Ports

Before considering a port complete, verify:

1. The schemas were placed under the correct family tree, not under an unrelated one.
2. Every `schemaLocation` inside the internal folder resolves locally.
3. Any required family-specific schema normalization succeeds.
4. `xgen` succeeds without generating broken imports or invalid struct tags.
5. Family-specific postprocessing reports the expected files.
6. The new folders are wired into the correct `internal/<family>/gen/gen.go`.
7. A public package exists or was updated under `pkg/<family>` when needed.
8. Fixture parsing works for the intended root documents.
9. Normalized marshal round-trip tests pass.
10. `go test ./...` passes, or the narrower package-level test scope is documented if the port is still in progress.
