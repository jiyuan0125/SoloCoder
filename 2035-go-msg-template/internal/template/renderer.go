package template

import (
	"fmt"
	"io"
	"strings"
)

type Renderer struct {
	partials    map[string][]Node
	partialLine map[string]int
	partialCol  map[string]int
}

func NewRenderer() *Renderer {
	return &Renderer{
		partials:    make(map[string][]Node),
		partialLine: make(map[string]int),
		partialCol:  make(map[string]int),
	}
}

func (r *Renderer) AddPartial(name string, content string, filename string) error {
	parser := NewParser(content, filename)
	nodes, err := parser.Parse()
	if err != nil {
		return err
	}
	r.partials[name] = nodes
	return nil
}

func (r *Renderer) Render(nodes []Node, data map[string]interface{}, w io.Writer) error {
	return r.renderNodes(nodes, data, w)
}

func (r *Renderer) renderNodes(nodes []Node, data map[string]interface{}, w io.Writer) error {
	for _, node := range nodes {
		switch n := node.(type) {
		case *TextNode:
			if _, err := io.WriteString(w, n.Content); err != nil {
				return err
			}
		case *VarNode:
			val := r.resolveValue(n.Name, data)
			if val == nil {
				if n.DefaultValue != "" {
					if _, err := io.WriteString(w, n.DefaultValue); err != nil {
						return err
					}
				}
			} else {
				if _, err := io.WriteString(w, fmt.Sprintf("%v", val)); err != nil {
					return err
				}
			}
		case *IfNode:
			if r.evalCondition(n.Condition, data) {
				if err := r.renderNodes(n.Children, data, w); err != nil {
					return err
				}
			}
		case *PartialNode:
			partialNodes, ok := r.partials[n.Name]
			if !ok {
				return &SyntaxError{
					File:   "(partial)",
					Line:   n.Line,
					Column: n.Column,
					Msg:    fmt.Sprintf("partial template not found: %s", n.Name),
				}
			}
			if err := r.renderNodes(partialNodes, data, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) resolveValue(name string, data map[string]interface{}) interface{} {
	parts := strings.Split(name, ".")
	var current interface{} = data

	for i, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil
			}
		} else {
			if i == 0 {
				if val, exists := data[part]; exists {
					current = val
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}

	return current
}

func (r *Renderer) evalCondition(cond string, data map[string]interface{}) bool {
	cond = strings.TrimSpace(cond)
	if cond == "" {
		return false
	}

	val := r.resolveValue(cond, data)
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int64, float64:
		return v != 0
	default:
		return true
	}
}
