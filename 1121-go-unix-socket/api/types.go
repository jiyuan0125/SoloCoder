package api

type HealthResponse struct {
	Status    string `json:"status"`
	PID       int    `json:"pid"`
	Timestamp int64  `json:"timestamp"`
}

type StatusResponse struct {
	PID          int    `json:"pid"`
	ListenAddr   string `json:"listen_addr"`
	IsPrimary    bool   `json:"is_primary"`
	RestartCount int    `json:"restart_count"`
	StartTime    int64  `json:"start_time"`
}

type RestartHistory struct {
	OldPID     int   `json:"old_pid"`
	NewPID     int   `json:"new_pid"`
	Timestamp  int64 `json:"timestamp"`
	Successful bool  `json:"successful"`
}

type RestartHistoryResponse struct {
	History []RestartHistory `json:"history"`
}

type EchoResponse struct {
	PID       int    `json:"pid"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
