package common

import "time"

type CompactionStrategy string

const (
	SizeTieredStrategy CompactionStrategy = "size-tiered"
	LeveledStrategy    CompactionStrategy = "leveled"
)

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PutResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetRequest struct {
	Key string `json:"key"`
}

type GetResponse struct {
	Exists bool   `json:"exists"`
	Value  string `json:"value,omitempty"`
	Error  string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type RangeRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RangeResponse struct {
	Results []KVPair `json:"results"`
	Error   string   `json:"error,omitempty"`
}

type CompactRequest struct {
}

type CompactResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type SwitchStrategyRequest struct {
	Strategy CompactionStrategy `json:"strategy"`
}

type SwitchStrategyResponse struct {
	Success  bool               `json:"success"`
	Strategy CompactionStrategy `json:"strategy,omitempty"`
	Error    string             `json:"error,omitempty"`
}

type LevelStats struct {
	Level     int   `json:"level"`
	FileCount int   `json:"file_count"`
	TotalSize int64 `json:"total_size"`
}

type StatsResponse struct {
	Levels       []LevelStats `json:"levels"`
	Strategy     string       `json:"strategy"`
	WriteAmplification float64 `json:"write_amplification"`
	TotalWrites  int64        `json:"total_writes"`
	UserWrites   int64        `json:"user_writes"`
}

type BatchPutRequest struct {
	Pairs []KVPair `json:"pairs"`
}

type BatchPutResponse struct {
	SuccessCount int    `json:"success_count"`
	Error        string `json:"error,omitempty"`
}

type Config struct {
	MemTableSize   int               `json:"mem_table_size"`
	Strategy       CompactionStrategy `json:"strategy"`
	DataDir        string            `json:"data_dir"`
}

type ConfigResponse struct {
	Current Config `json:"current"`
	Error   string `json:"error,omitempty"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}
