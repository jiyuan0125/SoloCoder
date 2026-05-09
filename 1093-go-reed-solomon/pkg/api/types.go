package api

type EncodeRequest struct {
	Data string `json:"data"`
	N    int    `json:"n"`
	M    int    `json:"m"`
}

type ShardInfo struct {
	ID       string `json:"id"`
	Index    int    `json:"index"`
	IsLost   bool   `json:"is_lost"`
	Checksum []byte `json:"checksum,omitempty"`
}

type EncodeResponse struct {
	BlockID     string      `json:"block_id"`
	N           int         `json:"n"`
	M           int         `json:"m"`
	Shards      []ShardInfo `json:"shards"`
	OriginalLen int         `json:"original_len"`
}

type DecodeRequest struct {
	BlockID  string   `json:"block_id"`
	ShardIDs []string `json:"shard_ids"`
}

type DecodeResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ShardStatus struct {
	ShardID  string `json:"shard_id"`
	Index    int    `json:"index"`
	IsAvailable bool `json:"is_available"`
}

type BlockStatus struct {
	BlockID     string        `json:"block_id"`
	N           int           `json:"n"`
	M           int           `json:"m"`
	TotalShards int           `json:"total_shards"`
	AvailableShards int       `json:"available_shards"`
	Shards      []ShardStatus `json:"shards"`
}

type StatusResponse struct {
	Blocks []BlockStatus `json:"blocks"`
}

type MarkLostRequest struct {
	BlockID string `json:"block_id"`
	ShardID string `json:"shard_id"`
}

type MarkLostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
