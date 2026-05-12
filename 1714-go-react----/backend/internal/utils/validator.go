package utils

import (
	"regexp"
	"unicode"
)

func IsValidCreditCode(code string) bool {
	if len(code) != 18 {
		return false
	}

	match, _ := regexp.MatchString(`^[0-9A-HJ-NPQRTUWXY]{18}$`, code)
	if !match {
		return false
	}

	return CheckCreditCodeChecksum(code)
}

func CheckCreditCodeChecksum(code string) bool {
	if len(code) != 18 {
		return false
	}

	weights := []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
	charMap := map[rune]int{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15, 'G': 16, 'H': 17,
		'J': 18, 'K': 19, 'L': 20, 'M': 21, 'N': 22, 'P': 23, 'Q': 24,
		'R': 25, 'T': 26, 'U': 27, 'W': 28, 'X': 29, 'Y': 30,
	}

	var sum int
	for i, char := range code[:17] {
		if val, ok := charMap[unicode.ToUpper(char)]; ok {
			sum += val * weights[i]
		} else {
			return false
		}
	}

	mod := sum % 31
	checkValue := 31 - mod
	if checkValue == 31 {
		checkValue = 0
	}

	lastChar := unicode.ToUpper(rune(code[17]))
	for char, val := range charMap {
		if val == checkValue && char == lastChar {
			return true
		}
	}

	return false
}

func IsValidPhone(phone string) bool {
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	if match {
		return true
	}
	match, _ = regexp.MatchString(`^0\d{2,3}-?\d{7,8}$`, phone)
	return match
}
