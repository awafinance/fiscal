package fiscal

import "github.com/awafinance/fiscal/pkg/fiscalerr"

// Sentinel errors re-exported from pkg/fiscalerr. These share identity with
// the sentinels in the per-family packages, so errors.Is works regardless
// of which family produced the error.
var (
	ErrEmptyDocument        = fiscalerr.ErrEmptyDocument
	ErrUnsupportedNamespace = fiscalerr.ErrUnsupportedNamespace
	ErrUnsupportedRoot      = fiscalerr.ErrUnsupportedRoot
)

// Typed errors re-exported as aliases so they refer to the same underlying
// type; errors.As works with either name.
type (
	UnsupportedNamespaceError = fiscalerr.UnsupportedNamespaceError
	UnsupportedRootError      = fiscalerr.UnsupportedRootError
)
