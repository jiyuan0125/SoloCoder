package spf

import (
	"sync"
)

type Topology struct {
	mu        sync.RWMutex
	nodes     map[string]struct{}
	adjacency map[string]map[string]Adjacency
}

func NewTopology() *Topology {
	return &Topology{
		nodes:     make(map[string]struct{}),
		adjacency: make(map[string]map[string]Adjacency),
	}
}

func (t *Topology) AddNode(nodeID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.nodes[nodeID]; !exists {
		t.nodes[nodeID] = struct{}{}
		if _, ok := t.adjacency[nodeID]; !ok {
			t.adjacency[nodeID] = make(map[string]Adjacency)
		}
	}
}

func (t *Topology) RemoveNode(nodeID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.nodes[nodeID]; !exists {
		return
	}
	delete(t.nodes, nodeID)
	for from := range t.adjacency {
		delete(t.adjacency[from], nodeID)
	}
	delete(t.adjacency, nodeID)
}

func (t *Topology) NodeExists(nodeID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.nodes[nodeID]
	return exists
}

func (t *Topology) AddLink(link Link) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.nodes[link.From]; !exists {
		t.nodes[link.From] = struct{}{}
	}
	if _, exists := t.nodes[link.To]; !exists {
		t.nodes[link.To] = struct{}{}
	}
	if _, ok := t.adjacency[link.From]; !ok {
		t.adjacency[link.From] = make(map[string]Adjacency)
	}
	existing, exists := t.adjacency[link.From][link.To]
	if !exists || IsNewerOrEqual(existing.SeqNum, link.SeqNum) {
		t.adjacency[link.From][link.To] = Adjacency{
			Neighbor: link.To,
			Cost:     link.Cost,
			SeqNum:   link.SeqNum,
		}
	}
}

func (t *Topology) RemoveLink(link Link) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if neighbors, ok := t.adjacency[link.From]; ok {
		delete(neighbors, link.To)
	}
	if neighbors, ok := t.adjacency[link.To]; ok {
		delete(neighbors, link.From)
	}
}

func (t *Topology) GetNodes() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	nodes := make([]string, 0, len(t.nodes))
	for id := range t.nodes {
		nodes = append(nodes, id)
	}
	return nodes
}

func (t *Topology) GetAdjacency(nodeID string) []Adjacency {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if neighbors, ok := t.adjacency[nodeID]; ok {
		result := make([]Adjacency, 0, len(neighbors))
		for _, adj := range neighbors {
			result = append(result, adj)
		}
		return result
	}
	return nil
}

func (t *Topology) GetAllAdjacency() map[string]map[string]Adjacency {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make(map[string]map[string]Adjacency)
	for node, neighbors := range t.adjacency {
		if _, exists := t.nodes[node]; !exists {
			continue
		}
		result[node] = make(map[string]Adjacency)
		for neighbor, adj := range neighbors {
			if _, exists := t.nodes[neighbor]; exists {
				result[node][neighbor] = adj
			}
		}
	}
	return result
}

func (t *Topology) Empty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.nodes) == 0
}
