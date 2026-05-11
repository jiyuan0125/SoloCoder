package api

type PointToIndexRequest struct {
	Order int    `json:"order"`
	X     uint64 `json:"x"`
	Y     uint64 `json:"y"`
}

type PointToIndexResponse struct {
	Index uint64 `json:"index"`
}

type IndexToPointRequest struct {
	Order int    `json:"order"`
	Index uint64 `json:"index"`
}

type IndexToPointResponse struct {
	X uint64 `json:"x"`
	Y uint64 `json:"y"`
}

type RangeQueryRequest struct {
	Order int    `json:"order"`
	MinX  uint64 `json:"min_x"`
	MaxX  uint64 `json:"max_x"`
	MinY  uint64 `json:"min_y"`
	MaxY  uint64 `json:"max_y"`
}

type RangeQueryResponse struct {
	Ranges []Range `json:"ranges"`
}

type Range struct {
	Min uint64 `json:"min"`
	Max uint64 `json:"max"`
}

type EstimateDistanceRequest struct {
	Order int    `json:"order"`
	Index1 uint64 `json:"index1"`
	Index2 uint64 `json:"index2"`
}

type EstimateDistanceResponse struct {
	Distance uint64 `json:"distance"`
}

type BatchPointToIndexRequest struct {
	Order     int      `json:"order"`
	Points    []Point  `json:"points"`
	Ascending bool     `json:"ascending"`
}

type BatchPointToIndexResponse struct {
	Indices []uint64 `json:"indices"`
}

type Point struct {
	X uint64 `json:"x"`
	Y uint64 `json:"y"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
