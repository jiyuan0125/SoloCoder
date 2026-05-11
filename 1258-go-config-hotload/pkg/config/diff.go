package config

import (
	"reflect"
)

func computeDiff(oldConfig, newConfig map[string]interface{}, refMap map[string][]string) []Change {
	changes := make([]Change, 0)

	for k, oldVal := range oldConfig {
		if newVal, ok := newConfig[k]; ok {
			compareAndCollect("", k, oldVal, newVal, refMap, &changes)
		} else {
			collectDeletes("", k, oldVal, refMap, &changes)
		}
	}

	for k, newVal := range newConfig {
		if _, ok := oldConfig[k]; !ok {
			collectAdds("", k, newVal, refMap, &changes)
		}
	}

	return changes
}

func compareAndCollect(prefix, key string, oldVal, newVal interface{}, refMap map[string][]string, changes *[]Change) {
	var path string
	if prefix == "" {
		path = key
	} else {
		path = prefix + "." + key
	}

	oldMap, oldIsMap := oldVal.(map[string]interface{})
	newMap, newIsMap := newVal.(map[string]interface{})

	if oldIsMap && newIsMap {
		compareMaps(path, oldMap, newMap, refMap, changes)
		return
	}

	oldArr, oldIsArr := oldVal.([]interface{})
	newArr, newIsArr := newVal.([]interface{})

	if oldIsArr && newIsArr {
		if !reflect.DeepEqual(oldArr, newArr) {
			*changes = append(*changes, Change{
				Path:     path,
				Old:      oldVal,
				New:      newVal,
				Type:     ChangeTypeModify,
				Affected: getAffectedPaths(path, refMap),
			})
		}
		return
	}

	if !reflect.DeepEqual(oldVal, newVal) {
		*changes = append(*changes, Change{
			Path:     path,
			Old:      oldVal,
			New:      newVal,
			Type:     ChangeTypeModify,
			Affected: getAffectedPaths(path, refMap),
		})
	}
}

func compareMaps(prefix string, oldMap, newMap map[string]interface{}, refMap map[string][]string, changes *[]Change) {
	for k, oldVal := range oldMap {
		if newVal, ok := newMap[k]; ok {
			compareAndCollect(prefix, k, oldVal, newVal, refMap, changes)
		} else {
			collectDeletes(prefix, k, oldVal, refMap, changes)
		}
	}

	for k, newVal := range newMap {
		if _, ok := oldMap[k]; !ok {
			collectAdds(prefix, k, newVal, refMap, changes)
		}
	}
}

func collectDeletes(prefix, key string, val interface{}, refMap map[string][]string, changes *[]Change) {
	var path string
	if prefix == "" {
		path = key
	} else {
		path = prefix + "." + key
	}

	*changes = append(*changes, Change{
		Path:     path,
		Old:      val,
		New:      nil,
		Type:     ChangeTypeDelete,
		Affected: getAffectedPaths(path, refMap),
	})

	if subMap, ok := val.(map[string]interface{}); ok {
		for k, subVal := range subMap {
			collectDeletes(path, k, subVal, refMap, changes)
		}
	}
}

func collectAdds(prefix, key string, val interface{}, refMap map[string][]string, changes *[]Change) {
	var path string
	if prefix == "" {
		path = key
	} else {
		path = prefix + "." + key
	}

	*changes = append(*changes, Change{
		Path:     path,
		Old:      nil,
		New:      val,
		Type:     ChangeTypeAdd,
		Affected: getAffectedPaths(path, refMap),
	})

	if subMap, ok := val.(map[string]interface{}); ok {
		for k, subVal := range subMap {
			collectAdds(path, k, subVal, refMap, changes)
		}
	}
}

func getAffectedPaths(path string, refMap map[string][]string) []string {
	result := make([]string, 0)

	if affected, ok := refMap[path]; ok {
		result = append(result, affected...)
	}

	for refPath, affected := range refMap {
		if startsWithPrefix(refPath, path) {
			for _, a := range affected {
				if !contains(result, a) {
					result = append(result, a)
				}
			}
		}
	}

	return result
}

func startsWithPrefix(fullPath, prefix string) bool {
	if prefix == "" {
		return false
	}

	if len(fullPath) < len(prefix) {
		return false
	}

	if fullPath == prefix {
		return true
	}

	if len(fullPath) > len(prefix) && fullPath[len(prefix)] == '.' {
		return fullPath[:len(prefix)] == prefix
	}

	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func applyChanges(config map[string]interface{}, changes []Change) map[string]interface{} {
	result := deepCopy(config)

	for _, change := range changes {
		applyChange(result, change)
	}

	return result
}

func applyChange(config map[string]interface{}, change Change) {
	parts := splitPath(change.Path)
	if len(parts) == 0 {
		return
	}

	current := config
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	lastPart := parts[len(parts)-1]
	switch change.Type {
	case ChangeTypeAdd, ChangeTypeModify:
		current[lastPart] = change.New
	case ChangeTypeDelete:
		delete(current, lastPart)
	}
}

func splitPath(path string) []string {
	return splitByDotExcludingBrackets(path)
}

func splitByDotExcludingBrackets(path string) []string {
	var parts []string
	var current string
	bracketDepth := 0

	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '.':
			if bracketDepth == 0 {
				if current != "" {
					parts = append(parts, current)
				}
				current = ""
			} else {
				current += string(c)
			}
		case '[':
			bracketDepth++
			current += string(c)
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			current += string(c)
		default:
			current += string(c)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func deepCopy(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = copyValue(v)
	}
	return result
}

func copyValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return deepCopy(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = copyValue(item)
		}
		return result
	default:
		return val
	}
}
