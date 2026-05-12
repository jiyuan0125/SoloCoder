package service

import (
	"errors"
	"fmt"
	"time"

	"confman/pkg/model"
	"confman/pkg/repository"
	"confman/pkg/utils"
	"gorm.io/gorm"
)

type ScheduleService struct {
	repo *repository.Repository
	db   *gorm.DB
}

func NewScheduleService(repo *repository.Repository, db *gorm.DB) *ScheduleService {
	return &ScheduleService{repo: repo, db: db}
}

func (s *ScheduleService) CheckSessionOverlap(meetingID uint, day int, startTime, endTime time.Time, excludeID *uint) (bool, error) {
	var count int64

	query := s.db.Model(&model.Session{}).
		Where("meeting_id = ? AND day = ?", meetingID, day).
		Where("start_time < ? AND end_time > ?", endTime, startTime)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *ScheduleService) CreateSession(session *model.Session) error {
	overlaps, err := s.CheckSessionOverlap(session.MeetingID, session.Day, session.StartTime, session.EndTime, nil)
	if err != nil {
		return err
	}
	if overlaps {
		return errors.New("session time overlaps with existing session")
	}

	session.UUID = utils.NewUUID()
	return s.db.Create(session).Error
}

func (s *ScheduleService) GetSessions(meetingID uint) ([]model.Session, error) {
	var sessions []model.Session
	result := s.db.Where("meeting_id = ?", meetingID).Order("day, start_time").Find(&sessions)
	return sessions, result.Error
}

func (s *ScheduleService) CheckAuthorConflict(paperID uint, day int, startTime, endTime time.Time) (bool, error) {
	var paper model.Paper
	err := s.db.Preload("Authors").First(&paper, paperID).Error
	if err != nil {
		return false, err
	}

	authorEmails := make([]string, 0, len(paper.Authors))
	for _, a := range paper.Authors {
		authorEmails = append(authorEmails, a.Email)
	}

	var conflictingSchedules []model.Schedule
	err = s.db.Joins("JOIN papers ON papers.id = schedules.paper_id").
		Joins("JOIN paper_authors ON paper_authors.paper_id = papers.id").
		Where("schedules.meeting_id = ? AND schedules.day = ?", paper.MeetingID, day).
		Where("schedules.start_time < ? AND schedules.end_time > ?", endTime, startTime).
		Where("schedules.paper_id != ?", paperID).
		Where("paper_authors.email IN ?", authorEmails).
		Find(&conflictingSchedules).Error

	if err != nil {
		return false, err
	}

	return len(conflictingSchedules) > 0, nil
}

func (s *ScheduleService) SchedulePaper(schedule *model.Schedule) error {
	var paper model.Paper
	err := s.db.First(&paper, schedule.PaperID).Error
	if err != nil {
		return err
	}

	if paper.Status != model.PaperStatusAccepted {
		return errors.New("only accepted papers can be scheduled")
	}

	if schedule.Day < 1 || schedule.Day > 3 {
		return errors.New("day must be between 1 and 3")
	}

	hasConflict, err := s.CheckAuthorConflict(schedule.PaperID, schedule.Day, schedule.StartTime, schedule.EndTime)
	if err != nil {
		return err
	}

	schedule.HasConflict = hasConflict
	schedule.UUID = utils.NewUUID()

	var existing model.Schedule
	err = s.db.Where("paper_id = ?", schedule.PaperID).First(&existing).Error
	if err == nil {
		return s.db.Model(&model.Schedule{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
			"day":         schedule.Day,
			"session":     schedule.Session,
			"start_time":  schedule.StartTime,
			"end_time":    schedule.EndTime,
			"session_id":  schedule.SessionID,
			"has_conflict": schedule.HasConflict,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	return s.db.Create(schedule).Error
}

func (s *ScheduleService) GetSchedule(meetingID uint) ([]model.Schedule, error) {
	var schedules []model.Schedule
	result := s.db.Preload("Paper.Authors").Where("meeting_id = ?", meetingID).Order("day, start_time").Find(&schedules)
	return schedules, result.Error
}

func (s *ScheduleService) DeleteSchedule(scheduleID uint) error {
	return s.db.Delete(&model.Schedule{}, scheduleID).Error
}

func (s *ScheduleService) GenerateDefaultSessions(meetingID uint, startDate time.Time) error {
	for day := 1; day <= 3; day++ {
		dayDate := startDate.AddDate(0, 0, day-1)

		morningStart := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), 9, 0, 0, 0, dayDate.Location())
		morningEnd := morningStart.Add(3 * time.Hour)
		morningSession := &model.Session{
			UUID:      utils.NewUUID(),
			MeetingID: meetingID,
			Day:       day,
			TimeOfDay: "morning",
			Name:      fmt.Sprintf("Day %d Morning Session", day),
			StartTime: morningStart,
			EndTime:   morningEnd,
		}
		if err := s.db.Create(morningSession).Error; err != nil {
			return err
		}

		afternoonStart := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), 14, 0, 0, 0, dayDate.Location())
		afternoonEnd := afternoonStart.Add(3 * time.Hour)
		afternoonSession := &model.Session{
			UUID:      utils.NewUUID(),
			MeetingID: meetingID,
			Day:       day,
			TimeOfDay: "afternoon",
			Name:      fmt.Sprintf("Day %d Afternoon Session", day),
			StartTime: afternoonStart,
			EndTime:   afternoonEnd,
		}
		if err := s.db.Create(afternoonSession).Error; err != nil {
			return err
		}
	}
	return nil
}
