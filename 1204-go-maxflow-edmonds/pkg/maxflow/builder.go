package maxflow

import "maxflow/pkg/api"

func BuildGraphFromRequest(req *api.FlowRequest) *Graph {
	g := NewGraph()

	for _, node := range req.Nodes {
		g.addNodeIfNotExists(node.ID)
	}

	for i, edge := range req.Edges {
		fromIdx := g.addNodeIfNotExists(edge.From)
		toIdx := g.addNodeIfNotExists(edge.To)

		capacity := edge.Capacity
		if capacity < 0 {
			capacity = 0
		}

		g.origEdges = append(g.origEdges, origEdge{
			from:     fromIdx,
			to:       toIdx,
			capacity: capacity,
			idx:      i,
		})

		g.addEdge(fromIdx, toIdx, capacity, i)
	}

	return g
}
