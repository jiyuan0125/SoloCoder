package ldapfilter

import (
	"encoding/json"
	"strings"
)

type NodeType int

const (
	NodeTypeSimple NodeType = iota
	NodeTypePresent
	NodeTypeSubstring
	NodeTypeAnd
	NodeTypeOr
	NodeTypeNot
)

type Operator int

const (
	OpEqual Operator = iota
	OpApprox
	OpGreaterEqual
	OpLessEqual
)

func (o Operator) String() string {
	switch o {
	case OpEqual:
		return "="
	case OpApprox:
		return "~="
	case OpGreaterEqual:
		return ">="
	case OpLessEqual:
		return "<="
	default:
		return "?"
	}
}

type Node interface {
	Type() NodeType
	String() string
}

type SimpleNode struct {
	Attribute string
	Operator  Operator
	Value     string
}

func (n *SimpleNode) Type() NodeType { return NodeTypeSimple }
func (n *SimpleNode) String() string {
	return n.Attribute + n.Operator.String() + n.Value
}

type PresentNode struct {
	Attribute string
}

func (n *PresentNode) Type() NodeType { return NodeTypePresent }
func (n *PresentNode) String() string {
	return n.Attribute + "=*"
}

type SubstringNode struct {
	Attribute string
	Initial   string
	Any       []string
	Final     string
}

func (n *SubstringNode) Type() NodeType { return NodeTypeSubstring }
func (n *SubstringNode) String() string {
	var sb strings.Builder
	sb.WriteString(n.Attribute)
	sb.WriteString("=")
	sb.WriteString(n.Initial)
	sb.WriteString("*")
	for _, a := range n.Any {
		sb.WriteString(a)
		sb.WriteString("*")
	}
	sb.WriteString(n.Final)
	return sb.String()
}

type AndNode struct {
	Children []Node
}

func (n *AndNode) Type() NodeType { return NodeTypeAnd }
func (n *AndNode) String() string {
	var sb strings.Builder
	sb.WriteString("&")
	for _, c := range n.Children {
		sb.WriteString("(")
		sb.WriteString(c.String())
		sb.WriteString(")")
	}
	return sb.String()
}

type OrNode struct {
	Children []Node
}

func (n *OrNode) Type() NodeType { return NodeTypeOr }
func (n *OrNode) String() string {
	var sb strings.Builder
	sb.WriteString("|")
	for _, c := range n.Children {
		sb.WriteString("(")
		sb.WriteString(c.String())
		sb.WriteString(")")
	}
	return sb.String()
}

type NotNode struct {
	Child Node
}

func (n *NotNode) Type() NodeType { return NodeTypeNot }
func (n *NotNode) String() string {
	return "!(" + n.Child.String() + ")"
}

type astNodeJSON struct {
	Type     string      `json:"type"`
	Children []astNodeJSON `json:"children,omitempty"`
	Attr     string      `json:"attribute,omitempty"`
	Op       string      `json:"operator,omitempty"`
	Value    string      `json:"value,omitempty"`
	Initial  string      `json:"initial,omitempty"`
	Any      []string    `json:"any,omitempty"`
	Final    string      `json:"final,omitempty"`
}

func ToASTJSON(n Node) interface{} {
	return toASTJSON(n)
}

func toASTJSON(n Node) astNodeJSON {
	switch node := n.(type) {
	case *SimpleNode:
		return astNodeJSON{
			Type:  "simple",
			Attr:  node.Attribute,
			Op:    node.Operator.String(),
			Value: node.Value,
		}
	case *PresentNode:
		return astNodeJSON{
			Type:  "present",
			Attr:  node.Attribute,
		}
	case *SubstringNode:
		return astNodeJSON{
			Type:    "substring",
			Attr:    node.Attribute,
			Initial: node.Initial,
			Any:     node.Any,
			Final:   node.Final,
		}
	case *AndNode:
		children := make([]astNodeJSON, len(node.Children))
		for i, c := range node.Children {
			children[i] = toASTJSON(c)
		}
		return astNodeJSON{
			Type:     "and",
			Children: children,
		}
	case *OrNode:
		children := make([]astNodeJSON, len(node.Children))
		for i, c := range node.Children {
			children[i] = toASTJSON(c)
		}
		return astNodeJSON{
			Type:     "or",
			Children: children,
		}
	case *NotNode:
		return astNodeJSON{
			Type:     "not",
			Children: []astNodeJSON{toASTJSON(node.Child)},
		}
	default:
		return astNodeJSON{Type: "unknown"}
	}
}

func MarshalAST(n Node) (string, error) {
	aj := toASTJSON(n)
	b, err := json.MarshalIndent(aj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func PrintTree(n Node) string {
	return printTree(n, 0)
}

func printTree(n Node, depth int) string {
	indent := strings.Repeat("  ", depth)
	var sb strings.Builder
	switch node := n.(type) {
	case *SimpleNode:
		sb.WriteString(indent)
		sb.WriteString("SIMPLE ")
		sb.WriteString(node.Attribute)
		sb.WriteString(node.Operator.String())
		sb.WriteString(node.Value)
	case *PresentNode:
		sb.WriteString(indent)
		sb.WriteString("PRESENT ")
		sb.WriteString(node.Attribute)
		sb.WriteString("=*")
	case *SubstringNode:
		sb.WriteString(indent)
		sb.WriteString("SUBSTRING ")
		sb.WriteString(node.Attribute)
		sb.WriteString("=")
		if node.Initial != "" {
			sb.WriteString(node.Initial)
		}
		sb.WriteString("*")
		for _, a := range node.Any {
			sb.WriteString(a)
			sb.WriteString("*")
		}
		if node.Final != "" {
			sb.WriteString(node.Final)
		}
	case *AndNode:
		sb.WriteString(indent)
		sb.WriteString("AND")
		for _, c := range node.Children {
			sb.WriteString("\n")
			sb.WriteString(printTree(c, depth+1))
		}
	case *OrNode:
		sb.WriteString(indent)
		sb.WriteString("OR")
		for _, c := range node.Children {
			sb.WriteString("\n")
			sb.WriteString(printTree(c, depth+1))
		}
	case *NotNode:
		sb.WriteString(indent)
		sb.WriteString("NOT")
		sb.WriteString("\n")
		sb.WriteString(printTree(node.Child, depth+1))
	}
	return sb.String()
}
