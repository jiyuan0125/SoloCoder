package convexhull

import (
	"sort"
)

type HullType int

const (
	HullEmpty HullType = iota
	HullPoint
	HullSegment
	HullPolygon
)

type ConvexHull struct {
	Points []Point
	Type   HullType
}

func (h *ConvexHull) IsEmpty() bool   { return h.Type == HullEmpty }
func (h *ConvexHull) IsPoint() bool   { return h.Type == HullPoint }
func (h *ConvexHull) IsSegment() bool { return h.Type == HullSegment }
func (h *ConvexHull) IsPolygon() bool { return h.Type == HullPolygon }
func (h *ConvexHull) Len() int        { return len(h.Points) }

func dedup(points []Point) []Point {
	if len(points) == 0 {
		return points
	}
	sorted := make([]Point, len(points))
	copy(sorted, points)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Less(sorted[j])
	})
	result := []Point{sorted[0]}
	for i := 1; i < len(sorted); i++ {
		if !sorted[i].Equal(sorted[i-1]) {
			result = append(result, sorted[i])
		}
	}
	return result
}

func andrewScan(points []Point) []Point {
	n := len(points)
	if n <= 1 {
		return points
	}
	lower := make([]Point, 0, n)
	for _, p := range points {
		for len(lower) >= 2 {
			a, b := lower[len(lower)-2], lower[len(lower)-1]
			cr := Cross(a, b, p)
			if Sign(cr) <= 0 {
				lower = lower[:len(lower)-1]
			} else {
				break
			}
		}
		lower = append(lower, p)
	}
	upper := make([]Point, 0, n)
	for i := n - 1; i >= 0; i-- {
		p := points[i]
		for len(upper) >= 2 {
			a, b := upper[len(upper)-2], upper[len(upper)-1]
			cr := Cross(a, b, p)
			if Sign(cr) <= 0 {
				upper = upper[:len(upper)-1]
			} else {
				break
			}
		}
		upper = append(upper, p)
	}
	if len(lower) > 0 && len(upper) > 0 {
		lower = lower[:len(lower)-1]
		upper = upper[:len(upper)-1]
	}
	return append(lower, upper...)
}

func classifiyHull(points []Point) *ConvexHull {
	n := len(points)
	if n == 0 {
		return &ConvexHull{Type: HullEmpty, Points: nil}
	}
	if n == 1 {
		return &ConvexHull{Type: HullPoint, Points: points}
	}
	if n == 2 {
		return &ConvexHull{Type: HullSegment, Points: points}
	}
	allCollinear := true
	a, b := points[0], points[1]
	for i := 2; i < n; i++ {
		if !IsCollinear(a, b, points[i]) {
			allCollinear = false
			break
		}
	}
	if allCollinear {
		minP, maxP := points[0], points[0]
		for _, p := range points {
			if p.Less(minP) {
				minP = p
			}
			if maxP.Less(p) {
				maxP = p
			}
		}
		return &ConvexHull{Type: HullSegment, Points: []Point{minP, maxP}}
	}
	return &ConvexHull{Type: HullPolygon, Points: points}
}

func Compute(points []Point) *ConvexHull {
	unique := dedup(points)
	n := len(unique)
	if n <= 2 {
		return classifiyHull(unique)
	}
	sorted := make([]Point, n)
	copy(sorted, unique)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Less(sorted[j])
	})
	hull := andrewScan(sorted)
	return classifiyHull(hull)
}

func Reverse(hull []Point) []Point {
	if len(hull) == 0 {
		return hull
	}
	result := make([]Point, len(hull))
	for i, j := 0, len(hull)-1; i < len(hull); i, j = i+1, j-1 {
		result[i] = hull[j]
	}
	return result
}
