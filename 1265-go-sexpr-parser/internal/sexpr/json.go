package sexpr

import (
	"encoding/json"
)

type JSONExpr struct {
	Type     string        `json:"type"`
	Value    interface{}   `json:"value,omitempty"`
	Elements []*JSONExpr   `json:"elements,omitempty"`
}

func ToJSON(expr Expr) *JSONExpr {
	switch v := expr.(type) {
	case *Atom:
		switch v.Type {
		case AtomInt:
			return &JSONExpr{Type: "integer", Value: v.IntValue}
		case AtomFloat:
			return &JSONExpr{Type: "float", Value: v.FloatValue}
		case AtomString:
			return &JSONExpr{Type: "string", Value: v.StringValue}
		case AtomSymbol:
			return &JSONExpr{Type: "symbol", Value: v.SymbolValue}
		}
	case *List:
		if v.IsNil() {
			return &JSONExpr{Type: "nil", Elements: []*JSONExpr{}}
		}
		elements := make([]*JSONExpr, len(v.Elements))
		for i, e := range v.Elements {
			elements[i] = ToJSON(e)
		}
		return &JSONExpr{Type: "list", Elements: elements}
	}
	return nil
}

func ToJSONString(expr Expr) (string, error) {
	j := ToJSON(expr)
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ToJSONAll(exprs []Expr) (string, error) {
	jexprs := make([]*JSONExpr, len(exprs))
	for i, e := range exprs {
		jexprs[i] = ToJSON(e)
	}
	data, err := json.MarshalIndent(jexprs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
