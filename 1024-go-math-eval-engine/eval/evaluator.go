package eval

type Variables map[string]interface{}

func Eval(node Node, vars Variables) *Value {
	switch node := node.(type) {
	case *Program:
		return Eval(node.Expression, vars)
	case *IntegerLiteral:
		return &Value{Type: VALUE_INT, Int: node.Value}
	case *FloatLiteral:
		return &Value{Type: VALUE_FLOAT, Float: node.Value}
	case *BooleanLiteral:
		return &Value{Type: VALUE_BOOL, Bool: node.Value}
	case *Identifier:
		if value, ok := vars[node.Value]; ok {
			return convertToValue(value)
		}
		return &Value{Type: VALUE_ERROR, Msg: "undefined variable: " + node.Value}
	case *PrefixExpression:
		right := Eval(node.Right, vars)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *InfixExpression:
		left := Eval(node.Left, vars)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, vars)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *CallExpression:
		fnIdent, ok := node.Function.(*Identifier)
		if !ok {
			return &Value{Type: VALUE_ERROR, Msg: "invalid function call"}
		}
		if !isBuiltinFunction(fnIdent.Value) {
			return &Value{Type: VALUE_ERROR, Msg: "undefined function: " + fnIdent.Value}
		}
		args := evalExpressions(node.Arguments, vars)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyBuiltinFunction(fnIdent.Value, args)
	}
	return nil
}

func evalExpressions(exps []Expression, vars Variables) []*Value {
	result := make([]*Value, 0)
	for _, e := range exps {
		evaluated := Eval(e, vars)
		if isError(evaluated) {
			return []*Value{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalPrefixExpression(operator string, right *Value) *Value {
	switch operator {
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	case "+":
		return evalPlusPrefixOperatorExpression(right)
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: " + operator}
}

func evalMinusPrefixOperatorExpression(right *Value) *Value {
	switch right.Type {
	case VALUE_INT:
		return &Value{Type: VALUE_INT, Int: -right.Int}
	case VALUE_FLOAT:
		return &Value{Type: VALUE_FLOAT, Float: -right.Float}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: -" + string(right.Type)}
}

func evalPlusPrefixOperatorExpression(right *Value) *Value {
	switch right.Type {
	case VALUE_INT:
		return &Value{Type: VALUE_INT, Int: right.Int}
	case VALUE_FLOAT:
		return &Value{Type: VALUE_FLOAT, Float: right.Float}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: +" + string(right.Type)}
}

func evalInfixExpression(operator string, left, right *Value) *Value {
	switch {
	case left.Type == VALUE_INT && right.Type == VALUE_INT:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type == VALUE_INT && right.Type == VALUE_FLOAT:
		left = &Value{Type: VALUE_FLOAT, Float: float64(left.Int)}
		return evalFloatInfixExpression(operator, left, right)
	case left.Type == VALUE_FLOAT && right.Type == VALUE_INT:
		right = &Value{Type: VALUE_FLOAT, Float: float64(right.Int)}
		return evalFloatInfixExpression(operator, left, right)
	case left.Type == VALUE_FLOAT && right.Type == VALUE_FLOAT:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type == VALUE_BOOL && right.Type == VALUE_BOOL:
		return evalBooleanInfixExpression(operator, left, right)
	case operator == "and" || operator == "or":
		return evalLogicalInfixExpression(operator, left, right)
	}
	return &Value{Type: VALUE_ERROR, Msg: "type mismatch: " + string(left.Type) + " " + operator + " " + string(right.Type)}
}

func evalIntegerInfixExpression(operator string, left, right *Value) *Value {
	leftVal := left.Int
	rightVal := right.Int
	switch operator {
	case "+":
		return &Value{Type: VALUE_INT, Int: leftVal + rightVal}
	case "-":
		return &Value{Type: VALUE_INT, Int: leftVal - rightVal}
	case "*":
		return &Value{Type: VALUE_INT, Int: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return &Value{Type: VALUE_ERROR, Msg: "division by zero"}
		}
		if leftVal%rightVal == 0 {
			return &Value{Type: VALUE_INT, Int: leftVal / rightVal}
		}
		return &Value{Type: VALUE_FLOAT, Float: float64(leftVal) / float64(rightVal)}
	case "<":
		return &Value{Type: VALUE_BOOL, Bool: leftVal < rightVal}
	case ">":
		return &Value{Type: VALUE_BOOL, Bool: leftVal > rightVal}
	case "<=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal <= rightVal}
	case ">=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal >= rightVal}
	case "==":
		return &Value{Type: VALUE_BOOL, Bool: leftVal == rightVal}
	case "!=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal != rightVal}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: " + operator}
}

func evalFloatInfixExpression(operator string, left, right *Value) *Value {
	leftVal := left.Float
	rightVal := right.Float
	switch operator {
	case "+":
		return &Value{Type: VALUE_FLOAT, Float: leftVal + rightVal}
	case "-":
		return &Value{Type: VALUE_FLOAT, Float: leftVal - rightVal}
	case "*":
		return &Value{Type: VALUE_FLOAT, Float: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return &Value{Type: VALUE_ERROR, Msg: "division by zero"}
		}
		return &Value{Type: VALUE_FLOAT, Float: leftVal / rightVal}
	case "<":
		return &Value{Type: VALUE_BOOL, Bool: leftVal < rightVal}
	case ">":
		return &Value{Type: VALUE_BOOL, Bool: leftVal > rightVal}
	case "<=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal <= rightVal}
	case ">=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal >= rightVal}
	case "==":
		return &Value{Type: VALUE_BOOL, Bool: leftVal == rightVal}
	case "!=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal != rightVal}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: " + operator}
}

func evalBooleanInfixExpression(operator string, left, right *Value) *Value {
	leftVal := left.Bool
	rightVal := right.Bool
	switch operator {
	case "and":
		return &Value{Type: VALUE_BOOL, Bool: leftVal && rightVal}
	case "or":
		return &Value{Type: VALUE_BOOL, Bool: leftVal || rightVal}
	case "==":
		return &Value{Type: VALUE_BOOL, Bool: leftVal == rightVal}
	case "!=":
		return &Value{Type: VALUE_BOOL, Bool: leftVal != rightVal}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: " + operator}
}

func evalLogicalInfixExpression(operator string, left, right *Value) *Value {
	leftBool := toBoolean(left)
	rightBool := toBoolean(right)
	switch operator {
	case "and":
		return &Value{Type: VALUE_BOOL, Bool: leftBool && rightBool}
	case "or":
		return &Value{Type: VALUE_BOOL, Bool: leftBool || rightBool}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unknown operator: " + operator}
}

func toBoolean(v *Value) bool {
	switch v.Type {
	case VALUE_INT:
		return v.Int != 0
	case VALUE_FLOAT:
		return v.Float != 0
	case VALUE_BOOL:
		return v.Bool
	}
	return false
}

func isError(v *Value) bool {
	if v != nil {
		return v.Type == VALUE_ERROR
	}
	return false
}

func convertToValue(value interface{}) *Value {
	switch v := value.(type) {
	case int:
		return &Value{Type: VALUE_INT, Int: int64(v)}
	case int32:
		return &Value{Type: VALUE_INT, Int: int64(v)}
	case int64:
		return &Value{Type: VALUE_INT, Int: v}
	case float32:
		return &Value{Type: VALUE_FLOAT, Float: float64(v)}
	case float64:
		return &Value{Type: VALUE_FLOAT, Float: v}
	case bool:
		return &Value{Type: VALUE_BOOL, Bool: v}
	}
	return &Value{Type: VALUE_ERROR, Msg: "unsupported variable type"}
}

func applyBuiltinFunction(fnName string, args []*Value) *Value {
	switch fnName {
	case "abs":
		return evalAbs(args)
	case "sqrt":
		return evalSqrt(args)
	case "max":
		return evalMax(args)
	case "min":
		return evalMin(args)
	case "ceil":
		return evalCeil(args)
	case "floor":
		return evalFloor(args)
	case "round":
		return evalRound(args)
	}
	return &Value{Type: VALUE_ERROR, Msg: "undefined function: " + fnName}
}
