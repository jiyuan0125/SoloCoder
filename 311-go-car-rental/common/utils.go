package common

import "time"

const DateFormat = "2006-01-02"

func ParseDate(dateStr string) (time.Time, error) {
	return time.ParseInLocation(DateFormat, dateStr, time.Local)
}

func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}
