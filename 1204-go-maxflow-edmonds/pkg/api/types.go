package api

type Node struct {
	ID string `json:"id"`
}

type Edge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Capacity int    `json:"capacity"`
}

type FlowRequest struct {
	Nodes  []Node `json:"nodes"`
	Edges  []Edge `json:"edges"`
	Source string `json:"source"`
	Sink   string `json:"sink"`
}

type FlowResult struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Flow    int    `json:"flow"`
	MaxFlow int    `json:"-"`
}

type FlowResponse struct {
	MaxFlow int          `json:"max_flow"`
	Flows   []FlowResult `json:"flows"`
}
