package common

type InsertRequest struct {
	GeoJSON interface{} `json:"geojson"`
}

type InsertResponse struct {
	Success bool   `json:"success"`
	Count   int    `json:"count"`
	Error   string `json:"error,omitempty"`
}

type RangeQueryRequest struct {
	MinLon float64 `json:"min_lon"`
	MinLat float64 `json:"min_lat"`
	MaxLon float64 `json:"max_lon"`
	MaxLat float64 `json:"max_lat"`
}

type PointQueryRequest struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type NearestNeighborRequest struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
	K   int     `json:"k"`
}

type QueryResponse struct {
	Success   bool        `json:"success"`
	Count     int         `json:"count"`
	Features  interface{} `json:"features"`
	Error     string      `json:"error,omitempty"`
}

type StatsResponse struct {
	Success bool `json:"success"`
	Count   int  `json:"count"`
}

type ClearResponse struct {
	Success bool `json:"success"`
}
