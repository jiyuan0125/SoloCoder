package regex

type NodeType int

const (
	NodeLiteral NodeType = iota
	NodeDot
	NodeConcat
	NodeAlternate
	NodeStar
	NodeEmpty
)

type Node struct {
	typ    NodeType
	value  byte
	left   *Node
	right  *Node
	child  *Node
}

func newLiteralNode(ch byte) *Node {
	return &Node{typ: NodeLiteral, value: ch}
}

func newDotNode() *Node {
	return &Node{typ: NodeDot}
}

func newEmptyNode() *Node {
	return &Node{typ: NodeEmpty}
}

func newConcatNode(left, right *Node) *Node {
	return &Node{typ: NodeConcat, left: left, right: right}
}

func newAlternateNode(left, right *Node) *Node {
	return &Node{typ: NodeAlternate, left: left, right: right}
}

func newStarNode(child *Node) *Node {
	return &Node{typ: NodeStar, child: child}
}
