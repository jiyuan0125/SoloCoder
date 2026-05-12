package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"trial-management-system/internal/models"
	"trial-management-system/internal/services"
)

func CreateSystemA(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		ConfigInfo string `json:"config_info"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a := &models.SystemA{
		Name:       req.Name,
		ConfigInfo: req.ConfigInfo,
	}

	if err := services.CreateSystemA(a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

func ListSystemA(c *gin.Context) {
	items, err := services.ListSystemA()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func GetSystemA(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	a, err := services.GetSystemAByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "系统A不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)
}

func CreateSystemB(c *gin.Context) {
	var req struct {
		SystemAID string `json:"system_a_id" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Details   string `json:"details"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	aID, err := uuid.Parse(req.SystemAID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系统A ID"})
		return
	}

	if _, err := services.GetSystemAByID(aID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "关联的系统A不存在"})
		return
	}

	b := &models.SystemB{
		SystemAID: aID,
		Name:      req.Name,
		Details:   req.Details,
	}

	if err := services.CreateSystemB(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, b)
}

func ListSystemB(c *gin.Context) {
	systemAID := c.Query("system_a_id")

	var items []models.SystemB
	db := services.GetDB()

	if systemAID != "" {
		aID, err := uuid.Parse(systemAID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的系统A ID"})
			return
		}
		db = db.Where("system_a_id = ?", aID)
	}

	if err := db.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetSystemB(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	b, err := services.GetSystemBByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "系统B不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, b)
}
