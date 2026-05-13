package generator

import (
	"unicode"
)

type StrengthLevel string

const (
	StrengthWeak   StrengthLevel = "弱"
	StrengthMedium StrengthLevel = "中"
	StrengthStrong StrengthLevel = "强"
)

type StrengthResult struct {
	Level    StrengthLevel
	Score    int
	Details  map[string]int
}

func (g *Generator) EvaluateStrength(password string) *StrengthResult {
	result := &StrengthResult{
		Details: make(map[string]int),
	}

	score := 0

	lengthScore := evaluateLength(password)
	score += lengthScore
	result.Details["length"] = lengthScore

	charTypesScore := evaluateCharTypes(password)
	score += charTypesScore
	result.Details["char_types"] = charTypesScore

	patternScore := evaluatePatterns(password)
	score += patternScore
	result.Details["patterns"] = patternScore

	result.Score = score

	if score < 40 {
		result.Level = StrengthWeak
	} else if score < 70 {
		result.Level = StrengthMedium
	} else {
		result.Level = StrengthStrong
	}

	return result
}

func evaluateLength(password string) int {
	length := len(password)
	if length < 8 {
		return 5
	} else if length < 12 {
		return 20
	} else if length < 16 {
		return 35
	} else {
		return 40
	}
}

func evaluateCharTypes(password string) int {
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		if unicode.IsLower(ch) {
			hasLower = true
		} else if unicode.IsUpper(ch) {
			hasUpper = true
		} else if unicode.IsDigit(ch) {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	count := 0
	if hasLower {
		count++
	}
	if hasUpper {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	switch count {
	case 1:
		return 0
	case 2:
		return 15
	case 3:
		return 30
	case 4:
		return 40
	}

	return 0
}

func evaluatePatterns(password string) int {
	score := 0

	if hasConsecutiveDigits(password) {
		score -= 10
	}

	if hasRepeatedChars(password) {
		score -= 15
	}

	if hasKeyboardPattern(password) {
		score -= 10
	}

	if score < -20 {
		score = -20
	}

	return score
}

func hasConsecutiveDigits(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if unicode.IsDigit(rune(password[i])) &&
			unicode.IsDigit(rune(password[i+1])) &&
			unicode.IsDigit(rune(password[i+2])) {
			if password[i+1]-password[i] == 1 && password[i+2]-password[i+1] == 1 {
				return true
			}
			if password[i]-password[i+1] == 1 && password[i+1]-password[i+2] == 1 {
				return true
			}
		}
	}
	return false
}

func hasRepeatedChars(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}
	return false
}

func hasKeyboardPattern(password string) bool {
	keyboardPatterns := []string{
		"qwerty", "asdfgh", "zxcvbn",
		"123456", "654321",
		"abcdef", "fedcba",
	}

	lowerPwd := toLower(password)
	for _, pattern := range keyboardPatterns {
		if containsSubstring(lowerPwd, pattern) {
			return true
		}
	}

	return false
}

func toLower(s string) string {
	result := make([]rune, 0, len(s))
	for _, ch := range s {
		result = append(result, unicode.ToLower(ch))
	}
	return string(result)
}

func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
