package common

type PutRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetResponse struct {
	Found bool   `json:"found"`
	Value string `json:"value,omitempty"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}

type KeysResponse struct {
	Keys []string `json:"keys"`
}

type ShardStats struct {
	ShardID    int  `json:"shard_id"`
	Count      int  `json:"count"`
	IsUneven   bool `json:"is_uneven"`
}

type StatsResponse struct {
	TotalCount int          `json:"total_count"`
	ShardCount int          `json:"shard_count"`
	Shards     []ShardStats `json:"shards"`
}
