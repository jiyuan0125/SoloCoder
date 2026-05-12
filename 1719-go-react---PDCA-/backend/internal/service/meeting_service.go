package service

import (
	"errors"
	"net/http"

	"medical-quality-system/internal/app"
	"medical-quality-system/internal/middleware"
	"medical-quality-system/internal/model"

	"gorm.io/gorm"
)

type MeetingService struct{}

func NewMeetingService() *MeetingService {
	return &MeetingService{}
}

func (s *MeetingService) Create(meeting *model.Meeting, actionItems []model.ActionItem) error {
	tx := app.DB.Begin()

	if err := tx.Create(meeting).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range actionItems {
		actionItems[i].MeetingID = meeting.ID
		actionItems[i].Completed = false
		if err := tx.Create(&actionItems[i]).Error; err != nil {
			tx.Rollback()
			return err
		}

		todo := model.Todo{
			Type:        model.TodoTypeAction,
			Title:       actionItems[i].Content,
			Description: "会议行动项",
			MeetingID:   &meeting.ID,
			Responsible: actionItems[i].Responsible,
			DueDate:   actionItems[i].DueDate,
			Status:    model.TodoStatusPending,
		}
		if err := tx.Create(&todo).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (s *MeetingService) List() ([]model.Meeting, error) {
	var meetings []model.Meeting
	err := app.DB.Order("date DESC").Find(&meetings).Error
	return meetings, err
}

func (s *MeetingService) Get(id uint) (*model.Meeting, error) {
	var meeting model.Meeting
	err := app.DB.First(&meeting, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, middleware.NewAppError(http.StatusNotFound, "会议记录不存在")
	}
	return &meeting, err
}

func (s *MeetingService) Update(id uint, meeting *model.Meeting) error {
	existing, err := s.Get(id)
	if err != nil {
		return err
	}
	return app.DB.Model(existing).Updates(meeting).Error
}

func (s *MeetingService) Delete(id uint) error {
	_, err := s.Get(id)
	if err != nil {
		return err
	}
	return app.DB.Delete(&model.Meeting{}, id).Error
}

func (s *MeetingService) GetActionItems(meetingID uint) ([]model.ActionItem, error) {
	var items []model.ActionItem
	err := app.DB.Where("meeting_id = ?", meetingID).Find(&items).Error
	return items, err
}
