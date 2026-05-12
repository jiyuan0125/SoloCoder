package service

import (
	"fmt"
	"medical-exam-system/internal/model"
	"medical-exam-system/internal/repository"
	"medical-exam-system/pkg/errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAllItems() []*model.ExamItem {
	return s.repo.GetAllItems()
}

func (s *Service) GetPackages(page, size int) ([]*model.Package, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return s.repo.GetPackages(page, size)
}

func (s *Service) GetPackageByID(id string) (*model.Package, error) {
	pkg, ok := s.repo.GetPackageByID(id)
	if !ok {
		return nil, errors.ErrPackageNotFound
	}
	return pkg, nil
}

func (s *Service) GetPackageItems(pkg *model.Package) []*model.ExamItem {
	items := []*model.ExamItem{}
	for _, itemID := range pkg.ItemIDs {
		if item, ok := s.repo.GetItemByID(itemID); ok {
			items = append(items, item)
		}
	}
	return items
}

func (s *Service) CreatePackage(pkg *model.Package) error {
	for _, itemID := range pkg.ItemIDs {
		if _, ok := s.repo.GetItemByID(itemID); !ok {
			return errors.ErrItemNotFound
		}
	}
	s.repo.CreatePackage(pkg)
	return nil
}

func (s *Service) UpdatePackage(pkg *model.Package) error {
	existing, ok := s.repo.GetPackageByID(pkg.ID)
	if !ok {
		return errors.ErrPackageNotFound
	}
	if len(pkg.ItemIDs) > 0 {
		for _, itemID := range pkg.ItemIDs {
			if _, ok := s.repo.GetItemByID(itemID); !ok {
				return errors.ErrItemNotFound
			}
		}
		existing.ItemIDs = pkg.ItemIDs
	}
	if pkg.Name != "" {
		existing.Name = pkg.Name
	}
	if pkg.Description != "" {
		existing.Description = pkg.Description
	}
	if pkg.Price > 0 {
		existing.Price = pkg.Price
	}
	s.repo.UpdatePackage(existing)
	return nil
}

func (s *Service) DeletePackage(id string) error {
	_, ok := s.repo.GetPackageByID(id)
	if !ok {
		return errors.ErrPackageNotFound
	}
	s.repo.DeletePackage(id)
	return nil
}

func (s *Service) CalculatePrice(packageID string, addItemIDs []string) (packagePrice, addItemsPrice, addItemsDiscount, addItemsFinalPrice, totalPrice int, err error) {
	pkg, ok := s.repo.GetPackageByID(packageID)
	if !ok {
		return 0, 0, 0, 0, 0, errors.ErrPackageNotFound
	}

	packagePrice = pkg.Price
	addItemsPrice = 0

	for _, itemID := range addItemIDs {
		item, ok := s.repo.GetItemByID(itemID)
		if !ok {
			return 0, 0, 0, 0, 0, errors.ErrItemNotFound
		}
		addItemsPrice += item.Price
	}

	addItemsFinalPrice = addItemsPrice
	addItemsDiscount = 0

	if addItemsPrice > 80000 {
		addItemsFinalPrice = addItemsPrice * 90 / 100
		addItemsDiscount = addItemsPrice - addItemsFinalPrice
	} else if addItemsPrice > 30000 {
		addItemsFinalPrice = addItemsPrice * 95 / 100
		addItemsDiscount = addItemsPrice - addItemsFinalPrice
	}

	totalPrice = packagePrice + addItemsFinalPrice

	return packagePrice, addItemsPrice, addItemsDiscount, addItemsFinalPrice, totalPrice, nil
}

func ValidateIDCard(idCard string) error {
	if len(idCard) != 18 {
		return errors.ErrInvalidIDCardLength
	}

	re := regexp.MustCompile(`^[0-9]{17}[0-9Xx]$`)
	if !re.MatchString(idCard) {
		return errors.ErrInvalidIDCardFormat
	}

	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

	sum := 0
	for i := 0; i < 17; i++ {
		num, _ := strconv.Atoi(string(idCard[i]))
		sum += num * weights[i]
	}

	mod := sum % 11
	expectedCode := checkCodes[mod]
	actualCode := strings.ToUpper(string(idCard[17]))

	if expectedCode != actualCode {
		return errors.ErrInvalidIDCardFormat
	}

	return nil
}

func (s *Service) GetAvailability(date string) (model.TimeSlot, int, int, bool) {
	return model.TimeSlotMorning, 90, s.repo.GetTimeSlotCount(date, model.TimeSlotMorning), false
}

func (s *Service) GetDayAvailability(dateStr string) (morningUsed, afternoonUsed int) {
	return s.repo.GetTimeSlotCount(dateStr, model.TimeSlotMorning),
		s.repo.GetTimeSlotCount(dateStr, model.TimeSlotAfternoon)
}

func (s *Service) CreateAppointment(customerName, idCard, phone, gender string, age int, packageID string, addItemIDs []string, examDate time.Time, timeSlot model.TimeSlot) (*model.Appointment, error) {
	if err := ValidateIDCard(idCard); err != nil {
		return nil, err
	}

	if examDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.ErrInvalidExamDate
	}

	dateStr := examDate.Format("20060102")

	if !s.repo.IncrementTimeSlotCount(dateStr, timeSlot) {
		return nil, errors.ErrTimeSlotFull
	}

	_, _, _, _, totalPrice, err := s.CalculatePrice(packageID, addItemIDs)
	if err != nil {
		s.repo.DecrementTimeSlotCount(dateStr, timeSlot)
		return nil, err
	}

	examNumber := s.repo.GenerateExamNumber(dateStr)

	appt := &model.Appointment{
		ExamNumber:   examNumber,
		CustomerName: customerName,
		IDCard:       idCard,
		Phone:        phone,
		Gender:       gender,
		Age:          age,
		PackageID:    packageID,
		AddItemIDs:   addItemIDs,
		TotalPrice:   totalPrice,
		ExamDate:     examDate,
		TimeSlot:     timeSlot,
		Status:       model.AppointmentStatusPending,
	}

	s.repo.CreateAppointment(appt)

	s.initExamResults(appt)

	return appt, nil
}

func (s *Service) initExamResults(appt *model.Appointment) {
	pkg, ok := s.repo.GetPackageByID(appt.PackageID)
	if !ok {
		return
	}

	allItemIDs := make(map[string]bool)
	for _, itemID := range pkg.ItemIDs {
		allItemIDs[itemID] = true
	}
	for _, itemID := range appt.AddItemIDs {
		allItemIDs[itemID] = true
	}

	for itemID := range allItemIDs {
		result := &model.ExamResult{
			AppointmentID: appt.ID,
			ItemID:        itemID,
			ResultValue:   nil,
			IsAbnormal:    false,
			IsCritical:    false,
			Status:        model.ResultStatusPending,
			ModifyCount:   0,
			ModifiedAt:    []time.Time{},
		}
		s.repo.CreateExamResult(result)
	}
}

func (s *Service) GetAppointmentByID(id string) (*model.Appointment, error) {
	appt, ok := s.repo.GetAppointmentByID(id)
	if !ok {
		return nil, errors.ErrAppointmentNotFound
	}
	return appt, nil
}

func (s *Service) GetAppointments(page, size int) ([]*model.Appointment, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return s.repo.GetAppointments(page, size)
}

func (s *Service) CancelAppointment(id string) error {
	appt, ok := s.repo.GetAppointmentByID(id)
	if !ok {
		return errors.ErrAppointmentNotFound
	}
	if appt.Status == model.AppointmentStatusCancelled {
		return errors.ErrAppointmentCancelled
	}

	dateStr := appt.ExamDate.Format("20060102")
	s.repo.DecrementTimeSlotCount(dateStr, appt.TimeSlot)

	appt.Status = model.AppointmentStatusCancelled
	s.repo.UpdateAppointment(appt)

	return nil
}

func (s *Service) GetExamResultsByAppointment(appointmentID string) []*model.ExamResult {
	return s.repo.GetExamResultsByAppointment(appointmentID)
}

func (s *Service) SubmitExamResult(resultID string, value float64) (*model.ExamResult, error) {
	result, ok := s.repo.GetExamResultByID(resultID)
	if !ok {
		return nil, errors.ErrResultNotFound
	}

	isFirstSubmission := result.Status == model.ResultStatusPending

	if !isFirstSubmission {
		if result.ModifyCount >= 3 {
			return nil, errors.ErrModifyLimitExceeded
		}
		result.ModifyCount++
		result.ModifiedAt = append(result.ModifiedAt, time.Now())
	}

	item, ok := s.repo.GetItemByID(result.ItemID)
	if !ok {
		return nil, errors.ErrItemNotFound
	}

	result.ResultValue = &value
	result.Status = model.ResultStatusCompleted

	result.IsAbnormal = false
	result.AbnormalType = ""
	result.IsCritical = false

	if item.RefRange.Min != nil {
		if value < *item.RefRange.Min {
			result.IsAbnormal = true
			result.AbnormalType = "偏低"
			if value < *item.RefRange.Min*0.5 {
				result.IsCritical = true
			}
		}
	}

	if item.RefRange.Max != nil {
		if value > *item.RefRange.Max {
			result.IsAbnormal = true
			result.AbnormalType = "偏高"
			if value > *item.RefRange.Max*1.5 {
				result.IsCritical = true
			}
		}
	}

	s.repo.UpdateExamResult(result)

	if result.IsCritical {
		appt, _ := s.repo.GetAppointmentByID(result.AppointmentID)
		s.repo.CreateCriticalAlert(&model.CriticalAlert{
			AppointmentID: result.AppointmentID,
			ExamNumber:    appt.ExamNumber,
			CustomerName:  appt.CustomerName,
			ItemID:        item.ID,
			ItemName:      item.Name,
			ResultValue:   value,
			RefRange:      formatRefRange(item.RefRange),
			Resolved:      false,
		})
	}

	s.tryGenerateReport(result.AppointmentID)

	return result, nil
}

func (s *Service) tryGenerateReport(appointmentID string) {
	appt, ok := s.repo.GetAppointmentByID(appointmentID)
	if !ok {
		return
	}

	if _, exists := s.repo.GetReportByAppointment(appointmentID); exists {
		return
	}

	results := s.repo.GetExamResultsByAppointment(appointmentID)
	if len(results) == 0 {
		return
	}

	allCompleted := true
	for _, r := range results {
		if r.Status != model.ResultStatusCompleted {
			allCompleted = false
			break
		}
	}

	if !allCompleted {
		return
	}

	pkg, _ := s.repo.GetPackageByID(appt.PackageID)

	reportItems := []model.ReportItem{}
	hasAbnormal := false

	for _, result := range results {
		item, _ := s.repo.GetItemByID(result.ItemID)
		reportItem := model.ReportItem{
			ItemID:       item.ID,
			ItemName:     item.Name,
			Department:   item.Department,
			ResultValue:  result.ResultValue,
			Unit:         item.Unit,
			RefRange:     formatRefRange(item.RefRange),
			IsAbnormal:   result.IsAbnormal,
			AbnormalType: result.AbnormalType,
			IsCritical:   result.IsCritical,
		}
		reportItems = append(reportItems, reportItem)
		if result.IsAbnormal {
			hasAbnormal = true
		}
	}

	report := &model.Report{
		AppointmentID:  appt.ID,
		ExamNumber:     appt.ExamNumber,
		CustomerName:   appt.CustomerName,
		Gender:         appt.Gender,
		Age:            appt.Age,
		PackageName:    pkg.Name,
		Items:          reportItems,
		HasAbnormal:    hasAbnormal,
		Status:         model.ReportStatusDraft,
	}

	s.repo.CreateReport(report)
}

func formatRefRange(r model.RefRange) string {
	if r.Min != nil && r.Max != nil {
		return fmt.Sprintf("%.2f-%.2f", *r.Min, *r.Max)
	} else if r.Min != nil {
		return fmt.Sprintf(">%.2f", *r.Min)
	} else if r.Max != nil {
		return fmt.Sprintf("<%.2f", *r.Max)
	}
	return ""
}

func (s *Service) GetRecentCriticalAlerts(limit int) []*model.CriticalAlert {
	return s.repo.GetRecentCriticalAlerts(limit)
}

func (s *Service) GetReports(page, size int) ([]*model.Report, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return s.repo.GetReports(page, size)
}

func (s *Service) GetReportByID(id string) (*model.Report, error) {
	report, ok := s.repo.GetReportByID(id)
	if !ok {
		return nil, errors.ErrReportNotFound
	}
	return report, nil
}

func (s *Service) UpdateReportContent(id, generalAdvice, followUpAdvice string) (*model.Report, error) {
	report, ok := s.repo.GetReportByID(id)
	if !ok {
		return nil, errors.ErrReportNotFound
	}
	if report.Status == model.ReportStatusPublished {
		return nil, errors.ErrInvalidReportStatus
	}
	report.GeneralAdvice = generalAdvice
	report.FollowUpAdvice = followUpAdvice
	s.repo.UpdateReport(report)
	return report, nil
}

func (s *Service) SubmitReportForReview(id string) (*model.Report, error) {
	report, ok := s.repo.GetReportByID(id)
	if !ok {
		return nil, errors.ErrReportNotFound
	}
	if report.Status != model.ReportStatusDraft {
		return nil, errors.ErrInvalidReportStatus
	}
	report.Status = model.ReportStatusReviewing
	s.repo.UpdateReport(report)
	return report, nil
}

func (s *Service) PublishReport(id string) (*model.Report, error) {
	report, ok := s.repo.GetReportByID(id)
	if !ok {
		return nil, errors.ErrReportNotFound
	}
	if report.Status != model.ReportStatusReviewing {
		return nil, errors.ErrInvalidReportStatus
	}
	now := time.Now()
	report.Status = model.ReportStatusPublished
	report.PublishedAt = &now
	s.repo.UpdateReport(report)
	return report, nil
}

func (s *Service) GetReportsByDateRange(startStr, endStr string) ([]*model.Report, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, errors.ErrInvalidDateRange
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return nil, errors.ErrInvalidDateRange
	}
	if start.After(end) {
		return nil, errors.ErrInvalidDateRange
	}
	return s.repo.GetReportsByDateRange(start, end), nil
}
