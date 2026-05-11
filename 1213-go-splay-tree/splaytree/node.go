package splaytree

type Node struct {
	Key    int64
	Value  int64
	Left   *Node
	Right  *Node
	Parent *Node
	Size   int64
	Sum    int64
	Add    int64
}

func newNode(key, value int64) *Node {
	return &Node{
		Key:   key,
		Value: value,
		Size:  1,
		Sum:   value,
	}
}

func (n *Node) update() {
	n.Size = 1
	n.Sum = n.Value
	if n.Left != nil {
		n.Size += n.Left.Size
		n.Sum += n.Left.Sum
	}
	if n.Right != nil {
		n.Size += n.Right.Size
		n.Sum += n.Right.Sum
	}
}

func (n *Node) push() {
	if n.Add == 0 {
		return
	}
	if n.Left != nil {
		n.Left.Value += n.Add
		n.Left.Sum += n.Left.Size * n.Add
		n.Left.Add += n.Add
	}
	if n.Right != nil {
		n.Right.Value += n.Add
		n.Right.Sum += n.Right.Size * n.Add
		n.Right.Add += n.Add
	}
	n.Add = 0
}
