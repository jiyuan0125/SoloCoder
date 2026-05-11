package baseconv

import (
	"errors"
	"strings"
)

func ValidateBase(base int) error {
	if base < MinBase || base > MaxBase {
		return errors.New("base must be between 2 and 62")
	}
	return nil
}

func ValidateNumber(s string, base int) error {
	if err := ValidateBase(base); err != nil {
		return err
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("empty string")
	}

	idx := 0
	if isSign(s[0]) {
		idx++
		if idx == len(s) {
			return errors.New("sign without digits")
		}
	}

	hasIntPart := false
	hasFracPart := false
	hasDot := false

	for i := idx; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			if hasDot {
				return errors.New("multiple decimal points")
			}
			hasDot = true
			continue
		}

		if !isValidForBase(c, base) {
			return errors.New("invalid character '" + string(c) + "' for base " + string(rune(base+'0')))
		}

		if !hasDot {
			hasIntPart = true
		} else {
			hasFracPart = true
		}
	}

	if !hasIntPart && !hasFracPart {
		return errors.New("no digits")
	}

	return nil
}
