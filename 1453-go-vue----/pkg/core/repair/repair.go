package repair

import (
	"errors"
	"fmt"
	"property-management/pkg/core"
	"property-management/pkg/core/models"
	"time"

	"github.com/google/uuid"
)

const (
	ConfirmDeadlineHours = 24
	ProcessDeadlineHours = 48
)

type Service struct {
	store *core.Store
}

func NewService(store *core.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ownerID string, repairType models.RepairType, description string, expectedTime time.Time) (*models.Repair, error) {
	s.store.Lock()
	defer s.store.Unlock()

	owner, ok := s.store.Users()[ownerID]
	if !ok {
		return nil, errors.New("owner not found")
	}

	now := time.Now()
	repair := &models.Repair{
		ID:           uuid.New().String(),
		OwnerID:      ownerID,
		OwnerName:    owner.Name,
		Building:     owner.Building,
		Unit:         owner.Unit,
		RepairType:   repairType,
		Description:  description,
		ExpectedTime: expectedTime,
		Status:       models.RepairStatusPendingDispatch,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.store.SetRepair(repair.ID, repair)
	return repair, nil
}

func (s *Service) Assign(repairID string, supervisorID string, staffID string) (*models.Repair, error) {
	s.store.Lock()
	defer s.store.Unlock()

	supervisor, ok := s.store.Users()[supervisorID]
	if !ok {
		return nil, errors.New("supervisor not found")
	}
	if supervisor.Role != models.RoleSupervisor {
		return nil, errors.New("only supervisor can assign repairs")
	}

	staff, ok := s.store.Users()[staffID]
	if !ok {
		return nil, errors.New("staff not found")
	}
	if staff.Role != models.RoleStaff {
		return nil, errors.New("assigned user must be staff")
	}

	repair, ok := s.store.Repairs()[repairID]
	if !ok {
		return nil, errors.New("repair not found")
	}

	if repair.Status != models.RepairStatusPendingDispatch && repair.Status != models.RepairStatusEscalated {
		return nil, errors.New("repair is not pending dispatch")
	}

	now := time.Now()
	repair.Status = models.RepairStatusAssigned
	repair.AssignedStaffID = staffID
	repair.AssignedStaffName = staff.Name
	repair.AssignedAt = now
	repair.UpdatedAt = now

	s.store.SetRepair(repair.ID, repair)
	return repair, nil
}

func (s *Service) StartProcess(repairID string, staffID string) (*models.Repair, error) {
	s.store.Lock()
	defer s.store.Unlock()

	repair, ok := s.store.Repairs()[repairID]
	if !ok {
		return nil, errors.New("repair not found")
	}

	if repair.AssignedStaffID != staffID {
		return nil, errors.New("you are not assigned to this repair")
	}

	if repair.Status != models.RepairStatusAssigned {
		return nil, errors.New("repair is not assigned")
	}

	repair.Status = models.RepairStatusProcessing
	repair.UpdatedAt = time.Now()

	s.store.SetRepair(repair.ID, repair)
	return repair, nil
}

func (s *Service) Complete(repairID string, staffID string, result string, imageURLs []string) (*models.Repair, error) {
	s.store.Lock()
	defer s.store.Unlock()

	repair, ok := s.store.Repairs()[repairID]
	if !ok {
		return nil, errors.New("repair not found")
	}

	if repair.AssignedStaffID != staffID {
		return nil, errors.New("you are not assigned to this repair")
	}

	if repair.Status != models.RepairStatusProcessing {
		return nil, errors.New("repair is not in processing status")
	}

	now := time.Now()
	repair.Status = models.RepairStatusCompleted
	repair.Result = result
	repair.ResultImageURLs = imageURLs
	repair.CompletedAt = now
	repair.UpdatedAt = now

	s.store.SetRepair(repair.ID, repair)
	return repair, nil
}

func (s *Service) Confirm(repairID string, ownerID string) (*models.Repair, error) {
	s.store.Lock()
	defer s.store.Unlock()

	repair, ok := s.store.Repairs()[repairID]
	if !ok {
		return nil, errors.New("repair not found")
	}

	if repair.OwnerID != ownerID {
		return nil, errors.New("you are not the owner of this repair")
	}

	if repair.Status != models.RepairStatusCompleted {
		return nil, errors.New("repair is not completed")
	}

	now := time.Now()
	repair.Status = models.RepairStatusClosed
	repair.ConfirmedAt = now
	repair.ClosedAt = now
	repair.UpdatedAt = now
	repair.IsAutoConfirmed = false

	s.store.SetRepair(repair.ID, repair)
	return repair, nil
}

func (s *Service) CheckAndAutoConfirm() {
	s.store.Lock()
	defer s.store.Unlock()

	now := time.Now()
	for _, repair := range s.store.Repairs() {
		if repair.Status == models.RepairStatusCompleted && !repair.CompletedAt.IsZero() {
			elapsed := now.Sub(repair.CompletedAt)
			if elapsed.Hours() >= ConfirmDeadlineHours {
				repair.Status = models.RepairStatusClosed
				repair.ConfirmedAt = now
				repair.ClosedAt = now
				repair.UpdatedAt = now
				repair.IsAutoConfirmed = true
				s.store.SetRepair(repair.ID, repair)
			}
		}
	}
}

func (s *Service) CheckAndEscalate() {
	s.store.Lock()
	defer s.store.Unlock()

	now := time.Now()
	for _, repair := range s.store.Repairs() {
		if (repair.Status == models.RepairStatusAssigned || repair.Status == models.RepairStatusProcessing) && !repair.AssignedAt.IsZero() {
			elapsed := now.Sub(repair.AssignedAt)
			if elapsed.Hours() >= ProcessDeadlineHours && repair.Status != models.RepairStatusEscalated {
				repair.Status = models.RepairStatusEscalated
				repair.UpdatedAt = now
				s.store.SetRepair(repair.ID, repair)
			}
		}
	}
}

func (s *Service) GetByID(id string) (*models.Repair, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	repair, ok := s.store.Repairs()[id]
	if !ok {
		return nil, errors.New("repair not found")
	}
	return repair, nil
}

func (s *Service) ListByOwner(ownerID string) []*models.Repair {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Repair
	for _, repair := range s.store.Repairs() {
		if repair.OwnerID == ownerID {
			result = append(result, repair)
		}
	}
	return result
}

func (s *Service) ListByStatus(status models.RepairStatus) []*models.Repair {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Repair
	for _, repair := range s.store.Repairs() {
		if repair.Status == status {
			result = append(result, repair)
		}
	}
	return result
}

func (s *Service) ListAll() []*models.Repair {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Repair
	for _, repair := range s.store.Repairs() {
		result = append(result, repair)
	}
	return result
}

func (s *Service) IsPublicFacility(repairType models.RepairType) bool {
	return repairType == models.RepairTypePublicFacility
}

func (s *Service) GenerateRepairBillID(repairID string) string {
	return fmt.Sprintf("R-%s", repairID)
}
