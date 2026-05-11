package fuzzy

func NormalizeWildcards(pattern string) string {
	result := []rune{}
	runes := []rune(pattern)
	prevStar := false
	
	for _, r := range runes {
		if r == '*' {
			if !prevStar {
				result = append(result, r)
				prevStar = true
			}
		} else {
			result = append(result, r)
			prevStar = false
		}
	}
	
	return string(result)
}

func WildcardMatch(pattern, text string, threshold int) (bool, int) {
	pattern = NormalizeWildcards(pattern)
	patternRunes := []rune(pattern)
	textRunes := []rune(text)
	
	m, n := len(patternRunes), len(textRunes)
	
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		for j := range dp[i] {
			dp[i][j] = -1
		}
	}
	
	dp[0][0] = 0
	
	for i := 1; i <= m; i++ {
		if patternRunes[i-1] == '*' {
			dp[i][0] = dp[i-1][0]
		} else {
			break
		}
	}
	
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if patternRunes[i-1] == '?' {
				if dp[i-1][j-1] != -1 {
					dp[i][j] = dp[i-1][j-1]
				}
			} else if patternRunes[i-1] == '*' {
				options := []int{}
				if dp[i-1][j] != -1 {
					options = append(options, dp[i-1][j])
				}
				if dp[i][j-1] != -1 {
					options = append(options, dp[i][j-1])
				}
				if len(options) > 0 {
					dp[i][j] = minInt(options...)
				}
			} else {
				if patternRunes[i-1] == textRunes[j-1] {
					if dp[i-1][j-1] != -1 {
						dp[i][j] = dp[i-1][j-1]
					}
				} else {
					options := []int{}
					if dp[i-1][j] != -1 {
						options = append(options, dp[i-1][j]+1)
					}
					if dp[i][j-1] != -1 {
						options = append(options, dp[i][j-1]+1)
					}
					if dp[i-1][j-1] != -1 {
						options = append(options, dp[i-1][j-1]+1)
					}
					if len(options) > 0 {
						dp[i][j] = minInt(options...)
					}
				}
			}
		}
	}
	
	distance := dp[m][n]
	if distance == -1 {
		return false, 0
	}
	
	return distance <= threshold, distance
}

func minInt(ints ...int) int {
	if len(ints) == 0 {
		return 0
	}
	minVal := ints[0]
	for _, val := range ints[1:] {
		if val < minVal {
			minVal = val
		}
	}
	return minVal
}
