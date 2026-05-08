package eval

import "math"

var builtinFunctions = []string{
	"abs", "sqrt", "max", "min", "ceil", "floor", "round",
}

func isBuiltinFunction(name string) bool {
	for _, fn := range builtinFunctions {
		if fn == name {
			return true
		}
	}
	return false
}

func evalAbs(args []*Value) *Value {
	if len(args) != 1 {
		return &Value{Type: VALUE_ERROR, Msg: "abs requires exactly 1 argument, got " + itoa(len(args))}
	}
	arg := args[0]
	switch arg.Type {
	case VALUE_INT:
		if arg.Int < 0 {
			return &Value{Type: VALUE_INT, Int: -arg.Int}
		}
		return &Value{Type: VALUE_INT, Int: arg.Int}
	case VALUE_FLOAT:
		return &Value{Type: VALUE_FLOAT, Float: math.Abs(arg.Float)}
	}
	return &Value{Type: VALUE_ERROR, Msg: "abs requires numeric argument"}
}

func evalSqrt(args []*Value) *Value {
	if len(args) != 1 {
		return &Value{Type: VALUE_ERROR, Msg: "sqrt requires exactly 1 argument, got " + itoa(len(args))}
	}
	arg := args[0]
	var val float64
	switch arg.Type {
	case VALUE_INT:
		val = float64(arg.Int)
	case VALUE_FLOAT:
		val = arg.Float
	default:
		return &Value{Type: VALUE_ERROR, Msg: "sqrt requires numeric argument"}
	}
	if val < 0 {
		return &Value{Type: VALUE_ERROR, Msg: "sqrt argument must be non-negative"}
	}
	result := math.Sqrt(val)
	if isInteger(result) {
		return &Value{Type: VALUE_INT, Int: int64(result)}
	}
	return &Value{Type: VALUE_FLOAT, Float: result}
}

func evalMax(args []*Value) *Value {
	if len(args) < 2 {
		return &Value{Type: VALUE_ERROR, Msg: "max requires at least 2 arguments, got " + itoa(len(args))}
	}
	var maxVal *Value
	for i, arg := range args {
		if arg.Type != VALUE_INT && arg.Type != VALUE_FLOAT {
			return &Value{Type: VALUE_ERROR, Msg: "max requires numeric arguments"}
		}
		if i == 0 {
			maxVal = arg
			continue
		}
		cmp := compare(maxVal, arg)
		if cmp < 0 {
			maxVal = arg
		}
	}
	return maxVal
}

func evalMin(args []*Value) *Value {
	if len(args) < 2 {
		return &Value{Type: VALUE_ERROR, Msg: "min requires at least 2 arguments, got " + itoa(len(args))}
	}
	var minVal *Value
	for i, arg := range args {
		if arg.Type != VALUE_INT && arg.Type != VALUE_FLOAT {
			return &Value{Type: VALUE_ERROR, Msg: "min requires numeric arguments"}
		}
		if i == 0 {
			minVal = arg
			continue
		}
		cmp := compare(minVal, arg)
		if cmp > 0 {
			minVal = arg
		}
	}
	return minVal
}

func evalCeil(args []*Value) *Value {
	if len(args) != 1 {
		return &Value{Type: VALUE_ERROR, Msg: "ceil requires exactly 1 argument, got " + itoa(len(args))}
	}
	arg := args[0]
	switch arg.Type {
	case VALUE_INT:
		return &Value{Type: VALUE_INT, Int: arg.Int}
	case VALUE_FLOAT:
		return &Value{Type: VALUE_INT, Int: int64(math.Ceil(arg.Float))}
	}
	return &Value{Type: VALUE_ERROR, Msg: "ceil requires numeric argument"}
}

func evalFloor(args []*Value) *Value {
	if len(args) != 1 {
		return &Value{Type: VALUE_ERROR, Msg: "floor requires exactly 1 argument, got " + itoa(len(args))}
	}
	arg := args[0]
	switch arg.Type {
	case VALUE_INT:
		return &Value{Type: VALUE_INT, Int: arg.Int}
	case VALUE_FLOAT:
		return &Value{Type: VALUE_INT, Int: int64(math.Floor(arg.Float))}
	}
	return &Value{Type: VALUE_ERROR, Msg: "floor requires numeric argument"}
}

func evalRound(args []*Value) *Value {
	if len(args) != 1 {
		return &Value{Type: VALUE_ERROR, Msg: "round requires exactly 1 argument, got " + itoa(len(args))}
	}
	arg := args[0]
	switch arg.Type {
	case VALUE_INT:
		return &Value{Type: VALUE_INT, Int: arg.Int}
	case VALUE_FLOAT:
		return &Value{Type: VALUE_INT, Int: int64(math.Round(arg.Float))}
	}
	return &Value{Type: VALUE_ERROR, Msg: "round requires numeric argument"}
}

func compare(a, b *Value) int {
	var aVal, bVal float64
	switch a.Type {
	case VALUE_INT:
		aVal = float64(a.Int)
	case VALUE_FLOAT:
		aVal = a.Float
	}
	switch b.Type {
	case VALUE_INT:
		bVal = float64(b.Int)
	case VALUE_FLOAT:
		bVal = b.Float
	}
	if aVal < bVal {
		return -1
	} else if aVal > bVal {
		return 1
	}
	return 0
}

func isInteger(f float64) bool {
	return f == float64(int64(f))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
