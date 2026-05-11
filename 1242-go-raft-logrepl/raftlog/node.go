package raftlog

import (
	"fmt"
	"raftlog/common"
	"sync"
)

type Node struct {
	mu      sync.RWMutex
	id      string
	address string
	entries []common.LogEntry
}

func NewNode(id, address string) *Node {
	return &Node{
		id:      id,
		address: address,
		entries: []common.LogEntry{},
	}
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Address() string {
	return n.address
}

func (n *Node) Entries() []common.LogEntry {
	n.mu.RLock()
	defer n.mu.RUnlock()
	copied := make([]common.LogEntry, len(n.entries))
	copy(copied, n.entries)
	return copied
}

func (n *Node) LastIndex() int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if len(n.entries) == 0 {
		return 0
	}
	return n.entries[len(n.entries)-1].Index
}

func (n *Node) LastTerm() int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if len(n.entries) == 0 {
		return 0
	}
	return n.entries[len(n.entries)-1].Term
}

func (n *Node) EntryCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.entries)
}

func (n *Node) EntryAt(index int64) (*common.LogEntry, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if index <= 0 || int(index) > len(n.entries) {
		return nil, false
	}
	entry := n.entries[index-1]
	return &entry, true
}

func (n *Node) Append(command string) common.LogEntry {
	n.mu.Lock()
	defer n.mu.Unlock()
	idx := int64(len(n.entries) + 1)
	entry := common.LogEntry{
		Index:   idx,
		Term:    1,
		Command: command,
	}
	n.entries = append(n.entries, entry)
	return entry
}

func (n *Node) AppendEntry(entry common.LogEntry) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	currentLen := int64(len(n.entries))
	if entry.Index <= currentLen {
		existing := n.entries[entry.Index-1]
		if existing.Term == entry.Term {
			return nil
		}
		n.entries = n.entries[:entry.Index-1]
	} else if entry.Index > currentLen+1 {
		return fmt.Errorf("gap detected: expected index %d, got %d", currentLen+1, entry.Index)
	}
	n.entries = append(n.entries, entry)
	return nil
}

func (n *Node) Reset() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.entries = []common.LogEntry{}
}
