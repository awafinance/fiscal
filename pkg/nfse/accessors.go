package nfse

import (
	"strconv"
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
	case d.infPedReg() != nil:
		return d.infPedReg().ChNFSe
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
	if d != nil && d.EventoNFSe != nil && d.EventoNFSe.InfEvento != nil && d.EventoNFSe.InfEvento.AmbGer != "" {
		return d.EventoNFSe.InfEvento.AmbGer
	}
	if reg := d.infPedReg(); reg != nil && reg.TpAmb != "" {
		return reg.TpAmb
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
	if reg := d.infPedReg(); reg != nil && reg.DhEvento != "" {
		return reg.DhEvento
	}
	if d != nil && d.EventoNFSe != nil && d.EventoNFSe.InfEvento != nil && d.EventoNFSe.InfEvento.DhProc != "" {
		return d.EventoNFSe.InfEvento.DhProc
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
	switch d.GetStatusCode() {
	case "100", "101", "102", "103", "107":
		return true
	default:
		return false
	}
}

func (d *Document) GetAmounts() []info.Amount {
	if d == nil {
		return nil
	}

	amounts := d.headlineAmounts()
	amounts = append(amounts, d.taxAmounts()...)
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
			{Type: "service", Value: nfseServiceAmount(d.infDPS())},
			{Type: "net", Value: v.VLiq},
			{Type: "tax_iss", Value: stringPtrValue(v.VISSQN)},
			{Type: "retained", Value: stringPtrValue(v.VTotalRet)},
		}
	}
	if d.DPS != nil && d.DPS.InfDPS != nil && d.DPS.InfDPS.Valores != nil && d.DPS.InfDPS.Valores.VServPrest != nil {
		return []info.Amount{{Type: "service", Value: d.DPS.InfDPS.Valores.VServPrest.VServ}}
	}
	return nil
}

func nfseServiceAmount(inf *TCInfDPS) string {
	if inf == nil || inf.Valores == nil || inf.Valores.VServPrest == nil {
		return ""
	}
	return inf.Valores.VServPrest.VServ
}

func (d *Document) taxAmounts() []info.Amount {
	inf := d.infDPS()
	if inf == nil || inf.Valores == nil || inf.Valores.Trib == nil {
		return nil
	}
	trib := inf.Valores.Trib
	var amounts []info.Amount
	if fed := trib.TribFed; fed != nil && fed.Piscofins != nil {
		amounts = append(amounts, nonZeroAmounts(
			info.Amount{Type: "tax_pis", Value: stringPtrValue(fed.Piscofins.VPis)},
			info.Amount{Type: "tax_cofins", Value: stringPtrValue(fed.Piscofins.VCofins)},
		)...)
	}
	if tot := trib.TotTrib; tot != nil && tot.VTotTrib != nil {
		amounts = append(amounts, nonZeroAmounts(
			info.Amount{Type: "taxes_fed", Value: tot.VTotTrib.VTotTribFed},
			info.Amount{Type: "taxes_est", Value: tot.VTotTrib.VTotTribEst},
			info.Amount{Type: "taxes_mun", Value: tot.VTotTrib.VTotTribMun},
		)...)
	}
	return amounts
}

func (d *Document) retentionAmounts() []info.Amount {
	inf := d.infDPS()
	if inf == nil || inf.Valores == nil || inf.Valores.Trib == nil {
		return nil
	}
	trib := inf.Valores.Trib

	var amounts []info.Amount

	// Municipal ISS retention: tpRetISSQN 2=retained by taker, 3=retained by intermediary.
	// The retained amount is vISSQN on the authorized NFSe's valores (computed by Sefin Nacional).
	if mun := trib.TribMun; mun != nil && (mun.TpRetISSQN == "2" || mun.TpRetISSQN == "3") {
		if d.NFSe != nil && d.NFSe.InfNFSe != nil && d.NFSe.InfNFSe.Valores != nil {
			amounts = append(amounts, info.Amount{
				Type:  "retained_iss",
				Value: stringPtrValue(d.NFSe.InfNFSe.Valores.VISSQN),
			})
		}
	}

	if fed := trib.TribFed; fed != nil {
		amounts = append(amounts,
			info.Amount{Type: "retained_irrf", Value: stringPtrValue(fed.VRetIRRF)},
			info.Amount{Type: "retained_csll", Value: stringPtrValue(fed.VRetCSLL)},
			info.Amount{Type: "retained_inss", Value: stringPtrValue(fed.VRetCP)},
		)
		if pc := fed.Piscofins; pc != nil && stringPtrValue(pc.TpRetPisCofins) == "1" {
			amounts = append(amounts,
				info.Amount{Type: "retained_pis", Value: stringPtrValue(pc.VPis)},
				info.Amount{Type: "retained_cofins", Value: stringPtrValue(pc.VCofins)},
			)
		}
	}

	return amounts
}

func (d *Document) GetCompetenceDate() string {
	if inf := d.infDPS(); inf != nil {
		return inf.DCompet
	}
	return ""
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
		d.providerParty(),
		d.takerParty(),
		d.intermediaryParty(),
	)
}

func (d *Document) providerParty() info.Party {
	party := info.Party{Role: "provider", Name: d.GetIssuer(), Document: d.GetIssuerDocument()}

	if emit := d.nfseEmit(); emit != nil {
		party.Name = firstNonEmpty(party.Name, emit.XNome)
		party.Document = firstNonEmpty(party.Document, firstStringPtr(emit.CNPJ, emit.CPF))
		party.MunicipalRegistration = firstNonEmpty(party.MunicipalRegistration, stringPtrValue(emit.IM))
		party.Phone = firstNonEmpty(party.Phone, stringPtrValue(emit.Fone))
		party.Email = firstNonEmpty(party.Email, stringPtrValue(emit.Email))
		party.Address = mergeAddress(party.Address, nfseIssuerAddress(emit.EnderNac))
	}

	if inf := d.infDPS(); inf != nil && inf.Prest != nil {
		prest := inf.Prest
		party.Name = firstNonEmpty(party.Name, stringPtrValue(prest.XNome))
		party.Document = firstNonEmpty(party.Document, firstStringPtr(prest.CNPJ, prest.CPF, prest.NIF))
		party.MunicipalRegistration = firstNonEmpty(party.MunicipalRegistration, stringPtrValue(prest.IM))
		party.Phone = firstNonEmpty(party.Phone, stringPtrValue(prest.Fone))
		party.Email = firstNonEmpty(party.Email, stringPtrValue(prest.Email))
		party.Address = mergeAddress(party.Address, nfseAddress(prest.End))
		if prest.RegTrib != nil {
			party.SimpleNationalOption = firstNonEmpty(party.SimpleNationalOption, prest.RegTrib.OpSimpNac)
			party.SimpleNationalRegime = firstNonEmpty(party.SimpleNationalRegime, stringPtrValue(prest.RegTrib.RegApTribSN))
			party.SpecialTaxRegime = firstNonEmpty(party.SpecialTaxRegime, prest.RegTrib.RegEspTrib)
		}
	}

	return party
}

func (d *Document) takerParty() info.Party {
	party := info.Party{Role: "taker", Name: d.GetRecipient(), Document: d.GetRecipientDocument()}
	if inf := d.infDPS(); inf != nil && inf.Toma != nil {
		fillPersonParty(&party, inf.Toma)
	}
	return party
}

func (d *Document) intermediaryParty() info.Party {
	party := info.Party{Role: "intermediary"}
	if inf := d.infDPS(); inf != nil {
		person := inf.Interm //nolint:misspell // Generated schema field name for <interm>.
		if person != nil {
			fillPersonParty(&party, person)
		}
	}
	return party
}

func fillPersonParty(party *info.Party, person *TCInfoPessoa) {
	if party == nil || person == nil {
		return
	}
	party.Name = firstNonEmpty(party.Name, person.XNome)
	party.Document = firstNonEmpty(party.Document, firstStringPtr(person.CNPJ, person.CPF, person.NIF))
	party.MunicipalRegistration = firstNonEmpty(party.MunicipalRegistration, stringPtrValue(person.IM))
	party.Phone = firstNonEmpty(party.Phone, stringPtrValue(person.Fone))
	party.Email = firstNonEmpty(party.Email, stringPtrValue(person.Email))
	party.Address = mergeAddress(party.Address, nfseAddress(person.End))
}

func (d *Document) nfseEmit() *TCEmitente {
	if d == nil || d.NFSe == nil || d.NFSe.InfNFSe == nil {
		return nil
	}
	return d.NFSe.InfNFSe.Emit
}

func nfseAddress(end *TCEndereco) *info.Address {
	if end == nil {
		return nil
	}
	address := &info.Address{
		Street:       end.XLgr,
		Number:       end.Nro,
		Complement:   stringPtrValue(end.XCpl),
		Neighborhood: end.XBairro,
	}
	if end.EndNac != nil {
		address.CityCode = end.EndNac.CMun
		address.PostalCode = end.EndNac.CEP
	}
	if end.EndExt != nil {
		address.CountryCode = end.EndExt.CPais
		address.PostalCode = end.EndExt.CEndPost
		address.CityName = end.EndExt.XCidade
		address.State = end.EndExt.XEstProvReg
	}
	return compactAddress(address)
}

func nfseIssuerAddress(end *TCEnderecoEmitente) *info.Address {
	if end == nil {
		return nil
	}
	return compactAddress(&info.Address{
		Street:       end.XLgr,
		Number:       end.Nro,
		Complement:   stringPtrValue(end.XCpl),
		Neighborhood: end.XBairro,
		PostalCode:   end.CEP,
		CityCode:     end.CMun,
		State:        end.UF,
	})
}

func mergeAddress(current, candidate *info.Address) *info.Address {
	if current == nil {
		return candidate
	}
	if candidate == nil {
		return current
	}
	current.Street = firstNonEmpty(current.Street, candidate.Street)
	current.Number = firstNonEmpty(current.Number, candidate.Number)
	current.Complement = firstNonEmpty(current.Complement, candidate.Complement)
	current.Neighborhood = firstNonEmpty(current.Neighborhood, candidate.Neighborhood)
	current.PostalCode = firstNonEmpty(current.PostalCode, candidate.PostalCode)
	current.CityCode = firstNonEmpty(current.CityCode, candidate.CityCode)
	current.CityName = firstNonEmpty(current.CityName, candidate.CityName)
	current.State = firstNonEmpty(current.State, candidate.State)
	current.CountryCode = firstNonEmpty(current.CountryCode, candidate.CountryCode)
	return compactAddress(current)
}

func compactAddress(address *info.Address) *info.Address {
	if address == nil || *address == (info.Address{}) {
		return nil
	}
	return address
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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

func nonZeroAmounts(amounts ...info.Amount) []info.Amount {
	out := make([]info.Amount, 0, len(amounts))
	for _, amount := range amounts {
		if isZeroAmount(amount.Value) {
			continue
		}
		out = append(out, amount)
	}
	return out
}

func isZeroAmount(value string) bool {
	if value == "" {
		return true
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	return f == 0
}

func compactParties(parties ...info.Party) []info.Party {
	out := make([]info.Party, 0, len(parties))
	for _, party := range parties {
		if partyHasData(party) {
			out = append(out, party)
		}
	}
	return out
}

func partyHasData(party info.Party) bool {
	return party.Name != "" ||
		party.Document != "" ||
		party.StateRegistration != "" ||
		party.MunicipalRegistration != "" ||
		party.Address != nil ||
		party.Phone != "" ||
		party.Email != "" ||
		party.SimpleNationalOption != "" ||
		party.SimpleNationalRegime != "" ||
		party.SpecialTaxRegime != ""
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
