package deepcopy

import (
	"errors"
	"reflect"
)

var (
	ErrPrototypeNotFound = errors.New("prototype not found")
	ErrUnsupportedType   = errors.New("不支持拷贝的类型")
)

type CompareMode string

const (
	CompareShallow   CompareMode = "shallow"
	CompareDeep      CompareMode = "deep"
	CompareReference CompareMode = "reference"
)

type CompareConfig struct {
	visited map[comparePair]struct{}
}

type comparePair struct {
	a uintptr
	b uintptr
}

func Compare(a, b interface{}, mode CompareMode) (bool, error) {
	if mode == CompareReference {
		return compareReference(a, b), nil
	}

	if a == nil && b == nil {
		return true, nil
	}

	if a == nil || b == nil {
		return false, nil
	}

	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Type() != valB.Type() {
		return false, nil
	}

	if mode == CompareShallow {
		return compareShallow(valA, valB), nil
	}

	config := &CompareConfig{
		visited: make(map[comparePair]struct{}),
	}

	return compareDeep(valA, valB, config), nil
}

func compareReference(a, b interface{}) bool {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != valB.Kind() {
		return false
	}

	switch valA.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return valA.Pointer() == valB.Pointer()
	default:
		return reflect.DeepEqual(a, b)
	}
}

func compareShallow(a, b reflect.Value) bool {
	switch a.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if a.IsNil() != b.IsNil() {
			return false
		}
		if a.IsNil() {
			return true
		}
		return a.Pointer() == b.Pointer()
	case reflect.Interface:
		if a.IsNil() != b.IsNil() {
			return false
		}
		if a.IsNil() {
			return true
		}
		return compareShallow(a.Elem(), b.Elem())
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !compareShallow(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Array:
		for i := 0; i < a.Len(); i++ {
			if !compareShallow(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a.Interface(), b.Interface())
	}
}

func compareDeep(a, b reflect.Value, config *CompareConfig) bool {
	if a.Kind() != b.Kind() {
		return false
	}

	switch a.Kind() {
	case reflect.Ptr:
		return compareDeepPtr(a, b, config)

	case reflect.Interface:
		return compareDeepInterface(a, b, config)

	case reflect.Slice:
		return compareDeepSlice(a, b, config)

	case reflect.Array:
		return compareDeepArray(a, b, config)

	case reflect.Map:
		return compareDeepMap(a, b, config)

	case reflect.Struct:
		return compareDeepStruct(a, b, config)

	default:
		return reflect.DeepEqual(a.Interface(), b.Interface())
	}
}

func compareDeepPtr(a, b reflect.Value, config *CompareConfig) bool {
	if a.IsNil() != b.IsNil() {
		return false
	}

	if a.IsNil() {
		return true
	}

	ptrA := a.Pointer()
	ptrB := b.Pointer()

	pair := comparePair{a: ptrA, b: ptrB}
	if _, visited := config.visited[pair]; visited {
		return true
	}

	config.visited[pair] = struct{}{}
	defer delete(config.visited, pair)

	return compareDeep(a.Elem(), b.Elem(), config)
}

func compareDeepInterface(a, b reflect.Value, config *CompareConfig) bool {
	if a.IsNil() != b.IsNil() {
		return false
	}

	if a.IsNil() {
		return true
	}

	return compareDeep(a.Elem(), b.Elem(), config)
}

func compareDeepSlice(a, b reflect.Value, config *CompareConfig) bool {
	if a.IsNil() != b.IsNil() {
		return false
	}

	if a.IsNil() {
		return true
	}

	if a.Len() != b.Len() {
		return false
	}

	if a.Len() > 0 {
		ptrA := a.Pointer()
		ptrB := b.Pointer()

		pair := comparePair{a: ptrA, b: ptrB}
		if _, visited := config.visited[pair]; visited {
			return true
		}

		config.visited[pair] = struct{}{}
		defer delete(config.visited, pair)
	}

	for i := 0; i < a.Len(); i++ {
		if !compareDeep(a.Index(i), b.Index(i), config) {
			return false
		}
	}

	return true
}

func compareDeepArray(a, b reflect.Value, config *CompareConfig) bool {
	if a.Len() != b.Len() {
		return false
	}

	for i := 0; i < a.Len(); i++ {
		if !compareDeep(a.Index(i), b.Index(i), config) {
			return false
		}
	}

	return true
}

func compareDeepMap(a, b reflect.Value, config *CompareConfig) bool {
	if a.IsNil() != b.IsNil() {
		return false
	}

	if a.IsNil() {
		return true
	}

	if a.Len() != b.Len() {
		return false
	}

	ptrA := a.Pointer()
	ptrB := b.Pointer()

	pair := comparePair{a: ptrA, b: ptrB}
	if _, visited := config.visited[pair]; visited {
		return true
	}

	config.visited[pair] = struct{}{}
	defer delete(config.visited, pair)

	for _, key := range a.MapKeys() {
		valA := a.MapIndex(key)
		valB := b.MapIndex(key)

		if !valB.IsValid() {
			return false
		}

		if !compareDeep(valA, valB, config) {
			return false
		}
	}

	return true
}

func compareDeepStruct(a, b reflect.Value, config *CompareConfig) bool {
	for i := 0; i < a.NumField(); i++ {
		if !compareDeep(a.Field(i), b.Field(i), config) {
			return false
		}
	}

	return true
}
