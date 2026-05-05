package protocol

const (
	CommandTypeStart    = "start"
	CommandTypeStop     = "stop"
	CommandTypeStatus   = "status"
	CommandTypeGetStats = "get_stats"
)

type LogFormat struct {
	Delimiter string
	Columns   []LogColumn
}

type LogColumn string

const (
	ColumnSQL        LogColumn = "sql"
	ColumnExecTime   LogColumn = "exec_time"
	ColumnScanRows   LogColumn = "scan_rows"
	ColumnLockWait   LogColumn = "lock_wait"
)

type Request struct {
	Command       string     `json:"command"`
	LogFilePath   string     `json:"log_file_path,omitempty"`
	LogFormat     *LogFormat `json:"log_format,omitempty"`
	MinExecTimeMs float64    `json:"min_exec_time_ms,omitempty"`
	TopN          int        `json:"top_n,omitempty"`
	SortBy        SortType   `json:"sort_by,omitempty"`
}

type SortType string

const (
	SortByAvgExecTime SortType = "avg_exec_time"
	SortByTotalExecTime SortType = "total_exec_time"
)

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Status  string      `json:"status,omitempty"`
	Stats   *Statistics `json:"stats,omitempty"`
}

type Statistics struct {
	TotalRowsParsed   int64        `json:"total_rows_parsed"`
	RowsSkipped       int64        `json:"rows_skipped"`
	UniqueTemplates   int64        `json:"unique_templates"`
	TemplateStats     []TemplateStat `json:"template_stats"`
}

type TemplateStat struct {
	Rank          int     `json:"rank"`
	SQLTemplate   string  `json:"sql_template"`
	Count         int64   `json:"count"`
	AvgExecTimeMs float64 `json:"avg_exec_time_ms"`
	MaxExecTimeMs float64 `json:"max_exec_time_ms"`
	TotalExecTimeMs float64 `json:"total_exec_time_ms"`
	AvgScanRows   float64 `json:"avg_scan_rows"`
}

type LogEntry struct {
	SQL          string
	ExecTimeMs   float64
	ScanRows     int64
	LockWaitMs   float64
}
