package main

import (
	"strings"

	"goroutinelab/api"
)

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
