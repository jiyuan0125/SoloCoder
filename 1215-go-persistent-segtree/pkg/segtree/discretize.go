package segtree

import (
	"sort"
)

type Discretizer struct {
	values  []int
	rankMap map[int]int
}

func NewDiscretizer(values []int) *Discretizer {
	d := &Discretizer{}
	d.build(values)
	return d
}

func (d *Discretizer) build(values []int) {
	if len(values) == 0 {
		d.values = []int{}
		d.rankMap = make(map[int]int)
		return
	}

	sorted := make([]int, len(values))
	copy(sorted, values)
	sort.Ints(sorted)

	unique := make([]int, 0, len(sorted))
	prev := sorted[0]
	unique = append(unique, prev)
	for i := 1; i < len(sorted); i++ {
		if sorted[i] != prev {
			prev = sorted[i]
			unique = append(unique, prev)
		}
	}

	d.values = unique
	d.rankMap = make(map[int]int, len(unique))
	for i, v := range unique {
		d.rankMap[v] = i + 1
	}
}

func (d *Discretizer) Rank(value int) int {
	return d.rankMap[value]
}

func (d *Discretizer) Value(rank int) int {
	if rank < 1 || rank > len(d.values) {
		return 0
	}
	return d.values[rank-1]
}

func (d *Discretizer) MaxRank() int {
	return len(d.values)
}

func (d *Discretizer) Empty() bool {
	return len(d.values) == 0
}
