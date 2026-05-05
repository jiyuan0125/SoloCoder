package aggregator

import (
	"sort"
	"sync"

	"slowquery/protocol"
)

type templateStats struct {
	count         int64
	totalExecTime float64
	maxExecTime   float64
	totalScanRows int64
}

type Aggregator struct {
	mu            sync.RWMutex
	stats         map[string]*templateStats
	totalParsed   int64
	rowsSkipped   int64
	minExecTimeMs float64
}

func NewAggregator(minExecTimeMs float64) *Aggregator {
	return &Aggregator{
		stats:         make(map[string]*templateStats),
		minExecTimeMs: minExecTimeMs,
	}
}

func (a *Aggregator) Add(entry *protocol.LogEntry, sqlTemplate string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.totalParsed++

	if a.minExecTimeMs > 0 && entry.ExecTimeMs < a.minExecTimeMs {
		return
	}

	stat, exists := a.stats[sqlTemplate]
	if !exists {
		stat = &templateStats{}
		a.stats[sqlTemplate] = stat
	}

	stat.count++
	stat.totalExecTime += entry.ExecTimeMs
	if entry.ExecTimeMs > stat.maxExecTime {
		stat.maxExecTime = entry.ExecTimeMs
	}
	stat.totalScanRows += entry.ScanRows
}

func (a *Aggregator) AddSkipped() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rowsSkipped++
}

func (a *Aggregator) GetSortedStats(sortBy protocol.SortType, topN int) *protocol.Statistics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	templates := make([]string, 0, len(a.stats))
	for tmpl := range a.stats {
		templates = append(templates, tmpl)
	}

	sort.Slice(templates, func(i, j int) bool {
		statsI := a.stats[templates[i]]
		statsJ := a.stats[templates[j]]

		switch sortBy {
		case protocol.SortByTotalExecTime:
			return statsI.totalExecTime > statsJ.totalExecTime
		case protocol.SortByAvgExecTime:
			fallthrough
		default:
			avgI := statsI.totalExecTime / float64(statsI.count)
			avgJ := statsJ.totalExecTime / float64(statsJ.count)
			return avgI > avgJ
		}
	})

	result := &protocol.Statistics{
		TotalRowsParsed: a.totalParsed,
		RowsSkipped:     a.rowsSkipped,
		UniqueTemplates: int64(len(templates)),
		TemplateStats:   make([]protocol.TemplateStat, 0, len(templates)),
	}

	limit := len(templates)
	if topN > 0 && topN < limit {
		limit = topN
	}

	for rank := 0; rank < limit; rank++ {
		tmpl := templates[rank]
		stat := a.stats[tmpl]

		result.TemplateStats = append(result.TemplateStats, protocol.TemplateStat{
			Rank:            rank + 1,
			SQLTemplate:     tmpl,
			Count:           stat.count,
			AvgExecTimeMs:   stat.totalExecTime / float64(stat.count),
			MaxExecTimeMs:   stat.maxExecTime,
			TotalExecTimeMs: stat.totalExecTime,
			AvgScanRows:     float64(stat.totalScanRows) / float64(stat.count),
		})
	}

	return result
}

func (a *Aggregator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats = make(map[string]*templateStats)
	a.totalParsed = 0
	a.rowsSkipped = 0
}
