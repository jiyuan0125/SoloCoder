package passwordstrength

import "math"

func calcTheoreticalEntropy(password string) float64 {
	chars := []rune(password)
	hasLower, hasUpper, hasDigit, hasSpecial := checkCharTypes(password)
	charSetSize := 0.0
	if hasLower {
		charSetSize += 26
	}
	if hasUpper {
		charSetSize += 26
	}
	if hasDigit {
		charSetSize += 10
	}
	if hasSpecial {
		charSetSize += 32
	}
	if charSetSize == 0 {
		charSetSize = 1
	}
	return float64(len(chars)) * math.Log2(charSetSize)
}

func calcEstimatedEntropy(password string, patterns *patternResult) float64 {
	base := calcTheoreticalEntropy(password)
	seqPenalty := float64(patterns.consecutiveCount) * 2.0
	repeatPenalty := float64(patterns.repeatedCount) * 3.0
	keyboardPenalty := float64(patterns.keyboardCount) * 5.0
	totalPenalty := seqPenalty + repeatPenalty + keyboardPenalty
	if totalPenalty >= base {
		return base * 0.3
	}
	return base - totalPenalty
}

type patternResult struct {
	consecutiveCount int
	repeatedCount    int
	keyboardCount    int
}

func analyzePatterns(password string) *patternResult {
	return &patternResult{
		consecutiveCount: detectConsecutiveSeq(password),
		repeatedCount:    detectRepeatedChars(password),
		keyboardCount:    detectKeyboardPatterns(password),
	}
}
