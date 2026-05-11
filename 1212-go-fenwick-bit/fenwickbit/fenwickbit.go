package fenwickbit

import "errors"

var (
	ErrInvalidIndex = errors.New("invalid index")
	ErrInvalidRange = errors.New("invalid range")
)

type FenwickTree struct {
	tree []int64
	size int
}

func New(size int) *FenwickTree {
	return &FenwickTree{
		tree: make([]int64, size+1),
		size: size,
	}
}

func (ft *FenwickTree) Size() int {
	return ft.size
}

func (ft *FenwickTree) Update(index int, delta int64) error {
	if index < 1 || index > ft.size {
		return ErrInvalidIndex
	}
	for index <= ft.size {
		ft.tree[index] += delta
		index += index & -index
	}
	return nil
}

func (ft *FenwickTree) QueryPrefix(index int) int64 {
	if index < 1 || index > ft.size {
		return 0
	}
	var sum int64
	for index > 0 {
		sum += ft.tree[index]
		index -= index & -index
	}
	return sum
}

func (ft *FenwickTree) QueryRange(l, r int) (int64, error) {
	if l < 1 || r < 1 || l > ft.size || r > ft.size || l > r {
		return 0, ErrInvalidRange
	}
	return ft.QueryPrefix(r) - ft.QueryPrefix(l-1), nil
}
