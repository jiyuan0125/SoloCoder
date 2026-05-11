package mercator

import (
	"math"
)

type TileResult struct {
	X       int64
	Y       int64
	Z       int
	Warning string
}

type TileBoundsResult struct {
	North   float64
	South   float64
	East    float64
	West    float64
	Warning string
}

func LatLngToTile(lat, lng float64, zoom int) TileResult {
	var warning string
	
	if zoom < MinZoom || zoom > MaxZoom {
		zoom = clampZoom(zoom)
		warning = "zoom level out of range, clamped"
	}
	
	lat = clampLat(lat)
	lng = normalizeLng(lng)
	
	latRad := lat * math.Pi / 180.0
	n := math.Pow(2.0, float64(zoom))
	
	xtile := int64(math.Floor((lng + 180.0) / 360.0 * n))
	ytile := int64(math.Floor((1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n))
	
	maxTile := int64(n) - 1
	if xtile > maxTile {
		xtile = 0
	}
	
	if xtile < 0 {
		xtile = 0
	}
	if ytile < 0 {
		ytile = 0
	}
	if ytile > maxTile {
		ytile = maxTile
	}
	
	return TileResult{X: xtile, Y: ytile, Z: zoom, Warning: warning}
}

func TileToLatLngBounds(tileX, tileY int64, zoom int) TileBoundsResult {
	var warning string
	
	if zoom < MinZoom || zoom > MaxZoom {
		zoom = clampZoom(zoom)
		warning = "zoom level out of range, clamped"
	}
	
	n := math.Pow(2.0, float64(zoom))
	
	west := tileXToLng(tileX, n)
	east := tileXToLng(tileX+1, n)
	north := tileYToLat(tileY, n)
	south := tileYToLat(tileY+1, n)
	
	return TileBoundsResult{
		North:   north,
		South:   south,
		East:    east,
		West:    west,
		Warning: warning,
	}
}

func tileXToLng(x int64, n float64) float64 {
	return float64(x)/n*360.0 - 180.0
}

func tileYToLat(y int64, n float64) float64 {
	return math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n))) * 180.0 / math.Pi
}

func clampZoom(zoom int) int {
	if zoom < MinZoom {
		return MinZoom
	}
	if zoom > MaxZoom {
		return MaxZoom
	}
	return zoom
}
