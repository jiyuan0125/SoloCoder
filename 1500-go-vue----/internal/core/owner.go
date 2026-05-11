package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OwnerService struct {
	store *Store
}

func NewOwnerService(store *Store) *OwnerService {
	return &OwnerService{store: store}
}

func (s *OwnerService) CreateOwner(name, phone, roomNumber string, area int64, status OwnerStatus) (*Owner, error) {
	if name == "" || phone == "" || roomNumber == "" {
		return nil, errors.New("name, phone, and room number are required")
	}
	if area <= 0 {
		return nil, errors.New("area must be positive")
	}
	if existing := s.store.GetOwnerByPhone(phone); existing != nil {
		return nil, errors.New("owner with this phone already exists")
	}

	owner := &Owner{
		ID:         uuid.NewString(),
		Name:       name,
		Phone:      phone,
		RoomNumber: roomNumber,
		Area:       area,
		Status:     status,
		VoteWeight: area,
	}
	s.store.AddOwner(owner)
	return owner, nil
}

func (s *OwnerService) GetOwner(id string) (*Owner, error) {
	owner := s.store.GetOwner(id)
	if owner == nil {
		return nil, errors.New("owner not found")
	}
	return owner, nil
}

func (s *OwnerService) ListOwners() []*Owner {
	return s.store.GetOwners()
}

func (s *OwnerService) CreateDelegate(principalID, agentID string) (*DelegateRelation, error) {
	if principalID == agentID {
		return nil, errors.New("cannot delegate to self")
	}
	if s.store.GetOwner(principalID) == nil {
		return nil, errors.New("principal owner not found")
	}
	if s.store.GetOwner(agentID) == nil {
		return nil, errors.New("agent owner not found")
	}

	delegate := &DelegateRelation{
		ID:          uuid.NewString(),
		PrincipalID: principalID,
		AgentID:     agentID,
		Status:      DelegateStatusPending,
		CreatedAt:   time.Now(),
	}
	s.store.AddDelegate(delegate)
	return delegate, nil
}

func (s *OwnerService) AcceptDelegate(delegateID, agentID string) (*DelegateRelation, error) {
	delegate := s.store.GetDelegate(delegateID)
	if delegate == nil {
		return nil, errors.New("delegate relation not found")
	}
	if delegate.AgentID != agentID {
		return nil, errors.New("not authorized to accept this delegate")
	}
	if delegate.Status != DelegateStatusPending {
		return nil, errors.New("delegate is not in pending status")
	}

	s.store.Lock()
	defer s.store.Unlock()

	if active := s.store.GetActiveDelegateForAgent(agentID); active != nil {
		return nil, errors.New("agent already has an active delegate")
	}

	now := time.Now()
	delegate.Status = DelegateStatusAccepted
	delegate.AcceptedAt = &now
	s.store.UpdateDelegate(delegate)

	return delegate, nil
}

func (s *OwnerService) RejectDelegate(delegateID, agentID string) (*DelegateRelation, error) {
	delegate := s.store.GetDelegate(delegateID)
	if delegate == nil {
		return nil, errors.New("delegate relation not found")
	}
	if delegate.AgentID != agentID {
		return nil, errors.New("not authorized to reject this delegate")
	}
	if delegate.Status != DelegateStatusPending {
		return nil, errors.New("delegate is not in pending status")
	}

	delegate.Status = DelegateStatusRejected
	s.store.UpdateDelegate(delegate)
	return delegate, nil
}

func (s *OwnerService) GetDelegates() []*DelegateRelation {
	return s.store.GetDelegates()
}

func (s *OwnerService) GetVoteWeight(ownerID string) (int64, error) {
	owner := s.store.GetOwner(ownerID)
	if owner == nil {
		return 0, errors.New("owner not found")
	}

	delegate := s.store.GetActiveDelegateForAgent(ownerID)
	if delegate != nil {
		principal := s.store.GetOwner(delegate.PrincipalID)
		if principal != nil {
			return principal.VoteWeight, nil
		}
	}

	return owner.VoteWeight, nil
}
