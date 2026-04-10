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

	decoder := xml.NewDecoder(bytes.NewReader(data))
	var elementStack []xml.Name
	rootWritten := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			start := xml.StartElement{
				Name: canonicalName(tok.Name, defaultNS, prefixByURI),
				Attr: canonicalAttrs(tok.Attr, defaultNS, prefixByURI, !rootWritten),
			}
			if err := enc.EncodeToken(start); err != nil {
				return err
			}
			elementStack = append(elementStack, start.Name)
			rootWritten = true
		case xml.EndElement:
			if len(elementStack) == 0 {
				return fmt.Errorf("xmlutil: unexpected end element %q", tok.Name.Local)
			}
			end := xml.EndElement{Name: elementStack[len(elementStack)-1]}
			elementStack = elementStack[:len(elementStack)-1]
			if err := enc.EncodeToken(end); err != nil {
				return err
			}
		case xml.CharData:
			if err := enc.EncodeToken(tok.Copy()); err != nil {
				return err
			}
		case xml.Comment:
			if err := enc.EncodeToken(tok.Copy()); err != nil {
				return err
			}
		case xml.Directive:
			if err := enc.EncodeToken(tok.Copy()); err != nil {
				return err
			}
		case xml.ProcInst:
			if err := enc.EncodeToken(tok); err != nil {
				return err
			}
		}
	}
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

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func collectNamespaces(data []byte) (string, map[string]string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	defaultNS := ""
	seen := map[string]struct{}{}
	rootSeen := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		if !rootSeen {
			rootSeen = true
			defaultNS = start.Name.Space
			if defaultNS == "" {
				for _, attr := range start.Attr {
					if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
						defaultNS = attr.Value
						break
					}
				}
			}
		}

		if start.Name.Space != "" && start.Name.Space != defaultNS {
			seen[start.Name.Space] = struct{}{}
		}
		for _, attr := range start.Attr {
			if attr.Name.Space == "" || attr.Name.Space == xmlnsNamespace {
				continue
			}
			if attr.Name.Space == defaultNS {
				continue
			}
			seen[attr.Name.Space] = struct{}{}
		}
	}

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

	return defaultNS, prefixByURI, nil
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
