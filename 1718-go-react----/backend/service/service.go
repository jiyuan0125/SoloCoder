package service

import (
	"clinical-path-backend/models"
	"clinical-path-backend/store"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) CreatePath(path *models.ClinicalPath) error {
	return s.store.CreatePath(path)
}

func (s *Service) UpdatePath(path *models.ClinicalPath) error {
	return s.store.UpdatePath(path)
}

func (s *Service) ListPaths() []*models.ClinicalPath {
	return s.store.ListPaths()
}

func (s *Service) GetPath(id int64) (*models.ClinicalPath, error) {
	return s.store.GetPath(id)
}

func (s *Service) CreateStage(stage *models.PathStage) error {
	return s.store.CreateStage(stage)
}

func (s *Service) ListStages(pathID int64) []*models.PathStage {
	return s.store.ListStages(pathID)
}

func (s *Service) CreateItem(item *models.OrderItem) error {
	return s.store.CreateItem(item)
}

func (s *Service) ListItems(pathID int64) []*models.OrderItem {
	return s.store.ListItems(pathID)
}

func (s *Service) ListItemsByStage(stageID int64) []*models.OrderItem {
	return s.store.ListItemsByStage(stageID)
}

func (s *Service) ListPatients() []*models.Patient {
	return s.store.ListPatients()
}

func (s *Service) GetPatient(id int64) (*models.Patient, error) {
	return s.store.GetPatient(id)
}

func (s *Service) GetPatientByNo(hospitalNo string) (*models.Patient, error) {
	return s.store.GetPatientByNo(hospitalNo)
}

func (s *Service) CreatePatient(patient *models.Patient) error {
	return s.store.CreatePatient(patient)
}

func diagnosisMatchesPath(diagnosis string, path *models.ClinicalPath) bool {
	icdCodes := strings.Split(path.ICDCodes, ",")
	for _, code := range icdCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if strings.Contains(diagnosis, code) {
			return true
		}
	}
	return false
}

func (s *Service) GenerateDailyOrders(pathID int64, enrollmentID int64, startDate time.Time) ([]*models.DailyOrder, error) {
	stages := s.store.ListStages(pathID)
	items := s.store.ListItems(pathID)

	itemsByStage := make(map[int64][]*models.OrderItem)
	for _, item := range items {
		itemsByStage[item.StageID] = append(itemsByStage[item.StageID], item)
	}

	orders := make([]*models.DailyOrder, 0)
	currentDate := startDate
	pathDay := 1

	for _, stage := range stages {
		stageItems := itemsByStage[stage.ID]
		for day := 0; day < stage.Days; day++ {
			for _, item := range stageItems {
				order := &models.DailyOrder{
					PathDay:    pathDay,
					OrderDate:  currentDate,
					StageID:    stage.ID,
					StageName:  stage.Name,
					ItemID:     item.ID,
					ItemName:   item.Name,
					Category:   item.Category,
					IsRequired: item.IsRequired,
					IsExecuted: false,
					IsVariation: false,
				}
				orders = append(orders, order)
			}
			currentDate = currentDate.AddDate(0, 0, 1)
			pathDay++
		}
	}

	return orders, nil
}

func (s *Service) EnrollPatient(patientID int64, pathID int64) (*models.PatientEnrollment, error) {
	patient, err := s.store.GetPatient(patientID)
	if err != nil {
		return nil, err
	}

	path, err := s.store.GetPath(pathID)
	if err != nil {
		return nil, err
	}

	if !diagnosisMatchesPath(patient.Diagnosis, path) {
		return nil, fmt.Errorf("患者诊断不在路径适用范围内")
	}

	if s.store.HasActiveEnrollment(patientID) {
		return nil, fmt.Errorf("患者已在其他路径中")
	}

	now := time.Now()
	enrollment := &models.PatientEnrollment{
		PatientID:         patientID,
		PathID:            pathID,
		PatientHospitalNo: patient.HospitalNo,
		PatientName:       patient.Name,
		Diagnosis:         patient.Diagnosis,
		EnrollDate:        now,
		ActualStartDate:   now,
		Status:            "active",
		VariationCount:    0,
		SuggestExit:       false,
	}

	orders, err := s.GenerateDailyOrders(pathID, 0, now)
	if err != nil {
		return nil, err
	}

	err = s.store.CreateEnrollment(enrollment, orders)
	if err != nil {
		return nil, err
	}

	return enrollment, nil
}

func (s *Service) ListEnrollments() []*models.PatientEnrollment {
	return s.store.ListEnrollments()
}

func (s *Service) GetEnrollment(id int64) (*models.PatientEnrollment, error) {
	return s.store.GetEnrollment(id)
}

func (s *Service) ListDailyOrders(enrollmentID int64) []*models.DailyOrder {
	return s.store.ListDailyOrders(enrollmentID)
}

func (s *Service) ExecuteOrder(orderID int64, executed bool) (*models.DailyOrder, error) {
	order, err := s.store.GetDailyOrder(orderID)
	if err != nil {
		return nil, err
	}

	enrollment, err := s.store.GetEnrollment(order.EnrollmentID)
	if err != nil {
		return nil, err
	}

	if enrollment.Status != "active" {
		return nil, fmt.Errorf("只能处理当前活动路径中的医嘱")
	}

	now := time.Now()
	if executed {
		order.IsExecuted = true
		order.ExecutedAt = &now
		order.IsVariation = false
	} else {
		if order.IsRequired {
			order.IsExecuted = false
			order.IsVariation = true
			
			variation := &models.VariationRecord{
				EnrollmentID:    enrollment.ID,
				PatientID:       enrollment.PatientID,
				PatientHospitalNo: enrollment.PatientHospitalNo,
				PathID:          enrollment.PathID,
				Date:            now,
				Content:         fmt.Sprintf("必选医嘱未执行: %s", order.ItemName),
				Reason:          "必选医嘱未执行",
				Type:            "可控变异",
				Action:          "系统自动标记",
			}
			err = s.store.AddVariationAndUpdateEnrollment(variation)
			if err != nil {
				return nil, err
			}
		} else {
			order.IsExecuted = false
			order.IsVariation = false
		}
	}

	err = s.store.UpdateDailyOrder(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *Service) RecordVariation(enrollmentID int64, date time.Time, content, reason, variationType, action string) (*models.VariationRecord, error) {
	enrollment, err := s.store.GetEnrollment(enrollmentID)
	if err != nil {
		return nil, err
	}

	patient, err := s.store.GetPatient(enrollment.PatientID)
	if err != nil {
		return nil, err
	}

	variation := &models.VariationRecord{
		EnrollmentID:      enrollmentID,
		PatientID:         patient.ID,
		PatientHospitalNo: patient.HospitalNo,
		PathID:            enrollment.PathID,
		Date:              date,
		Content:           content,
		Reason:            reason,
		Type:              variationType,
		Action:            action,
	}

	err = s.store.AddVariation(variation)
	if err != nil {
		return nil, err
	}

	return variation, nil
}

func (s *Service) ListVariations(enrollmentID int64) []*models.VariationRecord {
	return s.store.ListVariations(enrollmentID)
}

func (s *Service) ExitEnrollment(enrollmentID int64, exitReason string) error {
	return s.store.ExitEnrollment(enrollmentID, exitReason)
}

func (s *Service) CompleteEnrollment(enrollmentID int64, actualDays int, actualCost float64) error {
	return s.store.CompleteEnrollment(enrollmentID, actualDays, actualCost)
}

func roundToOneDecimal(v float64) float64 {
	return float64(int(v*10+0.5)) / 10.0
}

func (s *Service) CalculateQualityMetrics(year int, month time.Month) ([]*models.QualityMetrics, error) {
	paths := s.store.ListAllPaths()
	patients := s.store.ListAllPatients()

	metricsList := make([]*models.QualityMetrics, 0)

	for _, path := range paths {
		enrollments := s.store.ListEnrollmentsByPathAndMonth(path.ID, year, month)

		eligibleCount := 0
		for _, p := range patients {
			pYear, pMonth, _ := p.AdmissionDate.Date()
			if pYear == year && pMonth == month && diagnosisMatchesPath(p.Diagnosis, path) {
				eligibleCount++
			}
		}

		enrolledCount := len(enrollments)
		completedCount := 0
		exitedCount := 0
		variationPatientCount := 0
		totalStayDays := 0
		totalCost := 0.0

		for _, e := range enrollments {
			if e.Status == "completed" {
				completedCount++
				totalStayDays += e.ActualDays
				totalCost += e.ActualCost
			} else if e.Status == "exited" {
				exitedCount++
			}
			if e.VariationCount > 0 {
				variationPatientCount++
			}
		}

		enrollRate := 0.0
		if eligibleCount > 0 {
			enrollRate = roundToOneDecimal(float64(enrolledCount) / float64(eligibleCount) * 100)
		}

		completeRate := 0.0
		if enrolledCount > 0 {
			completeRate = roundToOneDecimal(float64(completedCount) / float64(enrolledCount) * 100)
		}

		variationRate := 0.0
		if enrolledCount > 0 {
			variationRate = roundToOneDecimal(float64(variationPatientCount) / float64(enrolledCount) * 100)
		}

		avgStayDays := 0.0
		if completedCount > 0 {
			avgStayDays = roundToOneDecimal(float64(totalStayDays) / float64(completedCount))
		}

		avgCost := 0.0
		if completedCount > 0 {
			avgCost = roundToOneDecimal(totalCost / float64(completedCount))
		}

		hasWarning := false
		warningReason := ""
		if completeRate < 70.0 {
			hasWarning = true
			warningReason = "完成率低于70%"
		}
		if variationRate > 40.0 {
			if hasWarning {
				warningReason += "；变异率高于40%"
			} else {
				hasWarning = true
				warningReason = "变异率高于40%"
			}
		}

		metrics := &models.QualityMetrics{
			PathID:          path.ID,
			PathCode:        path.Code,
			PathName:        path.Name,
			Month:           fmt.Sprintf("%04d-%02d", year, month),
			TotalEligible:   eligibleCount,
			EnrolledCount:   enrolledCount,
			CompletedCount:  completedCount,
			ExitedCount:     exitedCount,
			VariationCount:  variationPatientCount,
			EnrollRate:      enrollRate,
			CompleteRate:    completeRate,
			VariationRate:   variationRate,
			AvgStayDays:     avgStayDays,
			AvgCost:         avgCost,
			StandardMinDays: path.MinDays,
			StandardMaxDays: path.MaxDays,
			StandardMinCost: path.MinCost,
			StandardMaxCost: path.MaxCost,
			HasWarning:      hasWarning,
			WarningReason:   warningReason,
		}

		metricsList = append(metricsList, metrics)
	}

	return metricsList, nil
}
