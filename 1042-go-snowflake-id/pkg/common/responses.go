package common

type GenerateResponse struct {
	Success bool     `json:"success"`
	ID      int64    `json:"id,omitempty"`
	IDs     []int64  `json:"ids,omitempty"`
	Count   int      `json:"count,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type StatusResponse struct {
	Success         bool                   `json:"success"`
	LastTimestamp   int64                  `json:"last_timestamp"`
	CurrentTimestamp int64                 `json:"current_timestamp"`
	MachineID       int64                  `json:"machine_id"`
	LastSequence    int64                  `json:"last_sequence"`
	Epoch           int64                  `json:"epoch"`
	MaxSequence     int64                  `json:"max_sequence"`
}

type ParseResponse struct {
	Success        bool   `json:"success"`
	ID             int64  `json:"id"`
	Timestamp      int64  `json:"timestamp"`
	RealTimestamp  int64  `json:"real_timestamp"`
	MachineID      int64  `json:"machine_id"`
	Sequence       int64  `json:"sequence"`
	Error          string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
