package queue

import (
	"sync/atomic"
)

type Node struct {
	Value interface{}
	next  atomic.Pointer[TaggedNode]
}

type TaggedNode struct {
	node    *Node
	version uint64
}

func NewNode(v interface{}) *Node {
	return &Node{Value: v}
}

func newTaggedNode(n *Node, ver uint64) *TaggedNode {
	return &TaggedNode{node: n, version: ver}
}

type nodePool struct {
	stack  atomic.Pointer[poolNode]
	hits   atomic.Int64
	misses atomic.Int64
}

type poolNode struct {
	node *Node
	next *poolNode
}

func newNodePool() *nodePool {
	return &nodePool{}
}

func (p *nodePool) Get() *Node {
	for {
		head := p.stack.Load()
		if head == nil {
			p.misses.Add(1)
			return nil
		}
		next := head.next
		if p.stack.CompareAndSwap(head, next) {
			p.hits.Add(1)
			n := head.node
			head.next = nil
			return n
		}
	}
}

func (p *nodePool) Put(n *Node) {
	n.Value = nil
	n.next.Store(nil)
	pn := &poolNode{node: n}
	for {
		head := p.stack.Load()
		pn.next = head
		if p.stack.CompareAndSwap(head, pn) {
			return
		}
	}
}
