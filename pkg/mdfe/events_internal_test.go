package mdfe

import (
	"reflect"
	"testing"
)

func TestMDFeEventSpecsMatchDocumentFields(t *testing.T) {
	t.Parallel()

	docType := reflect.TypeFor[Document]()
	for _, spec := range mdfeEventSpecs {
		for _, kind := range []mdfeEventRootKind{mdfeSentEventRoot, mdfeRetEventRoot, mdfeProcEventRoot} {
			field, ok := docType.FieldByName(spec.docField(kind))
			if !ok {
				t.Fatalf("missing Document field %q for event %q", spec.docField(kind), spec.context)
			}
			if want := reflect.PointerTo(spec.rootType(kind)); field.Type != want {
				t.Fatalf("Document.%s type = %s, want %s", field.Name, field.Type, want)
			}
		}
	}
}

func TestMDFeEventSpecsMatchEnvelopeShape(t *testing.T) {
	t.Parallel()

	for _, spec := range mdfeEventSpecs {
		assertMDFeStringField(t, spec.rootType(mdfeSentEventRoot), "VersaoAttr")
		assertMDFeStringField(t, spec.rootType(mdfeRetEventRoot), "VersaoAttr")
		assertMDFeStringField(t, spec.rootType(mdfeProcEventRoot), "VersaoAttr")
		assertMDFePointerField(t, spec.rootType(mdfeSentEventRoot), "InfEvento")
		assertMDFePointerField(t, spec.rootType(mdfeRetEventRoot), "InfEvento")
		assertMDFePointerField(t, spec.rootType(mdfeSentEventRoot), "DsSignature")
		assertMDFeTypedPointerField(t, spec.rootType(mdfeProcEventRoot), "EventoMDFe", spec.rootType(mdfeSentEventRoot))
		assertMDFeTypedPointerField(t, spec.rootType(mdfeProcEventRoot), "RetEventoMDFe", spec.rootType(mdfeRetEventRoot))

		sentInf := mustMDFePointerField(t, spec.rootType(mdfeSentEventRoot), "InfEvento").Elem()
		assertMDFeStringField(t, sentInf, "ChMDFe")
		assertMDFePointerField(t, sentInf, "DetEvento")

		retInf := mustMDFePointerField(t, spec.rootType(mdfeRetEventRoot), "InfEvento").Elem()
		assertMDFeStringField(t, retInf, "TpAmb")
		assertMDFeStringField(t, retInf, "CStat")
	}
}

func assertMDFeStringField(t *testing.T, typ reflect.Type, name string) {
	t.Helper()

	fieldType := mustMDFeField(t, typ, name)
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.String {
		t.Fatalf("%s.%s type = %s, want string-compatible", typ, name, fieldType)
	}
}

func assertMDFePointerField(t *testing.T, typ reflect.Type, name string) {
	t.Helper()

	fieldType := mustMDFeField(t, typ, name)
	if fieldType.Kind() != reflect.Pointer {
		t.Fatalf("%s.%s type = %s, want pointer", typ, name, fieldType)
	}
}

func assertMDFeTypedPointerField(t *testing.T, typ reflect.Type, name string, elem reflect.Type) {
	t.Helper()

	fieldType := mustMDFeField(t, typ, name)
	if fieldType != reflect.PointerTo(elem) {
		t.Fatalf("%s.%s type = %s, want %s", typ, name, fieldType, reflect.PointerTo(elem))
	}
}

func mustMDFePointerField(t *testing.T, typ reflect.Type, name string) reflect.Type {
	t.Helper()

	fieldType := mustMDFeField(t, typ, name)
	if fieldType.Kind() != reflect.Pointer {
		t.Fatalf("%s.%s type = %s, want pointer", typ, name, fieldType)
	}
	return fieldType
}

func mustMDFeField(t *testing.T, typ reflect.Type, name string) reflect.Type {
	t.Helper()

	field, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("missing %s.%s", typ, name)
	}
	return field.Type
}
