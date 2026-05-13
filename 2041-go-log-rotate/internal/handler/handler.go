package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"log-rotate/internal/logger"
	"log-rotate/internal/repository"
)

type Handler struct {
	repo    *repository.Repository
	writer  *logger.LogWriter
	querier *logger.LogQuerier
}

func NewHandler(repo *repository.Repository, writer *logger.LogWriter, querier *logger.LogQuerier) *Handler {
	return &Handler{
		repo:    repo,
		writer:  writer,
		querier: querier,
	}
}

type WriteLogRequest struct {
	Level   string `json:"level" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type UpdateConfigRequest struct {
	MaxFileSizeMB   *int    `json:"max_file_size_mb"`
	RotateHour      *int    `json:"rotate_hour"`
	RotateMinute    *int    `json:"rotate_minute"`
	MaxArchiveFiles *int    `json:"max_archive_files"`
	LogDir          *string `json:"log_dir"`
}

func (h *Handler) WriteLog(c *gin.Context) {
	var req WriteLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Level == "" {
		req.Level = "INFO"
	}

	if err := h.writer.Write(req.Level, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetConfig(c *gin.Context) {
	cfg, err := h.repo.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.repo.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.MaxFileSizeMB != nil && *req.MaxFileSizeMB > 0 {
		cfg.MaxFileSizeMB = *req.MaxFileSizeMB
	}
	if req.RotateHour != nil && *req.RotateHour >= 0 && *req.RotateHour < 24 {
		cfg.RotateHour = *req.RotateHour
	}
	if req.RotateMinute != nil && *req.RotateMinute >= 0 && *req.RotateMinute < 60 {
		cfg.RotateMinute = *req.RotateMinute
	}
	if req.MaxArchiveFiles != nil && *req.MaxArchiveFiles > 0 {
		cfg.MaxArchiveFiles = *req.MaxArchiveFiles
	}
	if req.LogDir != nil && *req.LogDir != "" {
		cfg.LogDir = *req.LogDir
	}

	if err := h.repo.UpdateConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.writer.UpdateConfig(cfg)

	c.JSON(http.StatusOK, cfg)
}

func (h *Handler) GetStatus(c *gin.Context) {
	status := h.writer.GetStatus()
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetArchives(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "30")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 30
	}

	archives, err := h.repo.GetArchives(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"archives": archives,
		"count":    len(archives),
	})
}

func (h *Handler) QueryCurrentLog(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	lines, err := h.querier.QueryCurrent(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lines": lines,
		"count": len(lines),
	})
}

func (h *Handler) QueryArchiveLog(c *gin.Context) {
	fileName := c.Param("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	lines, err := h.querier.QueryArchive(fileName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_name": fileName,
		"lines":     lines,
		"count":     len(lines),
	})
}

func (h *Handler) TriggerRotate(c *gin.Context) {
	type rotateRequest struct {
		Force bool `json:"force"`
	}

	var req rotateRequest
	c.ShouldBindJSON(&req)

	currentStatus := h.writer.GetStatus()
	fileSize := currentStatus["current_file_size"].(int64)

	if req.Force || fileSize > 0 {
		h.writer.TriggerRotate()
		c.JSON(http.StatusOK, gin.H{"status": "rotate triggered"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "skip: current file is empty"})
	}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/log", h.WriteLog)
		api.POST("/log/rotate", h.TriggerRotate)

		api.GET("/config", h.GetConfig)
		api.PUT("/config", h.UpdateConfig)

		api.GET("/status", h.GetStatus)
		api.GET("/archives", h.GetArchives)

		api.GET("/log/current", h.QueryCurrentLog)
		api.GET("/log/archive/:filename", h.QueryArchiveLog)
	}
}
