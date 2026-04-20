package bpe_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	cnpj := "12345678000195"
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
	require.Contains(t, doc.GetAmounts(), info.Amount{Type: "ticket", Value: "120.50"})
	require.Contains(t, doc.GetParties(), info.Party{Role: "buyer", Name: "PASSAGEIRO TESTE", Document: "12345678000195"})
}

func TestDocumentGetEmitterDetail(t *testing.T) {
	data, err := os.ReadFile("../../testdata/bpe/v1_0/43180107364617000135630000000000081000000087-bpe.xml")
	require.NoError(t, err)

	doc, err := bpe.Parse(data)
	require.NoError(t, err)

	detail := doc.GetEmitterDetail()
	require.NotNil(t, detail)
	require.Empty(t, detail.TradeName)
	require.Equal(t, "1111111111", detail.IE)
	require.Equal(t, "516830", detail.IM)
	require.Equal(t, "1234567", detail.CNAE)
	require.Equal(t, "1", detail.CRT)
	require.Empty(t, detail.Phone)
	require.Empty(t, detail.Email)

	require.NotNil(t, detail.Address)
	require.Equal(t, "RUA ANTONIO DURO", detail.Address.Street)
	require.Equal(t, "870", detail.Address.Number)
	require.Equal(t, "CENTRO", detail.Address.Neighborhood)
	require.Equal(t, "4303509", detail.Address.CityCode)
	require.Equal(t, "CAMAQUA", detail.Address.CityName)
	require.Equal(t, "RS", detail.Address.State)
}

func TestDocumentGetEmitterDetailHandlesNilDocument(t *testing.T) {
	var doc *bpe.Document
	require.Nil(t, doc.GetEmitterDetail())

	require.Nil(t, (&bpe.Document{}).GetEmitterDetail())
}
