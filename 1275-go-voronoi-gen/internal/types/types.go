package types

import "encoding/json"

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Edge struct {
	Start Point `json:"start"`
	End   Point `json:"end"`
}

type VoronoiCell struct {
	Seed   Point   `json:"seed"`
	Points []Point `json:"points"`
	Edges  []Edge  `json:"edges"`
	Area   float64 `json:"area"`
}

type Diagram struct {
	Seeds   []Point        `json:"seeds"`
	Cells   []VoronoiCell  `json:"cells"`
	Edges   []Edge         `json:"edges"`
	Bounds  Bounds         `json:"bounds"`
}

type Bounds struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

type GenerateRequest struct {
	Seeds  []Point `json:"seeds"`
	Bounds Bounds  `json:"bounds,omitempty"`
}

type GenerateResponse struct {
	Diagram Diagram `json:"diagram"`
}

type NearestNeighborRequest struct {
	Point Point `json:"point"`
}

type NearestNeighborResponse struct {
	Seed       Point   `json:"seed"`
	Distance   float64 `json:"distance"`
	CellIndex  int     `json:"cell_index"`
}

type ClipRequest struct {
	Region Bounds `json:"region"`
}

type ClipResponse struct {
	Edges []Edge `json:"edges"`
}

type AreaRequest struct {
}

type AreaResponse struct {
	Areas []float64 `json:"areas"`
}

func (r *GenerateRequest) MarshalJSON() ([]byte, error) {
	type alias GenerateRequest
	return json.Marshal((*alias)(r))
}

func (r *GenerateRequest) UnmarshalJSON(data []byte) error {
	type alias GenerateRequest
	return json.Unmarshal(data, (*alias)(r))
}
