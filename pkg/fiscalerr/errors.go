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

// Sentinel errors. Callers use errors.Is to match them regardless of
// whether the concrete error is a plain wrap or one of the typed variants
// below.
var (
	ErrEmptyDocument        = errors.New("empty xml document")
	ErrUnsupportedNamespace = errors.New("unsupported namespace")
	ErrUnsupportedRoot      = errors.New("unsupported root element")
)

// UnsupportedNamespaceError is returned when the XML root belongs to a
// namespace that is not one of the supported fiscal families.
type UnsupportedNamespaceError struct {
	Namespace string
	Root      string
}

func (e *UnsupportedNamespaceError) Error() string {
	return fmt.Sprintf("parse fiscal: unsupported namespace %q root %q", e.Namespace, e.Root)
}

func (e *UnsupportedNamespaceError) Unwrap() error {
	return ErrUnsupportedNamespace
}

// UnsupportedRootError is returned when the XML root namespace matches a
// supported family but the root element itself is not a known envelope for
// that family.
type UnsupportedRootError struct {
	Family string
	Root   string
}

func (e *UnsupportedRootError) Error() string {
	return fmt.Sprintf("parse %s: unsupported root element %q", e.Family, e.Root)
}

func (e *UnsupportedRootError) Unwrap() error {
	return ErrUnsupportedRoot
}
