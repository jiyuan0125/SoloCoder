package cbor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

func ToJSON(v Value) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(jsonValue(v)); err != nil {
		return nil, err
	}
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

func jsonValue(v Value) interface{} {
	switch val := v.(type) {
	case uint64:
		return float64(val)
	case int64:
		return float64(val)
	case uint, uint8, uint16, uint32, int, int8, int16, int32:
		return val
	case float32:
		return float64(val)
	case float64:
		return val
	case bool:
		return val
	case nil:
		return nil
	case SimpleValue:
		return fmt.Sprintf("simple:%d", val)
	case []byte:
		return base64.StdEncoding.EncodeToString(val)
	case string:
		return val
	case []Value:
		arr := make([]interface{}, len(val))
		for i, vv := range val {
			arr[i] = jsonValue(vv)
		}
		return arr
	case []interface{}:
		arr := make([]interface{}, len(val))
		for i, vv := range val {
			arr[i] = jsonValue(vv)
		}
		return arr
	case *Map:
		obj := make(map[string]interface{})
		keys := make([]string, 0, val.Len())
		for _, entry := range val.Entries {
			keyStr := keyToString(entry.Key)
			keys = append(keys, keyStr)
			obj[keyStr] = jsonValue(entry.Value)
		}
		return obj
	case *BigInt:
		if val.Negative {
			n := new(big.Int).Neg(val.Int)
			return n.String()
		}
		return val.Int.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func keyToString(k Value) string {
	switch key := k.(type) {
	case string:
		return key
	case uint64:
		return fmt.Sprintf("%d", key)
	case int64:
		return fmt.Sprintf("%d", key)
	case uint, uint8, uint16, uint32, int, int8, int16, int32:
		return fmt.Sprintf("%v", key)
	case float32:
		return fmt.Sprintf("%f", key)
	case float64:
		return fmt.Sprintf("%f", key)
	case bool:
		return fmt.Sprintf("%t", key)
	case []byte:
		return base64.StdEncoding.EncodeToString(key)
	case *BigInt:
		if key.Negative {
			n := new(big.Int).Neg(key.Int)
			return n.String()
		}
		return key.Int.String()
	default:
		return fmt.Sprintf("%v", key)
	}
}

func FromJSON(data []byte) (Value, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return fromJSONValue(v), nil
}

func fromJSONValue(v interface{}) Value {
	switch val := v.(type) {
	case float64:
		if float64(int64(val)) == val {
			return int64(val)
		}
		return val
	case string:
		return val
	case bool:
		return val
	case nil:
		return nil
	case []interface{}:
		arr := make([]Value, len(val))
		for i, vv := range val {
			arr[i] = fromJSONValue(vv)
		}
		return arr
	case map[string]interface{}:
		m := NewMap()
		for k, vv := range val {
			m.Add(k, fromJSONValue(vv))
		}
		return m
	default:
		return val
	}
}
