package common

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type DistanceRequest struct {
	PointA        Coordinate `json:"pointA"`
	PointB        Coordinate `json:"pointB"`
	UseEllipsoid  bool       `json:"useEllipsoid"`
}

type DistanceResponse struct {
	DistanceKm float64 `json:"distanceKm"`
}

type PolylineRequest struct {
	Points       []Coordinate `json:"points"`
	UseEllipsoid bool         `json:"useEllipsoid"`
}

type PolylineLengthResponse struct {
	TotalLengthKm float64 `json:"totalLengthKm"`
	Simplified    bool    `json:"simplified"`
	OriginalCount int     `json:"originalCount"`
	SimplifiedCount int   `json:"simplifiedCount"`
}

type PointToPolylineRequest struct {
	Point        Coordinate   `json:"point"`
	Polyline     []Coordinate `json:"polyline"`
	UseEllipsoid bool         `json:"useEllipsoid"`
}

type PointToPolylineResponse struct {
	ClosestDistanceKm float64    `json:"closestDistanceKm"`
	ClosestPoint      Coordinate `json:"closestPoint"`
	SegmentIndex      int        `json:"segmentIndex"`
}

type BatchDistanceRequest struct {
	Center       Coordinate   `json:"center"`
	Targets      []Coordinate `json:"targets"`
	UseEllipsoid bool         `json:"useEllipsoid"`
}

type TargetDistance struct {
	Target     Coordinate `json:"target"`
	DistanceKm float64    `json:"distanceKm"`
}

type BatchDistanceResponse struct {
	Results []TargetDistance `json:"results"`
}

type CircleRequest struct {
	Center       Coordinate `json:"center"`
	RadiusKm     float64    `json:"radiusKm"`
	NumPoints    int        `json:"numPoints"`
	UseEllipsoid bool       `json:"useEllipsoid"`
}

type CircleResponse struct {
	Boundary []Coordinate `json:"boundary"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
