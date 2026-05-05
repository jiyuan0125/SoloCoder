package protocol

import "time"

const (
	DefaultPort = 3846
)

type RequestType string

const (
	RequestTypeValidate   RequestType = "validate"
	RequestTypeNextTime   RequestType = "next_time"
	RequestTypeNextNTimes RequestType = "next_n_times"
)

type ValidateRequest struct {
	Type    RequestType `json:"type"`
	Expr    string      `json:"expr"`
	Verbose bool        `json:"verbose,omitempty"`
}

type NextTimeRequest struct {
	Type    RequestType `json:"type"`
	Expr    string      `json:"expr"`
	Verbose bool        `json:"verbose,omitempty"`
}

type NextNTimesRequest struct {
	Type    RequestType `json:"type"`
	Expr    string      `json:"expr"`
	N       int         `json:"n"`
	Verbose bool        `json:"verbose,omitempty"`
}

type ValidateResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type ExecutionTime struct {
	Time   string        `json:"time"`
	Delay  time.Duration `json:"delay,omitempty"`
}

type NextTimeResponse struct {
	Valid          bool           `json:"valid"`
	Error          string         `json:"error,omitempty"`
	NextTime       *ExecutionTime `json:"next_time,omitempty"`
	WillNeverFire  bool           `json:"will_never_fire,omitempty"`
}

type NextNTimesResponse struct {
	Valid          bool            `json:"valid"`
	Error          string          `json:"error,omitempty"`
	NextTimes      []ExecutionTime `json:"next_times,omitempty"`
	WillNeverFire  bool            `json:"will_never_fire,omitempty"`
	Count          int             `json:"count"`
}
