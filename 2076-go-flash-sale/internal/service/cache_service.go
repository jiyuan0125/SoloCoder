package service

import (
	"encoding/json"
	"fmt"
	"time"

	"flashsale/internal/model"
	"flashsale/internal/repository"
)

type CacheService struct {
	cacheRepo    *repository.CacheRepository
	activityRepo *repository.ActivityRepository
	stockRepo    *repository.StockRepository
	reportRepo   *repository.ReportRepository
}

func NewCacheService(
	cacheRepo *repository.CacheRepository,
	activityRepo *repository.ActivityRepository,
	stockRepo *repository.StockRepository,
	reportRepo *repository.ReportRepository,
) *CacheService {
	return &CacheService{
		cacheRepo:    cacheRepo,
		activityRepo: activityRepo,
		stockRepo:    stockRepo,
		reportRepo:   reportRepo,
	}
}

func (s *CacheService) NotifyModule(moduleName, eventType string, activityID int64, data interface{}) error {
	var payload string
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		payload = string(jsonData)
	}

	notification := &model.CacheNotification{
		ModuleName: moduleName,
		EventType:  eventType,
		ActivityID: activityID,
		Payload:    payload,
	}

	return s.cacheRepo.CreateNotification(notification)
}

func (s *CacheService) NotifyActivityCreated(activityID int64) error {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return err
	}

	return s.NotifyModule("activity_service", "activity_created", activityID, map[string]interface{}{
		"activity_id":    activityID,
		"product_name":   activity.ProductName,
		"flash_price":    activity.FlashPrice,
		"start_time":     activity.StartTime.Format(time.RFC3339),
	})
}

func (s *CacheService) NotifyStockUpdated(activityID int64) error {
	stock, err := s.stockRepo.GetByActivityID(activityID)
	if err != nil {
		return err
	}

	return s.NotifyModule("stock_service", "stock_updated", activityID, map[string]interface{}{
		"activity_id":     activityID,
		"available_stock": stock.AvailableStock,
		"version":         stock.Version,
	})
}

func (s *CacheService) NotifyOrderCreated(order *model.Order) error {
	return s.NotifyModule("order_service", "order_created", order.ActivityID, map[string]interface{}{
		"order_no":   order.OrderNo,
		"activity_id": order.ActivityID,
		"user_id":    order.UserID,
		"price":      order.Price,
		"expire_at":  order.ExpireAt.Format(time.RFC3339),
	})
}

func (s *CacheService) NotifyOrderStatusChanged(orderNo string, activityID int64, status model.OrderStatus) error {
	return s.NotifyModule("order_service", "order_status_changed", activityID, map[string]interface{}{
		"order_no":   orderNo,
		"activity_id": activityID,
		"status":     string(status),
	})
}

func (s *CacheService) NotifyReportGenerated(report *model.Report) error {
	return s.NotifyModule("report_service", "report_generated", report.ActivityID, map[string]interface{}{
		"activity_id":     report.ActivityID,
		"total_orders":    report.TotalOrders,
		"paid_orders":     report.PaidOrders,
		"unpaid_orders":   report.UnpaidOrders,
		"cancelled_orders": report.CancelledOrders,
		"total_revenue":   report.TotalRevenue,
	})
}

func (s *CacheService) AllocateQuotas(activityID int64, segments map[string]float64) ([]*model.QuotaAllocation, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRepo.DeleteQuotaAllocations(activityID); err != nil {
		return nil, err
	}

	totalStock := activity.TotalStock
	allocations := make([]*model.QuotaAllocation, 0)

	for segmentName, ratio := range segments {
		quota := int(float64(totalStock) * ratio)
		allocation := &model.QuotaAllocation{
			ActivityID:   activityID,
			SegmentName:  segmentName,
			TotalQuota:   quota,
			AllocatedQuota: 0,
			Ratio:        ratio,
		}

		if err := s.cacheRepo.CreateQuotaAllocation(allocation); err != nil {
			return nil, err
		}

		allocations = append(allocations, allocation)
	}

	s.NotifyModule("quota_service", "quota_allocated", activityID, map[string]interface{}{
		"activity_id": activityID,
		"segments":    segments,
	})

	return allocations, nil
}

func (s *CacheService) RedistributeQuotas(activityID int64) error {
	stock, err := s.stockRepo.GetByActivityID(activityID)
	if err != nil {
		return err
	}

	allocations, err := s.cacheRepo.GetQuotaAllocations(activityID)
	if err != nil {
		return err
	}

	if len(allocations) == 0 {
		return fmt.Errorf("no quota allocations found for activity %d", activityID)
	}

	var totalAllocated int
	for _, alloc := range allocations {
		totalAllocated += alloc.AllocatedQuota
	}

	availableStock := stock.AvailableStock
	totalRemaining := availableStock

	var totalRatio float64
	for _, alloc := range allocations {
		totalRatio += alloc.Ratio
	}

	for _, alloc := range allocations {
		newQuota := int(float64(totalRemaining) * (alloc.Ratio / totalRatio))
		alloc.TotalQuota = newQuota

		if err := s.cacheRepo.UpdateQuotaAllocation(alloc); err != nil {
			return err
		}
	}

	s.NotifyModule("quota_service", "quota_redistributed", activityID, map[string]interface{}{
		"activity_id":     activityID,
		"available_stock": availableStock,
	})

	return nil
}

func (s *CacheService) GetQuotaAllocations(activityID int64) ([]*model.QuotaAllocation, error) {
	return s.cacheRepo.GetQuotaAllocations(activityID)
}
