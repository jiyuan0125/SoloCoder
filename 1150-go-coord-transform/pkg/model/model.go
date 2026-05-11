package model

type ConvertRequest struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Points []PointRequest `json:"points,omitempty"`
	Lng  *float64 `json:"lng,omitempty"`
	Lat  *float64 `json:"lat,omitempty"`
}

type PointRequest struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type ConvertResponse struct {
	Success bool          `json:"success"`
	Points  []PointResult `json:"points,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type PointResult struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}
