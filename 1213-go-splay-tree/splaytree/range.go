package splaytree

import "errors"

var ErrInvalidRange = errors.New("invalid range")

func (t *SplayTree) getKth(k int64) *Node {
	if t.root == nil || k <= 0 || k > t.root.Size {
		return nil
	}
	current := t.root
	for {
		current.push()
		leftSize := int64(0)
		if current.Left != nil {
			leftSize = current.Left.Size
		}
		if k == leftSize+1 {
			return current
		}
		if k <= leftSize {
			current = current.Left
		} else {
			k -= leftSize + 1
			current = current.Right
		}
	}
}

func (t *SplayTree) findSplit(key int64) (*Node, bool) {
	if t.root == nil {
		return nil, false
	}
	current := t.root
	var last *Node
	found := false
	for current != nil {
		last = current
		if current.Key == key {
			found = true
			break
		}
		if key < current.Key {
			current = current.Left
		} else {
			current = current.Right
		}
	}
	if last != nil {
		t.splayToRoot(last)
	}
	return last, found
}

func (t *SplayTree) RangeQueryByKey(l, r int64) (int64, error) {
	if t.root == nil {
		return 0, ErrInvalidRange
	}
	if l > r {
		return 0, ErrInvalidRange
	}
	keys := t.InOrder()
	values := t.InOrderValues()
	sum := int64(0)
	found := false
	for i, k := range keys {
		if k >= l && k <= r {
			sum += values[i]
			found = true
		}
	}
	if !found {
		return 0, ErrInvalidRange
	}
	return sum, nil
}

func (t *SplayTree) RangeAddByKey(l, r, delta int64) error {
	if t.root == nil {
		return ErrInvalidRange
	}
	if l > r {
		return ErrInvalidRange
	}
	keys := t.InOrder()
	values := t.InOrderValues()
	updated := false
	for i, k := range keys {
		if k >= l && k <= r {
			t.Put(k, values[i]+delta)
			updated = true
		}
	}
	if !updated {
		return ErrInvalidRange
	}
	return nil
}

func (t *SplayTree) RangeSetByKey(l, r, value int64) error {
	if t.root == nil {
		return ErrInvalidRange
	}
	if l > r {
		return ErrInvalidRange
	}
	keys := t.InOrder()
	updated := false
	for _, k := range keys {
		if k >= l && k <= r {
			t.Put(k, value)
			updated = true
		}
	}
	if !updated {
		return ErrInvalidRange
	}
	return nil
}

func (t *SplayTree) InOrder() []int64 {
	result := make([]int64, 0)
	t.inOrderHelper(t.root, &result)
	return result
}

func (t *SplayTree) inOrderHelper(node *Node, result *[]int64) {
	if node == nil {
		return
	}
	node.push()
	t.inOrderHelper(node.Left, result)
	*result = append(*result, node.Key)
	t.inOrderHelper(node.Right, result)
}

func (t *SplayTree) InOrderValues() []int64 {
	result := make([]int64, 0)
	t.inOrderValuesHelper(t.root, &result)
	return result
}

func (t *SplayTree) inOrderValuesHelper(node *Node, result *[]int64) {
	if node == nil {
		return
	}
	node.push()
	t.inOrderValuesHelper(node.Left, result)
	*result = append(*result, node.Value)
	t.inOrderValuesHelper(node.Right, result)
}

func (t *SplayTree) KeysInRange(l, r int64) []int64 {
	if t.root == nil || l > r {
		return nil
	}
	result := make([]int64, 0)
	keys := t.InOrder()
	for _, k := range keys {
		if k >= l && k <= r {
			result = append(result, k)
		}
	}
	return result
}

func (t *SplayTree) ValuesInRange(l, r int64) []int64 {
	if t.root == nil || l > r {
		return nil
	}
	result := make([]int64, 0)
	keys := t.InOrder()
	values := t.InOrderValues()
	for i, k := range keys {
		if k >= l && k <= r {
			result = append(result, values[i])
		}
	}
	return result
}
