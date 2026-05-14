package processor

import (
	"fmt"
	"time"

	"reportgen/internal/model"
)

type groupKey string

func ApplyGroupBy(rows []model.DataRow, groupBys []model.GroupBy) (map[groupKey][]model.DataRow, []string, error) {
	if len(groupBys) == 0 {
		return nil, nil, nil
	}

	validGroupBys := make([]model.GroupBy, 0, len(groupBys))
	groupKeys := make([]string, 0, len(groupBys))

	for _, gb := range groupBys {
		if len(rows) > 0 {
			if _, exists := rows[0].Values[gb.Field]; !exists {
				fmt.Printf("警告: 分组字段 '%s' 不存在，跳过该分组\n", gb.Field)
				continue
			}
		}
		validGroupBys = append(validGroupBys, gb)
		if gb.DateGroup != "" {
			groupKeys = append(groupKeys, fmt.Sprintf("%s_%s", gb.Field, gb.DateGroup))
		} else {
			groupKeys = append(groupKeys, gb.Field)
		}
	}

	if len(validGroupBys) == 0 {
		return nil, nil, nil
	}

	groups := make(map[groupKey][]model.DataRow)
	for _, row := range rows {
		key := generateGroupKey(row, validGroupBys)
		groups[key] = append(groups[key], row)
	}

	return groups, groupKeys, nil
}

func generateGroupKey(row model.DataRow, groupBys []model.GroupBy) groupKey {
	keyParts := make([]string, len(groupBys))
	for i, gb := range groupBys {
		val, exists := row.Values[gb.Field]
		if !exists {
			keyParts[i] = ""
			continue
		}

		if gb.DateGroup != "" {
			if t, ok := val.(time.Time); ok {
				keyParts[i] = formatDateByGroup(t, gb.DateGroup)
			} else {
				keyParts[i] = fmt.Sprintf("%v", val)
			}
		} else {
			keyParts[i] = fmt.Sprintf("%v", val)
		}
	}

	return groupKey(joinWithSeparator(keyParts, "|||"))
}

func formatDateByGroup(t time.Time, groupType model.DateGroupType) string {
	switch groupType {
	case model.DateGroupYear:
		return fmt.Sprintf("%d", t.Year())
	case model.DateGroupMonth:
		return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
	case model.DateGroupWeek:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case model.DateGroupDay:
		return t.Format("2006-01-02")
	default:
		return t.Format("2006-01-02")
	}
}

func joinWithSeparator(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

func CalculateAggregations(groups map[groupKey][]model.DataRow, aggregations []model.Aggregation, groupKeys []string) ([]model.GroupData, error) {
	if len(aggregations) == 0 {
		return nil, nil
	}

	var results []model.GroupData

	for key, rows := range groups {
		groupData := model.GroupData{
			Key:    parseGroupKey(key, groupKeys),
			Values: make(map[string]interface{}),
			Count:  len(rows),
		}

		for _, agg := range aggregations {
			alias := agg.Alias
			if alias == "" {
				if agg.Type == model.AggCount {
					alias = "count"
				} else {
					alias = fmt.Sprintf("%s_%s", agg.Field, agg.Type)
				}
			}

			value, err := calculateAggregation(rows, agg)
			if err != nil {
				return nil, err
			}
			groupData.Values[alias] = value
		}

		results = append(results, groupData)
	}

	return results, nil
}

func calculateAggregation(rows []model.DataRow, agg model.Aggregation) (interface{}, error) {
	if agg.Type == model.AggCount {
		return len(rows), nil
	}

	if len(rows) == 0 {
		return nil, nil
	}

	var values []float64
	for _, row := range rows {
		val, exists := row.Values[agg.Field]
		if !exists {
			continue
		}
		if num, ok := toFloat64(val); ok {
			values = append(values, num)
		}
	}

	if len(values) == 0 {
		return nil, nil
	}

	switch agg.Type {
	case model.AggSum:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum, nil
	case model.AggAvg:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values)), nil
	case model.AggMin:
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min, nil
	case model.AggMax:
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max, nil
	default:
		return nil, fmt.Errorf("unsupported aggregation type: %s", agg.Type)
	}
}

func parseGroupKey(key groupKey, groupKeys []string) model.GroupKey {
	parts := splitBySeparator(string(key), "|||")
	result := model.GroupKey{Fields: make(map[string]interface{})}

	for i, gk := range groupKeys {
		if i < len(parts) {
			result.Fields[gk] = parts[i]
		}
	}

	return result
}

func splitBySeparator(str string, sep string) []string {
	if str == "" {
		return []string{}
	}

	runes := []rune(str)
	sepRunes := []rune(sep)
	sepLen := len(sepRunes)

	var parts []string
	var current []rune

	for i := 0; i < len(runes); {
		if i+sepLen <= len(runes) && string(runes[i:i+sepLen]) == sep {
			parts = append(parts, string(current))
			current = nil
			i += sepLen
		} else {
			current = append(current, runes[i])
			i++
		}
	}
	parts = append(parts, string(current))

	return parts
}
