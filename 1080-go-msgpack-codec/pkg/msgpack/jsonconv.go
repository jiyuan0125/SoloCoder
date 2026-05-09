package msgpack

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type jsonExt struct {
	ExtType int8   `json:"__ext_type__"`
	Data    string `json:"__ext_data__"`
}

type jsonBinary struct {
	Binary string `json:"__binary__"`
}

func ToJSONCompatible(v interface{}) interface{} {
	switch val := v.(type) {
	case nil, bool, string, float64:
		return val
	case int64:
		return val
	case []byte:
		return jsonBinary{Binary: base64.StdEncoding.EncodeToString(val)}
	case time.Time:
		return val.Format(time.RFC3339Nano)
	case Ext:
		return jsonExt{
			ExtType: val.Type,
			Data:    base64.StdEncoding.EncodeToString(val.Data),
		}
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = ToJSONCompatible(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, item := range val {
			result[k] = ToJSONCompatible(item)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, item := range val {
			keyStr := fmt.Sprintf("%v", k)
			result[keyStr] = ToJSONCompatible(item)
		}
		return result
	default:
		return v
	}
}

func FromJSONCompatible(v interface{}) interface{} {
	switch val := v.(type) {
	case nil, bool, string, float64, int64:
		return val
	case float32:
		return float64(val)
	case int, int8, int16, int32, uint, uint8, uint16, uint32, uint64:
		return int64(val.(int64))
	case map[string]interface{}:
		if extType, ok := val["__ext_type__"].(float64); ok {
			if data, ok := val["__ext_data__"].(string); ok {
				decoded, err := base64.StdEncoding.DecodeString(data)
				if err == nil {
					return Ext{
						Type: int8(extType),
						Data: decoded,
					}
				}
			}
		}
		if binData, ok := val["__binary__"].(string); ok && len(val) == 1 {
			decoded, err := base64.StdEncoding.DecodeString(binData)
			if err == nil {
				return decoded
			}
		}
		result := make(map[string]interface{})
		for k, item := range val {
			result[k] = FromJSONCompatible(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = FromJSONCompatible(item)
		}
		return result
	default:
		return v
	}
}

func ToJSON(v interface{}) ([]byte, error) {
	compatible := ToJSONCompatible(v)
	return json.Marshal(compatible)
}

func FromJSON(data []byte) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return FromJSONCompatible(v), nil
}
