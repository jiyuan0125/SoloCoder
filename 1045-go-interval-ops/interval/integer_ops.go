package interval

import "sort"

func sortIntegerIntervals(intervals []*IntegerInterval) {
	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].Min == intervals[j].Min {
			return intervals[i].Max < intervals[j].Max
		}
		return intervals[i].Min < intervals[j].Min
	})
}

func MergeIntegers(intervals []*IntegerInterval) []*IntegerInterval {
	if len(intervals) == 0 {
		return []*IntegerInterval{}
	}
	if len(intervals) == 1 {
		return []*IntegerInterval{intervals[0]}
	}

	sorted := make([]*IntegerInterval, len(intervals))
	copy(sorted, intervals)
	sortIntegerIntervals(sorted)

	result := []*IntegerInterval{NewIntegerInterval(sorted[0].Min, sorted[0].Max)}
	for i := 1; i < len(sorted); i++ {
		last := result[len(result)-1]
		current := sorted[i]
		if last.OverlapsWith(current) || last.IsAdjacentTo(current) {
			newMin := min(last.Min, current.Min)
			newMax := max(last.Max, current.Max)
			result[len(result)-1] = NewIntegerInterval(newMin, newMax)
		} else {
			result = append(result, NewIntegerInterval(current.Min, current.Max))
		}
	}
	return result
}

func IntersectionIntegers(a, b []*IntegerInterval) []*IntegerInterval {
	if len(a) == 0 || len(b) == 0 {
		return []*IntegerInterval{}
	}

	aMerged := MergeIntegers(a)
	bMerged := MergeIntegers(b)

	result := []*IntegerInterval{}
	i, j := 0, 0
	for i < len(aMerged) && j < len(bMerged) {
		ia := aMerged[i]
		ib := bMerged[j]
		overlapStart := max(ia.Min, ib.Min)
		overlapEnd := min(ia.Max, ib.Max)
		if overlapStart <= overlapEnd {
			result = append(result, NewIntegerInterval(overlapStart, overlapEnd))
		}
		if ia.Max < ib.Max {
			i++
		} else {
			j++
		}
	}
	return result
}

func DifferenceIntegers(original, remove []*IntegerInterval) []*IntegerInterval {
	if len(original) == 0 {
		return []*IntegerInterval{}
	}
	if len(remove) == 0 {
		result := make([]*IntegerInterval, len(original))
		for i, iv := range original {
			result[i] = NewIntegerInterval(iv.Min, iv.Max)
		}
		return result
	}

	mergedOrig := MergeIntegers(original)
	mergedRemove := MergeIntegers(remove)

	result := []*IntegerInterval{}
	for _, iv := range mergedOrig {
		currentStart := iv.Min
		for _, rem := range mergedRemove {
			if rem.Max < currentStart {
				continue
			}
			if rem.Min > iv.Max {
				break
			}
			if rem.Min > currentStart {
				result = append(result, NewIntegerInterval(currentStart, rem.Min-1))
			}
			if rem.Max >= currentStart {
				currentStart = rem.Max + 1
			}
			if currentStart > iv.Max {
				break
			}
		}
		if currentStart <= iv.Max {
			result = append(result, NewIntegerInterval(currentStart, iv.Max))
		}
	}
	return result
}

func QueryPointIntegers(intervals []*IntegerInterval, point int) []*IntegerInterval {
	result := []*IntegerInterval{}
	for _, iv := range intervals {
		if iv.Contains(point) {
			result = append(result, NewIntegerInterval(iv.Min, iv.Max))
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
