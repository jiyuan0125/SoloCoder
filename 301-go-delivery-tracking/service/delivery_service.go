package service

import (
	"delivery-tracking/models"
	"delivery-tracking/repository"
	"errors"
	"time"
)

var (
	ValidStatuses = map[string]bool{
		"取餐确认": true,
		"商家出发": true,
		"到达小区": true,
		"已送达":   true,
	}

	ErrInvalidStatusName = errors.New("invalid status name")
	ErrInvalidLongitude  = errors.New("invalid longitude, must be between -180 and 180")
	ErrInvalidLatitude   = errors.New("invalid latitude, must be between -90 and 90")
	ErrDuplicateStatus   = errors.New("duplicate status")
)

type DeliveryService interface {
	ReportStatus(request *models.DeliveryStatusRequest) error
	ReportBatchStatuses(request *models.DeliveryStatusBatchRequest) []error
	GetTrajectory(orderNo string) (*models.DeliveryTrajectoryResponse, error)
}

type DeliveryServiceImpl struct {
	repo repository.DeliveryRepository
}

func NewDeliveryService(repo repository.DeliveryRepository) *DeliveryServiceImpl {
	return &DeliveryServiceImpl{repo: repo}
}

func (s *DeliveryServiceImpl) ValidateStatusName(statusName string) error {
	if !ValidStatuses[statusName] {
		return ErrInvalidStatusName
	}
	return nil
}

func (s *DeliveryServiceImpl) ValidateCoordinates(longitude, latitude float64) error {
	if longitude < -180 || longitude > 180 {
		return ErrInvalidLongitude
	}
	if latitude < -90 || latitude > 90 {
		return ErrInvalidLatitude
	}
	return nil
}

func (s *DeliveryServiceImpl) ReportStatus(request *models.DeliveryStatusRequest) error {
	if err := s.ValidateStatusName(request.StatusName); err != nil {
		return err
	}

	if err := s.ValidateCoordinates(request.Longitude, request.Latitude); err != nil {
		return err
	}

	receivedAt := time.Now().Unix()

	status := &models.DeliveryStatus{
		OrderNo:    request.OrderNo,
		StatusName: request.StatusName,
		Longitude:  request.Longitude,
		Latitude:   request.Latitude,
		OccurredAt: request.OccurredAt,
		ReceivedAt: receivedAt,
	}

	err := s.repo.InsertStatus(status)
	if err != nil {
		if err.Error() == "duplicate entry, status not inserted" {
			return ErrDuplicateStatus
		}
		return err
	}

	return nil
}

func (s *DeliveryServiceImpl) ReportBatchStatuses(request *models.DeliveryStatusBatchRequest) []error {
	var errs []error

	for _, req := range request.Statuses {
		err := s.ReportStatus(&req)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *DeliveryServiceImpl) GetTrajectory(orderNo string) (*models.DeliveryTrajectoryResponse, error) {
	statuses, err := s.repo.GetTrajectoryByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}

	response := &models.DeliveryTrajectoryResponse{
		OrderNo:  orderNo,
		Statuses: statuses,
	}

	return response, nil
}
