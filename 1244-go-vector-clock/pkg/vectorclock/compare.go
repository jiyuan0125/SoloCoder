package vectorclock

import "strings"

func Compare(a, b VectorClock) CausalityType {
	if a.Equal(b) {
		return CausalitySame
	}
	if a.LessThan(b) {
		return CausalityBefore
	}
	if b.LessThan(a) {
		return CausalityAfter
	}
	return CausalityConcurrent
}

func (a VectorClock) Equal(b VectorClock) bool {
	allNodes := a.unionKeys(b)
	for node := range allNodes {
		av := a.getValue(node)
		bv := b.getValue(node)
		if av != bv {
			return false
		}
	}
	return true
}

func (a VectorClock) LessThan(b VectorClock) bool {
	allNodes := a.unionKeys(b)
	hasStrictlyLess := false
	for node := range allNodes {
		av := a.getValue(node)
		bv := b.getValue(node)
		if av > bv {
			return false
		}
		if av < bv {
			hasStrictlyLess = true
		}
	}
	return hasStrictlyLess
}

func (a VectorClock) getValue(node string) uint64 {
	if strings.HasPrefix(node, "__") {
		return 0
	}
	if val, exists := a[node]; exists {
		return val
	}
	if historical, exists := a[HistoricalKey]; exists {
		return historical
	}
	return 0
}

func (a VectorClock) unionKeys(b VectorClock) map[string]struct{} {
	result := make(map[string]struct{})
	for k := range a {
		if k == HistoricalKey {
			continue
		}
		result[k] = struct{}{}
	}
	for k := range b {
		if k == HistoricalKey {
			continue
		}
		result[k] = struct{}{}
	}
	return result
}
