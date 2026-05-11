package antcolony

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
	AntCount           int     `json:"antCount"`
	InitialPheromone   float64 `json:"initialPheromone"`
	EvaporationRate    float64 `json:"evaporationRate"`
	Alpha              float64 `json:"alpha"`
	Beta               float64 `json:"beta"`
	MaxIterations      int     `json:"maxIterations"`
	ConvergenceThreshold float64 `json:"convergenceThreshold"`
	EliteWeight        float64 `json:"eliteWeight"`
}

func DefaultParameters() Parameters {
	return Parameters{
		AntCount:           20,
		InitialPheromone:   1.0,
		EvaporationRate:    0.1,
		Alpha:              1.0,
		Beta:               2.0,
		MaxIterations:      100,
		ConvergenceThreshold: 1e-6,
		EliteWeight:        2.0,
	}
}

type Path struct {
	Nodes  []string `json:"nodes"`
	Length float64  `json:"length"`
}

type SolveResult struct {
	Success          bool    `json:"success"`
	Error            string  `json:"error,omitempty"`
	Path             Path    `json:"path,omitempty"`
	Iterations       int     `json:"iterations"`
	Converged        bool    `json:"converged"`
	BestLengthHistory []float64 `json:"bestLengthHistory,omitempty"`
}

type SolveStatus struct {
	Status           string  `json:"status"`
	Progress         float64 `json:"progress"`
	CurrentIteration int     `json:"currentIteration"`
	BestLength       float64 `json:"bestLength"`
	Result           *SolveResult `json:"result,omitempty"`
}

const (
	StatusIdle       = "idle"
	StatusRunning    = "running"
	StatusCompleted  = "completed"
	StatusError      = "error"
)
