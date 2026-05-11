package suffixarray

func computeLCP(text string, sa []int, rank []int) []int {
	n := len(text)
	if n <= 1 {
		return []int{}
	}

	lcp := make([]int, n-1)
	k := 0

	for i := 0; i < n; i++ {
		if rank[i] == n-1 {
			k = 0
			continue
		}
		j := sa[rank[i]+1]
		for i+k < n && j+k < n && text[i+k] == text[j+k] {
			k++
		}
		lcp[rank[i]] = k
		if k > 0 {
			k--
		}
	}

	return lcp
}
