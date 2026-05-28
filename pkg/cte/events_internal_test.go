package cte

import (
	"reflect"
	"testing"
)

func TestCTeEventSpecsMatchDocumentFields(t *testing.T) {
	t.Parallel()

	docType := reflect.TypeFor[Document]()
	genericCount := 0
	for _, spec := range cteEventSpecs {
		if spec.generic {
			genericCount++
		}
		for _, kind := range []cteEventRootKind{cteSentEventRoot, cteRetEventRoot, cteProcEventRoot} {
			field, ok := docType.FieldByName(spec.docField(kind))
			if !ok {
				t.Fatalf("missing Document field %q for event %q", spec.docField(kind), spec.context)
			}
			if want := reflect.PointerTo(spec.rootType(kind)); field.Type != want {
				t.Fatalf("Document.%s type = %s, want %s", field.Name, field.Type, want)
			}
		}
	}
	if genericCount != 1 {
		t.Fatalf("generic CT-e event specs = %d, want 1", genericCount)
	}
}

func TestCTeEventSpecsMatchEnvelopeShape(t *testing.T) {
	t.Parallel()

	for _, spec := range cteEventSpecs {
		assertCTeStringField(t, spec.rootType(cteSentEventRoot), "VersaoAttr")
		assertCTeStringField(t, spec.rootType(cteRetEventRoot), "VersaoAttr")
		assertCTeStringField(t, spec.rootType(cteProcEventRoot), "VersaoAttr")
		assertCTePointerField(t, spec.rootType(cteSentEventRoot), "InfEvento")
		assertCTePointerField(t, spec.rootType(cteRetEventRoot), "InfEvento")
		assertCTePointerField(t, spec.rootType(cteSentEventRoot), "DsSignature")
		assertCTePointerField(t, spec.rootType(cteRetEventRoot), "DsSignature")
		assertCTeTypedPointerField(t, spec.rootType(cteProcEventRoot), "EventoCTe", spec.rootType(cteSentEventRoot))
		assertCTeTypedPointerField(t, spec.rootType(cteProcEventRoot), "RetEventoCTe", spec.rootType(cteRetEventRoot))

		sentInf := mustCTePointerField(t, spec.rootType(cteSentEventRoot), "InfEvento").Elem()
		assertCTeStringField(t, sentInf, "ChCTe")
		assertCTeStringField(t, sentInf, "TpAmb")
		assertCTeStringField(t, sentInf, "DhEvento")
		assertCTeStringField(t, sentInf, "TpEvento")
		assertCTePointerField(t, sentInf, "DetEvento")

		retInf := mustCTePointerField(t, spec.rootType(cteRetEventRoot), "InfEvento").Elem()
		assertCTeStringField(t, retInf, "ChCTe")
		assertCTeStringField(t, retInf, "TpAmb")
		assertCTeStringField(t, retInf, "NProt")
		assertCTeStringField(t, retInf, "CStat")
		assertCTeStringField(t, retInf, "XMotivo")
	}
}

func TestCTeStringPtrField(t *testing.T) {
	t.Parallel()

	type namedString string
	plain := "plain"
	named := namedString("named")
	value := reflect.ValueOf(struct {
		Plain   *string
		Named   *namedString
		Nil     *string
		NonPtr  string
		NonText *int
	}{
		Plain:   &plain,
		Named:   &named,
		NonPtr:  "value",
		NonText: new(int),
	})

	if got := cteStringPtrField(value, "Plain"); got == nil || *got != "plain" {
		t.Fatalf("Plain = %v, want plain", got)
	}
	if got := cteStringPtrField(value, "Named"); got == nil || *got != "named" {
		t.Fatalf("Named = %v, want named", got)
	}
	if got := cteStringPtrField(value, "Nil"); got != nil {
		t.Fatalf("Nil = %v, want nil", got)
	}
	if got := cteStringPtrField(value, "NonPtr"); got != nil {
		t.Fatalf("NonPtr = %v, want nil", got)
	}
	if got := cteStringPtrField(value, "NonText"); got != nil {
		t.Fatalf("NonText = %v, want nil", got)
	}
	if got := cteStringPtrField(value, "Missing"); got != nil {
		t.Fatalf("Missing = %v, want nil", got)
	}
}

func assertCTeStringField(t *testing.T, typ reflect.Type, name string) {
	t.Helper()

	fieldType := mustCTeField(t, typ, name)
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.String {
		t.Fatalf("%s.%s type = %s, want string-compatible", typ, name, fieldType)
	}
}

func assertCTePointerField(t *testing.T, typ reflect.Type, name string) {
	t.Helper()

	fieldType := mustCTeField(t, typ, name)
	if fieldType.Kind() != reflect.Pointer {
		t.Fatalf("%s.%s type = %s, want pointer", typ, name, fieldType)
	}
}

func assertCTeTypedPointerField(t *testing.T, typ reflect.Type, name string, elem reflect.Type) {
	t.Helper()

	fieldType := mustCTeField(t, typ, name)
	if fieldType != reflect.PointerTo(elem) {
		t.Fatalf("%s.%s type = %s, want %s", typ, name, fieldType, reflect.PointerTo(elem))
	}
}

func mustCTePointerField(t *testing.T, typ reflect.Type, name string) reflect.Type {
	t.Helper()

	fieldType := mustCTeField(t, typ, name)
	if fieldType.Kind() != reflect.Pointer {
		t.Fatalf("%s.%s type = %s, want pointer", typ, name, fieldType)
	}
	return fieldType
}

func mustCTeField(t *testing.T, typ reflect.Type, name string) reflect.Type {
	t.Helper()

	field, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("missing %s.%s", typ, name)
	}
	return field.Type
}
