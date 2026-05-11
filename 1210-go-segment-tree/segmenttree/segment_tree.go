package segmenttree

type SegmentTree struct {
	n    int
	sum  []int
	max  []int
	lazy []int
}

func New(initValues []int) *SegmentTree {
	n := len(initValues)
	st := &SegmentTree{
		n:    n,
		sum:  make([]int, 4*n),
		max:  make([]int, 4*n),
		lazy: make([]int, 4*n),
	}
	if n > 0 {
		st.build(1, 0, n-1, initValues)
	}
	return st
}

func (st *SegmentTree) build(node, l, r int, values []int) {
	if l == r {
		st.sum[node] = values[l]
		st.max[node] = values[l]
		return
	}
	mid := (l + r) / 2
	st.build(2*node, l, mid, values)
	st.build(2*node+1, mid+1, r, values)
	st.sum[node] = st.sum[2*node] + st.sum[2*node+1]
	st.max[node] = maxInt(st.max[2*node], st.max[2*node+1])
}

func (st *SegmentTree) pushDown(node, l, r int) {
	if st.lazy[node] == 0 || l == r {
		return
	}
	val := st.lazy[node]
	left := 2 * node
	rightNode := 2*node + 1
	mid := (l + r) / 2

	st.sum[left] += val * (mid - l + 1)
	st.max[left] += val
	st.lazy[left] += val

	st.sum[rightNode] += val * (r - (mid + 1) + 1)
	st.max[rightNode] += val
	st.lazy[rightNode] += val

	st.lazy[node] = 0
}

func (st *SegmentTree) rangeAdd(node, l, r, ul, ur, val int) {
	if ur < l || ul > r {
		return
	}
	if ul <= l && r <= ur {
		st.sum[node] += val * (r - l + 1)
		st.max[node] += val
		st.lazy[node] += val
		return
	}
	st.pushDown(node, l, r)
	mid := (l + r) / 2
	st.rangeAdd(2*node, l, mid, ul, ur, val)
	st.rangeAdd(2*node+1, mid+1, r, ul, ur, val)
	st.sum[node] = st.sum[2*node] + st.sum[2*node+1]
	st.max[node] = maxInt(st.max[2*node], st.max[2*node+1])
}

func (st *SegmentTree) rangeSum(node, l, r, ql, qr int) int {
	if qr < l || ql > r {
		return 0
	}
	if ql <= l && r <= qr {
		return st.sum[node]
	}
	st.pushDown(node, l, r)
	mid := (l + r) / 2
	leftSum := st.rangeSum(2*node, l, mid, ql, qr)
	rightSum := st.rangeSum(2*node+1, mid+1, r, ql, qr)
	return leftSum + rightSum
}

func (st *SegmentTree) rangeMax(node, l, r, ql, qr int) (bool, int) {
	if qr < l || ql > r {
		return false, 0
	}
	if ql <= l && r <= qr {
		return true, st.max[node]
	}
	st.pushDown(node, l, r)
	mid := (l + r) / 2
	leftOk, leftMax := st.rangeMax(2*node, l, mid, ql, qr)
	rightOk, rightMax := st.rangeMax(2*node+1, mid+1, r, ql, qr)
	if leftOk && rightOk {
		return true, maxInt(leftMax, rightMax)
	} else if leftOk {
		return true, leftMax
	} else if rightOk {
		return true, rightMax
	}
	return false, 0
}

func (st *SegmentTree) normalize(left, right int) (int, int, bool) {
	if left > right {
		left, right = right, left
	}
	if st.n == 0 {
		return 0, -1, false
	}
	if left < 0 {
		left = 0
	}
	if right >= st.n {
		right = st.n - 1
	}
	if left > right {
		return 0, -1, false
	}
	return left, right, true
}

func (st *SegmentTree) RangeAdd(left, right, value int) {
	l, r, valid := st.normalize(left, right)
	if !valid {
		return
	}
	st.rangeAdd(1, 0, st.n-1, l, r, value)
}

func (st *SegmentTree) RangeSum(left, right int) int {
	l, r, valid := st.normalize(left, right)
	if !valid {
		return 0
	}
	return st.rangeSum(1, 0, st.n-1, l, r)
}

func (st *SegmentTree) RangeMax(left, right int) int {
	l, r, valid := st.normalize(left, right)
	if !valid {
		return 0
	}
	_, res := st.rangeMax(1, 0, st.n-1, l, r)
	return res
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
