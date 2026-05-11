package main

import (
	"sort"

	"goroutinelab/analyzer"
	"goroutinelab/api"
)

type localAnalyzer struct {
	detector *detector
}

type detector struct {
	waitThresholdMinutes int
}

func newLocalAnalyzer(opts *api.Options) *localAnalyzer {
	threshold := 5
	if opts != nil && opts.WaitThresholdMinutes > 0 {
		threshold = opts.WaitThresholdMinutes
	}
	return &localAnalyzer{
		detector: &detector{waitThresholdMinutes: threshold},
	}
}

func (a *localAnalyzer) Analyze(text string, opts *api.Options) *api.Report {
	goroutines := analyzer.Parse(text)

	byBlockType := make(map[string]int)
	for _, g := range goroutines {
		byBlockType[g.State]++
	}

	suspects := make([]*api.GoroutineInfo, 0)
	for _, g := range goroutines {
		if a.detector.detect(g) {
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

func (d *detector) detect(g *api.GoroutineInfo) bool {
	if isKnownNormal(g) {
		return false
	}

	if g.StackTruncated {
		return false
	}

	var reasons []string

	switch g.State {
	case api.BlockTypeChanReceive:
		if d.isLongWait(g) {
			reasons = append(reasons, "channel receive blocked for extended period")
		}
	case api.BlockTypeChanSend:
		if d.isLongWait(g) {
			reasons = append(reasons, "channel send blocked for extended period")
		}
	case api.BlockTypeSelect:
		if d.isLongWait(g) && !hasProgress(g) {
			reasons = append(reasons, "select blocked without progress for extended period")
		}
	case api.BlockTypeIOWait:
		if d.isLongWait(g) && !hasTimeout(g) {
			reasons = append(reasons, "I/O wait with no timeout for extended period")
		}
	case api.BlockTypeSleep:
		if d.isLongWait(g) {
			reasons = append(reasons, "sleeping for unexpectedly long time")
		}
	case api.BlockTypeSemacquire:
		if d.isLongWait(g) {
			reasons = append(reasons, "semaphore/lock contention for extended period")
		}
	}

	if len(reasons) > 0 {
		g.Reasons = reasons
		return true
	}
	return false
}

func (d *detector) isLongWait(g *api.GoroutineInfo) bool {
	if !g.HasWaitTime {
		return false
	}
	return g.WaitMinutes >= d.waitThresholdMinutes
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
