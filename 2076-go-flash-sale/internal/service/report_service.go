package service

import (
	"time"

	"flashsale/internal/model"
	"flashsale/internal/repository"
)

type ReportService struct {
	reportRepo   *repository.ReportRepository
	orderRepo    *repository.OrderRepository
	activityRepo *repository.ActivityRepository
}

func NewReportService(
	reportRepo *repository.ReportRepository,
	orderRepo *repository.OrderRepository,
	activityRepo *repository.ActivityRepository,
) *ReportService {
	return &ReportService{
		reportRepo:   reportRepo,
		orderRepo:    orderRepo,
		activityRepo: activityRepo,
	}
}

func (s *ReportService) GenerateReport(activityID int64) (*model.Report, error) {
	orders, err := s.orderRepo.GetByActivityID(activityID)
	if err != nil {
		return nil, err
	}

	report := &model.Report{
		ActivityID: activityID,
	}

	var totalRevenue float64

	for _, order := range orders {
		report.TotalOrders++

		switch order.Status {
		case model.OrderStatusPaid:
			report.PaidOrders++
			totalRevenue += order.Price
		case model.OrderStatusPending:
			report.UnpaidOrders++
		case model.OrderStatusCancelled:
			report.CancelledOrders++
		}
	}

	report.TotalRevenue = totalRevenue

	if err := s.reportRepo.Upsert(report); err != nil {
		return nil, err
	}

	verifiedReport, err := s.VerifyReport(activityID)
	if err != nil {
		return nil, err
	}

	return verifiedReport, nil
}

func (s *ReportService) VerifyReport(activityID int64) (*model.Report, error) {
	orders, err := s.orderRepo.GetByActivityID(activityID)
	if err != nil {
		return nil, err
	}

	verifiedReport := &model.Report{
		ActivityID: activityID,
	}

	var totalRevenue float64

	for _, order := range orders {
		verifiedReport.TotalOrders++

		switch order.Status {
		case model.OrderStatusPaid:
			verifiedReport.PaidOrders++
			totalRevenue += order.Price
		case model.OrderStatusPending:
			verifiedReport.UnpaidOrders++
		case model.OrderStatusCancelled:
			verifiedReport.CancelledOrders++
		}
	}

	verifiedReport.TotalRevenue = totalRevenue

	if err := s.reportRepo.Upsert(verifiedReport); err != nil {
		return nil, err
	}

	if err := s.reportRepo.MarkVerified(activityID); err != nil {
		return nil, err
	}

	return s.reportRepo.GetByActivityID(activityID)
}

func (s *ReportService) GetReport(activityID int64) (*model.Report, error) {
	return s.reportRepo.GetByActivityID(activityID)
}

func (s *ReportService) ProcessEndedActivities(stockRepo *repository.StockRepository) error {
	activities, err := s.activityRepo.ListActive()
	if err != nil {
		return err
	}

	for _, activity := range activities {
		stock, err := stockRepo.GetByActivityID(activity.ID)
		if err != nil {
			continue
		}

		if stock.AvailableStock == 0 || time.Now().After(activity.StartTime.Add(24*time.Hour)) {
			_, err := s.GenerateReport(activity.ID)
			if err != nil {
				continue
			}
		}
	}

	return nil
}
