package core

import (
	"strings"
	"time"

	"community-activity-platform/pkg/common"
)

type ActivityService struct {
	storage *Storage
}

func NewActivityService(storage *Storage) *ActivityService {
	return &ActivityService{storage: storage}
}

func (s *ActivityService) Create(req common.CreateActivityRequest) ([]*common.Activity, error) {
	activities := []*common.Activity{}

	now := time.Now()
	seriesID := ""

	if req.Frequency == common.ActivityFrequencyRecurring {
		seriesID = generateID()
	}

	if req.Frequency == common.ActivityFrequencyRecurring {
		episodeCount := 1
		if req.EpisodeCount > 1 {
			episodeCount = req.EpisodeCount
		}

		for i := 1; i <= episodeCount; i++ {
			activity := &common.Activity{
				ID:                generateID(),
				Title:             req.Title,
				Description:       req.Description,
				StartTime:         req.StartTime.AddDate(0, 0, (i-1)*7),
				EndTime:           req.EndTime.AddDate(0, 0, (i-1)*7),
				Location:          req.Location,
				MaxParticipants:   req.MaxParticipants,
				RegistrationDeadline: req.RegistrationDeadline.AddDate(0, 0, (i-1)*7),
				ActivityType:      req.ActivityType,
				Tags:              req.Tags,
				Frequency:         req.Frequency,
				RecurringRule:     req.RecurringRule,
				SeriesID:          seriesID,
				EpisodeNumber:     i,
				Status:            common.ActivityStatusActive,
				CreateTime:        now,
				UpdateTime:        now,
			}

			s.storage.StoreActivity(activity)
			activities = append(activities, activity)
		}
	} else {
		activity := &common.Activity{
			ID:                generateID(),
			Title:             req.Title,
			Description:       req.Description,
			StartTime:         req.StartTime,
			EndTime:           req.EndTime,
			Location:          req.Location,
			MaxParticipants:   req.MaxParticipants,
			RegistrationDeadline: req.RegistrationDeadline,
			ActivityType:      req.ActivityType,
			Tags:              req.Tags,
			Frequency:         req.Frequency,
			Status:            common.ActivityStatusActive,
			CreateTime:        now,
			UpdateTime:        now,
		}

		s.storage.StoreActivity(activity)
		activities = append(activities, activity)
	}

	return activities, nil
}

func (s *ActivityService) GetByID(id string) (*common.Activity, error) {
	a, ok := s.storage.GetActivity(id)
	if !ok {
		return nil, common.ErrActivityNotFound
	}
	return a, nil
}

func (s *ActivityService) GetDetail(id string) (*common.ActivityDetailResponse, error) {
	a, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	registeredCount := s.storage.CountRegistrationsByActivity(id, common.RegistrationStatusRegistered)
	waitlistCount := s.storage.CountRegistrationsByActivity(id, common.RegistrationStatusWaitlist)
	availableSlots := a.MaxParticipants - registeredCount
	if availableSlots < 0 {
		availableSlots = 0
	}

	return &common.ActivityDetailResponse{
		Activity:        *a,
		RegisteredCount: registeredCount,
		WaitlistCount:   waitlistCount,
		AvailableSlots:  availableSlots,
	}, nil
}

func (s *ActivityService) List(req common.ListActivitiesRequest) ([]*common.ActivityDetailResponse, error) {
	activities := s.storage.ListActivities(func(a *common.Activity) bool {
		if req.Status != nil && a.Status != *req.Status {
			return false
		}
		if req.Type != nil && a.ActivityType != *req.Type {
			return false
		}
		return true
	})

	results := []*common.ActivityDetailResponse{}
	for _, a := range activities {
		detail, err := s.GetDetail(a.ID)
		if err == nil {
			results = append(results, detail)
		}
	}

	return results, nil
}

func (s *ActivityService) Update(id string, req common.UpdateActivityRequest) (*common.Activity, error) {
	a, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if a.Status != common.ActivityStatusActive {
		return nil, common.ErrActivityAlreadyCancelled
	}

	if req.Title != nil {
		a.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		a.Description = strings.TrimSpace(*req.Description)
	}
	if req.StartTime != nil {
		a.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		a.EndTime = *req.EndTime
	}
	if req.Location != nil {
		a.Location = strings.TrimSpace(*req.Location)
	}
	if req.MaxParticipants != nil {
		a.MaxParticipants = *req.MaxParticipants
	}
	if req.RegistrationDeadline != nil {
		a.RegistrationDeadline = *req.RegistrationDeadline
	}
	if req.ActivityType != nil {
		a.ActivityType = *req.ActivityType
	}
	if req.Tags != nil {
		a.Tags = *req.Tags
	}

	a.UpdateTime = time.Now()
	s.storage.StoreActivity(a)

	return a, nil
}

func (s *ActivityService) Cancel(id string) (*common.Activity, error) {
	a, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if a.Status == common.ActivityStatusCancelled {
		return nil, common.ErrActivityAlreadyCancelled
	}

	if a.Status == common.ActivityStatusCompleted {
		return nil, common.ErrActivityAlreadyCompleted
	}

	a.Status = common.ActivityStatusCancelled
	a.UpdateTime = time.Now()
	s.storage.StoreActivity(a)

	registrations := s.storage.ListRegistrationsByActivity(id, func(r *common.Registration) bool {
		return r.Status == common.RegistrationStatusRegistered || r.Status == common.RegistrationStatusWaitlist
	})

	for _, r := range registrations {
		r.Status = common.RegistrationStatusCancelled
		r.UpdateTime = time.Now()
		s.storage.StoreRegistration(r)
	}

	s.storage.ClearWaitlist(id)

	return a, nil
}
