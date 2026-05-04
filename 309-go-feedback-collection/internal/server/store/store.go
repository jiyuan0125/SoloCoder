package store

import (
	"errors"
	"time"

	"feedback-system/internal/server/model"
	"feedback-system/pkg/common"
)

var (
	ErrFeedbackNotFound = errors.New("feedback not found")
	ErrDailyLimitExceeded = errors.New("daily feedback limit exceeded")
)

type Store interface {
	CreateFeedback(feedback *model.Feedback) error
	GetFeedback(id string) (*model.Feedback, error)
	ListFeedbacks(page, pageSize int, filterType common.FeedbackType, filterStatus common.FeedbackStatus) ([]*model.Feedback, int, error)
	UpdateFeedback(feedback *model.Feedback) error
	GetUserDailyCount(userID string, date string) (int, error)
	IncrementUserDailyCount(userID string, date string) error
	GetStatistics() ([]common.TypeStats, int, float64, error)
}

func GetTodayDateString() string {
	return time.Now().Format("2006-01-02")
}
