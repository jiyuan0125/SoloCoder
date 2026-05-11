package bencode

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"unicode"
)

type JSONValue struct {
	ValueType string      `json:"type"`
	Value     interface{} `json:"value"`
}

func isPrintable(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func ToJSON(v Value) ([]byte, error) {
	converted := toJSONValue(v)
	return json.MarshalIndent(converted, "", "  ")
}

func toJSONValue(v Value) interface{} {
	switch val := v.(type) {
	case ByteString:
		s := string(val)
		if isPrintable(s) {
			return map[string]interface{}{
				"type":  "string",
				"value": s,
			}
		}
		return map[string]interface{}{
			"type":   "bytes",
			"base64": base64.StdEncoding.EncodeToString(val),
		}
	case Integer:
		return map[string]interface{}{
			"type":  "integer",
			"value": val.String(),
		}
	case List:
		list := make([]interface{}, len(val))
		for i, item := range val {
			list[i] = toJSONValue(item)
		}
		return map[string]interface{}{
			"type":  "list",
			"value": list,
		}
	case Dict:
		dict := make(map[string]interface{})
		for k, v := range val {
			dict[k] = toJSONValue(v)
		}
		return map[string]interface{}{
			"type":  "dict",
			"value": dict,
		}
	default:
		return map[string]interface{}{
			"type":  "unknown",
			"value": val,
		}
	}
}

func FromJSON(data []byte) (Value, error) {
	var raw map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	err := decoder.Decode(&raw)
	if err != nil {
		return nil, err
	}
	return fromJSONValue(raw)
}

func fromJSONValue(v interface{}) (Value, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		return fromJSONTaggedValue(val)
	case []interface{}:
		list := make(List, len(val))
		for i, item := range val {
			converted, err := fromJSONValue(item)
			if err != nil {
				return nil, err
			}
			list[i] = converted
		}
		return list, nil
	case string:
		return ByteString(val), nil
	case json.Number:
		n := string(val)
		if strings.Contains(n, ".") || strings.Contains(n, "e") || strings.Contains(n, "E") {
			return ByteString(n), nil
		}
		i := Integer{new(big.Int)}
		_, ok := i.SetString(n, 10)
		if ok {
			return i, nil
		}
		return ByteString(n), nil
	case float64:
		if float64(int64(val)) == val {
			return NewInteger(int64(val)), nil
		}
		return ByteString(formatFloat(val)), nil
	case bool:
		if val {
			return ByteString("true"), nil
		}
		return ByteString("false"), nil
	case nil:
		return ByteString(""), nil
	default:
		return ByteString(""), nil
	}
}

func fromJSONTaggedValue(m map[string]interface{}) (Value, error) {
	typeVal, hasType := m["type"]
	if !hasType {
		dict := make(Dict)
		for k, v := range m {
			converted, err := fromJSONValue(v)
			if err != nil {
				return nil, err
			}
			dict[k] = converted
		}
		return dict, nil
	}

	typeStr, ok := typeVal.(string)
	if !ok {
		dict := make(Dict)
		for k, v := range m {
			converted, err := fromJSONValue(v)
			if err != nil {
				return nil, err
			}
			dict[k] = converted
		}
		return dict, nil
	}

	valueRaw, hasValue := m["value"]
	base64Raw, hasBase64 := m["base64"]

	switch typeStr {
	case "string":
		if hasValue {
			if s, ok := valueRaw.(string); ok {
				return ByteString(s), nil
			}
		}
		return ByteString(""), nil
	case "bytes":
		if hasBase64 {
			if s, ok := base64Raw.(string); ok {
				decoded, err := base64.StdEncoding.DecodeString(s)
				if err != nil {
					return nil, err
				}
				return ByteString(decoded), nil
			}
		}
		return ByteString(""), nil
	case "integer":
		if hasValue {
			switch v := valueRaw.(type) {
			case string:
				i := Integer{new(big.Int)}
				_, ok := i.SetString(v, 10)
				if ok {
					return i, nil
				}
			case json.Number:
				i := Integer{new(big.Int)}
				_, ok := i.SetString(string(v), 10)
				if ok {
					return i, nil
				}
			case float64:
				return NewInteger(int64(v)), nil
			}
		}
		return NewInteger(0), nil
	case "list":
		var list List
		if hasValue {
			if arr, ok := valueRaw.([]interface{}); ok {
				list = make(List, len(arr))
				for i, item := range arr {
					converted, err := fromJSONValue(item)
					if err != nil {
						return nil, err
					}
					list[i] = converted
				}
			}
		}
		return list, nil
	case "dict":
		dict := make(Dict)
		if hasValue {
			if m2, ok := valueRaw.(map[string]interface{}); ok {
				for k, v := range m2 {
					converted, err := fromJSONValue(v)
					if err != nil {
						return nil, err
					}
					dict[k] = converted
				}
			}
		}
		return dict, nil
	default:
		dict := make(Dict)
		for k, v := range m {
			converted, err := fromJSONValue(v)
			if err != nil {
				return nil, err
			}
			dict[k] = converted
		}
		return dict, nil
	}
}
