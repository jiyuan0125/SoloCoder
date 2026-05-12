package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"confman/pkg/model"
	"confman/pkg/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	meetingService    *service.MeetingService
	paperService      *service.PaperService
	reviewService     *service.ReviewService
	scheduleService   *service.ScheduleService
	userService       *service.UserService
	notificationService *service.NotificationService
}

func New(
	meetingService *service.MeetingService,
	paperService *service.PaperService,
	reviewService *service.ReviewService,
	scheduleService *service.ScheduleService,
	userService *service.UserService,
	notificationService *service.NotificationService,
) *Handler {
	return &Handler{
		meetingService:    meetingService,
		paperService:      paperService,
		reviewService:     reviewService,
		scheduleService:   scheduleService,
		userService:       userService,
		notificationService: notificationService,
	}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	r.GET("/api/health", h.HealthCheck)

	api := r.Group("/api")
	{
		meetings := api.Group("/meetings")
		{
			meetings.GET("", h.ListMeetings)
			meetings.GET("/:id", h.GetMeeting)
			meetings.POST("", h.CreateMeeting)
			meetings.PUT("/:id", h.UpdateMeeting)
			meetings.DELETE("/:id", h.DeleteMeeting)
			meetings.GET("/:id/papers", h.ListMeetingPapers)
			meetings.GET("/:id/papers/export", h.ExportPapers)
			meetings.POST("/:id/sessions", h.CreateSession)
			meetings.GET("/:id/sessions", h.ListSessions)
			meetings.GET("/:id/schedule", h.GetMeetingSchedule)
		}

		papers := api.Group("/papers")
		{
			papers.GET("/:id", h.GetPaper)
			papers.POST("", h.CreatePaper)
			papers.POST("/:id/submit", h.SubmitPaper)
			papers.PUT("/:id/status", h.UpdatePaperStatus)
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("/assign", h.AssignReviewer)
			reviews.PUT("/:id", h.SubmitReview)
			reviews.GET("/reviewer/:id", h.GetReviewerReviews)
		}

		schedules := api.Group("/schedules")
		{
			schedules.POST("", h.SchedulePaper)
			schedules.DELETE("/:id", h.DeleteSchedule)
		}

		users := api.Group("/users")
		{
			users.GET("", h.ListUsers)
			users.GET("/reviewers", h.ListReviewers)
			users.POST("", h.CreateUser)
			users.GET("/:id/reminders", h.GetUserReminders)
			users.PUT("/reminders/:id/read", h.MarkReminderRead)
		}
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getUintParam(c *gin.Context, name string) (uint, error) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *Handler) ListMeetings(c *gin.Context) {
	meetings, err := h.meetingService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meetings)
}

func (h *Handler) GetMeeting(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	meeting, err := h.meetingService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "meeting not found"})
		return
	}
	c.JSON(http.StatusOK, meeting)
}

func (h *Handler) CreateMeeting(c *gin.Context) {
	var meeting model.Meeting
	if err := c.ShouldBindJSON(&meeting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.meetingService.Create(&meeting); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "meeting abbreviation already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !meeting.StartDate.IsZero() {
		_ = h.scheduleService.GenerateDefaultSessions(meeting.ID, meeting.StartDate)
	}

	c.JSON(http.StatusCreated, meeting)
}

func (h *Handler) UpdateMeeting(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.meetingService.Update(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	meeting, err := h.meetingService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	c.JSON(http.StatusOK, meeting)
}

func (h *Handler) DeleteMeeting(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.meetingService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ListMeetingPapers(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	papers, err := h.paperService.GetByMeetingID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, papers)
}

func (h *Handler) ExportPapers(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	data, err := h.paperService.ExportToJSON(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=papers.json")
	c.String(http.StatusOK, data)
}

func (h *Handler) GetPaper(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	paper, err := h.paperService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "paper not found"})
		return
	}

	c.JSON(http.StatusOK, paper)
}

type CreatePaperRequest struct {
	MeetingID uint                  `json:"meeting_id" binding:"required"`
	Title     string                `json:"title" binding:"required"`
	Abstract  string                `json:"abstract" binding:"required"`
	Keywords  string                `json:"keywords"`
	TopicArea string                `json:"topic_area"`
	Content   string                `json:"content"`
	Authors   []model.PaperAuthor   `json:"authors" binding:"required"`
}

func (h *Handler) CreatePaper(c *gin.Context) {
	var req CreatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meeting, err := h.meetingService.GetByID(req.MeetingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "meeting not found"})
		return
	}

	paper := &model.Paper{
		MeetingID: req.MeetingID,
		Title:     req.Title,
		Abstract:  req.Abstract,
		Keywords:  req.Keywords,
		TopicArea: req.TopicArea,
		Content:   req.Content,
		Authors:   req.Authors,
	}

	if err := h.paperService.Create(paper, meeting.Abbreviation); err != nil {
		if err.Error() == "paper title already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "paper title already exists in this meeting"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, paper)
}

func (h *Handler) SubmitPaper(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.paperService.Submit(id); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "submissions are not currently") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "only draft") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type UpdatePaperStatusRequest struct {
	Status model.PaperStatus `json:"status" binding:"required"`
}

func (h *Handler) UpdatePaperStatus(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdatePaperStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paperService.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type AssignReviewerRequest struct {
	PaperID    uint `json:"paper_id" binding:"required"`
	ReviewerID uint `json:"reviewer_id" binding:"required"`
	MeetingID  uint `json:"meeting_id" binding:"required"`
}

func (h *Handler) AssignReviewer(c *gin.Context) {
	var req AssignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.reviewService.AssignReviewer(req.PaperID, req.ReviewerID, req.MeetingID); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "reviewer cannot review more than 8") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "reviewer cannot review their own") {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type SubmitReviewRequest struct {
	Originality         int                        `json:"originality"`
	TechnicalQuality    int                        `json:"technical_quality"`
	Relevance           int                        `json:"relevance"`
	Clarity             int                        `json:"clarity"`
	Recommendation      model.ReviewRecommendation `json:"recommendation"`
	Comments            string                     `json:"comments"`
	ConfidentialComments string                    `json:"confidential_comments"`
}

func (h *Handler) SubmitReview(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tempReview := &model.Review{
		Originality:         req.Originality,
		TechnicalQuality:    req.TechnicalQuality,
		Relevance:           req.Relevance,
		Clarity:             req.Clarity,
		Recommendation:      req.Recommendation,
		Comments:            req.Comments,
	}

	validationErrors := h.reviewService.ValidateReview(tempReview)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	updates := map[string]interface{}{
		"originality":          req.Originality,
		"technical_quality":    req.TechnicalQuality,
		"relevance":            req.Relevance,
		"clarity":              req.Clarity,
		"recommendation":       req.Recommendation,
		"comments":             req.Comments,
		"confidential_comments": req.ConfidentialComments,
	}

	if err := h.reviewService.SubmitReview(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetReviewerReviews(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	reviews, err := h.reviewService.GetByReviewerID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (h *Handler) CreateSession(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var session model.Session
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session.MeetingID = id

	if err := h.scheduleService.CreateSession(&session); err != nil {
		if strings.Contains(err.Error(), "overlaps") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

func (h *Handler) ListSessions(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sessions, err := h.scheduleService.GetSessions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

type SchedulePaperRequest struct {
	MeetingID uint   `json:"meeting_id" binding:"required"`
	PaperID   uint   `json:"paper_id" binding:"required"`
	Day       int    `json:"day" binding:"required"`
	Session   string `json:"session"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

func (h *Handler) SchedulePaper(c *gin.Context) {
	var req SchedulePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := parseTime(req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format"})
		return
	}
	endTime, err := parseTime(req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
		return
	}

	schedule := &model.Schedule{
		MeetingID: req.MeetingID,
		PaperID:   req.PaperID,
		Day:       req.Day,
		Session:   req.Session,
		StartTime: startTime,
		EndTime:   endTime,
	}

	if err := h.scheduleService.SchedulePaper(schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func (h *Handler) GetMeetingSchedule(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	schedules, err := h.scheduleService.GetSchedule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedules)
}

func (h *Handler) DeleteSchedule(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.scheduleService.DeleteSchedule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.userService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) ListReviewers(c *gin.Context) {
	users, err := h.userService.GetReviewers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Create(&user); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetUserReminders(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	reminders, err := h.notificationService.GetReminders(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reminders)
}

func (h *Handler) MarkReminderRead(c *gin.Context) {
	id, err := getUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.notificationService.MarkReminderRead(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}
