package geofence

import "math"

func (p Polygon) Area() float64 {
	n := len(p.Points)
	if n < 3 {
		return 0
	}

	sum := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n

		lat1, lng1 := p.Points[i].ToRadians()
		lat2, lng2 := p.Points[j].ToRadians()

		sum += (lng2 - lng1) * (2 + math.Sin(lat1) + math.Sin(lat2))
	}

	area := math.Abs(sum * EarthRadius * EarthRadius / 2.0)
	return area
}
