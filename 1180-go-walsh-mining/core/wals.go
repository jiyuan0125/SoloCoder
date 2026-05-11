package core

import (
	"errors"
	"math"
)

const (
	DefaultMaxCandidateCount = 100000
)

type WALSConfig struct {
	MaxCandidateCount int
}

func DefaultConfig() *WALSConfig {
	return &WALSConfig{
		MaxCandidateCount: DefaultMaxCandidateCount,
	}
}

type WALSMiner struct {
	transactions []ItemSet
	weights      map[string]float64
	config       *WALSConfig
}

func NewWALSMiner(transactions []ItemSet, weights map[string]float64) *WALSMiner {
	return NewWALSMinerWithConfig(transactions, weights, DefaultConfig())
}

func NewWALSMinerWithConfig(transactions []ItemSet, weights map[string]float64, config *WALSConfig) *WALSMiner {
	if config == nil {
		config = DefaultConfig()
	}
	if weights == nil {
		weights = make(map[string]float64)
	}
	return &WALSMiner{
		transactions: transactions,
		weights:      weights,
		config:       config,
	}
}

func (m *WALSMiner) getWeight(item string) float64 {
	if w, ok := m.weights[item]; ok {
		return w
	}
	return 1.0
}

func (m *WALSMiner) computeContribution(transaction ItemSet, itemset ItemSet) float64 {
	minWeight := 1.0
	for item := range itemset {
		w := m.getWeight(item)
		if w < minWeight {
			minWeight = w
		}
	}
	return minWeight
}

func (m *WALSMiner) getItemWeightedSupport(item string, totalTransactions int) (float64, int, bool) {
	if totalTransactions == 0 {
		return 0, 0, false
	}
	var totalContribution float64
	var rawCount int
	for _, t := range m.transactions {
		if t.Contains(item) {
			totalContribution += m.computeContribution(t, NewItemSet([]string{item}))
			rawCount++
		}
	}
	return totalContribution / float64(totalTransactions), rawCount, true
}

func (m *WALSMiner) getItemSetWeightedSupport(itemset ItemSet, totalTransactions int) (float64, int, bool) {
	if totalTransactions == 0 {
		return 0, 0, false
	}
	var totalContribution float64
	var rawCount int
	for _, t := range m.transactions {
		if t.ContainsAll(itemset) {
			totalContribution += m.computeContribution(t, itemset)
			rawCount++
		}
	}
	return totalContribution / float64(totalTransactions), rawCount, true
}

func (m *WALSMiner) generateCandidates(prevFrequent []*FrequentItemSet, k int) []ItemSet {
	var candidates []ItemSet
	prevSets := make([]ItemSet, 0, len(prevFrequent))
	for _, fis := range prevFrequent {
		prevSets = append(prevSets, fis.Items)
	}

	for i := 0; i < len(prevSets); i++ {
		for j := i + 1; j < len(prevSets); j++ {
			set1 := prevSets[i]
			set2 := prevSets[j]
			items1 := set1.Items()
			items2 := set2.Items()

			canMerge := true
			for idx := 0; idx < k-2; idx++ {
				if items1[idx] != items2[idx] {
					canMerge = false
					break
				}
			}
			if !canMerge {
				continue
			}

			merged := set1.Union(set2)
			if merged.Size() != k {
				continue
			}

			hasPruned := false
			for _, subset := range generateAllSubsetsExceptFull(merged, k-1) {
				found := false
				for _, s := range prevSets {
					if equalItemSets(s, subset) {
						found = true
						break
					}
				}
				if !found {
					hasPruned = true
					break
				}
			}
			if !hasPruned {
				candidates = append(candidates, merged)
			}
		}
	}
	return candidates
}

func (m *WALSMiner) Mine(minSup float64, minConf float64) (*MiningResult, error) {
	if minSup < 0 || minSup > 1 {
		return nil, errors.New("minSup must be between 0 and 1 (inclusive)")
	}
	if minConf < 0 || minConf > 1 {
		return nil, errors.New("minConf must be between 0 and 1 (inclusive)")
	}

	totalTransactions := len(m.transactions)

	if totalTransactions == 0 {
		return &MiningResult{
			FrequentItemSets: []*FrequentItemSet{},
			Rules:            []*AssociationRule{},
			Stats: &MiningStats{
				ScansCount:      0,
				CandidatesCount: 0,
				FrequentCount:   0,
			},
		}, nil
	}

	stats := &MiningStats{}
	allFrequent := make([]*FrequentItemSet, 0)

	items := m.collectUniqueItems()

	scansCount := 1
	candidatesCount := len(items)
	stats.ScansCount = scansCount

	var k1Frequent []*FrequentItemSet
	for _, item := range items {
		wSup, rawCount, _ := m.getItemWeightedSupport(item, totalTransactions)
		if wSup >= minSup {
			k1Frequent = append(k1Frequent, &FrequentItemSet{
				Items:           NewItemSet([]string{item}),
				WeightedSupport: wSup,
				RawCount:        rawCount,
			})
		}
	}

	allFrequent = append(allFrequent, k1Frequent...)
	prevFrequent := k1Frequent
	stats.FrequentCount = len(k1Frequent)

	for k := 2; len(prevFrequent) > 0; k++ {
		scansCount++
		stats.ScansCount = scansCount

		candidates := m.generateCandidates(prevFrequent, k)
		candidatesCount += len(candidates)
		stats.CandidatesCount = candidatesCount

		if candidatesCount > m.config.MaxCandidateCount && minSup == 0 {
			break
		}

		var currFrequent []*FrequentItemSet
		for _, candidate := range candidates {
			wSup, rawCount, _ := m.getItemSetWeightedSupport(candidate, totalTransactions)
			if wSup >= minSup {
				currFrequent = append(currFrequent, &FrequentItemSet{
					Items:           candidate,
					WeightedSupport: wSup,
					RawCount:        rawCount,
				})
			}
		}

		allFrequent = append(allFrequent, currFrequent...)
		stats.FrequentCount = len(allFrequent)
		prevFrequent = currFrequent
	}

	rules := m.generateRules(allFrequent, minConf)

	return &MiningResult{
		FrequentItemSets: allFrequent,
		Rules:            rules,
		Stats:            stats,
	}, nil
}

func (m *WALSMiner) collectUniqueItems() []string {
	itemSet := make(map[string]struct{})
	for _, t := range m.transactions {
		if t.Size() == 0 {
			continue
		}
		for item := range t {
			itemSet[item] = struct{}{}
		}
	}
	items := make([]string, 0, len(itemSet))
	for item := range itemSet {
		items = append(items, item)
	}
	return items
}

func (m *WALSMiner) generateRules(frequent []*FrequentItemSet, minConf float64) []*AssociationRule {
	var rules []*AssociationRule
	supportMap := make(map[string]float64)
	for _, fis := range frequent {
		key := fis.Items.String()
		supportMap[key] = fis.WeightedSupport
	}

	for _, fis := range frequent {
		if fis.Items.Size() < 2 {
			continue
		}

		subsets := generateAllNonEmptyProperSubsets(fis.Items)
		for _, antecedent := range subsets {
			consequent := fis.Items.Difference(antecedent)
			if consequent.Size() == 0 {
				continue
			}

			union := antecedent.Union(consequent)
			unionSup := supportMap[union.String()]

			antecedentSup, ok := supportMap[antecedent.String()]
			if !ok || antecedentSup == 0 {
				continue
			}

			confidence := unionSup / antecedentSup
			if confidence >= minConf {
				rules = append(rules, &AssociationRule{
					Antecedent:      antecedent,
					Consequent:      consequent,
					Confidence:      confidence,
					WeightedSupport: unionSup,
				})
			}
		}
	}
	return rules
}

func equalItemSets(a, b ItemSet) bool {
	if a.Size() != b.Size() {
		return false
	}
	for item := range a {
		if !b.Contains(item) {
			return false
		}
	}
	return true
}

func generateAllNonEmptyProperSubsets(s ItemSet) []ItemSet {
	var subsets []ItemSet
	items := s.Items()
	n := len(items)
	if n == 0 {
		return subsets
	}

	maxMask := 1 << uint(n)
	for mask := 1; mask < maxMask-1; mask++ {
		subset := make(ItemSet)
		for i := 0; i < n; i++ {
			if mask&(1<<uint(i)) != 0 {
				subset.Add(items[i])
			}
		}
		subsets = append(subsets, subset)
	}
	return subsets
}

func generateAllSubsetsExceptFull(s ItemSet, size int) []ItemSet {
	var subsets []ItemSet
	items := s.Items()
	n := len(items)
	if n < size {
		return subsets
	}

	maxMask := 1 << uint(n)
	for mask := 0; mask < maxMask; mask++ {
		count := 0
		for i := 0; i < n; i++ {
			if mask&(1<<uint(i)) != 0 {
				count++
			}
		}
		if count == size {
			subset := make(ItemSet)
			for i := 0; i < n; i++ {
				if mask&(1<<uint(i)) != 0 {
					subset.Add(items[i])
				}
			}
			subsets = append(subsets, subset)
		}
	}
	return subsets
}

func RoundTo(x float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(x*shift) / shift
}
