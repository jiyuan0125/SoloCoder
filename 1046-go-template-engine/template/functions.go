package template

import (
	"fmt"
	"reflect"
	"strings"
)

func registerBuiltins(e *Engine) {
	e.RegisterFunction("upper", builtinUpper)
	e.RegisterFunction("lower", builtinLower)
	e.RegisterFunction("trim", builtinTrim)
	e.RegisterFunction("len", builtinLen)
	e.RegisterFunction("default", builtinDefault)
}

func builtinUpper(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("upper requires exactly 1 argument, got %d", len(args))
	}
	s, err := toString(args[0])
	if err != nil {
		return "", err
	}
	return strings.ToUpper(s), nil
}

func builtinLower(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("lower requires exactly 1 argument, got %d", len(args))
	}
	s, err := toString(args[0])
	if err != nil {
		return "", err
	}
	return strings.ToLower(s), nil
}

func builtinTrim(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("trim requires exactly 1 argument, got %d", len(args))
	}
	s, err := toString(args[0])
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func builtinLen(args ...interface{}) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("len requires exactly 1 argument, got %d", len(args))
	}
	v := reflect.ValueOf(args[0])
	var l int
	switch v.Kind() {
	case reflect.String:
		l = v.Len()
	case reflect.Array, reflect.Slice, reflect.Map:
		l = v.Len()
	case reflect.Invalid:
		l = 0
	default:
		return "", fmt.Errorf("len cannot be applied to type %T", args[0])
	}
	return fmt.Sprintf("%d", l), nil
}

func builtinDefault(args ...interface{}) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("default requires exactly 2 arguments, got %d", len(args))
	}
	val := args[0]
	def, err := toString(args[1])
	if err != nil {
		return "", err
	}
	if isEmpty(val) {
		return def, nil
	}
	s, err := toString(val)
	if err != nil {
		return def, nil
	}
	return s, nil
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return true
		}
		return isEmpty(rv.Elem().Interface())
	case reflect.Invalid:
		return true
	}
	return false
}

func toString(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return "", nil
		}
		return toString(rv.Elem().Interface())
	case reflect.String:
		return rv.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", rv.Float()), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", rv.Bool()), nil
	}
	return fmt.Sprintf("%v", v), nil
}
