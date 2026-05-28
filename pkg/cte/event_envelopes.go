package cte

type cteSentEventEnvelope struct {
	RootPresent bool
	InfPresent  bool
	AccessKey   string
	Environment string
	IssueDate   string
	EventType   string
	HasDetail   bool
	DetailXML   string
}

type cteRetEventEnvelope struct {
	RootPresent    bool
	InfPresent     bool
	AccessKey      string
	Environment    string
	ProtocolNumber string
	StatusCode     string
	StatusReason   string
}

type cteSentEventReader func(any) (cteSentEventEnvelope, bool)
type cteRetEventReader func(any) (cteRetEventEnvelope, bool)

var cteSentEventReaders = []cteSentEventReader{
	readEventoCTeSentEnvelope,
	readEventoCancCTeSentEnvelope,
	readEventoCECTeSentEnvelope,
	readEventoCancCECTeSentEnvelope,
	readEventoEPECCTeSentEnvelope,
	readEventoRegMultimodalSentEnvelope,
	readEventoGTVSentEnvelope,
	readEventoIECTeSentEnvelope,
	readEventoCancIECTeSentEnvelope,
	readEventoPrestDesacordoSentEnvelope,
	readEventoCancPrestDesacordoSentEnvelope,
	readEventoGenericoSentEnvelope,
}

var cteRetEventReaders = []cteRetEventReader{
	readEventoCTeRetEnvelope,
	readEventoCancCTeRetEnvelope,
	readEventoCECTeRetEnvelope,
	readEventoCancCECTeRetEnvelope,
	readEventoEPECCTeRetEnvelope,
	readEventoRegMultimodalRetEnvelope,
	readEventoGTVRetEnvelope,
	readEventoIECTeRetEnvelope,
	readEventoCancIECTeRetEnvelope,
	readEventoPrestDesacordoRetEnvelope,
	readEventoCancPrestDesacordoRetEnvelope,
	readEventoGenericoRetEnvelope,
}

func readCTeSentEventEnvelope(root any) (cteSentEventEnvelope, bool) {
	for _, read := range cteSentEventReaders {
		if env, ok := read(root); ok {
			return env, true
		}
	}
	return cteSentEventEnvelope{}, false
}

func readCTeRetEventEnvelope(root any) (cteRetEventEnvelope, bool) {
	for _, read := range cteRetEventReaders {
		if env, ok := read(root); ok {
			return env, true
		}
	}
	return cteRetEventEnvelope{}, false
}

func readEventoCTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoCTeInf(v), true
}

func readEventoCancCTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoCancCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoCancCTeInf(v), true
}

func readEventoCECTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoCECTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoCECTeInf(v), true
}

func readEventoCancCECTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoCancCECTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoCancCECTeInf(v), true
}

func readEventoEPECCTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoEPECCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoEPECCTeInf(v), true
}

func readEventoRegMultimodalSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoRegMultimodalCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoRegMultimodalInf(v), true
}

func readEventoGTVSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoGTVCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoGTVInf(v), true
}

func readEventoIECTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoIECTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoIECTeInf(v), true
}

func readEventoCancIECTeSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoCancIECTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoCancIECTeInf(v), true
}

func readEventoPrestDesacordoSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoPrestDesacordoCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoPrestDesacordoInf(v), true
}

func readEventoCancPrestDesacordoSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoCancPrestDesacordoCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoCancPrestDesacordoInf(v), true
}

func readEventoGenericoSentEnvelope(root any) (cteSentEventEnvelope, bool) {
	v, ok := root.(*EventoGenericoCTeEvento)
	if !ok {
		return cteSentEventEnvelope{}, false
	}
	return sentEnvelopeFromEventoGenericoInf(v), true
}

func readEventoCTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoCTeInf(v), true
}

func readEventoCancCTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoCancCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoCancCTeInf(v), true
}

func readEventoCECTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoCECTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoCECTeInf(v), true
}

func readEventoCancCECTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoCancCECTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoCancCECTeInf(v), true
}

func readEventoEPECCTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoEPECCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoEPECCTeInf(v), true
}

func readEventoRegMultimodalRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoRegMultimodalCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoRegMultimodalInf(v), true
}

func readEventoGTVRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoGTVCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoGTVInf(v), true
}

func readEventoIECTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoIECTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoIECTeInf(v), true
}

func readEventoCancIECTeRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoCancIECTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoCancIECTeInf(v), true
}

func readEventoPrestDesacordoRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoPrestDesacordoCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoPrestDesacordoInf(v), true
}

func readEventoCancPrestDesacordoRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoCancPrestDesacordoCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoCancPrestDesacordoInf(v), true
}

func readEventoGenericoRetEnvelope(root any) (cteRetEventEnvelope, bool) {
	v, ok := root.(*EventoGenericoCTeRetEvento)
	if !ok {
		return cteRetEventEnvelope{}, false
	}
	return retEnvelopeFromEventoGenericoInf(v), true
}

func sentEnvelopeFromEventoCTeInf(root *EventoCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoCTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoCancCTeInf(root *EventoCancCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoCancCTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoCECTeInf(root *EventoCECTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoCECTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoCancCECTeInf(root *EventoCancCECTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoCancCECTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoEPECCTeInf(root *EventoEPECCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoEPECCTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoRegMultimodalInf(root *EventoRegMultimodalCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoRegMultimodalInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoGTVInf(root *EventoGTVCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoGTVInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoIECTeInf(root *EventoIECTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoIECTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoCancIECTeInf(root *EventoCancIECTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoCancIECTeInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoPrestDesacordoInf(root *EventoPrestDesacordoCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoPrestDesacordoInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoCancPrestDesacordoInf(root *EventoCancPrestDesacordoCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoCancPrestDesacordoInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoGenericoInf(root *EventoGenericoCTeEvento) cteSentEventEnvelope {
	if root == nil {
		return cteSentEventEnvelope{}
	}
	return sentEnvelopeFromEventoGenericoInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoCTeInf(root *EventoCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoCTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoCancCTeInf(root *EventoCancCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoCancCTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoCECTeInf(root *EventoCECTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoCECTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoCancCECTeInf(root *EventoCancCECTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoCancCECTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoEPECCTeInf(root *EventoEPECCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoEPECCTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoRegMultimodalInf(root *EventoRegMultimodalCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoRegMultimodalInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoGTVInf(root *EventoGTVCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoGTVInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoIECTeInf(root *EventoIECTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoIECTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoCancIECTeInf(root *EventoCancIECTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoCancIECTeInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoPrestDesacordoInf(root *EventoPrestDesacordoCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoPrestDesacordoInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoCancPrestDesacordoInf(root *EventoCancPrestDesacordoCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoCancPrestDesacordoInfEvento(root.InfEvento)
}

func retEnvelopeFromEventoGenericoInf(root *EventoGenericoCTeRetEvento) cteRetEventEnvelope {
	if root == nil {
		return cteRetEventEnvelope{}
	}
	return retEnvelopeFromEventoGenericoInfEvento(root.InfEvento)
}

func sentEnvelopeFromEventoCTeInfEvento(inf *EventoCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoCancCTeInfEvento(inf *EventoCancCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoCECTeInfEvento(inf *EventoCECTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoCancCECTeInfEvento(inf *EventoCancCECTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoEPECCTeInfEvento(inf *EventoEPECCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoRegMultimodalInfEvento(inf *EventoRegMultimodalCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoGTVInfEvento(inf *EventoGTVCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoIECTeInfEvento(inf *EventoIECTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoCancIECTeInfEvento(inf *EventoCancIECTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoPrestDesacordoInfEvento(inf *EventoPrestDesacordoCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoCancPrestDesacordoInfEvento(inf *EventoCancPrestDesacordoCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	return cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
}

func sentEnvelopeFromEventoGenericoInfEvento(inf *EventoGenericoCTeAnonComplexInfEvento1) cteSentEventEnvelope {
	if inf == nil {
		return cteSentEventEnvelope{RootPresent: true}
	}
	env := cteSentEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: inf.ChCTe, Environment: inf.TpAmb, IssueDate: inf.DhEvento, EventType: inf.TpEvento, HasDetail: inf.DetEvento != nil}
	if inf.DetEvento != nil {
		env.DetailXML = inf.DetEvento.InnerXML
	}
	return env
}

func retEnvelopeFromEventoCTeInfEvento(inf *EventoCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoCancCTeInfEvento(inf *EventoCancCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoCECTeInfEvento(inf *EventoCECTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoCancCECTeInfEvento(inf *EventoCancCECTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoEPECCTeInfEvento(inf *EventoEPECCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoRegMultimodalInfEvento(inf *EventoRegMultimodalCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoGTVInfEvento(inf *EventoGTVCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoIECTeInfEvento(inf *EventoIECTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoCancIECTeInfEvento(inf *EventoCancIECTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoPrestDesacordoInfEvento(inf *EventoPrestDesacordoCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoCancPrestDesacordoInfEvento(inf *EventoCancPrestDesacordoCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}

func retEnvelopeFromEventoGenericoInfEvento(inf *EventoGenericoCTeAnonComplexInfEvento2) cteRetEventEnvelope {
	if inf == nil {
		return cteRetEventEnvelope{RootPresent: true}
	}
	return cteRetEventEnvelope{RootPresent: true, InfPresent: true, AccessKey: stringPtrValue(inf.ChCTe), Environment: inf.TpAmb, ProtocolNumber: stringPtrValue(inf.NProt), StatusCode: inf.CStat, StatusReason: typedStringPtrValue(inf.XMotivo)}
}
