package common

type PoolObject struct {
	Data string
}

type GetRequest struct {
	PoolID string `json:"pool_id"`
}

type GetResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Object     *PoolObject `json:"object,omitempty"`
}

type PutRequest struct {
	PoolID string      `json:"pool_id"`
	Object *PoolObject `json:"object"`
}

type PutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsRequest struct {
	PoolID string `json:"pool_id"`
}

type StatsResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Free      int    `json:"free"`
	Created   int64  `json:"created"`
	Discarded int64  `json:"discarded"`
	Hits      int64  `json:"hits"`
	Misses    int64  `json:"misses"`
	HitRate   float64 `json:"hit_rate"`
}

type CreatePoolRequest struct {
	PoolID      string `json:"pool_id"`
	MaxSize     int    `json:"max_size,omitempty"`
	IdleTimeout string `json:"idle_timeout,omitempty"`
}

type CreatePoolResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ClosePoolRequest struct {
	PoolID string `json:"pool_id"`
}

type ClosePoolResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
