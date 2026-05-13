package mapping

import (
	"strings"
)

func splitPath(path string) []string {
	return strings.Split(path, ".")
}

func getByPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := splitPath(path)
	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			val, ok := current[part]
			return val, ok
		}
		next, ok := current[part]
		if !ok {
			return nil, false
		}
		switch v := next.(type) {
		case map[string]interface{}:
			current = v
		default:
			return nil, false
		}
	}
	return nil, false
}

func setByPath(data map[string]interface{}, path string, value interface{}) {
	parts := splitPath(path)
	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		if next, ok := current[part]; ok {
			if m, ok := next.(map[string]interface{}); ok {
				current = m
			} else {
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			}
		} else {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}
}

func removeByPath(data map[string]interface{}, path string) {
	parts := splitPath(path)
	current := data
	var parents []map[string]interface{}
	var parentKeys []string
	
	for i, part := range parts {
		if i == len(parts)-1 {
			delete(current, part)
			for j := len(parents) - 1; j >= 0; j-- {
				parent := parents[j]
				key := parentKeys[j]
				if next, ok := parent[key]; ok {
					if m, ok := next.(map[string]interface{}); ok && len(m) == 0 {
						delete(parent, key)
					}
				}
			}
			return
		}
		next, ok := current[part]
		if !ok {
			return
		}
		if m, ok := next.(map[string]interface{}); ok {
			parents = append(parents, current)
			parentKeys = append(parentKeys, part)
			current = m
		} else {
			return
		}
	}
}

func ApplyXML2JSONMappings(data map[string]interface{}, mappings []FieldMapping) map[string]interface{} {
	if len(mappings) == 0 {
		return data
	}
	
	result := make(map[string]interface{})
	copyMap(result, data)
	
	for _, mapping := range mappings {
		if val, ok := getByPath(result, mapping.XMLPath); ok {
			setByPath(result, mapping.JSONPath, val)
			removeByPath(result, mapping.XMLPath)
		}
	}
	
	return result
}

func ApplyJSON2XMLMappings(data map[string]interface{}, mappings []FieldMapping) map[string]interface{} {
	if len(mappings) == 0 {
		return data
	}
	
	result := make(map[string]interface{})
	copyMap(result, data)
	
	for _, mapping := range mappings {
		if val, ok := getByPath(result, mapping.JSONPath); ok {
			setByPath(result, mapping.XMLPath, val)
			removeByPath(result, mapping.JSONPath)
		}
	}
	
	return result
}

func copyMap(dst, src map[string]interface{}) {
	for k, v := range src {
		switch val := v.(type) {
		case map[string]interface{}:
			newMap := make(map[string]interface{})
			copyMap(newMap, val)
			dst[k] = newMap
		case []interface{}:
			newArr := make([]interface{}, len(val))
			for i, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					newMap := make(map[string]interface{})
					copyMap(newMap, m)
					newArr[i] = newMap
				} else {
					newArr[i] = item
				}
			}
			dst[k] = newArr
		default:
			dst[k] = val
		}
	}
}
