package cte_test

import (
	"github.com/awafinance/fiscal/pkg/cte"
	"github.com/awafinance/fiscal/pkg/info"
)

func acceptsCTe(*cte.CTeCTe)                                {}
func acceptsEventoCancCTe(*cte.EventoCancCTeEvento)         {}
func acceptsEventoGenericoCTe(*cte.EventoGenericoCTeEvento) {}
func acceptsCTeOS(*cte.CTeOSCTeOS)                          {}

var _ = func(doc cte.Document) {
	acceptsCTe(doc.CTe)
	acceptsEventoCancCTe(doc.EventoCancCTe)
	acceptsEventoGenericoCTe(doc.EventoGenerico)
	acceptsCTeOS(doc.CTeOS)
}

var (
	_ info.BillingInfo           = (*cte.Document)(nil)
	_ info.AdditionalInformation = (*cte.Document)(nil)
	_ info.AmountsInfo           = (*cte.Document)(nil)
)
