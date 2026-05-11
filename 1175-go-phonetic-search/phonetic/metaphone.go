package phonetic

import (
	"strings"
)

func Metaphone(name string) string {
	name = filterLetters(name)
	if name == "" {
		return ""
	}

	upper := strings.ToUpper(name)
	runes := []rune(upper)

	result := make([]rune, 0, 4)

	i := 0
	n := len(runes)

	if n > 2 && runes[0] == 'S' && runes[1] == 'C' && runes[2] == 'H' {
		result = append(result, 'X')
		i = 3
	} else if n > 1 {
		switch {
		case runes[0] == 'A' && runes[1] == 'E':
			result = append(result, 'E')
			i = 2
		case runes[0] == 'P' && runes[1] == 'H':
			result = append(result, 'F')
			i = 2
		case (runes[0] == 'G' || runes[0] == 'K') && runes[1] == 'N':
			result = append(result, runes[1])
			i = 2
		case runes[0] == 'W' && runes[1] == 'R':
			result = append(result, 'R')
			i = 2
		case runes[0] == 'W' && runes[1] == 'H':
			result = append(result, 'W')
			i = 2
		}
	}

	if len(result) == 0 && n > 0 {
		if runes[0] == 'X' {
			result = append(result, 'S')
		} else {
			result = append(result, runes[0])
		}
		i = 1
	}

	for i < n && len(result) < 4 {
		current := runes[i]

		var (
			prev rune = 0
			next rune = 0
		)
		if i > 0 {
			prev = runes[i-1]
		}
		if i+1 < n {
			next = runes[i+1]
		}

		switch current {
		case 'B':
			if i+1 >= n && prev == 'M' {
				i++
				continue
			}
			result = append(result, 'P')
			i++

		case 'C':
			if i+1 < n {
				if runes[i+1] == 'I' || runes[i+1] == 'E' || runes[i+1] == 'Y' {
					if i+2 < n && string(runes[i:i+3]) == "CIA" {
						result = append(result, 'X')
						i += 3
						continue
					}
					result = append(result, 'S')
					i += 2
					continue
				}
				if runes[i+1] == 'K' {
					result = append(result, 'K')
					i += 2
					continue
				}
				if runes[i+1] == 'H' {
					result = append(result, 'X')
					i += 2
					continue
				}
			}
			result = append(result, 'K')
			i++

		case 'D':
			if i+2 < n && (runes[i+1] == 'G' && (runes[i+2] == 'E' || runes[i+2] == 'I' || runes[i+2] == 'Y')) {
				result = append(result, 'J')
				i += 3
				continue
			}
			result = append(result, 'T')
			i++

		case 'G':
			if i+1 < n {
				if runes[i+1] == 'N' {
					if i+2 >= n || !isVowel(runes[i+2]) {
						i++
						continue
					}
				}
				if runes[i+1] == 'H' {
					if i+2 >= n || !isVowel(runes[i+2]) {
						i++
						continue
					}
					if i > 0 && isVowel(prev) {
						result = append(result, 'F')
						i += 2
						continue
					}
				}
				if runes[i+1] == 'E' || runes[i+1] == 'I' || runes[i+1] == 'Y' {
					result = append(result, 'J')
					i += 2
					continue
				}
			}
			if i > 0 && (prev == 'D' || prev == 'G' || prev == 'C') {
				i++
				continue
			}
			result = append(result, 'K')
			i++

		case 'F', 'J', 'L', 'M', 'N', 'R':
			result = append(result, current)
			i++

		case 'K':
			if i > 0 && prev == 'C' {
				i++
				continue
			}
			result = append(result, 'K')
			i++

		case 'P':
			if i+1 < n && runes[i+1] == 'H' {
				result = append(result, 'F')
				i += 2
				continue
			}
			result = append(result, 'P')
			i++

		case 'Q':
			result = append(result, 'K')
			i++

		case 'S':
			if i+1 < n {
				if runes[i+1] == 'H' {
					result = append(result, 'X')
					i += 2
					continue
				}
				if (runes[i+1] == 'I' && i+2 < n) && (runes[i+2] == 'O' || runes[i+2] == 'A') {
					result = append(result, 'X')
					i += 3
					continue
				}
			}
			result = append(result, 'S')
			i++

		case 'T':
			if i+1 < n {
				if runes[i+1] == 'H' {
					result = append(result, '0')
					i += 2
					continue
				}
				if (runes[i+1] == 'I' && i+2 < n) && (runes[i+2] == 'O' || runes[i+2] == 'A') {
					result = append(result, 'X')
					i += 3
					continue
				}
			}
			result = append(result, 'T')
			i++

		case 'V':
			result = append(result, 'F')
			i++

		case 'W', 'Y':
			if i+1 < n && isVowel(runes[i+1]) {
				result = append(result, current)
				i++
				continue
			}
			i++

		case 'X':
			result = append(result, 'S')
			i++

		case 'Z':
			result = append(result, 'S')
			i++

		case 'A', 'E', 'I', 'O', 'U':
			if i == 0 {
				result = append(result, 'A')
			}
			i++

		case 'H':
			if i > 0 && i+1 < n && isVowel(prev) && isVowel(next) {
				result = append(result, 'H')
			}
			i++

		default:
			i++
		}
	}

	return string(result)
}
