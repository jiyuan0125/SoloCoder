package common

type Node struct {
	ID string `json:"id"`
}

type Edge struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Weight float64 `json:"weight"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Parameters struct {
	AntCount             int     `json:"antCount,omitempty"`
	InitialPheromone     float64 `json:"initialPheromone,omitempty"`
	EvaporationRate      float64 `json:"evaporationRate,omitempty"`
	Alpha                float64 `json:"alpha,omitempty"`
	Beta                 float64 `json:"beta,omitempty"`
	MaxIterations        int     `json:"maxIterations,omitempty"`
	ConvergenceThreshold float64 `json:"convergenceThreshold,omitempty"`
	EliteWeight          float64 `json:"eliteWeight,omitempty"`
}

type SubmitRequest struct {
	Graph      Graph      `json:"graph"`
	Start      string     `json:"start"`
	End        string     `json:"end"`
	Parameters Parameters `json:"parameters,omitempty"`
}

type SubmitResponse struct {
	TaskID    string `json:"taskId"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	IsLarge   bool   `json:"isLarge,omitempty"`
}

type StatusResponse struct {
	TaskID           string  `json:"taskId"`
	Status           string  `json:"status"`
	Progress         float64 `json:"progress"`
	CurrentIteration int     `json:"currentIteration"`
	BestLength       float64 `json:"bestLength"`
	Completed        bool    `json:"completed"`
}

type Path struct {
	Nodes  []string `json:"nodes"`
	Length float64  `json:"length"`
}

type ResultResponse struct {
	TaskID            string    `json:"taskId"`
	Success           bool      `json:"success"`
	Error             string    `json:"error,omitempty"`
	Path              Path      `json:"path,omitempty"`
	Iterations        int       `json:"iterations"`
	Converged         bool      `json:"converged"`
	BestLengthHistory []float64 `json:"bestLengthHistory,omitempty"`
}
