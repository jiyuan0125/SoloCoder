package uninorm

import (
	"unicode/utf8"
)

type Form int

const (
	NFC Form = iota
	NFD
	NFKC
	NFKD
)

func (f Form) String() string {
	switch f {
	case NFC:
		return "NFC"
	case NFD:
		return "NFD"
	case NFKC:
		return "NFKC"
	case NFKD:
		return "NFKD"
	default:
		return "Unknown"
	}
}

func ParseForm(s string) (Form, error) {
	switch s {
	case "NFC", "nfc":
		return NFC, nil
	case "NFD", "nfd":
		return NFD, nil
	case "NFKC", "nfkc":
		return NFKC, nil
	case "NFKD", "nfkd":
		return NFKD, nil
	default:
		return NFC, &InvalidFormError{Form: s}
	}
}

type InvalidFormError struct {
	Form string
}

func (e *InvalidFormError) Error() string {
	return "invalid normalization form: " + e.Form
}

const UnicodeVersion = "15.0"

func Normalize(input string, form Form) string {
	if input == "" {
		return ""
	}
	
	if isPureASCII(input) {
		return input
	}
	
	return customNormalize(input, form)
}

func isPureASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

type CharChange struct {
	Original      rune
	Normalized    []rune
	OriginalStr   string
	NormalizedStr string
}

func AnalyzeChanges(input string, form Form) ([]CharChange, string) {
	normalized := Normalize(input, form)
	
	if input == normalized {
		return []CharChange{}, normalized
	}
	
	inputRunes := []rune(input)
	
	var changes []CharChange
	
	for _, r := range inputRunes {
		original := string(r)
		normStr := Normalize(original, form)
		
		if original != normStr {
			changes = append(changes, CharChange{
				Original:      r,
				Normalized:    []rune(normStr),
				OriginalStr:   original,
				NormalizedStr: normStr,
			})
		}
	}
	
	return changes, normalized
}

func IsNormalized(input string, form Form) bool {
	return input == Normalize(input, form)
}
