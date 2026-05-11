package suffixarray

import (
	"sort"
)

type SuffixArray struct {
	sa   []int
	rank []int
	lcp  []int
	text string
}

func Build(text string) *SuffixArray {
	n := len(text)
	if n == 0 {
		return &SuffixArray{
			sa:   []int{0},
			rank: []int{0},
			lcp:  []int{},
			text: text,
		}
	}

	sa := make([]int, n)
	rank := make([]int, n)

	for i := 0; i < n; i++ {
		sa[i] = i
		rank[i] = int(text[i])
	}

	for k := 1; k < n; k <<= 1 {
		sort.Slice(sa, func(i, j int) bool {
			a, b := sa[i], sa[j]
			if rank[a] != rank[b] {
				return rank[a] < rank[b]
			}
			rankA := -1
			if a+k < n {
				rankA = rank[a+k]
			}
			rankB := -1
			if b+k < n {
				rankB = rank[b+k]
			}
			return rankA < rankB
		})

		newRank := make([]int, n)
		newRank[sa[0]] = 0
		for i := 1; i < n; i++ {
			prev, curr := sa[i-1], sa[i]
			same := rank[prev] == rank[curr]
			rankPrev := -1
			if prev+k < n {
				rankPrev = rank[prev+k]
			}
			rankCurr := -1
			if curr+k < n {
				rankCurr = rank[curr+k]
			}
			same = same && rankPrev == rankCurr
			if same {
				newRank[curr] = newRank[prev]
			} else {
				newRank[curr] = newRank[prev] + 1
			}
		}
		rank = newRank
	}

	lcp := computeLCP(text, sa, rank)

	return &SuffixArray{
		sa:   sa,
		rank: rank,
		lcp:  lcp,
		text: text,
	}
}

func (sa *SuffixArray) SA() []int {
	result := make([]int, len(sa.sa))
	copy(result, sa.sa)
	return result
}

func (sa *SuffixArray) Rank() []int {
	result := make([]int, len(sa.rank))
	copy(result, sa.rank)
	return result
}

func (sa *SuffixArray) LCP() []int {
	result := make([]int, len(sa.lcp))
	copy(result, sa.lcp)
	return result
}

func (sa *SuffixArray) Text() string {
	return sa.text
}
