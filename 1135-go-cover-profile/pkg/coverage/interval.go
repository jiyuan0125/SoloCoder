package coverage

import (
	"sort"
)

type LineInterval struct {
	Start int
	End   int
	Count int
}

func MergeIntervalsByFile(blocks []CoverBlock) map[string][]LineInterval {
	fileIntervals := make(map[string][]LineInterval)

	for _, block := range blocks {
		interval := LineInterval{
			Start: block.StartLine,
			End:   block.EndLine,
			Count: block.Count,
		}
		fileIntervals[block.File] = append(fileIntervals[block.File], interval)
	}

	for file, intervals := range fileIntervals {
		fileIntervals[file] = mergeIntervals(intervals)
	}

	return fileIntervals
}

func mergeIntervals(intervals []LineInterval) []LineInterval {
	if len(intervals) == 0 {
		return intervals
	}

	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].Start != intervals[j].Start {
			return intervals[i].Start < intervals[j].Start
		}
		return intervals[i].End < intervals[j].End
	})

	merged := []LineInterval{intervals[0]}

	for i := 1; i < len(intervals); i++ {
		last := &merged[len(merged)-1]
		current := intervals[i]

		if current.Start <= last.End {
			newCount := last.Count
			if current.Count > last.Count {
				newCount = current.Count
			}

			if current.End > last.End {
				last.End = current.End
			}
			last.Count = newCount
		} else {
			merged = append(merged, current)
		}
	}

	return merged
}

func GetCoveredLineNumbers(intervals []LineInterval) map[int]int {
	covered := make(map[int]int)

	for _, interval := range intervals {
		for line := interval.Start; line <= interval.End; line++ {
			if current, exists := covered[line]; !exists || interval.Count > current {
				covered[line] = interval.Count
			}
		}
	}

	return covered
}

func CountTotalLines(intervals []LineInterval) int {
	if len(intervals) == 0 {
		return 0
	}

	sorted := make([]LineInterval, len(intervals))
	copy(sorted, intervals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	total := 0
	current := sorted[0]

	for i := 1; i < len(sorted); i++ {
		if sorted[i].Start <= current.End {
			if sorted[i].End > current.End {
				current.End = sorted[i].End
			}
		} else {
			total += current.End - current.Start + 1
			current = sorted[i]
		}
	}

	total += current.End - current.Start + 1
	return total
}
