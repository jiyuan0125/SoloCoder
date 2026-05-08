package skiplist

const (
	maxLevel    = 32
	probability = 0.5
)

type node struct {
	key     string
	value   []byte
	forward []*node
}

func newNode(key string, value []byte, level int) *node {
	return &node{
		key:     key,
		value:   value,
		forward: make([]*node, level+1),
	}
}
