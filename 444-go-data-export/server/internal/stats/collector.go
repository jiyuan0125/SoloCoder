package stats

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"data-export/pkg/common"
)

type dailyStatsKey struct {
	year  int
	month int
	day   int
}

type dailyStatsData struct {
	totalTasks   int
	successTasks int
	failedTasks  int
	totalDuration int64
}

type templateStatsData struct {
	usageCount   int
	totalDuration int64
}

type StatsCollector struct {
	dailyMap    map[dailyStatsKey]*dailyStatsData
	templateMap map[string]*templateStatsData
	templateNames map[string]string
	mu          sync.RWMutex
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		dailyMap:      make(map[dailyStatsKey]*dailyStatsData),
		templateMap:   make(map[string]*templateStatsData),
		templateNames: make(map[string]string),
	}
}

func (s *StatsCollector) RecordTask(task *common.ExportTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := task.CreatedAt
	key := dailyStatsKey{
		year:  t.Year(),
		month: int(t.Month()),
		day:   t.Day(),
	}

	data := s.dailyMap[key]
	if data == nil {
		data = &dailyStatsData{}
		s.dailyMap[key] = data
	}

	data.totalTasks++
	if task.Status == common.TaskStatusCompleted {
		data.successTasks++
	} else if task.Status == common.TaskStatusFailed {
		data.failedTasks++
	}
	if task.DurationMs > 0 {
		data.totalDuration += task.DurationMs
	}

	if task.TemplateID != "" {
		tplData := s.templateMap[task.TemplateID]
		if tplData == nil {
			tplData = &templateStatsData{}
			s.templateMap[task.TemplateID] = tplData
		}
		tplData.usageCount++
		if task.DurationMs > 0 {
			tplData.totalDuration += task.DurationMs
		}
		if task.TemplateName != "" {
			s.templateNames[task.TemplateID] = task.TemplateName
		}
	}
}

func (s *StatsCollector) GetDailyStats(days int) []common.DailyStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []common.DailyStats
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		key := dailyStatsKey{
			year:  date.Year(),
			month: int(date.Month()),
			day:   date.Day(),
		}

		data := s.dailyMap[key]
		stat := common.DailyStats{
			Date:         date.Format("2006-01-02"),
			TotalTasks:   0,
			SuccessTasks: 0,
			FailedTasks:  0,
			AvgDurationMs: 0,
		}

		if data != nil {
			stat.TotalTasks = data.totalTasks
			stat.SuccessTasks = data.successTasks
			stat.FailedTasks = data.failedTasks
			if data.totalTasks > 0 {
				stat.AvgDurationMs = data.totalDuration / int64(data.totalTasks)
			}
		}

		result = append(result, stat)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result
}

func (s *StatsCollector) GetTopTemplates(topN int) []common.TemplateStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []common.TemplateStats

	for tplID, data := range s.templateMap {
		stat := common.TemplateStats{
			TemplateID:   tplID,
			TemplateName: s.templateNames[tplID],
			UsageCount:   data.usageCount,
		}
		if data.usageCount > 0 {
			stat.AvgDurationMs = data.totalDuration / int64(data.usageCount)
		}
		list = append(list, stat)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].UsageCount > list[j].UsageCount
	})

	if len(list) > topN {
		list = list[:topN]
	}

	return list
}

func (s *StatsCollector) GetOverallStats() struct {
	TotalTasks    int
	SuccessRate   string
	AvgDurationMs int64
} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalTasks int
	var successTasks int
	var totalDuration int64

	for _, data := range s.dailyMap {
		totalTasks += data.totalTasks
		successTasks += data.successTasks
		totalDuration += data.totalDuration
	}

	var successRate string
	if totalTasks > 0 {
		rate := float64(successTasks) / float64(totalTasks) * 100
		successRate = fmt.Sprintf("%.1f%%", rate)
	} else {
		successRate = "0%"
	}

	var avgDuration int64
	if totalTasks > 0 {
		avgDuration = totalDuration / int64(totalTasks)
	}

	return struct {
		TotalTasks    int
		SuccessRate   string
		AvgDurationMs int64
	}{
		TotalTasks:    totalTasks,
		SuccessRate:   successRate,
		AvgDurationMs: avgDuration,
	}
}

func (s *StatsCollector) GetExportStats(days int, topN int) *common.ExportStatsResponse {
	dailyStats := s.GetDailyStats(days)
	topTemplates := s.GetTopTemplates(topN)
	overall := s.GetOverallStats()

	resp := &common.ExportStatsResponse{
		DailyStats:   dailyStats,
		TopTemplates: topTemplates,
	}

	resp.OverallStats.TotalTasks = overall.TotalTasks
	resp.OverallStats.SuccessRate = overall.SuccessRate
	resp.OverallStats.AvgDurationMs = overall.AvgDurationMs

	return resp
}
