package spf

import (
	"container/heap"
	"math"
	"sync"
)

const (
	INFINITY = uint32(math.MaxUint32)
)

type SPFCalculator struct {
	topology     *Topology
	paths        map[string]map[string]*ShortestPath
	pathsMu      sync.RWMutex
	currentNodes map[string]struct{}
}

func NewSPFCalculator(topology *Topology) *SPFCalculator {
	return &SPFCalculator{
		topology:     topology,
		paths:        make(map[string]map[string]*ShortestPath),
		currentNodes: make(map[string]struct{}),
	}
}

func (spf *SPFCalculator) ComputeSPF(source string) error {
	if !spf.topology.NodeExists(source) {
		return &NodeNotFoundError{NodeID: source}
	}

	adjacency := spf.topology.GetAllAdjacency()
	nodes := spf.topology.GetNodes()
	nodeSet := make(map[string]struct{})
	for _, node := range nodes {
		nodeSet[node] = struct{}{}
	}

	distances := make(map[string]uint32)
	predecessors := make(map[string]map[string]struct{})
	visited := make(map[string]bool)

	for _, node := range nodes {
		distances[node] = INFINITY
		predecessors[node] = make(map[string]struct{})
		visited[node] = false
	}
	distances[source] = 0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{node: source, dist: 0})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*pqItem)
		currentNode := current.node
		currentDist := current.dist

		if visited[currentNode] {
			continue
		}
		visited[currentNode] = true

		if neighbors, ok := adjacency[currentNode]; ok {
			for neighbor, adj := range neighbors {
				newDist := currentDist + adj.Cost
				if newDist < currentDist {
					continue
				}

				if newDist < distances[neighbor] {
					distances[neighbor] = newDist
					predecessors[neighbor] = make(map[string]struct{})
					predecessors[neighbor][currentNode] = struct{}{}
					heap.Push(pq, &pqItem{node: neighbor, dist: newDist})
				} else if newDist == distances[neighbor] {
					predecessors[neighbor][currentNode] = struct{}{}
				}
			}
		}
	}

	result := make(map[string]*ShortestPath)
	for _, dest := range nodes {
		if dest == source {
			result[dest] = &ShortestPath{
				Destination: dest,
				Paths:       [][]string{{source}},
				TotalCost:   0,
				Valid:       true,
				Reachable:   true,
			}
			continue
		}

		dist := distances[dest]
		if dist == INFINITY {
			result[dest] = &ShortestPath{
				Destination: dest,
				Paths:       nil,
				TotalCost:   0,
				Valid:       false,
				Reachable:   false,
			}
		} else {
			paths := spf.findAllPaths(source, dest, predecessors)
			result[dest] = &ShortestPath{
				Destination: dest,
				Paths:       paths,
				TotalCost:   dist,
				Valid:       true,
				Reachable:   true,
			}
		}
	}

	spf.pathsMu.Lock()
	spf.paths[source] = result
	spf.currentNodes = nodeSet
	spf.pathsMu.Unlock()

	return nil
}

func (spf *SPFCalculator) findAllPaths(source, dest string, predecessors map[string]map[string]struct{}) [][]string {
	allPaths := make([][]string, 0)
	currentPath := []string{dest}
	spf.dfsPaths(source, dest, predecessors, currentPath, &allPaths)
	for i, path := range allPaths {
		for j, k := 0, len(path)-1; j < k; j, k = j+1, k-1 {
			path[j], path[k] = path[k], path[j]
		}
		allPaths[i] = path
	}
	return allPaths
}

func (spf *SPFCalculator) dfsPaths(source, current string, predecessors map[string]map[string]struct{}, path []string, allPaths *[][]string) {
	if current == source {
		pathCopy := make([]string, len(path))
		copy(pathCopy, path)
		*allPaths = append(*allPaths, pathCopy)
		return
	}

	predecessorMap, exists := predecessors[current]
	if !exists {
		return
	}

	for pred := range predecessorMap {
		newPath := append([]string{pred}, path...)
		spf.dfsPaths(source, pred, predecessors, newPath, allPaths)
	}
}

func (spf *SPFCalculator) GetShortestPaths(source string) ([]*ShortestPath, error) {
	spf.pathsMu.RLock()
	pathMap, exists := spf.paths[source]
	currentNodes := spf.currentNodes
	spf.pathsMu.RUnlock()

	if !exists {
		return nil, &SPFNotComputedError{NodeID: source}
	}

	currentTopologyNodes := make(map[string]struct{})
	for _, node := range spf.topology.GetNodes() {
		currentTopologyNodes[node] = struct{}{}
	}

	topologyChanged := false
	if len(currentTopologyNodes) != len(currentNodes) {
		topologyChanged = true
	} else {
		for node := range currentTopologyNodes {
			if _, exists := currentNodes[node]; !exists {
				topologyChanged = true
				break
			}
		}
	}

	result := make([]*ShortestPath, 0, len(pathMap))
	for _, sp := range pathMap {
		entry := &ShortestPath{
			Destination: sp.Destination,
			Paths:       make([][]string, len(sp.Paths)),
			TotalCost:   sp.TotalCost,
			Valid:       sp.Valid && !topologyChanged,
			Reachable:   sp.Reachable,
		}
		copy(entry.Paths, sp.Paths)

		if !topologyChanged && sp.Valid && sp.Reachable && sp.Paths != nil {
			for _, path := range sp.Paths {
				for _, node := range path {
					if _, exists := currentTopologyNodes[node]; !exists {
						entry.Valid = false
						break
					}
				}
				if !entry.Valid {
					break
				}
			}
		}

		result = append(result, entry)
	}

	return result, nil
}

type NodeNotFoundError struct {
	NodeID string
}

func (e *NodeNotFoundError) Error() string {
	return "node not found: " + e.NodeID
}

type SPFNotComputedError struct {
	NodeID string
}

func (e *SPFNotComputedError) Error() string {
	return "SPF not computed for source: " + e.NodeID
}
