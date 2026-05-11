package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func NewCluster() *Cluster {
	return &Cluster{
		nodes: make(map[int]*RaftNode),
	}
}

func (c *Cluster) CreateNodes(count int) []int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ids []int
	for i := 0; i < count; i++ {
		id := len(c.nodes) + 1
		node := NewRaftNode(id)
		c.nodes[id] = node
		ids = append(ids, id)
	}

	c.setupPeers()
	return ids
}

func (c *Cluster) AddNode(id int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.nodes[id]; exists {
		return fmt.Errorf("node %d already exists", id)
	}

	node := NewRaftNode(id)
	c.nodes[id] = node
	c.setupPeers()
	return nil
}

func (c *Cluster) setupPeers() {
	var nodes []*RaftNode
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}

	for i, node := range nodes {
		var peers []*RaftNode
		var peerChans []chan interface{}

		for j, p := range nodes {
			if i != j {
				peers = append(peers, p)
				peerChans = append(peerChans, p.getPeerChan())
			}
		}

		node.setPeers(peers, peerChans)
	}
}

func (c *Cluster) StartAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.nodes {
		node.Start()
	}
}

func (c *Cluster) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.nodes {
		node.Stop()
	}
}

func (c *Cluster) GetNode(id int) *RaftNode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nodes[id]
}

func (c *Cluster) GetAllNodes() map[int]*RaftNode {
	c.mu.Lock()
	defer c.mu.Unlock()

	nodes := make(map[int]*RaftNode)
	for k, v := range c.nodes {
		nodes[k] = v
	}
	return nodes
}

func (c *Cluster) GetLeader() *RaftNode {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, node := range c.nodes {
		if node.getState() == Leader {
			return node
		}
	}
	return nil
}

func (c *Cluster) SubmitLog(command string) (int, int, error) {
	leader := c.GetLeader()
	if leader == nil {
		return 0, 0, fmt.Errorf("no leader available")
	}
	return leader.SubmitLog(command)
}

func NewRaftNode(id int) *RaftNode {
	return &RaftNode{
		id:          id,
		state:       Follower,
		currentTerm: 0,
		votedFor:    -1,
		log:         make([]LogEntry, 0),
		commitIndex: -1,
		lastApplied: -1,
		stopCh:      make(chan struct{}),
	}
}

func (rn *RaftNode) setPeers(peers []*RaftNode, peerChans []chan interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	oldPeerCount := len(rn.peers)
	rn.peers = peers
	rn.peerChans = peerChans

	if rn.state == Leader {
		newPeerCount := len(peers)
		if newPeerCount > oldPeerCount {
			newNextIndex := make([]int, newPeerCount)
			newMatchIndex := make([]int, newPeerCount)
			copy(newNextIndex, rn.nextIndex)
			copy(newMatchIndex, rn.matchIndex)
			for i := oldPeerCount; i < newPeerCount; i++ {
				newNextIndex[i] = len(rn.log)
				newMatchIndex[i] = -1
			}
			rn.nextIndex = newNextIndex
			rn.matchIndex = newMatchIndex
		}
	}
}

func (rn *RaftNode) getPeerChan() chan interface{} {
	return make(chan interface{}, 100)
}

func (rn *RaftNode) getState() NodeState {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.state
}

func (rn *RaftNode) setState(state NodeState) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.state = state
}

func (rn *RaftNode) getCurrentTerm() int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.currentTerm
}

func (rn *RaftNode) Start() {
	go rn.run()
}

func (rn *RaftNode) Stop() {
	close(rn.stopCh)
}

func (rn *RaftNode) resetElectionTimer() {
	timeout := ElectionTimeoutMin + time.Duration(rand.Int63n(int64(ElectionTimeoutMax-ElectionTimeoutMin)))
	if rn.electionTimer != nil {
		rn.electionTimer.Stop()
	}
	rn.electionTimer = time.NewTimer(timeout)
}

func (rn *RaftNode) run() {
	rn.resetElectionTimer()

	for {
		select {
		case <-rn.stopCh:
			return
		case <-rn.electionTimer.C:
			if rn.getState() != Leader {
				rn.startElection()
			}
		}
	}
}

func (rn *RaftNode) startElection() {
	rn.mu.Lock()
	rn.currentTerm++
	rn.state = Candidate
	rn.votedFor = rn.id
	votes := 1
	term := rn.currentTerm
	lastLogIndex := len(rn.log) - 1
	lastLogTerm := 0
	if lastLogIndex >= 0 {
		lastLogTerm = rn.log[lastLogIndex].Term
	}
	rn.mu.Unlock()

	rn.resetElectionTimer()

	var wg sync.WaitGroup
	voteCh := make(chan bool, len(rn.peers))

	for i, peer := range rn.peers {
		wg.Add(1)
		go func(p *RaftNode, idx int) {
			defer wg.Done()
			args := RequestVoteArgs{
				Term:         term,
				CandidateID:  rn.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			reply := p.RequestVote(args)

			rn.mu.Lock()
			if reply.Term > rn.currentTerm {
				rn.currentTerm = reply.Term
				rn.state = Follower
				rn.votedFor = -1
				rn.mu.Unlock()
				voteCh <- false
				return
			}
			rn.mu.Unlock()

			voteCh <- reply.VoteGranted
		}(peer, i)
	}

	go func() {
		wg.Wait()
		close(voteCh)
	}()

	electionTimeout := time.NewTimer(ElectionTimeoutMax)
	defer electionTimeout.Stop()

	for {
		select {
		case v, ok := <-voteCh:
			if !ok {
				if rn.getState() == Candidate {
					rn.setState(Follower)
				}
				return
			}
			if v {
				votes++
				if votes*2 > len(rn.peers)+1 {
					rn.becomeLeader()
					return
				}
			}
		case <-electionTimeout.C:
			if rn.getState() == Candidate {
				rn.setState(Follower)
			}
			return
		}
	}
}

func (rn *RaftNode) becomeLeader() {
	rn.mu.Lock()
	rn.state = Leader
	rn.leaderID = rn.id
	peerCount := len(rn.peers)
	rn.nextIndex = make([]int, peerCount)
	rn.matchIndex = make([]int, peerCount)
	for i := range rn.nextIndex {
		rn.nextIndex[i] = len(rn.log)
		rn.matchIndex[i] = -1
	}
	rn.mu.Unlock()

	go rn.sendHeartbeats()
}

func (rn *RaftNode) sendHeartbeats() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rn.stopCh:
			return
		case <-ticker.C:
			if rn.getState() != Leader {
				return
			}
			rn.sendAppendEntriesToAll(false)
		}
	}
}

func (rn *RaftNode) sendAppendEntriesToAll(isHeartbeat bool) {
	rn.mu.Lock()
	leaderCommit := rn.commitIndex
	peers := rn.peers
	rn.mu.Unlock()

	for i, peer := range peers {
		go func(p *RaftNode, idx int) {
			for {
				rn.mu.Lock()
				if rn.state != Leader {
					rn.mu.Unlock()
					return
				}

				ni := rn.nextIndex[idx]
				prevLogIndex := ni - 1
				prevLogTerm := 0
				if prevLogIndex >= 0 && prevLogIndex < len(rn.log) {
					prevLogTerm = rn.log[prevLogIndex].Term
				}

				var entries []LogEntry
				if !isHeartbeat && ni <= len(rn.log) {
					entries = rn.log[ni:]
				}

				args := AppendEntriesArgs{
					Term:         rn.currentTerm,
					LeaderID:     rn.id,
					PrevLogIndex: prevLogIndex,
					PrevLogTerm:  prevLogTerm,
					Entries:      entries,
					LeaderCommit: leaderCommit,
				}
				rn.mu.Unlock()

				reply := p.AppendEntries(args)

				rn.mu.Lock()
				if reply.Term > rn.currentTerm {
					rn.currentTerm = reply.Term
					rn.state = Follower
					rn.votedFor = -1
					rn.mu.Unlock()
					return
				}

				if rn.state != Leader {
					rn.mu.Unlock()
					return
				}

				if reply.Success {
					rn.matchIndex[idx] = prevLogIndex + len(entries)
					rn.nextIndex[idx] = rn.matchIndex[idx] + 1
					rn.updateCommitIndex()
					rn.mu.Unlock()
					return
				} else {
					if reply.XTerm != -1 {
						conflictIndex := -1
						for j := len(rn.log) - 1; j >= 0; j-- {
							if rn.log[j].Term == reply.XTerm {
								conflictIndex = j
								break
							}
						}
						if conflictIndex != -1 {
							rn.nextIndex[idx] = conflictIndex + 1
						} else {
							rn.nextIndex[idx] = reply.XIndex
						}
					} else {
						rn.nextIndex[idx] = reply.XIndex
					}

					if rn.nextIndex[idx] < 0 {
						rn.nextIndex[idx] = 0
					}
					rn.mu.Unlock()

					if isHeartbeat {
						return
					}
				}
			}
		}(peer, i)
	}
}

func (rn *RaftNode) updateCommitIndex() {
	n := len(rn.log)
	for n = len(rn.log) - 1; n > rn.commitIndex; n-- {
		count := 1
		for i := range rn.peers {
			if rn.matchIndex[i] >= n {
				count++
			}
		}
		if count*2 > len(rn.peers)+1 && rn.log[n].Term == rn.currentTerm {
			break
		}
	}
	if n > rn.commitIndex {
		rn.commitIndex = n
	}
}

func (rn *RaftNode) RequestVote(args RequestVoteArgs) RequestVoteReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply := RequestVoteReply{
		Term:        rn.currentTerm,
		VoteGranted: false,
	}

	if args.Term < rn.currentTerm {
		return reply
	}

	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.state = Follower
		rn.votedFor = -1
	}

	lastLogIndex := len(rn.log) - 1
	lastLogTerm := 0
	if lastLogIndex >= 0 {
		lastLogTerm = rn.log[lastLogIndex].Term
	}

	logOk := (args.LastLogTerm > lastLogTerm) ||
		(args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)

	if (rn.votedFor == -1 || rn.votedFor == args.CandidateID) && logOk {
		rn.votedFor = args.CandidateID
		reply.VoteGranted = true
		rn.resetElectionTimer()
	}

	reply.Term = rn.currentTerm
	return reply
}

func (rn *RaftNode) AppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply := AppendEntriesReply{
		Term:    rn.currentTerm,
		Success: false,
		XTerm:   -1,
		XIndex:  0,
		XLen:    len(rn.log),
	}

	if args.Term < rn.currentTerm {
		return reply
	}

	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.votedFor = -1
	}

	rn.state = Follower
	rn.leaderID = args.LeaderID
	rn.lastHeartbeatAt = time.Now().UnixNano()
	rn.resetElectionTimer()

	if args.PrevLogIndex >= 0 {
		if args.PrevLogIndex >= len(rn.log) {
			reply.XIndex = len(rn.log)
			return reply
		}

		if rn.log[args.PrevLogIndex].Term != args.PrevLogTerm {
			reply.XTerm = rn.log[args.PrevLogIndex].Term
			for i := args.PrevLogIndex; i >= 0; i-- {
				if rn.log[i].Term != reply.XTerm {
					reply.XIndex = i + 1
					break
				}
				if i == 0 {
					reply.XIndex = 0
				}
			}
			return reply
		}
	}

	for i, entry := range args.Entries {
		logIndex := args.PrevLogIndex + 1 + i
		if logIndex >= len(rn.log) {
			rn.log = append(rn.log, entry)
		} else {
			if rn.log[logIndex].Term != entry.Term {
				rn.log = rn.log[:logIndex]
				rn.log = append(rn.log, entry)
			}
		}
	}

	if args.LeaderCommit > rn.commitIndex {
		lastNewIndex := args.PrevLogIndex + len(args.Entries)
		if args.LeaderCommit < lastNewIndex {
			rn.commitIndex = args.LeaderCommit
		} else {
			rn.commitIndex = lastNewIndex
		}
	}

	reply.Success = true
	reply.Term = rn.currentTerm
	return reply
}

func (rn *RaftNode) SubmitLog(command string) (int, int, error) {
	rn.mu.Lock()
	if rn.state != Leader {
		rn.mu.Unlock()
		return 0, 0, fmt.Errorf("not a leader")
	}

	entry := LogEntry{
		Term:    rn.currentTerm,
		Command: command,
	}
	rn.log = append(rn.log, entry)
	index := len(rn.log) - 1
	term := rn.currentTerm
	rn.mu.Unlock()

	rn.sendAppendEntriesToAll(false)

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return term, index, fmt.Errorf("timeout waiting for commit")
		case <-ticker.C:
			rn.mu.Lock()
			if rn.commitIndex >= index {
				rn.mu.Unlock()
				return term, index, nil
			}
			rn.mu.Unlock()
		}
	}
}
