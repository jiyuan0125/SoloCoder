package protocol

type CommandType string

const (
	CmdSync   CommandType = "sync"
	CmdStatus CommandType = "status"
	CmdHistory CommandType = "history"
)

type Request struct {
	Command     CommandType `json:"command"`
	SourceFile  string      `json:"source_file"`
	TargetFile  string      `json:"target_file"`
	KeyColumn   string      `json:"key_column"`
	Apply       bool        `json:"apply"`
}

type Response struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	Changes     *ChangeSummary `json:"changes,omitempty"`
	History     []SyncHistory `json:"history,omitempty"`
}

type ChangeSummary struct {
	Added    int `json:"added"`
	Modified int `json:"modified"`
	Deleted  int `json:"deleted"`
	Total    int `json:"total"`
	ChangeFile string `json:"change_file"`
}

type SyncHistory struct {
	ID         string `json:"id"`
	Timestamp  string `json:"timestamp"`
	SourceFile string `json:"source_file"`
	TargetFile string `json:"target_file"`
	Changes    ChangeSummary `json:"changes"`
	Applied    bool   `json:"applied"`
}

type ChangeRecord struct {
	Operation string            `json:"operation"`
	OldValues map[string]string `json:"old_values,omitempty"`
	NewValues map[string]string `json:"new_values,omitempty"`
}
