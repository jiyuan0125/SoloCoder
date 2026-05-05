package logparser

import (
	"strconv"
	"time"
)

var timeFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"02/Jan/2006:15:04:05 -0700",
	"02/Jan/2006:15:04:05 +0700",
	"02/Jan/2006:15:04:05",
	"Jan 02 15:04:05",
	"Jan  2 15:04:05",
	"2006/01/02 15:04:05",
	"2006-01-02",
	"02-01-2006",
	"01/02/2006",
	time.UnixDate,
	time.RubyDate,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.ANSIC,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
}

func parseTimestamp(timeStr string) any {
	if timeStr == "" {
		return nil
	}

	if ts, ok := tryParseUnixTimestamp(timeStr); ok {
		return ts
	}

	for _, format := range timeFormats {
		if t, err := time.ParseInLocation(format, timeStr, time.Local); err == nil {
			return t
		}
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	if len(timeStr) >= 4 && len(timeStr) <= 14 {
		if allDigits(timeStr) {
			if ts, ok := tryParseUnixTimestamp(timeStr); ok {
				return ts
			}
		}
	}

	return timeStr
}

func tryParseUnixTimestamp(timeStr string) (time.Time, bool) {
	if len(timeStr) == 0 {
		return time.Time{}, false
	}

	if ts, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		if ts > 1e18 {
			return time.Unix(0, ts), true
		} else if ts > 1e15 {
			return time.Unix(0, ts*1e3), true
		} else if ts > 1e12 {
			return time.Unix(ts/1e3, (ts%1e3)*1e6), true
		} else {
			return time.Unix(ts, 0), true
		}
	}

	if ts, err := strconv.ParseFloat(timeStr, 64); err == nil {
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), true
	}

	return time.Time{}, false
}

func allDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
