package annealing

import (
	"math"
)

type CoordinateType string

const (
	CoordinateTypeGeographic CoordinateType = "geographic"
	CoordinateTypeCartesian  CoordinateType = "cartesian"
)

type City struct {
	Name       string
	Latitude   float64
	Longitude  float64
	X          float64
	Y          float64
	CoordinateType CoordinateType
}

const (
	earthRadiusKm = 6371.0
	epsilon       = 1e-10
)

func (c City) Distance(other City) float64 {
	if c.CoordinateType == CoordinateTypeGeographic {
		return haversineDistance(c.Latitude, c.Longitude, other.Latitude, other.Longitude)
	}
	return cartesianDistance(c.X, c.Y, other.X, other.Y)
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := toRadians(lat1)
	lat2Rad := toRadians(lat2)
	lon1Rad := toRadians(lon1)
	lon2Rad := toRadians(lon2)

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func cartesianDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func CalculateTotalDistance(cities []City, order []int) float64 {
	if len(cities) <= 1 {
		return 0.0
	}
	var total float64
	n := len(cities)
	for i := 0; i < n; i++ {
		from := cities[order[i]]
		to := cities[order[(i+1)%n]]
		total += from.Distance(to)
	}
	return total
}

func IsSolutionBetter(current, candidate float64) bool {
	return candidate < current-epsilon
}
