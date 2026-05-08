package limiter

import (
	"time"
)

type gridLimiter struct {
	baseLimiter
	gridSize   time.Duration
	gridCount  int
	grids      map[int64]int
	lastAccess time.Time
}

func newGridLimiter(rule LimitRule) *gridLimiter {
	gridSize := rule.GridSize
	if gridSize <= 0 {
		gridSize = time.Second
	}

	gridCount := int(rule.Window / gridSize)
	if gridCount < 1 {
		gridCount = 1
	}

	return &gridLimiter{
		baseLimiter: baseLimiter{rule: rule},
		gridSize:    gridSize,
		gridCount:   gridCount,
		grids:       make(map[int64]int),
	}
}

func (g *gridLimiter) gridIndex(now time.Time) int64 {
	return now.UnixNano() / g.gridSize.Nanoseconds()
}

func (g *gridLimiter) cleanupExpired(now time.Time, currentIdx int64) {
	oldestIdx := currentIdx - int64(g.gridCount) + 1
	for idx := range g.grids {
		if idx < oldestIdx {
			delete(g.grids, idx)
		}
	}
}

func (g *gridLimiter) sumCounts(currentIdx int64) int {
	total := 0
	oldestIdx := currentIdx - int64(g.gridCount) + 1
	for idx, count := range g.grids {
		if idx >= oldestIdx && idx <= currentIdx {
			total += count
		}
	}
	return total
}

func (g *gridLimiter) findOldestActiveGrid(currentIdx int64) (int64, bool) {
	var oldestIdx int64
	found := false
	oldestThreshold := currentIdx - int64(g.gridCount) + 1

	for idx := range g.grids {
		if idx >= oldestThreshold && idx <= currentIdx {
			if !found || idx < oldestIdx {
				oldestIdx = idx
				found = true
			}
		}
	}
	return oldestIdx, found
}

func (g *gridLimiter) calculateRetryAfter(now time.Time, currentIdx int64) time.Duration {
	oldestIdx, found := g.findOldestActiveGrid(currentIdx)
	if !found {
		return 0
	}

	oldestGridEnd := time.Unix(0, (oldestIdx+1)*g.gridSize.Nanoseconds())
	waitTime := oldestGridEnd.Sub(now)
	if waitTime < 0 {
		waitTime = 0
	}
	return waitTime
}

func (g *gridLimiter) Allow() Result {
	return g.AllowN(1)
}

func (g *gridLimiter) AllowN(n int) Result {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	currentIdx := g.gridIndex(now)

	g.cleanupExpired(now, currentIdx)

	currentCount := g.sumCounts(currentIdx)

	if currentCount+n > g.rule.Limit {
		retryAfter := g.calculateRetryAfter(now, currentIdx)
		return Result{
			Allowed:    false,
			Remaining:  g.rule.Limit - currentCount,
			Limit:      g.rule.Limit,
			RetryAfter: retryAfter,
			WindowEnd:  time.Unix(0, (currentIdx+int64(g.gridCount))*g.gridSize.Nanoseconds()),
		}
	}

	g.grids[currentIdx] += n
	g.lastAccess = now

	newCount := g.sumCounts(currentIdx)
	return Result{
		Allowed:    true,
		Remaining:  g.rule.Limit - newCount,
		Limit:      g.rule.Limit,
		RetryAfter: 0,
		WindowEnd:  time.Unix(0, (currentIdx+int64(g.gridCount))*g.gridSize.Nanoseconds()),
	}
}

func (g *gridLimiter) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.grids = make(map[int64]int)
	g.lastAccess = time.Time{}
}

func (g *gridLimiter) Stats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	now := time.Now()
	currentIdx := g.gridIndex(now)
	total := g.sumCounts(currentIdx)

	return map[string]interface{}{
		"mode":       "grid",
		"limit":      g.rule.Limit,
		"window":     g.rule.Window.String(),
		"grid_size":  g.gridSize.String(),
		"grid_count": g.gridCount,
		"current":    total,
		"remaining":  g.rule.Limit - total,
		"last_access": g.lastAccess,
	}
}
