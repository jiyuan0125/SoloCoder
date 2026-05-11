package api

type ProblemSpec struct {
	Function     string   `json:"function"`
	CustomExpr   string   `json:"custom_expr,omitempty"`
	Variables    []string `json:"variables,omitempty"`
	Dimensions   int      `json:"dimensions"`
	LowerBound   []float64 `json:"lower_bound,omitempty"`
	UpperBound   []float64 `json:"upper_bound,omitempty"`
	Minimize     bool     `json:"minimize"`
}

type AlgorithmConfig struct {
	PopulationSize       int     `json:"population_size"`
	MaxGenerations       int     `json:"max_generations"`
	EliteCount           int     `json:"elite_count"`
	CrossoverRate        float64 `json:"crossover_rate"`
	MutationRate         float64 `json:"mutation_rate"`
	SBXIndex             float64 `json:"sbx_index"`
	InitialMutationSigma float64 `json:"initial_mutation_sigma"`
	FinalMutationSigma   float64 `json:"final_mutation_sigma"`
}

type OptimizeRequest struct {
	ID       string          `json:"id,omitempty"`
	Problem  *ProblemSpec    `json:"problem"`
	Config   *AlgorithmConfig `json:"config,omitempty"`
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskResponse struct {
	TaskID      string     `json:"task_id"`
	Status      TaskStatus `json:"status"`
	CreatedAt   string     `json:"created_at"`
	StartedAt   string     `json:"started_at,omitempty"`
	CompletedAt string     `json:"completed_at,omitempty"`
}

type Statistics struct {
	Generation   int       `json:"generation"`
	BestFitness  float64   `json:"best_fitness"`
	MeanFitness  float64   `json:"mean_fitness"`
	StdDeviation float64  `json:"std_deviation"`
	BestGenes    []float64 `json:"best_genes"`
}

type TaskResult struct {
	TaskID       string       `json:"task_id"`
	Status       TaskStatus   `json:"status"`
	CurrentGen   int          `json:"current_generation"`
	TotalGens    int          `json:"total_generations"`
	BestGenes    []float64    `json:"best_genes"`
	BestFitness  float64      `json:"best_fitness"`
	Statistics   []Statistics `json:"statistics,omitempty"`
	Error        string       `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ListFunctionsResponse struct {
	Functions []FunctionInfo `json:"functions"`
}

type FunctionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Minimize    bool   `json:"minimize"`
	DefaultLower []float64 `json:"default_lower"`
	DefaultUpper []float64 `json:"default_upper"`
}

type TaskListResponse struct {
	Tasks []TaskInfo `json:"tasks"`
}

type TaskInfo struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	Function  string     `json:"function"`
	CurrentGen int        `json:"current_generation"`
	BestFitness float64   `json:"best_fitness"`
	CreatedAt string      `json:"created_at"`
}
