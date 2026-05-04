package model

import (
	"time"

	"feedback-system/pkg/common"
)

type Feedback struct {
	ID          string
	UserID      string
	Type        common.FeedbackType
	Rating      int
	Description string
	Status      common.FeedbackStatus
	InternalNote string
	ProcessNote string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (f *Feedback) ToListItem() common.FeedbackListItem {
	return common.FeedbackListItem{
		ID:          f.ID,
		UserID:      f.UserID,
		Type:        f.Type,
		Rating:      f.Rating,
		Description: f.Description,
		Status:      f.Status,
		InternalNote: f.InternalNote,
		ProcessNote: f.ProcessNote,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

type UserDailyCount struct {
	UserID    string
	Date      string
	Count     int
	UpdatedAt time.Time
}
