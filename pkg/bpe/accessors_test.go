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

func TestDocumentGetEmitterDetailFromFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/bpe/v1_0/43190812345678000195630010000000011000000011-bpeProc.xml")
	require.NoError(t, err)

	doc, err := bpe.Parse(data)
	require.NoError(t, err)

	detail := doc.GetEmitterDetail()
	require.NotNil(t, detail)
	require.Equal(t, "12345678000195", doc.GetIssuerDocument())
	require.Nil(t, detail.Address)
}

func TestDocumentGetEmitterDetail(t *testing.T) {
	xFant := "TRANSPORTE RAPIDO"
	fone := "11999887766"
	email := "contato@transporte.com"
	doc := &bpe.Document{
		BPe: &bpe.TBPe{
			InfBPe: &bpe.TAnonComplexInfBPe2{
				Emit: &bpe.TAnonComplexEmit2{
					CNPJ:  "12345678000195",
					XNome: "TRANSPORTADORA TESTE",
					XFant: &xFant,
					IE:    "123456789",
					IM:    "54321",
					CNAE:  "4930202",
					CRT:   "3",
					EnderEmit: &bpe.TEndeEmi{
						XLgr:    "AV BRASIL",
						Nro:     "100",
						XBairro: "CENTRO",
						CMun:    "3550308",
						XMun:    "SAO PAULO",
						UF:      "SP",
						Fone:    &fone,
						Email:   &email,
					},
				},
			},
		},
	}

	detail := doc.GetEmitterDetail()
	require.NotNil(t, detail)
	require.Equal(t, "TRANSPORTE RAPIDO", detail.TradeName)
	require.Equal(t, "123456789", detail.IE)
	require.Equal(t, "54321", detail.IM)
	require.Equal(t, "4930202", detail.CNAE)
	require.Equal(t, "3", detail.CRT)
	require.Equal(t, "11999887766", detail.Phone)
	require.Equal(t, "contato@transporte.com", detail.Email)

	require.NotNil(t, detail.Address)
	require.Equal(t, "AV BRASIL", detail.Address.Street)
	require.Equal(t, "SAO PAULO", detail.Address.CityName)
	require.Equal(t, "SP", detail.Address.State)
}

func TestDocumentGetEmitterDetailHandlesNilDocument(t *testing.T) {
	var doc *bpe.Document
	require.Nil(t, doc.GetEmitterDetail())

	require.Nil(t, (&bpe.Document{}).GetEmitterDetail())
}
