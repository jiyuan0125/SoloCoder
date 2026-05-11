package convexhull

import "math"

func Perimeter(hull *ConvexHull) float64 {
	if hull == nil || hull.Len() <= 1 {
		return 0
	}
	if hull.Len() == 2 {
		return hull.Points[0].Distance(hull.Points[1])
	}
	total := 0.0
	n := len(hull.Points)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		total += hull.Points[i].Distance(hull.Points[j])
	}
	return total
}

func Area(hull *ConvexHull) float64 {
	if hull == nil || !hull.IsPolygon() {
		return 0
	}
	total := 0.0
	n := len(hull.Points)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		total += hull.Points[i].X*hull.Points[j].Y - hull.Points[j].X*hull.Points[i].Y
	}
	return math.Abs(total) / 2.0
}
