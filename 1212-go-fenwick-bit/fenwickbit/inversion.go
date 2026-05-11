package fenwickbit

import "sort"

func CountInversions(arr []int64) int64 {
	n := len(arr)
	if n < 2 {
		return 0
	}
	
	ranks := discretize(arr)
	ft := New(len(ranks))
	
	var invCount int64
	for i := n - 1; i >= 0; i-- {
		rank := ranks[arr[i]]
		invCount += ft.QueryPrefix(rank - 1)
		ft.Update(rank, 1)
	}
	
	return invCount
}

func discretize(arr []int64) map[int64]int {
	sorted := make([]int64, len(arr))
	copy(sorted, arr)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	
	ranks := make(map[int64]int)
	rank := 1
	for i := 0; i < len(sorted); i++ {
		if i == 0 || sorted[i] != sorted[i-1] {
			ranks[sorted[i]] = rank
			rank++
		}
	}
	
	return ranks
}
