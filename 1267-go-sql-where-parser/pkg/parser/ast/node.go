package ast

import (
	"encoding/json"
	"fmt"
)

type Node interface {
	NodeType() string
	Accept(v Visitor)
}

type Visitor interface {
	Visit(n Node)
}

type ComparisonOperator string

const (
	OpEq  ComparisonOperator = "="
	OpNeq ComparisonOperator = "!="
	OpGt  ComparisonOperator = ">"
	OpLt  ComparisonOperator = "<"
	OpGte ComparisonOperator = ">="
	OpLte ComparisonOperator = "<="
)

type LogicalOperator string

const (
	OpAnd LogicalOperator = "AND"
	OpOr  LogicalOperator = "OR"
)

type ComparisonNode struct {
	Column   string             `json:"column"`
	Operator ComparisonOperator `json:"operator"`
	Value    Value              `json:"value"`
}

func (n *ComparisonNode) NodeType() string { return "ComparisonNode" }

func (n *ComparisonNode) Accept(v Visitor) { v.Visit(n) }

func (n *ComparisonNode) MarshalJSON() ([]byte, error) {
	type Alias ComparisonNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type LogicalNode struct {
	Operator LogicalOperator `json:"operator"`
	Left     Node            `json:"left"`
	Right    Node            `json:"right"`
}

func (n *LogicalNode) NodeType() string { return "LogicalNode" }

func (n *LogicalNode) Accept(v Visitor) {
	if n.Left != nil {
		n.Left.Accept(v)
	}
	v.Visit(n)
	if n.Right != nil {
		n.Right.Accept(v)
	}
}

func (n *LogicalNode) MarshalJSON() ([]byte, error) {
	type Alias LogicalNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type NotNode struct {
	Operand Node `json:"operand"`
}

func (n *NotNode) NodeType() string { return "NotNode" }

func (n *NotNode) Accept(v Visitor) {
	if n.Operand != nil {
		n.Operand.Accept(v)
	}
	v.Visit(n)
}

func (n *NotNode) MarshalJSON() ([]byte, error) {
	type Alias NotNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type InNode struct {
	Column string  `json:"column"`
	Not    bool    `json:"not"`
	Values []Value `json:"values"`
}

func (n *InNode) NodeType() string { return "InNode" }

func (n *InNode) Accept(v Visitor) { v.Visit(n) }

func (n *InNode) MarshalJSON() ([]byte, error) {
	type Alias InNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type LikeNode struct {
	Column  string `json:"column"`
	Not     bool   `json:"not"`
	Pattern string `json:"pattern"`
	Escape  string `json:"escape,omitempty"`
}

func (n *LikeNode) NodeType() string { return "LikeNode" }

func (n *LikeNode) Accept(v Visitor) { v.Visit(n) }

func (n *LikeNode) MarshalJSON() ([]byte, error) {
	type Alias LikeNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type BetweenNode struct {
	Column string `json:"column"`
	Not    bool   `json:"not"`
	Lower  Value  `json:"lower"`
	Upper  Value  `json:"upper"`
}

func (n *BetweenNode) NodeType() string { return "BetweenNode" }

func (n *BetweenNode) Accept(v Visitor) { v.Visit(n) }

func (n *BetweenNode) MarshalJSON() ([]byte, error) {
	type Alias BetweenNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type NullCheckNode struct {
	Column string `json:"column"`
	Not    bool   `json:"not"`
}

func (n *NullCheckNode) NodeType() string { return "NullCheckNode" }

func (n *NullCheckNode) Accept(v Visitor) { v.Visit(n) }

func (n *NullCheckNode) MarshalJSON() ([]byte, error) {
	type Alias NullCheckNode
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  n.NodeType(),
		Alias: (*Alias)(n),
	})
}

type ValueType string

const (
	ValueTypeNumber  ValueType = "number"
	ValueTypeString  ValueType = "string"
	ValueTypeNull    ValueType = "null"
)

type Value struct {
	Type  ValueType   `json:"type"`
	Value interface{} `json:"value"`
}

func NewNumberValue(num float64) Value {
	return Value{Type: ValueTypeNumber, Value: num}
}

func NewStringValue(str string) Value {
	return Value{Type: ValueTypeString, Value: str}
}

func NewNullValue() Value {
	return Value{Type: ValueTypeNull, Value: nil}
}

func (v Value) String() string {
	switch v.Type {
	case ValueTypeNumber:
		return fmt.Sprintf("%v", v.Value)
	case ValueTypeString:
		return v.Value.(string)
	case ValueTypeNull:
		return "NULL"
	default:
		return "unknown"
	}
}
