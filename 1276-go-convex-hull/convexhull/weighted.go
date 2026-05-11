package convexhull

import (
	"math"
	"sort"
)

type WeightedPoint struct {
	Point
	Weight float64
}

func ComputeWeighted(points []WeightedPoint) *ConvexHull {
	if len(points) == 0 {
		return &ConvexHull{Type: HullEmpty}
	}
	maxW := points[0].Weight
	for _, p := range points {
		if p.Weight > maxW {
			maxW = p.Weight
		}
	}
	threshold := maxW * 0.5
	var selected []Point
	weighted := make([]WeightedPoint, len(points))
	copy(weighted, points)
	sort.SliceStable(weighted, func(i, j int) bool {
		if math.Abs(weighted[i].Weight-weighted[j].Weight) > epsilon {
			return weighted[i].Weight > weighted[j].Weight
		}
		return weighted[i].Point.Less(weighted[j].Point)
	})
	for _, p := range weighted {
		if p.Weight >= threshold {
			selected = append(selected, p.Point)
		}
	}
	if len(selected) == 0 {
		selected = append(selected, weighted[0].Point)
	}
	return Compute(selected)
}
