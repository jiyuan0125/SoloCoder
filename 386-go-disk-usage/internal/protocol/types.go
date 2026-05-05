package protocol

type ScanRequest struct {
	Path       string `json:"path"`
	Top        int    `json:"top"`
	MinSize    int64  `json:"min_size"`
	MaxDepth   int    `json:"max_depth"`
	Output     string `json:"output"`
	All        bool   `json:"all"`
	Summary    bool   `json:"summary"`
	Compare    bool   `json:"compare"`
}

type DirInfo struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Files      int    `json:"files"`
	Dirs       int    `json:"dirs"`
	Depth      int    `json:"depth"`
}

type ScanResult struct {
	Success      bool     `json:"success"`
	TotalSize  int64    `json:"total_size"`
	TotalFiles int     `json:"total_files"`
	TotalDirs  int     `json:"total_dirs"`
	Directories []DirInfo `json:"directories"`
	Warnings   []string `json:"warnings"`
	ScanTime   int64  `json:"scan_time"`
}

type CompareResult struct {
	OldScan   *ScanResult `json:"old_scan"`
	NewScan   *ScanResult `json:"new_scan"`
	GrowthDirs []GrowthDir  `json:"growth_dirs"`
}

type GrowthDir struct {
	Path        string `json:"path"`
	OldSize     int64  `json:"old_size"`
	NewSize     int64  `json:"new_size"`
	Growth      int64  `json:"growth"`
	OldFiles    int    `json:"old_files"`
	NewFiles    int    `json:"new_files"`
}

type MessageType int

const (
	MsgTypeScanRequest  MessageType = 1
	MsgTypeScanResponse MessageType = 2
	MsgTypeCompareRequest MessageType = 3
	MsgTypeCompareResponse MessageType = 4
	MsgTypeError        MessageType = 5
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}
