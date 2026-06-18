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

The root package can parse any supported fiscal XML by detecting the root
namespace:

```go
import "github.com/awafinance/fiscal"
```

It exposes the main entrypoint:

```go
Parse(data []byte) (*Document, error)
ParseReader(r io.Reader) (*Document, error)
```

The returned `fiscal.Document` is a tagged container with `Family`,
`RootName`, and one populated family document field:

```go
doc, err := fiscal.Parse(data)
if err != nil {
 panic(err)
}

switch doc.Family {
case fiscal.NFe:
 fmt.Println(doc.NFe.GetAccessKey())
case fiscal.MDFe:
 fmt.Println(doc.MDFe.GetRelatedDocuments())
}
```

Each document family is also exposed through its own package when callers
already know the family or need the family-specific typed API:

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
 "fmt"
 "os"

 "github.com/awafinance/fiscal"
)

func main() {
 data, err := os.ReadFile("document.xml")
 if err != nil {
  panic(err)
 }

 doc, err := fiscal.Parse(data)
 if err != nil {
  panic(err)
 }

 info := doc.Info()
 fmt.Println(info.GetAccessKey())
 fmt.Println(info.GetIssuer())
 fmt.Println(info.GetAmount())
}
```

## Document dispatch

`fiscal.Parse` first reads the XML root element namespace, routes the document
to the matching family package, and then unmarshals into the corresponding
schema-generated type.

For families that use generic event envelopes, parsing also dispatches by `tpEvento`, so event documents can be represented with concrete event-specific structs instead of a single generic event shape.

Examples of supported dispatch patterns:

- NF-e: invoice, processed invoice, consultation/status/inutilizacao roots, distribution, concrete events
- CT-e: document variants, processed wrappers, modal/event variants, distribution
- MDF-e: consultation roots, processed roots, distribution roots, concrete events
- BP-e: base documents and concrete event variants

## Convenience accessors

All supported family documents implement a common `DocumentInfo` interface.
The root `fiscal.Document.Info()` method returns that interface:

```go
type DocumentInfo interface {
 GetAccessKey() string
 GetVersion() string
 GetEnvironment() string
 GetNumber() string
 GetSeries() string
 GetModel() string
 GetIssueDate() string
 GetAmount() string
 GetIssuer() string
 GetIssuerDocument() string
 GetRecipient() string
 GetRecipientDocument() string
 GetProtocolNumber() string
 GetStatusCode() string
 GetStatusReason() string
 IsAuthorized() bool
}
```

These methods return the XML values as strings. The library does not parse,
round, sum, or normalize monetary values.

```go
doc, err := fiscal.Parse(data)
if err != nil {
 panic(err)
}

info := doc.Info()
fmt.Println(info.GetAccessKey())
fmt.Println(info.GetVersion())
fmt.Println(info.GetEnvironment())
fmt.Println(info.GetAmount())
fmt.Println(info.GetIssuerDocument())
```

Some fiscal concepts are not universal. For those, the library exposes small
optional interfaces. Consumers can type-assert only the detail they need:

```go
if amounts, ok := doc.Info().(fiscal.AmountsInfo); ok {
 for _, amount := range amounts.GetAmounts() {
  fmt.Println(amount.Type, amount.Value)
 }
}

if parties, ok := doc.Info().(fiscal.PartiesInfo); ok {
 for _, party := range parties.GetParties() {
  fmt.Println(party.Role, party.Name, party.Document)
 }
}

if related, ok := doc.Info().(fiscal.RelatedDocumentsInfo); ok {
 for _, ref := range related.GetRelatedDocuments() {
  fmt.Println(ref.Type, ref.AccessKey)
 }
}

if route, ok := doc.Info().(fiscal.RouteInfo); ok {
 fmt.Println(route.GetModal())
 fmt.Println(route.GetOrigin())
 fmt.Println(route.GetDestination())
}

if event, ok := doc.Info().(fiscal.LifecycleEventInfo); ok {
 fmt.Println(event.GetEventType())        // raw tpEvento ("110111") or NFS-e code ("e101101")
 fmt.Println(event.GetEventSequence())    // nSeqEvento
 fmt.Println(doc.Info().GetAccessKey())   // the referenced note
 fmt.Println(doc.Info().GetStatusCode())  // SEFAZ return, when the event carries one
}

if facts, ok := doc.Info().(fiscal.LifecycleEventFactsInfo); ok {
 if event := facts.GetLifecycleEventFacts(); event != nil {
  fmt.Println(event.RegistrationState) // "request" or "registered"
  fmt.Println(event.RequestNumber)     // nPedRegEvento
  fmt.Println(event.ProcessingTime)    // dhProc on registered events
 }
}
```

Optional interface support is intentionally grouped by concept:

- `AmountsInfo` returns raw amount fields such as NFe total, CTe service value,
  MDFe cargo value, BPe ticket value, and NFSe service/net values.
- `PartiesInfo` returns known parties with roles, such as issuer, recipient,
  provider, taker, sender, dispatcher, receiver, and buyer.
- `RelatedDocumentsInfo` returns document references such as linked NFe, CTe,
  MDF-e, or DCe access keys where the schema carries them. This also covers
  correction back-references: a CT-e complemento/substituto points at the
  original CT-e (`Type: "cte"`), and an NFS-e substitute note points at the
  superseded note (`Type: "nfse"`).
- `RouteInfo` returns modal, origin, and destination fields for transport and
  service documents where those concepts exist.
- `LifecycleEventInfo` returns the raw event type and sequence for event
  documents (cancelamento, substituição, manifestação, ...). The referenced
  note, date, and SEFAZ return ride the base accessors (`GetAccessKey`,
  `GetIssueDate`, `GetStatusCode`). `GetEventType` is never translated to a
  friendly name.
- `LifecycleEventFactsInfo` returns one normalized event fact record where the
  schema needs more than type and sequence. NFS-e uses this to distinguish a
  bare `pedRegEvento` request (`RegistrationState: "request"`) from a generated
  `evento` (`RegistrationState: "registered"`), and to expose `nPedRegEvento`
  and `dhProc` without inventing NF-e/CT-e status semantics.

The `pkg/info` package contains the shared structs and optional interface
definitions. The root package re-exports them as aliases, so callers can use
`fiscal.Amount`, `fiscal.Party`, `fiscal.RelatedDocument`, `fiscal.Location`,
and `fiscal.LifecycleEventFacts`.

### NFe Details

NFe also exposes convenience methods for line items and payments:

```go
for _, item := range doc.NFe.GetItems() {
 fmt.Println(item.Number, item.Code, item.Description, item.Amount)
}

for _, payment := range doc.NFe.GetPayments() {
 fmt.Println(payment.Method, payment.Amount)
}
```

These methods also return the raw XML string values.

### NFS-e Details

NFS-e exposes the accounting competence date, which determines which
period the invoice belongs to (distinct from the issue date):

```go
fmt.Println(doc.NFSe.GetCompetenceDate())
```

For NFS-e, `IsAuthorized()` is true for generated NFS-e status codes `100`,
`101`, `102`, `103`, and `107`. Event documents keep `GetStatusCode()` empty;
use `LifecycleEventFactsInfo` for NFS-e event registration facts.

### CT-e Details

CT-e exposes the raw document type (`tpCTe`): `0` normal, `1` complemento,
`3` substituto. The referenced original (for complemento/substituto) is surfaced
through `GetRelatedDocuments()`.

```go
fmt.Println(doc.CTe.GetType())
```

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

The repository also includes a small CLI for parsing fiscal XML documents. The family is detected automatically from the root namespace.

Run it from the repository root with:

```bash
go run ./cmd --help
go run ./cmd <xml>           # human-readable summary
go run ./cmd <xml> --json    # full typed document as JSON
```

The default output prints a summary built from the shared accessors
(access key, version, number/series, issuer, recipient, amount,
status, etc.) plus any extra sections the family exposes (amounts,
parties, route, related documents).

For example:

```bash
go run ./cmd testdata/nfe/35180834128745000152550010000476121675985748-nfe.xml
go run ./cmd --json testdata/cte/v4_0/43120178408960000182570010000000041000000047-cte.xml | jq '.cte'
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
