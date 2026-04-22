package fiscal

import (
	"bytes"
	"fmt"
	"io"

	"github.com/awafinance/fiscal/internal/xmlutil"
	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/awafinance/fiscal/pkg/cte"
	"github.com/awafinance/fiscal/pkg/fiscalerr"
	"github.com/awafinance/fiscal/pkg/mdfe"
	"github.com/awafinance/fiscal/pkg/nfe"
	"github.com/awafinance/fiscal/pkg/nfse"
)

type Family string

const (
	NFe  Family = "nfe"
	NFSe Family = "nfse"
	CTe  Family = "cte"
	MDFe Family = "mdfe"
	BPe  Family = "bpe"
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

	info DocumentInfo

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

	switch root.Space {
	case nfeNamespace:
		doc, err := nfe.Parse(data)
		return wrapNFe(doc, err)
	case nfseNamespace:
		doc, err := nfse.Parse(data)
		return wrapNFSe(doc, err)
	case cteNamespace:
		doc, err := cte.Parse(data)
		return wrapCTe(doc, err)
	case mdfeNamespace:
		doc, err := mdfe.Parse(data)
		return wrapMDFe(doc, err)
	case bpeNamespace:
		doc, err := bpe.Parse(data)
		return wrapBPe(doc, err)
	default:
		return nil, &fiscalerr.UnsupportedNamespaceError{Namespace: root.Space, Root: root.Local}
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
