package raft

type LogEntryInfo struct {
	Term    int
	Command string
}

type NodeInfo struct {
	ID              int
	State           string
	CurrentTerm     int
	VotedFor        int
	CommitIndex     int
	LastApplied     int
	LogEntries      []LogEntryInfo
	LeaderID        int
	LastHeartbeatAt int64
}

func (rn *RaftNode) GetInfo() NodeInfo {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	logEntries := make([]LogEntryInfo, len(rn.log))
	for i, e := range rn.log {
		logEntries[i] = LogEntryInfo{
			Term:    e.Term,
			Command: e.Command,
		}
	}

	return NodeInfo{
		ID:              rn.id,
		State:           string(rn.state),
		CurrentTerm:     rn.currentTerm,
		VotedFor:        rn.votedFor,
		CommitIndex:     rn.commitIndex,
		LastApplied:     rn.lastApplied,
		LogEntries:      logEntries,
		LeaderID:        rn.leaderID,
		LastHeartbeatAt: rn.lastHeartbeatAt,
	}
}

func (c *Cluster) GetAllNodeInfos() map[int]NodeInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	infos := make(map[int]NodeInfo)
	for id, node := range c.nodes {
		infos[id] = node.GetInfo()
	}
	return infos
}

func (c *Cluster) GetNodeInfo(id int) (NodeInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.nodes[id]
	if !exists {
		return NodeInfo{}, false
	}
	return node.GetInfo(), true
}

func (c *Cluster) NodeCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.nodes)
}
