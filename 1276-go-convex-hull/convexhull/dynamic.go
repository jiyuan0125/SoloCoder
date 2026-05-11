package convexhull

type DynamicConvexHull struct {
	hull *ConvexHull
}

func NewDynamicConvexHull() *DynamicConvexHull {
	return &DynamicConvexHull{
		hull: &ConvexHull{Type: HullEmpty},
	}
}

func (d *DynamicConvexHull) Add(p Point) {
	if d.hull == nil {
		d.hull = &ConvexHull{Type: HullEmpty}
	}
	switch d.hull.Type {
	case HullEmpty:
		d.hull = &ConvexHull{Type: HullPoint, Points: []Point{p}}
	case HullPoint:
		if !p.Equal(d.hull.Points[0]) {
			pts := []Point{d.hull.Points[0], p}
			d.hull = &ConvexHull{Type: HullSegment, Points: pts}
		}
	case HullSegment:
		a, b := d.hull.Points[0], d.hull.Points[1]
		if p.Equal(a) || p.Equal(b) {
			return
		}
		if IsCollinear(a, b, p) {
			pts := []Point{a, b, p}
			d.hull = classifiyHull(pts)
		} else {
			pts := []Point{a, b, p}
			d.hull = Compute(pts)
		}
	case HullPolygon:
		if PointInConvexHull(p, d.hull) {
			return
		}
		all := make([]Point, len(d.hull.Points)+1)
		copy(all, d.hull.Points)
		all[len(all)-1] = p
		d.hull = Compute(all)
	}
}

func (d *DynamicConvexHull) Hull() *ConvexHull {
	return d.hull
}

func (d *DynamicConvexHull) AddPoints(points []Point) {
	for _, p := range points {
		d.Add(p)
	}
}
