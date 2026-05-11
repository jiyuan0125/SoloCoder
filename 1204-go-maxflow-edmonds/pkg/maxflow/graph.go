package maxflow

type residualEdge struct {
	to     int
	rev    int
	cap    int
	origIdx int
}

type Graph struct {
	nodes     []string
	nodeMap   map[string]int
	adj       [][]residualEdge
	origEdges []origEdge
}

type origEdge struct {
	from     int
	to       int
	capacity int
	idx      int
}

func NewGraph() *Graph {
	return &Graph{
		nodeMap:   make(map[string]int),
		origEdges: make([]origEdge, 0),
	}
}

func (g *Graph) addNodeIfNotExists(id string) int {
	if idx, ok := g.nodeMap[id]; ok {
		return idx
	}
	idx := len(g.nodes)
	g.nodes = append(g.nodes, id)
	g.nodeMap[id] = idx
	g.adj = append(g.adj, []residualEdge{})
	return idx
}

func (g *Graph) addEdge(fromIdx, toIdx, capacity, origIdx int) {
	forwardIdx := len(g.adj[fromIdx])
	backwardIdx := len(g.adj[toIdx])

	g.adj[fromIdx] = append(g.adj[fromIdx], residualEdge{
		to:      toIdx,
		rev:     backwardIdx,
		cap:     capacity,
		origIdx: origIdx,
	})

	g.adj[toIdx] = append(g.adj[toIdx], residualEdge{
		to:      fromIdx,
		rev:     forwardIdx,
		cap:     0,
		origIdx: -1,
	})
}
