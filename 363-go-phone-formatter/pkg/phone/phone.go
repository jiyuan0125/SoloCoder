package phone

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidPhoneNumber = errors.New("invalid phone number")

type FormatType int

const (
	FormatDomestic FormatType = iota
	FormatInternational
	FormatPureNumber
)

func Normalize(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "-", "")
	input = strings.ReplaceAll(input, "(", "")
	input = strings.ReplaceAll(input, ")", "")

	if strings.HasPrefix(input, "0086") {
		input = "+86" + input[4:]
	}

	if strings.HasPrefix(input, "+86") {
		input = input[3:]
	}

	return input
}

func Format(input string, formatType FormatType) (string, error) {
	normalized := Normalize(input)

	if !Validate(normalized) {
		return "", ErrInvalidPhoneNumber
	}

	switch formatType {
	case FormatDomestic:
		return normalized[0:3] + " " + normalized[3:7] + " " + normalized[7:11], nil
	case FormatInternational:
		return "+86 " + normalized[0:3] + " " + normalized[3:7] + " " + normalized[7:11], nil
	case FormatPureNumber:
		return normalized, nil
	default:
		return "", ErrInvalidPhoneNumber
	}
}

func Validate(phone string) bool {
	if phone == "" {
		return false
	}

	normalized := Normalize(phone)

	if len(normalized) != 11 {
		return false
	}

	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, normalized)
	return match
}

func ExtractFromText(text string) []string {
	pattern := `(?:\+?86|0086)?[-\s()]*1[3-9][-\s()\d]{9,}`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(text, -1)

	var results []string
	seen := make(map[string]bool)

	for _, match := range matches {
		normalized := Normalize(match)
		if Validate(normalized) && !seen[normalized] {
			seen[normalized] = true
			results = append(results, normalized)
		}
	}

	return results
}
