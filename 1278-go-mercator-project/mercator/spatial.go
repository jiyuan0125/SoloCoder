package mercator

import (
	"math"
)

type DistanceResult struct {
	DistanceMeters     float64
	DistanceKilometers float64
	Warning            string
}

type AreaResult struct {
	AreaSquareMeters       float64
	AreaSquareKilometers   float64
	Warning                string
}

func HaversineDistance(lat1, lng1, lat2, lng2 float64) DistanceResult {
	var warning string
	
	lat1 = clampLat(lat1)
	lat2 = clampLat(lat2)
	lng1 = normalizeLng(lng1)
	lng2 = normalizeLng(lng2)
	
	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLng := (lng2 - lng1) * math.Pi / 180.0
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	distance := EarthRadiusMeters * c
	
	return DistanceResult{
		DistanceMeters:     distance,
		DistanceKilometers: distance / 1000.0,
		Warning:            warning,
	}
}

func SphericalPolygonArea(polygon []LatLng) AreaResult {
	var warning string
	
	if len(polygon) < 3 {
		return AreaResult{
			AreaSquareMeters:     0,
			AreaSquareKilometers: 0,
			Warning:              "polygon must have at least 3 points",
		}
	}
	
	if len(polygon) < 3 {
		return AreaResult{
			AreaSquareMeters:     0,
			AreaSquareKilometers: 0,
			Warning:              "polygon must have at least 3 points",
		}
	}
	
	var area float64
	n := len(polygon)
	
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		
		lat1 := clampLat(polygon[i].Lat)
		lng1 := normalizeLng(polygon[i].Lng)
		lat2 := clampLat(polygon[j].Lat)
		lng2 := normalizeLng(polygon[j].Lng)
		
		lat1Rad := lat1 * math.Pi / 180.0
		lng1Rad := lng1 * math.Pi / 180.0
		lat2Rad := lat2 * math.Pi / 180.0
		lng2Rad := lng2 * math.Pi / 180.0
		
		area += (lng2Rad - lng1Rad) * (2 + math.Sin(lat1Rad) + math.Sin(lat2Rad))
	}
	
	area = math.Abs(area * EarthRadiusMeters * EarthRadiusMeters / 2.0)
	
	return AreaResult{
		AreaSquareMeters:     area,
		AreaSquareKilometers: area / 1000000.0,
		Warning:              warning,
	}
}
