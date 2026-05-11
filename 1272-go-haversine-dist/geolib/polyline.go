package geolib

import (
	"math"
	"sort"
)

const (
	minDistanceThresholdMeters = 1.0
	minDistanceThresholdKm     = minDistanceThresholdMeters / 1000.0
)

type PointResult struct {
	Point      Coordinate
	DistanceKm float64
}

type Coordinate struct {
	Lat float64
	Lng float64
}

func SimplifyPolyline(points []Coordinate) []Coordinate {
	if len(points) <= 2 {
		result := make([]Coordinate, len(points))
		copy(result, points)
		return result
	}

	simplified := []Coordinate{points[0]}

	for i := 1; i < len(points); i++ {
		prev := simplified[len(simplified)-1]
		curr := points[i]

		dist := HaversineDistance(prev.Lat, prev.Lng, curr.Lat, curr.Lng)

		if dist >= minDistanceThresholdKm {
			simplified = append(simplified, curr)
		}
	}

	lastOriginal := points[len(points)-1]
	lastSimplified := simplified[len(simplified)-1]
	if len(simplified) > 0 && (lastSimplified.Lat != lastOriginal.Lat || lastSimplified.Lng != lastOriginal.Lng) {
		simplified = append(simplified, lastOriginal)
	}

	return simplified
}

func PolylineLength(points []Coordinate, useEllipsoid bool) (float64, int, int, error) {
	if len(points) < 2 {
		return 0, len(points), len(points), nil
	}

	for _, p := range points {
		if err := ValidateCoordinate(p.Lat, p.Lng); err != nil {
			return 0, 0, 0, err
		}
	}

	simplified := SimplifyPolyline(points)
	total := 0.0

	for i := 1; i < len(simplified); i++ {
		prev := simplified[i-1]
		curr := simplified[i]

		if prev.Lat == curr.Lat && prev.Lng == curr.Lng {
			continue
		}

		var dist float64
		if useEllipsoid {
			dist = VincentyApproxDistance(prev.Lat, prev.Lng, curr.Lat, curr.Lng)
		} else {
			dist = HaversineDistance(prev.Lat, prev.Lng, curr.Lat, curr.Lng)
		}

		total += dist
	}

	return total, len(points), len(simplified), nil
}

func ClosestPointOnSegment(p, a, b Coordinate, useEllipsoid bool) (Coordinate, float64) {
	if a.Lat == b.Lat && a.Lng == b.Lng {
		var dist float64
		if useEllipsoid {
			dist = VincentyApproxDistance(p.Lat, p.Lng, a.Lat, a.Lng)
		} else {
			dist = HaversineDistance(p.Lat, p.Lng, a.Lat, a.Lng)
		}
		return a, dist
	}

	phiP := toRad(p.Lat)
	lambdaP := toRad(p.Lng)
	phiA := toRad(a.Lat)
	lambdaA := toRad(a.Lng)
	phiB := toRad(b.Lat)
	lambdaB := toRad(b.Lng)

	deltaPhi := phiB - phiA
	deltaLambda := lambdaB - lambdaA

	t := ((phiP-phiA)*deltaPhi + (lambdaP-lambdaA)*deltaLambda) / (deltaPhi*deltaPhi + deltaLambda*deltaLambda)
	t = clamp(t, 0, 1)

	phiInterp := phiA + t*deltaPhi
	lambdaInterp := lambdaA + t*deltaLambda

	interpLat := toDeg(phiInterp)
	interpLng := toDeg(lambdaInterp)

	interpLng = normalizeLng(interpLng)

	interpLat = clamp(interpLat, -90, 90)

	closest := Coordinate{Lat: interpLat, Lng: interpLng}

	var dist float64
	if useEllipsoid {
		dist = VincentyApproxDistance(p.Lat, p.Lng, interpLat, interpLng)
	} else {
		dist = HaversineDistance(p.Lat, p.Lng, interpLat, interpLng)
	}

	return closest, dist
}

func normalizeLng(lng float64) float64 {
	for lng > 180 {
		lng -= 360
	}
	for lng < -180 {
		lng += 360
	}
	return lng
}

func PointToPolylineDistance(point Coordinate, polyline []Coordinate, useEllipsoid bool) (Coordinate, float64, int, error) {
	if err := ValidateCoordinate(point.Lat, point.Lng); err != nil {
		return Coordinate{}, 0, -1, err
	}

	for _, p := range polyline {
		if err := ValidateCoordinate(p.Lat, p.Lng); err != nil {
			return Coordinate{}, 0, -1, err
		}
	}

	if len(polyline) == 0 {
		return Coordinate{}, 0, -1, nil
	}

	if len(polyline) == 1 {
		var dist float64
		if useEllipsoid {
			dist = VincentyApproxDistance(point.Lat, point.Lng, polyline[0].Lat, polyline[0].Lng)
		} else {
			dist = HaversineDistance(point.Lat, point.Lng, polyline[0].Lat, polyline[0].Lng)
		}
		return polyline[0], dist, 0, nil
	}

	bestDist := math.Inf(1)
	var bestPoint Coordinate
	bestSegIdx := -1

	for i := 1; i < len(polyline); i++ {
		a := polyline[i-1]
		b := polyline[i]

		closest, dist := ClosestPointOnSegment(point, a, b, useEllipsoid)

		if dist < bestDist {
			bestDist = dist
			bestPoint = closest
			bestSegIdx = i - 1
		}
	}

	return bestPoint, bestDist, bestSegIdx, nil
}

func BatchDistances(center Coordinate, targets []Coordinate, useEllipsoid bool) ([]PointResult, error) {
	if err := ValidateCoordinate(center.Lat, center.Lng); err != nil {
		return nil, err
	}

	results := make([]PointResult, len(targets))

	for i, t := range targets {
		if err := ValidateCoordinate(t.Lat, t.Lng); err != nil {
			return nil, err
		}

		var dist float64
		if useEllipsoid {
			dist = VincentyApproxDistance(center.Lat, center.Lng, t.Lat, t.Lng)
		} else {
			dist = HaversineDistance(center.Lat, center.Lng, t.Lat, t.Lng)
		}

		results[i] = PointResult{
			Point:      t,
			DistanceKm: dist,
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceKm < results[j].DistanceKm
	})

	return results, nil
}
