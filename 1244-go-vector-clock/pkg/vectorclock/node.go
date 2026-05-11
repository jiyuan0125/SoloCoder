package vectorclock

import (
	"fmt"
	"sync"
	"time"
)

type Node struct {
	mu              sync.RWMutex
	nodeID          string
	clock           VectorClock
	events          []*Event
	nextEventID     uint64
	pruneThreshold  int
	historicalMax   uint64
	historicalNodes map[string]struct{}
}

func NewNode(nodeID string) *Node {
	return NewNodeWithThreshold(nodeID, DefaultPruneThreshold)
}

func NewNodeWithThreshold(nodeID string, pruneThreshold int) *Node {
	return &Node{
		nodeID:          nodeID,
		clock:           make(VectorClock),
		events:          []*Event{},
		nextEventID:     1,
		pruneThreshold:  pruneThreshold,
		historicalMax:   0,
		historicalNodes: make(map[string]struct{}),
	}
}

func (n *Node) NodeID() string {
	return n.nodeID
}

func (n *Node) CurrentClock() VectorClock {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.copyClock(n.clock)
}

func (n *Node) Events() []*Event {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]*Event, len(n.events))
	for i, e := range n.events {
		result[i] = &Event{
			EventID:      e.EventID,
			NodeID:       e.NodeID,
			Clock:        n.copyClock(e.Clock),
			Content:      e.Content,
			CreationTime: e.CreationTime,
		}
	}
	return result
}

func (n *Node) GetEvent(eventID uint64) (*Event, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, e := range n.events {
		if e.EventID == eventID {
			return &Event{
				EventID:      e.EventID,
				NodeID:       e.NodeID,
				Clock:        n.copyClock(e.Clock),
				Content:      e.Content,
				CreationTime: e.CreationTime,
			}, true
		}
	}
	return nil, false
}

func (n *Node) Produce(content string) *Event {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.incrementSelf()

	event := &Event{
		EventID:      n.nextEventID,
		NodeID:       n.nodeID,
		Clock:        n.copyClock(n.clock),
		Content:      content,
		CreationTime: time.Now(),
	}
	n.nextEventID++
	n.events = append(n.events, event)
	return event
}

func (n *Node) Receive(fromNodeID string, remoteClock VectorClock, content string) (*Event, error) {
	if fromNodeID == n.nodeID {
		return nil, fmt.Errorf("cannot receive from self")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.mergeRemoteClock(fromNodeID, remoteClock)
	n.incrementSelf()

	event := &Event{
		EventID:      n.nextEventID,
		NodeID:       n.nodeID,
		Clock:        n.copyClock(n.clock),
		Content:      content,
		CreationTime: time.Now(),
	}
	n.nextEventID++
	n.events = append(n.events, event)
	return event, nil
}

func (n *Node) incrementSelf() {
	n.clock[n.nodeID]++
}

func (n *Node) mergeRemoteClock(fromNodeID string, remoteClock VectorClock) {
	for node, value := range remoteClock {
		if node == HistoricalKey {
			continue
		}
		if _, exists := n.historicalNodes[node]; exists {
			if value > n.historicalMax {
				n.historicalMax = value
			}
			if _, isActive := n.clock[node]; !isActive {
				n.clock[node] = value
				delete(n.historicalNodes, node)
			}
		}
		if current, exists := n.clock[node]; exists {
			if value > current {
				n.clock[node] = value
			}
		} else {
			n.clock[node] = value
		}
	}

	if historicalVal, exists := remoteClock[HistoricalKey]; exists {
		if historicalVal > n.historicalMax {
			n.historicalMax = historicalVal
		}
	}
}

func (n *Node) copyClock(c VectorClock) VectorClock {
	result := make(VectorClock, len(c)+1)
	for k, v := range c {
		result[k] = v
	}
	if n.historicalMax > 0 {
		result[HistoricalKey] = n.historicalMax
	}
	return result
}

func (n *Node) Prune(retainCount int) int {
	n.mu.Lock()
	defer n.mu.Unlock()

	activeCount := len(n.clock)
	if activeCount <= n.pruneThreshold {
		return 0
	}

	recentNodes := n.getRecentNodes()
	if len(recentNodes) <= retainCount {
		return 0
	}

	toPrune := recentNodes[:len(recentNodes)-retainCount]
	prunedCount := 0

	for _, node := range toPrune {
		if node == n.nodeID {
			continue
		}
		if value, exists := n.clock[node]; exists {
			if value > n.historicalMax {
				n.historicalMax = value
			}
			delete(n.clock, node)
			n.historicalNodes[node] = struct{}{}
			prunedCount++
		}
	}

	return prunedCount
}

func (n *Node) getRecentNodes() []string {
	result := make([]string, 0, len(n.clock))
	for node := range n.clock {
		if node == HistoricalKey {
			continue
		}
		result = append(result, node)
	}
	return result
}

func (n *Node) HistoricalMax() uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.historicalMax
}

func (n *Node) HistoricalNodes() map[string]struct{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make(map[string]struct{}, len(n.historicalNodes))
	for k, v := range n.historicalNodes {
		result[k] = v
	}
	return result
}
