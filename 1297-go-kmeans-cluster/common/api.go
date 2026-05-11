package common

type Point []float64

type ClusterRequest struct {
	Points        []Point `json:"points"`
	K             *int    `json:"k,omitempty"`
	AutoK         bool    `json:"auto_k,omitempty"`
	MaxK          *int    `json:"max_k,omitempty"`
	MaxIterations *int    `json:"max_iterations,omitempty"`
	Epsilon       *float64 `json:"epsilon,omitempty"`
}

type ClusterResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message,omitempty"`
	Labels    []int          `json:"labels,omitempty"`
	Centers   []Point        `json:"centers,omitempty"`
	K         int            `json:"k,omitempty"`
	Iterations int           `json:"iterations,omitempty"`
	Converged bool           `json:"converged,omitempty"`
	Clusters  []ClusterInfo  `json:"clusters,omitempty"`
}

type WCSSRequest struct {
	Points        []Point `json:"points"`
	MaxK          int     `json:"max_k"`
	MaxIterations *int    `json:"max_iterations,omitempty"`
	Epsilon       *float64 `json:"epsilon,omitempty"`
}

type WCSSResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	WCSS    map[int]float64 `json:"wcss,omitempty"`
}

type ClusterInfo struct {
	Index       int     `json:"index"`
	Center      Point   `json:"center"`
	PointCount  int     `json:"point_count"`
	AvgDistance float64 `json:"avg_distance"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
