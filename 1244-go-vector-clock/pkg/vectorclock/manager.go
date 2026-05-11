package vectorclock

import (
	"fmt"
	"sync"
)

type NodeManager struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes: make(map[string]*Node),
	}
}

func (m *NodeManager) GetOrCreate(nodeID string) *Node {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node, exists := m.nodes[nodeID]; exists {
		return node
	}

	node := NewNode(nodeID)
	m.nodes[nodeID] = node
	return node
}

func (m *NodeManager) Get(nodeID string) (*Node, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	return node, exists
}

func (m *NodeManager) Produce(nodeID, content string) (*Event, error) {
	node := m.GetOrCreate(nodeID)
	event := node.Produce(content)
	return event, nil
}

func (m *NodeManager) Receive(targetNodeID, fromNodeID string, remoteClock VectorClock, content string) (*Event, error) {
	node := m.GetOrCreate(targetNodeID)
	event, err := node.Receive(fromNodeID, remoteClock, content)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (m *NodeManager) GetClock(nodeID string) (VectorClock, error) {
	node, exists := m.Get(nodeID)
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node.CurrentClock(), nil
}

func (m *NodeManager) GetEvents(nodeID string) ([]*Event, error) {
	node, exists := m.Get(nodeID)
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node.Events(), nil
}

func (m *NodeManager) Prune(nodeID string, retainCount int) (int, error) {
	node, exists := m.Get(nodeID)
	if !exists {
		return 0, fmt.Errorf("node %s not found", nodeID)
	}
	if retainCount <= 0 {
		retainCount = 10
	}
	return node.Prune(retainCount), nil
}

func CompareClocks(a, b VectorClock) CausalityType {
	return Compare(a, b)
}
