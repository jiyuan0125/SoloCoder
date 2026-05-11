package hilbert

import (
	"errors"
	"sort"
)

func RangeQuery(n int, minX, maxX, minY, maxY uint64) ([]Range, error) {
	if err := validateOrder(n); err != nil {
		return nil, err
	}

	size := uint64(1) << n
	if minX > maxX || minY > maxY || maxX >= size || maxY >= size {
		return nil, errors.New("invalid range coordinates")
	}

	if n == 0 {
		if minX == 0 && maxX == 0 && minY == 0 && maxY == 0 {
			return []Range{{0, 0}}, nil
		}
		return nil, errors.New("invalid range coordinates")
	}

	if minX == 0 && maxX == size-1 && minY == 0 && maxY == size-1 {
		maxD := (uint64(1) << (2 * n)) - 1
		return []Range{{0, maxD}}, nil
	}

	ranges := rangeQueryRecursive(n, minX, maxX, minY, maxY, 0, 0, 0)
	if len(ranges) == 0 {
		return nil, nil
	}

	ranges = mergeRanges(ranges)
	return ranges, nil
}

func rangeQueryRecursive(n int, queryMinX, queryMaxX, queryMinY, queryMaxY, currMinX, currMinY, currD uint64) []Range {
	size := uint64(1) << n

	currMaxX := currMinX + size - 1
	currMaxY := currMinY + size - 1

	if currMaxX < queryMinX || currMinX > queryMaxX || currMaxY < queryMinY || currMinY > queryMaxY {
		return nil
	}

	if queryMinX <= currMinX && queryMaxX >= currMaxX && queryMinY <= currMinY && queryMaxY >= currMaxY {
		maxD := currD + (uint64(1) << (2 * n)) - 1
		return []Range{{currD, maxD}}
	}

	if n == 1 {
		var ranges []Range
		subSize := size / 2
		order := [][]uint64{{0, 0}, {0, 1}, {1, 1}, {1, 0}}

		for i, offset := range order {
			x := currMinX + offset[0]*subSize
			y := currMinY + offset[1]*subSize

			if x >= queryMinX && x <= queryMaxX && y >= queryMinY && y <= queryMaxY {
				ranges = append(ranges, Range{currD + uint64(i), currD + uint64(i)})
			}
		}
		return ranges
	}

	subSize := size / 2
	subDSize := subSize * subSize

	ranges := make([]Range, 0)

	order := [][]uint64{{0, 0}, {0, 1}, {1, 1}, {1, 0}}
	offsets := []uint64{0, subDSize, 2 * subDSize, 3 * subDSize}

	for i, offset := range order {
		x := currMinX + offset[0]*subSize
		y := currMinY + offset[1]*subSize
		d := currD + offsets[i]

		subRanges := rangeQueryRecursive(n-1, queryMinX, queryMaxX, queryMinY, queryMaxY, x, y, d)
		ranges = append(ranges, subRanges...)
	}

	return ranges
}

func mergeRanges(ranges []Range) []Range {
	if len(ranges) <= 1 {
		return ranges
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Min < ranges[j].Min
	})

	merged := []Range{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if r.Min <= last.Max+1 {
			if r.Max > last.Max {
				last.Max = r.Max
			}
		} else {
			merged = append(merged, r)
		}
	}

	return merged
}

func EstimateDistance(n int, d1, d2 uint64) (uint64, error) {
	if err := validateOrder(n); err != nil {
		return 0, err
	}

	if err := validateIndex(n, d1); err != nil {
		return 0, err
	}

	if err := validateIndex(n, d2); err != nil {
		return 0, err
	}

	if d1 == d2 {
		return 0, nil
	}

	var diff uint64
	if d1 > d2 {
		diff = d1 - d2
	} else {
		diff = d2 - d1
	}

	approxDistance := uint64(float64(diff) / 1.5)
	return approxDistance, nil
}

func BatchPointToIndex(n int, points []Point, ascending bool) ([]uint64, error) {
	if err := validateOrder(n); err != nil {
		return nil, err
	}

	if len(points) == 0 {
		return []uint64{}, nil
	}

	indices := make([]uint64, len(points))
	for i, p := range points {
		d, err := PointToIndex(n, p)
		if err != nil {
			return nil, err
		}
		indices[i] = d
	}

	if ascending {
		sort.Slice(indices, func(i, j int) bool {
			return indices[i] < indices[j]
		})
	} else {
		sort.Slice(indices, func(i, j int) bool {
			return indices[i] > indices[j]
		})
	}

	return indices, nil
}
