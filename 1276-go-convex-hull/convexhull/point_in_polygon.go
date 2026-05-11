package convexhull

func PointInConvexHull(p Point, hull *ConvexHull) bool {
	if hull == nil {
		return false
	}
	switch hull.Type {
	case HullEmpty:
		return false
	case HullPoint:
		return p.Equal(hull.Points[0])
	case HullSegment:
		a, b := hull.Points[0], hull.Points[1]
		return pointOnSegment(p, a, b)
	case HullPolygon:
		return pointInConvexPolygon(p, hull.Points)
	default:
		return false
	}
}

func pointOnSegment(p, a, b Point) bool {
	if !IsCollinear(a, b, p) {
		return false
	}
	return (p.X-min(a.X, b.X) >= -epsilon) && (max(a.X, b.X)-p.X >= -epsilon) &&
		(p.Y-min(a.Y, b.Y) >= -epsilon) && (max(a.Y, b.Y)-p.Y >= -epsilon)
}

func pointInConvexPolygon(p Point, hull []Point) bool {
	n := len(hull)
	if n == 0 {
		return false
	}
	if n == 1 {
		return p.Equal(hull[0])
	}
	if n == 2 {
		return pointOnSegment(p, hull[0], hull[1])
	}
	hasNeg := false
	hasPos := false
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		cr := Sign(Cross(hull[i], hull[j], p))
		if cr < 0 {
			hasNeg = true
		} else if cr > 0 {
			hasPos = true
		}
		if hasNeg && hasPos {
			return false
		}
	}
	return true
}
