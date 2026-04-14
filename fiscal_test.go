package fiscal

import (
	"os"
	"strings"
	"testing"
)

func TestParseRoutesByNamespace(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		family   Family
		rootName string
	}{
		{
			name:     "nfe",
			path:     "testdata/nfe/42220575277525000178550030000292481295366801-procNFe.xml",
			family:   NFe,
			rootName: "nfeProc",
		},
		{
			name:     "nfse",
			path:     "testdata/nfse/v1_0/dps-simples.xml",
			family:   NFSe,
			rootName: "DPS",
		},
		{
			name:     "cte",
			path:     "testdata/cte/v4_0/43120178408960000182570010000000041000000047-cte.xml",
			family:   CTe,
			rootName: "CTe",
		},
		{
			name:     "mdfe",
			path:     "testdata/mdfe/v3_0/41190876676436000167580010000500001000437558-mdfe.xml",
			family:   MDFe,
			rootName: "MDFe",
		},
		{
			name:     "bpe",
			path:     "testdata/bpe/v1_0/43190812345678000195630010000000011000000011-bpe.xml",
			family:   BPe,
			rootName: "BPe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			doc, err := Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			if doc.Family != tt.family {
				t.Fatalf("Family = %q, want %q", doc.Family, tt.family)
			}
			if doc.RootName != tt.rootName {
				t.Fatalf("RootName = %q, want %q", doc.RootName, tt.rootName)
			}
		})
	}
}

func TestParseDetectsPrefixedRootNamespace(t *testing.T) {
	data := []byte(`<x:NFe xmlns:x="http://www.portalfiscal.inf.br/nfe"></x:NFe>`)

	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse succeeded, want NFe validation error")
	}
	if !strings.Contains(err.Error(), "parse nfe:") {
		t.Fatalf("error = %q, want NFe parser error", err)
	}
	if strings.Contains(err.Error(), "unsupported namespace") {
		t.Fatalf("error = %q, want namespace to be supported", err)
	}
}

func TestParseRejectsUnsupportedNamespace(t *testing.T) {
	_, err := Parse([]byte(`<doc xmlns="urn:example"></doc>`))
	if err == nil {
		t.Fatal("Parse succeeded, want unsupported namespace error")
	}
	if !strings.Contains(err.Error(), `unsupported namespace "urn:example" root "doc"`) {
		t.Fatalf("error = %q, want unsupported namespace error", err)
	}
}
