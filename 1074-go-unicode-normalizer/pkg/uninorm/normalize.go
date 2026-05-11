package uninorm

import (
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
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

	var nf norm.Form
	switch form {
	case NFC:
		nf = norm.NFC
	case NFD:
		nf = norm.NFD
	case NFKC:
		nf = norm.NFKC
	case NFKD:
		nf = norm.NFKD
	default:
		return input
	}

	return nf.String(input)
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
	OriginalRunes []rune
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
	normalizedRunes := []rune(normalized)

	changes := detectClusterChanges(inputRunes, normalizedRunes, form)

	return changes, normalized
}

func detectClusterChanges(input, output []rune, form Form) []CharChange {
	var changes []CharChange

	i, j := 0, 0

	for i < len(input) || j < len(output) {
		if i < len(input) && j < len(output) && input[i] == output[j] {
			i++
			j++
			continue
		}

		var inputCluster []rune
		if i < len(input) {
			inputCluster = extractCluster(input, i)
		}

		var outputCluster []rune
		if j < len(output) {
			outputCluster = extractCluster(output, j)
		}

		if len(inputCluster) == 0 && len(outputCluster) == 0 {
			break
		}

		clustersMatched := false
		if len(inputCluster) > 0 && len(outputCluster) > 0 {
			inputStr := string(inputCluster)
			outputStr := string(outputCluster)
			normalizedInput := Normalize(inputStr, form)

			if normalizedInput == outputStr {
				changes = append(changes, CharChange{
					OriginalRunes: inputCluster,
					Normalized:    outputCluster,
					OriginalStr:   inputStr,
					NormalizedStr: outputStr,
				})
				i += len(inputCluster)
				j += len(outputCluster)
				clustersMatched = true
			}
		}

		if !clustersMatched {
			if len(inputCluster) > 0 {
				inputStr := string(inputCluster)
				expectedOutput := Normalize(inputStr, form)
				expectedRunes := []rune(expectedOutput)

				changes = append(changes, CharChange{
					OriginalRunes: inputCluster,
					Normalized:    expectedRunes,
					OriginalStr:   inputStr,
					NormalizedStr: expectedOutput,
				})
				i += len(inputCluster)
			}
			if len(outputCluster) > 0 {
				j += len(outputCluster)
			}
		}
	}

	return changes
}

func extractCluster(runes []rune, start int) []rune {
	if start >= len(runes) {
		return nil
	}

	cluster := []rune{runes[start]}
	i := start + 1

	for i < len(runes) {
		r := runes[i]
		if isMark(r) {
			cluster = append(cluster, r)
			i++
		} else {
			break
		}
	}

	return cluster
}

func isMark(r rune) bool {
	category := runeCategory(r)
	return category == 'M'
}

func runeCategory(r rune) byte {
	switch {
	case r >= 0x0300 && r <= 0x036F:
		return 'M'
	case r >= 0x1AB0 && r <= 0x1AFF:
		return 'M'
	case r >= 0x1DC0 && r <= 0x1DFF:
		return 'M'
	case r >= 0x20D0 && r <= 0x20FF:
		return 'M'
	case r >= 0xFE20 && r <= 0xFE2F:
		return 'M'
	default:
		return 'L'
	}
}

func IsNormalized(input string, form Form) bool {
	var nf norm.Form
	switch form {
	case NFC:
		nf = norm.NFC
	case NFD:
		nf = norm.NFD
	case NFKC:
		nf = norm.NFKC
	case NFKD:
		nf = norm.NFKD
	default:
		return true
	}

	return nf.IsNormalString(input)
}
