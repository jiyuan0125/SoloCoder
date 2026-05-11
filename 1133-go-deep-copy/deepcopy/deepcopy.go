package deepcopy

import (
	"fmt"
	"reflect"
	"unsafe"
)

type CopyConfig struct {
	visited map[uintptr]reflect.Value
}

func DeepCopy(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}

	config := &CopyConfig{
		visited: make(map[uintptr]reflect.Value),
	}

	srcVal := reflect.ValueOf(src)
	dstVal, err := copyValue(srcVal, config)
	if err != nil {
		return nil, err
	}

	return dstVal.Interface(), nil
}

func copyValue(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	switch src.Kind() {
	case reflect.Invalid:
		return reflect.Value{}, fmt.Errorf("invalid value")

	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return copyPrimitive(src)

	case reflect.Ptr:
		return copyPtr(src, config)

	case reflect.Interface:
		return copyInterface(src, config)

	case reflect.Slice:
		return copySlice(src, config)

	case reflect.Array:
		return copyArray(src, config)

	case reflect.Map:
		return copyMap(src, config)

	case reflect.Struct:
		return copyStruct(src, config)

	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return reflect.Value{}, fmt.Errorf("不支持拷贝的类型: %s", src.Kind())

	default:
		return reflect.Value{}, fmt.Errorf("未知类型: %s", src.Kind())
	}
}

func copyPrimitive(src reflect.Value) (reflect.Value, error) {
	dst := reflect.New(src.Type()).Elem()
	dst.Set(src)
	return dst, nil
}

func getPointerAddr(v reflect.Value) uintptr {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return v.Pointer()
	}
	if v.CanAddr() {
		return v.Addr().Pointer()
	}
	return uintptr(unsafe.Pointer(reflect.ValueOf(v).Pointer()))
}

func copyPtr(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	ptr := src.Pointer()
	if copied, exists := config.visited[ptr]; exists {
		return copied, nil
	}

	dst := reflect.New(src.Type().Elem())
	config.visited[ptr] = dst

	elemVal, err := copyValue(src.Elem(), config)
	if err != nil {
		return reflect.Value{}, err
	}

	dst.Elem().Set(elemVal)
	return dst, nil
}

func copyInterface(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	elem := src.Elem()
	copied, err := copyValue(elem, config)
	if err != nil {
		return reflect.Value{}, err
	}

	dst := reflect.New(src.Type()).Elem()
	if copied.IsValid() {
		dst.Set(copied)
	}
	return dst, nil
}

func copySlice(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	if src.Len() > 0 && src.Index(0).CanAddr() {
		ptr := uintptr(unsafe.Pointer(src.Pointer()))
		if copied, exists := config.visited[ptr]; exists {
			return copied, nil
		}
	}

	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())

	if src.Len() > 0 && src.Index(0).CanAddr() {
		ptr := uintptr(unsafe.Pointer(src.Pointer()))
		config.visited[ptr] = dst
	}

	for i := 0; i < src.Len(); i++ {
		elem, err := copyValue(src.Index(i), config)
		if err != nil {
			return reflect.Value{}, err
		}
		if elem.IsValid() {
			dst.Index(i).Set(elem)
		}
	}

	return dst, nil
}

func copyArray(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	dst := reflect.New(src.Type()).Elem()

	for i := 0; i < src.Len(); i++ {
		elem, err := copyValue(src.Index(i), config)
		if err != nil {
			return reflect.Value{}, err
		}
		if elem.IsValid() {
			dst.Index(i).Set(elem)
		}
	}

	return dst, nil
}

func copyMap(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	ptr := src.Pointer()
	if copied, exists := config.visited[ptr]; exists {
		return copied, nil
	}

	dst := reflect.MakeMap(src.Type())
	config.visited[ptr] = dst

	for _, key := range src.MapKeys() {
		newKey, err := copyValue(key, config)
		if err != nil {
			return reflect.Value{}, err
		}

		newValue, err := copyValue(src.MapIndex(key), config)
		if err != nil {
			return reflect.Value{}, err
		}

		if newKey.IsValid() && newValue.IsValid() {
			dst.SetMapIndex(newKey, newValue)
		}
	}

	return dst, nil
}

func getStructFieldTag(field reflect.StructField) string {
	return field.Tag.Get("copy")
}

func copyStruct(src reflect.Value, config *CopyConfig) (reflect.Value, error) {
	dst := reflect.New(src.Type()).Elem()

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		fieldType := src.Type().Field(i)
		tag := getStructFieldTag(fieldType)

		if tag == "skip" {
			continue
		}

		if tag == "shallow" {
			dst.Field(i).Set(srcField)
			continue
		}

		copied, err := copyValue(srcField, config)
		if err != nil {
			return reflect.Value{}, err
		}
		if copied.IsValid() {
			dst.Field(i).Set(copied)
		}
	}

	return dst, nil
}
