package service

import (
	"time"

	"confman/pkg/model"
	"confman/pkg/repository"
	"confman/pkg/utils"
	"gorm.io/gorm"
)

type MeetingService struct {
	repo *repository.Repository
	db   *gorm.DB
}

func NewMeetingService(repo *repository.Repository, db *gorm.DB) *MeetingService {
	return &MeetingService{repo: repo, db: db}
}

func (s *MeetingService) GetAll() ([]model.Meeting, error) {
	var meetings []model.Meeting
	result := s.db.Preload("Chair").Preload("Responsible").Find(&meetings)
	return meetings, result.Error
}

func (s *MeetingService) GetByID(id uint) (*model.Meeting, error) {
	var meeting model.Meeting
	result := s.db.Preload("Chair").Preload("Responsible").First(&meeting, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &meeting, nil
}

func (s *MeetingService) Create(meeting *model.Meeting) error {
	meeting.UUID = utils.NewUUID()
	if meeting.Status == "" {
		meeting.Status = model.MeetingStatusPreparing
	}
	return s.db.Create(meeting).Error
}

func (s *MeetingService) Update(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.Meeting{}).Where("id = ?", id).Updates(updates).Error
}

func (s *MeetingService) Delete(id uint) error {
	return s.db.Delete(&model.Meeting{}, id).Error
}

func (s *MeetingService) IsAcceptingSubmissions(meetingID uint) (bool, error) {
	meeting, err := s.GetByID(meetingID)
	if err != nil {
		return false, err
	}
	if meeting.Status != model.MeetingStatusAcceptingSubmissions {
		return false, nil
	}
	now := time.Now()
	return now.Before(meeting.SubmissionDeadline), nil
}

func (s *MeetingService) GetPaperCount(meetingID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.Paper{}).Where("meeting_id = ?", meetingID).Count(&count).Error
	return count, err
}

func (s *MeetingService) GeneratePaperNumber(meetingAbbreviation string, meetingID uint) (string, error) {
	count, err := s.GetPaperCount(meetingID)
	if err != nil {
		return "", err
	}
	nextNum := count + 1
	return meetingAbbreviation + "-" + formatPaperNumber(int(nextNum)), nil
}

func formatPaperNumber(num int) string {
	if num < 10 {
		return "000" + string(rune('0'+num))
	} else if num < 100 {
		return "00" + string(rune('0'+num/10)) + string(rune('0'+num%10))
	} else if num < 1000 {
		return "0" + string(rune('0'+num/100)) + string(rune('0'+(num/10)%10)) + string(rune('0'+num%10))
	}
	return string(rune('0'+num/1000)) + string(rune('0'+(num/100)%10)) + string(rune('0'+(num/10)%10)) + string(rune('0'+num%10))
}
