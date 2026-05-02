package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type Role string

const (
	Follower  Role = "Follower"
	Candidate Role = "Candidate"
	Leader    Role = "Leader"
)

const (
	HeartbeatInterval = 50 * time.Millisecond
	MinElectionTimeout = 150 * time.Millisecond
	MaxElectionTimeout = 300 * time.Millisecond
	SnapshotThreshold = 100
)

type LogEntry struct {
	Term    int
	Index   int
	Command string
	Key     string
	Value   string
	Committed bool
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term      int
	Success   bool
	NextIndex int
}

type InstallSnapshotArgs struct {
	Term              int
	LeaderID          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

type InstallSnapshotReply struct {
	Term int
}

type RaftNode struct {
	mu sync.Mutex

	id          int
	port        int
	peers       map[int]string
	currentTerm int
	votedFor    int
	log         []LogEntry

	commitIndex int
	lastApplied int

	nextIndex  map[int]int
	matchIndex map[int]int

	state       Role
	stateMachine *StateMachine

	electionTimeout time.Duration
	lastHeartbeat   time.Time

	stopCh chan struct{}
}

var electionTimeouts = []time.Duration{
	150 * time.Millisecond,
	180 * time.Millisecond,
	210 * time.Millisecond,
	240 * time.Millisecond,
	270 * time.Millisecond,
}

func NewRaftNode(id int, port int, peers map[int]string) *RaftNode {
	node := &RaftNode{
		id:          id,
		port:        port,
		peers:       peers,
		currentTerm: 0,
		votedFor:    -1,
		log:         make([]LogEntry, 0),
		commitIndex: -1,
		lastApplied: -1,
		nextIndex:   make(map[int]int),
		matchIndex:  make(map[int]int),
		state:       Follower,
		stateMachine: NewStateMachine(),
		stopCh:      make(chan struct{}),
	}

	node.electionTimeout = electionTimeouts[id-1]
	node.lastHeartbeat = time.Now()

	for peerID := range peers {
		if peerID != id {
			node.nextIndex[peerID] = 0
			node.matchIndex[peerID] = -1
		}
	}

	return node
}

func (rn *RaftNode) Start() {
	go rn.run()
}

func (rn *RaftNode) Stop() {
	close(rn.stopCh)
}

func (rn *RaftNode) run() {
	for {
		select {
		case <-rn.stopCh:
			return
		default:
			rn.mu.Lock()
			state := rn.state
			rn.mu.Unlock()

			switch state {
			case Follower:
				rn.runFollower()
			case Candidate:
				rn.runCandidate()
			case Leader:
				rn.runLeader()
			}
		}
	}
}

func (rn *RaftNode) runFollower() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-rn.stopCh:
			return
		case <-ticker.C:
			rn.mu.Lock()
			if rn.state != Follower {
				rn.mu.Unlock()
				return
			}

			if time.Since(rn.lastHeartbeat) > rn.electionTimeout {
				rn.convertToCandidate()
			}
			rn.mu.Unlock()
		}
	}
}

func (rn *RaftNode) convertToCandidate() {
	rn.state = Candidate
	rn.currentTerm++
	rn.votedFor = rn.id
	rn.lastHeartbeat = time.Now()

	log.Printf("Node %d (term %d) converting to Candidate", rn.id, rn.currentTerm)
	go rn.startElection()
}

func (rn *RaftNode) startElection() {
	rn.mu.Lock()
	term := rn.currentTerm
	lastLogIndex := -1
	lastLogTerm := 0
	if len(rn.log) > 0 {
		lastLogIndex = rn.log[len(rn.log)-1].Index
		lastLogTerm = rn.log[len(rn.log)-1].Term
	}
	rn.mu.Unlock()

	votes := 1
	voteCh := make(chan bool, len(rn.peers)-1)

	for peerID, peerAddr := range rn.peers {
		if peerID == rn.id {
			continue
		}

		go func(pid int, addr string) {
			args := &RequestVoteArgs{
				Term:         term,
				CandidateID:  rn.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}

			reply, err := SendRequestVote(addr, args)
			if err != nil {
				voteCh <- false
				return
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()

			if reply.Term > rn.currentTerm {
				rn.currentTerm = reply.Term
				rn.convertToFollower()
				voteCh <- false
				return
			}

			if reply.Term == term && reply.VoteGranted {
				voteCh <- true
			} else {
				voteCh <- false
			}
		}(peerID, peerAddr)
	}

	timeout := time.After(rn.electionTimeout)
	for i := 0; i < len(rn.peers)-1; i++ {
		select {
		case <-timeout:
			rn.mu.Lock()
			if rn.state == Candidate {
				rn.convertToFollower()
			}
			rn.mu.Unlock()
			return
		case granted := <-voteCh:
			if granted {
				votes++
				if votes >= len(rn.peers)/2+1 {
					rn.mu.Lock()
					if rn.state == Candidate {
						rn.convertToLeader()
					}
					rn.mu.Unlock()
					return
				}
			}
		}
	}
}

func (rn *RaftNode) convertToFollower() {
	rn.state = Follower
	rn.votedFor = -1
	rn.lastHeartbeat = time.Now()
	log.Printf("Node %d (term %d) converting to Follower", rn.id, rn.currentTerm)
}

func (rn *RaftNode) convertToLeader() {
	rn.state = Leader

	for peerID := range rn.peers {
		if peerID != rn.id {
			rn.nextIndex[peerID] = len(rn.log)
			rn.matchIndex[peerID] = -1
		}
	}

	log.Printf("Node %d (term %d) elected as Leader", rn.id, rn.currentTerm)
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
			rn.mu.Lock()
			if rn.state != Leader {
				rn.mu.Unlock()
				return
			}

			for peerID, peerAddr := range rn.peers {
				if peerID == rn.id {
					continue
				}

				go rn.sendAppendEntries(peerID, peerAddr)
			}
			rn.mu.Unlock()
		}
	}
}

func (rn *RaftNode) sendAppendEntries(peerID int, peerAddr string) {
	rn.mu.Lock()
	if rn.state != Leader {
		rn.mu.Unlock()
		return
	}

	nextIdx := rn.nextIndex[peerID]
	prevLogIndex := nextIdx - 1
	prevLogTerm := 0
	if prevLogIndex >= 0 && prevLogIndex < len(rn.log) {
		prevLogTerm = rn.log[prevLogIndex].Term
	}

	var entries []LogEntry
	if nextIdx < len(rn.log) {
		entries = rn.log[nextIdx:]
	}

	args := &AppendEntriesArgs{
		Term:         rn.currentTerm,
		LeaderID:     rn.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: rn.commitIndex,
	}
	rn.mu.Unlock()

	reply, err := SendAppendEntries(peerAddr, args)
	if err != nil {
		return
	}

	rn.mu.Lock()
	defer rn.mu.Unlock()

	if reply.Term > rn.currentTerm {
		rn.currentTerm = reply.Term
		rn.convertToFollower()
		return
	}

	if rn.state != Leader || reply.Term != rn.currentTerm {
		return
	}

	if reply.Success {
		if len(entries) > 0 {
			rn.matchIndex[peerID] = entries[len(entries)-1].Index
			rn.nextIndex[peerID] = rn.matchIndex[peerID] + 1
		}

		rn.checkCommit()
	} else {
		if reply.NextIndex >= 0 {
			rn.nextIndex[peerID] = reply.NextIndex
		} else if rn.nextIndex[peerID] > 0 {
			rn.nextIndex[peerID]--
		}
	}
}

func (rn *RaftNode) checkCommit() {
	for i := rn.commitIndex + 1; i < len(rn.log); i++ {
		if rn.log[i].Term != rn.currentTerm {
			continue
		}

		matchCount := 1
		for peerID := range rn.peers {
			if peerID == rn.id {
				continue
			}
			if rn.matchIndex[peerID] >= i {
				matchCount++
			}
		}

		if matchCount >= len(rn.peers)/2+1 {
			rn.log[i].Committed = true
			rn.commitIndex = i

			rn.applyLog()
		}
	}
}

func (rn *RaftNode) applyLog() {
	for rn.lastApplied < rn.commitIndex {
		rn.lastApplied++
		entry := &rn.log[rn.lastApplied]
		rn.stateMachine.Apply(entry)
		log.Printf("Node %d applied log entry %d: %s %s=%s", rn.id, entry.Index, entry.Command, entry.Key, entry.Value)
	}

	rn.checkSnapshot()
}

func (rn *RaftNode) checkSnapshot() {
	if len(rn.log) > SnapshotThreshold {
		go rn.createSnapshot()
	}
}

func (rn *RaftNode) createSnapshot() {
	rn.mu.Lock()
	if len(rn.log) == 0 {
		rn.mu.Unlock()
		return
	}

	lastIncludedIndex := rn.lastApplied
	lastIncludedTerm := 0
	if lastIncludedIndex >= 0 && lastIncludedIndex < len(rn.log) {
		lastIncludedTerm = rn.log[lastIncludedIndex].Term
	}

	snapshotData, err := rn.stateMachine.CreateSnapshot()
	if err != nil {
		rn.mu.Unlock()
		log.Printf("Node %d failed to create snapshot: %v", rn.id, err)
		return
	}
	rn.mu.Unlock()

	rn.mu.Lock()
	defer rn.mu.Unlock()

	if lastIncludedIndex < 0 || lastIncludedIndex >= len(rn.log) {
		return
	}

	rn.log = rn.log[lastIncludedIndex+1:]
	for i := range rn.log {
		rn.log[i].Index = i
	}

	for peerID := range rn.peers {
		if peerID != rn.id {
			if rn.nextIndex[peerID] <= lastIncludedIndex+1 {
				rn.nextIndex[peerID] = 0
			} else {
				rn.nextIndex[peerID] -= lastIncludedIndex + 1
			}

			if rn.matchIndex[peerID] <= lastIncludedIndex {
				rn.matchIndex[peerID] = -1
			} else {
				rn.matchIndex[peerID] -= lastIncludedIndex + 1
			}
		}
	}

	if rn.commitIndex <= lastIncludedIndex {
		rn.commitIndex = -1
	} else {
		rn.commitIndex -= lastIncludedIndex + 1
	}

	if rn.lastApplied <= lastIncludedIndex {
		rn.lastApplied = -1
	} else {
		rn.lastApplied -= lastIncludedIndex + 1
	}

	log.Printf("Node %d created snapshot up to index %d (term %d)", rn.id, lastIncludedIndex, lastIncludedTerm)
	log.Printf("Snapshot data: %s", string(snapshotData))
}

func (rn *RaftNode) RequestVote(args *RequestVoteArgs) *RequestVoteReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply := &RequestVoteReply{
		Term:        rn.currentTerm,
		VoteGranted: false,
	}

	if args.Term < rn.currentTerm {
		return reply
	}

	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.convertToFollower()
		reply.Term = rn.currentTerm
	}

	if (rn.votedFor == -1 || rn.votedFor == args.CandidateID) && rn.isLogUpToDate(args.LastLogIndex, args.LastLogTerm) {
		rn.votedFor = args.CandidateID
		reply.VoteGranted = true
		rn.lastHeartbeat = time.Now()
		log.Printf("Node %d voted for Node %d in term %d", rn.id, args.CandidateID, rn.currentTerm)
	}

	return reply
}

func (rn *RaftNode) isLogUpToDate(candidateLastIndex int, candidateLastTerm int) bool {
	lastLogIndex := -1
	lastLogTerm := 0
	if len(rn.log) > 0 {
		lastLogIndex = rn.log[len(rn.log)-1].Index
		lastLogTerm = rn.log[len(rn.log)-1].Term
	}

	if candidateLastTerm != lastLogTerm {
		return candidateLastTerm > lastLogTerm
	}
	return candidateLastIndex >= lastLogIndex
}

func (rn *RaftNode) AppendEntries(args *AppendEntriesArgs) *AppendEntriesReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply := &AppendEntriesReply{
		Term:      rn.currentTerm,
		Success:   false,
		NextIndex: len(rn.log),
	}

	if args.Term < rn.currentTerm {
		return reply
	}

	rn.lastHeartbeat = time.Now()

	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.convertToFollower()
		reply.Term = rn.currentTerm
	}

	if rn.state == Candidate {
		rn.convertToFollower()
	}

	if args.PrevLogIndex >= 0 {
		if args.PrevLogIndex >= len(rn.log) {
			reply.NextIndex = len(rn.log)
			return reply
		}

		if rn.log[args.PrevLogIndex].Term != args.PrevLogTerm {
			for i := args.PrevLogIndex - 1; i >= 0; i-- {
				if rn.log[i].Term != rn.log[args.PrevLogIndex].Term {
					reply.NextIndex = i + 1
					return reply
				}
			}
			reply.NextIndex = 0
			return reply
		}
	}

	for i, entry := range args.Entries {
		idx := args.PrevLogIndex + 1 + i
		if idx < len(rn.log) {
			if rn.log[idx].Term != entry.Term {
				rn.log = rn.log[:idx]
				rn.log = append(rn.log, entry)
			}
		} else {
			rn.log = append(rn.log, entry)
		}
	}

	if len(args.Entries) > 0 {
		log.Printf("Node %d appended %d entries from Leader %d", rn.id, len(args.Entries), args.LeaderID)
	}

	if args.LeaderCommit > rn.commitIndex {
		newCommitIndex := args.LeaderCommit
		if len(rn.log) > 0 {
			lastLogIndex := rn.log[len(rn.log)-1].Index
			if newCommitIndex > lastLogIndex {
				newCommitIndex = lastLogIndex
			}
		}

		for i := rn.commitIndex + 1; i <= newCommitIndex; i++ {
			if i < len(rn.log) {
				rn.log[i].Committed = true
			}
		}
		rn.commitIndex = newCommitIndex
		rn.applyLog()
	}

	reply.Success = true
	reply.NextIndex = len(rn.log)
	return reply
}

func (rn *RaftNode) Set(key, value string) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != Leader {
		return &NotLeaderError{LeaderID: rn.findLeader()}
	}

	entry := LogEntry{
		Term:      rn.currentTerm,
		Index:     len(rn.log),
		Command:   "SET",
		Key:       key,
		Value:     value,
		Committed: false,
	}

	rn.log = append(rn.log, entry)
	log.Printf("Node %d (Leader) appended log entry %d: SET %s=%s", rn.id, entry.Index, key, value)

	return nil
}

func (rn *RaftNode) Get(key string) (string, bool) {
	return rn.stateMachine.Get(key)
}

func (rn *RaftNode) findLeader() int {
	for peerID := range rn.peers {
		if peerID != rn.id {
			return peerID
		}
	}
	return -1
}

type NotLeaderError struct {
	LeaderID int
}

func (e *NotLeaderError) Error() string {
	return "not leader"
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
