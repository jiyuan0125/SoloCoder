package radix

import (
	"strings"
)

type NodeType int

const (
	NodeStatic   NodeType = iota
	NodeParam
	NodeWildcard
)

type Node struct {
	Type     NodeType
	Prefix   string
	Children []*Node
	Value    interface{}
	IsEnd    bool
}

type Tree struct {
	Root *Node
}

func NewTree() *Tree {
	return &Tree{
		Root: &Node{
			Type:     NodeStatic,
			Prefix:   "",
			Children: []*Node{},
		},
	}
}

func (t *Tree) Insert(path string, value interface{}) {
	if path == "" {
		t.Root.Value = value
		t.Root.IsEnd = true
		return
	}
	t.insertRec(t.Root, path, value)
}

func (t *Tree) insertRec(node *Node, path string, value interface{}) {
	if path == "" {
		node.Value = value
		node.IsEnd = true
		return
	}

	segType, segContent, remaining := splitNextSegment(path)

	for _, child := range node.Children {
		if child.Type != segType {
			continue
		}

		if segType == NodeParam || segType == NodeWildcard {
			if child.Prefix == segContent {
				if remaining == "" {
					child.Value = value
					child.IsEnd = true
				} else {
					t.insertRec(child, remaining, value)
				}
				return
			}
			continue
		}

		commonLen := longestCommonPrefix(segContent, child.Prefix)

		if commonLen == 0 {
			continue
		}

		if commonLen == len(child.Prefix) {
			newPath := segContent[commonLen:] + remaining
			if newPath == "" {
				child.Value = value
				child.IsEnd = true
			} else {
				t.insertRec(child, newPath, value)
			}
			return
		}

		if commonLen == len(segContent) {
			oldPrefix := child.Prefix[commonLen:]
			oldNode := &Node{
				Type:     child.Type,
				Prefix:   oldPrefix,
				Children: child.Children,
				Value:    child.Value,
				IsEnd:    child.IsEnd,
			}

			child.Prefix = segContent
			child.Children = []*Node{oldNode}
			if remaining == "" {
				child.Value = value
				child.IsEnd = true
			} else {
				child.Value = nil
				child.IsEnd = false
				t.insertRec(child, remaining, value)
			}
			return
		}

		nodePrefix := child.Prefix[:commonLen]
		nodeSeg1 := child.Prefix[commonLen:]
		nodeSeg2 := segContent[commonLen:]

		child1 := &Node{
			Type:     child.Type,
			Prefix:   nodeSeg1,
			Children: child.Children,
			Value:    child.Value,
			IsEnd:    child.IsEnd,
		}

		child.Prefix = nodePrefix
		child.Children = []*Node{child1}
		child.Value = nil
		child.IsEnd = false

		if remaining == "" {
			child2 := &Node{
				Type:     NodeStatic,
				Prefix:   nodeSeg2,
				Children: []*Node{},
				Value:    value,
				IsEnd:    true,
			}
			child.Children = append(child.Children, child2)
		} else {
			child2 := &Node{
				Type:     NodeStatic,
				Prefix:   nodeSeg2,
				Children: []*Node{},
				Value:    nil,
				IsEnd:    false,
			}
			child.Children = append(child.Children, child2)
			t.insertRec(child2, remaining, value)
		}

		sortChildren(child.Children)
		return
	}

	newNode := &Node{
		Type:     segType,
		Prefix:   segContent,
		Children: []*Node{},
		Value:    nil,
		IsEnd:    false,
	}

	if remaining == "" {
		newNode.Value = value
		newNode.IsEnd = true
	}

	node.Children = append(node.Children, newNode)
	sortChildren(node.Children)

	if remaining != "" {
		t.insertRec(newNode, remaining, value)
	}
}

func splitNextSegment(path string) (NodeType, string, string) {
	if path == "" {
		return NodeStatic, "", ""
	}

	if path[0] == '*' {
		slashIdx := strings.Index(path, "/")
		if slashIdx == -1 {
			return NodeWildcard, path, ""
		}
		return NodeWildcard, path[:slashIdx], path[slashIdx:]
	}

	if path[0] == ':' {
		slashIdx := strings.Index(path, "/")
		if slashIdx == -1 {
			return NodeParam, path, ""
		}
		return NodeParam, path[:slashIdx], path[slashIdx:]
	}

	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == ':' || c == '*' {
			return NodeStatic, path[:i], path[i:]
		}
	}

	return NodeStatic, path, ""
}

func (t *Tree) Search(path string) (interface{}, map[string]string, bool) {
	params := make(map[string]string)
	return t.searchRec(t.Root, path, params)
}

func (t *Tree) searchRec(node *Node, path string, params map[string]string) (interface{}, map[string]string, bool) {
	if path == "" {
		if node.IsEnd {
			return node.Value, params, true
		}
		return nil, params, false
	}

	var wildcardValue interface{}
	var wildcardParams map[string]string
	var wildcardFound bool

	for _, child := range node.Children {
		switch child.Type {
		case NodeWildcard:
			if child.IsEnd {
				newParams := copyParams(params)
				if len(child.Prefix) > 1 {
					paramName := child.Prefix[1:]
					newParams[paramName] = path
				}
				wildcardValue = child.Value
				wildcardParams = newParams
				wildcardFound = true
			}
			continue

		case NodeParam:
			nextSlash := strings.Index(path, "/")
			var paramValue string
			var restPath string

			if nextSlash == -1 {
				paramValue = path
				restPath = ""
			} else {
				paramValue = path[:nextSlash]
				restPath = path[nextSlash:]
			}

			if paramValue == "" {
				continue
			}

			newParams := copyParams(params)
			if len(child.Prefix) > 1 {
				paramName := child.Prefix[1:]
				newParams[paramName] = paramValue
			}

			value, foundParams, found := t.searchRec(child, restPath, newParams)
			if found {
				return value, foundParams, true
			}

		case NodeStatic:
			if !strings.HasPrefix(path, child.Prefix) {
				continue
			}

			value, foundParams, found := t.searchRec(child, path[len(child.Prefix):], params)
			if found {
				return value, foundParams, true
			}
		}
	}

	if wildcardFound {
		return wildcardValue, wildcardParams, true
	}

	return nil, params, false
}

func (t *Tree) Delete(path string) bool {
	return t.deleteRec(t.Root, path, false)
}

func (t *Tree) deleteRec(node *Node, path string, isRoot bool) bool {
	if path == "" {
		if !node.IsEnd {
			return false
		}
		node.IsEnd = false
		node.Value = nil
		return true
	}

	segType, segContent, remaining := splitNextSegment(path)

	for i, child := range node.Children {
		if child.Type != segType {
			continue
		}

		if segType == NodeParam || segType == NodeWildcard {
			if child.Prefix != segContent {
				continue
			}

			found := t.deleteRec(child, remaining, false)
			if found {
				t.tryMerge(node, child, i)
			}
			return found
		}

		if !strings.HasPrefix(segContent, child.Prefix) {
			continue
		}

		commonLen := len(child.Prefix)
		newPath := segContent[commonLen:] + remaining

		found := t.deleteRec(child, newPath, false)
		if found {
			t.tryMerge(node, child, i)
		}
		return found
	}

	return false
}

func (t *Tree) tryMerge(parent *Node, child *Node, childIdx int) {
	if child.IsEnd {
		return
	}

	if len(child.Children) == 0 {
		parent.Children = append(parent.Children[:childIdx], parent.Children[childIdx+1:]...)
		return
	}

	if len(child.Children) == 1 {
		onlyChild := child.Children[0]
		if onlyChild.Type == NodeStatic && child.Type == NodeStatic {
			child.Prefix = child.Prefix + onlyChild.Prefix
			child.Children = onlyChild.Children
			child.Value = onlyChild.Value
			child.IsEnd = onlyChild.IsEnd
		}
	}
}

func (t *Tree) List() []string {
	result := []string{}
	t.listRec(t.Root, "", &result)
	return result
}

func (t *Tree) listRec(node *Node, current string, result *[]string) {
	fullPath := current + node.Prefix

	if node.IsEnd {
		*result = append(*result, fullPath)
	}

	for _, child := range node.Children {
		t.listRec(child, fullPath, result)
	}
}

func (t *Tree) Stats() Stats {
	nodeCount := 0
	maxHeight := 0
	t.traverse(t.Root, 0, &nodeCount, &maxHeight)

	return Stats{
		NodeCount: nodeCount,
		Height:    maxHeight,
		IsEmpty:   nodeCount == 1 && !t.Root.IsEnd && len(t.Root.Children) == 0,
	}
}

func (t *Tree) traverse(node *Node, depth int, nodeCount *int, maxHeight *int) {
	*nodeCount++
	if depth > *maxHeight {
		*maxHeight = depth
	}

	for _, child := range node.Children {
		t.traverse(child, depth+1, nodeCount, maxHeight)
	}
}

type Stats struct {
	NodeCount int
	Height    int
	IsEmpty   bool
}

func longestCommonPrefix(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return minLen
}

func sortChildren(children []*Node) {
	for i := 0; i < len(children); i++ {
		for j := i + 1; j < len(children); j++ {
			if compareNodes(children[i], children[j]) > 0 {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
}

func compareNodes(a, b *Node) int {
	if a.Type != b.Type {
		return int(a.Type) - int(b.Type)
	}
	return strings.Compare(a.Prefix, b.Prefix)
}

func copyParams(params map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		result[k] = v
	}
	return result
}
