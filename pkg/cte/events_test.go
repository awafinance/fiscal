package cte_test

import (
	"os"
	"testing"

	"github.com/awafinance/fiscal/pkg/cte"
	"github.com/awafinance/fiscal/pkg/info"
	"github.com/stretchr/testify/require"
)

func TestLifecycleEventAccessors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/cte/v4_0/cancel35150107565416000104570000000012301000012300-ped-eve.xml")
	require.NoError(t, err)

	doc, err := cte.Parse(data)
	require.NoError(t, err)

	require.Equal(t, "110111", doc.GetEventType())
	require.Equal(t, "1", doc.GetEventSequence())
	// The referenced note rides the base accessor for an event document.
	require.Equal(t, "35150107565416000104570000000012301000012300", doc.GetAccessKey())
}

func TestGetTypeAndSubstitutoBackRef(t *testing.T) {
	doc, err := cte.Parse([]byte(cteSubstitutoXML))
	require.NoError(t, err)

	require.Equal(t, "3", doc.GetType())
	require.Contains(t, doc.GetRelatedDocuments(), info.RelatedDocument{
		Type:      "cte",
		AccessKey: "43120178408960000182570010000000049000000048",
	})
}

func TestGetTypeNormal(t *testing.T) {
	doc, err := cte.Parse([]byte(cteWithCobrXML))
	require.NoError(t, err)
	require.Equal(t, "0", doc.GetType())
}

// Complemento (tpCTe=1) carries infCteComp and no infCTeNorm (they are an XSD
// choice); the back-ref must still surface.
func TestComplementoBackRef(t *testing.T) {
	doc, err := cte.Parse([]byte(cteComplementoXML))
	require.NoError(t, err)

	require.Equal(t, "1", doc.GetType())
	require.Contains(t, doc.GetRelatedDocuments(), info.RelatedDocument{
		Type:      "cte",
		AccessKey: "43120178408960000182570010000000049000000048",
	})
}

// CT-e OS (modelo 67) carries the same tpCTe back-references.
func TestCTeOSSubstitutoBackRef(t *testing.T) {
	doc, err := cte.Parse([]byte(cteOSSubstitutoXML))
	require.NoError(t, err)

	require.Equal(t, "3", doc.GetType())
	require.Contains(t, doc.GetRelatedDocuments(), info.RelatedDocument{
		Type:      "cte",
		AccessKey: "35170799999999999999670000000000269000000049",
	})
}

const cteComplementoXML = `<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe43120178408960000182570010000000041000000047" versao="4.00"><ide><cUF>43</cUF><cCT>00000004</cCT><CFOP>6353</CFOP><natOp>SERV</natOp><mod>57</mod><serie>1</serie><nCT>4</nCT><dhEmi>2012-01-06T17:25:56-02:00</dhEmi><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>7</cDV><tpAmb>2</tpAmb><tpCTe>1</tpCTe><procEmi>0</procEmi><verProc>104</verProc><cMunEnv>4213609</cMunEnv><xMunEnv>PORTO UNIAO</xMunEnv><UFEnv>SC</UFEnv><modal>01</modal><tpServ>0</tpServ><cMunIni>4213609</cMunIni><xMunIni>PORTO UNIAO</xMunIni><UFIni>SC</UFIni><cMunFim>4213609</cMunFim><xMunFim>PORTO UNIAO</xMunFim><UFFim>SC</UFFim><retira>0</retira><indIEToma>9</indIEToma></ide><emit><CNPJ>78408960000182</CNPJ><IE>ISENTO</IE><xNome>KERBER</xNome><enderEmit><xLgr>R</xLgr><nro>1</nro><xBairro>B</xBairro><cMun>4213609</cMun><xMun>PORTO UNIAO</xMun><CEP>89400000</CEP><UF>SC</UF></enderEmit></emit><infCteComp><chCTe>43120178408960000182570010000000049000000048</chCTe></infCteComp></infCte></CTe>`

const cteOSSubstitutoXML = `<CTeOS xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe35170799999999999999670000000000261309301440" versao="4.00"><ide><cUF>35</cUF><mod>67</mod><serie>4</serie><nCT>26</nCT><tpCTe>3</tpCTe><modal>01</modal></ide><emit><CNPJ>99999999999999</CNPJ><IE>ISENTO</IE><xNome>OS CARRIER</xNome></emit><infCTeNorm><infCteSub><chCte>35170799999999999999670000000000269000000049</chCte></infCteSub></infCTeNorm></infCte></CTeOS>`

const cteSubstitutoXML = `<CTe xmlns="http://www.portalfiscal.inf.br/cte"><infCte Id="CTe43120178408960000182570010000000041000000047" versao="4.00"><ide><cUF>43</cUF><cCT>00000004</cCT><CFOP>6353</CFOP><natOp>SERV</natOp><mod>57</mod><serie>1</serie><nCT>4</nCT><dhEmi>2012-01-06T17:25:56-02:00</dhEmi><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>7</cDV><tpAmb>2</tpAmb><tpCTe>3</tpCTe><procEmi>0</procEmi><verProc>104</verProc><cMunEnv>4213609</cMunEnv><xMunEnv>PORTO UNIAO</xMunEnv><UFEnv>SC</UFEnv><modal>01</modal><tpServ>0</tpServ><cMunIni>4213609</cMunIni><xMunIni>PORTO UNIAO</xMunIni><UFIni>SC</UFIni><cMunFim>4213609</cMunFim><xMunFim>PORTO UNIAO</xMunFim><UFFim>SC</UFFim><retira>0</retira><indIEToma>9</indIEToma></ide><emit><CNPJ>78408960000182</CNPJ><IE>ISENTO</IE><xNome>KERBER</xNome><enderEmit><xLgr>R</xLgr><nro>1</nro><xBairro>B</xBairro><cMun>4213609</cMun><xMun>PORTO UNIAO</xMun><CEP>89400000</CEP><UF>SC</UF></enderEmit></emit><infCTeNorm><infCteSub><chCte>43120178408960000182570010000000049000000048</chCte></infCteSub></infCTeNorm></infCte></CTe>`
