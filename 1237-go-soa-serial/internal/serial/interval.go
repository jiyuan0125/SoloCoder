package serial

import "sort"

type interval struct {
	start int64
	end   int64
}

type intervalList []interval

func newIntervalList() intervalList {
	return make(intervalList, 0)
}

func (il *intervalList) add(s, e int64) {
	newIntervals := append(*il, interval{start: s, end: e})
	sort.Slice(newIntervals, func(i, j int) bool {
		return newIntervals[i].start < newIntervals[j].start
	})
	merged := make(intervalList, 0, len(newIntervals))
	for _, iv := range newIntervals {
		n := len(merged)
		if n == 0 {
			merged = append(merged, iv)
			continue
		}
		last := &merged[n-1]
		if iv.start <= last.end+1 {
			if iv.end > last.end {
				last.end = iv.end
			}
		} else {
			merged = append(merged, iv)
		}
	}
	*il = merged
}

func (il *intervalList) findGaps(max int64) []interval {
	if max < 1 {
		return nil
	}
	var gaps []interval
	prev := int64(0)
	for _, iv := range *il {
		if iv.start > prev+1 {
			gaps = append(gaps, interval{start: prev + 1, end: iv.start - 1})
		}
		if iv.end > prev {
			prev = iv.end
		}
	}
	if prev < max {
		gaps = append(gaps, interval{start: prev + 1, end: max})
	}
	return gaps
}

func (il *intervalList) count() int64 {
	var total int64
	for _, iv := range *il {
		total += iv.end - iv.start + 1
	}
	return total
}

func (il *intervalList) max() int64 {
	if len(*il) == 0 {
		return 0
	}
	return (*il)[len(*il)-1].end
}
