package fiscal

import (
	"bytes"
	"fmt"
	"io"

	"github.com/awafinance/fiscal/internal/monetary"
	"github.com/awafinance/fiscal/internal/xmlutil"
	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/awafinance/fiscal/pkg/cte"
	"github.com/awafinance/fiscal/pkg/fiscalerr"
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/mdfe"
	"github.com/awafinance/fiscal/pkg/nfe"
	"github.com/awafinance/fiscal/pkg/nfse"
)

// Family is aliased to fiscalerr.Family so the typed-error Family field and
// the Document.Family field share a single underlying type.
type Family = fiscalerr.Family

const (
	NFe  = fiscalerr.NFe
	NFSe = fiscalerr.NFSe
	CTe  = fiscalerr.CTe
	MDFe = fiscalerr.MDFe
	BPe  = fiscalerr.BPe
)

const (
	nfeNamespace  = "http://www.portalfiscal.inf.br/nfe"
	nfseNamespace = "http://www.sped.fazenda.gov.br/nfse"
	cteNamespace  = "http://www.portalfiscal.inf.br/cte"
	mdfeNamespace = "http://www.portalfiscal.inf.br/mdfe"
	bpeNamespace  = "http://www.portalfiscal.inf.br/bpe"
)

type Document struct {
	Family   Family `json:"family"`
	RootName string `json:"rootName,omitempty"`

	info     DocumentInfo
	monetary *info.MonetaryInterpretation

	NFe  *nfe.Document  `json:"nfe,omitempty"`
	NFSe *nfse.Document `json:"nfse,omitempty"`
	CTe  *cte.Document  `json:"cte,omitempty"`
	MDFe *mdfe.Document `json:"mdfe,omitempty"`
	BPe  *bpe.Document  `json:"bpe,omitempty"`
}

func Parse(data []byte) (*Document, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("parse fiscal: %w", fiscalerr.ErrEmptyDocument)
	}
	root, err := xmlutil.ParseRootElement(data)
	if err != nil {
		return nil, fmt.Errorf("parse fiscal: read root: %w", err)
	}

	family, ok := familyForNamespace(root.Space)
	if !ok {
		return nil, fmt.Errorf("parse fiscal: %w", &fiscalerr.UnsupportedNamespaceError{Namespace: root.Space, Root: root.Local})
	}
	interpretation, err := monetary.Interpret(data, family, root.Local)
	if err != nil {
		return nil, fmt.Errorf("parse fiscal: %w", err)
	}

	var wrapped *Document
	switch root.Space {
	case nfeNamespace:
		doc, parseErr := nfe.Parse(data)
		wrapped, err = wrapNFe(doc, parseErr)
	case nfseNamespace:
		doc, parseErr := nfse.Parse(data)
		wrapped, err = wrapNFSe(doc, parseErr)
	case cteNamespace:
		doc, parseErr := cte.Parse(data)
		wrapped, err = wrapCTe(doc, parseErr)
	case mdfeNamespace:
		doc, parseErr := mdfe.Parse(data)
		wrapped, err = wrapMDFe(doc, parseErr)
	case bpeNamespace:
		doc, parseErr := bpe.Parse(data)
		wrapped, err = wrapBPe(doc, parseErr)
	}
	if err != nil {
		return nil, err
	}
	wrapped.monetary = interpretation
	return wrapped, nil
}

func familyForNamespace(namespace string) (Family, bool) {
	switch namespace {
	case nfeNamespace:
		return NFe, true
	case nfseNamespace:
		return NFSe, true
	case cteNamespace:
		return CTe, true
	case mdfeNamespace:
		return MDFe, true
	case bpeNamespace:
		return BPe, true
	default:
		return "", false
	}
}

func ParseReader(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("parse fiscal: read xml: %w", err)
	}
	return Parse(data)
}

func wrapNFe(doc *nfe.Document, err error) (*Document, error) {
	if err != nil {
		return nil, err
	}
	return &Document{Family: NFe, RootName: doc.RootName, info: doc, NFe: doc}, nil
}

func wrapNFSe(doc *nfse.Document, err error) (*Document, error) {
	if err != nil {
		return nil, err
	}
	return &Document{Family: NFSe, RootName: doc.RootName, info: doc, NFSe: doc}, nil
}

func wrapCTe(doc *cte.Document, err error) (*Document, error) {
	if err != nil {
		return nil, err
	}
	return &Document{Family: CTe, RootName: doc.RootName, info: doc, CTe: doc}, nil
}

func wrapMDFe(doc *mdfe.Document, err error) (*Document, error) {
	if err != nil {
		return nil, err
	}
	return &Document{Family: MDFe, RootName: doc.RootName, info: doc, MDFe: doc}, nil
}

func wrapBPe(doc *bpe.Document, err error) (*Document, error) {
	if err != nil {
		return nil, err
	}
	return &Document{Family: BPe, RootName: doc.RootName, info: doc, BPe: doc}, nil
}
