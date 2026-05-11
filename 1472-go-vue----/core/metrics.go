package core

import (
	"time"
	"taxisystem/common"
)

type MetricsManager struct {
	dm *DispatchManager
}

func NewMetricsManager(dm *DispatchManager) *MetricsManager {
	return &MetricsManager{dm: dm}
}

func (mm *MetricsManager) GetMetrics() *common.Metrics {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayStartUnix := todayStart.Unix()

	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	sevenDaysAgoUnix := sevenDaysAgo.Unix()

	orders := mm.dm.GetOrdersForMetrics()

	todayOrderCount := 0
	var totalAcceptTime float64
	acceptOrderCount := 0
	var totalFare float64
	completedWithFareCount := 0
	goodReviews := 0
	totalReviews := 0

	for _, order := range orders {
		if order.CreateTime >= todayStartUnix {
			todayOrderCount++
		}

		if order.Status == common.OrderStatusAccepted ||
			order.Status == common.OrderStatusInProgress ||
			order.Status == common.OrderStatusCompleted {
			if order.AcceptTime > 0 {
				acceptTime := float64(order.AcceptTime - order.CreateTime)
				if acceptTime > 0 {
					totalAcceptTime += acceptTime
					acceptOrderCount++
				}
			}
		}

		if order.Status == common.OrderStatusCompleted && order.ActualFare > 0 {
			totalFare += order.ActualFare
			completedWithFareCount++
		}

		if order.Rating != nil && order.Rating.CreateTime >= sevenDaysAgoUnix {
			totalReviews++
			if order.Rating.Stars >= 4 {
				goodReviews++
			}
		}
	}

	metrics := &common.Metrics{
		TodayOrderCount: todayOrderCount,
		AvgAcceptTime:   0,
		AvgTripFare:     0,
		GoodReviewRate:  0,
	}

	if acceptOrderCount > 0 {
		metrics.AvgAcceptTime = totalAcceptTime / float64(acceptOrderCount)
	}

	if completedWithFareCount > 0 {
		metrics.AvgTripFare = totalFare / float64(completedWithFareCount)
	}

	if totalReviews > 0 {
		metrics.GoodReviewRate = float64(goodReviews) / float64(totalReviews) * 100
	}

	return metrics
}
