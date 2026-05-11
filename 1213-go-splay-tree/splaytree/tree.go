package splaytree

import "errors"

var ErrKeyNotFound = errors.New("key not found")

type SplayTree struct {
	root *Node
}

func NewSplayTree() *SplayTree {
	return &SplayTree{}
}

func (t *SplayTree) findNode(key int64) *Node {
	if t.root == nil {
		return nil
	}
	current := t.root
	for {
		if current.Key == key {
			t.splayToRoot(current)
			return current
		}
		if key < current.Key {
			if current.Left == nil {
				t.splayToRoot(current)
				return nil
			}
			current = current.Left
		} else {
			if current.Right == nil {
				t.splayToRoot(current)
				return nil
			}
			current = current.Right
		}
	}
}

func (t *SplayTree) findMaxNode(node *Node) *Node {
	if node == nil {
		return nil
	}
	for node.Right != nil {
		node = node.Right
	}
	return node
}

func (t *SplayTree) Put(key, value int64) {
	if t.root == nil {
		t.root = newNode(key, value)
		return
	}

	current := t.root
	for {
		if current.Key == key {
			current.Value = value
			t.splayToRoot(current)
			return
		}
		if key < current.Key {
			if current.Left == nil {
				newNode := newNode(key, value)
				current.Left = newNode
				newNode.Parent = current
				t.splayToRoot(newNode)
				return
			}
			current = current.Left
		} else {
			if current.Right == nil {
				newNode := newNode(key, value)
				current.Right = newNode
				newNode.Parent = current
				t.splayToRoot(newNode)
				return
			}
			current = current.Right
		}
	}
}

func (t *SplayTree) Get(key int64) (int64, error) {
	node := t.findNode(key)
	if node == nil {
		return 0, ErrKeyNotFound
	}
	return node.Value, nil
}

func (t *SplayTree) Delete(key int64) error {
	if t.root == nil {
		return ErrKeyNotFound
	}

	node := t.findNode(key)
	if node == nil {
		return ErrKeyNotFound
	}

	if node.Left == nil && node.Right == nil {
		t.root = nil
		return nil
	}

	if node.Left == nil {
		t.root = node.Right
		t.root.Parent = nil
		return nil
	}

	if node.Right == nil {
		t.root = node.Left
		t.root.Parent = nil
		return nil
	}

	maxLeft := t.findMaxNode(node.Left)
	maxLeft.Right = node.Right
	node.Right.Parent = maxLeft

	t.root = node.Left
	t.root.Parent = nil

	t.splayToRoot(maxLeft)
	return nil
}

func (t *SplayTree) Root() *Node {
	return t.root
}

func (t *SplayTree) Size() int64 {
	if t.root == nil {
		return 0
	}
	return t.root.Size
}
