# fiscal

`fiscal` is a Go library for parsing and marshaling Brazilian fiscal XML documents with strongly typed structs generated from the official schemas.

It currently covers:

- NF-e
- NFS-e
- CT-e
- MDF-e
- BP-e

The library supports primary documents, processed wrappers, distribution responses, service/status and consultation roots where implemented, and typed event documents where applicable.

## Why this library

- Strongly typed Go structs generated from official XSDs
- Family-specific parsing with root dispatch
- Event-specific dispatch for families that reuse generic event envelopes
- Round-trip XML support through `xml.Marshal` / `xml.MarshalIndent`
- No runtime dependencies

## Install

```bash
go get github.com/awa/fiscal
```

## Packages

Each document family is exposed through its own package:

```go
import "github.com/awa/fiscal/pkg/nfe"
import "github.com/awa/fiscal/pkg/nfse"
import "github.com/awa/fiscal/pkg/cte"
import "github.com/awa/fiscal/pkg/mdfe"
import "github.com/awa/fiscal/pkg/bpe"
```

Each package exposes the same core entrypoint:

```go
Parse(data []byte) (*Document, error)
```

The returned `Document` is a tagged container where exactly one root field is expected to be populated.

## Basic usage

```go
package main

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/awa/fiscal/pkg/nfe"
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

	switch {
	case doc.NFe != nil:
		fmt.Println(doc.NFe.InfNFe.Emit.XNome)
	case doc.ProtNFe != nil:
		fmt.Println(doc.ProtNFe.InfProt.CStat)
	}

	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		panic(err)
	}

	_ = out
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
go generate ./internal/nfe/gen
go generate ./internal/cte/gen
go generate ./internal/mdfe/gen
go generate ./internal/bpe/gen
go generate ./internal/nfse/gen
```

Run tests with:

```bash
go test ./...
```

## License

[Apache-2.0](LICENSE)
