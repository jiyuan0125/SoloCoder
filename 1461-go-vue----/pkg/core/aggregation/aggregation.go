package aggregation

import (
	"time"

	"quality-trace/pkg/common"
	"quality-trace/pkg/core/batch"
	"quality-trace/pkg/core/inspection"
	"quality-trace/pkg/core/process"
)

type Service struct {
	batchSvc      *batch.Service
	processSvc    *process.Service
	inspectionSvc *inspection.Service
}

func NewService(b *batch.Service, p *process.Service, i *inspection.Service) *Service {
	return &Service{
		batchSvc:      b,
		processSvc:    p,
		inspectionSvc: i,
	}
}

func (s *Service) GetMetricsSummary(days int) (*common.MetricsSummaryResponse, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	batchMetrics, err := s.GetOneTimePassRate(startTime, endTime)
	if err != nil {
		return nil, err
	}

	processMetrics := s.GetProcessDefectRates(startTime, endTime)

	return &common.MetricsSummaryResponse{
		BatchMetrics:   *batchMetrics,
		ProcessMetrics: processMetrics,
	}, nil
}

func (s *Service) GetOneTimePassRate(startTime, endTime time.Time) (*common.BatchMetricsResponse, error) {
	filter := common.ListBatchesRequest{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	batches := s.batchSvc.ListBatches(filter)

	var eligibleBatches []*common.Batch
	for _, b := range batches {
		if (b.Status == common.BatchStatusCompleted || 
			b.Status == common.BatchStatusAbnormal || 
			b.Status == common.BatchStatusScrapped) && 
			s.inspectionSvc.HasFinalInspection(b.BatchID) {
			eligibleBatches = append(eligibleBatches, b)
		}
	}

	totalBatches := len(eligibleBatches)
	if totalBatches == 0 {
		return &common.BatchMetricsResponse{
			OneTimePassRate:      0,
			PeriodStart:          startTime.Format(time.RFC3339),
			PeriodEnd:            endTime.Format(time.RFC3339),
			TotalBatches:         0,
			OneTimePassBatches:   0,
		}, nil
	}

	oneTimePassCount := 0
	for _, b := range eligibleBatches {
		if b.IsOneTimePass {
			oneTimePassCount++
		}
	}

	rate := float64(oneTimePassCount) / float64(totalBatches) * 100

	return &common.BatchMetricsResponse{
		OneTimePassRate:      rate,
		PeriodStart:          startTime.Format(time.RFC3339),
		PeriodEnd:            endTime.Format(time.RFC3339),
		TotalBatches:         totalBatches,
		OneTimePassBatches:   oneTimePassCount,
	}, nil
}

func (s *Service) GetProcessDefectRates(startTime, endTime time.Time) []common.ProcessMetricsResponse {
	processRecords := s.processSvc.GetAllProcessRecords()

	type stats struct {
		total int
		fails int
	}

	processStats := make(map[string]*stats)

	for _, r := range processRecords {
		if r.StartTime.Before(startTime) || r.StartTime.After(endTime) {
			continue
		}

		if _, exists := processStats[r.ProcessName]; !exists {
			processStats[r.ProcessName] = &stats{}
		}
		processStats[r.ProcessName].total++
		if r.Result == common.ProcessResultFail {
			processStats[r.ProcessName].fails++
		}
	}

	var result []common.ProcessMetricsResponse
	for name, stat := range processStats {
		defectRate := 0.0
		if stat.total > 0 {
			defectRate = float64(stat.fails) / float64(stat.total) * 100
		}

		result = append(result, common.ProcessMetricsResponse{
			ProcessName: name,
			TotalCount:  stat.total,
			FailCount:   stat.fails,
			DefectRate:  defectRate,
		})
	}

	return result
}
