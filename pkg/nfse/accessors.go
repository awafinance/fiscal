package nfse

import (
	"strings"

	"github.com/awafinance/fiscal/pkg/info"
)

func (d *Document) GetAccessKey() string {
	switch {
	case d == nil:
		return ""
	case d.NFSe != nil && d.NFSe.InfNFSe != nil:
		return strings.TrimPrefix(d.NFSe.InfNFSe.IdAttr, "NFS")
	case d.DPS != nil && d.DPS.InfDPS != nil:
		return strings.TrimPrefix(d.DPS.InfDPS.IdAttr, "DPS")
	case d.PedRegEvento != nil && d.PedRegEvento.InfPedReg != nil:
		return d.PedRegEvento.InfPedReg.ChNFSe
	default:
		return ""
	}
}

func (d *Document) GetVersion() string {
	if d == nil {
		return ""
	}
	return d.VersaoAttr
}

func (d *Document) GetEnvironment() string {
	if inf := d.infDPS(); inf != nil {
		return inf.TpAmb
	}
	if d != nil && d.NFSe != nil && d.NFSe.InfNFSe != nil {
		return d.NFSe.InfNFSe.AmbGer
	}
	return ""
}

func (d *Document) GetNumber() string {
	switch {
	case d == nil:
		return ""
	case d.NFSe != nil && d.NFSe.InfNFSe != nil:
		return d.NFSe.InfNFSe.NNFSe
	case d.DPS != nil && d.DPS.InfDPS != nil:
		return d.DPS.InfDPS.NDPS
	default:
		return ""
	}
}

func (d *Document) GetSeries() string {
	if inf := d.infDPS(); inf != nil {
		return inf.Serie
	}
	return ""
}

func (d *Document) GetModel() string {
	return ""
}

func (d *Document) GetIssueDate() string {
	if inf := d.infDPS(); inf != nil {
		return inf.DhEmi
	}
	if d != nil && d.NFSe != nil && d.NFSe.InfNFSe != nil {
		return d.NFSe.InfNFSe.DhProc
	}
	return ""
}

func (d *Document) GetAmount() string {
	switch {
	case d == nil:
		return ""
	case d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.Valores != nil:
		return d.NFSe.InfNFSe.Valores.VLiq
	case d.DPS != nil && d.DPS.InfDPS != nil && d.DPS.InfDPS.Valores != nil && d.DPS.InfDPS.Valores.VServPrest != nil:
		return d.DPS.InfDPS.Valores.VServPrest.VServ
	default:
		return ""
	}
}

func (d *Document) GetIssuer() string {
	switch {
	case d == nil:
		return ""
	case d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.Emit != nil:
		return d.NFSe.InfNFSe.Emit.XNome
	case d.DPS != nil && d.DPS.InfDPS != nil && d.DPS.InfDPS.Prest != nil && d.DPS.InfDPS.Prest.XNome != nil:
		return *d.DPS.InfDPS.Prest.XNome
	default:
		return ""
	}
}

func (d *Document) GetIssuerDocument() string {
	switch {
	case d == nil:
		return ""
	case d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.Emit != nil:
		return firstStringPtr(d.NFSe.InfNFSe.Emit.CNPJ, d.NFSe.InfNFSe.Emit.CPF)
	case d.DPS != nil && d.DPS.InfDPS != nil && d.DPS.InfDPS.Prest != nil:
		return firstStringPtr(d.DPS.InfDPS.Prest.CNPJ, d.DPS.InfDPS.Prest.CPF, d.DPS.InfDPS.Prest.NIF)
	default:
		return ""
	}
}

func (d *Document) GetRecipient() string {
	if inf := d.infDPS(); inf != nil && inf.Toma != nil {
		return inf.Toma.XNome
	}
	return ""
}

func (d *Document) GetRecipientDocument() string {
	if inf := d.infDPS(); inf != nil && inf.Toma != nil {
		return firstStringPtr(inf.Toma.CNPJ, inf.Toma.CPF, inf.Toma.NIF)
	}
	return ""
}

func (d *Document) GetProtocolNumber() string {
	return ""
}

func (d *Document) GetStatusCode() string {
	if d != nil && d.NFSe != nil && d.NFSe.InfNFSe != nil {
		return d.NFSe.InfNFSe.CStat
	}
	return ""
}

func (d *Document) GetStatusReason() string {
	return ""
}

func (d *Document) IsAuthorized() bool {
	return d.GetStatusCode() == "100"
}

func (d *Document) GetAmounts() []info.Amount {
	if d == nil {
		return nil
	}

	amounts := d.headlineAmounts()
	amounts = append(amounts, d.retentionAmounts()...)

	if len(amounts) == 0 {
		return nil
	}
	return compactAmounts(amounts...)
}

func (d *Document) headlineAmounts() []info.Amount {
	if d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.Valores != nil {
		v := d.NFSe.InfNFSe.Valores
		return []info.Amount{
			{Type: "net", Value: v.VLiq},
			{Type: "iss", Value: stringPtrValue(v.VISSQN)},
			{Type: "retained", Value: stringPtrValue(v.VTotalRet)},
		}
	}
	if d.DPS != nil && d.DPS.InfDPS != nil && d.DPS.InfDPS.Valores != nil && d.DPS.InfDPS.Valores.VServPrest != nil {
		return []info.Amount{{Type: "service", Value: d.DPS.InfDPS.Valores.VServPrest.VServ}}
	}
	return nil
}

func (d *Document) retentionAmounts() []info.Amount {
	inf := d.infDPS()
	if inf == nil || inf.Valores == nil || inf.Valores.Trib == nil || inf.Valores.Trib.TribFed == nil {
		return nil
	}
	fed := inf.Valores.Trib.TribFed
	amounts := []info.Amount{
		{Type: "retained_irrf", Value: stringPtrValue(fed.VRetIRRF)},
		{Type: "retained_csll", Value: stringPtrValue(fed.VRetCSLL)},
		{Type: "retained_inss", Value: stringPtrValue(fed.VRetCP)},
	}
	if pc := fed.Piscofins; pc != nil && stringPtrValue(pc.TpRetPisCofins) == "1" {
		amounts = append(amounts,
			info.Amount{Type: "retained_pis", Value: stringPtrValue(pc.VPis)},
			info.Amount{Type: "retained_cofins", Value: stringPtrValue(pc.VCofins)},
		)
	}
	return amounts
}

func (d *Document) GetAdditionalInfo() string {
	values := make([]string, 0, 3)
	if inf := d.infDPS(); inf != nil && inf.Serv != nil {
		if inf.Serv.InfoCompl != nil {
			values = append(values, stringPtrValue(inf.Serv.InfoCompl.XInfComp))
		}
		if inf.Serv.CServ != nil {
			values = append(values, inf.Serv.CServ.XDescServ)
		}
	}
	if d != nil && d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.Valores != nil {
		values = append(values, stringPtrValue(d.NFSe.InfNFSe.Valores.XOutInf))
	}
	return joinNonEmpty("\n", values...)
}

func (d *Document) GetParties() []info.Party {
	return compactParties(
		info.Party{Role: "provider", Name: d.GetIssuer(), Document: d.GetIssuerDocument()},
		info.Party{Role: "taker", Name: d.GetRecipient(), Document: d.GetRecipientDocument()},
	)
}

func (d *Document) GetModal() string {
	return ""
}

func (d *Document) GetOrigin() info.Location {
	if inf := d.infDPS(); inf != nil && inf.Serv != nil && inf.Serv.LocPrest != nil {
		return info.Location{CountryCode: stringPtrValue(inf.Serv.LocPrest.CPaisPrestacao), CityCode: stringPtrValue(inf.Serv.LocPrest.CLocPrestacao)}
	}
	return info.Location{}
}

func (d *Document) GetDestination() info.Location {
	if d != nil && d.NFSe != nil && d.NFSe.InfNFSe != nil {
		return info.Location{CityCode: stringPtrValue(d.NFSe.InfNFSe.CLocIncid), CityName: stringPtrValue(d.NFSe.InfNFSe.XLocIncid)}
	}
	return info.Location{}
}

func (d *Document) infDPS() *TCInfDPS {
	switch {
	case d == nil:
		return nil
	case d.DPS != nil:
		return d.DPS.InfDPS
	case d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.DPS != nil:
		return d.NFSe.InfNFSe.DPS.InfDPS
	default:
		return nil
	}
}

func firstStringPtr(values ...*string) string {
	for _, value := range values {
		if value != nil && *value != "" {
			return *value
		}
	}
	return ""
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func compactAmounts(amounts ...info.Amount) []info.Amount {
	out := make([]info.Amount, 0, len(amounts))
	for _, amount := range amounts {
		if amount.Value != "" {
			out = append(out, amount)
		}
	}
	return out
}

func compactParties(parties ...info.Party) []info.Party {
	out := make([]info.Party, 0, len(parties))
	for _, party := range parties {
		if party.Name != "" || party.Document != "" {
			out = append(out, party)
		}
	}
	return out
}

func joinNonEmpty(sep string, values ...string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	return strings.Join(out, sep)
}
