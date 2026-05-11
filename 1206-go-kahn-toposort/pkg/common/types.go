package common

type Dependency struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type SortRequest struct {
	Tasks        []string     `json:"tasks"`
	Dependencies []Dependency `json:"dependencies"`
}

type SortResponse struct {
	Levels   [][]string `json:"levels"`
	HasCycle bool       `json:"has_cycle"`
	Error    string     `json:"error,omitempty"`
}
