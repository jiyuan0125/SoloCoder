package processor

import (
	"reportgen/internal/model"
)

func DetectAnomalies(rows []model.DataRow, rules []model.AnomalyRule) []model.ProcessedRow {
	if len(rules) == 0 {
		result := make([]model.ProcessedRow, len(rows))
		for i, row := range rows {
			result[i] = model.ProcessedRow{
				Values: make(map[string]model.ProcessedValue),
			}
			for k, v := range row.Values {
				result[i].Values[k] = model.ProcessedValue{
					Value:     v,
					IsAnomaly: false,
				}
			}
		}
		return result
	}

	result := make([]model.ProcessedRow, len(rows))
	for i, row := range rows {
		result[i] = model.ProcessedRow{
			Values: make(map[string]model.ProcessedValue),
		}
		for k, v := range row.Values {
			result[i].Values[k] = model.ProcessedValue{
				Value:     v,
				IsAnomaly: checkAnomaly(row, k, v, rules),
			}
		}
	}

	return result
}

func checkAnomaly(row model.DataRow, field string, value interface{}, rules []model.AnomalyRule) bool {
	for _, rule := range rules {
		if rule.Field != field {
			continue
		}

		if isAnomaly(value, rule.Operator, rule.Threshold) {
			return true
		}
	}
	return false
}

func isAnomaly(value interface{}, operator string, threshold interface{}) bool {
	switch operator {
	case "<", "lt":
		return compareNumbers(value, threshold, func(x, y float64) bool { return x < y })
	case "<=", "lte":
		return compareNumbers(value, threshold, func(x, y float64) bool { return x <= y })
	case ">", "gt":
		return compareNumbers(value, threshold, func(x, y float64) bool { return x > y })
	case ">=", "gte":
		return compareNumbers(value, threshold, func(x, y float64) bool { return x >= y })
	case "==", "=", "eq":
		return equals(value, threshold)
	case "!=", "ne":
		return !equals(value, threshold)
	default:
		return false
	}
}
