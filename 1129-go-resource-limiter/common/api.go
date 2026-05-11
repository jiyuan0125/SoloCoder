package common

type LimitMode string

const (
	ModeWait LimitMode = "wait"
	ModeFail LimitMode = "fail"
)

type RequestRequest struct {
	Key string `json:"key"`
}

type RequestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ConfigResponse struct {
	GlobalLimit int       `json:"global_limit"`
	KeyLimit    int       `json:"key_limit"`
	Mode        LimitMode `json:"mode"`
}

type ConfigUpdateRequest struct {
	GlobalLimit *int       `json:"global_limit,omitempty"`
	KeyLimit    *int       `json:"key_limit,omitempty"`
	Key         string     `json:"key,omitempty"`
	Mode        *LimitMode `json:"mode,omitempty"`
}

type KeyStats struct {
	Key       string `json:"key"`
	InUse     int    `json:"in_use"`
	Capacity  int    `json:"capacity"`
	Available int    `json:"available"`
	Rejected  int64  `json:"rejected"`
}

type StatsResponse struct {
	GlobalInUse     int                `json:"global_in_use"`
	GlobalCapacity  int                `json:"global_capacity"`
	GlobalAvailable int                `json:"global_available"`
	GlobalQueued    int64              `json:"global_queued"`
	GlobalRejected  int64              `json:"global_rejected"`
	Mode            LimitMode          `json:"mode"`
	KeyStats        map[string]KeyStats `json:"key_stats"`
}

type ModeUpdateRequest struct {
	Mode LimitMode `json:"mode"`
}
