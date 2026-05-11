package core

import (
	"time"

	"community-activity-platform/pkg/common"
)

type CheckInService struct {
	storage *Storage
}

func NewCheckInService(storage *Storage) *CheckInService {
	return &CheckInService{storage: storage}
}

func (s *CheckInService) CheckIn(req common.CheckInRequest) (*common.Registration, error) {
	activity, ok := s.storage.GetActivity(req.ActivityID)
	if !ok {
		return nil, common.ErrActivityNotFound
	}

	if activity.Status != common.ActivityStatusActive {
		return nil, common.ErrActivityAlreadyCancelled
	}

	now := time.Now()
	windowStart := activity.StartTime.Add(-30 * time.Minute)
	windowEnd := activity.StartTime.Add(15 * time.Minute)

	if now.Before(windowStart) {
		return nil, common.ErrCheckInNotYetOpen
	}
	if now.After(windowEnd) {
		return nil, common.ErrCheckInWindowClosed
	}

	registrations := s.storage.ListRegistrationsByActivity(req.ActivityID, func(r *common.Registration) bool {
		return r.Status == common.RegistrationStatusRegistered
	})

	var targetReg *common.Registration
	for _, r := range registrations {
		participant, ok := s.storage.GetParticipant(r.ParticipantID)
		if !ok {
			continue
		}
		if participant.PhoneLast4 == req.PhoneLast4 {
			targetReg = r
			break
		}
	}

	if targetReg == nil {
		return nil, common.ErrPhoneMismatch
	}

	if targetReg.CheckInStatus == common.CheckInStatusSuccess || targetReg.CheckInStatus == common.CheckInStatusLate {
		return targetReg, nil
	}

	checkInTime := now
	if now.After(activity.StartTime) {
		targetReg.CheckInStatus = common.CheckInStatusLate
	} else {
		targetReg.CheckInStatus = common.CheckInStatusSuccess
	}
	targetReg.CheckInTime = &checkInTime
	targetReg.Status = common.RegistrationStatusCheckedIn
	targetReg.UpdateTime = now
	s.storage.StoreRegistration(targetReg)

	return targetReg, nil
}

func (s *CheckInService) MarkNoShows(activityID string) error {
	activity, ok := s.storage.GetActivity(activityID)
	if !ok {
		return common.ErrActivityNotFound
	}

	if !isActivityEnded(activity.EndTime) {
		return common.ErrActivityNotEnded
	}

	registrations := s.storage.ListRegistrationsByActivity(activityID, func(r *common.Registration) bool {
		return r.Status == common.RegistrationStatusRegistered
	})

	for _, r := range registrations {
		r.Status = common.RegistrationStatusNoShow
		r.CheckInStatus = common.CheckInStatusNotCheckedIn
		r.UpdateTime = time.Now()
		s.storage.StoreRegistration(r)
	}

	return nil
}

func (s *CheckInService) GetCheckInStatus(activityID, participantID string) (common.CheckInStatus, error) {
	if _, ok := s.storage.GetActivity(activityID); !ok {
		return "", common.ErrActivityNotFound
	}

	reg, ok := s.storage.GetRegistrationByParticipantActivity(participantID, activityID)
	if !ok {
		return "", common.ErrRegistrationNotFound
	}

	return reg.CheckInStatus, nil
}
