package peucker

import (
	"math"

	"github.com/solo-coder/douglas-peucker/internal/api"
)

type SimplifyOptions struct {
	Mode           api.ThresholdMode
	Threshold      float64
	IsClosed       bool
	UniformOptions *api.UniformOptions
}

func Simplify(req *api.SimplifyRequest) *api.SimplifyResponse {
	originalCount := len(req.Points)
	if originalCount <= 2 {
		return createResponse(req.Points, req.Points)
	}

	threshold := req.Threshold
	if req.Mode == api.PercentageThresholdMode {
		totalLength := CalculateTotalLength(req.Points)
		threshold = totalLength * (req.Threshold / 100.0)
	}

	var simplifiedPoints []api.Point

	if req.IsClosed {
		simplifiedPoints = ProcessClosedLoop(req.Points, threshold)
	} else {
		simplifiedPoints = DouglasPeuckerIterative(req.Points, threshold)
	}

	if req.UniformOptions != nil {
		simplifiedPoints = applyUniformSampling(simplifiedPoints, req.UniformOptions)
	}

	if len(simplifiedPoints) < 2 {
		simplifiedPoints = []api.Point{req.Points[0], req.Points[len(req.Points)-1]}
	}

	return createResponse(req.Points, simplifiedPoints)
}

func createResponse(original, simplified []api.Point) *api.SimplifyResponse {
	originalCount := len(original)
	simplifiedCount := len(simplified)
	reductionRate := 0.0
	if originalCount > 0 {
		reductionRate = float64(originalCount-simplifiedCount) / float64(originalCount) * 100.0
	}

	originalArea := ShoelaceArea(original)
	simplifiedArea := ShoelaceArea(simplified)

	areaDeviation := 0.0
	if originalArea > 0 {
		areaDeviation = math.Abs(simplifiedArea-originalArea) / originalArea * 100.0
	}

	return &api.SimplifyResponse{
		Points:          simplified,
		OriginalCount:   originalCount,
		SimplifiedCount: simplifiedCount,
		ReductionRate:   reductionRate,
		AreaDeviation:   areaDeviation,
		OriginalArea:    originalArea,
		SimplifiedArea:  simplifiedArea,
		TotalLength:     CalculateTotalLength(original),
	}
}

func applyUniformSampling(points []api.Point, opts *api.UniformOptions) []api.Point {
	n := len(points)
	if n <= 2 {
		return append([]api.Point{}, points...)
	}

	if opts.MinDistance <= 0 && opts.MaxDistance <= 0 {
		return append([]api.Point{}, points...)
	}

	result := []api.Point{points[0]}

	for i := 1; i < n-1; i++ {
		distToLast := HaversineDistance(result[len(result)-1], points[i])

		shouldKeep := true

		if opts.MinDistance > 0 {
			if distToLast < opts.MinDistance {
				shouldKeep = false
			}
		}

		if shouldKeep {
			result = append(result, points[i])
		}
	}

	result = append(result, points[n-1])

	finalResult := []api.Point{result[0]}
	for i := 1; i < len(result)-1; i++ {
		dist := HaversineDistance(finalResult[len(finalResult)-1], result[i])
		if opts.MinDistance > 0 && dist < opts.MinDistance {
			continue
		}
		finalResult = append(finalResult, result[i])
	}
	finalResult = append(finalResult, result[len(result)-1])

	if opts.MaxDistance > 0 {
		finalResult = interpolateForMaxDistance(finalResult, opts.MaxDistance)
	}

	return finalResult
}

func interpolateForMaxDistance(points []api.Point, maxDist float64) []api.Point {
	if len(points) <= 2 || maxDist <= 0 {
		return append([]api.Point{}, points...)
	}

	var result []api.Point
	result = append(result, points[0])

	for i := 1; i < len(points); i++ {
		start := points[i-1]
		end := points[i]
		segmentDist := HaversineDistance(start, end)

		if segmentDist > maxDist {
			numPoints := int(math.Ceil(segmentDist/maxDist)) - 1
			for j := 1; j <= numPoints; j++ {
				t := float64(j) / float64(numPoints+1)
				interpPoint := api.Point{
					Lat: start.Lat + (end.Lat-start.Lat)*t,
					Lon: start.Lon + (end.Lon-start.Lon)*t,
				}
				result = append(result, interpPoint)
			}
		}

		result = append(result, end)
	}

	return result
}
