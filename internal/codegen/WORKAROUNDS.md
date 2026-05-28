# Codegen Workarounds

This directory contains the normalization and post-processing layer that adapts
`xgen` output into stable bindings for each document family.

The current generator version is pinned in [`mise.toml`](../../mise.toml) as:

- `github.com/xuri/xgen/cmd/xgen@v0.0.0-20260109003923-38cc1baf99f3`

## Why this exists

Some schema patterns are normalized before generation because `xgen` does not
emit usable Go for them. Other issues are corrected after generation because
the emitted Go has family-specific problems such as:

- `interface{}` fields where the XML payload is actually a string
- root aliases emitted as `string` instead of typed pointers
- event payload wrappers that need `innerxml` preservation
- duplicate generated files for imported fragments

## Current workaround inventory

### Shared patterns

- Replace generated `interface{}` fields with `string` in BPe, CTe, and NFe.
- Fix generated root aliases by reading the source XSD and rewriting `type Foo string` into `type Foo *TBar`.
- Add JSON tags after post-processing so generated structs remain usable across XML and JSON APIs.

### BPe

- Fix `Comp` field type mismatch in generated BPe core bindings.
- Preserve `detEvento` as `innerxml` in `eventoBPeTiposBasico_v1.00.xsd.go`.
- Rewrite selected `NProt *TProt` event fields to `NProt string`.

### CTe

- Remove modal support files that are generated as duplicate or incomplete fragments.
- Preserve `detEvento` as `innerxml` for the generic CT-e event package.
- Rewrite `DhCont`, `XJust`, and `CRT` to optional `*string` fields.
- Rewrite selected `NProt *TProt` and `NProtEvPrestDes *TProt` fields to `string`.
- Replace placeholder event payload structs with typed modal/event payloads.

### MDFe

- Rewrite `*TpAmb` to `*string`.
- Preserve `infModal` as `innerxml` in generated root and anonymous structs.
- Preserve `detEvento` as `innerxml` in `eventoMDFeTiposBasico_v3.00.xsd.go`.
- Rewrite selected `NProt *TProt` fields to `string`.

### NFe

- Remove duplicate generated imported fragments for known event families.
- Rewrite cancel-event `NProt *TProt` to `string`.
- Rewrite unsupported `evento_cce` enums to `*string`.

### NFSe

- Rewrite `xDesc interface{}` to `xDesc string`.

## Refactoring direction

The current code is now organized as explicit named workarounds, but most rules
still operate as text transformations. The next steps should be:

1. Add tests for normalization helpers and postprocess rules.
2. Convert the highest-risk Go rewrites from text replacement to AST-based rules.
3. Keep text replacements only for generator-shape compatibility issues that
   cannot be expressed more safely.
