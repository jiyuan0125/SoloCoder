package core

import (
	"sync"
	"time"

	"community-activity-platform/pkg/common"
)

type RegistrationService struct {
	storage    *Storage
	activityMu sync.Map
}

func NewRegistrationService(storage *Storage) *RegistrationService {
	return &RegistrationService{
		storage: storage,
	}
}

func (s *RegistrationService) Register(req common.RegisterActivityRequest) (*common.Registration, error) {
	activity, ok := s.storage.GetActivity(req.ActivityID)
	if !ok {
		return nil, common.ErrActivityNotFound
	}

	if activity.Status != common.ActivityStatusActive {
		return nil, common.ErrActivityAlreadyCancelled
	}

	now := time.Now()
	if now.After(activity.RegistrationDeadline) {
		return nil, common.ErrRegistrationClosed
	}

	if now.After(activity.StartTime) {
		return nil, common.ErrActivityAlreadyStarted
	}

	participant, ok := s.storage.GetParticipant(req.ParticipantID)
	if !ok {
		return nil, common.ErrParticipantNotFound
	}

	if existing, ok := s.storage.GetRegistrationByParticipantActivity(req.ParticipantID, req.ActivityID); ok {
		if existing.Status == common.RegistrationStatusRegistered || existing.Status == common.RegistrationStatusWaitlist {
			return nil, common.ErrAlreadyRegistered
		}
	}

	if err := s.checkPrerequisites(activity, participant, req); err != nil {
		return nil, err
	}

	mu, _ := s.activityMu.LoadOrStore(req.ActivityID, &sync.Mutex{})
	activityMutex := mu.(*sync.Mutex)
	activityMutex.Lock()
	defer activityMutex.Unlock()

	registeredCount := s.storage.CountRegistrationsByActivity(req.ActivityID, common.RegistrationStatusRegistered)

	registration := &common.Registration{
		ID:            generateID(),
		ActivityID:    req.ActivityID,
		ParticipantID: req.ParticipantID,
		Status:        common.RegistrationStatusRegistered,
		CheckInStatus: common.CheckInStatusNotYet,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}

	if registeredCount >= activity.MaxParticipants {
		registration.Status = common.RegistrationStatusWaitlist
		s.storage.StoreRegistration(registration)
		s.storage.AddToWaitlist(req.ActivityID, registration.ID)
		return registration, nil
	}

	s.storage.StoreRegistration(registration)
	return registration, nil
}

func (s *RegistrationService) checkPrerequisites(activity *common.Activity, participant *common.Participant, req common.RegisterActivityRequest) error {
	switch activity.ActivityType {
	case common.ActivityTypeParentChild:
		if req.AdultsCount < 1 || req.ChildrenCount < 1 {
			return common.ErrPrerequisitesNotMet
		}
	case common.ActivityTypeVolunteer:
		if participant.Age < 18 {
			return common.ErrPrerequisitesNotMet
		}
	}
	return nil
}

func (s *RegistrationService) Cancel(registrationID string) error {
	registration, ok := s.storage.GetRegistration(registrationID)
	if !ok {
		return common.ErrRegistrationNotFound
	}

	if registration.Status != common.RegistrationStatusRegistered && registration.Status != common.RegistrationStatusWaitlist {
		return nil
	}

	activity, ok := s.storage.GetActivity(registration.ActivityID)
	if !ok {
		return common.ErrActivityNotFound
	}

	if !is24HoursBefore(activity.StartTime) {
		return common.ErrCancellationPeriodPassed
	}

	registration.Status = common.RegistrationStatusCancelled
	registration.UpdateTime = time.Now()
	s.storage.StoreRegistration(registration)

	if registration.Status == common.RegistrationStatusWaitlist {
		s.storage.RemoveFromWaitlist(activity.ID, registration.ID)
		return nil
	}

	waitlistID, ok := s.storage.PopWaitlist(activity.ID)
	if ok {
		if waitlistReg, ok := s.storage.GetRegistration(waitlistID); ok {
			waitlistReg.Status = common.RegistrationStatusRegistered
			waitlistReg.UpdateTime = time.Now()
			s.storage.StoreRegistration(waitlistReg)
		}
	}

	return nil
}

func (s *RegistrationService) GetByID(id string) (*common.Registration, error) {
	r, ok := s.storage.GetRegistration(id)
	if !ok {
		return nil, common.ErrRegistrationNotFound
	}
	return r, nil
}

func (s *RegistrationService) ListByActivity(activityID string) ([]*common.Registration, error) {
	if _, ok := s.storage.GetActivity(activityID); !ok {
		return nil, common.ErrActivityNotFound
	}

	return s.storage.ListRegistrationsByActivity(activityID, nil), nil
}
