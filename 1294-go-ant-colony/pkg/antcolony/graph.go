package antcolony

import (
	"fmt"
	"math"
)

type internalGraph struct {
	nodeIndices map[string]int
	nodes       []string
	adjList     []map[int]float64
	edges       []Edge
}

func (g *internalGraph) getNodeIndex(nodeID string) (int, bool) {
	idx, ok := g.nodeIndices[nodeID]
	return idx, ok
}

func (g *internalGraph) getNeighbors(nodeIdx int) []int {
	neighbors := make([]int, 0, len(g.adjList[nodeIdx]))
	for n := range g.adjList[nodeIdx] {
		neighbors = append(neighbors, n)
	}
	return neighbors
}

func (g *internalGraph) getEdgeWeight(fromIdx, toIdx int) (float64, bool) {
	weight, ok := g.adjList[fromIdx][toIdx]
	return weight, ok
}

func buildInternalGraph(graph Graph) (*internalGraph, error) {
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("graph has no nodes")
	}

	nodeIndices := make(map[string]int, len(graph.Nodes))
	nodes := make([]string, len(graph.Nodes))
	for i, node := range graph.Nodes {
		if node.ID == "" {
			return nil, fmt.Errorf("node has empty ID")
		}
		if _, exists := nodeIndices[node.ID]; exists {
			return nil, fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIndices[node.ID] = i
		nodes[i] = node.ID
	}

	adjList := make([]map[int]float64, len(nodes))
	for i := range adjList {
		adjList[i] = make(map[int]float64)
	}

	for _, edge := range graph.Edges {
		fromIdx, ok := nodeIndices[edge.From]
		if !ok {
			return nil, fmt.Errorf("edge references unknown node: %s", edge.From)
		}
		toIdx, ok := nodeIndices[edge.To]
		if !ok {
			return nil, fmt.Errorf("edge references unknown node: %s", edge.To)
		}
		if edge.Weight <= 0 {
			return nil, fmt.Errorf("edge weight must be positive: %s -> %s (weight: %v)", edge.From, edge.To, edge.Weight)
		}
		if math.IsInf(edge.Weight, 1) {
			return nil, fmt.Errorf("edge weight cannot be infinity: %s -> %s", edge.From, edge.To)
		}
		adjList[fromIdx][toIdx] = edge.Weight
	}

	return &internalGraph{
		nodeIndices: nodeIndices,
		nodes:       nodes,
		adjList:     adjList,
		edges:       graph.Edges,
	}, nil
}

func isConnected(g *internalGraph, start, end int) bool {
	if start == end {
		return true
	}

	visited := make([]bool, len(g.nodes))
	queue := []int{start}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for neighbor := range g.adjList[current] {
			if neighbor == end {
				return true
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return false
}

func validateParameters(params Parameters) error {
	if params.AntCount <= 0 {
		return fmt.Errorf("antCount must be positive")
	}
	if params.InitialPheromone <= 0 {
		return fmt.Errorf("initialPheromone must be positive")
	}
	if params.EvaporationRate < 0 || params.EvaporationRate > 1 {
		return fmt.Errorf("evaporationRate must be between 0 and 1")
	}
	if params.Alpha < 0 {
		return fmt.Errorf("alpha must be non-negative")
	}
	if params.Beta < 0 {
		return fmt.Errorf("beta must be non-negative")
	}
	if params.MaxIterations <= 0 {
		return fmt.Errorf("maxIterations must be positive")
	}
	if params.ConvergenceThreshold < 0 {
		return fmt.Errorf("convergenceThreshold must be non-negative")
	}
	if params.EliteWeight < 0 {
		return fmt.Errorf("eliteWeight must be non-negative")
	}
	return nil
}

func checkLargeGraph(g *internalGraph) bool {
	nodeCount := len(g.nodes)
	edgeCount := 0
	for _, adj := range g.adjList {
		edgeCount += len(adj)
	}
	return nodeCount > 1000 || edgeCount > 100000
}
