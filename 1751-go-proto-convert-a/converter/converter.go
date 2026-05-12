package converter

import (
	"fmt"
	"strconv"
	"time"

	"protocol-converter/config"
	"protocol-converter/diagnostic"
)

func ConvertRequest(input map[string]interface{}, rules []config.FieldRule, path string) map[string]interface{} {
	return convert(input, rules, path, true)
}

func ConvertResponse(input map[string]interface{}, rules []config.FieldRule, path string) map[string]interface{} {
	return convert(input, rules, path, false)
}

func convert(input map[string]interface{}, rules []config.FieldRule, path string, toInternal bool) map[string]interface{} {
	output := make(map[string]interface{})
	processedFields := make(map[string]bool)

	for _, rule := range rules {
		var sourceKey, targetKey string
		if toInternal {
			sourceKey = rule.External
			targetKey = rule.Internal
		} else {
			sourceKey = rule.Internal
			targetKey = rule.External
		}

		processedFields[sourceKey] = true

		value, exists := input[sourceKey]
		if !exists {
			if rule.Default != nil {
				output[targetKey] = rule.Default
			}
			continue
		}

		convertedValue := convertValue(value, rule, toInternal, path)
		output[targetKey] = convertedValue
	}

	for key := range input {
		if !processedFields[key] {
			diagnostic.LogUnknownField(path, key)
		}
	}

	return output
}

func convertValue(value interface{}, rule config.FieldRule, toInternal bool, path string) interface{} {
	if rule.IsArray {
		arr, ok := value.([]interface{})
		if !ok {
			return value
		}

		var processedArr []interface{}
		for _, item := range arr {
			if len(rule.NestedRules) > 0 {
				if itemMap, ok := item.(map[string]interface{}); ok {
					processedArr = append(processedArr, convert(itemMap, rule.NestedRules, path, toInternal))
				} else {
					processedArr = append(processedArr, item)
				}
			} else {
				processedArr = append(processedArr, convertSingleValue(item, rule, toInternal))
			}
		}

		processedArr = filterArray(processedArr, rule)
		processedArr = sortArray(processedArr, rule)

		return processedArr
	}

	if len(rule.NestedRules) > 0 {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			return convert(nestedMap, rule.NestedRules, path, toInternal)
		}
	}

	return convertSingleValue(value, rule, toInternal)
}

func convertSingleValue(value interface{}, rule config.FieldRule, toInternal bool) interface{} {
	if rule.IsDate {
		return convertDate(value, toInternal)
	}

	if rule.Type != "" {
		return convertType(value, rule.Type)
	}

	return value
}

func convertDate(value interface{}, toInternal bool) interface{} {
	if toInternal {
		if str, ok := value.(string); ok {
			t, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return value
			}
			return t.Unix()
		}
	} else {
		if num, ok := value.(float64); ok {
			t := time.Unix(int64(num), 0)
			return t.Format(time.RFC3339)
		}
		if num, ok := value.(int64); ok {
			t := time.Unix(num, 0)
			return t.Format(time.RFC3339)
		}
	}
	return value
}

func convertType(value interface{}, targetType string) interface{} {
	switch targetType {
	case "string":
		return fmt.Sprintf("%v", value)
	case "int":
		switch v := value.(type) {
		case float64:
			return int64(v)
		case string:
			if num, err := strconv.ParseInt(v, 10, 64); err == nil {
				return num
			}
		}
	case "float":
		switch v := value.(type) {
		case int64:
			return float64(v)
		case int:
			return float64(v)
		case string:
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				return num
			}
		}
	case "bool":
		switch v := value.(type) {
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		}
	}
	return value
}
