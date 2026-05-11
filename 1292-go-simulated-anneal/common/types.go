package common

type CoordinateType string

const (
	CoordinateTypeGeographic CoordinateType = "geographic"
	CoordinateTypeCartesian  CoordinateType = "cartesian"
)

type City struct {
	Name           string         `json:"name"`
	Latitude       float64        `json:"latitude,omitempty"`
	Longitude      float64        `json:"longitude,omitempty"`
	X              float64        `json:"x,omitempty"`
	Y              float64        `json:"y,omitempty"`
	CoordinateType CoordinateType `json:"coordinateType"`
}

type SolverConfig struct {
	InitialTemperature       float64 `json:"initialTemperature,omitempty"`
	FinalTemperature         float64 `json:"finalTemperature,omitempty"`
	CoolingRate              float64 `json:"coolingRate,omitempty"`
	IterationsPerTemperature int     `json:"iterationsPerTemperature,omitempty"`
	AdaptiveCooling          bool    `json:"adaptiveCooling,omitempty"`
	AdaptiveThreshold        int     `json:"adaptiveThreshold,omitempty"`
	RandomSeed               int64   `json:"randomSeed,omitempty"`
}

type SolveRequest struct {
	Cities []City       `json:"cities"`
	Config *SolverConfig `json:"config,omitempty"`
}

type SolveResponse struct {
	TaskID string `json:"taskId"`
}

type ConvergencePoint struct {
	Temperature    float64 `json:"temperature"`
	ObjectiveValue float64 `json:"objectiveValue"`
}

type ProgressResponse struct {
	TaskID               string             `json:"taskId"`
	Status               string             `json:"status"`
	CurrentTemperature   float64            `json:"currentTemperature,omitempty"`
	CurrentIteration     int                `json:"currentIteration,omitempty"`
	CurrentDistance      float64            `json:"currentDistance,omitempty"`
	BestDistance         float64            `json:"bestDistance,omitempty"`
	InitialDistance      float64            `json:"initialDistance,omitempty"`
	ImprovementPct       float64            `json:"improvementPct,omitempty"`
	Path                 []int              `json:"path,omitempty"`
	CityNames            []string           `json:"cityNames,omitempty"`
	TotalDistance        float64            `json:"totalDistance,omitempty"`
	Convergence          []ConvergencePoint `json:"convergence,omitempty"`
	Duration             int64              `json:"duration,omitempty"`
	DurationStr          string             `json:"durationStr,omitempty"`
	TotalIterations      int                `json:"totalIterations,omitempty"`
	RandomSeed           int64              `json:"randomSeed,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
