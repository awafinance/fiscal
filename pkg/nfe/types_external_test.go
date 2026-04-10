package nfe_test

import "github.com/awafinance/fiscal/pkg/nfe"

func acceptsNFe(*nfe.NFeProcTNFe)         {}
func acceptsEventoCCe(*nfe.EventoCCeTEvento) {}

var _ = func(doc nfe.Document) {
	acceptsNFe(doc.NFe)
	acceptsEventoCCe(doc.EventoCCe)
}
