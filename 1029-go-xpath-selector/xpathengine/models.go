package xpathengine

type NodeType int

const (
	ElementNode NodeType = iota
	TextNode
	AttributeNode
	DocumentNode
	CommentNode
	ProcessingInstructionNode
)

type QName struct {
	Prefix string
	Local  string
}

type Node struct {
	Type       NodeType
	Name       QName
	Namespace  string
	Attributes []*Node
	Parent     *Node
	Children   []*Node
	Text       string
	Line       int
}

func (n *Node) IsElement() bool {
	return n.Type == ElementNode
}

func (n *Node) IsText() bool {
	return n.Type == TextNode
}

func (n *Node) IsAttribute() bool {
	return n.Type == AttributeNode
}

func (n *Node) GetLocalName() string {
	return n.Name.Local
}

func (n *Node) GetPrefix() string {
	return n.Name.Prefix
}

func (n *Node) GetNamespace() string {
	return n.Namespace
}

func (n *Node) GetAttribute(name QName) string {
	for _, attr := range n.Attributes {
		if attr.Name.Local == name.Local && attr.Namespace == n.Namespace {
			return attr.Text
		}
	}
	return ""
}

func (n *Node) GetChildElements() []*Node {
	var result []*Node
	for _, child := range n.Children {
		if child.Type == ElementNode {
			result = append(result, child)
		}
	}
	return result
}

func (n *Node) GetChildTexts() []*Node {
	var result []*Node
	for _, child := range n.Children {
		if child.Type == TextNode {
			result = append(result, child)
		}
	}
	return result
}

func (n *Node) StringValue() string {
	switch n.Type {
	case TextNode:
		return n.Text
	case AttributeNode:
		return n.Text
	case ElementNode, DocumentNode:
		var text string
		for _, child := range n.Children {
			if child.Type == TextNode {
				text += child.Text
			} else if child.Type == ElementNode {
				text += child.StringValue()
			}
		}
		return text
	default:
		return ""
	}
}

type XPathResult struct {
	Type  ResultType
	Value interface{}
}

type ResultType int

const (
	NodeSetResult ResultType = iota
	StringResult
	NumberResult
	BooleanResult
)

type NodeSet []*Node

func (r *XPathResult) NodeSet() (NodeSet, bool) {
	if r.Type != NodeSetResult {
		return nil, false
	}
	ns, ok := r.Value.(NodeSet)
	return ns, ok
}

func (r *XPathResult) String() (string, bool) {
	if r.Type != StringResult {
		return "", false
	}
	s, ok := r.Value.(string)
	return s, ok
}

func (r *XPathResult) Number() (float64, bool) {
	if r.Type != NumberResult {
		return 0, false
	}
	n, ok := r.Value.(float64)
	return n, ok
}

func (r *XPathResult) Boolean() (bool, bool) {
	if r.Type != BooleanResult {
		return false, false
	}
	b, ok := r.Value.(bool)
	return b, ok
}
