package peucker

import (
	"github.com/solo-coder/douglas-peucker/internal/api"
)

type stackItem struct {
	start int
	end   int
}

func DouglasPeuckerIterative(points []api.Point, threshold float64) []api.Point {
	n := len(points)
	if n <= 2 {
		return append([]api.Point{}, points...)
	}

	if threshold <= 0 {
		result := make([]api.Point, n)
		copy(result, points)
		return result
	}

	keep := make([]bool, n)
	keep[0] = true
	keep[n-1] = true

	stack := []stackItem{{start: 0, end: n - 1}}

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		start := item.start
		end := item.end

		if end-start <= 1 {
			continue
		}

		maxDist := 0.0
		maxIndex := start

		for i := start + 1; i < end; i++ {
			dist := PointToSegmentDistance(points[i], points[start], points[end])
			if dist > maxDist {
				maxDist = dist
				maxIndex = i
			}
		}

		if maxDist > threshold {
			keep[maxIndex] = true
			stack = append(stack, stackItem{start: maxIndex, end: end})
			stack = append(stack, stackItem{start: start, end: maxIndex})
		}
	}

	var result []api.Point
	for i, shouldKeep := range keep {
		if shouldKeep {
			result = append(result, points[i])
		}
	}

	return result
}

func ProcessClosedLoop(points []api.Point, threshold float64) []api.Point {
	n := len(points)
	if n <= 3 {
		return append([]api.Point{}, points...)
	}

	keep := make([]bool, n)
	for i := range keep {
		keep[i] = false
	}

	stack := []stackItem{{start: 0, end: n - 1}}

	for len(stack) > 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		start := item.start
		end := item.end

		if end-start <= 1 {
			continue
		}

		maxDist := 0.0
		maxIndex := start

		for i := start + 1; i < end; i++ {
			dist := PointToSegmentDistance(points[i], points[start], points[end])
			if dist > maxDist {
				maxDist = dist
				maxIndex = i
			}
		}

		if maxDist > threshold {
			keep[maxIndex] = true
			stack = append(stack, stackItem{start: maxIndex, end: end})
			stack = append(stack, stackItem{start: start, end: maxIndex})
		}
	}

	keep[0] = true
	keep[n-1] = true

	var result []api.Point
	for i, shouldKeep := range keep {
		if shouldKeep {
			result = append(result, points[i])
		}
	}

	return result
}
