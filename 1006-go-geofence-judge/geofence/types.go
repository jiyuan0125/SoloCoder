package geofence

import "math"

const (
	EarthRadius   = 6371000.0
	MinArea       = 100.0
	CoordPrecision = 1000000.0
	Epsilon       = 1e-9
)

type Point struct {
	Lat float64
	Lng float64
}

type Polygon struct {
	Points []Point
}

type Fence struct {
	ID          string
	Name        string
	Description string
	Polygon     Polygon
	Area        float64
}

func NewPoint(lat, lng float64) Point {
	return Point{
		Lat: round(lat, 6),
		Lng: normalizeLng(lng),
	}
}

func normalizeLng(lng float64) float64 {
	for lng > 180 {
		lng -= 360
	}
	for lng < -180 {
		lng += 360
	}
	return round(lng, 6)
}

func round(x float64, n int) float64 {
	shift := math.Pow(10, float64(n))
	return math.Round(x*shift) / shift
}

func (p Point) ToRadians() (float64, float64) {
	return p.Lat * math.Pi / 180, p.Lng * math.Pi / 180
}
