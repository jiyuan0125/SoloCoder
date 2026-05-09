package api

type NodeState string

const (
	Follower  NodeState = "follower"
	Candidate NodeState = "candidate"
	Leader    NodeState = "leader"
)

type LogEntry struct {
	Term    int
	Command string
}

type NodeInfo struct {
	ID              int
	State           NodeState
	CurrentTerm     int
	VotedFor        int
	CommitIndex     int
	LastApplied     int
	LogEntries      []LogEntry
	LeaderID        int
	LastHeartbeatAt int64
}

type CreateClusterRequest struct {
	NodeCount int `json:"node_count"`
}

type CreateClusterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Nodes   []int  `json:"nodes"`
}

type AddNodeRequest struct {
	NodeID int `json:"node_id"`
}

type AddNodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	NodeID  int    `json:"node_id"`
}

type SubmitLogRequest struct {
	Command string `json:"command"`
}

type SubmitLogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Term    int    `json:"term"`
	Index   int    `json:"index"`
}

type ListNodesResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Nodes   []NodeInfo `json:"nodes"`
}

type GetNodeResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Node    NodeInfo `json:"node"`
}

type TriggerElectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
