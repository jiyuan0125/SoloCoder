package api

type AnalyzeRequest struct {
	StackTrace string   `json:"stack_trace"`
	Options    *Options `json:"options,omitempty"`
}

type Options struct {
	WaitThresholdMinutes int `json:"wait_threshold_minutes,omitempty"`
	SuspectThreshold     int `json:"suspect_threshold,omitempty"`
}

type AnalyzeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Report  Report `json:"report,omitempty"`
}

type Report struct {
	TotalGoroutines int              `json:"total_goroutines"`
	ByBlockType     map[string]int   `json:"by_block_type"`
	SuspectGroups   []*SuspectGroup  `json:"suspect_groups"`
	HasSuspicious   bool             `json:"has_suspicious"`
}

type SuspectGroup struct {
	BlockType    string           `json:"block_type"`
	Count        int              `json:"count"`
	Goroutines   []*GoroutineInfo `json:"goroutines"`
}

type GoroutineInfo struct {
	ID             int           `json:"id"`
	State          string        `json:"state"`
	WaitMinutes    int           `json:"wait_minutes,omitempty"`
	HasWaitTime    bool          `json:"has_wait_time"`
	StackTruncated bool          `json:"stack_truncated"`
	UserStack      []StackFrame  `json:"user_stack"`
	RawStack       []string      `json:"raw_stack,omitempty"`
	Reasons        []string      `json:"reasons"`
	BlockingOn     *BlockingInfo `json:"blocking_on,omitempty"`
}

type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type BlockingInfo struct {
	Type        string `json:"type"`
	Function    string `json:"function,omitempty"`
	HasTimeout  bool   `json:"has_timeout,omitempty"`
}

const (
	BlockTypeChanReceive   = "chan_receive"
	BlockTypeChanSend      = "chan_send"
	BlockTypeSelect        = "select"
	BlockTypeIOWait        = "io_wait"
	BlockTypeSyscall       = "syscall"
	BlockTypeSleep         = "sleep"
	BlockTypeSemacquire    = "semacquire"
	BlockTypeUnknown       = "unknown"
)
