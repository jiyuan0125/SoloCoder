package geofence

import "errors"

var (
	ErrInvalidVertexCount = errors.New("polygon must have at least 3 vertices")
	ErrSelfIntersecting   = errors.New("polygon is self-intersecting")
)

func (p Polygon) Validate() error {
	if len(p.Points) < 3 {
		return ErrInvalidVertexCount
	}

	if p.hasSelfIntersection() {
		return ErrSelfIntersecting
	}

	return nil
}

func (p Polygon) hasSelfIntersection() bool {
	n := len(p.Points)
	points := append(p.Points, p.Points[0])

	for i := 0; i < n; i++ {
		for j := i + 2; j < n; j++ {
			if i == 0 && j == n-1 {
				continue
			}
			if segmentsIntersect(
				points[i], points[i+1],
				points[j], points[j+1],
			) {
				return true
			}
		}
	}
	return false
}
