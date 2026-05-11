package peucker

import (
	"math"

	"github.com/solo-coder/douglas-peucker/internal/api"
)

func PointToSegmentDistance(point, start, end api.Point) float64 {
	dx := end.Lon - start.Lon
	dy := end.Lat - start.Lat

	if dx == 0 && dy == 0 {
		return HaversineDistance(point, start)
	}

	t := ((point.Lon-start.Lon)*dx + (point.Lat-start.Lat)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))

	projX := start.Lon + t*dx
	projY := start.Lat + t*dy

	projPoint := api.Point{Lat: projY, Lon: projX}

	return HaversineDistance(point, projPoint)
}

func ShoelaceArea(points []api.Point) float64 {
	n := len(points)
	if n < 3 {
		return 0
	}

	var area float64
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += points[i].Lon*points[j].Lat - points[j].Lon*points[i].Lat
	}

	area = math.Abs(area) / 2

	minLat, maxLat := points[0].Lat, points[0].Lat
	for _, p := range points {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
	}
	centerLat := toRadians((minLat + maxLat) / 2)
	latScale := math.Cos(centerLat)

	degToMeter := earthRadius * math.Pi / 180.0
	areaScale := degToMeter * degToMeter * latScale

	return area * areaScale
}
