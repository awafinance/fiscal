package cte

type cteEventEnvelope struct {
	RootPresent    bool
	InfPresent     bool
	AccessKey      string
	Environment    string
	IssueDate      string
	EventType      string
	ProtocolNumber string
	StatusCode     string
	StatusReason   string
	HasDetail      bool
	DetailXML      string
}

func readCTeSentEventEnvelope(root any) (cteEventEnvelope, bool) {
	spec, value, ok := cteEventSpecForRoot(root, cteSentEventRoot)
	if !ok {
		return cteEventEnvelope{}, false
	}
	if value.IsNil() {
		return cteEventEnvelope{}, true
	}

	inf := cteField(value, "InfEvento")
	if !cteHasValue(inf) {
		return cteEventEnvelope{RootPresent: true}, true
	}

	detail := cteField(inf, "DetEvento")
	env := cteEventEnvelope{
		RootPresent: true,
		InfPresent:  true,
		AccessKey:   cteStringField(inf, "ChCTe"),
		Environment: cteStringField(inf, "TpAmb"),
		IssueDate:   cteStringField(inf, "DhEvento"),
		EventType:   cteStringField(inf, "TpEvento"),
		HasDetail:   cteHasValue(detail),
	}
	if spec.captureDetailXML {
		env.DetailXML = cteStringField(detail, "InnerXML")
	}
	return env, true
}

func readCTeRetEventEnvelope(root any) (cteEventEnvelope, bool) {
	_, value, ok := cteEventSpecForRoot(root, cteRetEventRoot)
	if !ok {
		return cteEventEnvelope{}, false
	}
	if value.IsNil() {
		return cteEventEnvelope{}, true
	}

	inf := cteField(value, "InfEvento")
	if !cteHasValue(inf) {
		return cteEventEnvelope{RootPresent: true}, true
	}

	return cteEventEnvelope{
		RootPresent:    true,
		InfPresent:     true,
		AccessKey:      cteStringField(inf, "ChCTe"),
		Environment:    cteStringField(inf, "TpAmb"),
		ProtocolNumber: cteStringField(inf, "NProt"),
		StatusCode:     cteStringField(inf, "CStat"),
		StatusReason:   cteStringField(inf, "XMotivo"),
	}, true
}
