package testfuncs

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"strings"
)

type CustomFunction struct {
	Expression string
	variables  []string
	compiled   func(map[string]float64) float64
}

func ParseCustomFunction(expr string, variables []string) (*CustomFunction, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, errors.New("empty expression")
	}

	fullExpr := wrapExpression(expr)
	astExpr, err := parser.ParseExpr(fullExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid expression: %v", err)
	}

	compiled, err := compileExpression(astExpr, variables)
	if err != nil {
		return nil, err
	}

	return &CustomFunction{
		Expression: expr,
		variables:  append([]string{}, variables...),
		compiled:   compiled,
	}, nil
}

func (cf *CustomFunction) Evaluate(x []float64) float64 {
	vars := make(map[string]float64)
	for i, name := range cf.variables {
		if i < len(x) {
			vars[name] = x[i]
		}
	}
	return cf.compiled(vars)
}

func (cf *CustomFunction) Variables() []string {
	return append([]string{}, cf.variables...)
}

func wrapExpression(expr string) string {
	return fmt.Sprintf("func() float64 { return %s }()", expr)
}

func compileExpression(node ast.Expr, allowedVars []string) (func(map[string]float64) float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		if n.Kind == token.FLOAT || n.Kind == token.INT {
			val, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				return nil, err
			}
			return func(map[string]float64) float64 { return val }, nil
		}
		return nil, fmt.Errorf("unsupported literal: %s", n.Value)

	case *ast.Ident:
		if n.Name == "pi" {
			return func(map[string]float64) float64 { return math.Pi }, nil
		}
		if n.Name == "e" {
			return func(map[string]float64) float64 { return math.E }, nil
		}
		for _, v := range allowedVars {
			if strings.EqualFold(v, n.Name) {
				varName := n.Name
				return func(vars map[string]float64) float64 {
					if val, ok := vars[varName]; ok {
						return val
					}
					return vars[strings.ToLower(varName)]
				}, nil
			}
		}
		return nil, fmt.Errorf("unknown variable: %s (allowed: %v)", n.Name, allowedVars)

	case *ast.ParenExpr:
		return compileExpression(n.X, allowedVars)

	case *ast.BinaryExpr:
		left, err := compileExpression(n.X, allowedVars)
		if err != nil {
			return nil, err
		}
		right, err := compileExpression(n.Y, allowedVars)
		if err != nil {
			return nil, err
		}

		switch n.Op {
		case token.ADD:
			return func(vars map[string]float64) float64 { return left(vars) + right(vars) }, nil
		case token.SUB:
			return func(vars map[string]float64) float64 { return left(vars) - right(vars) }, nil
		case token.MUL:
			return func(vars map[string]float64) float64 { return left(vars) * right(vars) }, nil
		case token.QUO:
			return func(vars map[string]float64) float64 { return left(vars) / right(vars) }, nil
		case token.REM:
			return func(vars map[string]float64) float64 {
				a := left(vars)
				b := right(vars)
				return math.Mod(a, b)
			}, nil
		default:
			return nil, fmt.Errorf("unsupported operator: %s", n.Op)
		}

	case *ast.CallExpr:
		fn, ok := n.Fun.(*ast.Ident)
		if !ok {
			return nil, errors.New("unsupported function call")
		}

		var args []func(map[string]float64) float64
		for _, arg := range n.Args {
			compiled, err := compileExpression(arg, allowedVars)
			if err != nil {
				return nil, err
			}
			args = append(args, compiled)
		}

		funcName := strings.ToLower(fn.Name)
		switch funcName {
		case "sin":
			return func(vars map[string]float64) float64 { return math.Sin(args[0](vars)) }, nil
		case "cos":
			return func(vars map[string]float64) float64 { return math.Cos(args[0](vars)) }, nil
		case "tan":
			return func(vars map[string]float64) float64 { return math.Tan(args[0](vars)) }, nil
		case "exp":
			return func(vars map[string]float64) float64 { return math.Exp(args[0](vars)) }, nil
		case "log":
			return func(vars map[string]float64) float64 { return math.Log(args[0](vars)) }, nil
		case "log10":
			return func(vars map[string]float64) float64 { return math.Log10(args[0](vars)) }, nil
		case "sqrt":
			return func(vars map[string]float64) float64 { return math.Sqrt(args[0](vars)) }, nil
		case "abs":
			return func(vars map[string]float64) float64 { return math.Abs(args[0](vars)) }, nil
		case "pow":
			if len(args) != 2 {
				return nil, errors.New("pow requires 2 arguments")
			}
			return func(vars map[string]float64) float64 { return math.Pow(args[0](vars), args[1](vars)) }, nil
		case "min":
			if len(args) < 2 {
				return nil, errors.New("min requires at least 2 arguments")
			}
			return func(vars map[string]float64) float64 {
				result := args[0](vars)
				for _, a := range args[1:] {
					val := a(vars)
					if val < result {
						result = val
					}
				}
				return result
			}, nil
		case "max":
			if len(args) < 2 {
				return nil, errors.New("max requires at least 2 arguments")
			}
			return func(vars map[string]float64) float64 {
				result := args[0](vars)
				for _, a := range args[1:] {
					val := a(vars)
					if val > result {
						result = val
					}
				}
				return result
			}, nil
		default:
			return nil, fmt.Errorf("unsupported function: %s", fn.Name)
		}

	case *ast.UnaryExpr:
		operand, err := compileExpression(n.X, allowedVars)
		if err != nil {
			return nil, err
		}
		switch n.Op {
		case token.SUB:
			return func(vars map[string]float64) float64 { return -operand(vars) }, nil
		case token.ADD:
			return func(vars map[string]float64) float64 { return operand(vars) }, nil
		default:
			return nil, fmt.Errorf("unsupported unary operator: %s", n.Op)
		}

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", node)
	}
}
