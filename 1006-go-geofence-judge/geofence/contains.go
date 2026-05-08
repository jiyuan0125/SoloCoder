package geofence

import "math"

func (p Polygon) Contains(point Point) bool {
	n := len(p.Points)
	if n == 0 {
		return false
	}

	shiftedPoly, shiftedPoint := p.shiftForDateline(point)

	points := append(shiftedPoly.Points, shiftedPoly.Points[0])

	for i := 0; i < n; i++ {
		if pointOnSegment(shiftedPoint, points[i], points[i+1]) {
			return true
		}
	}

	count := 0
	for i := 0; i < n; i++ {
		a := points[i]
		b := points[i+1]

		if math.Abs(a.Lat-b.Lat) < Epsilon {
			continue
		}

		if shiftedPoint.Lat > math.Max(a.Lat, b.Lat)+Epsilon {
			continue
		}
		if shiftedPoint.Lat < math.Min(a.Lat, b.Lat)-Epsilon {
			continue
		}

		t := (shiftedPoint.Lat - a.Lat) / (b.Lat - a.Lat)
		intersectLng := a.Lng + t*(b.Lng-a.Lng)

		if intersectLng >= shiftedPoint.Lng-Epsilon {
			if a.Lat < shiftedPoint.Lat+Epsilon && b.Lat > shiftedPoint.Lat-Epsilon {
				count++
			} else if b.Lat < shiftedPoint.Lat+Epsilon && a.Lat > shiftedPoint.Lat-Epsilon {
				count++
			}
		}
	}

	return count%2 == 1
}

func (p Polygon) shiftForDateline(point Point) (Polygon, Point) {
	normalized := make([]Point, len(p.Points))
	for i, pt := range p.Points {
		normalized[i] = NewPoint(pt.Lat, pt.Lng)
	}

	normalizedPoint := NewPoint(point.Lat, point.Lng)

	hasEast := false
	hasWest := false
	for _, pt := range normalized {
		if pt.Lng > 90 {
			hasEast = true
		}
		if pt.Lng < -90 {
			hasWest = true
		}
	}

	if !hasEast || !hasWest {
		return Polygon{Points: normalized}, normalizedPoint
	}

	shift := 180.0
	shiftedPoly := make([]Point, len(normalized))
	for i, pt := range normalized {
		shiftedPoly[i] = NewPoint(pt.Lat, pt.Lng+shift)
	}
	shiftedPt := NewPoint(normalizedPoint.Lat, normalizedPoint.Lng+shift)

	return Polygon{Points: shiftedPoly}, shiftedPt
}

func (p Polygon) MinDistance(point Point) float64 {
	n := len(p.Points)
	if n == 0 {
		return math.Inf(1)
	}

	points := append(p.Points, p.Points[0])
	minDist := math.Inf(1)

	for i := 0; i < n; i++ {
		dist := distanceToSegment(point, points[i], points[i+1])
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}
