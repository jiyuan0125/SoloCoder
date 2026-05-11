package phonetic

import (
	"strings"
)

var soundexMap = map[rune]byte{
	'B': '1', 'F': '1', 'P': '1', 'V': '1',
	'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
	'D': '3', 'T': '3',
	'L': '4',
	'M': '5', 'N': '5',
	'R': '6',
}

var soundexFirstCharMap = map[rune]byte{
	'A': '0', 'E': '0', 'I': '0', 'O': '0', 'U': '0', 'H': '0', 'W': '0', 'Y': '0',
	'B': '1', 'F': '1', 'P': '1', 'V': '1',
	'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
	'D': '3', 'T': '3',
	'L': '4',
	'M': '5', 'N': '5',
	'R': '6',
}

func getSoundexCode(r rune) (byte, bool) {
	switch r {
	case 'B', 'F', 'P', 'V':
		return '1', true
	case 'C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z':
		return '2', true
	case 'D', 'T':
		return '3', true
	case 'L':
		return '4', true
	case 'M', 'N':
		return '5', true
	case 'R':
		return '6', true
	}
	return 0, false
}

func Soundex(name string) string {
	name = filterLetters(name)
	if name == "" {
		return ""
	}

	upper := strings.ToUpper(name)
	runes := []rune(upper)

	firstChar := runes[0]
	result := []rune{firstChar}

	var lastCode byte
	if code, ok := getSoundexCode(firstChar); ok {
		lastCode = code
	}

	for i := 1; i < len(runes); i++ {
		r := runes[i]
		code, ok := getSoundexCode(r)
		if !ok {
			if r == 'H' || r == 'W' {
				continue
			}
			lastCode = 0
			continue
		}

		if code != lastCode {
			result = append(result, rune(code))
			lastCode = code
		}

		if len(result) >= 4 {
			break
		}
	}

	for len(result) < 4 {
		result = append(result, '0')
	}

	return string(result[:4])
}
