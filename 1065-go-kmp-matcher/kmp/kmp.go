package kmp

import "strings"

func BuildNext(pattern string) []int {
	n := len(pattern)
	next := make([]int, n)
	if n == 0 {
		return next
	}
	next[0] = 0
	for i := 1; i < n; i++ {
		j := next[i-1]
		for j > 0 && pattern[i] != pattern[j] {
			j = next[j-1]
		}
		if pattern[i] == pattern[j] {
			j++
		}
		next[i] = j
	}
	return next
}

type MatchOptions struct {
	CaseInsensitive bool
}

func SingleMatch(text, pattern string, opts MatchOptions) []int {
	n := len(text)
	m := len(pattern)
	if m == 0 || n < m {
		return []int{}
	}

	workText := text
	workPattern := pattern
	if opts.CaseInsensitive {
		workText = strings.ToLower(text)
		workPattern = strings.ToLower(pattern)
	}

	next := BuildNext(workPattern)
	results := []int{}
	j := 0
	for i := 0; i < n; i++ {
		for j > 0 && workText[i] != workPattern[j] {
			j = next[j-1]
		}
		if workText[i] == workPattern[j] {
			j++
		}
		if j == m {
			results = append(results, i-m+1)
			j = next[j-1]
		}
	}
	return results
}

func MultiMatch(text string, patterns []string, opts MatchOptions) map[string][]int {
	result := make(map[string][]int)
	for _, pattern := range patterns {
		positions := SingleMatch(text, pattern, opts)
		if positions == nil {
			positions = []int{}
		}
		result[pattern] = positions
	}
	return result
}
