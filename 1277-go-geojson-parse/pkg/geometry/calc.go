package geometry

import "math"

const epsilon = 1e-9

func Distance(a, b Point) float64 {
	dlon := b.Lon - a.Lon
	dlat := b.Lat - a.Lat
	return math.Sqrt(dlon*dlon + dlat*dlat)
}

func SquaredDistance(a, b Point) float64 {
	dlon := b.Lon - a.Lon
	dlat := b.Lat - a.Lat
	return dlon*dlon + dlat*dlat
}

func PointDistanceToLineString(p Point, ls LineString) float64 {
	if len(ls.Points) < 2 {
		return math.Inf(1)
	}
	
	minDist := math.Inf(1)
	for i := 0; i < len(ls.Points)-1; i++ {
		dist := PointDistanceToSegment(p, ls.Points[i], ls.Points[i+1])
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

func PointDistanceToSegment(p, a, b Point) float64 {
	ap := NewPoint(p.Lon-a.Lon, p.Lat-a.Lat)
	ab := NewPoint(b.Lon-a.Lon, b.Lat-a.Lat)
	
	abLenSq := ab.Lon*ab.Lon + ab.Lat*ab.Lat
	if abLenSq < epsilon {
		return Distance(p, a)
	}
	
	t := (ap.Lon*ab.Lon + ap.Lat*ab.Lat) / abLenSq
	t = math.Max(0, math.Min(1, t))
	
	proj := NewPoint(a.Lon+t*ab.Lon, a.Lat+t*ab.Lat)
	return Distance(p, proj)
}

func PointDistanceToPolygon(p Point, poly Polygon) float64 {
	if len(poly.Rings) == 0 {
		return math.Inf(1)
	}
	
	if PointInPolygon(p, poly) {
		return 0
	}
	
	minDist := math.Inf(1)
	for _, ring := range poly.Rings {
		dist := PointDistanceToRing(p, ring)
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

func PointDistanceToRing(p Point, ring []Point) float64 {
	if len(ring) < 2 {
		return math.Inf(1)
	}
	
	minDist := math.Inf(1)
	for i := 0; i < len(ring)-1; i++ {
		dist := PointDistanceToSegment(p, ring[i], ring[i+1])
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

func PointDistanceToGeometry(p Point, g Geometry) float64 {
	switch g.Type {
	case GeometryTypePoint:
		if g.Point != nil {
			return Distance(p, *g.Point)
		}
	case GeometryTypeLineString:
		if g.LineString != nil {
			return PointDistanceToLineString(p, *g.LineString)
		}
	case GeometryTypePolygon:
		if g.Polygon != nil {
			return PointDistanceToPolygon(p, *g.Polygon)
		}
	case GeometryTypeMultiPoint:
		if g.MultiPoint != nil {
			minDist := math.Inf(1)
			for _, pt := range g.MultiPoint.Points {
				dist := Distance(p, pt)
				if dist < minDist {
					minDist = dist
				}
			}
			return minDist
		}
	case GeometryTypeMultiLineString:
		if g.MultiLineString != nil {
			minDist := math.Inf(1)
			for _, ls := range g.MultiLineString.LineStrings {
				dist := PointDistanceToLineString(p, ls)
				if dist < minDist {
					minDist = dist
				}
			}
			return minDist
		}
	case GeometryTypeMultiPolygon:
		if g.MultiPolygon != nil {
			minDist := math.Inf(1)
			for _, poly := range g.MultiPolygon.Polygons {
				dist := PointDistanceToPolygon(p, poly)
				if dist < minDist {
					minDist = dist
				}
			}
			return minDist
		}
	}
	return math.Inf(1)
}

func PointInPolygon(p Point, poly Polygon) bool {
	if len(poly.Rings) == 0 {
		return false
	}
	
	if !PointInRing(p, poly.Rings[0]) {
		return false
	}
	
	for i := 1; i < len(poly.Rings); i++ {
		if PointInRing(p, poly.Rings[i]) {
			return false
		}
	}
	
	return true
}

func PointInRing(p Point, ring []Point) bool {
	if len(ring) < 4 {
		return false
	}
	
	n := len(ring)
	inside := false
	
	for i, j := 0, n-1; i < n; i, j = i+1, i {
		ri := ring[i]
		rj := ring[j]
		
		if ((ri.Lat > p.Lat) != (rj.Lat > p.Lat)) &&
			(p.Lon < (rj.Lon-ri.Lon)*(p.Lat-ri.Lat)/(rj.Lat-ri.Lat)+ri.Lon) {
			inside = !inside
		}
	}
	
	return inside
}

func SegmentsIntersect(a1, a2, b1, b2 Point) bool {
	d1 := Direction(b1, b2, a1)
	d2 := Direction(b1, b2, a2)
	d3 := Direction(a1, a2, b1)
	d4 := Direction(a1, a2, b2)
	
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	
	if d1 == 0 && OnSegment(b1, b2, a1) {
		return true
	}
	if d2 == 0 && OnSegment(b1, b2, a2) {
		return true
	}
	if d3 == 0 && OnSegment(a1, a2, b1) {
		return true
	}
	if d4 == 0 && OnSegment(a1, a2, b2) {
		return true
	}
	
	return false
}

func Direction(pi, pj, pk Point) float64 {
	return (pk.Lon-pi.Lon)*(pj.Lat-pi.Lat) - (pj.Lon-pi.Lon)*(pk.Lat-pi.Lat)
}

func OnSegment(pi, pj, pk Point) bool {
	return math.Min(pi.Lon, pj.Lon) <= pk.Lon+epsilon && pk.Lon-epsilon <= math.Max(pi.Lon, pj.Lon) &&
		math.Min(pi.Lat, pj.Lat) <= pk.Lat+epsilon && pk.Lat-epsilon <= math.Max(pi.Lat, pj.Lat)
}

func LineStringsIntersect(a, b LineString) bool {
	for i := 0; i < len(a.Points)-1; i++ {
		for j := 0; j < len(b.Points)-1; j++ {
			if SegmentsIntersect(a.Points[i], a.Points[i+1], b.Points[j], b.Points[j+1]) {
				return true
			}
		}
	}
	return false
}

func RingsIntersect(a, b []Point) bool {
	for i := 0; i < len(a)-1; i++ {
		for j := 0; j < len(b)-1; j++ {
			if SegmentsIntersect(a[i], a[i+1], b[j], b[j+1]) {
				return true
			}
		}
	}
	return false
}

func PolygonsIntersect(a, b Polygon) bool {
	if len(a.Rings) == 0 || len(b.Rings) == 0 {
		return false
	}
	
	mbrA := PolygonMBR(a)
	mbrB := PolygonMBR(b)
	if !mbrA.Intersects(mbrB) {
		return false
	}
	
	if len(a.Rings[0]) > 0 && PointInPolygon(a.Rings[0][0], b) {
		return true
	}
	if len(b.Rings[0]) > 0 && PointInPolygon(b.Rings[0][0], a) {
		return true
	}
	
	for _, ringA := range a.Rings {
		for _, ringB := range b.Rings {
			if RingsIntersect(ringA, ringB) {
				return true
			}
		}
	}
	
	return false
}

func GeometriesIntersect(a, b Geometry) bool {
	mbrA := GeometryMBR(a)
	mbrB := GeometryMBR(b)
	if !mbrA.Intersects(mbrB) {
		return false
	}
	
	switch a.Type {
	case GeometryTypePoint:
		if a.Point == nil {
			return false
		}
		switch b.Type {
		case GeometryTypePoint:
			return b.Point != nil && math.Abs(a.Point.Lon-b.Point.Lon) < epsilon && math.Abs(a.Point.Lat-b.Point.Lat) < epsilon
		case GeometryTypeLineString:
			return b.LineString != nil && PointDistanceToLineString(*a.Point, *b.LineString) < epsilon
		case GeometryTypePolygon:
			return b.Polygon != nil && PointInPolygon(*a.Point, *b.Polygon)
		case GeometryTypeMultiPoint:
			if b.MultiPoint == nil {
				return false
			}
			for _, p := range b.MultiPoint.Points {
				if math.Abs(a.Point.Lon-p.Lon) < epsilon && math.Abs(a.Point.Lat-p.Lat) < epsilon {
					return true
				}
			}
			return false
		case GeometryTypeMultiLineString:
			if b.MultiLineString == nil {
				return false
			}
			for _, ls := range b.MultiLineString.LineStrings {
				if PointDistanceToLineString(*a.Point, ls) < epsilon {
					return true
				}
			}
			return false
		case GeometryTypeMultiPolygon:
			if b.MultiPolygon == nil {
				return false
			}
			for _, poly := range b.MultiPolygon.Polygons {
				if PointInPolygon(*a.Point, poly) {
					return true
				}
			}
			return false
		}
	case GeometryTypeLineString:
		if a.LineString == nil {
			return false
		}
		switch b.Type {
		case GeometryTypeLineString:
			return b.LineString != nil && LineStringsIntersect(*a.LineString, *b.LineString)
		case GeometryTypePolygon:
			return b.Polygon != nil && lineStringPolygonIntersect(*a.LineString, *b.Polygon)
		case GeometryTypeMultiLineString:
			if b.MultiLineString == nil {
				return false
			}
			for _, ls := range b.MultiLineString.LineStrings {
				if LineStringsIntersect(*a.LineString, ls) {
					return true
				}
			}
			return false
		case GeometryTypeMultiPolygon:
			if b.MultiPolygon == nil {
				return false
			}
			for _, poly := range b.MultiPolygon.Polygons {
				if lineStringPolygonIntersect(*a.LineString, poly) {
					return true
				}
			}
			return false
		}
	case GeometryTypePolygon:
		if a.Polygon == nil {
			return false
		}
		switch b.Type {
		case GeometryTypePolygon:
			return b.Polygon != nil && PolygonsIntersect(*a.Polygon, *b.Polygon)
		case GeometryTypeMultiPolygon:
			if b.MultiPolygon == nil {
				return false
			}
			for _, poly := range b.MultiPolygon.Polygons {
				if PolygonsIntersect(*a.Polygon, poly) {
					return true
				}
			}
			return false
		}
	case GeometryTypeMultiPoint:
		if a.MultiPoint == nil {
			return false
		}
		for _, p := range a.MultiPoint.Points {
			if pointGeometryIntersect(p, b) {
				return true
			}
		}
		return false
	case GeometryTypeMultiLineString:
		if a.MultiLineString == nil {
			return false
		}
		for _, ls := range a.MultiLineString.LineStrings {
			if lineStringGeometryIntersect(ls, b) {
				return true
			}
		}
		return false
	case GeometryTypeMultiPolygon:
		if a.MultiPolygon == nil {
			return false
		}
		for _, poly := range a.MultiPolygon.Polygons {
			if polygonGeometryIntersect(poly, b) {
				return true
			}
		}
		return false
	}
	
	return false
}

func pointGeometryIntersect(p Point, g Geometry) bool {
	switch g.Type {
	case GeometryTypePoint:
		return g.Point != nil && math.Abs(p.Lon-g.Point.Lon) < epsilon && math.Abs(p.Lat-g.Point.Lat) < epsilon
	case GeometryTypeLineString:
		return g.LineString != nil && PointDistanceToLineString(p, *g.LineString) < epsilon
	case GeometryTypePolygon:
		return g.Polygon != nil && PointInPolygon(p, *g.Polygon)
	}
	return false
}

func lineStringGeometryIntersect(ls LineString, g Geometry) bool {
	switch g.Type {
	case GeometryTypeLineString:
		return g.LineString != nil && LineStringsIntersect(ls, *g.LineString)
	case GeometryTypePolygon:
		return g.Polygon != nil && lineStringPolygonIntersect(ls, *g.Polygon)
	}
	return false
}

func polygonGeometryIntersect(poly Polygon, g Geometry) bool {
	switch g.Type {
	case GeometryTypePolygon:
		return g.Polygon != nil && PolygonsIntersect(poly, *g.Polygon)
	}
	return false
}

func lineStringPolygonIntersect(ls LineString, poly Polygon) bool {
	if len(poly.Rings) == 0 {
		return false
	}
	
	if len(ls.Points) > 0 && PointInPolygon(ls.Points[0], poly) {
		return true
	}
	
	for _, ring := range poly.Rings {
		for i := 0; i < len(ring)-1; i++ {
			for j := 0; j < len(ls.Points)-1; j++ {
				if SegmentsIntersect(ring[i], ring[i+1], ls.Points[j], ls.Points[j+1]) {
					return true
				}
			}
		}
	}
	
	return false
}

func RingArea(ring []Point) float64 {
	if len(ring) < 3 {
		return 0
	}
	
	n := len(ring)
	area := 0.0
	
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += ring[i].Lon * ring[j].Lat
		area -= ring[j].Lon * ring[i].Lat
	}
	
	return math.Abs(area) / 2.0
}

func RingSignedArea(ring []Point) float64 {
	if len(ring) < 3 {
		return 0
	}
	
	n := len(ring)
	area := 0.0
	
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += ring[i].Lon * ring[j].Lat
		area -= ring[j].Lon * ring[i].Lat
	}
	
	return area / 2.0
}

func RingIsClockwise(ring []Point) bool {
	return RingSignedArea(ring) < 0
}

func RingIsCounterClockwise(ring []Point) bool {
	return RingSignedArea(ring) > 0
}

func ReverseRing(ring []Point) {
	for i, j := 0, len(ring)-1; i < j; i, j = i+1, j-1 {
		ring[i], ring[j] = ring[j], ring[i]
	}
}

func FixPolygonRingOrientation(poly *Polygon) {
	if poly == nil || len(poly.Rings) == 0 {
		return
	}
	
	if len(poly.Rings[0]) > 0 && RingIsClockwise(poly.Rings[0]) {
		ReverseRing(poly.Rings[0])
	}
	
	for i := 1; i < len(poly.Rings); i++ {
		if len(poly.Rings[i]) > 0 && RingIsCounterClockwise(poly.Rings[i]) {
			ReverseRing(poly.Rings[i])
		}
	}
}

func EnsureRingClosed(ring []Point) []Point {
	if len(ring) == 0 {
		return ring
	}
	
	first := ring[0]
	last := ring[len(ring)-1]
	
	if math.Abs(first.Lon-last.Lon) < epsilon && math.Abs(first.Lat-last.Lat) < epsilon {
		return ring
	}
	
	closed := make([]Point, len(ring)+1)
	copy(closed, ring)
	closed[len(ring)] = first
	return closed
}

func PolygonArea(poly Polygon) float64 {
	if len(poly.Rings) == 0 {
		return 0
	}
	
	area := RingArea(poly.Rings[0])
	
	for i := 1; i < len(poly.Rings); i++ {
		area -= RingArea(poly.Rings[i])
	}
	
	if area < 0 {
		return 0
	}
	return area
}
