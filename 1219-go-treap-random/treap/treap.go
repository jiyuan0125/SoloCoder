package treap

import (
	"errors"
	"math/rand"
)

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrEmptyTree      = errors.New("tree is empty")
	ErrNoPrev         = errors.New("no predecessor")
	ErrNoNext         = errors.New("no successor")
)

type Node struct {
	Key      int64
	Value    int64
	Priority int64
	Left     *Node
	Right    *Node
}

type Treap struct {
	root *Node
	rng  *rand.Rand
}

func New(src rand.Source) *Treap {
	t := &Treap{}
	if src != nil {
		t.rng = rand.New(src)
	}
	return t
}

func (t *Treap) randomPriority() int64 {
	if t.rng != nil {
		return t.rng.Int63()
	}
	return rand.Int63()
}

func rightRotate(y *Node) *Node {
	x := y.Left
	T2 := x.Right

	x.Right = y
	y.Left = T2

	return x
}

func leftRotate(x *Node) *Node {
	y := x.Right
	T2 := y.Left

	y.Left = x
	x.Right = T2

	return y
}

func (t *Treap) Insert(key int64, value int64) error {
	t.root = t.insert(t.root, key, value)
	return nil
}

func (t *Treap) insert(node *Node, key int64, value int64) *Node {
	if node == nil {
		return &Node{
			Key:      key,
			Value:    value,
			Priority: t.randomPriority(),
		}
	}

	if key == node.Key {
		node.Value = value
		return node
	}

	if key < node.Key {
		node.Left = t.insert(node.Left, key, value)
		if node.Left.Priority < node.Priority {
			node = rightRotate(node)
		}
	} else {
		node.Right = t.insert(node.Right, key, value)
		if node.Right.Priority < node.Priority {
			node = leftRotate(node)
		}
	}

	return node
}

func (t *Treap) Search(key int64) (int64, error) {
	node := t.search(t.root, key)
	if node == nil {
		return 0, ErrKeyNotFound
	}
	return node.Value, nil
}

func (t *Treap) search(node *Node, key int64) *Node {
	if node == nil {
		return nil
	}
	if key == node.Key {
		return node
	}
	if key < node.Key {
		return t.search(node.Left, key)
	}
	return t.search(node.Right, key)
}

func (t *Treap) Delete(key int64) error {
	if t.root == nil {
		return ErrEmptyTree
	}
	newRoot, deleted := t.delete(t.root, key)
	if !deleted {
		return ErrKeyNotFound
	}
	t.root = newRoot
	return nil
}

func (t *Treap) delete(node *Node, key int64) (*Node, bool) {
	if node == nil {
		return nil, false
	}

	var deleted bool
	if key < node.Key {
		node.Left, deleted = t.delete(node.Left, key)
		return node, deleted
	}
	if key > node.Key {
		node.Right, deleted = t.delete(node.Right, key)
		return node, deleted
	}

	if node.Left == nil {
		return node.Right, true
	}
	if node.Right == nil {
		return node.Left, true
	}

	if node.Left.Priority < node.Right.Priority {
		node = rightRotate(node)
		node.Right, _ = t.delete(node.Right, key)
	} else {
		node = leftRotate(node)
		node.Left, _ = t.delete(node.Left, key)
	}
	return node, true
}

func (t *Treap) Predecessor(key int64) (*Node, error) {
	if t.root == nil {
		return nil, ErrEmptyTree
	}
	node := t.predecessor(t.root, key, nil)
	if node == nil {
		return nil, ErrNoPrev
	}
	return node, nil
}

func (t *Treap) predecessor(node *Node, key int64, currentBest *Node) *Node {
	if node == nil {
		return currentBest
	}

	if node.Key == key {
		if node.Left != nil {
			return t.maxNode(node.Left)
		}
		return currentBest
	}

	if node.Key < key {
		return t.predecessor(node.Right, key, node)
	}
	return t.predecessor(node.Left, key, currentBest)
}

func (t *Treap) maxNode(node *Node) *Node {
	for node.Right != nil {
		node = node.Right
	}
	return node
}

func (t *Treap) Successor(key int64) (*Node, error) {
	if t.root == nil {
		return nil, ErrEmptyTree
	}
	node := t.successor(t.root, key, nil)
	if node == nil {
		return nil, ErrNoNext
	}
	return node, nil
}

func (t *Treap) successor(node *Node, key int64, currentBest *Node) *Node {
	if node == nil {
		return currentBest
	}

	if node.Key == key {
		if node.Right != nil {
			return t.minNode(node.Right)
		}
		return currentBest
	}

	if node.Key > key {
		return t.successor(node.Left, key, node)
	}
	return t.successor(node.Right, key, currentBest)
}

func (t *Treap) minNode(node *Node) *Node {
	for node.Left != nil {
		node = node.Left
	}
	return node
}

func (t *Treap) InOrder() []*Node {
	var result []*Node
	t.inOrder(t.root, &result)
	return result
}

func (t *Treap) inOrder(node *Node, result *[]*Node) {
	if node == nil {
		return
	}
	t.inOrder(node.Left, result)
	*result = append(*result, node)
	t.inOrder(node.Right, result)
}
