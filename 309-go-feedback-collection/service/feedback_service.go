package service

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"feedback-collection/dao"
	"feedback-collection/models"
)

const (
	MaxDescriptionLength = 500
	MaxNoteLength        = 500
	MaxDailyFeedbacks    = 5
	MinRating            = 1
	MaxRating            = 5
	DefaultPageSize      = 20
	MaxPageSize          = 100
)

var (
	ErrInvalidRating       = errors.New("invalid rating: must be between 1 and 5")
	ErrInvalidFeedbackType = errors.New("invalid feedback type: must be one of [feature, bug, complaint]")
	ErrInvalidStatus       = errors.New("invalid status: must be one of [pending, processing, closed]")
	ErrDescriptionTooLong  = errors.New("description too long: maximum 500 characters")
	ErrNoteTooLong         = errors.New("note too long: maximum 500 characters")
	ErrDailyLimitExceeded  = errors.New("daily feedback limit exceeded: maximum 5 per day")
	ErrFeedbackNotFound    = errors.New("feedback not found")
	ErrProcessingNoteRequired = errors.New("processing note required when closing a bug report")
	ErrInvalidPage         = errors.New("invalid page: must be greater than 0")
	ErrInvalidPageSize     = errors.New("invalid page size: must be between 1 and 100")
	ErrUserIDRequired      = errors.New("user_id is required")
)

func isValidFeedbackType(feedbackType string) bool {
	for _, t := range models.ValidFeedbackTypes {
		if string(t) == feedbackType {
			return true
		}
	}
	return false
}

func isValidStatus(status string) bool {
	for _, s := range models.ValidFeedbackStatuses {
		if string(s) == status {
			return true
		}
	}
	return false
}

func parseRating(rating interface{}) (int, error) {
	switch v := rating.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if v == float64(int(v)) {
			return int(v), nil
		}
		return 0, ErrInvalidRating
	case string:
		ratingInt, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, ErrInvalidRating
		}
		return ratingInt, nil
	default:
		return 0, ErrInvalidRating
	}
}

func CreateFeedback(req *models.CreateFeedbackRequest) (*models.Feedback, error) {
	if req.UserID == "" {
		return nil, ErrUserIDRequired
	}

	if !isValidFeedbackType(req.FeedbackType) {
		return nil, ErrInvalidFeedbackType
	}

	rating, err := parseRating(req.Rating)
	if err != nil {
		return nil, err
	}

	if rating < MinRating || rating > MaxRating {
		return nil, ErrInvalidRating
	}

	if utf8.RuneCountInString(req.Description) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	count, err := dao.CountUserFeedbacksToday(req.UserID)
	if err != nil {
		return nil, err
	}

	if count >= MaxDailyFeedbacks {
		return nil, ErrDailyLimitExceeded
	}

	feedback := &models.Feedback{
		UserID:       req.UserID,
		FeedbackType: models.FeedbackType(req.FeedbackType),
		Rating:       rating,
		Description:  req.Description,
		Status:       models.StatusPending,
	}

	err = dao.CreateFeedback(feedback)
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

func GetFeedbackList(feedbackType, status string, page, pageSize int) ([]*models.Feedback, *models.Pagination, error) {
	if page <= 0 {
		return nil, nil, ErrInvalidPage
	}

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	if feedbackType != "" && !isValidFeedbackType(feedbackType) {
		return nil, nil, ErrInvalidFeedbackType
	}

	if status != "" && !isValidStatus(status) {
		return nil, nil, ErrInvalidStatus
	}

	feedbacks, total, err := dao.GetFeedbackList(feedbackType, status, page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	pagination := &models.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	return feedbacks, pagination, nil
}

func UpdateFeedbackStatus(feedbackID int64, req *models.UpdateFeedbackStatusRequest) error {
	if !isValidStatus(req.Status) {
		return ErrInvalidStatus
	}

	feedback, err := dao.GetFeedbackByID(feedbackID)
	if err != nil {
		return err
	}

	if feedback == nil {
		return ErrFeedbackNotFound
	}

	if req.Status == string(models.StatusClosed) && feedback.FeedbackType == models.FeedbackTypeBug {
		if strings.TrimSpace(req.ProcessingNote) == "" {
			return ErrProcessingNoteRequired
		}
	}

	return dao.UpdateFeedbackStatus(feedbackID, models.FeedbackStatus(req.Status), req.ProcessingNote)
}

func AddInternalNote(feedbackID int64, req *models.AddInternalNoteRequest) error {
	if utf8.RuneCountInString(req.Note) > MaxNoteLength {
		return ErrNoteTooLong
	}

	feedback, err := dao.GetFeedbackByID(feedbackID)
	if err != nil {
		return err
	}

	if feedback == nil {
		return ErrFeedbackNotFound
	}

	return dao.AddInternalNote(feedbackID, req.Note)
}

func GetStatistics() ([]*models.FeedbackStatistics, error) {
	return dao.GetStatistics()
}

func GetFeedbackByID(feedbackID int64) (*models.Feedback, error) {
	feedback, err := dao.GetFeedbackByID(feedbackID)
	if err != nil {
		return nil, err
	}

	if feedback == nil {
		return nil, ErrFeedbackNotFound
	}

	return feedback, nil
}
