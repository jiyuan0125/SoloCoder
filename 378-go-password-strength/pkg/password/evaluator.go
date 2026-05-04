package password

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidPassword = errors.New("密码不能为空或全为空格")

func (e *Evaluator) Evaluate(password string) (*Result, error) {
	trimmed := strings.TrimSpace(password)
	if trimmed == "" {
		return nil, ErrInvalidPassword
	}

	result := &Result{
		Level:       LevelWeak,
		Score:       0,
		Suggestions: []string{},
	}

	if len(password) < 6 {
		result.Suggestions = append(result.Suggestions, "建议密码长度至少6位")
		return result, nil
	}

	if e.isWeakPassword(password) {
		result.Suggestions = append(result.Suggestions, "建议避免使用常见弱密码")
		return result, nil
	}

	score := 0
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		if unicode.IsLower(c) {
			hasLower = true
		} else if unicode.IsUpper(c) {
			hasUpper = true
		} else if unicode.IsDigit(c) {
			hasDigit = true
		} else if isSpecialChar(c) {
			hasSpecial = true
		}
	}

	if len(password) >= 12 {
		score += 2
	} else if len(password) >= 8 {
		score += 1
	} else {
		result.Suggestions = append(result.Suggestions, "建议密码长度至少8位")
	}

	if hasLower {
		score += 1
	} else {
		result.Suggestions = append(result.Suggestions, "建议增加小写字母")
	}

	if hasUpper {
		score += 1
	} else {
		result.Suggestions = append(result.Suggestions, "建议增加大写字母")
	}

	if hasDigit {
		score += 1
	} else {
		result.Suggestions = append(result.Suggestions, "建议增加数字")
	}

	if hasSpecial {
		score += 1
	} else {
		result.Suggestions = append(result.Suggestions, "建议增加特殊字符")
	}

	penalty := 0

	if e.hasConsecutiveRepeats(password) {
		penalty += 1
		result.Suggestions = append(result.Suggestions, "建议避免连续重复字符超过3个")
	}

	charTypeCount := 0
	if hasLower {
		charTypeCount++
	}
	if hasUpper {
		charTypeCount++
	}
	if hasDigit {
		charTypeCount++
	}
	if hasSpecial {
		charTypeCount++
	}

	if charTypeCount == 1 {
		penalty += 1
		result.Suggestions = append(result.Suggestions, "建议使用多种字符类型组合")
	}

	if e.hasKeyboardSequence(password) {
		penalty += 1
		result.Suggestions = append(result.Suggestions, "建议避免使用键盘上相邻的按键序列")
	}

	score -= penalty
	if score < 0 {
		score = 0
	}

	result.Score = score

	if score <= 2 {
		result.Level = LevelWeak
	} else if score <= 4 {
		result.Level = LevelMedium
	} else if score <= 5 {
		result.Level = LevelStrong
	} else {
		result.Level = LevelVeryStrong
	}

	return result, nil
}

func isSpecialChar(c rune) bool {
	return c >= 33 && c <= 126 && !unicode.IsLetter(c) && !unicode.IsDigit(c)
}

func (e *Evaluator) hasConsecutiveRepeats(password string) bool {
	if len(password) < 4 {
		return false
	}

	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count > 3 {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

func Evaluate(password string) (*Result, error) {
	e := NewEvaluator()
	return e.Evaluate(password)
}
