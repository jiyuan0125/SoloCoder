package passwordstrength

import (
	"github.com/passwordstrength/common"
)

const (
	maxLengthScore    = 25.0
	maxDiversityScore = 25.0
	maxPatternPenalty = 20.0
	maxEntropyBonus   = 30.0
)

func Evaluate(password string) *common.EvaluateResponse {
	if IsCommonPassword(password) {
		return buildCommonResponse()
	}
	hasLower, hasUpper, hasDigit, hasSpecial := checkCharTypes(password)
	patterns := analyzePatterns(password)
	lengthScore := calcLengthScore(password)
	diversityScore := calcDiversityScore(hasLower, hasUpper, hasDigit, hasSpecial)
	patternPenalty := calcPatternPenalty(patterns)
	entropyTheoretical := calcTheoreticalEntropy(password)
	entropyEstimated := calcEstimatedEntropy(password, patterns)
	entropyBonus := calcEntropyBonus(entropyEstimated)
	total := lengthScore + diversityScore - patternPenalty + entropyBonus
	if total < 0 {
		total = 0
	}
	if total > 100 {
		total = 100
	}
	dimensions := []common.DimensionScore{
		{Name: "长度评分", Score: lengthScore, Max: maxLengthScore},
		{Name: "字符多样性", Score: diversityScore, Max: maxDiversityScore},
		{Name: "模式扣分项", Score: maxPatternPenalty - patternPenalty, Max: maxPatternPenalty},
		{Name: "熵值加分", Score: entropyBonus, Max: maxEntropyBonus},
	}
	suggestions := buildSuggestions(password, hasLower, hasUpper, hasDigit, hasSpecial, patterns)
	return &common.EvaluateResponse{
		TotalScore:  total,
		Level:       determineLevel(total),
		Dimensions:  dimensions,
		Entropy:     common.EntropyInfo{Theoretical: entropyTheoretical, Estimated: entropyEstimated},
		Suggestions: suggestions,
		IsCommon:    false,
	}
}

func calcLengthScore(password string) float64 {
	l := len(password)
	switch {
	case l < 8:
		return 0
	case l >= 12:
		return maxLengthScore
	default:
		return maxLengthScore * 0.5
	}
}

func calcDiversityScore(hasLower, hasUpper, hasDigit, hasSpecial bool) float64 {
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
	return (float64(count) / 4.0) * maxDiversityScore
}

func calcPatternPenalty(patterns *patternResult) float64 {
	seqPenalty := float64(patterns.consecutiveCount) * 4.0
	repeatPenalty := float64(patterns.repeatedCount) * 5.0
	keyboardPenalty := float64(patterns.keyboardCount) * 6.0
	total := seqPenalty + repeatPenalty + keyboardPenalty
	if total > maxPatternPenalty {
		total = maxPatternPenalty
	}
	return total
}

func calcEntropyBonus(estimatedEntropy float64) float64 {
	if estimatedEntropy <= 0 {
		return 0
	}
	bonus := (estimatedEntropy / 100.0) * maxEntropyBonus
	if bonus > maxEntropyBonus {
		bonus = maxEntropyBonus
	}
	return bonus
}

func determineLevel(total float64) common.StrengthLevel {
	switch {
	case total >= 80:
		return common.LevelVeryStrong
	case total >= 60:
		return common.LevelStrong
	case total >= 40:
		return common.LevelMedium
	default:
		return common.LevelWeak
	}
}

func buildSuggestions(password string, hasLower, hasUpper, hasDigit, hasSpecial bool, patterns *patternResult) []string {
	var suggestions []string
	if len(password) < 8 {
		suggestions = append(suggestions, "密码太短，建议至少8位，最好12位以上")
	} else if len(password) < 12 {
		suggestions = append(suggestions, "密码长度尚可，建议增加到12位以上获得更高安全性")
	}
	if !hasLower {
		suggestions = append(suggestions, "缺少小写字母，建议添加a-z中的字符")
	}
	if !hasUpper {
		suggestions = append(suggestions, "缺少大写字母，建议添加A-Z中的字符")
	}
	if !hasDigit {
		suggestions = append(suggestions, "缺少数字，建议添加0-9中的字符")
	}
	if !hasSpecial {
		suggestions = append(suggestions, "缺少特殊字符，建议添加!@#$%^&*()中的字符")
	}
	if patterns.consecutiveCount > 0 {
		suggestions = append(suggestions, "检测到连续递增或递减字符序列（如abc、123），建议避免此类模式")
	}
	if patterns.repeatedCount > 0 {
		suggestions = append(suggestions, "检测到重复字符（如aaa、111），建议避免同一字符连续出现3次以上")
	}
	if patterns.keyboardCount > 0 {
		suggestions = append(suggestions, "检测到键盘连续按键模式（如qwerty、asdf），建议避免此类模式")
	}
	return suggestions
}

func buildCommonResponse() *common.EvaluateResponse {
	return &common.EvaluateResponse{
		TotalScore:  0,
		Level:       common.LevelWeak,
		Dimensions:  []common.DimensionScore{},
		Entropy:     common.EntropyInfo{},
		Suggestions: []string{"密码在常见弱密码字典中，建议立即更换"},
		IsCommon:    true,
	}
}
