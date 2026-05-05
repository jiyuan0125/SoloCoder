package cron

import (
	"fmt"
	"strconv"
	"strings"
)

func Parse(expr string) (*CronExpr, error) {
	parts := strings.Fields(expr)
	if len(parts) != 6 {
		return nil, &ParseError{
			Message: fmt.Sprintf("expected 6 fields, got %d", len(parts)),
			Pos:     0,
			Field:   "all",
		}
	}

	fieldTypes := []FieldType{FieldSecond, FieldMinute, FieldHour, FieldDayOfMonth, FieldMonth, FieldDayOfWeek}
	cron := &CronExpr{}

	for i, part := range parts {
		fieldType := fieldTypes[i]
		field, err := parseField(part, fieldType)
		if err != nil {
			return nil, err
		}

		switch fieldType {
		case FieldSecond:
			cron.Second = field
		case FieldMinute:
			cron.Minute = field
		case FieldHour:
			cron.Hour = field
		case FieldDayOfMonth:
			cron.DayOfMonth = field
		case FieldMonth:
			cron.Month = field
		case FieldDayOfWeek:
			cron.DayOfWeek = field
		}
	}

	return cron, nil
}

func parseField(part string, fieldType FieldType) (*Field, error) {
	minVal := fieldRanges[fieldType][0]
	maxVal := fieldRanges[fieldType][1]

	if part == "*" {
		return createWildcardField(fieldType, minVal, maxVal), nil
	}

	var values []int
	parts := strings.Split(part, ",")

	for _, p := range parts {
		vals, err := parseSubPart(p, fieldType, minVal, maxVal)
		if err != nil {
			return nil, err
		}
		values = append(values, vals...)
	}

	values = deduplicateAndSort(values)

	return &Field{
		Type:  fieldType,
		Range: values,
	}, nil
}

func parseSubPart(part string, fieldType FieldType, minVal, maxVal int) ([]int, error) {
	slashParts := strings.SplitN(part, "/", 2)

	var step int
	var err error

	if len(slashParts) == 2 {
		stepStr := slashParts[1]
		step, err = strconv.Atoi(stepStr)
		if err != nil {
			return nil, &ParseError{
				Message: fmt.Sprintf("invalid step value: %s", stepStr),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}
		if step == 0 {
			return nil, &ParseError{
				Message: "step value cannot be zero",
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}
		if step < 0 {
			return nil, &ParseError{
				Message: fmt.Sprintf("step value cannot be negative: %d", step),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}
	} else {
		step = 1
	}

	rangePart := slashParts[0]

	if rangePart == "*" {
		return generateRange(minVal, maxVal, step), nil
	}

	dashParts := strings.SplitN(rangePart, "-", 2)

	if len(dashParts) == 2 {
		startStr := dashParts[0]
		endStr := dashParts[1]

		start, err := strconv.Atoi(startStr)
		if err != nil {
			return nil, &ParseError{
				Message: fmt.Sprintf("invalid range start: %s", startStr),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}

		end, err := strconv.Atoi(endStr)
		if err != nil {
			return nil, &ParseError{
				Message: fmt.Sprintf("invalid range end: %s", endStr),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}

		if fieldType == FieldDayOfWeek {
			if start == 7 {
				start = 0
			}
			if end == 7 {
				end = 0
			}
		}

		if start < minVal || start > maxVal {
			return nil, &ParseError{
				Message: fmt.Sprintf("range start %d out of bounds [%d, %d]", start, minVal, maxVal),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}

		if end < minVal || end > maxVal {
			return nil, &ParseError{
				Message: fmt.Sprintf("range end %d out of bounds [%d, %d]", end, minVal, maxVal),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}

		if start > end {
			return nil, &ParseError{
				Message: fmt.Sprintf("range start %d is greater than end %d", start, end),
				Pos:     0,
				Field:   fieldNames[fieldType],
			}
		}

		return generateRange(start, end, step), nil
	}

	val, err := strconv.Atoi(rangePart)
	if err != nil {
		return nil, &ParseError{
			Message: fmt.Sprintf("invalid value: %s", rangePart),
			Pos:     0,
			Field:   fieldNames[fieldType],
		}
	}

	if fieldType == FieldDayOfWeek {
		if val == 7 {
			val = 0
		}
	}

	if val < minVal || val > maxVal {
		return nil, &ParseError{
			Message: fmt.Sprintf("value %d out of bounds [%d, %d]", val, minVal, maxVal),
			Pos:     0,
			Field:   fieldNames[fieldType],
		}
	}

	return []int{val}, nil
}

func createWildcardField(fieldType FieldType, minVal, maxVal int) *Field {
	return &Field{
		Type:  fieldType,
		Range: generateRange(minVal, maxVal, 1),
	}
}

func generateRange(start, end, step int) []int {
	var result []int
	for i := start; i <= end; i += step {
		result = append(result, i)
	}
	return result
}

func deduplicateAndSort(values []int) []int {
	if len(values) == 0 {
		return values
	}

	seen := make(map[int]bool)
	var result []int

	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
