package raft

import (
	"sync"
	"time"
)

type NodeState string

const (
	Follower  NodeState = "follower"
	Candidate NodeState = "candidate"
	Leader    NodeState = "leader"
)

const (
	ElectionTimeoutMin = 150 * time.Millisecond
	ElectionTimeoutMax = 300 * time.Millisecond
	HeartbeatInterval  = 50 * time.Millisecond
)

type LogEntry struct {
	Term    int
	Command string
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
	Term    int
	Success bool
	XTerm   int
	XIndex  int
	XLen    int
}

type RaftNode struct {
	mu sync.Mutex

	id    int
	state NodeState

	currentTerm int
	votedFor    int
	log         []LogEntry

	commitIndex int
	lastApplied int

	nextIndex  []int
	matchIndex []int

	electionTimer    *time.Timer
	heartbeatTimer   *time.Ticker
	heartbeatTicker  *time.Ticker
	electionResetCh  chan struct{}
	heartbeatResetCh chan struct{}

	peers     []*RaftNode
	peerChans []chan interface{}

	stopCh chan struct{}

	lastHeartbeatAt int64
	leaderID        int
}

type Cluster struct {
	mu    sync.Mutex
	nodes map[int]*RaftNode
}

type Message struct {
	From int
	Type string
	Args interface{}
	Reply chan interface{}
}
