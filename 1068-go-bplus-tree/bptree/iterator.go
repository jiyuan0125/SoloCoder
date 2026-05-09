package bptree

type Iterator struct {
	tree      *BPTree
	direction int
	current   *node
	index     int
}

const (
	Forward  = 0
	Backward = 1
)

func (t *BPTree) NewIterator(direction int) *Iterator {
	var current *node
	var index int

	if direction == Forward {
		current = t.head
		index = 0
	} else {
		current = t.tail
		if current != nil {
			index = current.keyCount() - 1
		} else {
			index = -1
		}
	}

	return &Iterator{
		tree:      t,
		direction: direction,
		current:   current,
		index:     index,
	}
}

func (it *Iterator) Next() *KeyValue {
	if it.current == nil || it.index < 0 || it.index >= it.current.keyCount() {
		return nil
	}

	kv := &KeyValue{
		Key:   it.current.keys[it.index],
		Value: it.current.values[it.index],
	}

	if it.direction == Forward {
		it.index++
		if it.index >= it.current.keyCount() {
			it.current = it.current.next
			it.index = 0
		}
	} else {
		it.index--
		if it.index < 0 {
			it.current = it.current.prev
			if it.current != nil {
				it.index = it.current.keyCount() - 1
			}
		}
	}

	return kv
}

func (it *Iterator) HasNext() bool {
	return it.current != nil && it.index >= 0 && it.index < it.current.keyCount()
}

func (t *BPTree) Scan(direction int, batchSize int, startOffset int) []*KeyValue {
	if batchSize <= 0 {
		batchSize = 10
	}

	var result []*KeyValue
	it := t.NewIterator(direction)

	for i := 0; i < startOffset && it.HasNext(); i++ {
		it.Next()
	}

	for len(result) < batchSize && it.HasNext() {
		kv := it.Next()
		if kv != nil {
			result = append(result, kv)
		}
	}

	return result
}

func (t *BPTree) ForwardScan(batchSize int, startOffset int) []*KeyValue {
	return t.Scan(Forward, batchSize, startOffset)
}

func (t *BPTree) BackwardScan(batchSize int, startOffset int) []*KeyValue {
	return t.Scan(Backward, batchSize, startOffset)
}
