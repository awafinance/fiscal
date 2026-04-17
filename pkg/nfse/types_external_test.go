package nfse_test

import (
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/awafinance/fiscal/pkg/nfse"
)

func acceptsDPS(*nfse.TCDPS)                {}
func acceptsPedRegEvento(*nfse.TCPedRegEvt) {}

var _ = func(doc nfse.Document) {
	acceptsDPS(doc.DPS)
	acceptsPedRegEvento(doc.PedRegEvento)
}

var (
	_ info.AdditionalInformation = (*nfse.Document)(nil)
	_ info.AmountsInfo           = (*nfse.Document)(nil)
)
