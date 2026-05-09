package rbtree

import (
	"errors"
)

type color bool

const (
	red   color = true
	black color = false
)

type Node struct {
	key    int
	color  color
	parent *Node
	left   *Node
	right  *Node
	size   int
}

type Tree struct {
	root *Node
	nil  *Node
	size int
}

func (t *Tree) nodeSize(n *Node) int {
	if n == t.nil {
		return 0
	}
	return n.size
}

func (t *Tree) updateSize(n *Node) {
	n.size = 1 + t.nodeSize(n.left) + t.nodeSize(n.right)
}

func New() *Tree {
	nilNode := &Node{color: black}
	nilNode.left = nilNode
	nilNode.right = nilNode
	nilNode.parent = nilNode
	return &Tree{
		root: nilNode,
		nil:  nilNode,
		size: 0,
	}
}

func (t *Tree) Size() int {
	return t.size
}

func (t *Tree) leftRotate(x *Node) {
	y := x.right
	x.right = y.left
	if y.left != t.nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == t.nil {
		t.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
	t.updateSize(x)
	t.updateSize(y)
}

func (t *Tree) rightRotate(y *Node) {
	x := y.left
	y.left = x.right
	if x.right != t.nil {
		x.right.parent = y
	}
	x.parent = y.parent
	if y.parent == t.nil {
		t.root = x
	} else if y == y.parent.right {
		y.parent.right = x
	} else {
		y.parent.left = x
	}
	x.right = y
	y.parent = x
	t.updateSize(y)
	t.updateSize(x)
}

func (t *Tree) insertFixup(z *Node) {
	for z.parent.color == red {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y.color == red {
				z.parent.color = black
				y.color = black
				z.parent.parent.color = red
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					t.leftRotate(z)
				}
				z.parent.color = black
				z.parent.parent.color = red
				t.rightRotate(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y.color == red {
				z.parent.color = black
				y.color = black
				z.parent.parent.color = red
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					t.rightRotate(z)
				}
				z.parent.color = black
				z.parent.parent.color = red
				t.leftRotate(z.parent.parent)
			}
		}
	}
	t.root.color = black
}

func (t *Tree) Insert(key int) {
	z := &Node{
		key:   key,
		color: red,
		left:  t.nil,
		right: t.nil,
		size:  1,
	}
	y := t.nil
	x := t.root
	for x != t.nil {
		y = x
		x.size++
		if z.key < x.key {
			x = x.left
		} else {
			x = x.right
		}
	}
	z.parent = y
	if y == t.nil {
		t.root = z
	} else if z.key < y.key {
		y.left = z
	} else {
		y.right = z
	}
	t.size++
	t.insertFixup(z)
}

func (t *Tree) transplant(u, v *Node) {
	if u.parent == t.nil {
		t.root = v
	} else if u == u.parent.left {
		u.parent.left = v
	} else {
		u.parent.right = v
	}
	v.parent = u.parent
}

func (t *Tree) minimum(x *Node) *Node {
	for x.left != t.nil {
		x = x.left
	}
	return x
}

func (t *Tree) deleteFixup(x *Node) {
	for x != t.root && x.color == black {
		if x == x.parent.left {
			w := x.parent.right
			if w.color == red {
				w.color = black
				x.parent.color = red
				t.leftRotate(x.parent)
				w = x.parent.right
			}
			if w.left.color == black && w.right.color == black {
				w.color = red
				x = x.parent
			} else {
				if w.right.color == black {
					w.left.color = black
					w.color = red
					t.rightRotate(w)
					w = x.parent.right
				}
				w.color = x.parent.color
				x.parent.color = black
				w.right.color = black
				t.leftRotate(x.parent)
				x = t.root
			}
		} else {
			w := x.parent.left
			if w.color == red {
				w.color = black
				x.parent.color = red
				t.rightRotate(x.parent)
				w = x.parent.left
			}
			if w.right.color == black && w.left.color == black {
				w.color = red
				x = x.parent
			} else {
				if w.left.color == black {
					w.right.color = black
					w.color = red
					t.leftRotate(w)
					w = x.parent.left
				}
				w.color = x.parent.color
				x.parent.color = black
				w.left.color = black
				t.rightRotate(x.parent)
				x = t.root
			}
		}
	}
	x.color = black
}

func (t *Tree) updateSizeUp(n *Node) {
	for n != t.nil {
		t.updateSize(n)
		n = n.parent
	}
}

func (t *Tree) Delete(key int) bool {
	z := t.root
	for z != t.nil {
		if key == z.key {
			break
		} else if key < z.key {
			z = z.left
		} else {
			z = z.right
		}
	}
	if z == t.nil {
		return false
	}

	y := z
	yOriginalColor := y.color
	var x *Node

	if z.left == t.nil {
		x = z.right
		t.transplant(z, z.right)
		t.updateSizeUp(x.parent)
	} else if z.right == t.nil {
		x = z.left
		t.transplant(z, z.left)
		t.updateSizeUp(x.parent)
	} else {
		y = t.minimum(z.right)
		yOriginalColor = y.color
		x = y.right
		if y.parent == z {
			x.parent = y
		} else {
			t.transplant(y, y.right)
			y.right = z.right
			y.right.parent = y
			t.updateSizeUp(x.parent)
		}
		t.transplant(z, y)
		y.left = z.left
		y.left.parent = y
		y.color = z.color
		t.updateSizeUp(y)
	}

	t.size--
	if yOriginalColor == black {
		t.deleteFixup(x)
	}
	return true
}

func (t *Tree) RangeQuery(low, high int) []int {
	if low > high {
		return []int{}
	}
	if t.size == 0 {
		return []int{}
	}
	result := []int{}
	t.rangeQueryHelper(t.root, low, high, &result)
	return result
}

func (t *Tree) rangeQueryHelper(n *Node, low, high int, result *[]int) {
	if n == t.nil {
		return
	}
	if n.key >= low {
		t.rangeQueryHelper(n.left, low, high, result)
	}
	if n.key >= low && n.key <= high {
		*result = append(*result, n.key)
	}
	if n.key <= high {
		t.rangeQueryHelper(n.right, low, high, result)
	}
}

func (t *Tree) KthLargest(k int) (int, error) {
	if t.size == 0 {
		return 0, errors.New("tree is empty")
	}
	if k < 1 || k > t.size {
		return 0, errors.New("k is out of range")
	}
	target := t.size - k + 1
	n := t.root
	for n != t.nil {
		leftSize := t.nodeSize(n.left)
		if target == leftSize+1 {
			return n.key, nil
		} else if target <= leftSize {
			n = n.left
		} else {
			target -= leftSize + 1
			n = n.right
		}
	}
	return 0, errors.New("element not found")
}

func (t *Tree) validateRBProperties() bool {
	if t.root.color == red {
		return false
	}
	return t.validateRBPropertiesHelper(t.root, make(map[*Node]bool), 0, -1)
}

func (t *Tree) validateRBPropertiesHelper(n *Node, visited map[*Node]bool, pathBlackCount int, leafBlackCount int) bool {
	if n == t.nil {
		if n.color != black {
			return false
		}
		if leafBlackCount == -1 {
			leafBlackCount = pathBlackCount
		} else if pathBlackCount != leafBlackCount {
			return false
		}
		return leafBlackCount != -1
	}

	if visited[n] {
		return false
	}
	visited[n] = true

	if n.color == red {
		if n.left.color == red || n.right.color == red {
			return false
		}
	}

	currentPathBlackCount := pathBlackCount
	if n.color == black {
		currentPathBlackCount++
	}

	expectedSize := 1 + t.nodeSize(n.left) + t.nodeSize(n.right)
	if n.size != expectedSize {
		return false
	}

	leftResult := t.validateRBPropertiesHelper(n.left, visited, currentPathBlackCount, leafBlackCount)
	if !leftResult {
		return false
	}

	return t.validateRBPropertiesHelper(n.right, visited, currentPathBlackCount, leafBlackCount)
}
