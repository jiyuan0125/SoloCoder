package convexhull

func FilterBoundaryPoints(points []Point) []Point {
	hull := Compute(points)
	if hull.IsEmpty() {
		return nil
	}
	result := make([]Point, len(hull.Points))
	copy(result, hull.Points)
	return result
}
