package raftlog

import (
	"errors"
	"raftlog/common"
	"sync"
)

var (
	ErrNodeNotFound = errors.New("node not found")
)

type ConsistencyCheckResult struct {
	Consistent    bool
	MatchedCount  int
	Node1ID       string
	Node1Total    int
	Node2ID       string
	Node2Total    int
	MismatchIndex int64
	Node1Entry    *common.LogEntry
	Node2Entry    *common.LogEntry
}

type Cluster struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

func NewCluster(nodeInfos []common.NodeInfo) *Cluster {
	nodes := make(map[string]*Node)
	for _, info := range nodeInfos {
		nodes[info.ID] = NewNode(info.ID, info.Address)
	}
	return &Cluster{nodes: nodes}
}

func (c *Cluster) Submit(nodeID, command string) error {
	c.mu.RLock()
	src, ok := c.nodes[nodeID]
	c.mu.RUnlock()
	if !ok {
		return ErrNodeNotFound
	}
	entry := src.Append(command)
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, target := range c.nodes {
		if target.ID() == nodeID {
			continue
		}
		_ = target.AppendEntry(entry)
	}
	return nil
}

func (c *Cluster) GetLog(nodeID string) ([]common.LogEntry, error) {
	c.mu.RLock()
	node, ok := c.nodes[nodeID]
	c.mu.RUnlock()
	if !ok {
		return nil, ErrNodeNotFound
	}
	return node.Entries(), nil
}

func (c *Cluster) Reset(nodeID string) error {
	c.mu.RLock()
	node, ok := c.nodes[nodeID]
	c.mu.RUnlock()
	if !ok {
		return ErrNodeNotFound
	}
	node.Reset()
	return nil
}

func (c *Cluster) CheckConsistency(nodeID1, nodeID2 string) (ConsistencyCheckResult, error) {
	c.mu.RLock()
	n1, ok1 := c.nodes[nodeID1]
	n2, ok2 := c.nodes[nodeID2]
	c.mu.RUnlock()
	if !ok1 || !ok2 {
		return ConsistencyCheckResult{}, ErrNodeNotFound
	}
	return CheckLogsConsistency(n1, n2), nil
}

func (c *Cluster) AllNodeStatus() []common.NodeStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	statuses := make([]common.NodeStatus, 0, len(c.nodes))
	for _, node := range c.nodes {
		statuses = append(statuses, common.NodeStatus{
			ID:         node.ID(),
			Address:    node.Address(),
			EntryCount: node.EntryCount(),
			LastIndex:  node.LastIndex(),
			LastTerm:   node.LastTerm(),
			Entries:    node.Entries(),
		})
	}
	return statuses
}

func CheckLogsConsistency(n1, n2 *Node) ConsistencyCheckResult {
	e1 := n1.Entries()
	e2 := n2.Entries()
	minLen := len(e1)
	if len(e2) < minLen {
		minLen = len(e2)
	}
	for i := 0; i < minLen; i++ {
		a := e1[i]
		b := e2[i]
		if a.Index != b.Index || a.Term != b.Term || a.Command != b.Command {
			var ae, be *common.LogEntry
			ac := a
			bc := b
			ae = &ac
			be = &bc
			return ConsistencyCheckResult{
				Consistent:    false,
				MatchedCount:  i,
				Node1ID:       n1.ID(),
				Node1Total:    len(e1),
				Node2ID:       n2.ID(),
				Node2Total:    len(e2),
				MismatchIndex: a.Index,
				Node1Entry:    ae,
				Node2Entry:    be,
			}
		}
	}
	if len(e1) != len(e2) {
		mismatchIdx := int64(minLen + 1)
		var ae, be *common.LogEntry
		if len(e1) > minLen {
			ac := e1[minLen]
			ae = &ac
		}
		if len(e2) > minLen {
			bc := e2[minLen]
			be = &bc
		}
		return ConsistencyCheckResult{
			Consistent:    false,
			MatchedCount:  minLen,
			Node1ID:       n1.ID(),
			Node1Total:    len(e1),
			Node2ID:       n2.ID(),
			Node2Total:    len(e2),
			MismatchIndex: mismatchIdx,
			Node1Entry:    ae,
			Node2Entry:    be,
		}
	}
	return ConsistencyCheckResult{
		Consistent:   true,
		MatchedCount: len(e1),
		Node1ID:      n1.ID(),
		Node1Total:   len(e1),
		Node2ID:      n2.ID(),
		Node2Total:   len(e2),
	}
}
