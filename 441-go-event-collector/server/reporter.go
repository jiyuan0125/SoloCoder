package server

import (
	"sync"
	"time"

	"event-collector/common"
)

type DailyReporter struct {
	storage      *MemoryStorage
	reports      map[string]*common.DailyReport
	mu           sync.RWMutex
	lastReport   string
}

func NewDailyReporter(storage *MemoryStorage) *DailyReporter {
	reporter := &DailyReporter{
		storage: storage,
		reports: make(map[string]*common.DailyReport),
	}
	go reporter.startDailyReport()
	return reporter
}

func (r *DailyReporter) startDailyReport() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.generateDailyReport()
	}
}

func (r *DailyReporter) generateDailyReport() {
	today := time.Now().Format("2006-01-02")

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lastReport == today {
		return
	}

	events := r.storage.GetTodayEvents()
	eventSummary := make(map[string]common.EventDaily)

	for _, event := range events {
		daily, exists := eventSummary[event.Name]
		if !exists {
			daily = common.EventDaily{
				TotalCount:  0,
				UniqueUsers: make([]string, 0),
			}
		}

		daily.TotalCount++

		userExists := false
		for _, userID := range daily.UniqueUsers {
			if userID == event.UserID {
				userExists = true
				break
			}
		}
		if !userExists {
			daily.UniqueUsers = append(daily.UniqueUsers, event.UserID)
		}

		eventSummary[event.Name] = daily
	}

	report := &common.DailyReport{
		Date:         today,
		EventSummary: eventSummary,
	}

	r.reports[today] = report
	r.lastReport = today

	r.cleanupOldReports()
}

func (r *DailyReporter) cleanupOldReports() {
	cutoffDate := time.Now().AddDate(0, 0, -common.ReportRetentionDays).Format("2006-01-02")

	for date := range r.reports {
		if date < cutoffDate {
			delete(r.reports, date)
		}
	}
}

func (r *DailyReporter) GetReports() []*common.DailyReport {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reports := make([]*common.DailyReport, 0, len(r.reports))
	for _, report := range r.reports {
		reports = append(reports, report)
	}

	return reports
}

func (r *DailyReporter) GetReport(date string) (*common.DailyReport, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, exists := r.reports[date]
	return report, exists
}
