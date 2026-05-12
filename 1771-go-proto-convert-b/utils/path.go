package utils

import (
	"strconv"
	"strings"
)

func GetByPath(data map[string]interface{}, path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}
	
	parts := splitPath(path)
	var current interface{} = data
	
	for _, part := range parts {
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		
		val, exists := currentMap[part]
		if !exists {
			return nil, false
		}
		current = val
	}
	
	return current, true
}

func SetByPath(data map[string]interface{}, path string, value interface{}) bool {
	if path == "" {
		return false
	}
	
	parts := splitPath(path)
	current := data
	
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, exists := current[part]
		if !exists {
			current[part] = make(map[string]interface{})
			current = current[part].(map[string]interface{})
		} else {
			nextMap, ok := next.(map[string]interface{})
			if !ok {
				return false
			}
			current = nextMap
		}
	}
	
	current[parts[len(parts)-1]] = value
	return true
}

func splitPath(path string) []string {
	path = strings.TrimPrefix(path, ".")
	path = strings.TrimSuffix(path, ".")
	return strings.Split(path, ".")
}

func ConvertType(value interface{}, sourceType, targetType string) (interface{}, error) {
	switch targetType {
	case "string":
		return convertToString(value)
	case "int", "integer":
		return convertToInt(value)
	case "float", "number":
		return convertToFloat(value)
	case "bool", "boolean":
		return convertToBool(value)
	case "array":
		return convertToArray(value)
	case "object", "map":
		return convertToMap(value)
	default:
		return value, nil
	}
}

func convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", &ConvertError{Type: "string"}
	}
}

func convertToInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, &ConvertError{Type: "int"}
	}
}

func convertToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, &ConvertError{Type: "float"}
	}
}

func convertToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, &ConvertError{Type: "bool"}
	}
}

func convertToArray(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil
	default:
		return nil, &ConvertError{Type: "array"}
	}
}

func convertToMap(value interface{}) (map[string]interface{}, error) {
	v, ok := value.(map[string]interface{})
	if !ok {
		return nil, &ConvertError{Type: "object"}
	}
	return v, nil
}

type ConvertError struct {
	Type string
}

func (e *ConvertError) Error() string {
	return "failed to convert to " + e.Type
}
