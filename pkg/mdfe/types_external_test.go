package mdfe_test

import "github.com/awafinance/fiscal/pkg/mdfe"

func acceptsMDFe(*mdfe.MDFeTMDFe)              {}
func acceptsEventoCancel(*mdfe.EventoCancelTEvento) {}

var _ = func(doc mdfe.Document) {
	acceptsMDFe(doc.MDFe)
	acceptsEventoCancel(doc.EventoCancMDFe)
}
