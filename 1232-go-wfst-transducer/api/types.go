package api

type Arc struct {
	From   int     `json:"from"`
	Input  string  `json:"input"`
	Output string  `json:"output"`
	Weight float64 `json:"weight"`
	Next   int     `json:"next"`
}

type CreateRequest struct {
	States []int `json:"states,omitempty"`
	Arcs   []Arc `json:"arcs"`
}

type CreateResponse struct {
	Success   bool   `json:"success"`
	WFSTID    string `json:"wfst_id"`
	NumStates int    `json:"num_states"`
	NumArcs   int    `json:"num_arcs"`
	Message   string `json:"message,omitempty"`
}

type SearchRequest struct {
	WFSTID  string   `json:"wfst_id"`
	Input   []string `json:"input"`
}

type SearchResponse struct {
	Success       bool      `json:"success"`
	Path          []int     `json:"path"`
	InputSeq      []string  `json:"input_seq"`
	OutputSeq     []string  `json:"output_seq"`
	Weights       []float64 `json:"weights"`
	TotalWeight   float64   `json:"total_weight"`
	Message       string    `json:"message,omitempty"`
}

type ComposeRequest struct {
	WFSTIDA string `json:"wfst_id_a"`
	WFSTIDB string `json:"wfst_id_b"`
}

type ComposeResponse struct {
	Success   bool   `json:"success"`
	WFSTID    string `json:"wfst_id"`
	NumStates int    `json:"num_states"`
	NumArcs   int    `json:"num_arcs"`
	Message   string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
