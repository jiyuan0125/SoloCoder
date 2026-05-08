package rpn

import (
	"math"
)

type Function func(args []float64) (float64, error)

var builtinFunctions = map[string]Function{
	"abs": funcAbs,
	"max": funcMax,
	"min": funcMin,
	"sqrt": funcSqrt,
	"sin": funcSin,
	"cos": funcCos,
	"tan": funcTan,
}

func GetFunction(name string) (Function, bool) {
	f, ok := builtinFunctions[name]
	return f, ok
}

func funcAbs(args []float64) (float64, error) {
	if len(args) != 1 {
		return 0, &FunctionError{Msg: "abs requires exactly 1 argument"}
	}
	return math.Abs(args[0]), nil
}

func funcMax(args []float64) (float64, error) {
	if len(args) < 2 {
		return 0, &FunctionError{Msg: "max requires at least 2 arguments"}
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		if args[i] > result {
			result = args[i]
		}
	}
	return result, nil
}

func funcMin(args []float64) (float64, error) {
	if len(args) < 2 {
		return 0, &FunctionError{Msg: "min requires at least 2 arguments"}
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		if args[i] < result {
			result = args[i]
		}
	}
	return result, nil
}

func funcSqrt(args []float64) (float64, error) {
	if len(args) != 1 {
		return 0, &FunctionError{Msg: "sqrt requires exactly 1 argument"}
	}
	if args[0] < 0 {
		return 0, &FunctionError{Msg: "sqrt of negative number"}
	}
	return math.Sqrt(args[0]), nil
}

func funcSin(args []float64) (float64, error) {
	if len(args) != 1 {
		return 0, &FunctionError{Msg: "sin requires exactly 1 argument"}
	}
	return math.Sin(args[0]), nil
}

func funcCos(args []float64) (float64, error) {
	if len(args) != 1 {
		return 0, &FunctionError{Msg: "cos requires exactly 1 argument"}
	}
	return math.Cos(args[0]), nil
}

func funcTan(args []float64) (float64, error) {
	if len(args) != 1 {
		return 0, &FunctionError{Msg: "tan requires exactly 1 argument"}
	}
	return math.Tan(args[0]), nil
}
