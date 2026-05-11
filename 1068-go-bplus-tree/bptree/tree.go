package bptree

import (
	"math"
)

type BPTree struct {
	root   *node
	order  int
	minKeys int
	opts   *Options
	head   *node
	tail   *node
}

func New(opts ...*Options) *BPTree {
	var options *Options
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}
	options.validate()

	tree := &BPTree{
		order:  options.Order,
		minKeys: int(math.Ceil(float64(options.Order)/2)) - 1,
		opts:   options,
	}

	tree.root = newLeafNode(options.Order)
	tree.head = tree.root
	tree.tail = tree.root

	return tree
}

func (t *BPTree) Insert(key string, value interface{}) {
	if t.root.isLeaf {
		t.insertIntoLeaf(t.root, key, value)
		if t.root.keyCount() >= t.order {
			t.splitLeaf(t.root)
		}
		return
	}

	leaf := t.findLeaf(key)
	t.insertIntoLeaf(leaf, key, value)

	if leaf.keyCount() >= t.order {
		t.splitLeaf(leaf)
	}
}

func (t *BPTree) insertIntoLeaf(leaf *node, key string, value interface{}) {
	idx := 0
	for idx < leaf.keyCount() && leaf.keys[idx] < key {
		idx++
	}

	if !t.opts.AllowDuplicates {
		if idx < leaf.keyCount() && leaf.keys[idx] == key {
			leaf.values[idx] = value
			return
		}
	}

	leaf.keys = append(leaf.keys[:idx], append([]string{key}, leaf.keys[idx:]...)...)
	leaf.values = append(leaf.values[:idx], append([]interface{}{value}, leaf.values[idx:]...)...)
}

func (t *BPTree) splitLeaf(leaf *node) {
	mid := leaf.keyCount() / 2

	newLeaf := newLeafNode(t.order)
	newLeaf.keys = append(newLeaf.keys, leaf.keys[mid:]...)
	newLeaf.values = append(newLeaf.values, leaf.values[mid:]...)
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]

	if leaf.next != nil {
		leaf.next.prev = newLeaf
	}
	newLeaf.prev = leaf
	newLeaf.next = leaf.next
	leaf.next = newLeaf

	if newLeaf.next == nil {
		t.tail = newLeaf
	}

	parent := leaf.parent
	if parent == nil {
		t.createNewRoot(leaf, newLeaf)
		return
	}

	t.insertIntoParent(parent, leaf, newLeaf.keys[0], newLeaf)
}

func (t *BPTree) createNewRoot(left *node, right *node) {
	newRoot := newInternalNode(t.order)
	newRoot.keys = append(newRoot.keys, right.keys[0])
	newRoot.children = append(newRoot.children, left, right)
	left.parent = newRoot
	right.parent = newRoot
	t.root = newRoot
}

func (t *BPTree) insertIntoParent(parent *node, left *node, key string, right *node) {
	idx := 0
	for idx < parent.childCount() && parent.children[idx] != left {
		idx++
	}

	parent.keys = append(parent.keys[:idx], append([]string{key}, parent.keys[idx:]...)...)
	parent.children = append(parent.children[:idx+1], append([]*node{right}, parent.children[idx+1:]...)...)
	right.parent = parent

	if parent.keyCount() >= t.order {
		t.splitInternal(parent)
	}
}

func (t *BPTree) splitInternal(node *node) {
	mid := node.keyCount() / 2
	promotedKey := node.keys[mid]

	newNode := newInternalNode(t.order)
	newNode.keys = append(newNode.keys, node.keys[mid+1:]...)
	newNode.children = append(newNode.children, node.children[mid+1:]...)
	node.keys = node.keys[:mid]
	node.children = node.children[:mid+1]

	for _, child := range newNode.children {
		child.parent = newNode
	}

	parent := node.parent
	if parent == nil {
		newRoot := newInternalNode(t.order)
		newRoot.keys = append(newRoot.keys, promotedKey)
		newRoot.children = append(newRoot.children, node, newNode)
		node.parent = newRoot
		newNode.parent = newRoot
		t.root = newRoot
		return
	}

	t.insertIntoParent(parent, node, promotedKey, newNode)
}

func (t *BPTree) Search(key string) (interface{}, bool) {
	if t.root == nil {
		return nil, false
	}

	leaf := t.findLeaf(key)
	if leaf == nil {
		return nil, false
	}

	for i, k := range leaf.keys {
		if k == key {
			return leaf.values[i], true
		}
	}

	return nil, false
}

func (t *BPTree) findLeaf(key string) *node {
	if t.root == nil || t.root.isLeaf {
		return t.root
	}

	current := t.root
	for !current.isLeaf {
		idx := 0
		for idx < current.keyCount() && key >= current.keys[idx] {
			idx++
		}
		current = current.children[idx]
	}

	return current
}

func (t *BPTree) Delete(key string) bool {
	if t.root == nil {
		return false
	}

	leaf := t.findLeaf(key)
	if leaf == nil {
		return false
	}

	idx := -1
	for i, k := range leaf.keys {
		if k == key {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	leaf.keys = append(leaf.keys[:idx], leaf.keys[idx+1:]...)
	leaf.values = append(leaf.values[:idx], leaf.values[idx+1:]...)

	if leaf.keyCount() < t.minKeys {
		t.handleUnderflow(leaf)
	}

	return true
}

func (t *BPTree) handleUnderflow(node *node) {
	if node == t.root {
		if !node.isLeaf && node.childCount() == 1 {
			t.root = node.children[0]
			t.root.parent = nil
		}
		return
	}

	parent := node.parent
	nodeIdx := t.findChildIndex(parent, node)

	if nodeIdx > 0 {
		leftSibling := parent.children[nodeIdx-1]
		if node.isLeaf {
			if leftSibling.keyCount() > t.minKeys {
				t.borrowFromLeftLeaf(node, leftSibling, nodeIdx)
				return
			}
		} else {
			if leftSibling.keyCount() > t.minKeys {
				t.borrowFromLeftInternal(node, leftSibling, nodeIdx)
				return
			}
		}
	}

	if nodeIdx < parent.childCount()-1 {
		rightSibling := parent.children[nodeIdx+1]
		if node.isLeaf {
			if rightSibling.keyCount() > t.minKeys {
				t.borrowFromRightLeaf(node, rightSibling, nodeIdx)
				return
			}
		} else {
			if rightSibling.keyCount() > t.minKeys {
				t.borrowFromRightInternal(node, rightSibling, nodeIdx)
				return
			}
		}
	}

	if nodeIdx > 0 {
		leftSibling := parent.children[nodeIdx-1]
		if node.isLeaf {
			t.mergeLeaves(leftSibling, node, nodeIdx-1)
		} else {
			t.mergeInternals(leftSibling, node, nodeIdx-1)
		}
	} else {
		rightSibling := parent.children[nodeIdx+1]
		if node.isLeaf {
			t.mergeLeaves(node, rightSibling, nodeIdx)
		} else {
			t.mergeInternals(node, rightSibling, nodeIdx)
		}
	}
}

func (t *BPTree) findChildIndex(parent *node, child *node) int {
	for i, c := range parent.children {
		if c == child {
			return i
		}
	}
	return -1
}

func (t *BPTree) borrowFromLeftLeaf(node *node, left *node, nodeIdx int) {
	movedKey := left.keys[left.keyCount()-1]
	movedValue := left.values[left.valueCount()-1]

	left.keys = left.keys[:left.keyCount()-1]
	left.values = left.values[:left.valueCount()-1]

	node.keys = append([]string{movedKey}, node.keys...)
	node.values = append([]interface{}{movedValue}, node.values...)

	node.parent.keys[nodeIdx-1] = movedKey
}

func (t *BPTree) borrowFromRightLeaf(node *node, right *node, nodeIdx int) {
	movedKey := right.keys[0]
	movedValue := right.values[0]

	right.keys = right.keys[1:]
	right.values = right.values[1:]

	node.keys = append(node.keys, movedKey)
	node.values = append(node.values, movedValue)

	node.parent.keys[nodeIdx] = right.keys[0]
}

func (t *BPTree) borrowFromLeftInternal(target *node, left *node, nodeIdx int) {
	parent := target.parent

	movedKey := parent.keys[nodeIdx-1]
	movedChild := left.children[left.childCount()-1]

	parent.keys[nodeIdx-1] = left.keys[left.keyCount()-1]
	left.keys = left.keys[:left.keyCount()-1]
	left.children = left.children[:left.childCount()-1]

	target.keys = append([]string{movedKey}, target.keys...)
	target.children = append([]*node{movedChild}, target.children...)
	movedChild.parent = target
}

func (t *BPTree) borrowFromRightInternal(target *node, right *node, nodeIdx int) {
	parent := target.parent

	movedKey := parent.keys[nodeIdx]
	movedChild := right.children[0]

	parent.keys[nodeIdx] = right.keys[0]
	right.keys = right.keys[1:]
	right.children = right.children[1:]

	target.keys = append(target.keys, movedKey)
	target.children = append(target.children, movedChild)
	movedChild.parent = target
}

func (t *BPTree) mergeLeaves(left *node, right *node, nodeIdx int) {
	left.keys = append(left.keys, right.keys...)
	left.values = append(left.values, right.values...)

	if right.next != nil {
		right.next.prev = left
	}
	left.next = right.next

	if right == t.tail {
		t.tail = left
	}

	parent := right.parent
	parent.keys = append(parent.keys[:nodeIdx], parent.keys[nodeIdx+1:]...)
	parent.children = append(parent.children[:nodeIdx+1], parent.children[nodeIdx+2:]...)

	if parent.keyCount() < t.minKeys {
		t.handleUnderflow(parent)
	}
}

func (t *BPTree) mergeInternals(left *node, right *node, nodeIdx int) {
	parent := right.parent
	separator := parent.keys[nodeIdx]

	left.keys = append(left.keys, separator)
	left.keys = append(left.keys, right.keys...)
	left.children = append(left.children, right.children...)

	for _, child := range right.children {
		child.parent = left
	}

	parent.keys = append(parent.keys[:nodeIdx], parent.keys[nodeIdx+1:]...)
	parent.children = append(parent.children[:nodeIdx+1], parent.children[nodeIdx+2:]...)

	if parent.keyCount() < t.minKeys {
		t.handleUnderflow(parent)
	}
}

func (t *BPTree) Range(start string, end string) []*KeyValue {
	var result []*KeyValue

	if t.head == nil {
		return result
	}

	current := t.findLeaf(start)

	for current != nil {
		for i, key := range current.keys {
			if key > end {
				return result
			}
			if key >= start {
				result = append(result, &KeyValue{
					Key:   key,
					Value: current.values[i],
				})
			}
		}
		current = current.next
	}

	return result
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func (t *BPTree) Size() int {
	count := 0
	current := t.head
	for current != nil {
		count += current.keyCount()
		current = current.next
	}
	return count
}
