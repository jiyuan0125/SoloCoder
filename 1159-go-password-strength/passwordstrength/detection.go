package passwordstrength

import (
	"unicode"
)

func checkCharTypes(password string) (hasLower, hasUpper, hasDigit, hasSpecial bool) {
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsSymbol(r) || unicode.IsPunct(r):
			hasSpecial = true
		}
	}
	return
}

func detectConsecutiveSeq(password string) int {
	if len(password) < 3 {
		return 0
	}
	count := 0
	chars := []rune(password)
	for i := 0; i < len(chars)-2; i++ {
		c1 := chars[i]
		c2 := chars[i+1]
		c3 := chars[i+2]
		isInc := c2 == c1+1 && c3 == c2+1
		isDec := c2 == c1-1 && c3 == c2-1
		if isInc || isDec {
			count++
		}
	}
	return count
}

func detectRepeatedChars(password string) int {
	if len(password) < 3 {
		return 0
	}
	count := 0
	chars := []rune(password)
	for i := 0; i < len(chars)-2; i++ {
		if chars[i] == chars[i+1] && chars[i] == chars[i+2] {
			count++
		}
	}
	return count
}

func detectKeyboardPatterns(password string) int {
	keyboardRows := []string{
		"qwertyuiop",
		"asdfghjkl",
		"zxcvbnm",
		"1234567890",
		"QWERTYUIOP",
		"ASDFGHJKL",
		"ZXCVBNM",
	}
	lowPwd := toLower(password)
	patternCount := 0
	for _, row := range keyboardRows {
		patternCount += countPatternMatches(lowPwd, row)
	}
	return patternCount
}

func countPatternMatches(password, row string) int {
	count := 0
	if len(row) < 4 {
		return 0
	}
	for start := 0; start <= len(row)-4; start++ {
		pattern := row[start : start+4]
		if containsSub(password, pattern) {
			count++
		}
	}
	return count
}

func containsSub(s, sub string) bool {
	if len(sub) == 0 || len(s) < len(sub) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = unicode.ToLower(r)
	}
	return string(runes)
}
