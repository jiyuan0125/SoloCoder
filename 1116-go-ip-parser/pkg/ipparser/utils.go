package ipparser

import (
	"strconv"
	"strings"
)

func parseOctet(s string) (uint64, error) {
	base := 10
	str := s

	if len(str) >= 2 && str[0] == '0' {
		if str[1] == 'x' || str[1] == 'X' {
			base = 16
			str = str[2:]
		} else {
			base = 8
		}
	}

	if base == 8 {
		for _, c := range str {
			if c < '0' || c > '7' {
				return 0, &ParseError{Message: "invalid octal digit: " + string(c)}
			}
		}
	}

	val, err := strconv.ParseUint(str, base, 32)
	if err != nil {
		return 0, &ParseError{Message: err.Error()}
	}

	return val, nil
}

func parseHexGroup(s string) (uint16, error) {
	if len(s) == 0 {
		return 0, &ParseError{Message: "empty hex group"}
	}
	if len(s) > 4 {
		return 0, &ParseError{Message: "hex group too long"}
	}

	val, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0, &ParseError{Message: err.Error()}
	}

	return uint16(val), nil
}

func isIPv4Segment(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || (c == 'x' || c == 'X')) {
			return false
		}
	}

	_, err := parseOctet(s)
	return err == nil
}

func countDots(s string) int {
	return strings.Count(s, ".")
}

func countColons(s string) int {
	return strings.Count(s, ":")
}

func hasDoubleColon(s string) bool {
	return strings.Contains(s, "::")
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}
