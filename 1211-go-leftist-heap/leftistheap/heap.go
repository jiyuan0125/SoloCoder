package leftistheap

import "errors"

type HeapType int

const (
	MinHeap HeapType = iota
	MaxHeap
)

type Comparator func(a, b int64) bool

type Node struct {
	value int64
	dist  int
	left  *Node
	right *Node
}

type Heap struct {
	root       *Node
	comparator Comparator
	heapType   HeapType
}

var ErrEmptyHeap = errors.New("heap is empty")

func defaultMinComparator(a, b int64) bool {
	return a < b
}

func defaultMaxComparator(a, b int64) bool {
	return a > b
}

func NewHeap(heapType HeapType) *Heap {
	var comparator Comparator
	if heapType == MinHeap {
		comparator = defaultMinComparator
	} else {
		comparator = defaultMaxComparator
	}
	return &Heap{
		root:       nil,
		comparator: comparator,
		heapType:   heapType,
	}
}

func NewHeapWithComparator(comparator Comparator) *Heap {
	return &Heap{
		root:       nil,
		comparator: comparator,
	}
}

func (h *Heap) IsEmpty() bool {
	return h.root == nil
}

func (h *Heap) Insert(value int64) {
	newNode := &Node{
		value: value,
		dist:  0,
		left:  nil,
		right: nil,
	}
	newHeap := &Heap{
		root:       newNode,
		comparator: h.comparator,
		heapType:   h.heapType,
	}
	merged := h.merge(h.root, newHeap.root)
	h.root = merged
}

func (h *Heap) Top() (int64, error) {
	if h.IsEmpty() {
		return 0, ErrEmptyHeap
	}
	return h.root.value, nil
}

func (h *Heap) Pop() (int64, error) {
	if h.IsEmpty() {
		return 0, ErrEmptyHeap
	}
	value := h.root.value
	left := h.root.left
	right := h.root.right
	h.root = h.merge(left, right)
	return value, nil
}

func (h *Heap) Merge(other *Heap) {
	if h.comparator == nil {
		h.comparator = other.comparator
		h.heapType = other.heapType
	}
	h.root = h.merge(h.root, other.root)
	other.root = nil
}

func (h *Heap) MergeHeaps(h1, h2 *Heap) *Heap {
	result := NewHeapWithComparator(h.comparator)
	result.root = h.merge(h1.root, h2.root)
	h1.root = nil
	h2.root = nil
	return result
}

func (h *Heap) merge(node1, node2 *Node) *Node {
	if node1 == nil {
		return node2
	}
	if node2 == nil {
		return node1
	}

	if !h.comparator(node1.value, node2.value) {
		node1, node2 = node2, node1
	}

	node1.right = h.merge(node1.right, node2)

	if node1.left == nil || (node1.right != nil && node1.left.dist < node1.right.dist) {
		node1.left, node1.right = node1.right, node1.left
	}

	if node1.right == nil {
		node1.dist = 0
	} else {
		node1.dist = node1.right.dist + 1
	}

	return node1
}

func (h *Heap) Size() int {
	return countNodes(h.root)
}

func countNodes(node *Node) int {
	if node == nil {
		return 0
	}
	return 1 + countNodes(node.left) + countNodes(node.right)
}
