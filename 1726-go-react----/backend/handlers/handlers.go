package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"paper-management-platform/models"
	"paper-management-platform/store"
)

const (
	ReviewTimeoutDays = 30
	MinScore          = 1
	MaxScore          = 10
)

func getCurrentUser(c *gin.Context) *models.User {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "author1"
	}
	user, _ := store.GetStore().GetUser(userID)
	return user
}

func addVersionHistory(paper *models.Paper, newStatus models.PaperStatus, operatorID, operatorName, description string) {
	history := models.VersionHistory{
		ID:            store.GenerateHistoryID(),
		PaperID:       paper.ID,
		Status:        newStatus,
		OperatorID:    operatorID,
		OperatorName:  operatorName,
		ChangedAt:     time.Now(),
		Description:   description,
	}
	paper.VersionHistory = append(paper.VersionHistory, history)
	paper.UpdatedAt = time.Now()
}

func parseKeywords(keywordsStr string) []string {
	if keywordsStr == "" {
		return []string{}
	}
	parts := strings.Split(keywordsStr, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func isReviewerPaperAuthor(paper *models.Paper, reviewerID string) bool {
	for _, author := range paper.Authors {
		if author.ID == reviewerID {
			return true
		}
	}
	return false
}

func calculateReviewMetrics(paper *models.Paper) (avgScore float64, hasReject bool, allSubmitted bool, allOrTimeout bool) {
	if len(paper.AssignedReviewers) == 0 {
		return 0, false, false, false
	}

	now := time.Now()
	totalScore := 0
	submittedCount := 0
	allSubmitted = true
	allOrTimeout = true

	for _, reviewer := range paper.AssignedReviewers {
		if reviewer.Submitted {
			submittedCount++
		} else {
			allSubmitted = false
			daysSince := now.Sub(reviewer.AssignedAt).Hours() / 24
			if daysSince <= float64(ReviewTimeoutDays) {
				allOrTimeout = false
			}
		}
	}

	for _, review := range paper.Reviews {
		totalScore += review.Score
		if review.Decision == models.DecisionReject {
			hasReject = true
		}
	}

	if submittedCount > 0 {
		avgScore = float64(totalScore) / float64(submittedCount)
	}

	return avgScore, hasReject, allSubmitted, allOrTimeout
}

func shouldAutoReject(avgScore float64, hasReject bool) bool {
	if avgScore < 5 {
		return true
	}
	if hasReject && avgScore < 6 {
		return true
	}
	return false
}

func GetUsers(c *gin.Context) {
	users := store.GetStore().ListUsers()
	c.JSON(http.StatusOK, users)
}

func CreatePaper(c *gin.Context) {
	var req models.CreatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	user := getCurrentUser(c)
	now := time.Now()

	paper := &models.Paper{
		ID:              store.GenerateID(),
		Title:           req.Title,
		Abstract:        req.Abstract,
		Keywords:        parseKeywords(req.Keywords),
		Authors:         req.Authors,
		SubjectCategory: req.SubjectCategory,
		TargetJournal:   req.TargetJournal,
		Status:          models.StatusDraft,
		CreatedBy:       user.ID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	addVersionHistory(paper, models.StatusDraft, user.ID, user.Name, "创建论文草稿")

	if err := store.GetStore().CreatePaper(paper); err != nil {
		if err.Error() == "title already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{Error: "conflict", Message: "论文标题已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, paper)
}

func ListPapers(c *gin.Context) {
	statusFilter := c.Query("status")
	keywordFilter := c.Query("keyword")
	sortBy := c.Query("sort")

	papers := store.GetStore().ListPapers()

	filtered := make([]*models.Paper, 0)
	for _, p := range papers {
		if statusFilter != "" && string(p.Status) != statusFilter {
			continue
		}
		if keywordFilter != "" {
			matched := false
			for _, kw := range p.Keywords {
				if strings.Contains(strings.ToLower(kw), strings.ToLower(keywordFilter)) {
					matched = true
					break
				}
			}
			if !matched && !strings.Contains(strings.ToLower(p.Title), strings.ToLower(keywordFilter)) {
				continue
			}
		}
		filtered = append(filtered, p)
	}

	if sortBy == "created_at" || sortBy == "" {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		})
	}

	c.JSON(http.StatusOK, filtered)
}

func GetPaper(c *gin.Context) {
	id := c.Param("id")
	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}
	c.JSON(http.StatusOK, paper)
}

func UpdatePaper(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	var req models.UpdatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusDraft && paper.Status != models.StatusRevision {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只有草稿或修改中状态可以修改"})
		return
	}

	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		if req.Title != "" && req.Title != p.Title {
			if err := store.GetStore().CheckTitleUnique(req.Title, id); err != nil {
				return err
			}
			p.Title = req.Title
		}
		if req.Abstract != "" {
			p.Abstract = req.Abstract
		}
		if req.Keywords != "" {
			p.Keywords = parseKeywords(req.Keywords)
		}
		if len(req.Authors) > 0 {
			p.Authors = req.Authors
		}
		if req.SubjectCategory != "" {
			p.SubjectCategory = req.SubjectCategory
		}
		if req.TargetJournal != "" {
			p.TargetJournal = req.TargetJournal
		}
		return nil
	})

	if err != nil {
		if err.Error() == "title already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{Error: "conflict", Message: "论文标题已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	addVersionHistory(updatedPaper, updatedPaper.Status, user.ID, user.Name, "更新论文信息")

	c.JSON(http.StatusOK, updatedPaper)
}

func SubmitForReview(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusDraft && paper.Status != models.StatusRevision {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能从草稿或修改中状态提交审稿"})
		return
	}

	newStatus := models.StatusPendingReview
	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		p.Status = newStatus
		addVersionHistory(p, newStatus, user.ID, user.Name, "提交审稿，等待分配审稿人")
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	c.JSON(http.StatusOK, updatedPaper)
}

func AssignReviewers(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	if user.Role != models.RoleEditor && user.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "只有编辑或管理员可以分配审稿人"})
		return
	}

	var req models.AssignReviewersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusPendingReview {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能在待审稿状态分配审稿人"})
		return
	}

	for _, reviewer := range req.Reviewers {
		if isReviewerPaperAuthor(paper, reviewer.ID) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "审稿人不能评审自己参与的论文"})
			return
		}
	}

	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		now := time.Now()
		for _, r := range req.Reviewers {
			p.AssignedReviewers = append(p.AssignedReviewers, models.Reviewer{
				ID:          r.ID,
				Name:        r.Name,
				Affiliation: r.Affiliation,
				AssignedAt:  now,
				Submitted:   false,
			})
		}
		p.Status = models.StatusInReview
		addVersionHistory(p, models.StatusInReview, user.ID, user.Name, "分配审稿人，进入审稿中")
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	c.JSON(http.StatusOK, updatedPaper)
}

func SubmitReview(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	if user.Role != models.RoleReviewer && user.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "只有审稿人或管理员可以提交评审"})
		return
	}

	var req models.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	if req.Score < MinScore || req.Score > MaxScore {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "分数必须在1到10之间"})
		return
	}

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusInReview {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能在审稿中状态提交评审"})
		return
	}

	if isReviewerPaperAuthor(paper, req.ReviewerID) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "审稿人不能评审自己参与的论文"})
		return
	}

	foundAssigned := false
	for i, r := range paper.AssignedReviewers {
		if r.ID == req.ReviewerID {
			if paper.AssignedReviewers[i].Submitted {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "已经提交过评审"})
				return
			}
			foundAssigned = true
			break
		}
	}
	if !foundAssigned {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "该审稿人未被分配到此论文"})
		return
	}

	reviewerName := user.Name
	for _, r := range paper.AssignedReviewers {
		if r.ID == req.ReviewerID {
			reviewerName = r.Name
			break
		}
	}

	var updatedPaper *models.Paper
	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		for i := range p.AssignedReviewers {
			if p.AssignedReviewers[i].ID == req.ReviewerID {
				p.AssignedReviewers[i].Submitted = true
			}
		}

		review := models.Review{
			ID:           store.GenerateReviewID(),
			ReviewerID:   req.ReviewerID,
			ReviewerName: reviewerName,
			Decision:     req.Decision,
			Score:        req.Score,
			Comments:     req.Comments,
			SubmittedAt:  time.Now(),
		}
		p.Reviews = append(p.Reviews, review)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ = store.GetStore().GetPaper(id)

	avgScore, hasReject, allSubmitted, allOrTimeout := calculateReviewMetrics(updatedPaper)

	var finalStatus models.PaperStatus
	var desc string

	if allOrTimeout {
		if shouldAutoReject(avgScore, hasReject) {
			finalStatus = models.StatusRejected
			desc = "自动拒稿：平均分=" + strconv.FormatFloat(avgScore, 'f', 2, 64)
		} else if allSubmitted {
			if hasReject {
				finalStatus = models.StatusRevision
				desc = "需要修改后重审"
			} else {
				finalStatus = models.StatusAccepted
				desc = "审稿通过"
			}
		}
	}

	if finalStatus != "" {
		err = store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
			p.Status = finalStatus
			addVersionHistory(p, finalStatus, user.ID, user.Name, desc)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
			return
		}
		updatedPaper, _ = store.GetStore().GetPaper(id)
	} else {
		addVersionHistory(updatedPaper, updatedPaper.Status, user.ID, user.Name, "审稿人 "+reviewerName+" 提交评审")
	}

	c.JSON(http.StatusOK, updatedPaper)
}

func ApproveAction(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusInReview {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能在审稿中状态执行审批"})
		return
	}

	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		p.Status = models.StatusAccepted
		addVersionHistory(p, models.StatusAccepted, user.ID, user.Name, "人工审核通过")
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	c.JSON(http.StatusOK, updatedPaper)
}

func RejectAction(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusInReview {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能在审稿中状态执行拒稿"})
		return
	}

	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		p.Status = models.StatusRejected
		addVersionHistory(p, models.StatusRejected, user.ID, user.Name, "人工拒稿")
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	c.JSON(http.StatusOK, updatedPaper)
}

func CancelAction(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusPendingReview && paper.Status != models.StatusInReview {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能在待审稿或审稿中状态取消"})
		return
	}

	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		p.Status = models.StatusDraft
		addVersionHistory(p, models.StatusDraft, user.ID, user.Name, "撤回审稿，返回草稿")
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	c.JSON(http.StatusOK, updatedPaper)
}

func PublishPaper(c *gin.Context) {
	id := c.Param("id")
	user := getCurrentUser(c)

	if user.Role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "只有管理员可以发表论文"})
		return
	}

	var req models.PublishPaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	if paper.Status != models.StatusAccepted {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "只能从已录用状态发表"})
		return
	}

	err := store.GetStore().UpdatePaper(id, func(p *models.Paper) error {
		p.Status = models.StatusPublished
		p.PublicationInfo = &models.PublicationInfo{
			JournalName: req.PublicationInfo.JournalName,
			Volume:      req.PublicationInfo.Volume,
			Issue:       req.PublicationInfo.Issue,
			PageStart:   req.PublicationInfo.PageStart,
			PageEnd:     req.PublicationInfo.PageEnd,
			DOI:         req.PublicationInfo.DOI,
		}
		addVersionHistory(p, models.StatusPublished, user.ID, user.Name, "论文已发表，DOI: "+req.PublicationInfo.DOI)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	updatedPaper, _ := store.GetStore().GetPaper(id)
	c.JSON(http.StatusOK, updatedPaper)
}

func GetVersionHistory(c *gin.Context) {
	id := c.Param("id")
	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	history := make([]models.VersionHistory, len(paper.VersionHistory))
	copy(history, paper.VersionHistory)
	sort.Slice(history, func(i, j int) bool {
		return history[i].ChangedAt.After(history[j].ChangedAt)
	})

	c.JSON(http.StatusOK, history)
}

func GetReviews(c *gin.Context) {
	id := c.Param("id")
	paper, exists := store.GetStore().GetPaper(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "论文不存在"})
		return
	}

	c.JSON(http.StatusOK, paper.Reviews)
}
