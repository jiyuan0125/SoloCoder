package converter

import (
	"sort"

	"protocol-converter/config"
)

func filterArray(arr []interface{}, rule config.FieldRule) []interface{} {
	if rule.ArrayFilter == "" {
		return arr
	}

	var filtered []interface{}
	for _, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if _, exists := itemMap[rule.ArrayFilter]; exists {
				filtered = append(filtered, item)
			}
		} else {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func sortArray(arr []interface{}, rule config.FieldRule) []interface{} {
	if rule.ArraySortBy == "" {
		return arr
	}

	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	sort.Slice(sorted, func(i, j int) bool {
		itemI, okI := sorted[i].(map[string]interface{})
		itemJ, okJ := sorted[j].(map[string]interface{})

		if !okI || !okJ {
			return i < j
		}

		valI, existsI := itemI[rule.ArraySortBy]
		valJ, existsJ := itemJ[rule.ArraySortBy]

		if !existsI || !existsJ {
			return existsI
		}

		numI, okINum := toFloat64(valI)
		numJ, okJNum := toFloat64(valJ)

		if okINum && okJNum {
			if rule.ArraySortAsc {
				return numI < numJ
			}
			return numI > numJ
		}

		strI, okIStr := valI.(string)
		strJ, okJStr := valJ.(string)

		if okIStr && okJStr {
			if rule.ArraySortAsc {
				return strI < strJ
			}
			return strI > strJ
		}

		return i < j
	})

	return sorted
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}
