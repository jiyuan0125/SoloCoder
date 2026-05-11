package euler

func (g *Graph) isConnectedExcluding(excludeEdges map[int]bool) bool {
	nodesWithEdges := g.getNodesWithEdges()
	if len(nodesWithEdges) == 0 {
		return true
	}

	var start Node
	for n := range nodesWithEdges {
		start = n
		break
	}

	visited := make(map[Node]bool)
	stack := []Node{start}
	visited[start] = true

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, edgeIdx := range g.adj[current] {
			if excludeEdges[edgeIdx] {
				continue
			}
			other, err := g.OtherEnd(edgeIdx, current)
			if err != nil {
				continue
			}
			if !visited[other] {
				visited[other] = true
				stack = append(stack, other)
			}
		}
	}

	for n := range nodesWithEdges {
		if !visited[n] {
			return false
		}
	}
	return true
}

func (g *Graph) isBridge(edgeIdx int, excludeEdges map[int]bool) bool {
	excludeEdges[edgeIdx] = true
	defer delete(excludeEdges, edgeIdx)

	e := g.Edges[edgeIdx]
	from := e.From
	to := e.To

	visited := make(map[Node]bool)
	stack := []Node{from}
	visited[from] = true

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if current == to {
			return false
		}

		for _, adjEdgeIdx := range g.adj[current] {
			if excludeEdges[adjEdgeIdx] {
				continue
			}
			other, err := g.OtherEnd(adjEdgeIdx, current)
			if err != nil {
				continue
			}
			if !visited[other] {
				visited[other] = true
				stack = append(stack, other)
			}
		}
	}

	return true
}

func (g *Graph) IsConnected() bool {
	return g.isConnectedExcluding(make(map[int]bool))
}

func (g *Graph) ValidateConnectivity() bool {
	if !g.hasEdges() {
		return true
	}
	return g.IsConnected()
}
