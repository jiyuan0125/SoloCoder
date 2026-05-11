package coordconv

import (
	"errors"
	"math"
)

const (
	pi    = 3.14159265358979323846
	a     = 6378245.0
	ee    = 0.00669342162296594323
	xPi   = pi * 3000.0 / 180.0
	bd09k = 2.0 * pi / 365.2422
)

type CoordSystem string

const (
	WGS84 CoordSystem = "wgs84"
	GCJ02 CoordSystem = "gcj02"
	BD09  CoordSystem = "bd09"
)

type Point struct {
	Lng float64
	Lat float64
}

func ValidRange(lng, lat float64) error {
	if lng < -180 || lng > 180 {
		return errors.New("longitude out of range [-180, 180]")
	}
	if lat < -90 || lat > 90 {
		return errors.New("latitude out of range [-90, 90]")
	}
	return nil
}

func inChina(lng, lat float64) bool {
	return lng >= 72.004 && lng <= 137.8347 && lat >= 0.8293 && lat <= 55.8271
}

func delta(lng, lat float64) (float64, float64) {
	dLat := transformLat(lng-105.0, lat-35.0)
	dLng := transformLng(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(radLat) * pi)
	return dLng, dLat
}

func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*pi) + 20.0*math.Sin(2.0*x*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*pi) + 40.0*math.Sin(y/3.0*pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*pi) + 320*math.Sin(y*pi/30.0)) * 2.0 / 3.0
	return ret
}

func transformLng(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*pi) + 20.0*math.Sin(2.0*x*pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*pi) + 40.0*math.Sin(x/3.0*pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*pi) + 300.0*math.Sin(x/30.0*pi)) * 2.0 / 3.0
	return ret
}

func WGS84ToGCJ02(p Point) (Point, error) {
	if err := ValidRange(p.Lng, p.Lat); err != nil {
		return Point{}, err
	}
	if !inChina(p.Lng, p.Lat) {
		return p, nil
	}
	dLng, dLat := delta(p.Lng, p.Lat)
	return Point{
		Lng: roundTo6(p.Lng + dLng),
		Lat: roundTo6(p.Lat + dLat),
	}, nil
}

func GCJ02ToWGS84(p Point) (Point, error) {
	if err := ValidRange(p.Lng, p.Lat); err != nil {
		return Point{}, err
	}
	if !inChina(p.Lng, p.Lat) {
		return p, nil
	}
	dLng, dLat := delta(p.Lng, p.Lat)
	wgsLng := p.Lng - dLng
	wgsLat := p.Lat - dLat
	dLng, dLat = delta(wgsLng, wgsLat)
	return Point{
		Lng: roundTo6(p.Lng - dLng),
		Lat: roundTo6(p.Lat - dLat),
	}, nil
}

func GCJ02ToBD09(p Point) (Point, error) {
	if err := ValidRange(p.Lng, p.Lat); err != nil {
		return Point{}, err
	}
	if !inChina(p.Lng, p.Lat) {
		return p, nil
	}
	x := p.Lng
	y := p.Lat
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*xPi)
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*xPi)
	return Point{
		Lng: roundTo6(z*math.Cos(theta) + 0.0065),
		Lat: roundTo6(z*math.Sin(theta) + 0.006),
	}, nil
}

func BD09ToGCJ02(p Point) (Point, error) {
	if err := ValidRange(p.Lng, p.Lat); err != nil {
		return Point{}, err
	}
	if !inChina(p.Lng, p.Lat) {
		return p, nil
	}
	x := p.Lng - 0.0065
	y := p.Lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPi)
	return Point{
		Lng: roundTo6(z * math.Cos(theta)),
		Lat: roundTo6(z * math.Sin(theta)),
	}, nil
}

func WGS84ToBD09(p Point) (Point, error) {
	gcj, err := WGS84ToGCJ02(p)
	if err != nil {
		return Point{}, err
	}
	return GCJ02ToBD09(gcj)
}

func BD09ToWGS84(p Point) (Point, error) {
	gcj, err := BD09ToGCJ02(p)
	if err != nil {
		return Point{}, err
	}
	return GCJ02ToWGS84(gcj)
}

func Convert(from, to CoordSystem, p Point) (Point, error) {
	switch {
	case from == to:
		return p, nil
	case from == WGS84 && to == GCJ02:
		return WGS84ToGCJ02(p)
	case from == WGS84 && to == BD09:
		return WGS84ToBD09(p)
	case from == GCJ02 && to == WGS84:
		return GCJ02ToWGS84(p)
	case from == GCJ02 && to == BD09:
		return GCJ02ToBD09(p)
	case from == BD09 && to == WGS84:
		return BD09ToWGS84(p)
	case from == BD09 && to == GCJ02:
		return BD09ToGCJ02(p)
	default:
		return Point{}, errors.New("unsupported coordinate system conversion")
	}
}

func roundTo6(f float64) float64 {
	return math.Round(f*1e6) / 1e6
}
