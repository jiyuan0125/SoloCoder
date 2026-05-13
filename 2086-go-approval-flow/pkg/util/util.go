package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func ToJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func FromJSON(data string, v interface{}) error {
	if data == "" || data == "null" {
		return nil
	}
	return json.Unmarshal([]byte(data), v)
}

func SliceToJSON[T any](s []T) string {
	if len(s) == 0 {
		return "[]"
	}
	data, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func JSONToSlice[T any](data string) ([]T, error) {
	if data == "" || data == "null" {
		return []T{}, nil
	}
	var result []T
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IntToBool(i int) bool {
	return i != 0
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func StrPtr(s string) *string {
	return &s
}

func PtrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func MapToJSON(m map[string]interface{}) string {
	if m == nil {
		return "{}"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func JSONToMap(data string) (map[string]interface{}, error) {
	if data == "" || data == "null" {
		return map[string]interface{}{}, nil
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}

func FormatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
