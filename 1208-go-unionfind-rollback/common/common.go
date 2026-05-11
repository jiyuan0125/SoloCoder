package common

type Action struct {
	Type string `json:"type"`
	Node string `json:"node,omitempty"`
	A    string `json:"a,omitempty"`
	B    string `json:"b,omitempty"`
	Step int    `json:"step,omitempty"`
}

type ActionResult struct {
	Type     string      `json:"type"`
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Step     int         `json:"step"`
}

type Request struct {
	Actions []Action `json:"actions"`
}

type Response struct {
	Results []ActionResult `json:"results"`
	Success bool           `json:"success"`
}
