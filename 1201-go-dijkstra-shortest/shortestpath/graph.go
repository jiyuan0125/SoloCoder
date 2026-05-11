package shortestpath

import (
	"container/heap"
	"math"
)

type Graph struct {
	nodes  map[string]bool
	adj    map[string][]edge
}

type edge struct {
	to     string
	weight float64
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]bool),
		adj:   make(map[string][]edge),
	}
}

func (g *Graph) HasNode(node string) bool {
	return g.nodes[node]
}

func (g *Graph) AddNode(node string) {
	if !g.nodes[node] {
		g.nodes[node] = true
		g.adj[node] = []edge{}
	}
}

func (g *Graph) AddEdge(from, to string, weight float64) error {
	if weight <= 0 {
		return ErrNegativeWeight
	}
	g.AddNode(from)
	g.AddNode(to)
	g.adj[from] = append(g.adj[from], edge{to: to, weight: weight})
	return nil
}

func (g *Graph) IsEmpty() bool {
	return len(g.nodes) == 0
}

func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

type ShortestPathResult struct {
	TotalTime float64
	Path      []string
	EdgeCount int
}

type AllPathsResult struct {
	Distances map[string]float64
}

type item struct {
	node     string
	distance float64
	index    int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].distance < pq[j].distance
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (g *Graph) ShortestPath(start, end string) (*ShortestPathResult, error) {
	if g.IsEmpty() {
		return nil, ErrEmptyGraph
	}
	if !g.HasNode(start) {
		return nil, ErrNodeNotFound
	}
	if !g.HasNode(end) {
		return nil, ErrNodeNotFound
	}
	if start == end {
		return &ShortestPathResult{
			TotalTime: 0,
			Path:      []string{start},
			EdgeCount: 0,
		}, nil
	}

	dist := make(map[string]float64)
	prev := make(map[string]string)
	visited := make(map[string]bool)

	for node := range g.nodes {
		dist[node] = math.Inf(1)
	}
	dist[start] = 0

	pq := make(priorityQueue, 0)
	heap.Init(&pq)
	heap.Push(&pq, &item{node: start, distance: 0})

	for pq.Len() > 0 {
		curr := heap.Pop(&pq).(*item)
		u := curr.node

		if visited[u] {
			continue
		}
		visited[u] = true

		if u == end {
			break
		}

		for _, e := range g.adj[u] {
			v := e.to
			alt := dist[u] + e.weight

			if alt < dist[v] {
				dist[v] = alt
				prev[v] = u
				heap.Push(&pq, &item{node: v, distance: alt})
			}
		}
	}

	if math.IsInf(dist[end], 1) {
		return nil, ErrNoPath
	}

	path := []string{}
	for v := end; v != ""; v = prev[v] {
		path = append([]string{v}, path...)
		if v == start {
			break
		}
	}

	if path[0] != start {
		return nil, ErrNoPath
	}

	return &ShortestPathResult{
		TotalTime: dist[end],
		Path:      path,
		EdgeCount: len(path) - 1,
	}, nil
}

func (g *Graph) AllShortestPaths(start string) (*AllPathsResult, error) {
	if g.IsEmpty() {
		return nil, ErrEmptyGraph
	}
	if !g.HasNode(start) {
		return nil, ErrNodeNotFound
	}

	dist := make(map[string]float64)
	visited := make(map[string]bool)

	for node := range g.nodes {
		dist[node] = math.Inf(1)
	}
	dist[start] = 0

	pq := make(priorityQueue, 0)
	heap.Init(&pq)
	heap.Push(&pq, &item{node: start, distance: 0})

	for pq.Len() > 0 {
		curr := heap.Pop(&pq).(*item)
		u := curr.node

		if visited[u] {
			continue
		}
		visited[u] = true

		for _, e := range g.adj[u] {
			v := e.to
			alt := dist[u] + e.weight

			if alt < dist[v] {
				dist[v] = alt
				heap.Push(&pq, &item{node: v, distance: alt})
			}
		}
	}

	result := &AllPathsResult{
		Distances: make(map[string]float64),
	}
	for node, d := range dist {
		if !math.IsInf(d, 1) {
			result.Distances[node] = d
		}
	}

	return result, nil
}
