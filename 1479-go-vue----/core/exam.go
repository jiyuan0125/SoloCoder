package core

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/drivingschool/common"
)

const CancellationApprovalThreshold = 3 * 24 * time.Hour
const WaitingConfirmationTimeout = 24 * time.Hour

func (s *Store) CreateExamPlan(date time.Time, subject common.Subject, venue string, totalQuota int) (*common.ExamPlan, error) {
	if venue == "" {
		return nil, errors.New("考场地址不能为空")
	}
	if totalQuota <= 0 {
		return nil, errors.New("名额必须为正数")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	planID := fmt.Sprintf("EXP_%s_%d", date.Format("20060102"), time.Now().Unix())
	plan := &common.ExamPlan{
		ID:         planID,
		Date:       date,
		Subject:    subject,
		Venue:      venue,
		TotalQuota: totalQuota,
		UsedQuota:  0,
	}

	s.examPlans[planID] = plan
	return plan, nil
}

func (s *Store) GetExamPlan(planID string) (*common.ExamPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plan, exists := s.examPlans[planID]
	if !exists {
		return nil, errors.New("考试计划不存在")
	}
	return plan, nil
}

func (s *Store) ListExamPlans() []*common.ExamPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plans := make([]*common.ExamPlan, 0, len(s.examPlans))
	for _, plan := range s.examPlans {
		plans = append(plans, plan)
	}
	return plans
}

func (s *Store) BookExam(studentID, examPlanID string) (*common.ExamBooking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.students[studentID]
	if !exists {
		return nil, errors.New("学员不存在")
	}

	examPlan, exists := s.examPlans[examPlanID]
	if !exists {
		return nil, errors.New("考试计划不存在")
	}

	nextSubject, err := s.getNextSubjectUnlocked(studentID)
	if err != nil {
		return nil, err
	}

	if examPlan.Subject != nextSubject {
		return nil, errors.New("只能预约下一科目的考试")
	}

	if err := s.checkStudyHoursRequirementUnlocked(studentID, examPlan.Subject); err != nil {
		return nil, err
	}

	for _, booking := range s.examBookings {
		if booking.StudentID == studentID && booking.ExamPlanID == examPlanID &&
			(booking.Status == common.ExamBookingStatusPending || booking.Status == common.ExamBookingStatusConfirmed) {
			return nil, errors.New("已预约该考试")
		}
	}

	waitOrder := 0
	isWaiting := false
	status := common.ExamBookingStatusPending

	if examPlan.UsedQuota >= examPlan.TotalQuota {
		isWaiting = true
		waitOrder = s.getNextWaitOrder(examPlanID)
		status = common.ExamBookingStatusPending
	} else {
		examPlan.UsedQuota++
		status = common.ExamBookingStatusConfirmed
	}

	bookingID := fmt.Sprintf("EBK_%d", time.Now().UnixNano())
	booking := &common.ExamBooking{
		ID:          bookingID,
		StudentID:   studentID,
		ExamPlanID:  examPlanID,
		Status:      status,
		IsWaiting:   isWaiting,
		WaitOrder:   waitOrder,
		CreatedAt:   time.Now(),
	}

	s.examBookings[bookingID] = booking

	return booking, nil
}

func (s *Store) getNextWaitOrder(examPlanID string) int {
	maxOrder := 0
	for _, booking := range s.examBookings {
		if booking.ExamPlanID == examPlanID && booking.IsWaiting {
			if booking.WaitOrder > maxOrder {
				maxOrder = booking.WaitOrder
			}
		}
	}
	return maxOrder + 1
}

func (s *Store) ConfirmWaitingExam(bookingID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, exists := s.examBookings[bookingID]
	if !exists {
		return errors.New("预约不存在")
	}

	if !booking.IsWaiting {
		return errors.New("该预约不在候补队列中")
	}

	if time.Since(booking.CreatedAt) > WaitingConfirmationTimeout {
		booking.Status = common.ExamBookingStatusExpired
		return errors.New("候补确认已超时")
	}

	examPlan := s.examPlans[booking.ExamPlanID]
	if examPlan.UsedQuota >= examPlan.TotalQuota {
		return errors.New("名额已被其他候补学员占用")
	}

	booking.IsWaiting = false
	booking.Status = common.ExamBookingStatusConfirmed
	examPlan.UsedQuota++
	now := time.Now()
	booking.ConfirmedAt = &now

	s.reorderWaitingQueue(booking.ExamPlanID)

	return nil
}

func (s *Store) reorderWaitingQueue(examPlanID string) {
	var waitingBookings []*common.ExamBooking
	for _, booking := range s.examBookings {
		if booking.ExamPlanID == examPlanID && booking.IsWaiting {
			waitingBookings = append(waitingBookings, booking)
		}
	}

	sort.Slice(waitingBookings, func(i, j int) bool {
		return waitingBookings[i].WaitOrder < waitingBookings[j].WaitOrder
	})

	for i, booking := range waitingBookings {
		booking.WaitOrder = i + 1
	}
}

func (s *Store) CancelExamBooking(studentID, bookingID string) (needsApproval bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, exists := s.examBookings[bookingID]
	if !exists {
		return false, errors.New("预约不存在")
	}

	if booking.StudentID != studentID {
		return false, errors.New("无权取消他人预约")
	}

	if booking.Status == common.ExamBookingStatusCancelled ||
		booking.Status == common.ExamBookingStatusExpired {
		return false, errors.New("预约已取消或过期")
	}

	examPlan := s.examPlans[booking.ExamPlanID]
	timeUntilExam := time.Until(examPlan.Date)

	if timeUntilExam < CancellationApprovalThreshold {
		s.pendingCancellations[bookingID] = struct{}{}
		return true, nil
	}

	s.processCancellation(booking)

	return false, nil
}

func (s *Store) ApproveCancelExam(bookingID string, approved bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pendingCancellations[bookingID]; !exists {
		return errors.New("没有待审批的取消请求")
	}

	booking, exists := s.examBookings[bookingID]
	if !exists {
		return errors.New("预约不存在")
	}

	delete(s.pendingCancellations, bookingID)

	if !approved {
		return nil
	}

	s.processCancellation(booking)

	return nil
}

func (s *Store) processCancellation(booking *common.ExamBooking) {
	examPlan := s.examPlans[booking.ExamPlanID]

	booking.Status = common.ExamBookingStatusCancelled

	if booking.IsWaiting {
		s.reorderWaitingQueue(booking.ExamPlanID)
		return
	}

	examPlan.UsedQuota--

	nextWaiting := s.getNextWaitingBooking(booking.ExamPlanID)
	if nextWaiting != nil {
		nextWaiting.WaitOrder = 0
	}
}

func (s *Store) getNextWaitingBooking(examPlanID string) *common.ExamBooking {
	var next *common.ExamBooking
	for _, booking := range s.examBookings {
		if booking.ExamPlanID == examPlanID && booking.IsWaiting {
			if next == nil || booking.WaitOrder < next.WaitOrder {
				next = booking
			}
		}
	}
	return next
}

func (s *Store) ListExamBookings(studentID string) []*common.ExamBooking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bookings := make([]*common.ExamBooking, 0)
	for _, booking := range s.examBookings {
		if studentID == "" || booking.StudentID == studentID {
			bookings = append(bookings, booking)
		}
	}
	return bookings
}

func (s *Store) getNextSubjectUnlocked(studentID string) (common.Subject, error) {
	student, exists := s.students[studentID]
	if !exists {
		return 0, errors.New("学员不存在")
	}

	for subject := common.Subject1; subject <= common.Subject4; subject++ {
		if student.SubjectStatus[subject] != common.SubjectStatusPassed {
			return subject, nil
		}
	}

	return 0, errors.New("所有科目已通过")
}

func (s *Store) checkStudyHoursRequirementUnlocked(studentID string, subject common.Subject) error {
	student, exists := s.students[studentID]
	if !exists {
		return errors.New("学员不存在")
	}

	if student.VehicleType == common.VehicleTypeC1 {
		switch subject {
		case common.Subject2:
			if student.StudyHours[subject] < 24 {
				return errors.New("C1车型科目二需要至少24学时")
			}
		case common.Subject3:
			if student.StudyHours[subject] < 16 {
				return errors.New("C1车型科目三需要至少16学时")
			}
		}
	}

	return nil
}
