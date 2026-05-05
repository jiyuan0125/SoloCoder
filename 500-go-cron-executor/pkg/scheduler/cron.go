package scheduler

import (
	"time"
)

type CronParser struct {
}

func NewCronParser() *CronParser {
	return &CronParser{}
}

func (p *CronParser) NextTime(expr string, from time.Time) (time.Time, error) {
	return p.parseCron(expr, from)
}

func (p *CronParser) parseCron(expr string, from time.Time) (time.Time, error) {
	minute, hour, dom, month, dow, err := p.parseCronFields(expr)
	if err != nil {
		return time.Time{}, err
	}

	next := from.Truncate(time.Minute).Add(time.Minute)
	end := next.AddDate(5, 0, 0)

	for next.Before(end) {
		if p.matchField(next.Minute(), minute) &&
			p.matchField(next.Hour(), hour) &&
			p.matchField(next.Day(), dom) &&
			p.matchField(int(next.Month()), month) &&
			p.matchField(int(next.Weekday()), dow) {
			return next, nil
		}
		next = next.Add(time.Minute)
	}

	return time.Time{}, nil
}

func (p *CronParser) parseCronFields(expr string) (minute, hour, dom, month, dow map[int]bool, err error) {
	parts := splitCron(expr)
	if len(parts) != 5 {
		return nil, nil, nil, nil, nil, nil
	}

	minute, err = p.parseField(parts[0], 0, 59)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	hour, err = p.parseField(parts[1], 0, 23)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	dom, err = p.parseField(parts[2], 1, 31)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	month, err = p.parseField(parts[3], 1, 12)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	dow, err = p.parseField(parts[4], 0, 6)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return minute, hour, dom, month, dow, nil
}

func splitCron(expr string) []string {
	var parts []string
	var current []byte

	for i := 0; i < len(expr); i++ {
		if expr[i] == ' ' || expr[i] == '\t' {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = nil
			}
		} else {
			current = append(current, expr[i])
		}
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

func (p *CronParser) parseField(field string, min, max int) (map[int]bool, error) {
	result := make(map[int]bool)

	if field == "*" {
		for i := min; i <= max; i++ {
			result[i] = true
		}
		return result, nil
	}

	if len(field) >= 2 && field[0:2] == "*/" {
		step := parseInt(field[2:])
		if step <= 0 {
			step = 1
		}
		for i := min; i <= max; i += step {
			result[i] = true
		}
		return result, nil
	}

	parts := splitBy(field, ',')
	for _, part := range parts {
		if contains(part, '-') {
			rangeParts := splitBy(part, '-')
			if len(rangeParts) == 2 {
				start := parseInt(rangeParts[0])
				end := parseInt(rangeParts[1])
				if start < min {
					start = min
				}
				if end > max {
					end = max
				}
				for i := start; i <= end; i++ {
					result[i] = true
				}
			}
		} else {
			val := parseInt(part)
			if val >= min && val <= max {
				result[val] = true
			}
		}
	}

	return result, nil
}

func splitBy(s string, sep byte) []string {
	var parts []string
	var current []byte

	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = nil
			}
		} else {
			current = append(current, s[i])
		}
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

func parseInt(s string) int {
	result := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + int(s[i]-'0')
		}
	}
	return result
}

func (p *CronParser) matchField(value int, field map[int]bool) bool {
	return field[value]
}
