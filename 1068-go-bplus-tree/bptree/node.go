package bptree

type node struct {
	keys     []string
	children []*node
	values   []interface{}
	isLeaf   bool
	parent   *node
	prev     *node
	next     *node
}

func newLeafNode(order int) *node {
	return &node{
		keys:   make([]string, 0, order-1),
		values: make([]interface{}, 0, order-1),
		isLeaf: true,
	}
}

func newInternalNode(order int) *node {
	return &node{
		keys:     make([]string, 0, order-1),
		children: make([]*node, 0, order),
		isLeaf:   false,
	}
}

func (n *node) keyCount() int {
	return len(n.keys)
}

func (n *node) childCount() int {
	return len(n.children)
}

func (n *node) valueCount() int {
	return len(n.values)
}
