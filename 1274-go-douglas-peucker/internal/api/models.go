package api

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type UniformOptions struct {
	MinDistance float64 `json:"min_distance,omitempty"`
	MaxDistance float64 `json:"max_distance,omitempty"`
}

type ThresholdMode string

const (
	FixedThresholdMode    ThresholdMode = "fixed"
	PercentageThresholdMode ThresholdMode = "percentage"
)

type SimplifyRequest struct {
	Points         []Point        `json:"points"`
	Mode           ThresholdMode  `json:"mode"`
	Threshold      float64        `json:"threshold"`
	IsClosed       bool           `json:"is_closed,omitempty"`
	UniformOptions *UniformOptions `json:"uniform_options,omitempty"`
}

type SimplifyResponse struct {
	Points           []Point  `json:"points"`
	OriginalCount    int      `json:"original_count"`
	SimplifiedCount  int      `json:"simplified_count"`
	ReductionRate    float64  `json:"reduction_rate"`
	AreaDeviation    float64  `json:"area_deviation"`
	OriginalArea     float64  `json:"original_area"`
	SimplifiedArea   float64  `json:"simplified_area"`
	TotalLength      float64  `json:"total_length"`
}
