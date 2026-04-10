package bpe_test

import "github.com/awafinance/fiscal/pkg/bpe"

func acceptsBPe(*bpe.TBPe)            {}
func acceptsCancel(*bpe.CancelTEvento) {}

var _ = func(doc bpe.Document) {
	acceptsBPe(doc.BPe)
	acceptsCancel(doc.EventoCancBPe)
}
