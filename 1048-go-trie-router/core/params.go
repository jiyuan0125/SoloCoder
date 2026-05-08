package core

import (
	"strconv"
	"strings"
)

func validateParamType(value string, paramType ParamType) bool {
	switch paramType {
	case ParamTypeString:
		return true
	case ParamTypeInt:
		_, err := strconv.ParseInt(value, 10, 64)
		return err == nil
	case ParamTypeBool:
		_, err := strconv.ParseBool(strings.ToLower(value))
		return err == nil
	default:
		return true
	}
}

func copyMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
