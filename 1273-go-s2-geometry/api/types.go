package api

type LatLngToCellIDRequest struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Level int     `json:"level"`
}

type LatLngToCellIDResponse struct {
	CellID string `json:"cell_id"`
	Face   int    `json:"face"`
	Level  int    `json:"level"`
}

type CellIDToLatLngRequest struct {
	CellID string `json:"cell_id"`
}

type CellIDToLatLngResponse struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	LatLo float64 `json:"lat_lo"`
	LatHi float64 `json:"lat_hi"`
	LngLo float64 `json:"lng_lo"`
	LngHi float64 `json:"lng_hi"`
	Face  int     `json:"face"`
	Level int     `json:"level"`
}

type ContainsRequest struct {
	Container string `json:"container"`
	Contained string `json:"contained"`
}

type ContainsResponse struct {
	Contains bool `json:"contains"`
}

type NeighborsRequest struct {
	CellID string `json:"cell_id"`
	Level  int    `json:"level"`
}

type NeighborsResponse struct {
	Neighbors []string `json:"neighbors"`
}

type CoveringRequest struct {
	LatLo    float64 `json:"lat_lo"`
	LatHi    float64 `json:"lat_hi"`
	LngLo    float64 `json:"lng_lo"`
	LngHi    float64 `json:"lng_hi"`
	MinLevel int     `json:"min_level"`
	MaxLevel int     `json:"max_level"`
	MaxCells int     `json:"max_cells"`
}

type CoveringResponse struct {
	CellIDs []string `json:"cell_ids"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
