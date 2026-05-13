package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templateRegex = regexp.MustCompile(`\$\{response\.([^}]+)\}`)

func ApplyTemplate(template interface{}, results []ServiceResult) interface{} {
	responseMap := make(map[string]interface{})
	for _, r := range results {
		responseMap[r.Name] = r.Response
	}

	return deepApply(template, responseMap)
}

func deepApply(value interface{}, responseMap map[string]interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return replacePlaceholders(v, responseMap)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = deepApply(val, responseMap)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = deepApply(val, responseMap)
		}
		return result
	default:
		return value
	}
}

func replacePlaceholders(s string, responseMap map[string]interface{}) interface{} {
	matches := templateRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return s
	}

	if len(matches) == 1 && matches[0][0] == s {
		return getValueByPath(responseMap, matches[0][1])
	}

	result := templateRegex.ReplaceAllStringFunc(s, func(match string) string {
		parts := templateRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		val := getValueByPath(responseMap, parts[1])
		return fmt.Sprintf("%v", val)
	})

	return result
}

func getValueByPath(responseMap map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}

	serviceName := parts[0]
	serviceResp, ok := responseMap[serviceName]
	if !ok {
		return nil
	}

	if len(parts) == 1 {
		return serviceResp
	}

	return navigatePath(serviceResp, parts[1:])
}

func navigatePath(data interface{}, pathParts []string) interface{} {
	if len(pathParts) == 0 {
		return data
	}

	current := data
	for _, part := range pathParts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil
			}
			current = val
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil
			}
			current = v[idx]
		default:
			return nil
		}
	}

	return current
}
