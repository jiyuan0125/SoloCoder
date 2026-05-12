package services

import (
	"teaching-evaluation-system/pkg/database"
	"teaching-evaluation-system/pkg/models"
)

type UpdateFeedbackRequest struct {
	ImprovementPlan string `json:"improvement_plan"`
}

func GetFeedbackItems(teacherID *uint) ([]models.FeedbackItem, error) {
	var items []models.FeedbackItem
	query := database.DB.Preload("CourseResult").Order("created_at DESC")
	if teacherID != nil {
		query = query.Where("teacher_id = ?", *teacherID)
	}
	err := query.Find(&items).Error
	return items, err
}

func GetFeedbackItem(itemID uint) (*models.FeedbackItem, error) {
	var item models.FeedbackItem
	err := database.DB.Preload("CourseResult").First(&item, itemID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func UpdateImprovementPlan(itemID uint, req UpdateFeedbackRequest) error {
	var item models.FeedbackItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		return err
	}

	item.ImprovementPlan = req.ImprovementPlan
	item.Status = models.FeedbackStatusPlanFilled
	return database.DB.Save(&item).Error
}

func UpdateFeedbackStatus(itemID uint, status models.FeedbackStatus) error {
	var item models.FeedbackItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		return err
	}

	item.Status = status
	return database.DB.Save(&item).Error
}
