package topk

type elementWithFreq struct {
	element float64
	freq    int
}

type maxHeap []elementWithFreq

func (h maxHeap) Len() int { return len(h) }

func (h maxHeap) Less(i, j int) bool {
	if h[i].freq == h[j].freq {
		return h[i].element > h[j].element
	}
	return h[i].freq > h[j].freq
}

func (h maxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *maxHeap) Push(x interface{}) {
	*h = append(*h, x.(elementWithFreq))
}

func (h *maxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
