package suffixarray

import "strings"

type RepeatedSubstring struct {
	Substring string
	Positions []int
}

type SubstringOccurrences struct {
	Count     int
	Positions []int
}

func (sa *SuffixArray) LongestRepeatedSubstrings() []RepeatedSubstring {
	n := len(sa.sa)
	if n <= 1 || len(sa.lcp) == 0 {
		return []RepeatedSubstring{}
	}

	maxLen := 0
	for _, l := range sa.lcp {
		if l > maxLen {
			maxLen = l
		}
	}

	if maxLen == 0 {
		return []RepeatedSubstring{}
	}

	seen := make(map[string]*RepeatedSubstring)
	for i, lcp := range sa.lcp {
		if lcp == maxLen {
			start := sa.sa[i]
			substr := sa.text[start : start+maxLen]
			if rs, ok := seen[substr]; ok {
				rs.Positions = append(rs.Positions, sa.sa[i+1])
			} else {
				seen[substr] = &RepeatedSubstring{
					Substring: substr,
					Positions: []int{start, sa.sa[i+1]},
				}
			}
		}
	}

	result := make([]RepeatedSubstring, 0, len(seen))
	for _, rs := range seen {
		result = append(result, *rs)
	}

	return result
}

func (sa *SuffixArray) CountOccurrences(pattern string) *SubstringOccurrences {
	if len(pattern) == 0 {
		n := len(sa.text)
		positions := make([]int, n)
		for i := 0; i < n; i++ {
			positions[i] = i
		}
		return &SubstringOccurrences{
			Count:     n,
			Positions: positions,
		}
	}

	n := len(sa.text)
	m := len(pattern)

	if m > n {
		return &SubstringOccurrences{
			Count:     0,
			Positions: []int{},
		}
	}

	left := sa.findLeftBound(pattern)
	right := sa.findRightBound(pattern)

	if left >= right {
		return &SubstringOccurrences{
			Count:     0,
			Positions: []int{},
		}
	}

	positions := make([]int, right-left)
	for i := left; i < right; i++ {
		positions[i-left] = sa.sa[i]
	}

	return &SubstringOccurrences{
		Count:     right - left,
		Positions: positions,
	}
}

func (sa *SuffixArray) findLeftBound(pattern string) int {
	low, high := 0, len(sa.sa)
	m := len(pattern)
	n := len(sa.text)

	for low < high {
		mid := (low + high) / 2
		idx := sa.sa[mid]
		end := idx + m
		if end > n {
			end = n
		}
		if pattern > sa.text[idx:end] {
			low = mid + 1
		} else {
			high = mid
		}
	}

	return low
}

func (sa *SuffixArray) findRightBound(pattern string) int {
	low, high := 0, len(sa.sa)
	m := len(pattern)
	n := len(sa.text)

	for low < high {
		mid := (low + high) / 2
		idx := sa.sa[mid]
		end := idx + m
		if end > n {
			end = n
		}
		if strings.HasPrefix(sa.text[idx:end], pattern) || pattern > sa.text[idx:end] {
			low = mid + 1
		} else {
			high = mid
		}
	}

	return low
}
