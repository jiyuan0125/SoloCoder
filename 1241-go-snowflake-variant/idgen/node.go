package idgen

import (
	"sync"

	"idgen/api"
)

type Node struct {
	NodeID       uint16
	Name         string
	Remark       string
	RegisteredAt int64
	Generator    *IDGenerator
}

type NodeManager struct {
	mu      sync.RWMutex
	nodes   map[uint16]*Node
	names   map[string]uint16
	nextID  uint16
	maxID   uint16
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes:  make(map[uint16]*Node),
		names:  make(map[string]uint16),
		nextID: api.ReservedNodeID + 1,
		maxID:  api.MaxNodeID,
	}
}

func (m *NodeManager) RegisterNode(name, remark string, registeredAt int64) (uint16, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.names[name]; exists {
		return 0, ErrNodeAlreadyExists
	}

	nodeID, err := m.allocateNodeID()
	if err != nil {
		return 0, err
	}

	node := &Node{
		NodeID:       nodeID,
		Name:         name,
		Remark:       remark,
		RegisteredAt: registeredAt,
		Generator:    NewIDGenerator(nodeID),
	}

	m.nodes[nodeID] = node
	m.names[name] = nodeID

	return nodeID, nil
}

func (m *NodeManager) UnregisterNode(nodeID uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, exists := m.nodes[nodeID]
	if !exists {
		return ErrNodeNotRegistered
	}

	delete(m.nodes, nodeID)
	delete(m.names, node.Name)

	return nil
}

func (m *NodeManager) GetNode(nodeID uint16) (*Node, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, exists := m.nodes[nodeID]
	return node, exists
}

func (m *NodeManager) ListNodes() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		list = append(list, node)
	}
	return list
}

func (m *NodeManager) allocateNodeID() (uint16, error) {
	start := m.nextID
	for {
		if _, exists := m.nodes[m.nextID]; !exists {
			nodeID := m.nextID
			m.nextID = (m.nextID + 1) % (m.maxID + 1)
			if m.nextID == api.ReservedNodeID {
				m.nextID = api.ReservedNodeID + 1
			}
			return nodeID, nil
		}

		m.nextID = (m.nextID + 1) % (m.maxID + 1)
		if m.nextID == api.ReservedNodeID {
			m.nextID = api.ReservedNodeID + 1
		}

		if m.nextID == start {
			return 0, ErrNoAvailableNodeID
		}
	}
}
