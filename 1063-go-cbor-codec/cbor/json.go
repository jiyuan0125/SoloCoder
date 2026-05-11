package cbor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
)

func ToJSON(v Value) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := &jsonEncoder{w: buf}
	if err := enc.encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type jsonEncoder struct {
	w io.Writer
}

func (e *jsonEncoder) encode(v Value) error {
	switch val := v.(type) {
	case uint64:
		return e.writeFloat(float64(val))
	case int64:
		return e.writeFloat(float64(val))
	case uint:
		return e.writeInt(int64(val))
	case uint8:
		return e.writeInt(int64(val))
	case uint16:
		return e.writeInt(int64(val))
	case uint32:
		return e.writeInt(int64(val))
	case int:
		return e.writeInt(int64(val))
	case int8:
		return e.writeInt(int64(val))
	case int16:
		return e.writeInt(int64(val))
	case int32:
		return e.writeInt(int64(val))
	case float32:
		return e.writeFloat(float64(val))
	case float64:
		return e.writeFloat(val)
	case bool:
		return e.writeBool(val)
	case nil:
		return e.writeNull()
	case SimpleValue:
		return e.writeString(fmt.Sprintf("simple:%d", val))
	case []byte:
		return e.writeString(base64.StdEncoding.EncodeToString(val))
	case string:
		return e.writeString(val)
	case []Value:
		return e.encodeArray(val)
	case []interface{}:
		arr := make([]Value, len(val))
		for i, vv := range val {
			arr[i] = vv
		}
		return e.encodeArray(arr)
	case *Map:
		return e.encodeMap(val)
	case *BigInt:
		if val.Negative {
			n := new(big.Int).Neg(val.Int)
			return e.writeString(n.String())
		}
		return e.writeString(val.Int.String())
	default:
		return e.writeString(fmt.Sprintf("%v", v))
	}
}

func (e *jsonEncoder) writeInt(n int64) error {
	_, err := fmt.Fprintf(e.w, "%d", n)
	return err
}

func (e *jsonEncoder) writeFloat(f float64) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
}

func (e *jsonEncoder) writeBool(b bool) error {
	if b {
		_, err := e.w.Write([]byte("true"))
		return err
	}
	_, err := e.w.Write([]byte("false"))
	return err
}

func (e *jsonEncoder) writeNull() error {
	_, err := e.w.Write([]byte("null"))
	return err
}

func (e *jsonEncoder) writeString(s string) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
}

func (e *jsonEncoder) encodeArray(arr []Value) error {
	if _, err := e.w.Write([]byte("[")); err != nil {
		return err
	}
	for i, v := range arr {
		if i > 0 {
			if _, err := e.w.Write([]byte(",")); err != nil {
				return err
			}
		}
		if err := e.encode(v); err != nil {
			return err
		}
	}
	if _, err := e.w.Write([]byte("]")); err != nil {
		return err
	}
	return nil
}

func (e *jsonEncoder) encodeMap(m *Map) error {
	if _, err := e.w.Write([]byte("{")); err != nil {
		return err
	}
	for i, entry := range m.Entries {
		if i > 0 {
			if _, err := e.w.Write([]byte(",")); err != nil {
				return err
			}
		}
		keyStr := keyToString(entry.Key)
		if err := e.writeString(keyStr); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(":")); err != nil {
			return err
		}
		if err := e.encode(entry.Value); err != nil {
			return err
		}
	}
	if _, err := e.w.Write([]byte("}")); err != nil {
		return err
	}
	return nil
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
