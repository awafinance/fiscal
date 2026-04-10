package cte_test

import "github.com/awafinance/fiscal/pkg/cte"

func acceptsCTe(*cte.CTeCTe)                        {}
func acceptsEventoCancCTe(*cte.EventoCancCTeEvento) {}
func acceptsCTeOS(*cte.CTeOSCTeOS)                  {}

var _ = func(doc cte.Document) {
	acceptsCTe(doc.CTe)
	acceptsEventoCancCTe(doc.EventoCancCTe)
	acceptsCTeOS(doc.CTeOS)
}
