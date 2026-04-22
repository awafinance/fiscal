// Package fiscalerr holds the sentinel and typed errors returned by the
// top-level fiscal package and the per-family parsers. Consumers typically
// import them through github.com/awafinance/fiscal (see the aliases there);
// this package exists so the family packages and the top-level package can
// share a single identity without an import cycle.
package fiscalerr

import (
	"errors"
	"fmt"
)

// Family identifies one of the supported fiscal document families. It is
// defined here (rather than in the top-level fiscal package) so that typed
// errors can carry a Family field without creating an import cycle with the
// per-family parsers.
type Family string

const (
	NFe  Family = "nfe"
	NFSe Family = "nfse"
	CTe  Family = "cte"
	MDFe Family = "mdfe"
	BPe  Family = "bpe"
)

// Sentinel errors. Callers match them with errors.Is. The typed errors
// below Unwrap to the matching sentinel, so either concrete shape works.
// The sentinel messages are context-free; callers wrap them with fmt.Errorf
// to add a "parse <family>:" prefix.
var (
	ErrEmptyDocument        = errors.New("empty xml document")
	ErrUnsupportedNamespace = errors.New("unsupported namespace")
	ErrUnsupportedRoot      = errors.New("unsupported root element")
)

// UnsupportedNamespaceError is returned when the XML root belongs to a
// namespace that is not one of the supported fiscal families. Its Error()
// text is context-free; callers wrap with fmt.Errorf to add a prefix.
type UnsupportedNamespaceError struct {
	Namespace string
	Root      string
}

func (e *UnsupportedNamespaceError) Error() string {
	return fmt.Sprintf("unsupported namespace %q root %q", e.Namespace, e.Root)
}

func (e *UnsupportedNamespaceError) Unwrap() error {
	return ErrUnsupportedNamespace
}

// UnsupportedRootError is returned when the XML root namespace matches a
// supported family but the root element itself is not a known envelope for
// that family. Its Error() text is context-free; callers wrap with
// fmt.Errorf to add a prefix.
type UnsupportedRootError struct {
	Family Family
	Root   string
}

func (e *UnsupportedRootError) Error() string {
	return fmt.Sprintf("unsupported root element %q", e.Root)
}

func (e *UnsupportedRootError) Unwrap() error {
	return ErrUnsupportedRoot
}
