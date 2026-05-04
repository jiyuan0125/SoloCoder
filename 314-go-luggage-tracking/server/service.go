package server

import (
	"fmt"
	"sort"
	"time"

	"luggage-tracking/common"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) Scan(req common.ScanRequest) error {
	if !common.ValidateLuggageTag(req.TagNumber) {
		return common.ErrInvalidLuggageTag
	}

	if !common.ValidateStage(req.Stage) {
		return common.ErrInvalidStage
	}

	luggage, exists := s.store.GetLuggage(req.TagNumber)

	if !exists {
		if req.FlightNumber == "" {
			return common.ErrLuggageNotFound
		}
		luggage = &common.Luggage{
			TagNumber:    req.TagNumber,
			FlightNumber: req.FlightNumber,
			Status:       common.StatusNormal,
			Stages:       []common.LuggageStage{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			Completed:    false,
		}
	} else {
		if luggage.Completed || luggage.Status == common.StatusFlightCancel {
			return common.ErrLuggageAlreadyExists
		}
		if req.FlightNumber != "" && luggage.FlightNumber != req.FlightNumber {
			luggage.FlightNumber = req.FlightNumber
		}
	}

	for _, stage := range luggage.Stages {
		if stage.Stage == req.Stage {
			return common.ErrStageAlreadyScanned
		}
	}

	newStage := common.LuggageStage{
		Stage:        req.Stage,
		ScannedAt:    time.Now(),
		Operator:     req.Operator,
		IsBackfilled: false,
	}

	luggage.Stages = append(luggage.Stages, newStage)
	sortStagesByOrder(luggage.Stages)

	if req.Stage == common.StageConveyor {
		luggage.Completed = true
		luggage.Status = common.StatusCompleted
	}

	luggage.UpdatedAt = time.Now()

	if !exists {
		if err := s.store.CreateLuggage(luggage); err != nil {
			return err
		}
	} else {
		if err := s.store.UpdateLuggage(luggage); err != nil {
			return err
		}
	}

	s.logOperation(req.Operator, common.RoleStaff, "scan",
		fmt.Sprintf("扫码上报行李 %s, 环节: %s", req.TagNumber, common.StageName[req.Stage]))

	return nil
}

func (s *Service) Backfill(req common.BackfillRequest) error {
	if !common.ValidateLuggageTag(req.TagNumber) {
		return common.ErrInvalidLuggageTag
	}

	if !common.ValidateStage(req.Stage) {
		return common.ErrInvalidStage
	}

	luggage, exists := s.store.GetLuggage(req.TagNumber)
	if !exists {
		return common.ErrLuggageNotFound
	}

	if luggage.Completed || luggage.Status == common.StatusFlightCancel {
		return common.ErrLuggageAlreadyExists
	}

	for _, stage := range luggage.Stages {
		if stage.Stage == req.Stage {
			return common.ErrStageAlreadyScanned
		}
	}

	scannedAt := req.ScannedAt
	if scannedAt.IsZero() {
		scannedAt = time.Now()
	}

	newStage := common.LuggageStage{
		Stage:        req.Stage,
		ScannedAt:    scannedAt,
		Operator:     req.Operator,
		IsBackfilled: true,
	}

	luggage.Stages = append(luggage.Stages, newStage)
	sortStagesByOrder(luggage.Stages)

	if req.Stage == common.StageConveyor {
		luggage.Completed = true
		luggage.Status = common.StatusCompleted
	}

	luggage.UpdatedAt = time.Now()

	if err := s.store.UpdateLuggage(luggage); err != nil {
		return err
	}

	s.logOperation(req.Operator, common.RoleStaff, "backfill",
		fmt.Sprintf("补录行李 %s, 环节: %s", req.TagNumber, common.StageName[req.Stage]))

	s.logAudit(req.Operator, "backfill",
		fmt.Sprintf("补录行李 %s 环节 %s, 时间: %s", req.TagNumber, req.Stage, scannedAt.Format(time.RFC3339)))

	return nil
}

func (s *Service) GetLuggage(tagNumber string) (*common.LuggageResponse, error) {
	if !common.ValidateLuggageTag(tagNumber) {
		return nil, common.ErrInvalidLuggageTag
	}

	luggage, exists := s.store.GetLuggage(tagNumber)
	if !exists {
		return nil, common.ErrLuggageNotFound
	}

	return s.toLuggageResponse(luggage), nil
}

func (s *Service) GetFlightLuggages(flightNumber string, date string) (*common.FlightLuggageResponse, error) {
	if flightNumber == "" {
		return nil, common.ErrInvalidFlightNumber
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	luggages := s.store.GetLuggagesByFlight(flightNumber, date)

	responses := make([]common.LuggageResponse, 0, len(luggages))
	for _, l := range luggages {
		responses = append(responses, *s.toLuggageResponse(l))
	}

	return &common.FlightLuggageResponse{
		FlightNumber: flightNumber,
		Date:         date,
		Luggages:     responses,
	}, nil
}

func (s *Service) CancelFlight(req common.FlightCancelRequest) error {
	if req.FlightNumber == "" {
		return common.ErrInvalidFlightNumber
	}

	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	luggages := s.store.GetLuggagesByFlight(req.FlightNumber, date)

	for _, luggage := range luggages {
		if !luggage.Completed && luggage.Status != common.StatusFlightCancel {
			luggage.Status = common.StatusFlightCancel
			luggage.UpdatedAt = time.Now()
			s.store.UpdateLuggage(luggage)
		}
	}

	s.logOperation(req.Operator, common.RoleStaff, "cancel_flight",
		fmt.Sprintf("取消航班 %s, 日期: %s", req.FlightNumber, date))

	s.logAudit(req.Operator, "cancel_flight",
		fmt.Sprintf("取消航班 %s, 日期: %s, 影响行李数: %d", req.FlightNumber, date, len(luggages)))

	return nil
}

func (s *Service) CheckAnomalies() {
	luggages := s.store.GetActiveLuggages()
	now := time.Now()

	for _, luggage := range luggages {
		if len(luggage.Stages) == 0 {
			continue
		}

		lastStage := luggage.Stages[len(luggage.Stages)-1]
		elapsed := now.Sub(lastStage.ScannedAt)

		if elapsed.Minutes() > float64(common.AnomalyTimeoutMinutes) {
			if luggage.Status == common.StatusNormal {
				luggage.Status = common.StatusAnomaly
				luggage.UpdatedAt = now
				s.store.UpdateLuggage(luggage)

				s.logOperation("system", common.RoleAdmin, "anomaly_detect",
					fmt.Sprintf("行李 %s 标记为异常, 最后环节: %s, 超时: %.0f分钟",
						luggage.TagNumber, common.StageName[lastStage.Stage], elapsed.Minutes()))
			}
		}
	}
}

func (s *Service) GetOperationLogs() []common.OperationLog {
	return s.store.GetOperationLogs()
}

func (s *Service) GetAuditLogs() []common.AuditLog {
	return s.store.GetAuditLogs()
}

func (s *Service) toLuggageResponse(luggage *common.Luggage) *common.LuggageResponse {
	currentStage := ""
	if len(luggage.Stages) > 0 {
		lastStage := luggage.Stages[len(luggage.Stages)-1]
		currentStage = common.StageName[lastStage.Stage]
	}

	return &common.LuggageResponse{
		TagNumber:    luggage.TagNumber,
		FlightNumber: luggage.FlightNumber,
		Status:       luggage.Status,
		Stages:       luggage.Stages,
		CurrentStage: currentStage,
		CreatedAt:    luggage.CreatedAt,
	}
}

func sortStagesByOrder(stages []common.LuggageStage) {
	sort.Slice(stages, func(i, j int) bool {
		return common.StageOrder[stages[i].Stage] < common.StageOrder[stages[j].Stage]
	})
}

func (s *Service) logOperation(operator, role, action, details string) {
	s.store.AddOperationLog(common.OperationLog{
		Operator: operator,
		Role:     role,
		Action:   action,
		Details:  details,
	})
}

func (s *Service) logAudit(operator, action, details string) {
	s.store.AddAuditLog(common.AuditLog{
		Operator: operator,
		Action:   action,
		Details:  details,
	})
}
