package externalsort

import "container/heap"

type heapItem struct {
	value   int64
	chunkID int
	index   int
}

type minHeap []*heapItem

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].value < h[j].value }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }

func (h *minHeap) Push(x interface{}) {
	item := x.(*heapItem)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

func (h *minHeap) PushItem(value int64, chunkID int) {
	heap.Push(h, &heapItem{value: value, chunkID: chunkID})
}

func (h *minHeap) PopItem() (int64, int, bool) {
	if h.Len() == 0 {
		return 0, -1, false
	}
	item := heap.Pop(h).(*heapItem)
	return item.value, item.chunkID, true
}

func newMinHeap() *minHeap {
	h := &minHeap{}
	heap.Init(h)
	return h
}
