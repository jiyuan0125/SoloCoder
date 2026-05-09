package gossip

import (
	"sync"
)

type NodeStatus string

const (
	StatusAlive   NodeStatus = "alive"
	StatusSuspect NodeStatus = "suspect"
	StatusDead    NodeStatus = "dead"
)

type VersionedValue struct {
	Value   string
	Version int64
}

type Node struct {
	ID           string
	Status       NodeStatus
	Storage      map[string]*VersionedValue
	mu           sync.RWMutex
	pingFailures int
	suspectRounds int
	MaxPingFailures int
	MaxSuspectRounds int
}

func NewNode(id string) *Node {
	return &Node{
		ID:               id,
		Status:           StatusAlive,
		Storage:          make(map[string]*VersionedValue),
		MaxPingFailures:  3,
		MaxSuspectRounds: 5,
	}
}

func (n *Node) SetStatus(status NodeStatus) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Status = status
}

func (n *Node) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Status
}

func (n *Node) Get(key string) (string, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if v, ok := n.Storage[key]; ok {
		return v.Value, true
	}
	return "", false
}

func (n *Node) Put(key string, value string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if existing, ok := n.Storage[key]; ok {
		existing.Value = value
		existing.Version++
	} else {
		n.Storage[key] = &VersionedValue{
			Value:   value,
			Version: 1,
		}
	}
}

func (n *Node) PutWithVersion(key string, value string, version int64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if existing, ok := n.Storage[key]; ok {
		if version > existing.Version {
			existing.Value = value
			existing.Version = version
		}
	} else {
		n.Storage[key] = &VersionedValue{
			Value:   value,
			Version: version,
		}
	}
}

func (n *Node) GetStorage() map[string]*VersionedValue {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make(map[string]*VersionedValue)
	for k, v := range n.Storage {
		result[k] = &VersionedValue{
			Value:   v.Value,
			Version: v.Version,
		}
	}
	return result
}

func (n *Node) GetDigest() map[string]int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make(map[string]int64)
	for k, v := range n.Storage {
		result[k] = v.Version
	}
	return result
}

func (n *Node) RecordPingSuccess() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pingFailures = 0
	n.suspectRounds = 0
	if n.Status == StatusSuspect {
		n.Status = StatusAlive
	}
}

func (n *Node) RecordPingFailure() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pingFailures++
	if n.Status == StatusAlive && n.pingFailures >= n.MaxPingFailures {
		n.Status = StatusSuspect
		n.suspectRounds = 0
	}
}

func (n *Node) IncrementSuspectRound() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.Status == StatusSuspect {
		n.suspectRounds++
		if n.suspectRounds >= n.MaxSuspectRounds {
			n.Status = StatusDead
		}
	}
}

func (n *Node) SimulateFailure() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pingFailures = n.MaxPingFailures
	n.Status = StatusSuspect
	n.suspectRounds = 0
}

func (n *Node) SimulateRecovery() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pingFailures = 0
	n.suspectRounds = 0
	n.Status = StatusAlive
}
