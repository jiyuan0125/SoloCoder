package model

type ActionType string

const (
	ActionRangeAdd ActionType = "range_add"
	ActionRangeSum ActionType = "range_sum"
	ActionRangeMax ActionType = "range_max"
)

type Action struct {
	Type  ActionType `json:"type"`
	Left  int        `json:"left"`
	Right int        `json:"right"`
	Value int        `json:"value,omitempty"`
}

type Request struct {
	InitValues []int    `json:"init_values"`
	Actions    []Action `json:"actions"`
}

type Response struct {
	Results []int `json:"results"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
