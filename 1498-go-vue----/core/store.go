package core

import (
	"sync"

	"housekeeping-training/common"
)

type Store struct {
	mu sync.RWMutex

	courses        map[string]*common.Course
	schedules      map[string]*common.ClassSchedule
	students       map[string]*common.Student
	enrollments    map[string]*common.Enrollment
	attendance     map[string]*common.AttendanceRecord
	exams          map[string]*common.Exam
	certificates   map[string]*common.Certificate
	employers      map[string]*common.Employer
	jobPostings    map[string]*common.JobPosting
	recommendations map[string]*common.Recommendation
	refunds        map[string]*common.RefundRequest

	nextCourseID        int
	nextScheduleID      int
	nextStudentID       int
	nextEnrollmentID    int
	nextAttendanceID    int
	nextExamID          int
	nextCertificateID   int
	nextEmployerID      int
	nextJobPostingID    int
	nextRecommendationID int
	nextRefundID        int
	certSequence        int
}

func NewStore() *Store {
	return &Store{
		courses:        make(map[string]*common.Course),
		schedules:      make(map[string]*common.ClassSchedule),
		students:       make(map[string]*common.Student),
		enrollments:    make(map[string]*common.Enrollment),
		attendance:     make(map[string]*common.AttendanceRecord),
		exams:          make(map[string]*common.Exam),
		certificates:   make(map[string]*common.Certificate),
		employers:      make(map[string]*common.Employer),
		jobPostings:    make(map[string]*common.JobPosting),
		recommendations: make(map[string]*common.Recommendation),
		refunds:        make(map[string]*common.RefundRequest),
		certSequence:   1,
	}
}

func (s *Store) Lock()   { s.mu.Lock() }
func (s *Store) Unlock() { s.mu.Unlock() }
func (s *Store) RLock()  { s.mu.RLock() }
func (s *Store) RUnlock(){ s.mu.RUnlock() }
