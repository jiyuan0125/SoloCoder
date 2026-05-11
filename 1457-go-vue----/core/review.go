package core

import (
	"complaint-system/common"
	"errors"
	"time"
)

func (s *Service) ReviewTicket(ticketNo string, req *common.ReviewRequest) error {
	if req.ReviewResult != common.ReviewSatisfied && req.ReviewResult != common.ReviewBasiclySatisfied && req.ReviewResult != common.ReviewDissatisfied {
		return errors.New("invalid review_result")
	}

	t, ok := s.store.get(ticketNo)
	if !ok {
		return ErrTicketNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != common.StatusPendingReview {
		return ErrInvalidStatus
	}

	now := time.Now()
	t.ReviewResult = &req.ReviewResult
	t.ReviewRemark = req.Remark
	t.LastReviewTime = &now

	if req.ReviewResult == common.ReviewDissatisfied {
		if t.RetryCount >= maxRetryCount {
			t.IsEscalated = true
			t.Status = common.StatusProcessing
		} else {
			t.RetryCount++
			t.Status = common.StatusProcessing
			t.ProcessingResult = ""
			t.ProcessedAt = nil
		}
	} else {
		t.Status = common.StatusClosed
		t.ClosedAt = &now
	}

	return nil
}

func (s *Service) GetStatistics() *common.StatisticsResponse {
	tickets := s.store.list()
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	var todayNewCount int
	var processingCount int
	var closedCount int
	var totalProcessingHours float64
	var validProcessingCount int
	var reviewTotal int
	var reviewSatisfiedCount int

	for _, t := range tickets {
		t.mu.RLock()
		if t.CreatedAt.After(startOfDay) || t.CreatedAt.Equal(startOfDay) {
			todayNewCount++
		}

		switch t.Status {
		case common.StatusDispatched, common.StatusProcessing:
			processingCount++
		case common.StatusClosed:
			closedCount++
			if !t.IsTimeout && t.ProcessedAt != nil {
				duration := t.ProcessedAt.Sub(t.CreatedAt)
				totalProcessingHours += duration.Hours()
				validProcessingCount++
			}
		}

		if t.LastReviewTime != nil {
			if t.LastReviewTime.After(thirtyDaysAgo) {
				reviewTotal++
				if *t.ReviewResult == common.ReviewSatisfied || *t.ReviewResult == common.ReviewBasiclySatisfied {
					reviewSatisfiedCount++
				}
			}
		}
		t.mu.RUnlock()
	}

	var avgProcessingHours float64
	if validProcessingCount > 0 {
		avgProcessingHours = totalProcessingHours / float64(validProcessingCount)
	}

	var reviewSatisfaction float64
	if reviewTotal > 0 {
		reviewSatisfaction = float64(reviewSatisfiedCount) / float64(reviewTotal) * 100
	}

	return &common.StatisticsResponse{
		TodayNewCount:      todayNewCount,
		ProcessingCount:    processingCount,
		ClosedCount:        closedCount,
		AvgProcessingHours: avgProcessingHours,
		ReviewSatisfaction:  reviewSatisfaction,
	}
}
