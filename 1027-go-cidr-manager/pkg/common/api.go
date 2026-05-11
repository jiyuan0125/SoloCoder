package common

type ParseRequest struct {
	CIDR string `json:"cidr"`
}

type ParseResponse struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	CIDR       string `json:"cidr,omitempty"`
	Network    string `json:"network,omitempty"`
	Broadcast  string `json:"broadcast,omitempty"`
	FirstUsable string `json:"first_usable,omitempty"`
	LastUsable  string `json:"last_usable,omitempty"`
	Prefix     int    `json:"prefix,omitempty"`
	Size       uint64 `json:"size,omitempty"`
	Version    string `json:"version,omitempty"`
}

type ContainsRequest struct {
	CIDR string `json:"cidr"`
	IP   string `json:"ip"`
}

type ContainsResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Contained bool   `json:"contained"`
}

type MergeRequest struct {
	CIDRs []string `json:"cidrs"`
}

type MergeResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Merged  []string `json:"merged,omitempty"`
}

type AllocateRequest struct {
	PoolID   string `json:"pool_id"`
	Prefix   int    `json:"prefix"`
}

type AllocateResponse struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Allocated  string `json:"allocated,omitempty"`
	PoolID     string `json:"pool_id,omitempty"`
}

type ReleaseRequest struct {
	PoolID string `json:"pool_id"`
	CIDR   string `json:"cidr"`
}

type ReleaseResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	PoolID  string `json:"pool_id,omitempty"`
}

type ExcludeRequest struct {
	PoolID string `json:"pool_id"`
	CIDR   string `json:"cidr"`
}

type ExcludeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	PoolID  string `json:"pool_id,omitempty"`
}

type CreatePoolRequest struct {
	PoolID string `json:"pool_id"`
	CIDR   string `json:"cidr"`
}

type CreatePoolResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	PoolID  string `json:"pool_id,omitempty"`
}

type PoolStatusRequest struct {
	PoolID string `json:"pool_id"`
}

type PoolStatusResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	PoolID    string   `json:"pool_id,omitempty"`
	Available []string `json:"available,omitempty"`
	Used      []string `json:"used,omitempty"`
}
