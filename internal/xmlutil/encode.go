package xmlutil

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

const (
	xmlnsNamespace = "xmlns"
	xmlNamespace   = "http://www.w3.org/XML/1998/namespace"
	dsNamespace    = "http://www.w3.org/2000/09/xmldsig#"
)

// EncodeCanonical marshals a value and rewrites the resulting token stream to:
// - declare namespaces once on the root element
// - keep the document namespace as the default namespace
// - use a stable prefix for XMLDSig nodes
// - keep xmlns declarations ahead of regular attributes
func EncodeCanonical(enc *xml.Encoder, value any) error {
	data, err := xml.Marshal(value)
	if err != nil {
		return err
	}

	defaultNS, prefixByURI, err := collectNamespaces(data)
	if err != nil {
		return err
	}

	state := &canonicalWriter{
		enc:         enc,
		defaultNS:   defaultNS,
		prefixByURI: prefixByURI,
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := state.writeToken(tok); err != nil {
			return err
		}
	}
}

type canonicalWriter struct {
	enc          *xml.Encoder
	defaultNS    string
	prefixByURI  map[string]string
	elementStack []xml.Name
	rootWritten  bool
}

func (w *canonicalWriter) writeToken(tok xml.Token) error {
	switch tok := tok.(type) {
	case xml.StartElement:
		return w.writeStart(tok)
	case xml.EndElement:
		return w.writeEnd(tok)
	case xml.CharData:
		return w.enc.EncodeToken(tok.Copy())
	case xml.Comment:
		return w.enc.EncodeToken(tok.Copy())
	case xml.Directive:
		return w.enc.EncodeToken(tok.Copy())
	case xml.ProcInst:
		return w.enc.EncodeToken(tok)
	}
	return nil
}

func (w *canonicalWriter) writeStart(tok xml.StartElement) error {
	start := xml.StartElement{
		Name: canonicalName(tok.Name, w.defaultNS, w.prefixByURI),
		Attr: canonicalAttrs(tok.Attr, w.defaultNS, w.prefixByURI, !w.rootWritten),
	}
	if err := w.enc.EncodeToken(start); err != nil {
		return err
	}
	w.elementStack = append(w.elementStack, start.Name)
	w.rootWritten = true
	return nil
}

func (w *canonicalWriter) writeEnd(tok xml.EndElement) error {
	if len(w.elementStack) == 0 {
		return fmt.Errorf("xmlutil: unexpected end element %q", tok.Name.Local)
	}
	end := xml.EndElement{Name: w.elementStack[len(w.elementStack)-1]}
	w.elementStack = w.elementStack[:len(w.elementStack)-1]
	return w.enc.EncodeToken(end)
}

func ParseRootName(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var rootName string

	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if rootName == "" {
					return "", err
				}
				return rootName, nil
			}
			return rootName, err
		}

		if start, ok := tok.(xml.StartElement); ok && rootName == "" {
			rootName = start.Name.Local
		}
	}
}

func ParseRootElement(data []byte) (xml.Name, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		tok, err := decoder.Token()
		if err != nil {
			return xml.Name{}, err
		}

		if start, ok := tok.(xml.StartElement); ok {
			return start.Name, nil
		}
	}
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func collectNamespaces(data []byte) (string, map[string]string, error) {
	defaultNS, seen, err := scanNamespaces(data)
	if err != nil {
		return "", nil, err
	}
	return defaultNS, assignPrefixes(seen), nil
}

func scanNamespaces(data []byte) (string, map[string]struct{}, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	defaultNS := ""
	seen := map[string]struct{}{}
	rootSeen := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return defaultNS, seen, nil
			}
			return "", nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		if !rootSeen {
			rootSeen = true
			defaultNS = rootDefaultNS(start)
		}

		recordElementNamespaces(start, defaultNS, seen)
	}
}

func rootDefaultNS(start xml.StartElement) string {
	if start.Name.Space != "" {
		return start.Name.Space
	}
	for _, attr := range start.Attr {
		if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
			return attr.Value
		}
	}
	return ""
}

func recordElementNamespaces(start xml.StartElement, defaultNS string, seen map[string]struct{}) {
	if start.Name.Space != "" && start.Name.Space != defaultNS {
		seen[start.Name.Space] = struct{}{}
	}
	for _, attr := range start.Attr {
		if attr.Name.Space == "" || attr.Name.Space == xmlnsNamespace || attr.Name.Space == defaultNS {
			continue
		}
		seen[attr.Name.Space] = struct{}{}
	}
}

func assignPrefixes(seen map[string]struct{}) map[string]string {
	prefixByURI := map[string]string{}
	if _, ok := seen[xmlNamespace]; ok {
		prefixByURI[xmlNamespace] = "xml"
		delete(seen, xmlNamespace)
	}
	if _, ok := seen[dsNamespace]; ok {
		prefixByURI[dsNamespace] = "ds"
		delete(seen, dsNamespace)
	}

	uris := make([]string, 0, len(seen))
	for uri := range seen {
		uris = append(uris, uri)
	}
	slices.Sort(uris)
	for idx, uri := range uris {
		prefixByURI[uri] = fmt.Sprintf("ns%d", idx+1)
	}
	return prefixByURI
}

func canonicalName(name xml.Name, defaultNS string, prefixByURI map[string]string) xml.Name {
	local := qualify(name.Space, name.Local, defaultNS, prefixByURI, true)
	return xml.Name{Local: local}
}

func canonicalAttrs(attrs []xml.Attr, defaultNS string, prefixByURI map[string]string, root bool) []xml.Attr {
	out := make([]xml.Attr, 0, len(attrs)+len(prefixByURI)+1)
	if root {
		if defaultNS != "" {
			out = append(out, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: defaultNS})
		}

		pairs := make([]struct {
			uri    string
			prefix string
		}, 0, len(prefixByURI))
		for uri, prefix := range prefixByURI {
			if prefix == "" || prefix == "xml" {
				continue
			}
			pairs = append(pairs, struct {
				uri    string
				prefix string
			}{uri: uri, prefix: prefix})
		}
		slices.SortFunc(pairs, func(a, b struct {
			uri    string
			prefix string
		}) int {
			return strings.Compare(a.prefix, b.prefix)
		})
		for _, pair := range pairs {
			out = append(out, xml.Attr{
				Name:  xml.Name{Local: "xmlns:" + pair.prefix},
				Value: pair.uri,
			})
		}
	}

	for _, attr := range attrs {
		if isNamespaceDecl(attr) {
			continue
		}
		out = append(out, xml.Attr{
			Name:  xml.Name{Local: qualify(attr.Name.Space, attr.Name.Local, defaultNS, prefixByURI, false)},
			Value: attr.Value,
		})
	}

	return out
}

func qualify(space, local, defaultNS string, prefixByURI map[string]string, element bool) string {
	switch {
	case local == "":
		return ""
	case space == "", space == defaultNS:
		return local
	case !element && space == defaultNS:
		return local
	}

	if prefix, ok := prefixByURI[space]; ok && prefix != "" {
		return prefix + ":" + local
	}

	return local
}

func isNamespaceDecl(attr xml.Attr) bool {
	return attr.Name.Space == xmlnsNamespace || (attr.Name.Space == "" && strings.HasPrefix(attr.Name.Local, "xmlns"))
}
