package service

import (
	"errors"
	"time"

	"confman/pkg/model"
	"confman/pkg/repository"
	"confman/pkg/utils"
	"gorm.io/gorm"
)

type ReviewService struct {
	repo *repository.Repository
	db   *gorm.DB
}

func NewReviewService(repo *repository.Repository, db *gorm.DB) *ReviewService {
	return &ReviewService{repo: repo, db: db}
}

func (s *ReviewService) GetReviewerCount(reviewerID uint, meetingID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.Review{}).
		Joins("JOIN papers ON papers.id = reviews.paper_id").
		Where("reviews.reviewer_id = ? AND papers.meeting_id = ?", reviewerID, meetingID).
		Count(&count).Error
	return count, err
}

func (s *ReviewService) IsReviewerAuthor(paperID uint, reviewerID uint) (bool, error) {
	var paper model.Paper
	err := s.db.Preload("Authors").First(&paper, paperID).Error
	if err != nil {
		return false, err
	}

	var reviewer model.User
	err = s.db.First(&reviewer, reviewerID).Error
	if err != nil {
		return false, err
	}

	for _, author := range paper.Authors {
		if author.Email == reviewer.Email {
			return true, nil
		}
	}
	return false, nil
}

func (s *ReviewService) AssignReviewer(paperID uint, reviewerID uint, meetingID uint) error {
	count, err := s.GetReviewerCount(reviewerID, meetingID)
	if err != nil {
		return err
	}
	if count >= 8 {
		return errors.New("reviewer cannot review more than 8 papers in the same round")
	}

	isAuthor, err := s.IsReviewerAuthor(paperID, reviewerID)
	if err != nil {
		return err
	}
	if isAuthor {
		return errors.New("reviewer cannot review their own paper")
	}

	var existing model.Review
	err = s.db.Where("paper_id = ? AND reviewer_id = ?", paperID, reviewerID).First(&existing).Error
	if err == nil {
		return errors.New("reviewer already assigned to this paper")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	review := &model.Review{
		UUID:       utils.NewUUID(),
		PaperID:    paperID,
		ReviewerID: reviewerID,
	}
	return s.db.Create(review).Error
}

func (s *ReviewService) ValidateReview(review *model.Review) []string {
	var errors []string

	if review.Originality < 1 || review.Originality > 5 {
		errors = append(errors, "originality score must be between 1 and 5")
	}
	if review.TechnicalQuality < 1 || review.TechnicalQuality > 5 {
		errors = append(errors, "technical quality score must be between 1 and 5")
	}
	if review.Relevance < 1 || review.Relevance > 5 {
		errors = append(errors, "relevance score must be between 1 and 5")
	}
	if review.Clarity < 1 || review.Clarity > 5 {
		errors = append(errors, "clarity score must be between 1 and 5")
	}

	validRecs := []model.ReviewRecommendation{
		model.StrongAccept, model.WeakAccept, model.Borderline,
		model.WeakReject, model.StrongReject,
	}
	isValidRec := false
	for _, r := range validRecs {
		if r == review.Recommendation {
			isValidRec = true
			break
		}
	}
	if !isValidRec {
		errors = append(errors, "invalid recommendation")
	}

	if len(review.Comments) < 100 {
		errors = append(errors, "comments must be at least 100 characters")
	}

	return errors
}

func (s *ReviewService) SubmitReview(reviewID uint, updates map[string]interface{}) error {
	now := time.Now()
	updates["submitted_at"] = &now

	return s.db.Model(&model.Review{}).Where("id = ?", reviewID).Updates(updates).Error
}

func (s *ReviewService) AggregateReviews(paperID uint) (model.PaperStatus, error) {
	var reviews []model.Review
	err := s.db.Where("paper_id = ? AND submitted_at IS NOT NULL", paperID).Find(&reviews).Error
	if err != nil {
		return "", err
	}

	if len(reviews) == 0 {
		return "", errors.New("no submitted reviews yet")
	}

	hasStrongAccept := false
	hasStrongReject := false
	hasAccept := false

	for _, r := range reviews {
		switch r.Recommendation {
		case model.StrongAccept:
			hasStrongAccept = true
			hasAccept = true
		case model.WeakAccept:
			hasAccept = true
		case model.StrongReject:
			hasStrongReject = true
		}
	}

	if hasAccept && !hasStrongReject {
		return model.PaperStatusAccepted, nil
	}
	if hasStrongReject && !hasStrongAccept {
		return model.PaperStatusRejected, nil
	}

	return "", nil
}

func (s *ReviewService) GetByReviewerID(reviewerID uint) ([]model.Review, error) {
	var reviews []model.Review
	result := s.db.Preload("Reviewer").Where("reviewer_id = ?", reviewerID).Find(&reviews)
	return reviews, result.Error
}

func (s *ReviewService) GetByPaperID(paperID uint) ([]model.Review, error) {
	var reviews []model.Review
	result := s.db.Preload("Reviewer").Where("paper_id = ?", paperID).Find(&reviews)
	return reviews, result.Error
}
