package api

type EncodeRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Precision int     `json:"precision"`
}

type EncodeResponse struct {
	Geohash string `json:"geohash"`
}

type DecodeRequest struct {
	Geohash string `json:"geohash"`
}

type DecodeResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type BoundingBoxRequest struct {
	Geohash string `json:"geohash"`
}

type BoundingBoxResponse struct {
	MinLat float64 `json:"min_lat"`
	MaxLat float64 `json:"max_lat"`
	MinLng float64 `json:"min_lng"`
	MaxLng float64 `json:"max_lng"`
}

type NeighborsRequest struct {
	Geohash string `json:"geohash"`
}

type NeighborsResponse struct {
	North     string `json:"north"`
	Northeast string `json:"northeast"`
	East      string `json:"east"`
	Southeast string `json:"southeast"`
	South     string `json:"south"`
	Southwest string `json:"southwest"`
	West      string `json:"west"`
	Northwest string `json:"northwest"`
}

type ProximitySearchRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusKm  float64 `json:"radius_km"`
}

type ProximitySearchResponse struct {
	Geohashes []string `json:"geohashes"`
	Count     int      `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
