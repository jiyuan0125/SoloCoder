package euler

import (
	"fmt"
	"strings"
)

type Node string

type Edge struct {
	From Node
	To   Node
}

type Graph struct {
	Nodes map[Node]bool
	Edges []Edge
	adj   map[Node][]int
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[Node]bool),
		Edges: make([]Edge, 0),
		adj:   make(map[Node][]int),
	}
}

func (g *Graph) AddNode(n Node) {
	g.Nodes[n] = true
	if _, ok := g.adj[n]; !ok {
		g.adj[n] = make([]int, 0)
	}
}

func (g *Graph) AddEdge(from, to Node) {
	g.AddNode(from)
	g.AddNode(to)
	edgeIdx := len(g.Edges)
	g.Edges = append(g.Edges, Edge{From: from, To: to})
	g.adj[from] = append(g.adj[from], edgeIdx)
	g.adj[to] = append(g.adj[to], edgeIdx)
}

func (g *Graph) Degree(n Node) int {
	return len(g.adj[n])
}

func (g *Graph) OddDegreeNodes() []Node {
	odds := make([]Node, 0)
	for n := range g.Nodes {
		if g.Degree(n)%2 != 0 {
			odds = append(odds, n)
		}
	}
	return odds
}

func (g *Graph) OtherEnd(edgeIdx int, from Node) (Node, error) {
	if edgeIdx < 0 || edgeIdx >= len(g.Edges) {
		return "", fmt.Errorf("edge index out of bounds")
	}
	e := g.Edges[edgeIdx]
	if e.From == from {
		return e.To, nil
	}
	if e.To == from {
		return e.From, nil
	}
	return "", fmt.Errorf("node %s not in edge %d", from, edgeIdx)
}

func (g *Graph) String() string {
	var sb strings.Builder
	sb.WriteString("Graph{\n")
	sb.WriteString(fmt.Sprintf("  Nodes: %v\n", g.nodeList()))
	sb.WriteString("  Edges:\n")
	for i, e := range g.Edges {
		sb.WriteString(fmt.Sprintf("    %d: %s-%s\n", i, e.From, e.To))
	}
	sb.WriteString("}")
	return sb.String()
}

func (g *Graph) nodeList() []Node {
	list := make([]Node, 0, len(g.Nodes))
	for n := range g.Nodes {
		list = append(list, n)
	}
	return list
}

func (g *Graph) hasEdges() bool {
	return len(g.Edges) > 0
}

func (g *Graph) getNodesWithEdges() map[Node]bool {
	hasEdge := make(map[Node]bool)
	for _, e := range g.Edges {
		hasEdge[e.From] = true
		hasEdge[e.To] = true
	}
	return hasEdge
}
