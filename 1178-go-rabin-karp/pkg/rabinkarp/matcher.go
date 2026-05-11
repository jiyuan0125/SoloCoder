package rabinkarp

import (
	"math/big"
	"sync"
)

const (
	baseConst    = 911382629
	modulusConst = 18446744073709551557
)

var (
	baseBig    = new(big.Int).SetUint64(baseConst)
	modulusBig = new(big.Int).SetUint64(modulusConst)
)

type MatchResult struct {
	Pattern string
	Index   int
}

type CollisionStats struct {
	TotalChecks    int
	FalsePositives int
	HashBuckets    map[uint64]int
}

type PatternGroup struct {
	patterns  map[string]bool
	hashToStr map[uint64][]string
	powBase   uint64
}

type Matcher struct {
	mu         sync.RWMutex
	byLen      map[int]*PatternGroup
	lengths    []int
	stats      CollisionStats
}

func mulModBig(a, b uint64) uint64 {
	ab := new(big.Int).Mul(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	ab.Mod(ab, modulusBig)
	return ab.Uint64()
}

func addModBig(a, b uint64) uint64 {
	sum := new(big.Int).Add(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	sum.Mod(sum, modulusBig)
	return sum.Uint64()
}

func subModBig(a, b uint64) uint64 {
	diff := new(big.Int).Sub(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	diff.Mod(diff, modulusBig)
	return diff.Uint64()
}

func powModBig(exp uint64) uint64 {
	result := new(big.Int).Exp(baseBig, new(big.Int).SetUint64(exp), modulusBig)
	return result.Uint64()
}

func hashString(s string) uint64 {
	h := uint64(0)
	for i := 0; i < len(s); i++ {
		h = addModBig(mulModBig(h, baseConst), uint64(s[i]))
	}
	return h
}

func NewMatcher() *Matcher {
	return &Matcher{
		byLen: make(map[int]*PatternGroup),
		stats: CollisionStats{
			HashBuckets: make(map[uint64]int),
		},
	}
}

func (m *Matcher) ensureGroup(length int) *PatternGroup {
	if g, ok := m.byLen[length]; ok {
		return g
	}
	g := &PatternGroup{
		patterns:  make(map[string]bool),
		hashToStr: make(map[uint64][]string),
		powBase:   powModBig(uint64(length - 1)),
	}
	m.byLen[length] = g
	m.lengths = append(m.lengths, length)
	return g
}

func (m *Matcher) Add(pattern string) bool {
	if len(pattern) == 0 {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	length := len(pattern)
	group := m.ensureGroup(length)

	if group.patterns[pattern] {
		return false
	}

	group.patterns[pattern] = true
	h := hashString(pattern)
	group.hashToStr[h] = append(group.hashToStr[h], pattern)
	if len(group.hashToStr[h]) > 1 {
		m.stats.HashBuckets[h] = len(group.hashToStr[h])
	}

	return true
}

func (m *Matcher) AddBatch(patterns []string) (added, skipped int) {
	for _, p := range patterns {
		if m.Add(p) {
			added++
		} else {
			skipped++
		}
	}
	return added, skipped
}

func (m *Matcher) Remove(pattern string) bool {
	if len(pattern) == 0 {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	length := len(pattern)
	group, ok := m.byLen[length]
	if !ok {
		return false
	}

	if !group.patterns[pattern] {
		return false
	}

	delete(group.patterns, pattern)
	h := hashString(pattern)
	if list, ok := group.hashToStr[h]; ok {
		for i, s := range list {
			if s == pattern {
				group.hashToStr[h] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(group.hashToStr[h]) == 0 {
			delete(group.hashToStr, h)
			delete(m.stats.HashBuckets, h)
		}
	}

	if len(group.patterns) == 0 {
		delete(m.byLen, length)
		for i, l := range m.lengths {
			if l == length {
				m.lengths = append(m.lengths[:i], m.lengths[i+1:]...)
				break
			}
		}
	}

	return true
}

func (m *Matcher) Patterns() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []string
	for _, group := range m.byLen {
		for p := range group.patterns {
			result = append(result, p)
		}
	}
	return result
}

func (m *Matcher) PatternsByLength() map[int][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[int][]string)
	for length, group := range m.byLen {
		list := make([]string, 0, len(group.patterns))
		for p := range group.patterns {
			list = append(list, p)
		}
		result[length] = list
	}
	return result
}

func (m *Matcher) CollisionStats() CollisionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	buckets := make(map[uint64]int, len(m.stats.HashBuckets))
	for k, v := range m.stats.HashBuckets {
		buckets[k] = v
	}

	return CollisionStats{
		TotalChecks:    m.stats.TotalChecks,
		FalsePositives: m.stats.FalsePositives,
		HashBuckets:    buckets,
	}
}

func (m *Matcher) Search(text string) []MatchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []MatchResult
	n := len(text)

	for _, length := range m.lengths {
		if length > n {
			continue
		}

		group := m.byLen[length]
		if len(group.patterns) == 0 {
			continue
		}

		matches := m.searchInGroup(text, length, group)
		results = append(results, matches...)
	}

	return results
}

func (m *Matcher) searchInGroup(text string, length int, group *PatternGroup) []MatchResult {
	var results []MatchResult
	n := len(text)

	currentHash := hashString(text[:length])

	for i := 0; i <= n-length; i++ {
		if candidates, ok := group.hashToStr[currentHash]; ok {
			window := text[i : i+length]
			m.stats.TotalChecks++

			found := false
			for _, p := range candidates {
				if p == window {
					results = append(results, MatchResult{
						Pattern: p,
						Index:   i,
					})
					found = true
				}
			}

			if !found {
				m.stats.FalsePositives++
			}
		}

		if i < n-length {
			currentHash = addModBig(
				mulModBig(
					subModBig(currentHash, mulModBig(uint64(text[i]), group.powBase)),
					baseConst,
				),
				uint64(text[i+length]),
			)
		}
	}

	return results
}

func Hash(s string) uint64 {
	return hashString(s)
}

func HashWithLen(s string, length int) (uint64, bool) {
	if length <= 0 || length > len(s) {
		return 0, false
	}
	return hashString(s[:length]), true
}
