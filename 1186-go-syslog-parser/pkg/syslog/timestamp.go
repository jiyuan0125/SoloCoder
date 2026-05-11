package syslog

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	rfc3339Regex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(\.\d{1,6})?(Z|[+-]\d{2}:?\d{2})$`)
	rfc3164Regex = regexp.MustCompile(`^([A-Za-z]{3})\s+(\d{1,2})\s+(\d{2}):(\d{2}):(\d{2})$`)
)

func ParseTimestamp(ts string) (time.Time, string, error) {
	ts = strings.TrimSpace(ts)

	if ts == "-" {
		return time.Time{}, "nil", nil
	}

	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999999Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t, "rfc3339", nil
		}
	}

	normalizedTS := normalizeRFC3339Timestamp(ts)
	for _, format := range formats {
		if t, err := time.Parse(format, normalizedTS); err == nil {
			return t, "rfc3339", nil
		}
	}

	if rfc3164Regex.MatchString(ts) {
		t, err := parseRFC3164Timestamp(ts)
		if err == nil {
			return t, "rfc3164", nil
		}
	}

	return time.Time{}, "", fmt.Errorf("unable to parse timestamp: %s", ts)
}

func normalizeRFC3339Timestamp(ts string) string {
	if len(ts) < 6 {
		return ts
	}

	lastIdx := len(ts) - 1
	if ts[lastIdx] == 'Z' {
		return ts
	}

	for i := len(ts) - 5; i >= 0 && i >= len(ts)-10; i-- {
		if ts[i] == '+' || ts[i] == '-' {
			offsetPart := ts[i:]
			if len(offsetPart) == 5 && offsetPart[3] != ':' {
				return ts[:i] + offsetPart[:3] + ":" + offsetPart[3:]
			}
			break
		}
	}

	return ts
}

func parseRFC3164Timestamp(ts string) (time.Time, error) {
	monthMap := map[string]time.Month{
		"Jan": time.January, "Feb": time.February, "Mar": time.March,
		"Apr": time.April, "May": time.May, "Jun": time.June,
		"Jul": time.July, "Aug": time.August, "Sep": time.September,
		"Oct": time.October, "Nov": time.November, "Dec": time.December,
	}

	matches := rfc3164Regex.FindStringSubmatch(ts)
	if len(matches) != 6 {
		return time.Time{}, fmt.Errorf("invalid RFC 3164 timestamp format")
	}

	monthStr := strings.Title(strings.ToLower(matches[1]))
	dayStr := fmt.Sprintf("%02s", matches[2])
	hour := matches[3]
	minute := matches[4]
	second := matches[5]

	month, ok := monthMap[monthStr]
	if !ok {
		return time.Time{}, fmt.Errorf("invalid month: %s", monthStr)
	}

	now := time.Now()
	year := now.Year()

	parsedTime, err := time.Parse("2006 01 02 15:04:05", fmt.Sprintf("%d %02d %s %s:%s:%s", year, month, dayStr, hour, minute, second))
	if err != nil {
		return time.Time{}, err
	}

	if parsedTime.After(now.Add(24 * time.Hour)) {
		parsedTime = parsedTime.AddDate(-1, 0, 0)
	}

	return parsedTime, nil
}

func FormatTimestamp(t time.Time, formatType string) string {
	if t.IsZero() {
		return "-"
	}

	if formatType == "rfc3164" {
		return t.Format("Jan _2 15:04:05")
	}

	return t.Format("2006-01-02T15:04:05.000000Z07:00")
}
