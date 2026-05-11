package core

import (
	"fmt"
	"sync"
	"time"
)

type MilestoneService struct {
	store *Store
	mu    sync.Mutex
}

func NewMilestoneService(store *Store) *MilestoneService {
	return &MilestoneService{store: store}
}

type CompleteMilestoneInput struct {
	ContractID  string
	MilestoneID string
	ActualDate  time.Time
	Approved    bool
	Operator    string
}

func (s *MilestoneService) CompleteMilestone(input CompleteMilestoneInput) (*Milestone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	contract, err := s.store.GetContract(input.ContractID)
	if err != nil {
		return nil, err
	}

	if contract.Status != ContractStatusActive {
		return nil, ErrContractAlreadyCompleted
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

	if milestone.Status == MilestoneStatusCompleted || milestone.Status == MilestoneStatusApproved {
		return nil, ErrMilestoneAlreadyCompleted
	}

	oldStatus := milestone.Status

	isOffHours := false
	if contract.ServiceWindow != nil {
		isOffHours = !isWithinServiceWindow(input.ActualDate, contract.ServiceWindow)
	}

	actualDateCopy := input.ActualDate
	milestone.ActualDate = &actualDateCopy
	milestone.IsOffHours = isOffHours

	if input.Approved {
		milestone.Status = MilestoneStatusApproved
		milestone.Payment.Status = PaymentStatusDue
	} else {
		milestone.Status = MilestoneStatusCompleted
	}

	err = s.store.UpdateContract(contract)
	if err != nil {
		return nil, err
	}

	s.store.AddAuditLog(&AuditLog{
		ContractID: contract.ID,
		Operation:  "COMPLETE_MILESTONE",
		Operator:   input.Operator,
		Field:      fmt.Sprintf("milestone.%s.status", milestone.ID),
		OldValue:   string(oldStatus),
		NewValue:   string(milestone.Status),
	})

	if milestone.IsOffHours {
		s.store.AddAuditLog(&AuditLog{
			ContractID: contract.ID,
			Operation:  "MILESTONE_OFF_HOURS",
			Operator:   input.Operator,
			Field:      fmt.Sprintf("milestone.%s.is_off_hours", milestone.ID),
			OldValue:   "false",
			NewValue:   "true",
		})
	}

	s.checkContractCompletion(contract, input.Operator)

	return milestone, nil
}

func (s *MilestoneService) checkContractCompletion(contract *Contract, operator string) {
	allApproved := true
	for _, m := range contract.Milestones {
		if m.Status != MilestoneStatusApproved {
			allApproved = false
			break
		}
	}

	if allApproved {
		oldStatus := contract.Status
		contract.Status = ContractStatusCompleted
		s.store.UpdateContract(contract)

		s.store.AddAuditLog(&AuditLog{
			ContractID: contract.ID,
			Operation:  "COMPLETE_CONTRACT",
			Operator:   operator,
			Field:      "contract.status",
			OldValue:   string(oldStatus),
			NewValue:   string(ContractStatusCompleted),
		})
	}
}

func isWithinServiceWindow(t time.Time, window *ServiceWindow) bool {
	if window == nil {
		return true
	}

	weekday := int(t.Weekday())
	weekdayMatch := false
	for _, wd := range window.Weekdays {
		if wd == weekday {
			weekdayMatch = true
			break
		}
	}
	if !weekdayMatch {
		return false
	}

	startTime, err := parseTimeString(window.StartTime)
	if err != nil {
		return true
	}
	endTime, err := parseTimeString(window.EndTime)
	if err != nil {
		return true
	}

	currentMinutes := t.Hour()*60 + t.Minute()
	startMinutes := startTime.Hour*60 + startTime.Minute
	endMinutes := endTime.Hour*60 + endTime.Minute

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

type timeParts struct {
	Hour   int
	Minute int
}

func parseTimeString(s string) (*timeParts, error) {
	var h, m int
	_, err := fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil {
		return nil, err
	}
	return &timeParts{Hour: h, Minute: m}, nil
}

func (s *MilestoneService) GetMilestone(contractID, milestoneID string) (*Milestone, error) {
	return s.store.GetMilestone(contractID, milestoneID)
}

func (s *MilestoneService) GetProgress(contractID string) (int, error) {
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return 0, err
	}

	completedPct := 0
	for _, m := range contract.Milestones {
		if m.Status == MilestoneStatusApproved {
			completedPct += m.PlannedPercentage
		}
	}
	return completedPct, nil
}

func (s *MilestoneService) GetProgressDetails(contractID string) (map[string]interface{}, error) {
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return nil, err
	}

	completedPct := 0
	completedCount := 0
	totalCount := len(contract.Milestones)
	offHoursCount := 0

	for _, m := range contract.Milestones {
		if m.Status == MilestoneStatusApproved {
			completedPct += m.PlannedPercentage
			completedCount++
		}
		if m.IsOffHours {
			offHoursCount++
		}
	}

	return map[string]interface{}{
		"contract_id":       contractID,
		"progress_percent":  completedPct,
		"completed_count":   completedCount,
		"total_milestones":  totalCount,
		"off_hours_count":   offHoursCount,
		"contract_status":   string(contract.Status),
	}, nil
}
