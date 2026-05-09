package api

type AnalysisRequest struct {
	Filename string `json:"filename"`
	Source   string `json:"source"`
}

type FunctionComplexity struct {
	Name       string `json:"name"`
	Line       int    `json:"line"`
	Complexity int    `json:"complexity"`
}

type CallGraph struct {
	Adjacency map[string][]string `json:"adjacency"`
}

type UnusedVariable struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Name     string `json:"name"`
}

type UnusedImport struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Path     string `json:"path"`
}

type SyntaxError struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

type AnalysisResponse struct {
	Complexities    []FunctionComplexity `json:"complexities"`
	CallGraph       CallGraph            `json:"call_graph"`
	UnusedVariables []UnusedVariable     `json:"unused_variables"`
	UnusedImports   []UnusedImport       `json:"unused_imports"`
	SyntaxErrors    []SyntaxError        `json:"syntax_errors,omitempty"`
}
