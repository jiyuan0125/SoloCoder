package core

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type PaymentService struct {
	store *Store
	mu    sync.Mutex
}

func NewPaymentService(store *Store) *PaymentService {
	return &PaymentService{store: store}
}

type PayMilestoneInput struct {
	ContractID  string
	MilestoneID string
	PaidAmount  int64
	Operator    string
}

func (s *PaymentService) PayMilestone(input PayMilestoneInput) (*Payment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	contract, err := s.store.GetContract(input.ContractID)
	if err != nil {
		return nil, err
	}

	var milestone *Milestone
	for _, m := range contract.Milestones {
		if m.ID == input.MilestoneID {
			milestone = m
			break
		}
	}
	if milestone == nil {
		return nil, ErrMilestoneNotFound
	}

	if milestone.Status != MilestoneStatusApproved {
		return nil, ErrMilestoneNotApproved
	}

	if milestone.Payment.Status == PaymentStatusPaid {
		return nil, ErrPaymentAlreadyPaid
	}

	if input.PaidAmount < 0 {
		return nil, ErrInvalidAmount
	}

	oldStatus := milestone.Payment.Status
	oldPaidAmount := milestone.Payment.PaidAmount

	paidDate := time.Now()
	paidAmount := input.PaidAmount
	milestone.Payment.PaidAmount = &paidAmount
	milestone.Payment.PaidDate = &paidDate
	milestone.Payment.Status = PaymentStatusPaid

	adjustment := paidAmount - milestone.Payment.DueAmount
	if adjustment != 0 {
		s.store.AddAdjustment(&ContractAdjustment{
			ContractID:       contract.ID,
			PaymentID:        milestone.Payment.ID,
			AdjustmentAmount: adjustment,
			Reason:           fmt.Sprintf("Payment difference: paid=%d, due=%d", paidAmount, milestone.Payment.DueAmount),
		})

		s.store.AddAuditLog(&AuditLog{
			ContractID: contract.ID,
			Operation:  "PAYMENT_ADJUSTMENT",
			Operator:   input.Operator,
			Field:      fmt.Sprintf("payment.%s.adjustment", milestone.Payment.ID),
			OldValue:   "0",
			NewValue:   strconv.FormatInt(adjustment, 10),
		})
	}

	err = s.store.UpdateContract(contract)
	if err != nil {
		return nil, err
	}

	oldPaidStr := ""
	if oldPaidAmount != nil {
		oldPaidStr = strconv.FormatInt(*oldPaidAmount, 10)
	}

	s.store.AddAuditLog(&AuditLog{
		ContractID: contract.ID,
		Operation:  "PAY_MILESTONE",
		Operator:   input.Operator,
		Field:      fmt.Sprintf("payment.%s.status", milestone.Payment.ID),
		OldValue:   string(oldStatus),
		NewValue:   string(PaymentStatusPaid),
	})

	s.store.AddAuditLog(&AuditLog{
		ContractID: contract.ID,
		Operation:  "PAY_MILESTONE_AMOUNT",
		Operator:   input.Operator,
		Field:      fmt.Sprintf("payment.%s.paid_amount", milestone.Payment.ID),
		OldValue:   oldPaidStr,
		NewValue:   strconv.FormatInt(paidAmount, 10),
	})

	return milestone.Payment, nil
}

func (s *PaymentService) GetPayment(contractID, milestoneID string) (*Payment, error) {
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return nil, err
	}

	for _, m := range contract.Milestones {
		if m.ID == milestoneID {
			return m.Payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (s *PaymentService) GetContractPayments(contractID string) ([]*Payment, error) {
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return nil, err
	}

	payments := []*Payment{}
	for _, m := range contract.Milestones {
		payments = append(payments, m.Payment)
	}
	return payments, nil
}

func (s *PaymentService) GetPaymentSummary(contractID string) (map[string]interface{}, error) {
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return nil, err
	}

	totalDue := int64(0)
	totalPaid := int64(0)
	paidCount := 0
	pendingCount := 0

	for _, m := range contract.Milestones {
		totalDue += m.Payment.DueAmount
		if m.Payment.Status == PaymentStatusPaid && m.Payment.PaidAmount != nil {
			totalPaid += *m.Payment.PaidAmount
			paidCount++
		} else {
			pendingCount++
		}
	}

	return map[string]interface{}{
		"contract_id":    contractID,
		"total_due":      totalDue,
		"total_paid":     totalPaid,
		"outstanding":    totalDue - totalPaid,
		"paid_count":     paidCount,
		"pending_count":  pendingCount,
	}, nil
}
