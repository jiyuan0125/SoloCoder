package queue

import (
	"sync/atomic"
)

type LockFreeQueue struct {
	head atomic.Pointer[TaggedNode]
	tail atomic.Pointer[TaggedNode]

	depth        atomic.Int64
	casSuccess   atomic.Int64
	casFailure   atomic.Int64
	allocsTotal  atomic.Int64
	enqueueTotal atomic.Int64

	pool *nodePool
}

func NewLockFree() *LockFreeQueue {
	dummy := &Node{}
	q := &LockFreeQueue{
		pool: newNodePool(),
	}
	initial := newTaggedNode(dummy, 0)
	q.head.Store(initial)
	q.tail.Store(initial)
	return q
}

func (q *LockFreeQueue) Enqueue(v interface{}) {
	var newNode *Node

	for {
		newNode = q.pool.Get()
		if newNode != nil {
			newNode.Value = v
			newNode.next.Store(nil)
			break
		}

		newNode = NewNode(v)
		q.allocsTotal.Add(1)
		break
	}

	for {
		tailTagged := q.tail.Load()
		if tailTagged == nil || tailTagged.node == nil {
			continue
		}
		tailNode := tailTagged.node

		nextTagged := tailNode.next.Load()

		currentTail := q.tail.Load()
		if tailTagged != currentTail {
			continue
		}

		if nextTagged == nil {
			newTagged := newTaggedNode(newNode, tailTagged.version+1)
			if tailNode.next.CompareAndSwap(nil, newTagged) {
				q.tail.CompareAndSwap(tailTagged, newTagged)
				q.depth.Add(1)
				q.enqueueTotal.Add(1)
				q.casSuccess.Add(1)
				return
			} else {
				q.casFailure.Add(1)
			}
		} else {
			advanceTagged := newTaggedNode(nextTagged.node, tailTagged.version+1)
			q.tail.CompareAndSwap(tailTagged, advanceTagged)
		}
	}
}

func (q *LockFreeQueue) Dequeue() (interface{}, bool) {
	for {
		headTagged := q.head.Load()
		if headTagged == nil || headTagged.node == nil {
			continue
		}
		headNode := headTagged.node

		tailTagged := q.tail.Load()
		nextTagged := headNode.next.Load()

		currentHead := q.head.Load()
		if headTagged != currentHead {
			continue
		}

		if headTagged == tailTagged {
			if nextTagged == nil {
				return nil, false
			}
			advanceTagged := newTaggedNode(nextTagged.node, tailTagged.version+1)
			q.tail.CompareAndSwap(tailTagged, advanceTagged)
		} else {
			if nextTagged == nil || nextTagged.node == nil {
				continue
			}
			v := nextTagged.node.Value
			newHeadTagged := newTaggedNode(nextTagged.node, headTagged.version+1)
			if q.head.CompareAndSwap(headTagged, newHeadTagged) {
				q.depth.Add(-1)
				q.pool.Put(headNode)
				return v, true
			}
		}
	}
}

func (q *LockFreeQueue) Depth() int64 {
	return q.depth.Load()
}

func (q *LockFreeQueue) CASSuccess() int64 {
	return q.casSuccess.Load()
}

func (q *LockFreeQueue) CASFailure() int64 {
	return q.casFailure.Load()
}

func (q *LockFreeQueue) AllocsTotal() int64 {
	return q.allocsTotal.Load()
}

func (q *LockFreeQueue) PoolHits() int64 {
	return q.pool.hits.Load()
}

func (q *LockFreeQueue) PoolMisses() int64 {
	return q.pool.misses.Load()
}

func (q *LockFreeQueue) EnqueueTotal() int64 {
	return q.enqueueTotal.Load()
}
