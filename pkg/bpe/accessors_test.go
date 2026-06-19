package bpe_test

import (
	"testing"

	"github.com/awafinance/fiscal/pkg/bpe"
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/stretchr/testify/require"
)

func TestDocumentConvenienceAccessors(t *testing.T) {
	cnpj := "12345678000195"
	cpf := "12345678901"
	buyerIE := "110042490114"
	cep := "01001000"
	emitPhone := "1133334444"
	emitEmail := "emit@example.com"
	buyerPhone := "1155556666"
	buyerEmail := "buyer@example.com"
	passengerPhone := "1177778888"
	passengerEmail := "passenger@example.com"
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
					IE:    "123456789",
					XNome: "TRANSPORTADORA TESTE",
					IM:    "98765",
					CRT:   "1",
					EnderEmit: &bpe.TEndeEmi{
						XLgr:    "RUA A",
						Nro:     "10",
						XBairro: "CENTRO",
						CMun:    "3550308",
						XMun:    "SAO PAULO",
						CEP:     &cep,
						UF:      "SP",
						Fone:    &emitPhone,
						Email:   &emitEmail,
					},
				},
				Comp: &bpe.TAnonComplexComp12{
					XNome: "PASSAGEIRO TESTE",
					CNPJ:  &cnpj,
					IE:    &buyerIE,
					EnderComp: &bpe.TEndereco{
						XLgr:    "AV B",
						Nro:     "20",
						XBairro: "BAIRRO",
						CMun:    "3304557",
						XMun:    "RIO DE JANEIRO",
						CEP:     &cep,
						UF:      "RJ",
						Fone:    &buyerPhone,
						Email:   &buyerEmail,
					},
				},
				InfPassagem: &bpe.TAnonComplexInfPassagem1{
					InfPassageiro: &bpe.TAnonComplexInfPassageiro1{
						XNome: "PASSAGEIRO TESTE",
						CPF:   &cpf,
						Fone:  &passengerPhone,
						Email: &passengerEmail,
					},
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
	issuer := requireParty(t, doc.GetParties(), "issuer")
	require.Equal(t, "123456789", issuer.StateRegistration)
	require.Equal(t, "98765", issuer.MunicipalRegistration)
	require.Equal(t, "1133334444", issuer.Phone)
	require.Equal(t, "emit@example.com", issuer.Email)
	require.Equal(t, "1", issuer.SimpleNationalOption)
	require.Equal(t, "01001000", issuer.Address.PostalCode)
	buyer := requireParty(t, doc.GetParties(), "buyer")
	require.Equal(t, "PASSAGEIRO TESTE", buyer.Name)
	require.Equal(t, "12345678000195", buyer.Document)
	require.Equal(t, "110042490114", buyer.StateRegistration)
	require.Equal(t, "1155556666", buyer.Phone)
	require.Equal(t, "buyer@example.com", buyer.Email)
	require.Equal(t, "3304557", buyer.Address.CityCode)
	passenger := requireParty(t, doc.GetParties(), "passenger")
	require.Equal(t, "PASSAGEIRO TESTE", passenger.Name)
	require.Equal(t, "12345678901", passenger.Document)
	require.Equal(t, "1177778888", passenger.Phone)
	require.Equal(t, "passenger@example.com", passenger.Email)
}

func requireParty(t *testing.T, parties []info.Party, role string) info.Party {
	t.Helper()
	for _, party := range parties {
		if party.Role == role {
			return party
		}
	}
	require.Failf(t, "party not found", "role %q in %#v", role, parties)
	return info.Party{}
}
