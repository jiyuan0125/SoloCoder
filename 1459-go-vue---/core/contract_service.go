package core

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ContractService struct {
	store *Store
	mu    sync.Mutex
}

func NewContractService(store *Store) *ContractService {
	return &ContractService{store: store}
}

type CreateMilestoneInput struct {
	Name              string
	PlannedDate       time.Time
	PlannedPercentage int
	Owner             string
}

type CreateContractInput struct {
	ContractNo    string
	PartyA        string
	PartyB        string
	TotalAmount   int64
	SignDate      time.Time
	ContractType  string
	Milestones    []CreateMilestoneInput
	ServiceWindow *ServiceWindow
	Operator      string
}

func (s *ContractService) CreateContract(input CreateContractInput) (*Contract, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if input.TotalAmount < 0 {
		return nil, ErrInvalidAmount
	}

	totalPct := 0
	for _, m := range input.Milestones {
		if m.PlannedPercentage <= 0 || m.PlannedPercentage > 100 {
			return nil, ErrInvalidPercentage
		}
		totalPct += m.PlannedPercentage
	}
	if totalPct != 100 {
		return nil, ErrMilestonePercentageSum
	}

	contract := &Contract{
		ContractNo:   input.ContractNo,
		PartyA:       input.PartyA,
		PartyB:       input.PartyB,
		TotalAmount:  input.TotalAmount,
		SignDate:     input.SignDate,
		ContractType: input.ContractType,
		Status:       ContractStatusActive,
		ServiceWindow: input.ServiceWindow,
	}

	for _, m := range input.Milestones {
		plannedAmount := calculateAmount(input.TotalAmount, m.PlannedPercentage)
		milestone := &Milestone{
			ID:                uuid.New().String(),
			ContractID:        "",
			Name:              m.Name,
			PlannedDate:       m.PlannedDate,
			PlannedPercentage: m.PlannedPercentage,
			PlannedAmount:     plannedAmount,
			Owner:             m.Owner,
			Status:            MilestoneStatusPending,
			IsOffHours:        false,
			Payment: &Payment{
				ID:          uuid.New().String(),
				MilestoneID: "",
				ContractID:  "",
				DueAmount:   plannedAmount,
				PaidAmount:  nil,
				PaidDate:    nil,
				Status:      PaymentStatusPending,
			},
		}
		milestone.ContractID = contract.ID
		milestone.Payment.MilestoneID = milestone.ID
		milestone.Payment.ContractID = contract.ID
		contract.Milestones = append(contract.Milestones, milestone)
	}

	created, err := s.store.CreateContract(contract)
	if err != nil {
		return nil, err
	}

	for _, m := range created.Milestones {
		m.ContractID = created.ID
		m.Payment.ContractID = created.ID
		m.Payment.MilestoneID = m.ID
	}
	s.store.UpdateContract(created)

	s.store.AddAuditLog(&AuditLog{
		ContractID: created.ID,
		Operation:  "CREATE_CONTRACT",
		Operator:   input.Operator,
		Field:      "contract",
		OldValue:   "",
		NewValue:   fmt.Sprintf("Created contract %s", created.ContractNo),
	})

	return created, nil
}

func (s *ContractService) GetContract(id string) (*Contract, error) {
	return s.store.GetContract(id)
}

func (s *ContractService) ListContracts() []*Contract {
	return s.store.ListContracts()
}

type UpdateContractAmountInput struct {
	ContractID      string
	NewTotalAmount  int64
	Operator        string
}

func (s *ContractService) UpdateContractAmount(input UpdateContractAmountInput) (*Contract, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	contract, err := s.store.GetContract(input.ContractID)
	if err != nil {
		return nil, err
	}

	if contract.Status != ContractStatusActive {
		return nil, ErrContractAlreadyCompleted
	}

	if input.NewTotalAmount < 0 {
		return nil, ErrInvalidAmount
	}

	oldAmount := contract.TotalAmount
	if oldAmount == input.NewTotalAmount {
		return contract, nil
	}

	completedPlannedPct := 0
	completedDueTotal := int64(0)
	for _, m := range contract.Milestones {
		if m.Status == MilestoneStatusApproved || m.Status == MilestoneStatusCompleted {
			completedPlannedPct += m.PlannedPercentage
			completedDueTotal += m.Payment.DueAmount
		}
	}

	remainingPct := 100 - completedPlannedPct
	if remainingPct < 0 {
		remainingPct = 0
	}

	for _, m := range contract.Milestones {
		if m.Status == MilestoneStatusPending || m.Status == MilestoneStatusInProgress {
			if remainingPct > 0 {
				newDueAmount := calculateAmount(input.NewTotalAmount, m.PlannedPercentage)
				m.PlannedAmount = newDueAmount
				if m.Payment.Status == PaymentStatusPending {
					m.Payment.DueAmount = newDueAmount
				}
			}
		}
	}

	contract.TotalAmount = input.NewTotalAmount
	contract.UpdatedAt = time.Now()

	err = s.store.UpdateContract(contract)
	if err != nil {
		return nil, err
	}

	s.store.AddAuditLog(&AuditLog{
		ContractID: contract.ID,
		Operation:  "UPDATE_TOTAL_AMOUNT",
		Operator:   input.Operator,
		Field:      "total_amount",
		OldValue:   strconv.FormatInt(oldAmount, 10),
		NewValue:   strconv.FormatInt(input.NewTotalAmount, 10),
	})

	return contract, nil
}

func calculateAmount(total int64, percentage int) int64 {
	return (total * int64(percentage)) / 100
}
