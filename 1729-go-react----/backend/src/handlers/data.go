package handlers

import (
	"net/http"
	"research-collaboration/src/models"
	"research-collaboration/src/storage"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	store *storage.Storage
}

func NewDataHandler(store *storage.Storage) *DataHandler {
	return &DataHandler{store: store}
}

func (h *DataHandler) UploadData(c *gin.Context) {
	var req struct {
		Name        string           `json:"name" binding:"required"`
		Description string           `json:"description"`
		Format      models.DataFormat `json:"format" binding:"required"`
		FileSize    int64            `json:"file_size" binding:"required"`
		ProjectID   string           `json:"project_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, ok := h.store.GetProject(req.ProjectID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "user-1"
	}

	data := &models.SharedData{
		Name:        req.Name,
		Description: req.Description,
		Format:      req.Format,
		FileSize:    req.FileSize,
		ProjectID:   req.ProjectID,
		UploaderID:  userID,
		CreatedAt:   time.Now(),
	}

	h.store.CreateSharedData(data)
	c.JSON(http.StatusCreated, data)
}

func (h *DataHandler) GetData(c *gin.Context) {
	id := c.Param("id")
	data, ok := h.store.GetSharedData(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "data not found"})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "user-1"
	}

	if !h.store.IsProjectMember(userID, data.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a project member"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *DataHandler) ListData(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "user-1"
	}

	projectID := c.Query("project_id")
	search := strings.ToLower(c.Query("search"))

	var dataList []*models.SharedData
	if projectID != "" {
		dataList = h.store.GetSharedDataByProject(projectID)
	} else {
		dataList = h.store.GetAllSharedData()
	}

	filtered := []*models.SharedData{}
	for _, d := range dataList {
		if !h.store.IsProjectMember(userID, d.ProjectID) {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(d.Name), search) &&
			!strings.Contains(strings.ToLower(d.Description), search) {
			continue
		}
		filtered = append(filtered, d)
	}

	c.JSON(http.StatusOK, filtered)
}

func (h *DataHandler) DownloadData(c *gin.Context) {
	id := c.Param("id")
	data, ok := h.store.GetSharedData(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "data not found"})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "user-1"
	}

	if !h.store.IsProjectMember(userID, data.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a project member"})
		return
	}

	log := &models.DownloadLog{
		DataID:     data.ID,
		UserID:     userID,
		DownloadedAt: time.Now(),
	}
	h.store.CreateDownloadLog(log)

	todo := &models.TodoItem{
		UserID:    userID,
		Action:    "downloaded_data",
		RefID:     data.ID,
		RefType:   "shared_data",
		CreatedAt: time.Now(),
	}
	h.store.CreateTodo(todo)

	c.JSON(http.StatusOK, gin.H{
		"message": "download recorded",
		"data":    data,
		"log_id":  log.ID,
	})
}

func (h *DataHandler) GetDownloadLogs(c *gin.Context) {
	dataID := c.Param("id")
	_, ok := h.store.GetSharedData(dataID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "data not found"})
		return
	}

	logs := h.store.GetDownloadLogsByData(dataID)
	c.JSON(http.StatusOK, logs)
}
