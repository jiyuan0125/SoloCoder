package convexhull

import "math"

func Intersection(h1, h2 *ConvexHull) *ConvexHull {
	if h1 == nil || h2 == nil || h1.IsEmpty() || h2.IsEmpty() {
		return &ConvexHull{Type: HullEmpty}
	}
	if h1.IsPoint() {
		if PointInConvexHull(h1.Points[0], h2) {
			return &ConvexHull{Type: HullPoint, Points: []Point{h1.Points[0]}}
		}
		return &ConvexHull{Type: HullEmpty}
	}
	if h2.IsPoint() {
		if PointInConvexHull(h2.Points[0], h1) {
			return &ConvexHull{Type: HullPoint, Points: []Point{h2.Points[0]}}
		}
		return &ConvexHull{Type: HullEmpty}
	}
	if h1.IsSegment() && h2.IsSegment() {
		return segmentSegmentIntersection(h1.Points[0], h1.Points[1], h2.Points[0], h2.Points[1])
	}
	if h1.IsSegment() {
		return clipSegmentByConvex(h1.Points[0], h1.Points[1], h2)
	}
	if h2.IsSegment() {
		return clipSegmentByConvex(h2.Points[0], h2.Points[1], h1)
	}
	return sutherlandHodgman(h1.Points, h2.Points)
}

func segmentSegmentIntersection(a1, a2, b1, b2 Point) *ConvexHull {
	da := a2.Sub(a1)
	db := b2.Sub(b1)
	d := da.Cross(db)
	if Sign(d) == 0 {
		return collinearSegmentIntersection(a1, a2, b1, b2)
	}
	ab := b1.Sub(a1)
	t := ab.Cross(db) / d
	u := ab.Cross(da) / d
	if t >= -epsilon && t <= 1+epsilon && u >= -epsilon && u <= 1+epsilon {
		t = clamp01(t)
		pt := Point{a1.X + t*da.X, a1.Y + t*da.Y}
		return &ConvexHull{Type: HullPoint, Points: []Point{pt}}
	}
	return &ConvexHull{Type: HullEmpty}
}

func collinearSegmentIntersection(a1, a2, b1, b2 Point) *ConvexHull {
	if !IsCollinear(a1, a2, b1) {
		return &ConvexHull{Type: HullEmpty}
	}
	aMin, aMax := segmentMinMax(a1, a2)
	bMin, bMax := segmentMinMax(b1, b2)
	if aMax.Less(bMin) || bMax.Less(aMin) {
		return &ConvexHull{Type: HullEmpty}
	}
	start := bMin
	if aMin.Less(start) {
		start = aMin
	}
	end := aMax
	if end.Less(bMax) {
		end = bMax
	}
	if start.Equal(end) {
		return &ConvexHull{Type: HullPoint, Points: []Point{start}}
	}
	return &ConvexHull{Type: HullSegment, Points: []Point{start, end}}
}

func segmentMinMax(a, b Point) (Point, Point) {
	if a.Less(b) {
		return a, b
	}
	return b, a
}

func clipSegmentByConvex(a, b Point, hull *ConvexHull) *ConvexHull {
	if hull.IsPoint() {
		if pointOnSegment(hull.Points[0], a, b) {
			return &ConvexHull{Type: HullPoint, Points: []Point{hull.Points[0]}}
		}
		return &ConvexHull{Type: HullEmpty}
	}
	if hull.IsSegment() {
		return segmentSegmentIntersection(a, b, hull.Points[0], hull.Points[1])
	}
	pts := hull.Points
	n := len(pts)
	t0, t1 := 0.0, 1.0
	dir := b.Sub(a)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		edge := pts[j].Sub(pts[i])
		normal := Point{-edge.Y, edge.X}
		diff := a.Sub(pts[i])
		denom := normal.Dot(dir)
		numer := -normal.Dot(diff)
		if Sign(denom) == 0 {
			if Sign(numer) < 0 {
				return &ConvexHull{Type: HullEmpty}
			}
			continue
		}
		t := numer / denom
		if denom > 0 {
			if t > t1 {
				return &ConvexHull{Type: HullEmpty}
			}
			if t > t0 {
				t0 = t
			}
		} else {
			if t < t0 {
				return &ConvexHull{Type: HullEmpty}
			}
			if t < t1 {
				t1 = t
			}
		}
	}
	if t0 > t1+epsilon {
		return &ConvexHull{Type: HullEmpty}
	}
	t0 = clamp01(t0)
	t1 = clamp01(t1)
	p0 := Point{a.X + t0*dir.X, a.Y + t0*dir.Y}
	if math.Abs(t0-t1) < epsilon {
		return &ConvexHull{Type: HullPoint, Points: []Point{p0}}
	}
	p1 := Point{a.X + t1*dir.X, a.Y + t1*dir.Y}
	return &ConvexHull{Type: HullSegment, Points: []Point{p0, p1}}
}

func sutherlandHodgman(subject, clip []Point) *ConvexHull {
	output := make([]Point, len(subject))
	copy(output, subject)
	clipN := len(clip)
	for i := 0; i < clipN; i++ {
		if len(output) == 0 {
			break
		}
		j := (i + 1) % clipN
		cp1, cp2 := clip[i], clip[j]
		input := output
		output = nil
		s := input[len(input)-1]
		for _, e := range input {
			if inside(e, cp1, cp2) {
				if !inside(s, cp1, cp2) {
					intersect := lineLineIntersection(s, e, cp1, cp2)
					output = append(output, intersect)
				}
				output = append(output, e)
			} else if inside(s, cp1, cp2) {
				intersect := lineLineIntersection(s, e, cp1, cp2)
				output = append(output, intersect)
			}
			s = e
		}
	}
	unique := dedup(output)
	if len(unique) <= 2 {
		return classifiyHull(unique)
	}
	return Compute(unique)
}

func inside(p, a, b Point) bool {
	return Sign(Cross(a, b, p)) >= 0
}

func lineLineIntersection(a1, a2, b1, b2 Point) Point {
	da := a2.Sub(a1)
	db := b2.Sub(b1)
	ab := b1.Sub(a1)
	t := ab.Cross(db) / da.Cross(db)
	return Point{a1.X + t*da.X, a1.Y + t*da.Y}
}

func clamp01(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}
