package jsonpath

import (
	"fmt"
	"strconv"
)

func (l *Literal) Eval(current interface{}, root interface{}) (interface{}, error) {
	return l.Value, nil
}

func (p *PathExpression) Eval(current interface{}, root interface{}) (interface{}, error) {
	start := root
	if p.Rel {
		start = current
	}
	results := []interface{}{start}
	for _, step := range p.Steps {
		results = evalPathStepAll(results, step)
		if len(results) == 0 {
			return nil, nil
		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

func evalPathStepAll(nodes []interface{}, step PathStep) []interface{} {
	out := []interface{}{}
	for _, n := range nodes {
		out = append(out, evalPathStep(n, step)...)
	}
	return out
}

func evalPathStep(node interface{}, step PathStep) []interface{} {
	switch s := step.(type) {
	case *DotStep:
		if m, ok := node.(map[string]interface{}); ok {
			if v, ok := m[s.Name]; ok {
				return []interface{}{v}
			}
		}
	case *DotDotStep:
		return collectDescendants(node, s.Name)
	case *BracketStep:
		out := []interface{}{}
		for _, sel := range s.Selectors {
			out = append(out, applyBracketStepSelector(node, sel)...)
		}
		return out
	}
	return nil
}

func applyBracketStepSelector(node interface{}, sel BracketStepSelector) []interface{} {
	switch s := sel.(type) {
	case *NameStepSelector:
		if m, ok := node.(map[string]interface{}); ok {
			if v, ok := m[s.Name]; ok {
				return []interface{}{v}
			}
		}
	case *IndexStepSelector:
		if arr, ok := node.([]interface{}); ok {
			idx := s.Index
			if idx < 0 {
				idx = len(arr) + idx
			}
			if idx >= 0 && idx < len(arr) {
				return []interface{}{arr[idx]}
			}
		}
	case *WildcardStepSelector:
		if m, ok := node.(map[string]interface{}); ok {
			out := []interface{}{}
			for _, v := range m {
				out = append(out, v)
			}
			return out
		}
		if arr, ok := node.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

func (b *BinaryExpr) Eval(current interface{}, root interface{}) (interface{}, error) {
	left, err := b.Left.Eval(current, root)
	if err != nil {
		return nil, err
	}
	right, err := b.Right.Eval(current, root)
	if err != nil {
		return nil, err
	}
	return applyBinaryOp(b.Op, left, right), nil
}

func applyBinaryOp(op BinaryOp, left, right interface{}) interface{} {
	switch op {
	case BinaryAnd:
		if truthy(left) && truthy(right) {
			return true
		}
		return false
	case BinaryOr:
		if truthy(left) || truthy(right) {
			return true
		}
		return false
	case BinaryEqual:
		return valuesEqual(left, right)
	case BinaryNotEqual:
		return !valuesEqual(left, right)
	default:
		return applyCompareOp(op, left, right)
	}
}

func truthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0
	case []interface{}:
		return len(x) > 0
	case map[string]interface{}:
		return len(x) > 0
	default:
		return true
	}
}

func valuesEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func applyCompareOp(op BinaryOp, left, right interface{}) interface{} {
	ln, lOK := toNumber(left)
	rn, rOK := toNumber(right)
	if lOK && rOK {
		switch op {
		case BinaryLess:
			return ln < rn
		case BinaryLessEqual:
			return ln <= rn
		case BinaryGreater:
			return ln > rn
		case BinaryGreaterEqual:
			return ln >= rn
		}
	}
	ls, lOK := toString(left)
	rs, rOK := toString(right)
	if lOK && rOK {
		switch op {
		case BinaryLess:
			return ls < rs
		case BinaryLessEqual:
			return ls <= rs
		case BinaryGreater:
			return ls > rs
		case BinaryGreaterEqual:
			return ls >= rs
		}
	}
	return false
}

func toNumber(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case string:
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func toString(v interface{}) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case int:
		return strconv.Itoa(x), true
	case bool:
		return strconv.FormatBool(x), true
	}
	return "", false
}
