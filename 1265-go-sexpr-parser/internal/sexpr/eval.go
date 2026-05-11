package sexpr

import (
	"fmt"
)

type Env struct {
	parent *Env
	vars   map[string]Expr
}

func NewEnv(parent *Env) *Env {
	return &Env{
		parent: parent,
		vars:   make(map[string]Expr),
	}
}

func (e *Env) Get(name string) (Expr, bool) {
	if val, ok := e.vars[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Env) Set(name string, value Expr) {
	e.vars[name] = value
}

func (e *Env) Define(name string, value Expr) {
	e.vars[name] = value
}

type EvalError struct {
	Msg string
	Pos Position
}

func (e *EvalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
}

func evalError(msg string, pos Position) error {
	return &EvalError{Msg: msg, Pos: pos}
}

func Eval(expr Expr, env *Env) (Expr, error) {
	switch v := expr.(type) {
	case *Atom:
		switch v.Type {
		case AtomInt, AtomFloat, AtomString:
			return v, nil
		case AtomSymbol:
			if v.SymbolValue == "true" {
				return NewSymbol("true", v.pos), nil
			}
			if v.SymbolValue == "false" {
				return NewSymbol("false", v.pos), nil
			}
			if val, ok := env.Get(v.SymbolValue); ok {
				return val, nil
			}
			return nil, evalError(fmt.Sprintf("undefined variable: %s", v.SymbolValue), v.pos)
		}
	case *List:
		if v.IsNil() {
			return v, nil
		}

		first := v.Elements[0]
		switch f := first.(type) {
		case *Atom:
			if f.Type == AtomSymbol {
				switch f.SymbolValue {
				case "quote":
					if len(v.Elements) != 2 {
						return nil, evalError("quote requires exactly 1 argument", v.pos)
					}
					return v.Elements[1], nil
				case "define":
					return evalDefine(v, env)
				case "if":
					return evalIf(v, env)
				case "+":
					return evalArithmetic(v, env, "+")
				case "-":
					return evalArithmetic(v, env, "-")
				case "*":
					return evalArithmetic(v, env, "*")
				case "/":
					return evalArithmetic(v, env, "/")
				case "=":
					return evalComparison(v, env, "=")
				case "<":
					return evalComparison(v, env, "<")
				case ">":
					return evalComparison(v, env, ">")
				}
			}
		}

		return evalFunctionCall(v, env)
	}

	return nil, evalError("invalid expression", expr.Pos())
}

func evalDefine(list *List, env *Env) (Expr, error) {
	if len(list.Elements) < 3 {
		return nil, evalError("define requires at least 2 arguments", list.pos)
	}

	name, ok := list.Elements[1].(*Atom)
	if !ok || name.Type != AtomSymbol {
		return nil, evalError("define requires a symbol as first argument", list.Elements[1].Pos())
	}

	value, err := Eval(list.Elements[2], env)
	if err != nil {
		return nil, err
	}

	env.Define(name.SymbolValue, value)
	return value, nil
}

func evalIf(list *List, env *Env) (Expr, error) {
	if len(list.Elements) != 4 {
		return nil, evalError("if requires exactly 3 arguments (condition, then, else)", list.pos)
	}

	cond, err := Eval(list.Elements[1], env)
	if err != nil {
		return nil, err
	}

	if isTruthy(cond) {
		return Eval(list.Elements[2], env)
	}
	return Eval(list.Elements[3], env)
}

func isTruthy(expr Expr) bool {
	if expr.IsNil() {
		return false
	}
	if atom, ok := expr.(*Atom); ok && atom.Type == AtomSymbol && atom.SymbolValue == "false" {
		return false
	}
	return true
}

func evalArithmetic(list *List, env *Env, op string) (Expr, error) {
	if len(list.Elements) < 2 {
		return nil, evalError(fmt.Sprintf("%s requires at least 1 argument", op), list.pos)
	}

	var resultNum bool
	var intResult int64
	var floatResult float64

	for i := 1; i < len(list.Elements); i++ {
		arg, err := Eval(list.Elements[i], env)
		if err != nil {
			return nil, err
		}

		atom, ok := arg.(*Atom)
		if !ok || (atom.Type != AtomInt && atom.Type != AtomFloat) {
			return nil, evalError(fmt.Sprintf("%s requires numeric arguments", op), arg.Pos())
		}

		if i == 1 {
			if atom.Type == AtomInt {
				resultNum = false
				intResult = atom.IntValue
			} else {
				resultNum = true
				floatResult = atom.FloatValue
			}
			if op == "-" && len(list.Elements) == 2 {
				if resultNum {
					floatResult = -floatResult
				} else {
					intResult = -intResult
				}
			}
			continue
		}

		var argInt int64
		var argFloat float64
		var argIsFloat bool

		if atom.Type == AtomInt {
			argInt = atom.IntValue
			argFloat = float64(atom.IntValue)
		} else {
			argIsFloat = true
			argFloat = atom.FloatValue
			argInt = int64(atom.FloatValue)
		}

		if resultNum || argIsFloat {
			if !resultNum {
				floatResult = float64(intResult)
				resultNum = true
			}
			switch op {
			case "+":
				floatResult += argFloat
			case "-":
				floatResult -= argFloat
			case "*":
				floatResult *= argFloat
			case "/":
				if argFloat == 0 {
					return nil, evalError("division by zero", arg.Pos())
				}
				floatResult /= argFloat
			}
		} else {
			switch op {
			case "+":
				intResult += argInt
			case "-":
				intResult -= argInt
			case "*":
				intResult *= argInt
			case "/":
				if argInt == 0 {
					return nil, evalError("division by zero", arg.Pos())
				}
				intResult /= argInt
			}
		}
	}

	if resultNum {
		return NewFloat(floatResult, list.pos), nil
	}
	return NewInt(intResult, list.pos), nil
}

func evalComparison(list *List, env *Env, op string) (Expr, error) {
	if len(list.Elements) != 3 {
		return nil, evalError(fmt.Sprintf("%s requires exactly 2 arguments", op), list.pos)
	}

	left, err := Eval(list.Elements[1], env)
	if err != nil {
		return nil, err
	}

	right, err := Eval(list.Elements[2], env)
	if err != nil {
		return nil, err
	}

	result := compare(left, right, op)
	if result {
		return NewSymbol("true", list.pos), nil
	}
	return NewSymbol("false", list.pos), nil
}

func compare(a, b Expr, op string) bool {
	atomA, okA := a.(*Atom)
	atomB, okB := b.(*Atom)

	if !okA || !okB {
		return false
	}

	var numA, numB float64
	var isNumA, isNumB bool

	switch atomA.Type {
	case AtomInt:
		numA = float64(atomA.IntValue)
		isNumA = true
	case AtomFloat:
		numA = atomA.FloatValue
		isNumA = true
	}

	switch atomB.Type {
	case AtomInt:
		numB = float64(atomB.IntValue)
		isNumB = true
	case AtomFloat:
		numB = atomB.FloatValue
		isNumB = true
	}

	if isNumA && isNumB {
		switch op {
		case "=":
			return numA == numB
		case "<":
			return numA < numB
		case ">":
			return numA > numB
		}
	}

	if atomA.Type == atomB.Type {
		switch atomA.Type {
		case AtomString:
			switch op {
			case "=":
				return atomA.StringValue == atomB.StringValue
			}
		case AtomSymbol:
			switch op {
			case "=":
				return atomA.SymbolValue == atomB.SymbolValue
			}
		}
	}

	return false
}

func evalFunctionCall(list *List, env *Env) (Expr, error) {
	return list, nil
}

func EvalString(input string) (Expr, error) {
	exprs, err := Parse(input)
	if err != nil {
		return nil, err
	}

	if len(exprs) == 0 {
		return nil, evalError("no expression to evaluate", Position{Line: 1, Column: 1})
	}

	env := NewEnv(nil)
	var last Expr
	for _, expr := range exprs {
		last, err = Eval(expr, env)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}
