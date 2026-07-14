package fiscal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNFeMonetaryInterpretationPreservesW16FactsAndChecks(t *testing.T) {
	t.Parallel()

	doc, err := Parse([]byte(nfeMonetaryXML))
	require.NoError(t, err)

	interpretation := doc.MonetaryInterpretation()
	require.NotNil(t, interpretation)
	require.Equal(t, MonetaryContractVersion, interpretation.ContractVersion)
	require.Equal(t, "NF-e PL 010e v1.02", interpretation.Provenance.Schema)
	require.Equal(t, "55", interpretation.Provenance.Model)
	require.Equal(t, "4.00", interpretation.Provenance.Layout)

	require.Equal(t, "145.00", requireMonetaryFact(t, interpretation, "nfe.document.total.ICMSTot.vNF", 1).Value)
	require.Equal(t, "2.00", requireMonetaryFact(t, interpretation, "nfe.document.total.ICMSTot.vSeg", 1).Value)
	require.Equal(t, "1.00", requireMonetaryFact(t, interpretation, "nfe.document.total.ICMSTot.vOutro", 1).Value)
	require.Equal(t, "8.00", requireMonetaryFact(t, interpretation, "nfe.document.total.ISSQNtot.vServ", 1).Value)
	require.Equal(t, "0.00", requireMonetaryFact(t, interpretation, "nfe.document.total.ICMSTot.vICMS", 1).Value)

	w16 := requireMonetaryCheck(t, interpretation, "W16-10")
	require.Equal(t, MonetaryCheckPassed, w16.Status)
	require.Len(t, w16.Calculations, 2)
	require.Equal(t, "145.00", w16.Calculations[0].Calculated)

	w60 := requireMonetaryCheck(t, interpretation, "W60-10")
	require.Equal(t, MonetaryCheckPassed, w60.Status)
	require.Equal(t, "150.00", w60.DeclaredValue)
}

func TestCTeMonetaryInterpretationPreservesReceivableAndComponents(t *testing.T) {
	t.Parallel()

	doc, err := Parse([]byte(cteMonetaryXML))
	require.NoError(t, err)

	interpretation := doc.MonetaryInterpretation()
	require.NotNil(t, interpretation)
	require.Equal(t, "CT-e 4.00 through NT 2026.002 v1.00", interpretation.Provenance.Schema)
	require.Equal(t, "12480.50", requireMonetaryFact(t, interpretation, "cte.document.vPrest.vTPrest", 1).Value)
	require.Equal(t, "12000.00", requireMonetaryFact(t, interpretation, "cte.document.vPrest.vRec", 1).Value)

	freight := requireMonetaryFact(t, interpretation, "cte.document.vPrest.Comp.vComp", 1)
	require.Equal(t, "11820.00", freight.Value)
	require.Equal(t, "FRETE PESO", freight.Label)
	require.Equal(t, MonetaryScopeComponent, freight.Scope)
	toll := requireMonetaryFact(t, interpretation, "cte.document.vPrest.Comp.vComp", 2)
	require.Equal(t, "660.50", toll.Value)
	require.Equal(t, "PEDAGIO", toll.Label)

	require.Equal(t, MonetaryCheckPassed, requireMonetaryCheck(t, interpretation, "G048/H056").Status)
}

func TestNFSeMonetaryInterpretationSeparatesActualAndApproximateTaxes(t *testing.T) {
	t.Parallel()

	doc, err := Parse([]byte(nfseMonetaryXML))
	require.NoError(t, err)

	interpretation := doc.MonetaryInterpretation()
	require.NotNil(t, interpretation)
	require.Equal(t, "NFS-e Nacional v1.01 2026-02-09", interpretation.Provenance.Schema)

	pis := requireMonetaryFact(t, interpretation, "nfse.dps.valores.trib.tribFed.piscofins.vPis", 1)
	require.Equal(t, MonetaryFactActualTax, pis.Kind)
	require.Nil(t, pis.Retention)
	cofins := requireMonetaryFact(t, interpretation, "nfse.dps.valores.trib.tribFed.piscofins.vCofins", 1)
	require.Equal(t, MonetaryFactActualTax, cofins.Kind)
	require.Nil(t, cofins.Retention)

	aggregate := requireMonetaryFact(t, interpretation, "nfse.dps.valores.trib.tribFed.vRetCSLL", 1)
	require.Equal(t, MonetaryFactRetention, aggregate.Kind)
	require.Equal(t, []string{"PIS", "COFINS", "CSLL"}, aggregate.TaxTypes)
	require.NotNil(t, aggregate.Retention)
	require.Equal(t, "aggregate_unallocated", aggregate.Retention.Allocation)

	iss := requireMonetaryFact(t, interpretation, "nfse.document.valores.vISSQN", 1)
	require.NotNil(t, iss.Retention)
	require.Equal(t, "taker", iss.Retention.Actor)

	federalBurden := requireMonetaryFact(t, interpretation, "nfse.dps.valores.trib.totTrib.vTotTrib.vTotTribFed", 1)
	require.Equal(t, MonetaryFactApproximateTax, federalBurden.Kind)
	require.Equal(t, "36.50", federalBurden.Value)
	stateBurden := requireMonetaryFact(t, interpretation, "nfse.dps.valores.trib.totTrib.vTotTrib.vTotTribEst", 1)
	require.Equal(t, "0.00", stateBurden.Value, "an explicit zero must not be deleted")
	require.Equal(t, MonetaryFactApproximateTaxRate,
		requireMonetaryFact(t, interpretation, "nfse.dps.valores.trib.totTrib.pTotTrib.pTotTribFed", 1).Kind)

	require.Equal(t, "913.50", requireMonetaryFact(t, interpretation, "nfse.document.IBSCBS.totCIBS.vTotNF", 1).Value)
	require.Equal(t, MonetaryCheckPassed, requireMonetaryCheck(t, interpretation, "NFSE-vLiq").Status)
	require.Equal(t, MonetaryCheckPassed, requireMonetaryCheck(t, interpretation, "NFSE-vTotNF").Status)
}

func TestMonetaryInterpretationRejectsUnknownRelease(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte(`<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte versao="3.00"><ide><mod>57</mod></ide></infCte></CTe>`))
	require.ErrorIs(t, err, ErrUnsupportedRelease)
	var releaseErr *UnsupportedReleaseError
	require.ErrorAs(t, err, &releaseErr)
	require.Equal(t, CTe, releaseErr.Family)
	require.Equal(t, "3.00", releaseErr.Version)
}

func requireMonetaryFact(
	t *testing.T,
	interpretation *MonetaryInterpretation,
	code string,
	occurrence int,
) MonetaryFact {
	t.Helper()
	for _, fact := range interpretation.Facts {
		if fact.Code == code && fact.Occurrence == occurrence {
			return fact
		}
	}
	require.Failf(t, "monetary fact not found", "%s occurrence %d in %#v", code, occurrence, interpretation.Facts)
	return MonetaryFact{}
}

func requireMonetaryCheck(t *testing.T, interpretation *MonetaryInterpretation, rule string) MonetaryCheck {
	t.Helper()
	for _, check := range interpretation.Checks {
		if check.Rule == rule {
			return check
		}
	}
	require.Failf(t, "monetary check not found", "%s in %#v", rule, interpretation.Checks)
	return MonetaryCheck{}
}

const nfeMonetaryXML = `<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe35260712345678000195550010000000011000000010" versao="4.00">
    <ide><mod>55</mod></ide>
    <emit><CNPJ>12345678000195</CNPJ></emit>
    <det nItem="1">
      <prod><CFOP>5102</CFOP><vProd>100.00</vProd></prod>
      <imposto>
        <PISST><vPIS>2.00</vPIS><indSomaPISST>1</indSomaPISST></PISST>
        <COFINSST><vCOFINS>3.00</vCOFINS><indSomaCOFINSST>1</indSomaCOFINSST></COFINSST>
      </imposto>
      <vItem>150.00</vItem>
    </det>
    <total>
      <ICMSTot>
        <vBC>0.00</vBC><vICMS>0.00</vICMS><vICMSDeson>2.00</vICMSDeson>
        <vFCP>0.00</vFCP><vBCST>0.00</vBCST><vST>10.00</vST><vFCPST>1.00</vFCPST><vFCPSTRet>0.00</vFCPSTRet>
        <vICMSMonoReten>3.00</vICMSMonoReten><vProd>100.00</vProd><vFrete>4.00</vFrete><vSeg>2.00</vSeg>
        <vDesc>5.00</vDesc><vII>5.00</vII><vIPI>6.00</vIPI><vIPIDevol>7.00</vIPIDevol>
        <vPIS>0.00</vPIS><vCOFINS>0.00</vCOFINS><vOutro>1.00</vOutro><vNF>145.00</vNF><vTotTrib>22.00</vTotTrib>
      </ICMSTot>
      <ISSQNtot><vServ>8.00</vServ></ISSQNtot>
      <vNFTot>150.00</vNFTot>
    </total>
  </infNFe>
</NFe>`

const cteMonetaryXML = `<CTe xmlns="http://www.portalfiscal.inf.br/cte">
  <infCte Id="CTe35260712345678000195570010000000011000000010" versao="4.00">
    <ide><mod>57</mod><dhEmi>2026-07-01T10:00:00-03:00</dhEmi></ide>
    <emit><CNPJ>12345678000195</CNPJ></emit>
    <vPrest>
      <vTPrest>12480.50</vTPrest><vRec>12000.00</vRec>
      <Comp><xNome>FRETE PESO</xNome><vComp>11820.00</vComp></Comp>
      <Comp><xNome>PEDAGIO</xNome><vComp>660.50</vComp></Comp>
    </vPrest>
    <imp><vTotTrib>0.00</vTotTrib></imp>
  </infCte>
</CTe>`

const nfseMonetaryXML = `<NFSe xmlns="http://www.sped.fazenda.gov.br/nfse" versao="1.01">
  <infNFSe Id="NFS00000000000000000000000000000000000000000001">
    <dhProc>2026-07-01T10:00:00-03:00</dhProc>
    <emit><CNPJ>12345678000195</CNPJ></emit>
    <valores><vISSQN>50.00</vISSQN><vTotalRet>86.50</vTotalRet><vLiq>913.50</vLiq></valores>
    <IBSCBS><totCIBS><vTotNF>913.50</vTotNF></totCIBS></IBSCBS>
    <DPS versao="1.00"><infDPS Id="DPS00000000000000000000000000000000000000000001">
      <valores>
        <vServPrest><vServ>1000.00</vServ></vServPrest>
        <trib>
          <tribMun><tpRetISSQN>2</tpRetISSQN></tribMun>
          <tribFed>
            <piscofins><vPis>6.50</vPis><vCofins>30.00</vCofins><tpRetPisCofins>3</tpRetPisCofins></piscofins>
            <vRetCSLL>36.50</vRetCSLL>
          </tribFed>
          <totTrib>
            <vTotTrib><vTotTribFed>36.50</vTotTribFed><vTotTribEst>0.00</vTotTribEst><vTotTribMun>50.00</vTotTribMun></vTotTrib>
            <pTotTrib><pTotTribFed>3.65</pTotTribFed><pTotTribEst>0.00</pTotTribEst><pTotTribMun>5.00</pTotTribMun></pTotTrib>
          </totTrib>
        </trib>
      </valores>
    </infDPS></DPS>
  </infNFSe>
</NFSe>`
