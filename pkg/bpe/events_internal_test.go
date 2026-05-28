package bpe

import (
	"reflect"
	"testing"
)

func TestBPeEventSpecsMatchDocumentFields(t *testing.T) {
	t.Parallel()

	docType := reflect.TypeFor[Document]()
	for _, spec := range bpeEventSpecs {
		for _, kind := range []bpeEventRootKind{bpeSentEventRoot, bpeRetEventRoot, bpeProcEventRoot} {
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

func TestBPeEventSpecsMatchEnvelopeShape(t *testing.T) {
	t.Parallel()

	for _, spec := range bpeEventSpecs {
		assertBPeStringField(t, spec.rootType(bpeSentEventRoot), "VersaoAttr")
		assertBPeStringField(t, spec.rootType(bpeRetEventRoot), "VersaoAttr")
		assertBPeStringField(t, spec.rootType(bpeProcEventRoot), "VersaoAttr")
		assertBPePointerField(t, spec.rootType(bpeSentEventRoot), "InfEvento")
		assertBPePointerField(t, spec.rootType(bpeRetEventRoot), "InfEvento")
		assertBPePointerField(t, spec.rootType(bpeSentEventRoot), "DsSignature")
		assertBPePointerField(t, spec.rootType(bpeRetEventRoot), "DsSignature")
		assertBPeTypedPointerField(t, spec.rootType(bpeProcEventRoot), "EventoBPe", spec.rootType(bpeSentEventRoot))
		assertBPeTypedPointerField(t, spec.rootType(bpeProcEventRoot), "RetEventoBPe", spec.rootType(bpeRetEventRoot))

		sentInf := mustBPePointerField(t, spec.rootType(bpeSentEventRoot), "InfEvento").Elem()
		assertBPeStringField(t, sentInf, "ChBPe")
		assertBPePointerField(t, sentInf, "DetEvento")

		retInf := mustBPePointerField(t, spec.rootType(bpeRetEventRoot), "InfEvento").Elem()
		assertBPeStringField(t, retInf, "TpAmb")
		assertBPeStringField(t, retInf, "CStat")
	}
}

func assertBPeStringField(t *testing.T, typ reflect.Type, name string) {
	t.Helper()

	fieldType := mustBPeField(t, typ, name)
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.String {
		t.Fatalf("%s.%s type = %s, want string-compatible", typ, name, fieldType)
	}
}

func assertBPePointerField(t *testing.T, typ reflect.Type, name string) {
	t.Helper()

	fieldType := mustBPeField(t, typ, name)
	if fieldType.Kind() != reflect.Pointer {
		t.Fatalf("%s.%s type = %s, want pointer", typ, name, fieldType)
	}
}

func assertBPeTypedPointerField(t *testing.T, typ reflect.Type, name string, elem reflect.Type) {
	t.Helper()

	fieldType := mustBPeField(t, typ, name)
	if fieldType != reflect.PointerTo(elem) {
		t.Fatalf("%s.%s type = %s, want %s", typ, name, fieldType, reflect.PointerTo(elem))
	}
}

func mustBPePointerField(t *testing.T, typ reflect.Type, name string) reflect.Type {
	t.Helper()

	fieldType := mustBPeField(t, typ, name)
	if fieldType.Kind() != reflect.Pointer {
		t.Fatalf("%s.%s type = %s, want pointer", typ, name, fieldType)
	}
	return fieldType
}

func mustBPeField(t *testing.T, typ reflect.Type, name string) reflect.Type {
	t.Helper()

	field, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("missing %s.%s", typ, name)
	}
	return field.Type
}
