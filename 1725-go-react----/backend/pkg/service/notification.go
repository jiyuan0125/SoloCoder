package service

import (
	"encoding/json"
	"log"
	"time"

	"confman/pkg/model"
	"confman/pkg/repository"
	"confman/pkg/utils"
	"gorm.io/gorm"
)

type NotificationService struct {
	repo *repository.Repository
	db   *gorm.DB
}

func NewNotificationService(repo *repository.Repository, db *gorm.DB) *NotificationService {
	return &NotificationService{repo: repo, db: db}
}

func (s *NotificationService) LogAudit(action, entityType, entityID string, details interface{}, userID *uint, hasError bool, errorMessage string) {
	detailsJSON := ""
	if details != nil {
		if bytes, err := json.Marshal(details); err == nil {
			detailsJSON = string(bytes)
		}
	}

	logEntry := &model.AuditLog{
		UUID:         utils.NewUUID(),
		Action:       action,
		EntityType:   entityType,
		EntityID:     entityID,
		Details:      detailsJSON,
		UserID:       userID,
		HasError:     hasError,
		ErrorMessage: errorMessage,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(logEntry).Error; err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	if hasError {
		s.notifyResponsible(entityType, entityID, errorMessage)
	}
}

func (s *NotificationService) notifyResponsible(entityType, entityID, errorMessage string) {
	var responsibleID *uint

	switch entityType {
	case "meeting":
		var meeting model.Meeting
		if err := s.db.Select("responsible_id").Where("id = ?", entityID).First(&meeting).Error; err == nil {
			responsibleID = meeting.ResponsibleID
		}
	case "paper":
		var paper model.Paper
		if err := s.db.Select("meeting_id").Where("id = ?", entityID).First(&paper).Error; err == nil {
			var meeting model.Meeting
			if err := s.db.Select("responsible_id").Where("id = ?", paper.MeetingID).First(&meeting).Error; err == nil {
				responsibleID = meeting.ResponsibleID
			}
		}
	}

	if responsibleID != nil {
		log.Printf("NOTIFICATION: Sending error notification to user %d: %s", *responsibleID, errorMessage)
	}
}

func (s *NotificationService) CreateReminder(meetingID uint, reminderType string, targetUserID uint, paperID *uint, message string) error {
	reminder := &model.Reminder{
		UUID:         utils.NewUUID(),
		MeetingID:    meetingID,
		Type:         reminderType,
		TargetUserID: targetUserID,
		PaperID:      paperID,
		Message:      message,
		IsRead:       false,
		SentAt:       time.Now(),
		CreatedAt:    time.Now(),
	}
	return s.db.Create(reminder).Error
}

func (s *NotificationService) CheckAndSendReminders() {
	log.Println("Checking for reminders to send...")

	var meetings []model.Meeting
	if err := s.db.Find(&meetings).Error; err != nil {
		log.Printf("Error fetching meetings: %v", err)
		return
	}

	now := time.Now()

	for _, meeting := range meetings {
		submissionDaysUntil := int(meeting.SubmissionDeadline.Sub(now).Hours() / 24)
		if submissionDaysUntil == 7 {
			s.sendSubmissionReminders(&meeting)
		}

		reviewDaysUntil := int(meeting.ReviewDeadline.Sub(now).Hours() / 24)
		if reviewDaysUntil == 3 {
			s.sendReviewReminders(&meeting)
		}

		if meeting.Status == model.MeetingStatusNotifying {
			s.sendAcceptanceReminders(&meeting)
		}
	}
}

func (s *NotificationService) sendSubmissionReminders(meeting *model.Meeting) {
	log.Printf("Sending submission reminders for meeting %s", meeting.Abbreviation)

	var papers []model.Paper
	if err := s.db.Preload("Authors").Where("meeting_id = ? AND status = ?", meeting.ID, model.PaperStatusDraft).Find(&papers).Error; err != nil {
		log.Printf("Error fetching draft papers: %v", err)
		return
	}

	for _, paper := range papers {
		for _, author := range paper.Authors {
			user, err := s.getUserByEmail(author.Email)
			if err != nil || user == nil {
				continue
			}
			msg := "投稿截止还有7天，请尽快完成您的论文投稿"
			paperID := paper.ID
			if err := s.CreateReminder(meeting.ID, "submission_reminder", user.ID, &paperID, msg); err != nil {
				log.Printf("Error creating reminder: %v", err)
			}
		}
	}
}

func (s *NotificationService) sendReviewReminders(meeting *model.Meeting) {
	log.Printf("Sending review reminders for meeting %s", meeting.Abbreviation)

	var reviews []model.Review
	if err := s.db.Joins("JOIN papers ON papers.id = reviews.paper_id").
		Where("papers.meeting_id = ? AND reviews.submitted_at IS NULL", meeting.ID).
		Find(&reviews).Error; err != nil {
		log.Printf("Error fetching pending reviews: %v", err)
		return
	}

	for _, review := range reviews {
		msg := "审稿截止还有3天，请尽快完成您的评审任务"
		if err := s.CreateReminder(meeting.ID, "review_reminder", review.ReviewerID, &review.PaperID, msg); err != nil {
			log.Printf("Error creating reminder: %v", err)
		}
	}
}

func (s *NotificationService) sendAcceptanceReminders(meeting *model.Meeting) {
	log.Printf("Sending acceptance reminders for meeting %s", meeting.Abbreviation)

	var papers []model.Paper
	if err := s.db.Preload("Authors").Where("meeting_id = ? AND status = ?", meeting.ID, model.PaperStatusAccepted).Find(&papers).Error; err != nil {
		log.Printf("Error fetching accepted papers: %v", err)
		return
	}

	for _, paper := range papers {
		for _, author := range paper.Authors {
			user, err := s.getUserByEmail(author.Email)
			if err != nil || user == nil {
				continue
			}
			msg := "您的论文已被录用，请确认是否参会"
			paperID := paper.ID
			if err := s.CreateReminder(meeting.ID, "acceptance_confirmation", user.ID, &paperID, msg); err != nil {
				log.Printf("Error creating reminder: %v", err)
			}
		}
	}
}

func (s *NotificationService) getUserByEmail(email string) (*model.User, error) {
	var user model.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *NotificationService) GetReminders(userID uint) ([]model.Reminder, error) {
	var reminders []model.Reminder
	result := s.db.Where("target_user_id = ?", userID).Order("created_at DESC").Find(&reminders)
	return reminders, result.Error
}

func (s *NotificationService) MarkReminderRead(reminderID uint) error {
	return s.db.Model(&model.Reminder{}).Where("id = ?", reminderID).Update("is_read", true).Error
}
