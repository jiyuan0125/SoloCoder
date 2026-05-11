package core

import (
	"errors"
	"fmt"
	"time"

	"housekeeping-training/common"
)

func (s *Store) RecordExam(enrollmentID, examDateStr string, writtenScore, practicalScore int, isRetake bool) (*common.Exam, error) {
	if enrollmentID == "" {
		return nil, errors.New("enrollment ID is required")
	}
	if examDateStr == "" {
		return nil, errors.New("exam date is required")
	}
	if writtenScore < 0 || writtenScore > 50 {
		return nil, errors.New("written score must be between 0 and 50")
	}
	if practicalScore < 0 || practicalScore > 50 {
		return nil, errors.New("practical score must be between 0 and 50")
	}

	examDate, err := time.Parse("2006-01-02", examDateStr)
	if err != nil {
		return nil, errors.New("invalid exam date format, use YYYY-MM-DD")
	}

	s.Lock()
	defer s.Unlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return nil, errors.New("enrollment not found")
	}

	hasPassedExam := false
	hasTakenFirstExam := false
	hasTakenRetake := false

	for _, e := range s.exams {
		if e.EnrollmentID != enrollmentID {
			continue
		}
		if e.Status == common.ExamResultPassed {
			hasPassedExam = true
		}
		if e.IsRetake {
			hasTakenRetake = true
		} else {
			hasTakenFirstExam = true
		}
	}

	if hasPassedExam {
		return nil, errors.New("student has already passed this exam")
	}

	if isRetake {
		if !hasTakenFirstExam {
			return nil, errors.New("must take first exam before retake")
		}
		if hasTakenRetake {
			return nil, errors.New("retake already used")
		}
	} else {
		if hasTakenFirstExam {
			return nil, errors.New("first exam already recorded, use retake")
		}
	}

	totalScore := writtenScore + practicalScore
	status := common.ExamResultFailed
	if totalScore >= 60 {
		status = common.ExamResultPassed
	}

	course, ok := s.courses[s.schedules[enrollment.ScheduleID].CourseID]
	retakeFee := 0
	if isRetake {
		retakeFee = int(float64(course.Fee) * 0.1)
		if retakeFee < 50 {
			retakeFee = 50
		}
	}

	s.nextExamID++
	exam := &common.Exam{
		ID:             fmt.Sprintf("EX%04d", s.nextExamID),
		EnrollmentID:   enrollmentID,
		ExamDate:       examDate,
		WrittenScore:   writtenScore,
		PracticalScore: practicalScore,
		TotalScore:     totalScore,
		Status:         status,
		IsRetake:       isRetake,
		RetakeFee:      retakeFee,
	}

	s.exams[exam.ID] = exam

	if status == common.ExamResultPassed {
		s.createCertificate(enrollment.StudentID, s.schedules[enrollment.ScheduleID].CourseID, exam.ID)
	}

	return exam, nil
}

func (s *Store) createCertificate(studentID, courseID, examID string) {
	now := time.Now()
	yearMonth := fmt.Sprintf("%04d%02d", now.Year(), now.Month())
	certNo := fmt.Sprintf("HZ%s%04d", yearMonth, s.certSequence)
	s.certSequence++

	s.nextCertificateID++
	cert := &common.Certificate{
		ID:        fmt.Sprintf("CE%04d", s.nextCertificateID),
		StudentID: studentID,
		CourseID:  courseID,
		ExamID:    examID,
		CertNo:    certNo,
		IssueDate: now,
	}
	s.certificates[cert.ID] = cert
}

func (s *Store) GetExam(id string) (*common.Exam, error) {
	s.RLock()
	defer s.RUnlock()

	exam, ok := s.exams[id]
	if !ok {
		return nil, errors.New("exam not found")
	}
	return exam, nil
}

func (s *Store) ListExams() []*common.Exam {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Exam, 0, len(s.exams))
	for _, e := range s.exams {
		result = append(result, e)
	}
	return result
}

func (s *Store) GetCertificate(id string) (*common.Certificate, error) {
	s.RLock()
	defer s.RUnlock()

	cert, ok := s.certificates[id]
	if !ok {
		return nil, errors.New("certificate not found")
	}
	return cert, nil
}

func (s *Store) ListCertificates() []*common.Certificate {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Certificate, 0, len(s.certificates))
	for _, c := range s.certificates {
		result = append(result, c)
	}
	return result
}

func (s *Store) GetStudentCertificates(studentID string) []*common.Certificate {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Certificate, 0)
	for _, c := range s.certificates {
		if c.StudentID == studentID {
			result = append(result, c)
		}
	}
	return result
}

func (s *Store) HasCertificate(studentID, courseID string) bool {
	s.RLock()
	defer s.RUnlock()

	for _, c := range s.certificates {
		if c.StudentID == studentID && c.CourseID == courseID {
			return true
		}
	}
	return false
}

func (s *Store) HasTakenExam(enrollmentID string) bool {
	s.RLock()
	defer s.RUnlock()

	for _, e := range s.exams {
		if e.EnrollmentID == enrollmentID {
			return true
		}
	}
	return false
}
