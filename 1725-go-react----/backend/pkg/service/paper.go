package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"confman/pkg/model"
	"confman/pkg/repository"
	"confman/pkg/utils"
	"gorm.io/gorm"
)

type PaperService struct {
	repo           *repository.Repository
	db             *gorm.DB
	meetingService *MeetingService
}

func NewPaperService(repo *repository.Repository, db *gorm.DB, meetingService *MeetingService) *PaperService {
	return &PaperService{repo: repo, db: db, meetingService: meetingService}
}

type FormatCheckError struct {
	Field   string
	Message string
}

func (e *FormatCheckError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (s *PaperService) CheckFormat(paper *model.Paper) []*FormatCheckError {
	var errors []*FormatCheckError

	if len(paper.Abstract) < 200 {
		errors = append(errors, &FormatCheckError{Field: "abstract", Message: "摘要至少200字"})
	}
	if len(paper.Abstract) > 500 {
		errors = append(errors, &FormatCheckError{Field: "abstract", Message: "摘要最多500字"})
	}

	keywords := strings.Split(paper.Keywords, ",")
	for i := range keywords {
		keywords[i] = strings.TrimSpace(keywords[i])
	}
	filteredKeywords := make([]string, 0)
	for _, k := range keywords {
		if k != "" {
			filteredKeywords = append(filteredKeywords, k)
		}
	}
	if len(filteredKeywords) < 3 {
		errors = append(errors, &FormatCheckError{Field: "keywords", Message: "关键词至少3个"})
	}
	if len(filteredKeywords) > 5 {
		errors = append(errors, &FormatCheckError{Field: "keywords", Message: "关键词最多5个"})
	}

	if len(paper.Authors) == 0 {
		errors = append(errors, &FormatCheckError{Field: "authors", Message: "至少需要1位作者"})
	}

	hasFirstAuthor := false
	hasCorresponding := false
	for i, author := range paper.Authors {
		if author.Name == "" {
			errors = append(errors, &FormatCheckError{Field: "authors", Message: fmt.Sprintf("第%d位作者姓名不能为空", i+1)})
		}
		if author.Email == "" {
			errors = append(errors, &FormatCheckError{Field: "authors", Message: fmt.Sprintf("第%d位作者邮箱不能为空", i+1)})
		}
		if author.IsFirstAuthor {
			hasFirstAuthor = true
		}
		if author.IsCorresponding {
			hasCorresponding = true
		}
	}

	if !hasFirstAuthor {
		errors = append(errors, &FormatCheckError{Field: "authors", Message: "需要标记第一作者"})
	}
	if !hasCorresponding {
		errors = append(errors, &FormatCheckError{Field: "authors", Message: "需要标记通讯作者"})
	}

	return errors
}

func (s *PaperService) GetByMeetingID(meetingID uint) ([]model.Paper, error) {
	var papers []model.Paper
	result := s.db.Preload("Authors").Preload("Reviews").Preload("Schedule").Where("meeting_id = ?", meetingID).Find(&papers)
	return papers, result.Error
}

func (s *PaperService) GetByID(id uint) (*model.Paper, error) {
	var paper model.Paper
	result := s.db.Preload("Authors").Preload("Reviews.Reviewer").Preload("Schedule").First(&paper, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &paper, nil
}

func (s *PaperService) Create(paper *model.Paper, meetingAbbreviation string) error {
	existing := &model.Paper{}
	err := s.db.Where("meeting_id = ? AND title = ?", paper.MeetingID, paper.Title).First(existing).Error
	if err == nil {
		return errors.New("paper title already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	paper.UUID = utils.NewUUID()
	if paper.Status == "" {
		paper.Status = model.PaperStatusDraft
	}

	paperNumber, err := s.meetingService.GeneratePaperNumber(meetingAbbreviation, paper.MeetingID)
	if err != nil {
		return err
	}
	paper.PaperNumber = paperNumber

	tx := s.db.Begin()
	if err := tx.Create(paper).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range paper.Authors {
		paper.Authors[i].PaperID = paper.ID
		paper.Authors[i].Order = i
		if err := tx.Create(&paper.Authors[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *PaperService) Submit(paperID uint) error {
	paper, err := s.GetByID(paperID)
	if err != nil {
		return err
	}

	if paper.Status != model.PaperStatusDraft {
		return errors.New("only draft papers can be submitted")
	}

	accepting, err := s.meetingService.IsAcceptingSubmissions(paper.MeetingID)
	if err != nil {
		return err
	}
	if !accepting {
		return errors.New("submissions are not currently being accepted")
	}

	formatErrors := s.CheckFormat(paper)
	if len(formatErrors) > 0 {
		errMsgs := make([]string, 0, len(formatErrors))
		for _, e := range formatErrors {
			errMsgs = append(errMsgs, e.Error())
		}
		return errors.New(strings.Join(errMsgs, "; "))
	}

	updates := map[string]interface{}{
		"status":          model.PaperStatusFormatChecking,
		"submission_time": time.Now(),
	}
	return s.db.Model(&model.Paper{}).Where("id = ?", paperID).Updates(updates).Error
}

func (s *PaperService) UpdateStatus(paperID uint, newStatus model.PaperStatus) error {
	validTransitions := map[model.PaperStatus][]model.PaperStatus{
		model.PaperStatusDraft:         {model.PaperStatusSubmitted},
		model.PaperStatusSubmitted:     {model.PaperStatusFormatChecking},
		model.PaperStatusFormatChecking: {model.PaperStatusUnderReview, model.PaperStatusRevision},
		model.PaperStatusUnderReview:   {model.PaperStatusAccepted, model.PaperStatusRejected, model.PaperStatusRevision},
		model.PaperStatusRevision:      {model.PaperStatusUnderReview, model.PaperStatusAccepted, model.PaperStatusRejected},
		model.PaperStatusAccepted:      {model.PaperStatusPublished},
		model.PaperStatusRejected:      {},
		model.PaperStatusPublished:     {},
	}

	paper, err := s.GetByID(paperID)
	if err != nil {
		return err
	}

	validNexts, ok := validTransitions[paper.Status]
	if !ok {
		return errors.New("invalid current status")
	}

	isValid := false
	for _, s := range validNexts {
		if s == newStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("invalid status transition")
	}

	return s.db.Model(&model.Paper{}).Where("id = ?", paperID).Update("status", newStatus).Error
}

func (s *PaperService) GetAllForExport(meetingID uint) ([]map[string]interface{}, error) {
	var papers []model.Paper
	result := s.db.Preload("Authors").Preload("Reviews").Preload("Schedule").Where("meeting_id = ?", meetingID).Find(&papers)
	if result.Error != nil {
		return nil, result.Error
	}

	exportData := make([]map[string]interface{}, 0, len(papers))
	for _, p := range papers {
		authors := make([]string, 0, len(p.Authors))
		for _, a := range p.Authors {
			authors = append(authors, fmt.Sprintf("%s (%s)", a.Name, a.Email))
		}

		item := map[string]interface{}{
			"id":            p.ID,
			"paper_number":  p.PaperNumber,
			"title":         p.Title,
			"topic_area":    p.TopicArea,
			"status":        p.Status,
			"authors":       strings.Join(authors, "; "),
			"submitted_at":  p.SubmissionTime.Format(time.RFC3339),
		}

		if len(p.Reviews) > 0 {
			avgOriginality := 0.0
			avgQuality := 0.0
			avgRelevance := 0.0
			avgClarity := 0.0
			recommendations := make([]string, 0)
			for _, r := range p.Reviews {
				if r.SubmittedAt != nil {
					avgOriginality += float64(r.Originality)
					avgQuality += float64(r.TechnicalQuality)
					avgRelevance += float64(r.Relevance)
					avgClarity += float64(r.Clarity)
					recommendations = append(recommendations, string(r.Recommendation))
				}
			}
			submittedCount := len(recommendations)
			if submittedCount > 0 {
				item["avg_originality"] = avgOriginality / float64(submittedCount)
				item["avg_quality"] = avgQuality / float64(submittedCount)
				item["avg_relevance"] = avgRelevance / float64(submittedCount)
				item["avg_clarity"] = avgClarity / float64(submittedCount)
				item["recommendations"] = strings.Join(recommendations, ", ")
			}
		}

		if p.Schedule != nil {
			item["scheduled_day"] = p.Schedule.Day
			item["scheduled_session"] = p.Schedule.Session
			item["scheduled_start"] = p.Schedule.StartTime.Format(time.RFC3339)
			item["has_conflict"] = p.Schedule.HasConflict
		}

		exportData = append(exportData, item)
	}

	return exportData, nil
}

func (s *PaperService) ExportToJSON(meetingID uint) (string, error) {
	data, err := s.GetAllForExport(meetingID)
	if err != nil {
		return "", err
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
