package core

import (
	"time"

	"community-activity-platform/pkg/common"
)

type FeedbackService struct {
	storage *Storage
}

func NewFeedbackService(storage *Storage) *FeedbackService {
	return &FeedbackService{storage: storage}
}

func (s *FeedbackService) Submit(req common.SubmitFeedbackRequest) (*common.Feedback, error) {
	activity, ok := s.storage.GetActivity(req.ActivityID)
	if !ok {
		return nil, common.ErrActivityNotFound
	}

	if !isActivityEnded(activity.EndTime) {
		return nil, common.ErrActivityNotEnded
	}

	if isFeedbackPeriodEnded(activity.EndTime) {
		return nil, common.ErrFeedbackPeriodPassed
	}

	if _, ok := s.storage.GetFeedbackByParticipantActivity(req.ParticipantID, req.ActivityID); ok {
		return nil, common.ErrAlreadySubmittedFeedback
	}

	reg, ok := s.storage.GetRegistrationByParticipantActivity(req.ParticipantID, req.ActivityID)
	if !ok {
		return nil, common.ErrRegistrationNotFound
	}

	if reg.CheckInStatus != common.CheckInStatusSuccess && reg.CheckInStatus != common.CheckInStatusLate {
		return nil, common.ErrPrerequisitesNotMet
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, common.ErrInvalidRating
	}

	feedback := &common.Feedback{
		ID:            generateID(),
		ActivityID:    req.ActivityID,
		ParticipantID: req.ParticipantID,
		Rating:        req.Rating,
		Comments:      req.Comments,
		CreateTime:    time.Now(),
	}

	s.storage.StoreFeedback(feedback)
	return feedback, nil
}

func (s *FeedbackService) SendFeedbackNotifications(activityID string) error {
	activity, ok := s.storage.GetActivity(activityID)
	if !ok {
		return common.ErrActivityNotFound
	}

	if !isActivityEnded(activity.EndTime) {
		return common.ErrActivityNotEnded
	}

	registrations := s.storage.ListRegistrationsByActivity(activityID, func(r *common.Registration) bool {
		return r.CheckInStatus == common.CheckInStatusSuccess || r.CheckInStatus == common.CheckInStatusLate
	})

	for _, r := range registrations {
		participant, ok := s.storage.GetParticipant(r.ParticipantID)
		if !ok {
			continue
		}
		s.sendNotification(participant, activity)
	}

	return nil
}

func (s *FeedbackService) sendNotification(participant *common.Participant, activity *common.Activity) {
}

func (s *FeedbackService) GenerateSummary(activityID string) (*common.ActivitySummary, error) {
	activity, ok := s.storage.GetActivity(activityID)
	if !ok {
		return nil, common.ErrActivityNotFound
	}

	if !isActivityEnded(activity.EndTime) {
		return nil, common.ErrActivityNotEnded
	}

	feedbackPeriodEnd := activity.EndTime.Add(72 * time.Hour)
	now := time.Now()
	if now.Before(feedbackPeriodEnd) {
		return nil, common.ErrActivityNotEnded
	}

	registrations := s.storage.ListRegistrationsByActivity(activityID, nil)
	feedbacks := s.storage.ListFeedbacksByActivity(activityID)

	totalRegistrations := 0
	checkInCount := 0
	totalRating := 0

	for _, r := range registrations {
		if r.Status == common.RegistrationStatusRegistered || r.Status == common.RegistrationStatusCheckedIn || r.Status == common.RegistrationStatusNoShow {
			totalRegistrations++
		}
		if r.CheckInStatus == common.CheckInStatusSuccess || r.CheckInStatus == common.CheckInStatusLate {
			checkInCount++
		}
	}

	for _, f := range feedbacks {
		totalRating += f.Rating
	}

	checkInRate := 0.0
	if totalRegistrations > 0 {
		checkInRate = float64(checkInCount) / float64(totalRegistrations)
	}

	averageRating := 0.0
	if len(feedbacks) > 0 {
		averageRating = float64(totalRating) / float64(len(feedbacks))
	}

	summary := &common.ActivitySummary{
		ActivityID:       activityID,
		TotalRegistrations: totalRegistrations,
		CheckInCount:     checkInCount,
		CheckInRate:    checkInRate,
		AverageRating: averageRating,
		FeedbackCount:  len(feedbacks),
		GenerateTime:  time.Now(),
	}

	s.storage.StoreActivitySummary(summary)

	if checkInRate < 0.5 {
		todo := &common.AnalysisTodo{
			ID:         generateID(),
			ActivityID:   activityID,
			CheckInRate:  checkInRate,
			CreateTime:   time.Now(),
			Resolved:     false,
		}
		s.storage.StoreAnalysisTodo(todo)
	}

	activity.Status = common.ActivityStatusCompleted
	activity.UpdateTime = time.Now()
	s.storage.StoreActivity(activity)

	return summary, nil
}

func (s *FeedbackService) GetSummary(activityID string) (*common.ActivitySummary, error) {
	summary, ok := s.storage.GetActivitySummary(activityID)
	if !ok {
		return nil, common.ErrActivityNotFound
	}
	return summary, nil
}

func (s *FeedbackService) ListAnalysisTodos() ([]*common.AnalysisTodo, error) {
	return s.storage.ListAnalysisTodos(), nil
}
