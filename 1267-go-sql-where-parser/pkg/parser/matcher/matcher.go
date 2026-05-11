package matcher

import (
	"fmt"

	"sqlparser/pkg/parser/ast"
)

func Match(node ast.Node, data map[string]interface{}) (bool, error) {
	return matchNode(node, data)
}

func matchNode(node ast.Node, data map[string]interface{}) (bool, error) {
	switch n := node.(type) {
	case *ast.ComparisonNode:
		return matchComparison(n, data)
	case *ast.LogicalNode:
		return matchLogical(n, data)
	case *ast.NotNode:
		return matchNot(n, data)
	case *ast.InNode:
		return matchIn(n, data)
	case *ast.LikeNode:
		return matchLike(n, data)
	case *ast.BetweenNode:
		return matchBetween(n, data)
	case *ast.NullCheckNode:
		return matchNullCheck(n, data)
	default:
		return false, fmt.Errorf("unknown node type: %s", node.NodeType())
	}
}

func matchComparison(n *ast.ComparisonNode, data map[string]interface{}) (bool, error) {
	colVal, exists := data[n.Column]
	if !exists {
		return false, fmt.Errorf("column not found: %s", n.Column)
	}

	if colVal == nil {
		return false, nil
	}

	return compareValues(colVal, n.Operator, n.Value)
}

func compareValues(a interface{}, op ast.ComparisonOperator, b ast.Value) (bool, error) {
	switch b.Type {
	case ast.ValueTypeNumber:
		aNum, err := toFloat64(a)
		if err != nil {
			return false, err
		}
		bNum := b.Value.(float64)
		return compareNumbers(aNum, op, bNum), nil
	case ast.ValueTypeString:
		aStr := fmt.Sprintf("%v", a)
		bStr := b.Value.(string)
		return compareStrings(aStr, op, bStr), nil
	case ast.ValueTypeNull:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported value type: %s", b.Type)
	}
}

func compareNumbers(a float64, op ast.ComparisonOperator, b float64) bool {
	switch op {
	case ast.OpEq:
		return a == b
	case ast.OpNeq:
		return a != b
	case ast.OpGt:
		return a > b
	case ast.OpLt:
		return a < b
	case ast.OpGte:
		return a >= b
	case ast.OpLte:
		return a <= b
	default:
		return false
	}
}

func compareStrings(a string, op ast.ComparisonOperator, b string) bool {
	switch op {
	case ast.OpEq:
		return a == b
	case ast.OpNeq:
		return a != b
	case ast.OpGt:
		return a > b
	case ast.OpLt:
		return a < b
	case ast.OpGte:
		return a >= b
	case ast.OpLte:
		return a <= b
	default:
		return false
	}
}

func matchLogical(n *ast.LogicalNode, data map[string]interface{}) (bool, error) {
	left, err := matchNode(n.Left, data)
	if err != nil {
		return false, err
	}

	if n.Operator == ast.OpAnd {
		if !left {
			return false, nil
		}
		right, err := matchNode(n.Right, data)
		if err != nil {
			return false, err
		}
		return right, nil
	} else if n.Operator == ast.OpOr {
		if left {
			return true, nil
		}
		right, err := matchNode(n.Right, data)
		if err != nil {
			return false, err
		}
		return right, nil
	}

	return false, fmt.Errorf("unknown logical operator: %s", n.Operator)
}

func matchNot(n *ast.NotNode, data map[string]interface{}) (bool, error) {
	result, err := matchNode(n.Operand, data)
	if err != nil {
		return false, err
	}
	return !result, nil
}

func matchIn(n *ast.InNode, data map[string]interface{}) (bool, error) {
	colVal, exists := data[n.Column]
	if !exists {
		return false, fmt.Errorf("column not found: %s", n.Column)
	}

	if colVal == nil {
		result := n.Not
		return result, nil
	}

	found := false
	for _, val := range n.Values {
		match, err := valueEquals(colVal, val)
		if err != nil {
			return false, err
		}
		if match {
			found = true
			break
		}
	}

	if n.Not {
		return !found, nil
	}
	return found, nil
}

func valueEquals(a interface{}, b ast.Value) (bool, error) {
	switch b.Type {
	case ast.ValueTypeNumber:
		aNum, err := toFloat64(a)
		if err != nil {
			return false, err
		}
		bNum := b.Value.(float64)
		return aNum == bNum, nil
	case ast.ValueTypeString:
		aStr := fmt.Sprintf("%v", a)
		bStr := b.Value.(string)
		return aStr == bStr, nil
	case ast.ValueTypeNull:
		return a == nil, nil
	default:
		return false, fmt.Errorf("unsupported value type: %s", b.Type)
	}
}

func matchLike(n *ast.LikeNode, data map[string]interface{}) (bool, error) {
	colVal, exists := data[n.Column]
	if !exists {
		return false, fmt.Errorf("column not found: %s", n.Column)
	}

	if colVal == nil {
		return n.Not, nil
	}

	colStr := fmt.Sprintf("%v", colVal)
	matched := likeMatch(colStr, n.Pattern, n.Escape)

	if n.Not {
		return !matched, nil
	}
	return matched, nil
}

func likeMatch(text, pattern, escape string) bool {
	if escape == "" {
		escape = "\\"
	}
	escapeRune := rune(escape[0])

	textRunes := []rune(text)
	patternRunes := []rune(pattern)

	return likeMatchRecursive(textRunes, patternRunes, 0, 0, escapeRune)
}

func likeMatchRecursive(text, pattern []rune, tIdx, pIdx int, escape rune) bool {
	if pIdx >= len(pattern) {
		return tIdx >= len(text)
	}

	if pattern[pIdx] == escape && pIdx+1 < len(pattern) {
		pIdx++
		if tIdx >= len(text) {
			return false
		}
		if text[tIdx] == pattern[pIdx] {
			return likeMatchRecursive(text, pattern, tIdx+1, pIdx+1, escape)
		}
		return false
	}

	if pattern[pIdx] == '%' {
		if pIdx+1 >= len(pattern) {
			return true
		}
		for i := tIdx; i <= len(text); i++ {
			if likeMatchRecursive(text, pattern, i, pIdx+1, escape) {
				return true
			}
		}
		return false
	}

	if pattern[pIdx] == '_' {
		if tIdx >= len(text) {
			return false
		}
		return likeMatchRecursive(text, pattern, tIdx+1, pIdx+1, escape)
	}

	if tIdx >= len(text) {
		return false
	}

	if text[tIdx] == pattern[pIdx] {
		return likeMatchRecursive(text, pattern, tIdx+1, pIdx+1, escape)
	}

	return false
}

func matchBetween(n *ast.BetweenNode, data map[string]interface{}) (bool, error) {
	colVal, exists := data[n.Column]
	if !exists {
		return false, fmt.Errorf("column not found: %s", n.Column)
	}

	if colVal == nil {
		return n.Not, nil
	}

	var lower, upper float64
	var colNum float64
	var err error

	if n.Lower.Type == ast.ValueTypeNumber && n.Upper.Type == ast.ValueTypeNumber {
		lower = n.Lower.Value.(float64)
		upper = n.Upper.Value.(float64)
		colNum, err = toFloat64(colVal)
		if err != nil {
			return false, err
		}
		matched := colNum >= lower && colNum <= upper
		if n.Not {
			return !matched, nil
		}
		return matched, nil
	}

	colStr := fmt.Sprintf("%v", colVal)
	lowerStr := n.Lower.String()
	upperStr := n.Upper.String()

	matched := colStr >= lowerStr && colStr <= upperStr
	if n.Not {
		return !matched, nil
	}
	return matched, nil
}

func matchNullCheck(n *ast.NullCheckNode, data map[string]interface{}) (bool, error) {
	colVal, exists := data[n.Column]
	if !exists {
		return n.Not, nil
	}

	isNull := colVal == nil
	if n.Not {
		return !isNull, nil
	}
	return isNull, nil
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string to number: %s", val)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert to number: %v", v)
	}
}
