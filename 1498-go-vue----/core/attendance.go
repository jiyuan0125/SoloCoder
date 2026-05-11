package core

import (
	"errors"
	"fmt"
	"time"

	"housekeeping-training/common"
)

func (s *Store) RecordAttendance(enrollmentID, classDateStr string, status common.AttendanceStatus) (*common.AttendanceRecord, error) {
	if enrollmentID == "" {
		return nil, errors.New("enrollment ID is required")
	}
	if classDateStr == "" {
		return nil, errors.New("class date is required")
	}
	if status != common.AttendancePresent && status != common.AttendanceLeave && status != common.AttendanceAbsent {
		return nil, errors.New("invalid attendance status")
	}

	classDate, err := time.Parse("2006-01-02", classDateStr)
	if err != nil {
		return nil, errors.New("invalid class date format, use YYYY-MM-DD")
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.enrollments[enrollmentID]; !ok {
		return nil, errors.New("enrollment not found")
	}

	s.nextAttendanceID++
	record := &common.AttendanceRecord{
		ID:           fmt.Sprintf("A%04d", s.nextAttendanceID),
		EnrollmentID: enrollmentID,
		ClassDate:    classDate,
		Status:       status,
	}
	s.attendance[record.ID] = record
	return record, nil
}

func (s *Store) GetAttendanceRate(enrollmentID string) (float64, error) {
	s.RLock()
	defer s.RUnlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return 0, errors.New("enrollment not found")
	}

	schedule, ok := s.schedules[enrollment.ScheduleID]
	if !ok {
		return 0, errors.New("schedule not found")
	}

	total := 0
	present := 0

	for _, a := range s.attendance {
		if a.EnrollmentID == enrollmentID {
			total++
			if a.Status == common.AttendancePresent {
				present++
			}
		}
	}

	if total == 0 {
		total = schedule.TotalClasses
	}

	return float64(present) / float64(total), nil
}

func (s *Store) CanTakeExam(enrollmentID string) (bool, string, error) {
	rate, err := s.GetAttendanceRate(enrollmentID)
	if err != nil {
		return false, "", err
	}
	if rate < 0.8 {
		return false, fmt.Sprintf("attendance rate is %.1f%%, below 80%%", rate*100), nil
	}
	return true, "", nil
}

func (s *Store) ListAttendanceRecords(enrollmentID string) []*common.AttendanceRecord {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.AttendanceRecord, 0)
	for _, a := range s.attendance {
		if a.EnrollmentID == enrollmentID {
			result = append(result, a)
		}
	}
	return result
}
