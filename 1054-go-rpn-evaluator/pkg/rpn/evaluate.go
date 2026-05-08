package rpn

import (
	"math"
	"strconv"
)

const epsilon = 1e-12

func EvaluateRPN(terms []RPNTerm, vars *VariableStore) (float64, error) {
	var stack []float64

	for _, term := range terms {
		switch term.Type {
		case TermNumber:
			val, err := strconv.ParseFloat(term.Value, 64)
			if err != nil {
				return 0, &ParseError{Msg: "invalid number: " + term.Value}
			}
			stack = append(stack, val)

		case TermIdentifier:
			if vars == nil {
				return 0, &VariableError{Msg: "undefined variable: " + term.Value}
			}
			val, ok := vars.Get(term.Value)
			if !ok {
				return 0, &VariableError{Msg: "undefined variable: " + term.Value}
			}
			stack = append(stack, val)

		case TermOperator:
			if term.IsUnary {
				if len(stack) < 1 {
					return 0, &EvalError{Msg: "insufficient operands for unary operator"}
				}
				a := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				switch term.Value {
				case "-":
					stack = append(stack, -a)
				case "+":
					stack = append(stack, a)
				default:
					return 0, &EvalError{Msg: "unknown unary operator: " + term.Value}
				}
			} else {
				if len(stack) < 2 {
					return 0, &EvalError{Msg: "insufficient operands for binary operator"}
				}
				b := stack[len(stack)-1]
				a := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				switch term.Value {
				case "+":
					stack = append(stack, a+b)
				case "-":
					stack = append(stack, a-b)
				case "*":
					stack = append(stack, a*b)
				case "/":
					if math.Abs(b) < epsilon {
						return 0, &EvalError{Msg: "division by zero"}
					}
					stack = append(stack, a/b)
				default:
					return 0, &EvalError{Msg: "unknown operator: " + term.Value}
				}
			}

		case TermFunction:
			if len(stack) < term.ArgCount {
				return 0, &EvalError{Msg: "insufficient arguments for function: " + term.Value}
			}
			args := make([]float64, term.ArgCount)
			for i := 0; i < term.ArgCount; i++ {
				args[term.ArgCount-1-i] = stack[len(stack)-1-i]
			}
			stack = stack[:len(stack)-term.ArgCount]

			fn, ok := GetFunction(term.Value)
			if !ok {
				return 0, &FunctionError{Msg: "undefined function: " + term.Value}
			}

			result, err := fn(args)
			if err != nil {
				return 0, err
			}
			stack = append(stack, result)
		}
	}

	if len(stack) != 1 {
		return 0, &EvalError{Msg: "invalid expression"}
	}

	return stack[0], nil
}

func EvaluateInfix(expr string, vars *VariableStore) (float64, error) {
	terms, err := InfixToRPN(expr)
	if err != nil {
		return 0, err
	}
	return EvaluateRPN(terms, vars)
}

func FormatResult(val float64) string {
	if math.Abs(val-math.Round(val)) < epsilon {
		return strconv.FormatInt(int64(math.Round(val)), 10)
	}
	return strconv.FormatFloat(val, 'f', -1, 64)
}
