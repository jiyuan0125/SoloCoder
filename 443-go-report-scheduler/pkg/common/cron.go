package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CronField struct {
	Name     string
	MinValue int
	MaxValue int
}

var CronFields = []CronField{
	{Name: "秒", MinValue: 0, MaxValue: 59},
	{Name: "分", MinValue: 0, MaxValue: 59},
	{Name: "时", MinValue: 0, MaxValue: 23},
	{Name: "日", MinValue: 1, MaxValue: 31},
	{Name: "月", MinValue: 1, MaxValue: 12},
	{Name: "周", MinValue: 0, MaxValue: 7},
}

func ValidateCronExpression(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) != 6 {
		return fmt.Errorf("cron表达式必须包含6个字段，实际有%d个", len(fields))
	}

	for i, field := range fields {
		cf := CronFields[i]
		if err := validateCronField(field, cf); err != nil {
			return fmt.Errorf("%s字段验证失败: %w", cf.Name, err)
		}
	}

	return nil
}

func validateCronField(field string, cf CronField) error {
	if field == "*" {
		return nil
	}

	if strings.Contains(field, ",") {
		values := strings.Split(field, ",")
		for _, v := range values {
			if err := validateSingleValue(v, cf); err != nil {
				return err
			}
		}
		return nil
	}

	if strings.Contains(field, "-") {
		rangeParts := strings.Split(field, "-")
		if len(rangeParts) != 2 {
			return fmt.Errorf("范围格式错误: %s", field)
		}
		start, err := parseValue(rangeParts[0], cf)
		if err != nil {
			return err
		}
		end, err := parseValue(rangeParts[1], cf)
		if err != nil {
			return err
		}
		if start > end {
			return fmt.Errorf("范围起始值不能大于结束值: %s", field)
		}
		if start < cf.MinValue || end > cf.MaxValue {
			return fmt.Errorf("范围超出允许范围 [%d-%d]: %s", cf.MinValue, cf.MaxValue, field)
		}
		return nil
	}

	if strings.Contains(field, "/") {
		stepParts := strings.Split(field, "/")
		if len(stepParts) != 2 {
			return fmt.Errorf("步长格式错误: %s", field)
		}
		base := stepParts[0]
		stepStr := stepParts[1]
		
		if base != "*" {
			if err := validateSingleValue(base, cf); err != nil {
				return err
			}
		}
		
		step, err := strconv.Atoi(stepStr)
		if err != nil {
			return fmt.Errorf("步长值无效: %s", stepStr)
		}
		if step <= 0 {
			return fmt.Errorf("步长必须大于0: %d", step)
		}
		return nil
	}

	return validateSingleValue(field, cf)
}

func validateSingleValue(value string, cf CronField) error {
	val, err := parseValue(value, cf)
	if err != nil {
		return err
	}

	if cf.Name == "周" {
		if val == 7 {
			val = 0
		}
	}

	if val < cf.MinValue || val > cf.MaxValue {
		if cf.Name == "周" && val == 7 {
			return nil
		}
		return fmt.Errorf("值 %d 超出允许范围 [%d-%d]", val, cf.MinValue, cf.MaxValue)
	}

	return nil
}

func parseValue(value string, cf CronField) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("空值")
	}

	val, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("无效的数值: %s", trimmed)
	}

	return val, nil
}

func GetNextRunTime(cronExpr string, fromTime time.Time) (time.Time, error) {
	if err := ValidateCronExpression(cronExpr); err != nil {
		return time.Time{}, err
	}

	fields := strings.Fields(cronExpr)
	
	now := fromTime.Add(time.Second)
	
	for i := 0; i < 365*24*60*60; i++ {
		candidate := now.Add(time.Duration(i) * time.Second)
		
		if matchesCronField(fields[0], candidate.Second(), CronFields[0]) &&
			matchesCronField(fields[1], candidate.Minute(), CronFields[1]) &&
			matchesCronField(fields[2], candidate.Hour(), CronFields[2]) &&
			matchesDayOfMonth(fields[3], candidate.Day(), candidate.Month(), candidate.Year()) &&
			matchesCronField(fields[4], int(candidate.Month()), CronFields[4]) &&
			matchesDayOfWeek(fields[5], int(candidate.Weekday())) {
			return candidate, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法计算下次执行时间")
}

func matchesCronField(field string, value int, cf CronField) bool {
	if field == "*" {
		return true
	}

	if cf.Name == "周" && value == 0 {
		if matchesCronField(field, 7, CronField{Name: "周", MinValue: 0, MaxValue: 6}) {
			return true
		}
	}

	if strings.Contains(field, ",") {
		values := strings.Split(field, ",")
		for _, v := range values {
			if matchesSingleValue(v, value, cf) {
				return true
			}
		}
		return false
	}

	if strings.Contains(field, "-") {
		rangeParts := strings.Split(field, "-")
		start, _ := parseValue(rangeParts[0], cf)
		end, _ := parseValue(rangeParts[1], cf)
		return value >= start && value <= end
	}

	if strings.Contains(field, "/") {
		stepParts := strings.Split(field, "/")
		base := stepParts[0]
		stepStr := stepParts[1]
		step, _ := strconv.Atoi(stepStr)

		var start int
		if base == "*" {
			start = cf.MinValue
		} else {
			start, _ = parseValue(base, cf)
		}

		if value < start {
			return false
		}
		return (value-start)%step == 0
	}

	return matchesSingleValue(field, value, cf)
}

func matchesSingleValue(field string, value int, cf CronField) bool {
	val, err := parseValue(field, cf)
	if err != nil {
		return false
	}

	if cf.Name == "周" {
		if val == 7 {
			val = 0
		}
		if value == 7 {
			value = 0
		}
	}

	return val == value
}

func matchesDayOfMonth(field string, day int, month time.Month, year int) bool {
	cf := CronFields[3]
	
	maxDay := daysInMonth(month, year)
	if day > maxDay {
		return false
	}

	return matchesCronField(field, day, cf)
}

func matchesDayOfWeek(field string, dayOfWeek int) bool {
	cf := CronFields[5]
	
	if dayOfWeek == 0 {
		if matchesCronField(field, 7, cf) {
			return true
		}
	}
	
	return matchesCronField(field, dayOfWeek, cf)
}

func daysInMonth(month time.Month, year int) int {
	switch month {
	case time.February:
		if isLeapYear(year) {
			return 29
		}
		return 28
	case time.April, time.June, time.September, time.November:
		return 30
	default:
		return 31
	}
}

func isLeapYear(year int) bool {
	if year%4 != 0 {
		return false
	}
	if year%100 != 0 {
		return true
	}
	if year%400 != 0 {
		return false
	}
	return true
}
