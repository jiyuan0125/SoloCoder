package core

import (
	"errors"
	"fmt"
	"time"

	"housekeeping-training/common"
)

func (s *Store) CreateStudent(name, idCard, phone, education string) (*common.Student, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if idCard == "" {
		return nil, errors.New("ID card is required")
	}
	if phone == "" {
		return nil, errors.New("phone is required")
	}

	s.Lock()
	defer s.Unlock()

	s.nextStudentID++
	student := &common.Student{
		ID:         fmt.Sprintf("ST%04d", s.nextStudentID),
		Name:       name,
		IDCard:     idCard,
		Phone:      phone,
		Education:  education,
		EnrollDate: time.Now(),
	}
	s.students[student.ID] = student
	return student, nil
}

func (s *Store) GetStudent(id string) (*common.Student, error) {
	s.RLock()
	defer s.RUnlock()

	student, ok := s.students[id]
	if !ok {
		return nil, errors.New("student not found")
	}
	return student, nil
}

func (s *Store) ListStudents() []*common.Student {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Student, 0, len(s.students))
	for _, st := range s.students {
		result = append(result, st)
	}
	return result
}

func (s *Store) EnrollStudent(studentID, scheduleID string) (*common.Enrollment, error) {
	if studentID == "" {
		return nil, errors.New("student ID is required")
	}
	if scheduleID == "" {
		return nil, errors.New("schedule ID is required")
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.students[studentID]; !ok {
		return nil, errors.New("student not found")
	}

	schedule, ok := s.schedules[scheduleID]
	if !ok {
		return nil, errors.New("schedule not found")
	}

	course, ok := s.courses[schedule.CourseID]
	if !ok {
		return nil, errors.New("course not found")
	}

	if schedule.Enrolled >= course.MaxCapacity {
		return nil, errors.New("schedule is full")
	}

	if s.hasScheduleConflict(studentID, schedule) {
		return nil, errors.New("time conflict with existing enrollment")
	}

	halfFee := course.Fee / 2

	s.nextEnrollmentID++
	enrollment := &common.Enrollment{
		ID:            fmt.Sprintf("E%04d", s.nextEnrollmentID),
		StudentID:     studentID,
		ScheduleID:    scheduleID,
		EnrollDate:    time.Now(),
		FirstPayment:  halfFee,
		SecondPayment: course.Fee - halfFee,
		PaidFirst:     false,
		PaidSecond:    false,
	}

	s.enrollments[enrollment.ID] = enrollment
	schedule.Enrolled++

	return enrollment, nil
}

func (s *Store) hasScheduleConflict(studentID string, newSchedule *common.ClassSchedule) bool {
	for _, e := range s.enrollments {
		if e.StudentID != studentID {
			continue
		}
		existingSchedule, ok := s.schedules[e.ScheduleID]
		if !ok {
			continue
		}
		if existingSchedule.TimeSlot == newSchedule.TimeSlot {
			return true
		}
	}
	return false
}

func (s *Store) GetEnrollment(id string) (*common.Enrollment, error) {
	s.RLock()
	defer s.RUnlock()

	enrollment, ok := s.enrollments[id]
	if !ok {
		return nil, errors.New("enrollment not found")
	}
	return enrollment, nil
}

func (s *Store) ListEnrollments() []*common.Enrollment {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Enrollment, 0, len(s.enrollments))
	for _, e := range s.enrollments {
		result = append(result, e)
	}
	return result
}

func (s *Store) ListStudentEnrollments(studentID string) []*common.Enrollment {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Enrollment, 0)
	for _, e := range s.enrollments {
		if e.StudentID == studentID {
			result = append(result, e)
		}
	}
	return result
}

func (s *Store) PayFirstInstallment(enrollmentID string) error {
	s.Lock()
	defer s.Unlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return errors.New("enrollment not found")
	}
	if enrollment.PaidFirst {
		return errors.New("first installment already paid")
	}
	enrollment.PaidFirst = true
	return nil
}

func (s *Store) PaySecondInstallment(enrollmentID string) error {
	s.Lock()
	defer s.Unlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return errors.New("enrollment not found")
	}
	if !enrollment.PaidFirst {
		return errors.New("first installment not paid yet")
	}
	if enrollment.PaidSecond {
		return errors.New("second installment already paid")
	}

	schedule, ok := s.schedules[enrollment.ScheduleID]
	if !ok {
		return errors.New("schedule not found")
	}

	classesTaken := s.countClassesTaken(enrollmentID)
	if classesTaken*2 < schedule.TotalClasses {
		return errors.New("course not yet halfway")
	}

	enrollment.PaidSecond = true
	return nil
}

func (s *Store) countClassesTaken(enrollmentID string) int {
	count := 0
	for _, a := range s.attendance {
		if a.EnrollmentID == enrollmentID {
			count++
		}
	}
	return count
}
