package analyzer

import (
	"sort"

	"goroutinelab/api"
)

type Analyzer struct {
	Detector *Detector
}

func New() *Analyzer {
	return &Analyzer{
		Detector: NewDetector(),
	}
}

func (a *Analyzer) Analyze(text string, opts *api.Options) *api.Report {
	if opts == nil {
		opts = &api.Options{}
	}
	if opts.WaitThresholdMinutes > 0 {
		a.Detector.WaitThresholdMinutes = opts.WaitThresholdMinutes
	}

	goroutines := Parse(text)

	byBlockType := make(map[string]int)
	for _, g := range goroutines {
		byBlockType[g.State]++
	}

	suspects := make([]*api.GoroutineInfo, 0)
	for _, g := range goroutines {
		if a.Detector.Detect(g) {
			suspects = append(suspects, g)
		}
	}

	groups := groupByBlockType(suspects)
	for _, g := range groups {
		sortByWaitMinutesDesc(g.Goroutines)
	}

	report := &api.Report{
		TotalGoroutines: len(goroutines),
		ByBlockType:     byBlockType,
		SuspectGroups:   groups,
		HasSuspicious:   len(suspects) > 0,
	}

	return report
}

func groupByBlockType(suspects []*api.GoroutineInfo) []*api.SuspectGroup {
	groupMap := make(map[string]*api.SuspectGroup)
	for _, g := range suspects {
		group, ok := groupMap[g.State]
		if !ok {
			group = &api.SuspectGroup{BlockType: g.State}
			groupMap[g.State] = group
		}
		group.Goroutines = append(group.Goroutines, g)
		group.Count++
	}

	result := make([]*api.SuspectGroup, 0, len(groupMap))
	for _, g := range groupMap {
		result = append(result, g)
	}
	return result
}

func sortByWaitMinutesDesc(goroutines []*api.GoroutineInfo) {
	sort.Slice(goroutines, func(i, j int) bool {
		if goroutines[i].HasWaitTime && goroutines[j].HasWaitTime {
			return goroutines[i].WaitMinutes > goroutines[j].WaitMinutes
		}
		return goroutines[i].HasWaitTime
	})
}
