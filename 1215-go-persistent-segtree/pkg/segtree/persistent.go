package segtree

import (
	"errors"
)

var (
	ErrEmptySegmentTree = errors.New("segment tree is empty")
	ErrInvalidRange     = errors.New("invalid range: l > r")
	ErrInvalidK         = errors.New("k is out of range")
	ErrInvalidVersion   = errors.New("invalid version")
	ErrEmptyArray       = errors.New("array is empty")
)

type node struct {
	left, right int
	count       int
}

type PersistentSegmentTree struct {
	nodes      []node
	roots      []int
	discretizer *Discretizer
	maxRank    int
}

func New() *PersistentSegmentTree {
	return &PersistentSegmentTree{
		nodes:   []node{{left: -1, right: -1, count: 0}},
		roots:   []int{0},
		maxRank: 0,
	}
}

func (p *PersistentSegmentTree) Build(values []int) error {
	if len(values) == 0 {
		return ErrEmptyArray
	}

	discretizer := NewDiscretizer(values)
	p.discretizer = discretizer
	p.maxRank = discretizer.MaxRank()
	p.nodes = []node{{left: -1, right: -1, count: 0}}
	p.roots = []int{0}

	for _, v := range values {
		rank := discretizer.Rank(v)
		newRoot := p.update(p.roots[len(p.roots)-1], 1, p.maxRank, rank, 1)
		p.roots = append(p.roots, newRoot)
	}

	return nil
}

func (p *PersistentSegmentTree) CreateEmptyVersion() int {
	return len(p.roots) - 1
}

func (p *PersistentSegmentTree) Insert(version int, value int) (int, error) {
	if p.discretizer == nil {
		return 0, ErrEmptySegmentTree
	}
	if version < 0 || version >= len(p.roots) {
		return 0, ErrInvalidVersion
	}

	rank := p.discretizer.Rank(value)
	if rank == 0 {
		return 0, errors.New("value not in original array")
	}

	newRoot := p.update(p.roots[version], 1, p.maxRank, rank, 1)
	p.roots = append(p.roots, newRoot)
	return len(p.roots) - 1, nil
}

func (p *PersistentSegmentTree) update(root int, l, r, pos, delta int) int {
	newNodeIdx := len(p.nodes)
	newNode := p.nodes[root]
	newNode.count += delta
	p.nodes = append(p.nodes, newNode)

	if l == r {
		return newNodeIdx
	}

	mid := (l + r) >> 1
	if pos <= mid {
		leftChild := p.nodes[root].left
		if leftChild == -1 {
			p.nodes[newNodeIdx].left = p.update(0, l, mid, pos, delta)
		} else {
			p.nodes[newNodeIdx].left = p.update(leftChild, l, mid, pos, delta)
		}
	} else {
		rightChild := p.nodes[root].right
		if rightChild == -1 {
			p.nodes[newNodeIdx].right = p.update(0, mid+1, r, pos, delta)
		} else {
			p.nodes[newNodeIdx].right = p.update(rightChild, mid+1, r, pos, delta)
		}
	}

	return newNodeIdx
}

func (p *PersistentSegmentTree) QueryKth(l, r, k int) (int, error) {
	if p.discretizer == nil {
		return 0, ErrEmptySegmentTree
	}
	if l < 1 || r > len(p.roots)-1 {
		return 0, ErrInvalidRange
	}
	if l > r {
		return 0, ErrInvalidRange
	}
	if k < 1 {
		return 0, ErrInvalidK
	}

	total := r - l + 1
	if k > total {
		return 0, ErrInvalidK
	}

	rank := p.queryKth(p.roots[l-1], p.roots[r], 1, p.maxRank, k)
	return p.discretizer.Value(rank), nil
}

func (p *PersistentSegmentTree) queryKth(rootL, rootR int, l, r, k int) int {
	if l == r {
		return l
	}

	mid := (l + r) >> 1
	
	leftCount := 0
	if p.nodes[rootR].left != -1 {
		leftCount += p.nodes[p.nodes[rootR].left].count
	}
	if p.nodes[rootL].left != -1 {
		leftCount -= p.nodes[p.nodes[rootL].left].count
	}

	if k <= leftCount {
		leftL := p.nodes[rootL].left
		leftR := p.nodes[rootR].left
		if leftL == -1 {
			leftL = 0
		}
		if leftR == -1 {
			leftR = 0
		}
		return p.queryKth(leftL, leftR, l, mid, k)
	} else {
		rightL := p.nodes[rootL].right
		rightR := p.nodes[rootR].right
		if rightL == -1 {
			rightL = 0
		}
		if rightR == -1 {
			rightR = 0
		}
		return p.queryKth(rightL, rightR, mid+1, r, k-leftCount)
	}
}

func (p *PersistentSegmentTree) PointQuery(version int, value int) (int, error) {
	if p.discretizer == nil {
		return 0, ErrEmptySegmentTree
	}
	if version < 0 || version >= len(p.roots) {
		return 0, ErrInvalidVersion
	}

	rank := p.discretizer.Rank(value)
	if rank == 0 {
		return 0, errors.New("value not in original array")
	}

	return p.pointQuery(p.roots[version], 1, p.maxRank, rank), nil
}

func (p *PersistentSegmentTree) pointQuery(root int, l, r, pos int) int {
	if root == -1 {
		return 0
	}
	if l == r {
		return p.nodes[root].count
	}

	mid := (l + r) >> 1
	if pos <= mid {
		return p.pointQuery(p.nodes[root].left, l, mid, pos)
	}
	return p.pointQuery(p.nodes[root].right, mid+1, r, pos)
}

func (p *PersistentSegmentTree) RangeSum(version int, ql, qr int) (int, error) {
	if p.discretizer == nil {
		return 0, ErrEmptySegmentTree
	}
	if version < 0 || version >= len(p.roots) {
		return 0, ErrInvalidVersion
	}
	if ql < 1 || qr > p.maxRank {
		return 0, ErrInvalidRange
	}
	if ql > qr {
		return 0, ErrInvalidRange
	}

	return p.rangeSum(p.roots[version], 1, p.maxRank, ql, qr), nil
}

func (p *PersistentSegmentTree) rangeSum(root int, l, r, ql, qr int) int {
	if root == -1 {
		return 0
	}
	if qr < l || ql > r {
		return 0
	}
	if ql <= l && r <= qr {
		return p.nodes[root].count
	}

	mid := (l + r) >> 1
	return p.rangeSum(p.nodes[root].left, l, mid, ql, qr) +
		p.rangeSum(p.nodes[root].right, mid+1, r, ql, qr)
}

func (p *PersistentSegmentTree) VersionCount() int {
	return len(p.roots)
}

func (p *PersistentSegmentTree) MaxRank() int {
	return p.maxRank
}
