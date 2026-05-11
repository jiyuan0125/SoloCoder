package analyzer

import (
	"strings"

	"goroutinelab/api"
)

const DefaultWaitThresholdMinutes = 5

type Detector struct {
	WaitThresholdMinutes int
}

func NewDetector() *Detector {
	return &Detector{WaitThresholdMinutes: DefaultWaitThresholdMinutes}
}

func (d *Detector) Detect(g *api.GoroutineInfo) bool {
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

func (d *Detector) isLongWait(g *api.GoroutineInfo) bool {
	if !g.HasWaitTime {
		return false
	}
	return g.WaitMinutes >= d.WaitThresholdMinutes
}

func isKnownNormal(g *api.GoroutineInfo) bool {
	if g.ID == 1 {
		return true
	}
	for _, f := range g.UserStack {
		if isGCStack(f.Function) {
			return true
		}
	}
	return false
}

func isGCStack(fn string) bool {
	gcMarkers := []string{
		"runtime.gcBgMarkWorker",
		"runtime.gcBgMarkStartWorkers",
		"runtime.runfinq",
		"runtime.sysmon",
		"runtime.checkdead",
		"runtime.notetsleepg",
		"runtime.scavenge",
		"runtime.forcegchelper",
		"runtime.gchelper",
	}
	for _, m := range gcMarkers {
		if strings.Contains(fn, m) {
			return true
		}
	}
	return false
}

func hasProgress(g *api.GoroutineInfo) bool {
	return false
}

func hasTimeout(g *api.GoroutineInfo) bool {
	for _, f := range g.UserStack {
		fn := f.Function
		if strings.Contains(fn, "context.WithTimeout") ||
			strings.Contains(fn, "context.WithDeadline") ||
			strings.Contains(fn, "SetWriteDeadline") ||
			strings.Contains(fn, "SetReadDeadline") ||
			strings.Contains(fn, "SetDeadline") {
			return true
		}
	}
	return false
}
