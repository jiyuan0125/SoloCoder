package api

type CreatePoolRequest struct {
	TargetAddress    string `json:"target_address"`
	MinIdle          int    `json:"min_idle"`
	MaxActive        int    `json:"max_active"`
	IdleTimeoutSec   int    `json:"idle_timeout_sec"`
	MaxLifetimeSec   int    `json:"max_lifetime_sec"`
	AcquireTimeoutMs int    `json:"acquire_timeout_ms"`
	HealthCheckMs    int    `json:"health_check_ms"`
}

type CreatePoolResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type AcquireConnectionResponse struct {
	Success bool   `json:"success"`
	ConnID  string `json:"conn_id,omitempty"`
	Message string `json:"message,omitempty"`
}

type ReleaseConnectionRequest struct {
	ConnID string `json:"conn_id"`
}

type ReleaseConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type PoolStatsResponse struct {
	Success    bool   `json:"success"`
	MinIdle    int    `json:"min_idle,omitempty"`
	MaxActive  int    `json:"max_active,omitempty"`
	TotalConns int    `json:"total_conns,omitempty"`
	IdleConns  int    `json:"idle_conns,omitempty"`
	ActiveConns int   `json:"active_conns,omitempty"`
	Message    string `json:"message,omitempty"`
}
