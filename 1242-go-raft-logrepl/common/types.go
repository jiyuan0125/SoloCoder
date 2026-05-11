package common

type LogEntry struct {
	Index   int64  `json:"index"`
	Term    int64  `json:"term"`
	Command string `json:"command"`
}

type NodeInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

type SubmitRequest struct {
	NodeID  string `json:"node_id"`
	Command string `json:"command"`
}

type SubmitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type LogResponse struct {
	Success bool        `json:"success"`
	NodeID  string      `json:"node_id"`
	Entries []LogEntry  `json:"entries,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ConsistencyCheckRequest struct {
	NodeID1 string `json:"node_id1"`
	NodeID2 string `json:"node_id2"`
}

type ConsistencyCheckResponse struct {
	Success bool `json:"success"`
	
	Consistent     bool       `json:"consistent"`
	MatchedCount   int        `json:"matched_count,omitempty"`
	Node1ID        string     `json:"node1_id"`
	Node1Total     int        `json:"node1_total"`
	Node2ID        string     `json:"node2_id"`
	Node2Total     int        `json:"node2_total"`
	
	MismatchIndex  int64      `json:"mismatch_index,omitempty"`
	Node1Entry     *LogEntry  `json:"node1_entry,omitempty"`
	Node2Entry     *LogEntry  `json:"node2_entry,omitempty"`
	Message        string     `json:"message,omitempty"`
}

type ResetRequest struct {
	NodeID string `json:"node_id"`
}

type ResetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type NodeStatus struct {
	ID         string     `json:"id"`
	Address    string     `json:"address"`
	EntryCount int        `json:"entry_count"`
	LastIndex  int64      `json:"last_index"`
	LastTerm   int64      `json:"last_term"`
	Entries    []LogEntry `json:"entries"`
}

type StatusResponse struct {
	Success bool         `json:"success"`
	Nodes   []NodeStatus `json:"nodes,omitempty"`
	Message string       `json:"message,omitempty"`
}
