package bmsearch

import (
	"fmt"
)

const ASCIIMax = 256

type PreprocessResult struct {
	Pattern string
	BmBc    []int
	BmGs    []int
}

func BuildBadCharTable(pattern string) []int {
	m := len(pattern)
	bmBc := make([]int, ASCIIMax)
	for i := range bmBc {
		bmBc[i] = m
	}
	for i := 0; i < m; i++ {
		c := pattern[i]
		bmBc[c] = m - 1 - i
	}
	return bmBc
}

func BuildZArray(s string) []int {
	n := len(s)
	if n == 0 {
		return nil
	}
	Z := make([]int, n)
	Z[0] = n
	l, r := 0, 0
	for i := 1; i < n; i++ {
		if i > r {
			l, r = i, i
			for r < n && s[r-l] == s[r] {
				r++
			}
			Z[i] = r - l
			r--
		} else {
			k := i - l
			if Z[k] < r-i+1 {
				Z[i] = Z[k]
			} else {
				l = i
				for r < n && s[r-l] == s[r] {
					r++
				}
				Z[i] = r - l
				r--
			}
		}
	}
	return Z
}

func BuildGoodSuffixTable(pattern string) []int {
	m := len(pattern)
	if m == 0 {
		return nil
	}

	bmGs := make([]int, m)
	for i := range bmGs {
		bmGs[i] = m
	}

	for L := 1; L < m; L++ {
		for i := m - L - 1; i >= 0; i-- {
			match := true
			for k := 0; k < L; k++ {
				if pattern[i+k] != pattern[m-L+k] {
					match = false
					break
				}
			}
			if match {
				if bmGs[m-L-1] == m {
					bmGs[m-L-1] = m - L - i
				}
			}
		}
	}

	for L := m - 1; L >= 0; L-- {
		match := true
		for k := 0; k < L; k++ {
			if pattern[k] != pattern[m-L+k] {
				match = false
				break
			}
		}
		if match {
			for i := 0; i <= m-L-1; i++ {
				if bmGs[i] == m {
					bmGs[i] = m - L
				}
			}
		}
	}

	return bmGs
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func Preprocess(pattern string) (*PreprocessResult, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern is empty")
	}
	bmBc := BuildBadCharTable(pattern)
	bmGs := BuildGoodSuffixTable(pattern)
	return &PreprocessResult{
		Pattern: pattern,
		BmBc:    bmBc[:],
		BmGs:    bmGs,
	}, nil
}

func Search(text, pattern string) []int {
	n, m := len(text), len(pattern)
	if m == 0 || n == 0 || m > n {
		return []int{}
	}

	bmBc := BuildBadCharTable(pattern)
	bmGs := BuildGoodSuffixTable(pattern)

	return searchWithTables(text, pattern, bmBc, bmGs)
}

func SearchWithPreprocess(text string, prep *PreprocessResult) []int {
	if prep == nil {
		return []int{}
	}
	n, m := len(text), len(prep.Pattern)
	if m == 0 || n == 0 || m > n {
		return []int{}
	}
	return searchWithTables(text, prep.Pattern, prep.BmBc, prep.BmGs)
}

func searchWithTables(text, pattern string, bmBc []int, bmGs []int) []int {
	n, m := len(text), len(pattern)
	result := []int{}

	i := 0
	for i <= n-m {
		j := m - 1
		for j >= 0 && pattern[j] == text[i+j] {
			j--
		}
		if j < 0 {
			result = append(result, i)
			if m > 0 && bmGs[0] > 0 {
				i += bmGs[0]
			} else {
				i += 1
			}
		} else {
			bcShift := bmBc[text[i+j]] - (m - 1 - j)
			if bcShift < 1 {
				bcShift = 1
			}
			if j == m-1 {
				i += bcShift
			} else {
				gsShift := bmGs[j]
				if bcShift > gsShift {
					i += bcShift
				} else {
					i += gsShift
				}
			}
		}
	}

	return result
}
