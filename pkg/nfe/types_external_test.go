package nfe_test

import (
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfe"
)

func acceptsNFe(*nfe.NFeProcTNFe)            {}
func acceptsEventoCCe(*nfe.EventoCCeTEvento) {}

var _ = func(doc nfe.Document) {
	acceptsNFe(doc.NFe)
	acceptsEventoCCe(doc.EventoCCe)
}

var (
	_ info.PaymentsInfo          = (*nfe.Document)(nil)
	_ info.BillingInfo           = (*nfe.Document)(nil)
	_ info.AdditionalInformation = (*nfe.Document)(nil)
	_ info.AmountsInfo           = (*nfe.Document)(nil)
)
