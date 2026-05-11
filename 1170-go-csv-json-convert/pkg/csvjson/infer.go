package csvjson

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	numberPattern   = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	trimQuotePattern = regexp.MustCompile(`^"(.*)"$`)
)

func InferType(value string, maxDigits int) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.EqualFold(trimmed, "null") {
		return nil
	}
	if m := trimQuotePattern.FindStringSubmatch(trimmed); len(m) == 2 {
		inner := strings.ReplaceAll(m[1], "\"\"", "\"")
		return InferType(inner, maxDigits)
	}
	lower := strings.ToLower(trimmed)
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}
	if numberPattern.MatchString(trimmed) {
		abs := strings.TrimLeft(trimmed, "-")
		if idx := strings.Index(abs, "."); idx >= 0 {
			abs = abs[:idx]
		}
		digits := len(abs)
		if digits <= maxDigits {
			if strings.Contains(trimmed, ".") {
				if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
					return f
				}
			} else {
				if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
					return i
				}
			}
		}
	}
	return trimmed
}
