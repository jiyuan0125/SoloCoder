package engine

import (
	"strconv"
	"strings"
)

func EvaluateNode(n *Node, env map[string]string) (bool, error) {
	if n == nil {
		return false, nil
	}

	switch n.Type {
	case NodeTypeLogic:
		return evaluateLogic(n, env)
	case NodeTypeComparison:
		return evaluateComparison(n, env)
	default:
		return false, nil
	}
}

func evaluateLogic(n *Node, env map[string]string) (bool, error) {
	if len(n.Children) < 2 {
		return false, nil
	}

	left, err := EvaluateNode(n.Children[0], env)
	if err != nil {
		return false, err
	}

	if n.LogicOp == LogicOpAND {
		if !left {
			return false, nil
		}
	} else if n.LogicOp == LogicOpOR {
		if left {
			return true, nil
		}
	}

	right, err := EvaluateNode(n.Children[1], env)
	if err != nil {
		return false, err
	}

	if n.LogicOp == LogicOpAND {
		return left && right, nil
	}
	return left || right, nil
}

func evaluateComparison(n *Node, env map[string]string) (bool, error) {
	leftRaw, ok := env[n.LeftVar]
	if !ok {
		return false, nil
	}
	left := strings.TrimSpace(leftRaw)
	right := strings.TrimSpace(n.RightValue)

	leftNum, leftNumOk := tryParseFloat(left)
	rightNum, rightNumOk := tryParseFloat(right)

	if leftNumOk && rightNumOk {
		return compareNumbers(leftNum, rightNum, n.CompOp), nil
	}

	return compareStrings(left, right, n.CompOp), nil
}

func tryParseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

func compareNumbers(a, b float64, op CompOp) bool {
	switch op {
	case CompOpGT:
		return a > b
	case CompOpGTE:
		return a >= b
	case CompOpLT:
		return a < b
	case CompOpLTE:
		return a <= b
	case CompOpEQ:
		return a == b
	case CompOpNEQ:
		return a != b
	}
	return false
}

func compareStrings(a, b string, op CompOp) bool {
	switch op {
	case CompOpGT:
		return a > b
	case CompOpGTE:
		return a >= b
	case CompOpLT:
		return a < b
	case CompOpLTE:
		return a <= b
	case CompOpEQ:
		return a == b
	case CompOpNEQ:
		return a != b
	}
	return false
}

func EvaluateCondition(condition string, env map[string]string) (bool, error) {
	node, err := ParseExpression(condition)
	if err != nil {
		return false, err
	}
	return EvaluateNode(node, env)
}
