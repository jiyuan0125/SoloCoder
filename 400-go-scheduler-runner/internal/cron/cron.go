package cron

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidCronFormat = errors.New("invalid cron expression format")
	ErrInvalidFieldValue = errors.New("invalid field value")
)

type CronField struct {
	Minute     map[int]bool
	Hour       map[int]bool
	DayOfMonth map[int]bool
	Month      map[int]bool
	DayOfWeek  map[int]bool
}

func ParseCron(expr string) (*CronField, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, ErrInvalidCronFormat
	}

	minute, err := parseField(fields[0], 0, 59)
	if err != nil {
		return nil, err
	}

	hour, err := parseField(fields[1], 0, 23)
	if err != nil {
		return nil, err
	}

	dayOfMonth, err := parseField(fields[2], 1, 31)
	if err != nil {
		return nil, err
	}

	month, err := parseField(fields[3], 1, 12)
	if err != nil {
		return nil, err
	}

	dayOfWeek, err := parseField(fields[4], 0, 7)
	if err != nil {
		return nil, err
	}

	return &CronField{
		Minute:     minute,
		Hour:       hour,
		DayOfMonth: dayOfMonth,
		Month:      month,
		DayOfWeek:  dayOfWeek,
	}, nil
}

func parseField(field string, min, max int) (map[int]bool, error) {
	result := make(map[int]bool)

	if field == "*" {
		for i := min; i <= max; i++ {
			result[i] = true
		}
		return result, nil
	}

	if strings.Contains(field, "/") {
		parts := strings.SplitN(field, "/", 2)
		base := parts[0]
		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return nil, ErrInvalidFieldValue
		}

		start := min
		end := max

		if base != "*" {
			if strings.Contains(base, "-") {
				rangeParts := strings.SplitN(base, "-", 2)
				start, err = strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, ErrInvalidFieldValue
				}
				end, err = strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, ErrInvalidFieldValue
				}
			} else {
				start, err = strconv.Atoi(base)
				if err != nil {
					return nil, ErrInvalidFieldValue
				}
				end = max
			}
		}

		for i := start; i <= end; i += step {
			if i < min || i > max {
				continue
			}
			result[i] = true
		}
		return result, nil
	}

	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, part := range parts {
			if strings.Contains(part, "-") {
				rangeParts := strings.SplitN(part, "-", 2)
				start, err := strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, ErrInvalidFieldValue
				}
				end, err := strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, ErrInvalidFieldValue
				}
				for i := start; i <= end; i++ {
					if i < min || i > max {
						return nil, ErrInvalidFieldValue
					}
					result[i] = true
				}
			} else {
				val, err := strconv.Atoi(part)
				if err != nil {
					return nil, ErrInvalidFieldValue
				}
				if val < min || val > max {
					return nil, ErrInvalidFieldValue
				}
				result[val] = true
			}
		}
		return result, nil
	}

	if strings.Contains(field, "-") {
		parts := strings.SplitN(field, "-", 2)
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, ErrInvalidFieldValue
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, ErrInvalidFieldValue
		}
		for i := start; i <= end; i++ {
			if i < min || i > max {
				return nil, ErrInvalidFieldValue
			}
			result[i] = true
		}
		return result, nil
	}

	val, err := strconv.Atoi(field)
	if err != nil {
		return nil, ErrInvalidFieldValue
	}
	if val < min || val > max {
		return nil, ErrInvalidFieldValue
	}
	result[val] = true
	return result, nil
}

func (cf *CronField) Next(t time.Time) time.Time {
	t = t.Truncate(time.Minute).Add(time.Minute)

	for count := 0; count < 100000; count++ {
		if cf.Matches(t) {
			return t
		}
		t = t.Add(time.Minute)
	}

	return time.Time{}
}

func (cf *CronField) Matches(t time.Time) bool {
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	weekday := int(t.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	if !cf.Minute[minute] {
		return false
	}
	if !cf.Hour[hour] {
		return false
	}
	if !cf.Month[month] {
		return false
	}

	dayMatch := cf.DayOfMonth[day]
	weekMatch := cf.DayOfWeek[weekday] || (weekday == 7 && cf.DayOfWeek[0])

	if len(cf.DayOfMonth) == 31 && len(cf.DayOfWeek) <= 1 {
		return dayMatch
	} else if len(cf.DayOfWeek) == 8 && len(cf.DayOfMonth) < 31 {
		return dayMatch
	} else if len(cf.DayOfMonth) == 31 && len(cf.DayOfWeek) == 8 {
		return true
	} else {
		return dayMatch || weekMatch
	}
}
