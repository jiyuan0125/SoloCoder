package kruskal

import (
	"sort"
)

type Edge struct {
	From string
	To   string
	Cost int
}

type Graph struct {
	Nodes []string
	Edges []Edge
}

type MSTResult struct {
	Edges       []Edge
	TotalCost   int
	IsConnected bool
	Isolated    []string
}

type UnionFind struct {
	parent map[string]string
	rank   map[string]int
}

func NewUnionFind(nodes []string) *UnionFind {
	uf := &UnionFind{
		parent: make(map[string]string),
		rank:   make(map[string]int),
	}
	for _, node := range nodes {
		uf.parent[node] = node
		uf.rank[node] = 0
	}
	return uf
}

func (uf *UnionFind) Find(node string) string {
	if uf.parent[node] != node {
		uf.parent[node] = uf.Find(uf.parent[node])
	}
	return uf.parent[node]
}

func (uf *UnionFind) Union(node1, node2 string) bool {
	root1 := uf.Find(node1)
	root2 := uf.Find(node2)

	if root1 == root2 {
		return false
	}

	if uf.rank[root1] < uf.rank[root2] {
		uf.parent[root1] = root2
	} else if uf.rank[root1] > uf.rank[root2] {
		uf.parent[root2] = root1
	} else {
		uf.parent[root2] = root1
		uf.rank[root1]++
	}

	return true
}

func (uf *UnionFind) IsConnected(node1, node2 string) bool {
	return uf.Find(node1) == uf.Find(node2)
}

func (uf *UnionFind) GetConnectedComponents() map[string][]string {
	components := make(map[string][]string)
	for node := range uf.parent {
		root := uf.Find(node)
		components[root] = append(components[root], node)
	}
	return components
}

func KruskalMST(graph Graph) MSTResult {
	edges := make([]Edge, len(graph.Edges))
	copy(edges, graph.Edges)

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Cost < edges[j].Cost
	})

	uf := NewUnionFind(graph.Nodes)
	var mstEdges []Edge
	totalCost := 0

	for _, edge := range edges {
		if !uf.IsConnected(edge.From, edge.To) {
			uf.Union(edge.From, edge.To)
			mstEdges = append(mstEdges, edge)
			totalCost += edge.Cost
		}
	}

	components := uf.GetConnectedComponents()
	isConnected := len(components) == 1
	var isolated []string

	if !isConnected {
		for _, component := range components {
			if len(component) == 1 {
				isolated = append(isolated, component[0])
			}
		}
	}

	return MSTResult{
		Edges:       mstEdges,
		TotalCost:   totalCost,
		IsConnected: isConnected,
		Isolated:    isolated,
	}
}

func RemoveDuplicateEdges(edges []Edge) []Edge {
	type edgeKey struct {
		from, to string
	}
	seen := make(map[edgeKey]bool)
	var result []Edge

	for _, edge := range edges {
		key1 := edgeKey{edge.From, edge.To}
		key2 := edgeKey{edge.To, edge.From}

		if !seen[key1] && !seen[key2] {
			seen[key1] = true
			result = append(result, edge)
		}
	}

	return result
}
