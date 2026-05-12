package handlers

import (
	"net/http"
	"time"

	"health-supervision-system/internal/models"
	"health-supervision-system/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateNoticeRequest struct {
	UnitID      uint   `json:"unit_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	NoticeType  string `json:"notice_type" binding:"required"`
	DurationDays int   `json:"duration_days"`
	RelatedID   *uint  `json:"related_id"`
}

func ListNotices(c *gin.Context) {
	status := c.Query("status")
	unitID := c.Query("unit_id")

	var notices []models.PublicNotice
	query := storage.DB.Preload("Unit")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if unitID != "" {
		query = query.Where("unit_id = ?", unitID)
	}

	if err := query.Order("created_at DESC").Find(&notices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notices})
}

func GetNotice(c *gin.Context) {
	id := c.Param("id")
	var notice models.PublicNotice

	if err := storage.DB.Preload("Unit").First(&notice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "notice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notice})
}

func PublishNotice(c *gin.Context) {
	var req CreateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var unit models.SupervisedUnit
	if err := storage.DB.First(&unit, req.UnitID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unit"})
		return
	}

	durationDays := 30
	if req.DurationDays > 0 {
		durationDays = req.DurationDays
	}

	now := time.Now()
	expiryDate := now.AddDate(0, 0, durationDays)

	notice := models.PublicNotice{
		UnitID:      req.UnitID,
		Title:       req.Title,
		Content:     req.Content,
		NoticeType:  req.NoticeType,
		PublishDate: now,
		ExpiryDate:  expiryDate,
		Status:      models.NoticeStatusActive,
		RelatedID:   req.RelatedID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := storage.DB.Create(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create notice"})
		return
	}

	noticeID := notice.ID
	storage.CreateAuditLog(0, "system", "发布", "公示", &noticeID, "发布公示: "+notice.Title, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"data": notice})
}

func UpdateNoticeExpiry(c *gin.Context) {
	id := c.Param("id")
	var notice models.PublicNotice

	if err := storage.DB.First(&notice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "notice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notice"})
		return
	}

	var req struct {
		DurationDays int `json:"duration_days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notice.ExpiryDate = notice.PublishDate.AddDate(0, 0, req.DurationDays)
	notice.UpdatedAt = time.Now()

	if err := storage.DB.Save(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notice"})
		return
	}

	noticeID := notice.ID
	storage.CreateAuditLog(0, "system", "更新", "公示", &noticeID, "更新公示期限: "+notice.Title, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"data": notice})
}

func ExpireNotices() error {
	now := time.Now()

	result := storage.DB.Model(&models.PublicNotice{}).
		Where("status = ? AND expiry_date <= ?", models.NoticeStatusActive, now).
		Update("status", models.NoticeStatusHistory)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		storage.CreateAuditLog(0, "system", "自动更新", "公示", nil, "自动将过期公示移入历史公示，数量: "+string(rune(result.RowsAffected)), "system")
	}

	return nil
}

func CheckAndExpireNotices(c *gin.Context) {
	if err := ExpireNotices(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to expire notices"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "checked and expired notices"})
}

func DeleteNotice(c *gin.Context) {
	id := c.Param("id")
	var notice models.PublicNotice

	if err := storage.DB.First(&notice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "notice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notice"})
		return
	}

	if err := storage.DB.Delete(&notice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notice"})
		return
	}

	noticeID := notice.ID
	storage.CreateAuditLog(0, "system", "删除", "公示", &noticeID, "删除公示: "+notice.Title, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "notice deleted"})
}
