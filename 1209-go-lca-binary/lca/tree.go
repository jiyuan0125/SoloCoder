package lca

import (
	"fmt"
	"math/bits"
	"strings"
)

type FamilyTree struct {
	nodes      map[string]*Node
	root       *Node
	maxDepth   int
	logMax     int
	up         map[string][]*Node
	preprocessed bool
}

type Node struct {
	ID       string
	Depth    int
	Children []*Node
	Parent   *Node
}

func NewFamilyTree() *FamilyTree {
	return &FamilyTree{
		nodes: make(map[string]*Node),
		up:    make(map[string][]*Node),
	}
}

func (ft *FamilyTree) AddNode(id string, parentID *string) error {
	if _, exists := ft.nodes[id]; exists {
		return fmt.Errorf("node %s already exists", id)
	}

	node := &Node{ID: id}

	if parentID == nil || *parentID == "" {
		if ft.root != nil {
			return fmt.Errorf("cannot have multiple root nodes: existing root is %s, trying to add %s", ft.root.ID, id)
		}
		ft.root = node
		ft.nodes[id] = node
		return nil
	}

	parent, exists := ft.nodes[*parentID]
	if !exists {
		return fmt.Errorf("parent node %s not found for child %s", *parentID, id)
	}

	node.Parent = parent
	parent.Children = append(parent.Children, node)
	ft.nodes[id] = node

	return nil
}

func (ft *FamilyTree) ValidateAndBuild() error {
	if ft.root == nil {
		return fmt.Errorf("no root node found (node with parent_id null or empty string)")
	}

	if ft.hasCycle() {
		return fmt.Errorf("cycle detected in the family tree")
	}

	ft.calculateDepths()
	return nil
}

func (ft *FamilyTree) hasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(*Node) bool
	dfs = func(node *Node) bool {
		visited[node.ID] = true
		recStack[node.ID] = true

		for _, child := range node.Children {
			if !visited[child.ID] {
				if dfs(child) {
					return true
				}
			} else if recStack[child.ID] {
				return true
			}
		}

		recStack[node.ID] = false
		return false
	}

	return dfs(ft.root)
}

func (ft *FamilyTree) calculateDepths() {
	ft.maxDepth = 0
	queue := []*Node{ft.root}
	ft.root.Depth = 0

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node.Depth > ft.maxDepth {
			ft.maxDepth = node.Depth
		}

		for _, child := range node.Children {
			child.Depth = node.Depth + 1
			queue = append(queue, child)
		}
	}

	if ft.maxDepth == 0 {
		ft.logMax = 1
	} else {
		ft.logMax = bits.Len(uint(ft.maxDepth)) + 1
	}
}

func (ft *FamilyTree) Preprocess() {
	if ft.preprocessed {
		return
	}

	for id, node := range ft.nodes {
		ft.up[id] = make([]*Node, ft.logMax)
		ft.up[id][0] = node.Parent
		if node == ft.root {
			ft.up[id][0] = node
		}
	}

	for k := 1; k < ft.logMax; k++ {
		for id := range ft.nodes {
			if ft.up[id][k-1] != nil {
				ft.up[id][k] = ft.up[ft.up[id][k-1].ID][k-1]
			}
		}
	}

	ft.preprocessed = true
}

func (ft *FamilyTree) LCA(a, b string) (string, error) {
	nodeA, exists := ft.nodes[a]
	if !exists {
		return "", fmt.Errorf("node %s not found", a)
	}

	nodeB, exists := ft.nodes[b]
	if !exists {
		return "", fmt.Errorf("node %s not found", b)
	}

	if nodeA == nodeB {
		return nodeA.ID, nil
	}

	if nodeA.Depth < nodeB.Depth {
		nodeA, nodeB = nodeB, nodeA
		a, b = b, a
	}

	nodeA = ft.upToDepth(a, nodeB.Depth)

	if nodeA == nodeB {
		return nodeA.ID, nil
	}

	for k := ft.logMax - 1; k >= 0; k-- {
		if ft.up[nodeA.ID][k] != ft.up[nodeB.ID][k] {
			nodeA = ft.up[nodeA.ID][k]
			nodeB = ft.up[nodeB.ID][k]
		}
	}

	return nodeA.Parent.ID, nil
}

func (ft *FamilyTree) upToDepth(nodeID string, targetDepth int) *Node {
	node := ft.nodes[nodeID]

	if node.Depth == targetDepth {
		return node
	}

	steps := node.Depth - targetDepth

	for k := ft.logMax - 1; k >= 0; k-- {
		if steps >= (1 << k) {
			node = ft.up[node.ID][k]
			steps -= (1 << k)
		}
	}

	return node
}

func (n *Node) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Node %s (depth %d)", n.ID, n.Depth))
	if n.Parent != nil {
		sb.WriteString(fmt.Sprintf(", parent: %s", n.Parent.ID))
	}
	if len(n.Children) > 0 {
		sb.WriteString(", children: [")
		for i, child := range n.Children {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(child.ID)
		}
		sb.WriteString("]")
	}
	return sb.String()
}
