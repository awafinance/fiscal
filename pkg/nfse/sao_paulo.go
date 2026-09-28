package nfse

import (
	"encoding/xml"
	"errors"
	"fmt"
)

// SaoPauloNFSe holds facts from a São Paulo municipal NFe export.
// ChaveNotaNacional is source data, not a validated national access key.
type SaoPauloNFSe struct {
	XMLName  xml.Name `xml:"NFe" json:"-"`
	ChaveNFe struct {
		InscricaoPrestador string `xml:"InscricaoPrestador"`
		NumeroNFe          string `xml:"NumeroNFe"`
		CodigoVerificacao  string `xml:"CodigoVerificacao"`
		ChaveNotaNacional  string `xml:"ChaveNotaNacional"`
	} `xml:"ChaveNFe"`
	DataEmissaoNFe     string `xml:"DataEmissaoNFe"`
	DataFatoGeradorNFe string `xml:"DataFatoGeradorNFe"`
	CPFCNPJPrestador   struct {
		CNPJ string `xml:"CNPJ"`
		CPF  string `xml:"CPF"`
	} `xml:"CPFCNPJPrestador"`
	RazaoSocialPrestador string `xml:"RazaoSocialPrestador"`
	CPFCNPJTomador       struct {
		CNPJ string `xml:"CNPJ"`
		CPF  string `xml:"CPF"`
	} `xml:"CPFCNPJTomador"`
	RazaoSocialTomador string `xml:"RazaoSocialTomador"`
	StatusNFe          string `xml:"StatusNFe"`
	ValorServicos      string `xml:"ValorServicos"`
	ValorDeducoes      string `xml:"ValorDeducoes"`
	ValorFinalCobrado  string `xml:"ValorFinalCobrado"`
	ValorISS           string `xml:"ValorISS"`
	ValorPIS           string `xml:"ValorPIS"`
	ValorCOFINS        string `xml:"ValorCOFINS"`
	ValorINSS          string `xml:"ValorINSS"`
	ValorIR            string `xml:"ValorIR"`
	ValorCSLL          string `xml:"ValorCSLL"`
	RetencaoPisCofins  string `xml:"RetencaoPisCofins"`
	ISSRetido          bool   `xml:"ISSRetido"`
	Discriminacao      string `xml:"Discriminacao"`
}

func parseSaoPaulo(data []byte) (*Document, error) {
	var parsed SaoPauloNFSe
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse nfse: decode Sao Paulo NFe: %w", err)
	}
	if parsed.XMLName.Space != "" || parsed.ChaveNFe.InscricaoPrestador == "" ||
		parsed.ChaveNFe.NumeroNFe == "" || parsed.ChaveNFe.CodigoVerificacao == "" ||
		firstNonEmpty(parsed.CPFCNPJPrestador.CNPJ, parsed.CPFCNPJPrestador.CPF) == "" {
		return nil, errors.New("parse nfse: missing Sao Paulo invoice identity")
	}
	return &Document{RootName: "NFe", SaoPaulo: &parsed}, nil
}
