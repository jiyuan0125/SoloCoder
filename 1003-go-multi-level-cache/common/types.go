package common

import "time"

type GetRequest struct {
	Key string `json:"key"`
}

type GetResponse struct {
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Found  bool        `json:"found"`
	Expire time.Time   `json:"expire"`
}

type SetRequest struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	MemoryTTL time.Duration `json:"memory_ttl"`
	FileTTL  time.Duration `json:"file_ttl"`
}

type SetResponse struct {
	Success bool `json:"success"`
}

type DeleteRequest struct {
	Key    string `json:"key"`
	Prefix string `json:"prefix"`
}

type DeleteResponse struct {
	Deleted int  `json:"deleted"`
	Success bool `json:"success"`
}

type ClearRequest struct {
}

type ClearResponse struct {
	Success bool `json:"success"`
}

type StatsResponse struct {
	MemoryHits      int64   `json:"memory_hits"`
	MemoryMisses    int64   `json:"memory_misses"`
	FileHits        int64   `json:"file_hits"`
	FileMisses      int64   `json:"file_misses"`
	HitRate         float64 `json:"hit_rate"`
	MemoryCount     int     `json:"memory_count"`
	FileCount       int     `json:"file_count"`
	MemoryCapacity  int     `json:"memory_capacity"`
	MemoryUsage     float64 `json:"memory_usage"`
}
