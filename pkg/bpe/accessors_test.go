package bpe_test

import (
	"testing"

	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	cnpj := "12345678000195"
	vTotTrib := "12.34"
	doc := &bpe.Document{
		BPe: &bpe.TBPe{
			InfBPe: &bpe.TAnonComplexInfBPe2{
				IdAttr: "BPe43190812345678000195630010000000011000000011",
				Ide: &bpe.TAnonComplexIde2{
					Mod:   "63",
					Serie: "1",
					NBP:   "1",
					DhEmi: "2019-08-01T10:00:00-03:00",
				},
				Emit: &bpe.TAnonComplexEmit2{
					CNPJ:  cnpj,
					XNome: "TRANSPORTADORA TESTE",
				},
				Comp: &bpe.TAnonComplexComp12{
					XNome: "PASSAGEIRO TESTE",
					CNPJ:  &cnpj,
				},
				InfValorBPe: &bpe.TAnonComplexInfValorBPe1{VBP: "120.50"},
				Imp: &bpe.TAnonComplexImp2{
					VTotTrib: &vTotTrib,
					ICMS: &bpe.TImp{
						ICMS00: &bpe.TAnonComplexICMS001{CST: "00", VBC: "120.50", PICMS: "10.00", VICMS: "12.05"},
					},
				},
			},
		},
	}

	require.Equal(t, "43190812345678000195630010000000011000000011", doc.GetAccessKey())
	require.Empty(t, doc.GetVersion())
	require.Empty(t, doc.GetEnvironment())
	require.Equal(t, "1", doc.GetNumber())
	require.Equal(t, "1", doc.GetSeries())
	require.Equal(t, "63", doc.GetModel())
	require.Equal(t, "2019-08-01T10:00:00-03:00", doc.GetIssueDate())
	require.Equal(t, "120.50", doc.GetAmount())
	require.Equal(t, "TRANSPORTADORA TESTE", doc.GetIssuer())
	require.Equal(t, "12345678000195", doc.GetIssuerDocument())
	require.Equal(t, "PASSAGEIRO TESTE", doc.GetRecipient())
	require.Equal(t, "12345678000195", doc.GetRecipientDocument())
	require.False(t, doc.IsAuthorized())
	amounts := doc.GetAmounts()
	require.Contains(t, amounts, info.Amount{Type: "ticket", Value: "120.50"})
	require.Contains(t, amounts, info.Amount{Type: "taxes", Value: "12.34"})
	require.Contains(t, amounts, info.Amount{Type: "tax_icms", Value: "12.05"})
	require.Contains(t, doc.GetParties(), info.Party{Role: "buyer", Name: "PASSAGEIRO TESTE", Document: "12345678000195"})
}
