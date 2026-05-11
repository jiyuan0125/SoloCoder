package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type refResolver struct {
	refPattern *regexp.Regexp
}

func newRefResolver() *refResolver {
	return &refResolver{
		refPattern: regexp.MustCompile(`\$\{([^}]+)\}`),
	}
}

func (r *refResolver) resolveRefs(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = r.resolveValue(v, data)
	}
	return result
}

func (r *refResolver) resolveValue(v interface{}, root map[string]interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return r.resolveString(val, root)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, sub := range val {
			result[k] = r.resolveValue(sub, root)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = r.resolveValue(item, root)
		}
		return result
	default:
		return v
	}
}

func (r *refResolver) resolveString(s string, root map[string]interface{}) interface{} {
	matches := r.refPattern.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return s
	}

	if len(matches) == 1 && matches[0][0] == s {
		path := matches[0][1]
		return getValueByPath(root, path)
	}

	result := s
	for _, match := range matches {
		refVar := match[0]
		path := match[1]
		refVal := getValueByPath(root, path)
		if refVal != nil {
			result = strings.Replace(result, refVar, fmt.Sprintf("%v", refVal), 1)
		}
	}
	return result
}

func getValueByPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data
	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return nil
		}
		if i == len(parts)-1 {
			return val
		}
		sub, ok := val.(map[string]interface{})
		if !ok {
			return nil
		}
		current = sub
	}
	return nil
}

func findRefsInValue(v interface{}) []string {
	refs := make([]string, 0)
	switch val := v.(type) {
	case string:
		pattern := regexp.MustCompile(`\$\{([^}]+)\}`)
		matches := pattern.FindAllStringSubmatch(val, -1)
		for _, m := range matches {
			refs = append(refs, m[1])
		}
	case map[string]interface{}:
		for _, sub := range val {
			refs = append(refs, findRefsInValue(sub)...)
		}
	case []interface{}:
		for _, item := range val {
			refs = append(refs, findRefsInValue(item)...)
		}
	}
	return refs
}

func parseFile(filePath string, format ConfigFormat, enableRef bool) (map[string]interface{}, map[string][]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	config, err := parseData(data, format)
	if err != nil {
		return nil, nil, err
	}

	refMap := make(map[string][]string)
	if enableRef {
		refMap = buildRefMap(config)
		resolved := newRefResolver().resolveRefs(config)
		return resolved, refMap, nil
	}

	return config, refMap, nil
}

func parseData(data []byte, format ConfigFormat) (map[string]interface{}, error) {
	var result map[string]interface{}
	switch format {
	case FormatJSON:
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
	case FormatYAML:
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
	return normalizeMap(result), nil
}

func normalizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = normalizeValue(v)
	}
	return result
}

func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return normalizeMap(val)
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, vv := range val {
			result[fmt.Sprintf("%v", k)] = normalizeValue(vv)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = normalizeValue(item)
		}
		return result
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case bool:
		return val
	case string:
		return val
	case nil:
		return nil
	default:
		return val
	}
}

func buildRefMap(config map[string]interface{}) map[string][]string {
	refMap := make(map[string][]string)
	walkMap(config, "", func(path string, value interface{}) {
		refs := findRefsInValue(value)
		for _, ref := range refs {
			refMap[ref] = append(refMap[ref], path)
		}
	})

	for _, v := range refMap {
		sort.Strings(v)
	}
	return refMap
}

func walkMap(m map[string]interface{}, prefix string, fn func(path string, value interface{})) {
	for k, v := range m {
		var path string
		if prefix == "" {
			path = k
		} else {
			path = prefix + "." + k
		}
		fn(path, v)

		if sub, ok := v.(map[string]interface{}); ok {
			walkMap(sub, path, fn)
		} else if arr, ok := v.([]interface{}); ok {
			for i, item := range arr {
				if subItem, ok := item.(map[string]interface{}); ok {
					walkMap(subItem, fmt.Sprintf("%s[%d]", path, i), fn)
				}
			}
		}
	}
}
