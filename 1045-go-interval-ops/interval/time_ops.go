package interval

import (
	"sort"
	"time"
)

func sortTimeIntervals(intervals []*TimeInterval) {
	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].Start.Equal(intervals[j].Start) {
			return intervals[i].End.Before(intervals[j].End)
		}
		return intervals[i].Start.Before(intervals[j].Start)
	})
}

func MergeTimes(intervals []*TimeInterval) []*TimeInterval {
	if len(intervals) == 0 {
		return []*TimeInterval{}
	}
	if len(intervals) == 1 {
		zone := time.UTC
		if intervals[0].Zone != nil {
			zone = intervals[0].Zone
		}
		return []*TimeInterval{NewTimeInterval(intervals[0].Start, intervals[0].End, zone)}
	}

	sorted := make([]*TimeInterval, len(intervals))
	copy(sorted, intervals)
	sortTimeIntervals(sorted)

	zone := time.UTC
	if sorted[0].Zone != nil {
		zone = sorted[0].Zone
	}

	result := []*TimeInterval{NewTimeInterval(sorted[0].Start, sorted[0].End, zone)}
	for i := 1; i < len(sorted); i++ {
		last := result[len(result)-1]
		current := sorted[i]
		if last.OverlapsWith(current) || last.IsAdjacentTo(current) {
			newStart := minTime(last.Start, current.Start)
			newEnd := maxTime(last.End, current.End)
			result[len(result)-1] = NewTimeInterval(newStart, newEnd, zone)
		} else {
			result = append(result, NewTimeInterval(current.Start, current.End, zone))
		}
	}
	return result
}

func IntersectionTimes(a, b []*TimeInterval) []*TimeInterval {
	if len(a) == 0 || len(b) == 0 {
		return []*TimeInterval{}
	}

	aMerged := MergeTimes(a)
	bMerged := MergeTimes(b)

	zone := time.UTC
	if aMerged[0].Zone != nil {
		zone = aMerged[0].Zone
	}

	result := []*TimeInterval{}
	i, j := 0, 0
	for i < len(aMerged) && j < len(bMerged) {
		ia := aMerged[i]
		ib := bMerged[j]
		overlapStart := maxTime(ia.Start, ib.Start)
		overlapEnd := minTime(ia.End, ib.End)
		if overlapStart.Before(overlapEnd) || overlapStart.Equal(overlapEnd) {
			result = append(result, NewTimeInterval(overlapStart, overlapEnd, zone))
		}
		if ia.End.Before(ib.End) {
			i++
		} else {
			j++
		}
	}
	return result
}

func DifferenceTimes(original, remove []*TimeInterval) []*TimeInterval {
	if len(original) == 0 {
		return []*TimeInterval{}
	}
	if len(remove) == 0 {
		zone := time.UTC
		if len(original) > 0 && original[0].Zone != nil {
			zone = original[0].Zone
		}
		result := make([]*TimeInterval, len(original))
		for i, iv := range original {
			result[i] = NewTimeInterval(iv.Start, iv.End, zone)
		}
		return result
	}

	mergedOrig := MergeTimes(original)
	mergedRemove := MergeTimes(remove)

	zone := time.UTC
	if len(mergedOrig) > 0 && mergedOrig[0].Zone != nil {
		zone = mergedOrig[0].Zone
	}

	result := []*TimeInterval{}
	for _, iv := range mergedOrig {
		currentStart := iv.Start
		for _, rem := range mergedRemove {
			if rem.End.Before(currentStart) {
				continue
			}
			if rem.Start.After(iv.End) {
				break
			}
			if rem.Start.After(currentStart) {
				end := rem.Start.Add(-time.Second)
				if !currentStart.After(end) {
					result = append(result, NewTimeInterval(currentStart, end, zone))
				}
			}
			if !rem.End.Before(currentStart) {
				currentStart = rem.End.Add(time.Second)
			}
			if currentStart.After(iv.End) {
				break
			}
		}
		if !currentStart.After(iv.End) {
			result = append(result, NewTimeInterval(currentStart, iv.End, zone))
		}
	}
	return result
}

func QueryPointTimes(intervals []*TimeInterval, point time.Time) []*TimeInterval {
	result := []*TimeInterval{}
	for _, iv := range intervals {
		if iv.Contains(point) {
			zone := time.UTC
			if iv.Zone != nil {
				zone = iv.Zone
			}
			result = append(result, NewTimeInterval(iv.Start, iv.End, zone))
		}
	}
	return result
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
