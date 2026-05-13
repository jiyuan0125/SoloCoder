package processor

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"reportgen/internal/model"
)

func ApplyFilters(rows []model.DataRow, filters []model.Filter) []model.DataRow {
	if len(filters) == 0 {
		return rows
	}

	var result []model.DataRow
	for _, row := range rows {
		if matchAllFilters(row, filters) {
			result = append(result, row)
		}
	}
	return result
}

func matchAllFilters(row model.DataRow, filters []model.Filter) bool {
	for _, f := range filters {
		if !matchFilter(row, f) {
			return false
		}
	}
	return true
}

func matchFilter(row model.DataRow, f model.Filter) bool {
	val, exists := row.Values[f.Field]
	if !exists {
		return false
	}

	return compare(val, f.Operator, f.Value)
}

func compare(a interface{}, op model.FilterOperator, b interface{}) bool {
	switch op {
	case model.FilterEq:
		return equals(a, b)
	case model.FilterNe:
		return !equals(a, b)
	case model.FilterGt:
		return compareNumbers(a, b, func(x, y float64) bool { return x > y })
	case model.FilterGte:
		return compareNumbers(a, b, func(x, y float64) bool { return x >= y })
	case model.FilterLt:
		return compareNumbers(a, b, func(x, y float64) bool { return x < y })
	case model.FilterLte:
		return compareNumbers(a, b, func(x, y float64) bool { return x <= y })
	case model.FilterLike:
		return like(a, b)
	case model.FilterIn:
		return in(a, b)
	case model.FilterBetween:
		return between(a, b)
	default:
		return false
	}
}

func equals(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if t, ok := a.(time.Time); ok {
		if bt, ok := toTime(b); ok {
			return t.Equal(bt)
		}
	}

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Type() != vb.Type() {
		if numA, ok := toFloat64(a); ok {
			if numB, ok := toFloat64(b); ok {
				return numA == numB
			}
		}
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}

	return a == b
}

func compareNumbers(a, b interface{}, cmp func(float64, float64) bool) bool {
	numA, okA := toFloat64(a)
	numB, okB := toFloat64(b)
	if !okA || !okB {
		return false
	}
	return cmp(numA, numB)
}

func like(a, b interface{}) bool {
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)

	strB = strings.ReplaceAll(strB, "%", ".*")
	strB = strings.ReplaceAll(strB, "_", ".")

	return strings.Contains(strA, strB) || strings.HasPrefix(strA, strB) || strings.HasSuffix(strA, strB)
}

func in(a, b interface{}) bool {
	vb := reflect.ValueOf(b)
	if vb.Kind() != reflect.Slice && vb.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < vb.Len(); i++ {
		if equals(a, vb.Index(i).Interface()) {
			return true
		}
	}
	return false
}

func between(a, b interface{}) bool {
	vb := reflect.ValueOf(b)
	if vb.Kind() != reflect.Slice && vb.Kind() != reflect.Array || vb.Len() < 2 {
		return false
	}

	min := vb.Index(0).Interface()
	max := vb.Index(1).Interface()

	return compareNumbers(a, min, func(x, y float64) bool { return x >= y }) &&
		compareNumbers(a, max, func(x, y float64) bool { return x <= y })
}

func toFloat64(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case int16:
		return float64(x), true
	case int8:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint64:
		return float64(x), true
	case uint32:
		return float64(x), true
	case time.Time:
		return float64(x.Unix()), true
	default:
		return 0, false
	}
}

func toTime(v interface{}) (time.Time, bool) {
	switch x := v.(type) {
	case time.Time:
		return x, true
	case string:
		layouts := []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006/01/02",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, x); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}
