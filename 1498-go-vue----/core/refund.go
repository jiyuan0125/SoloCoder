package core

import (
	"errors"
	"fmt"
	"time"

	"housekeeping-training/common"
)

func (s *Store) RequestRefund(enrollmentID, reason string) (*common.RefundRequest, error) {
	if enrollmentID == "" {
		return nil, errors.New("enrollment ID is required")
	}

	s.Lock()
	defer s.Unlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return nil, errors.New("enrollment not found")
	}

	if s.HasTakenExam(enrollmentID) {
		return nil, errors.New("cannot refund after exam")
	}

	for _, r := range s.refunds {
		if r.EnrollmentID == enrollmentID {
			return nil, errors.New("refund request already exists")
		}
	}

	schedule, ok := s.schedules[enrollment.ScheduleID]
	if !ok {
		return nil, errors.New("schedule not found")
	}

	classesTaken := s.countClassesTaken(enrollmentID)
	totalClasses := schedule.TotalClasses
	if totalClasses == 0 {
		totalClasses = 1
	}

	var totalPaid int
	if enrollment.PaidFirst {
		totalPaid += enrollment.FirstPayment
	}
	if enrollment.PaidSecond {
		totalPaid += enrollment.SecondPayment
	}

	percentageCompleted := float64(classesTaken) / float64(totalClasses)
	usedAmount := int(float64(totalPaid) * percentageCompleted)
	refundAmount := totalPaid - usedAmount

	if refundAmount < 0 {
		refundAmount = 0
	}

	s.nextRefundID++
	refund := &common.RefundRequest{
		ID:           fmt.Sprintf("RF%04d", s.nextRefundID),
		EnrollmentID: enrollmentID,
		RequestDate:  time.Now(),
		RefundAmount: refundAmount,
		Reason:       reason,
	}
	s.refunds[refund.ID] = refund

	delete(s.enrollments, enrollmentID)
	if schedule.Enrolled > 0 {
		schedule.Enrolled--
	}

	return refund, nil
}

func (s *Store) GetRefund(id string) (*common.RefundRequest, error) {
	s.RLock()
	defer s.RUnlock()

	refund, ok := s.refunds[id]
	if !ok {
		return nil, errors.New("refund request not found")
	}
	return refund, nil
}

func (s *Store) ListRefunds() []*common.RefundRequest {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.RefundRequest, 0, len(s.refunds))
	for _, r := range s.refunds {
		result = append(result, r)
	}
	return result
}
