package peucker

import (
	"math"

	"github.com/solo-coder/douglas-peucker/internal/api"
)

const earthRadius = 6371000.0

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func HaversineDistance(p1, p2 api.Point) float64 {
	lat1 := toRadians(p1.Lat)
	lat2 := toRadians(p2.Lat)
	deltaLat := toRadians(p2.Lat - p1.Lat)
	deltaLon := toRadians(p2.Lon - p1.Lon)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func CalculateTotalLength(points []api.Point) float64 {
	if len(points) < 2 {
		return 0
	}

	var total float64
	for i := 1; i < len(points); i++ {
		total += HaversineDistance(points[i-1], points[i])
	}
	return total
}
