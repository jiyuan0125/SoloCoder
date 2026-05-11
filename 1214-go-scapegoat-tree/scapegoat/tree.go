package scapegoat

import (
	"errors"
	"math"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrEmptyTree   = errors.New("tree is empty")
)

const (
	alpha         = 0.7
	deleteRatio   = 0.5
)

type node struct {
	key       int64
	value     int64
	deleted   bool
	left      *node
	right     *node
}

type Tree struct {
	root        *node
	size        int
	totalNodes  int
}

func NewTree() *Tree {
	return &Tree{}
}

func (t *Tree) Size() int {
	return t.size
}

func (t *Tree) Insert(key int64, value int64) {
	path := []*node{}
	t.root = t.insert(t.root, key, value, &path)
	
	if len(path) == 0 {
		return
	}
	
	for i := len(path) - 1; i >= 0; i-- {
		scg := path[i]
		if t.isUnbalanced(scg) {
			t.rebuildSubtree(scg)
			break
		}
	}
	
	if t.shouldFullRebuild() {
		t.fullRebuild()
	}
}

func (t *Tree) insert(n *node, key int64, value int64, path *[]*node) *node {
	if n == nil {
		t.size++
		t.totalNodes++
		return &node{key: key, value: value}
	}
	
	*path = append(*path, n)
	
	if key == n.key {
		if n.deleted {
			n.deleted = false
			n.value = value
			t.size++
		} else {
			n.value = value
		}
		return n
	}
	
	if key < n.key {
		n.left = t.insert(n.left, key, value, path)
	} else {
		n.right = t.insert(n.right, key, value, path)
	}
	
	return n
}

func (t *Tree) Search(key int64) (int64, error) {
	n := t.root
	for n != nil {
		if key == n.key {
			if n.deleted {
				return 0, ErrKeyNotFound
			}
			return n.value, nil
		}
		if key < n.key {
			n = n.left
		} else {
			n = n.right
		}
	}
	return 0, ErrKeyNotFound
}

func (t *Tree) Delete(key int64) error {
	if t.root == nil {
		return ErrEmptyTree
	}
	
	n := t.root
	for n != nil {
		if key == n.key {
			if n.deleted {
				return ErrKeyNotFound
			}
			n.deleted = true
			t.size--
			
			if t.shouldFullRebuild() {
				t.fullRebuild()
			}
			return nil
		}
		if key < n.key {
			n = n.left
		} else {
			n = n.right
		}
	}
	
	return ErrKeyNotFound
}

func (t *Tree) isUnbalanced(n *node) bool {
	total := t.countAll(n)
	if total <= 2 {
		return false
	}
	leftSize := t.countAll(n.left)
	rightSize := t.countAll(n.right)
	
	maxSize := math.Max(float64(leftSize), float64(rightSize))
	return maxSize > alpha*float64(total)
}

func (t *Tree) countAll(n *node) int {
	if n == nil {
		return 0
	}
	return 1 + t.countAll(n.left) + t.countAll(n.right)
}

func (t *Tree) shouldFullRebuild() bool {
	if t.totalNodes == 0 {
		return false
	}
	deletedCount := t.totalNodes - t.size
	return float64(deletedCount) > deleteRatio*float64(t.totalNodes)
}

func (t *Tree) rebuildSubtree(scg *node) {
	nodes := []*node{}
	t.inorder(scg, &nodes)
	
	if len(nodes) <= 2 {
		return
	}
	
	newRoot := t.buildBalanced(nodes, 0, len(nodes)-1)
	
	if scg == t.root {
		t.root = newRoot
	} else {
		parent := t.findParent(t.root, scg)
		if parent != nil {
			if parent.left == scg {
				parent.left = newRoot
			} else {
				parent.right = newRoot
			}
		}
	}
}

func (t *Tree) inorder(n *node, nodes *[]*node) {
	if n == nil {
		return
	}
	t.inorder(n.left, nodes)
	*nodes = append(*nodes, n)
	t.inorder(n.right, nodes)
}

func (t *Tree) buildBalanced(nodes []*node, start, end int) *node {
	if start > end {
		return nil
	}
	
	mid := (start + end) / 2
	n := nodes[mid]
	
	n.left = t.buildBalanced(nodes, start, mid-1)
	n.right = t.buildBalanced(nodes, mid+1, end)
	
	return n
}

func (t *Tree) findParent(root, target *node) *node {
	if root == nil || root == target {
		return nil
	}
	if root.left == target || root.right == target {
		return root
	}
	if target.key < root.key {
		return t.findParent(root.left, target)
	}
	return t.findParent(root.right, target)
}

func (t *Tree) fullRebuild() {
	if t.root == nil {
		return
	}
	
	nodes := []*node{}
	t.inorderActive(t.root, &nodes)
	
	t.totalNodes = len(nodes)
	
	if t.totalNodes == 0 {
		t.root = nil
		return
	}
	
	t.root = t.buildBalancedActive(nodes, 0, len(nodes)-1)
}

func (t *Tree) inorderActive(n *node, nodes *[]*node) {
	if n == nil {
		return
	}
	t.inorderActive(n.left, nodes)
	if !n.deleted {
		*nodes = append(*nodes, n)
	}
	t.inorderActive(n.right, nodes)
}

func (t *Tree) buildBalancedActive(nodes []*node, start, end int) *node {
	if start > end {
		return nil
	}
	
	mid := (start + end) / 2
	n := nodes[mid]
	n.deleted = false
	
	n.left = t.buildBalancedActive(nodes, start, mid-1)
	n.right = t.buildBalancedActive(nodes, mid+1, end)
	
	return n
}
