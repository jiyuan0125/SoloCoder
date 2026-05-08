package replay

import (
	"encoding/json"
	"net/http"
	"strings"

	"http-replay/api"
)

func applyReplaceRules(reqURL string, rules []api.ReplaceRule) string {
	for _, rule := range rules {
		if rule.OldPrefix == "" {
			continue
		}
		if strings.HasPrefix(reqURL, rule.OldPrefix) {
			reqURL = strings.Replace(reqURL, rule.OldPrefix, rule.NewPrefix, 1)
		}
	}
	return reqURL
}

func applyVariableReplacements(body string, variableValues []api.VariableValue, recordedVars map[string]string) string {
	if len(variableValues) == 0 || len(recordedVars) == 0 {
		return body
	}

	replacements := make(map[string]string)
	for _, vv := range variableValues {
		if oldValue, exists := recordedVars[vv.Name]; exists {
			replacements[oldValue] = vv.Value
		}
	}

	result := body
	for oldVal, newVal := range replacements {
		result = strings.ReplaceAll(result, oldVal, newVal)
	}

	return result
}

func applyVariableReplacementsToHeaders(headers http.Header, variableValues []api.VariableValue, recordedVars map[string]string) {
	if len(variableValues) == 0 || len(recordedVars) == 0 {
		return
	}

	replacements := make(map[string]string)
	for _, vv := range variableValues {
		if oldValue, exists := recordedVars[vv.Name]; exists {
			replacements[oldValue] = vv.Value
		}
	}

	for key, values := range headers {
		for i, val := range values {
			for oldVal, newVal := range replacements {
				values[i] = strings.ReplaceAll(val, oldVal, newVal)
			}
		}
		headers[key] = values
	}
}

func ReplaceJSONPath(data []byte, path string, newValue string) ([]byte, bool) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 || parts[0] != "$" {
		return data, false
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, false
	}

	current := obj
	for i := 1; i < len(parts)-1; i++ {
		key := parts[i]
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return data, false
			}
		} else {
			return data, false
		}
	}

	if m, ok := current.(map[string]interface{}); ok {
		lastKey := parts[len(parts)-1]
		if _, exists := m[lastKey]; exists {
			m[lastKey] = newValue
			result, err := json.Marshal(obj)
			if err != nil {
				return data, false
			}
			return result, true
		}
	}

	return data, false
}

func GetJSONPathString(data []byte, path string) (string, bool) {
	return extractJSONPath(data, path)
}
