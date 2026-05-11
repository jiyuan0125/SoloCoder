package hclparser

import (
	"encoding/json"
	"fmt"
)

type Node interface {
	node()
	String() string
}

type Body struct {
	Statements []Statement `json:"statements"`
}

func (b *Body) node()        {}
func (b *Body) String() string { return "Body" }

type Statement interface {
	Node
	statement()
}

type Block struct {
	Type       string        `json:"type"`
	Labels     []string      `json:"labels"`
	Body       *Body         `json:"body"`
}

func (b *Block) node()      {}
func (b *Block) statement() {}
func (b *Block) String() string {
	return fmt.Sprintf("Block{Type: %s, Labels: %v}", b.Type, b.Labels)
}

type Attribute struct {
	Key   string    `json:"key"`
	Value Expression `json:"value"`
}

func (a *Attribute) node()      {}
func (a *Attribute) statement() {}
func (a *Attribute) String() string {
	return fmt.Sprintf("Attribute{Key: %s}", a.Key)
}

type Expression interface {
	Node
	expression()
}

type LiteralValue struct {
	Value interface{} `json:"value"`
}

func (l *LiteralValue) node()       {}
func (l *LiteralValue) expression() {}
func (l *LiteralValue) String() string {
	return fmt.Sprintf("LiteralValue{%v}", l.Value)
}

type VariableReference struct {
	Path []string      `json:"path"`
	Indices []Expression `json:"indices,omitempty"`
}

func (v *VariableReference) node()       {}
func (v *VariableReference) expression() {}
func (v *VariableReference) String() string {
	return fmt.Sprintf("VariableReference{Path: %v}", v.Path)
}

type FunctionCall struct {
	Name          string       `json:"name"`
	PositionalArgs []Expression `json:"positional_args,omitempty"`
	NamedArgs     map[string]Expression `json:"named_args,omitempty"`
}

func (f *FunctionCall) node()       {}
func (f *FunctionCall) expression() {}
func (f *FunctionCall) String() string {
	return fmt.Sprintf("FunctionCall{Name: %s}", f.Name)
}

type ConditionalExpr struct {
	Condition  Expression `json:"condition"`
	TrueValue  Expression `json:"true_value"`
	FalseValue Expression `json:"false_value"`
}

func (c *ConditionalExpr) node()       {}
func (c *ConditionalExpr) expression() {}
func (c *ConditionalExpr) String() string {
	return "ConditionalExpr"
}

type BinaryOp struct {
	Left  Expression `json:"left"`
	Op    string     `json:"op"`
	Right Expression `json:"right"`
}

func (b *BinaryOp) node()       {}
func (b *BinaryOp) expression() {}
func (b *BinaryOp) String() string {
	return fmt.Sprintf("BinaryOp{Op: %s}", b.Op)
}

type UnaryOp struct {
	Op     string     `json:"op"`
	Operand Expression `json:"operand"`
}

func (u *UnaryOp) node()       {}
func (u *UnaryOp) expression() {}
func (u *UnaryOp) String() string {
	return fmt.Sprintf("UnaryOp{Op: %s}", u.Op)
}

type ListExpr struct {
	Elements []Expression `json:"elements"`
}

func (l *ListExpr) node()       {}
func (l *ListExpr) expression() {}
func (l *ListExpr) String() string {
	return fmt.Sprintf("ListExpr{%d elements}", len(l.Elements))
}

type MapExpr struct {
	Elements map[string]Expression `json:"elements"`
}

func (m *MapExpr) node()       {}
func (m *MapExpr) expression() {}
func (m *MapExpr) String() string {
	return fmt.Sprintf("MapExpr{%d elements}", len(m.Elements))
}

type IndexExpr struct {
	Collection Expression `json:"collection"`
	Index      Expression `json:"index"`
}

func (i *IndexExpr) node()       {}
func (i *IndexExpr) expression() {}
func (i *IndexExpr) String() string {
	return "IndexExpr"
}

type HeredocValue struct {
	Delimiter   string `json:"delimiter"`
	Content     string `json:"content"`
	StripIndent bool   `json:"strip_indent"`
	HasInterp   bool   `json:"has_interp"`
}

func (h *HeredocValue) node()       {}
func (h *HeredocValue) expression() {}
func (h *HeredocValue) String() string {
	return fmt.Sprintf("HeredocValue{Delimiter: %s}", h.Delimiter)
}

func (b *Body) MarshalJSON() ([]byte, error) {
	type Alias Body
	blocks := make([]*Block, 0)
	attrs := make([]*Attribute, 0)
	for _, stmt := range b.Statements {
		switch s := stmt.(type) {
		case *Block:
			blocks = append(blocks, s)
		case *Attribute:
			attrs = append(attrs, s)
		}
	}
	return json.Marshal(&struct {
		*Alias
		Blocks     []*Block     `json:"blocks,omitempty"`
		Attributes []*Attribute `json:"attributes,omitempty"`
	}{
		Alias:      (*Alias)(b),
		Blocks:     blocks,
		Attributes: attrs,
	})
}

func (f *FunctionCall) MarshalJSON() ([]byte, error) {
	type Alias FunctionCall
	named := make([]struct {
		Key   string     `json:"key"`
		Value Expression `json:"value"`
	}, 0, len(f.NamedArgs))
	for k, v := range f.NamedArgs {
		named = append(named, struct {
			Key   string     `json:"key"`
			Value Expression `json:"value"`
		}{Key: k, Value: v})
	}
	return json.Marshal(&struct {
		*Alias
		NamedArgsList []struct {
			Key   string     `json:"key"`
			Value Expression `json:"value"`
		} `json:"named_args_list,omitempty"`
	}{
		Alias:         (*Alias)(f),
		NamedArgsList: named,
	})
}

func (m *MapExpr) MarshalJSON() ([]byte, error) {
	type Alias MapExpr
	elements := make([]struct {
		Key   string     `json:"key"`
		Value Expression `json:"value"`
	}, 0, len(m.Elements))
	for k, v := range m.Elements {
		elements = append(elements, struct {
			Key   string     `json:"key"`
			Value Expression `json:"value"`
		}{Key: k, Value: v})
	}
	return json.Marshal(&struct {
		*Alias
		ElementsList []struct {
			Key   string     `json:"key"`
			Value Expression `json:"value"`
		} `json:"elements_list,omitempty"`
	}{
		Alias:        (*Alias)(m),
		ElementsList: elements,
	})
}

type ParseResult struct {
	Body *Body `json:"body"`
}
