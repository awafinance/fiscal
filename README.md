# fiscal

`fiscal` is a Go library for parsing and marshaling Brazilian fiscal XML
documents with strongly typed structs generated from the adapted
official XSD schemas.

It currently covers:

- NF-e
- NFS-e
- CT-e
- MDF-e
- BP-e

The library supports primary documents, processed wrappers, distribution
responses, service/status and consultation roots where implemented, and
typed event documents where applicable.

## Setup

```bash
git clone git@github.com:awafinance/fiscal.git
cd fiscal
mise trust && mise install
```

## Packages

Each document family is exposed through its own package:

```go
import "github.com/awafinance/fiscal/pkg/nfe"
import "github.com/awafinance/fiscal/pkg/nfse"
import "github.com/awafinance/fiscal/pkg/cte"
import "github.com/awafinance/fiscal/pkg/mdfe"
import "github.com/awafinance/fiscal/pkg/bpe"
```

Each package exposes the same core entrypoint:

```go
Parse(data []byte) (*Document, error)
ParseReader(r io.Reader) (*Document, error)
```

The returned `Document` is a tagged container where exactly one root field is
expected to be populated.

The generated XML schema types are also part of the public API through each
family package, so external consumers can build and pass typed values without
importing `internal/...`.

```go
var inf *nfe.NFeProcTNFe
doc := &nfe.Document{NFe: inf}
```

## Basic usage

```go
package main

import (
 "encoding/xml"
 "fmt"
 "os"

 "github.com/awafinance/fiscal/pkg/nfe"
)

func main() {
 data, err := os.ReadFile("nfe.xml")
 if err != nil {
  panic(err)
 }

 doc, err := nfe.Parse(data)
 if err != nil {
  panic(err)
 }

 ...
}
```

## Document dispatch

`Parse` first detects the XML root element and then unmarshals into the corresponding schema-generated type.

For families that use generic event envelopes, parsing also dispatches by `tpEvento`, so event documents can be represented with concrete event-specific structs instead of a single generic event shape.

Examples of supported dispatch patterns:

- NF-e: invoice, processed invoice, consultation/status/inutilizacao roots, distribution, concrete events
- CT-e: document variants, processed wrappers, modal/event variants, distribution
- MDF-e: consultation roots, processed roots, distribution roots, concrete events
- BP-e: base documents and concrete event variants

## Round-trip behavior

Parsed documents can be marshaled back through the standard library encoder:

```go
out, err := xml.MarshalIndent(doc, "", "  ")
```

The library implements custom `MarshalXML` logic so the original supported root is preserved.

JSON output also includes `rootName` so parsed documents can preserve their original root selection when marshaled back to XML after a JSON round-trip.

## Limitations

- A `Document` must contain exactly one supported root field. Setting multiple root fields is invalid.
- XML round-tripping preserves the parsed root only when `rootName` is available. Parsed documents populate it automatically, and JSON output now carries it as well.
- NF-e marshaling emits `nfeProc` when protocol data is present. A document with `ProtNFe` is not re-encoded as bare `NFe`.
- The supported typed API is the alias surface exported from `pkg/<family>/types.go`. Depending directly on `internal/...` generated packages is not supported.

## CLI

The repository also includes a small CLI for parsing fiscal XML documents and writing the parsed document as JSON.

Run it from the repository root with:

```bash
go run ./cmd --help
go run ./cmd nfe <xml>
go run ./cmd nfse <xml>
go run ./cmd cte <xml>
go run ./cmd mdfe <xml>
go run ./cmd bpe <xml>
```

Each document command takes one XML file path. For example:

```bash
go run ./cmd nfe testdata/nfe/35180834128745000152550010000476121675985748-nfe.xml
```

## Project structure

- `pkg/<family>` contains the public parsing API
- `internal/<family>/schemas` contains normalized XSDs used for generation
- `internal/<family>/gen` contains generated Go bindings
- `internal/<family>/tools/codegen` contains normalization and post-processing helpers

Generation is schema-driven. The generated code is post-processed to fix XML namespace handling, normalize anonymous types, and keep marshal/unmarshal behavior stable across families.

## Development

Regenerate bindings for a family with:

```bash
mise run gen
```

Run tests with:

```bash
mise run test
```

Check all commands available:

```bash
mise tasks
```

## Acknowledgements

This project was inspired by [nfelib](https://github.com/akretion/nfelib).
Thanks to the `nfelib` maintainers for the reference implementation and
for the test files that helped shape and validate this library.

## License

[Apache-2.0](LICENSE)
