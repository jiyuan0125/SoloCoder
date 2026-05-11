package core

import "sort"

type ItemSet map[string]struct{}

func NewItemSet(items []string) ItemSet {
	s := make(ItemSet)
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func (s ItemSet) Add(item string) {
	s[item] = struct{}{}
}

func (s ItemSet) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s ItemSet) ContainsAll(other ItemSet) bool {
	for item := range other {
		if !s.Contains(item) {
			return false
		}
	}
	return true
}

func (s ItemSet) Intersection(other ItemSet) ItemSet {
	result := make(ItemSet)
	for item := range s {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

func (s ItemSet) Union(other ItemSet) ItemSet {
	result := make(ItemSet)
	for item := range s {
		result.Add(item)
	}
	for item := range other {
		result.Add(item)
	}
	return result
}

func (s ItemSet) Difference(other ItemSet) ItemSet {
	result := make(ItemSet)
	for item := range s {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

func (s ItemSet) Size() int {
	return len(s)
}

func (s ItemSet) Items() []string {
	result := make([]string, 0, len(s))
	for item := range s {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func (s ItemSet) String() string {
	items := s.Items()
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}

func (s ItemSet) Copy() ItemSet {
	result := make(ItemSet)
	for item := range s {
		result.Add(item)
	}
	return result
}

type FrequentItemSet struct {
	Items          ItemSet
	WeightedSupport float64
	RawCount       int
}

type AssociationRule struct {
	Antecedent      ItemSet
	Consequent      ItemSet
	Confidence      float64
	WeightedSupport float64
}

type MiningStats struct {
	ScansCount      int
	CandidatesCount int
	FrequentCount   int
}

type MiningResult struct {
	FrequentItemSets []*FrequentItemSet
	Rules            []*AssociationRule
	Stats            *MiningStats
}
