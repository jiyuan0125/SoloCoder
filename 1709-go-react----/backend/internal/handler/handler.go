package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"organdonation/internal/models"
	"organdonation/internal/service"
	apperrors "organdonation/pkg/errors"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	appErr := apperrors.FromError(err)
	c.JSON(appErr.Code, gin.H{"error": appErr.Message})
}

func getPageAndSize(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func paginatedResponse(data interface{}, page, size, total int) models.PaginatedResponse {
	totalPages := total / size
	if total%size != 0 {
		totalPages++
	}
	return models.PaginatedResponse{
		Data:       data,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")

	donors := api.Group("/donors")
	{
		donors.POST("", h.CreateDonor)
		donors.GET("", h.ListDonors)
		donors.GET("/:id", h.GetDonor)
		donors.GET("/:id/sub", h.GetDonorSub)
	}

	recipients := api.Group("/recipients")
	{
		recipients.POST("", h.CreateRecipient)
		recipients.GET("", h.ListRecipients)
		recipients.GET("/:id", h.GetRecipient)
		recipients.GET("/:id/sub", h.GetRecipientSub)
	}

	organs := api.Group("/organs")
	{
		organs.POST("/:id/assess", h.AssessOrgan)
		organs.POST("/:id/match", h.StartMatching)
		organs.POST("/:id/confirm", h.ConfirmMatch)
		organs.POST("/:id/acquire", h.AcquireOrgan)
	}

	transplants := api.Group("/transplants")
	{
		transplants.POST("", h.CreateTransplant)
		transplants.GET("", h.ListTransplants)
		transplants.GET("/:id", h.GetTransplant)
		transplants.PUT("/:id/status", h.UpdatePostOpStatus)
		transplants.POST("/:id/followup", h.AddFollowUp)
	}

	todos := api.Group("/todos")
	{
		todos.GET("", h.ListTodos)
		todos.PUT("/:id/status", h.UpdateTodoStatus)
	}

	locations := api.Group("/locations")
	{
		locations.POST("", h.CreateLocation)
		locations.GET("", h.ListLocations)
	}

	personnel := api.Group("/personnel")
	{
		personnel.POST("", h.CreatePersonnel)
		personnel.GET("", h.ListPersonnel)
	}

	reports := api.Group("/reports")
	{
		reports.POST("/generate", h.GenerateReport)
		reports.GET("", h.ListReports)
	}

	api.POST("/upgrade-urgency", h.UpgradeUrgency)
}

func (h *Handler) CreateDonor(c *gin.Context) {
	var donor models.Donor
	if err := c.ShouldBindJSON(&donor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RegisterDonor(context.Background(), &donor); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, donor)
}

func (h *Handler) ListDonors(c *gin.Context) {
	page, size := getPageAndSize(c)
	donors, total := h.svc.ListDonors(context.Background(), page, size)
	c.JSON(http.StatusOK, paginatedResponse(donors, page, size, total))
}

func (h *Handler) GetDonor(c *gin.Context) {
	id := c.Param("id")
	donor, exists := h.svc.GetDonor(context.Background(), id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "捐献者不存在"})
		return
	}
	c.JSON(http.StatusOK, donor)
}

func (h *Handler) GetDonorSub(c *gin.Context) {
	id := c.Param("id")
	donor, exists := h.svc.GetDonor(context.Background(), id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "捐献者不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"organs": donor.Organs})
}

func (h *Handler) CreateRecipient(c *gin.Context) {
	var recipient models.Recipient
	if err := c.ShouldBindJSON(&recipient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RegisterRecipient(context.Background(), &recipient); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, recipient)
}

func (h *Handler) ListRecipients(c *gin.Context) {
	page, size := getPageAndSize(c)
	recipients, total := h.svc.ListRecipients(context.Background(), page, size)
	c.JSON(http.StatusOK, paginatedResponse(recipients, page, size, total))
}

func (h *Handler) GetRecipient(c *gin.Context) {
	id := c.Param("id")
	recipient, exists := h.svc.GetRecipient(context.Background(), id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "受体不存在"})
		return
	}
	c.JSON(http.StatusOK, recipient)
}

func (h *Handler) GetRecipientSub(c *gin.Context) {
	id := c.Param("id")
	recipient, exists := h.svc.GetRecipient(context.Background(), id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "受体不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"matched_organ_id": recipient.MatchedOrganID})
}

type AssessRequest struct {
	FunctionScore         int  `json:"function_score"`
	HasVesselAbnormality  bool `json:"has_vessel_abnormality"`
	HasOtherLesions       bool `json:"has_other_lesions"`
}

func (h *Handler) AssessOrgan(c *gin.Context) {
	organID := c.Param("id")
	var req AssessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assessment := &models.OrganAssessment{
		FunctionScore:        req.FunctionScore,
		HasVesselAbnormality: req.HasVesselAbnormality,
		HasOtherLesions:      req.HasOtherLesions,
	}

	if err := h.svc.AssessOrgan(context.Background(), organID, assessment); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "评估完成"})
}

func (h *Handler) StartMatching(c *gin.Context) {
	organID := c.Param("id")
	if err := h.svc.StartMatching(context.Background(), organID); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "匹配已启动"})
}

type ConfirmRequest struct {
	RecipientID string `json:"recipient_id"`
	Confirmed   bool   `json:"confirmed"`
}

func (h *Handler) ConfirmMatch(c *gin.Context) {
	organID := c.Param("id")
	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ConfirmMatch(context.Background(), organID, req.RecipientID, req.Confirmed); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "确认完成"})
}

type AcquireRequest struct {
	ColdIschemiaHours int `json:"cold_ischemia_hours"`
}

func (h *Handler) AcquireOrgan(c *gin.Context) {
	organID := c.Param("id")
	var req AcquireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.AcquireOrgan(context.Background(), organID, req.ColdIschemiaHours); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "器官已获取"})
}

func (h *Handler) CreateTransplant(c *gin.Context) {
	var transplant models.TransplantRecord
	if err := c.ShouldBindJSON(&transplant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.CreateTransplant(context.Background(), &transplant); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, transplant)
}

func (h *Handler) ListTransplants(c *gin.Context) {
	page, size := getPageAndSize(c)
	transplants, total := h.svc.ListTransplants(context.Background(), page, size)
	c.JSON(http.StatusOK, paginatedResponse(transplants, page, size, total))
}

func (h *Handler) GetTransplant(c *gin.Context) {
	id := c.Param("id")
	transplant, exists := h.svc.GetTransplant(context.Background(), id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "移植记录不存在"})
		return
	}
	c.JSON(http.StatusOK, transplant)
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdatePostOpStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdatePostOpStatus(context.Background(), id, models.PostOpStatus(req.Status)); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

func (h *Handler) AddFollowUp(c *gin.Context) {
	transplantID := c.Param("id")
	var followUp models.FollowUp
	if err := c.ShouldBindJSON(&followUp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	followUp.TransplantID = transplantID

	if err := h.svc.AddFollowUp(context.Background(), &followUp); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, followUp)
}

func (h *Handler) ListTodos(c *gin.Context) {
	page, size := getPageAndSize(c)
	todos, total := h.svc.ListTodos(context.Background(), page, size)
	c.JSON(http.StatusOK, paginatedResponse(todos, page, size, total))
}

func (h *Handler) UpdateTodoStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateTodoStatus(context.Background(), id, models.TodoStatus(req.Status)); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "待办状态已更新"})
}

func (h *Handler) CreateLocation(c *gin.Context) {
	var location models.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.svc.CreateLocation(context.Background(), &location)
	c.JSON(http.StatusCreated, location)
}

func (h *Handler) ListLocations(c *gin.Context) {
	locations := h.svc.ListLocations(context.Background())
	c.JSON(http.StatusOK, locations)
}

func (h *Handler) CreatePersonnel(c *gin.Context) {
	var personnel models.Personnel
	if err := c.ShouldBindJSON(&personnel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.svc.CreatePersonnel(context.Background(), &personnel)
	c.JSON(http.StatusCreated, personnel)
}

func (h *Handler) ListPersonnel(c *gin.Context) {
	personnel := h.svc.ListPersonnel(context.Background())
	c.JSON(http.StatusOK, personnel)
}

func (h *Handler) GenerateReport(c *gin.Context) {
	report := h.svc.GenerateWeeklyReport(context.Background())
	c.JSON(http.StatusCreated, report)
}

func (h *Handler) ListReports(c *gin.Context) {
	reports := h.svc.ListReports(context.Background())
	c.JSON(http.StatusOK, reports)
}

func (h *Handler) UpgradeUrgency(c *gin.Context) {
	h.svc.UpgradeUrgencyLevel(context.Background())
	c.JSON(http.StatusOK, gin.H{"message": "紧急程度升级检查完成"})
}
