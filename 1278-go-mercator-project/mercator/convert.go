package mercator

import (
	"fmt"
	"math"
)

type ConvertResult struct {
	X       float64
	Y       float64
	Warning string
}

func LatLngToMercator(lat, lng float64) ConvertResult {
	var warning string
	
	if lat > LatMax {
		lat = LatMax
		warning = fmt.Sprintf("latitude exceeds max value, clamped to %.6f", LatMax)
	} else if lat < LatMin {
		lat = LatMin
		warning = fmt.Sprintf("latitude exceeds min value, clamped to %.6f", LatMin)
	}
	
	latRad := lat * math.Pi / 180.0
	lngRad := lng * math.Pi / 180.0
	
	x := EarthRadiusMeters * lngRad
	
	sinLat := math.Sin(latRad)
	y := EarthRadiusMeters * math.Log((1+sinLat)/(1-sinLat)) / 2
	
	y = clampY(y)
	
	return ConvertResult{X: x, Y: y, Warning: warning}
}

func MercatorToLatLng(x, y float64) ConvertResult {
	var warning string
	
	if y > MercatorMaxY {
		y = MercatorMaxY
		warning = "mercator Y exceeds max value, clamped"
	} else if y < MercatorMinY {
		y = MercatorMinY
		warning = "mercator Y exceeds min value, clamped"
	}
	
	lng := x / EarthRadiusMeters
	lat := math.Atan(math.Sinh(y / EarthRadiusMeters))
	
	lngDeg := lng * 180.0 / math.Pi
	latDeg := lat * 180.0 / math.Pi
	
	latDeg = clampLat(latDeg)
	lngDeg = normalizeLng(lngDeg)
	
	return ConvertResult{X: lngDeg, Y: latDeg, Warning: warning}
}

func clampY(y float64) float64 {
	if y > MercatorMaxY {
		return MercatorMaxY
	}
	if y < MercatorMinY {
		return MercatorMinY
	}
	return y
}

func clampLat(lat float64) float64 {
	if lat > LatMax {
		return LatMax
	}
	if lat < LatMin {
		return LatMin
	}
	return lat
}

func normalizeLng(lng float64) float64 {
	for lng > 180.0 {
		lng -= 360.0
	}
	for lng < -180.0 {
		lng += 360.0
	}
	return lng
}
