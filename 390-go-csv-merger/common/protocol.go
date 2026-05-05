package common

type CommandType string

const (
	CmdMergeCSV CommandType = "MERGE_CSV"
	CmdPing     CommandType = "PING"
)

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type MergeRequest struct {
	Command       CommandType `json:"command"`
	InputFiles    []string    `json:"input_files"`
	OutputFile    string      `json:"output_file"`
	DedupColumns  []string    `json:"dedup_columns,omitempty"`
	SortColumn    string      `json:"sort_column,omitempty"`
	SortOrder     SortOrder   `json:"sort_order,omitempty"`
	CustomHeader  []string    `json:"custom_header,omitempty"`
}

type MergeResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
	RecordsMerged int    `json:"records_merged,omitempty"`
	Error         string `json:"error,omitempty"`
}

type PingRequest struct {
	Command CommandType `json:"command"`
}

type PingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
