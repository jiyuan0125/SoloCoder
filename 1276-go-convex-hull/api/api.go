package api

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WeightedPoint struct {
	Point
	Weight float64 `json:"weight"`
}

type HullType string

const (
	HullTypeEmpty   HullType = "empty"
	HullTypePoint   HullType = "point"
	HullTypeSegment HullType = "segment"
	HullTypePolygon HullType = "polygon"
)

type ConvexHullResponse struct {
	Type   HullType `json:"type"`
	Points []Point  `json:"points"`
}

type ComputeRequest struct {
	Points []Point `json:"points"`
}

type WeightedComputeRequest struct {
	Points []WeightedPoint `json:"points"`
}

type ComputeResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	Hull    ConvexHullResponse `json:"hull,omitempty"`
}

type PerimeterAreaRequest struct {
	Points []Point `json:"points"`
}

type PerimeterAreaResponse struct {
	Success   bool    `json:"success"`
	Error     string  `json:"error,omitempty"`
	Perimeter float64 `json:"perimeter"`
	Area      float64 `json:"area"`
	Hull      ConvexHullResponse `json:"hull,omitempty"`
}

type PointInHullRequest struct {
	Point  Point   `json:"point"`
	HullPoints []Point `json:"hull_points"`
}

type PointInHullResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Inside  bool   `json:"inside"`
}

type IntersectionRequest struct {
	Hull1 []Point `json:"hull1"`
	Hull2 []Point `json:"hull2"`
}

type IntersectionResponse struct {
	Success         bool              `json:"success"`
	Error           string            `json:"error,omitempty"`
	Intersection    ConvexHullResponse `json:"intersection,omitempty"`
	IntersectionArea float64           `json:"intersection_area"`
}

type FilterBoundaryRequest struct {
	Points []Point `json:"points"`
}

type FilterBoundaryResponse struct {
	Success  bool    `json:"success"`
	Error    string  `json:"error,omitempty"`
	Boundary []Point `json:"boundary"`
}

type DynamicAddRequest struct {
	SessionID string  `json:"session_id"`
	Points    []Point `json:"points"`
}

type DynamicAddResponse struct {
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
	SessionID string            `json:"session_id"`
	Hull      ConvexHullResponse `json:"hull"`
}

type DynamicGetRequest struct {
	SessionID string `json:"session_id"`
}

type DynamicGetResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	Hull    ConvexHullResponse `json:"hull"`
}
