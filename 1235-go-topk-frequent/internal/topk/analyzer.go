package topk

import (
	"container/heap"
	"sort"

	"topk-frequent/internal/sketch"
)

type TopKItem struct {
	Element   string
	Freq      uint64
	Uncertain bool
}

type Analyzer struct {
	sketch *sketch.CountMinSketch
	k      int
	h      *minHeap
	set    map[string]struct{}
}

func New(width, depth, k int) *Analyzer {
	if k <= 0 {
		k = 10
	}
	h := make(minHeap, 0, k)
	return &Analyzer{
		sketch: sketch.New(width, depth),
		k:      k,
		h:      &h,
		set:    make(map[string]struct{}),
	}
}

func (a *Analyzer) Add(item string) {
	a.sketch.Add(item)
	a.updateTopK(item)
}

func (a *Analyzer) AddBatch(items []string) {
	unique := make(map[string]uint64)
	for _, item := range items {
		unique[item]++
	}

	for item, n := range unique {
		a.sketch.AddN(item, n)
	}

	for item := range unique {
		a.updateTopK(item)
	}
}

func (a *Analyzer) updateTopK(item string) {
	freq := a.sketch.Count(item)

	if _, exists := a.set[item]; exists {
		a.updateExistingItem(item, freq)
		return
	}

	if a.h.Len() < a.k {
		heap.Push(a.h, Item{Element: item, Freq: freq})
		a.set[item] = struct{}{}
		return
	}

	if a.h.Len() > 0 {
		minItem := (*a.h)[0]
		if freq > minItem.Freq {
			delete(a.set, minItem.Element)
			heap.Pop(a.h)
			heap.Push(a.h, Item{Element: item, Freq: freq})
			a.set[item] = struct{}{}
		}
	}
}

func (a *Analyzer) updateExistingItem(item string, newFreq uint64) {
	for i := range *a.h {
		if (*a.h)[i].Element == item {
			(*a.h)[i].Freq = newFreq
			break
		}
	}
	heap.Init(a.h)
}

func (a *Analyzer) TopK() []TopKItem {
	return a.TopKWithK(a.k)
}

func (a *Analyzer) TopKWithK(k int) []TopKItem {
	if k <= 0 {
		k = a.k
	}

	h := make(minHeap, 0, k)
	for item := range a.set {
		freq := a.sketch.Count(item)
		if h.Len() < k {
			heap.Push(&h, Item{Element: item, Freq: freq})
		} else if h.Len() > 0 && freq > h[0].Freq {
			heap.Pop(&h)
			heap.Push(&h, Item{Element: item, Freq: freq})
		}
	}

	result := make([]TopKItem, 0, h.Len())
	for h.Len() > 0 {
		item := heap.Pop(&h).(Item)
		uncertain := a.isUncertain(item.Element)
		result = append(result, TopKItem{
			Element:   item.Element,
			Freq:      item.Freq,
			Uncertain: uncertain,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Freq != result[j].Freq {
			return result[i].Freq > result[j].Freq
		}
		return result[i].Element < result[j].Element
	})

	return result
}

func (a *Analyzer) isUncertain(item string) bool {
	counts := a.sketch.Counts(item)
	if len(counts) < 2 {
		return false
	}

	minCount := counts[0]
	for _, c := range counts {
		if c < minCount {
			minCount = c
		}
	}

	threshold := float64(minCount) * 0.5
	for _, c := range counts {
		if float64(c-minCount) > threshold {
			return true
		}
	}
	return false
}

func (a *Analyzer) Reset() {
	a.sketch.Reset()
	*a.h = make(minHeap, 0, a.k)
	a.set = make(map[string]struct{})
}

func (a *Analyzer) K() int {
	return a.k
}

func (a *Analyzer) Width() int {
	return a.sketch.Width()
}

func (a *Analyzer) Depth() int {
	return a.sketch.Depth()
}
