package monetary

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

type node struct {
	name       xml.Name
	attrs      []xml.Attr
	text       strings.Builder
	children   []*node
	parent     *node
	occurrence int
}

//nolint:gocognit,gocyclo // XML token nesting is clearest as one small state machine.
func parseTree(data []byte) (*node, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true

	var root *node
	var current *node
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode xml tree: %w", err)
		}

		switch token := token.(type) {
		case xml.StartElement:
			child := &node{name: token.Name, attrs: token.Attr, parent: current, occurrence: 1}
			if current == nil {
				if root != nil {
					return nil, errors.New("decode xml tree: multiple root elements")
				}
				root = child
			} else {
				for _, sibling := range current.children {
					if sibling.name == child.name {
						child.occurrence++
					}
				}
				current.children = append(current.children, child)
			}
			current = child
		case xml.CharData:
			if current != nil {
				current.text.Write(token)
			}
		case xml.EndElement:
			if current == nil || current.name != token.Name {
				return nil, fmt.Errorf("decode xml tree: unexpected closing element %s", token.Name.Local)
			}
			current = current.parent
		}
	}

	if root == nil {
		return nil, errors.New("decode xml tree: missing root element")
	}
	if current != nil {
		return nil, fmt.Errorf("decode xml tree: unclosed element %s", current.name.Local)
	}
	return root, nil
}

func (n *node) value() string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.text.String())
}

func (n *node) attr(name string) string {
	if n == nil {
		return ""
	}
	for _, attr := range n.attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func (n *node) child(name string) *node {
	if n == nil {
		return nil
	}
	for _, child := range n.children {
		if child.name.Local == name {
			return child
		}
	}
	return nil
}

func (n *node) childValue(name string) string {
	return n.child(name).value()
}

func (n *node) descendants(name string) []*node {
	var found []*node
	var visit func(*node)
	visit = func(current *node) {
		if current.name.Local == name {
			found = append(found, current)
		}
		for _, child := range current.children {
			visit(child)
		}
	}
	if n != nil {
		visit(n)
	}
	return found
}

func (n *node) firstDescendant(names ...string) *node {
	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}
	var visit func(*node) *node
	visit = func(current *node) *node {
		if _, ok := wanted[current.name.Local]; ok {
			return current
		}
		for _, child := range current.children {
			if found := visit(child); found != nil {
				return found
			}
		}
		return nil
	}
	if n == nil {
		return nil
	}
	return visit(n)
}

func (n *node) nearestAncestor(names ...string) *node {
	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}
	for current := n; current != nil; current = current.parent {
		if _, ok := wanted[current.name.Local]; ok {
			return current
		}
	}
	return nil
}

func (n *node) path() []*node {
	var reversed []*node
	for current := n; current != nil; current = current.parent {
		reversed = append(reversed, current)
	}
	path := make([]*node, len(reversed))
	for i := range reversed {
		path[len(reversed)-1-i] = reversed[i]
	}
	return path
}
