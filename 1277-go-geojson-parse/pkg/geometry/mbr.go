package geometry

import "math"

func NewMBR(minLon, minLat, maxLon, maxLat float64) MBR {
	return MBR{
		MinLon: minLon,
		MinLat: minLat,
		MaxLon: maxLon,
		MaxLat: maxLat,
	}
}

func (m MBR) Width() float64 {
	if m.CrossesDateLine() {
		return (180 - m.MinLon) + (m.MaxLon + 180)
	}
	return m.MaxLon - m.MinLon
}

func (m MBR) Height() float64 {
	return m.MaxLat - m.MinLat
}

func (m MBR) Area() float64 {
	return m.Width() * m.Height()
}

func (m MBR) CrossesDateLine() bool {
	return m.MinLon > m.MaxLon
}

func (m MBR) Contains(p Point) bool {
	if m.CrossesDateLine() {
		if p.Lon >= m.MinLon || p.Lon <= m.MaxLon {
			return p.Lat >= m.MinLat && p.Lat <= m.MaxLat
		}
		return false
	}
	return p.Lon >= m.MinLon && p.Lon <= m.MaxLon &&
		p.Lat >= m.MinLat && p.Lat <= m.MaxLat
}

func (m MBR) Intersects(other MBR) bool {
	if m.CrossesDateLine() && other.CrossesDateLine() {
		return m.MaxLat >= other.MinLat && m.MinLat <= other.MaxLat
	}
	
	if m.CrossesDateLine() || other.CrossesDateLine() {
		latIntersects := m.MaxLat >= other.MinLat && m.MinLat <= other.MaxLat
		if !latIntersects {
			return false
		}
		
		var a, b MBR
		if m.CrossesDateLine() {
			a, b = m, other
		} else {
			a, b = other, m
		}
		
		return (b.MaxLon >= a.MinLon || b.MinLon <= a.MaxLon)
	}
	
	return m.MaxLon >= other.MinLon && m.MinLon <= other.MaxLon &&
		m.MaxLat >= other.MinLat && m.MinLat <= other.MaxLat
}

func (m MBR) Union(other MBR) MBR {
	var minLon, maxLon float64
	
	if m.CrossesDateLine() || other.CrossesDateLine() {
		a, b := m, other
		
		aContainsLeft := a.MinLon <= b.MinLon && a.MinLon <= b.MaxLon
		bContainsLeft := b.MinLon <= a.MinLon && b.MinLon <= a.MaxLon
		
		if (a.CrossesDateLine() && b.CrossesDateLine()) ||
			(a.CrossesDateLine() && (aContainsLeft || !bContainsLeft)) ||
			(b.CrossesDateLine() && (bContainsLeft || !aContainsLeft)) {
			minLon = math.Min(a.MinLon, b.MinLon)
			maxLon = math.Max(a.MaxLon, b.MaxLon)
		} else {
			normalized := normalizeDateLineMBR(m, other)
			minLon = normalized.MinLon
			maxLon = normalized.MaxLon
		}
	} else {
		minLon = math.Min(m.MinLon, other.MinLon)
		maxLon = math.Max(m.MaxLon, other.MaxLon)
	}
	
	minLat := math.Min(m.MinLat, other.MinLat)
	maxLat := math.Max(m.MaxLat, other.MaxLat)
	
	return NewMBR(minLon, minLat, maxLon, maxLat)
}

func normalizeDateLineMBR(a, b MBR) MBR {
	var leftMBR, rightMBR MBR
	if a.MaxLon < b.MinLon {
		leftMBR, rightMBR = a, b
	} else {
		leftMBR, rightMBR = b, a
	}
	
	width1 := (180 - rightMBR.MinLon) + (leftMBR.MaxLon + 180)
	width2 := rightMBR.MaxLon - leftMBR.MinLon
	
	if width1 < width2 {
		return NewMBR(rightMBR.MinLon, math.Min(a.MinLat, b.MinLat), leftMBR.MaxLon, math.Max(a.MaxLat, b.MaxLat))
	}
	
	return NewMBR(leftMBR.MinLon, math.Min(a.MinLat, b.MinLat), rightMBR.MaxLon, math.Max(a.MaxLat, b.MaxLat))
}

func (m MBR) ExpandToInclude(p Point) MBR {
	if m.CrossesDateLine() {
		if p.Lon > m.MinLon || p.Lon < m.MaxLon {
			return NewMBR(m.MinLon, math.Min(m.MinLat, p.Lat), m.MaxLon, math.Max(m.MaxLat, p.Lat))
		}
		
		dist1 := p.Lon - m.MaxLon
		dist2 := m.MinLon - p.Lon
		
		if dist1 < dist2 {
			return NewMBR(m.MinLon, math.Min(m.MinLat, p.Lat), p.Lon, math.Max(m.MaxLat, p.Lat))
		} else {
			return NewMBR(p.Lon, math.Min(m.MinLat, p.Lat), m.MaxLon, math.Max(m.MaxLat, p.Lat))
		}
	}
	
	minLon := math.Min(m.MinLon, p.Lon)
	maxLon := math.Max(m.MaxLon, p.Lon)
	minLat := math.Min(m.MinLat, p.Lat)
	maxLat := math.Max(m.MaxLat, p.Lat)
	
	return NewMBR(minLon, minLat, maxLon, maxLat)
}

func PointMBR(p Point) MBR {
	return NewMBR(p.Lon, p.Lat, p.Lon, p.Lat)
}

func PointsMBR(points []Point) MBR {
	if len(points) == 0 {
		return MBR{}
	}
	
	mbr := PointMBR(points[0])
	for i := 1; i < len(points); i++ {
		mbr = mbr.ExpandToInclude(points[i])
	}
	return mbr
}

func LineStringMBR(ls LineString) MBR {
	return PointsMBR(ls.Points)
}

func RingMBR(ring []Point) MBR {
	return PointsMBR(ring)
}

func PolygonMBR(poly Polygon) MBR {
	if len(poly.Rings) == 0 {
		return MBR{}
	}
	
	mbr := RingMBR(poly.Rings[0])
	for i := 1; i < len(poly.Rings); i++ {
		ringMBR := RingMBR(poly.Rings[i])
		mbr = mbr.Union(ringMBR)
	}
	return mbr
}

func GeometryMBR(g Geometry) MBR {
	switch g.Type {
	case GeometryTypePoint:
		if g.Point != nil {
			return PointMBR(*g.Point)
		}
	case GeometryTypeLineString:
		if g.LineString != nil {
			return LineStringMBR(*g.LineString)
		}
	case GeometryTypePolygon:
		if g.Polygon != nil {
			return PolygonMBR(*g.Polygon)
		}
	case GeometryTypeMultiPoint:
		if g.MultiPoint != nil {
			return PointsMBR(g.MultiPoint.Points)
		}
	case GeometryTypeMultiLineString:
		if g.MultiLineString != nil {
			if len(g.MultiLineString.LineStrings) == 0 {
				return MBR{}
			}
			mbr := LineStringMBR(g.MultiLineString.LineStrings[0])
			for i := 1; i < len(g.MultiLineString.LineStrings); i++ {
				mbr = mbr.Union(LineStringMBR(g.MultiLineString.LineStrings[i]))
			}
			return mbr
		}
	case GeometryTypeMultiPolygon:
		if g.MultiPolygon != nil {
			if len(g.MultiPolygon.Polygons) == 0 {
				return MBR{}
			}
			mbr := PolygonMBR(g.MultiPolygon.Polygons[0])
			for i := 1; i < len(g.MultiPolygon.Polygons); i++ {
				mbr = mbr.Union(PolygonMBR(g.MultiPolygon.Polygons[i]))
			}
			return mbr
		}
	}
	return MBR{}
}

func FeatureMBR(f Feature) MBR {
	return GeometryMBR(f.Geometry)
}
