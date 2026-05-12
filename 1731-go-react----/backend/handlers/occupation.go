package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"skill-cert/models"
	"skill-cert/storage"
)

type OccupationHandler struct {
	store *storage.Storage
}

func NewOccupationHandler(store *storage.Storage) *OccupationHandler {
	return &OccupationHandler{store: store}
}

type CreateOccupationRequest struct {
	Name     string                  `json:"name" binding:"required"`
	Code     string                  `json:"code" binding:"required"`
	Industry models.IndustryCategory `json:"industry" binding:"required"`
	Levels   []models.LevelConfig    `json:"levels" binding:"required"`
}

func (h *OccupationHandler) CreateOccupation(c *gin.Context) {
	var req CreateOccupationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	occ := &models.Occupation{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Code:     req.Code,
		Industry: req.Industry,
		Levels:   req.Levels,
	}

	h.store.SaveOccupation(occ)
	c.JSON(http.StatusCreated, occ)
}

func (h *OccupationHandler) GetOccupations(c *gin.Context) {
	occs := h.store.GetOccupations()
	c.JSON(http.StatusOK, occs)
}

func (h *OccupationHandler) GetOccupation(c *gin.Context) {
	id := c.Param("id")
	occ := h.store.GetOccupation(id)
	if occ == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "职业不存在"})
		return
	}
	c.JSON(http.StatusOK, occ)
}

type UpdateOccupationRequest struct {
	Name     string                  `json:"name"`
	Code     string                  `json:"code"`
	Industry models.IndustryCategory `json:"industry"`
	Levels   []models.LevelConfig    `json:"levels"`
}

func (h *OccupationHandler) UpdateOccupation(c *gin.Context) {
	id := c.Param("id")
	occ := h.store.GetOccupation(id)
	if occ == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "职业不存在"})
		return
	}

	var req UpdateOccupationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		occ.Name = req.Name
	}
	if req.Code != "" {
		occ.Code = req.Code
	}
	if req.Industry != "" {
		occ.Industry = req.Industry
	}
	if req.Levels != nil {
		occ.Levels = req.Levels
	}

	h.store.SaveOccupation(occ)
	c.JSON(http.StatusOK, occ)
}

func (h *OccupationHandler) DeleteOccupation(c *gin.Context) {
	id := c.Param("id")
	occ := h.store.GetOccupation(id)
	if occ == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "职业不存在"})
		return
	}

	h.store.DeleteOccupation(id)
	c.JSON(http.StatusNoContent, nil)
}
