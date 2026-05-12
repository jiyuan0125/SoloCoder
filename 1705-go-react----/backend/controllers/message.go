package controllers

import (
	"net/http"
	"telemedicine/config"
	"telemedicine/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func SendMessage(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")
	userName := c.GetString("userName")

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.First(&consultation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if consultation.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "问诊已结束不能再发消息"})
		return
	}

	if userRole == "grassroot" && consultation.GrassrootDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
		return
	}
	if userRole == "expert" {
		if consultation.ExpertDoctorID == nil || *consultation.ExpertDoctorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
			return
		}
	}

	message := models.Message{
		ConsultationID: consultation.ID,
		SenderID:       userID,
		SenderRole:     userRole,
		SenderName:     userName,
		Content:        req.Content,
	}

	if err := config.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送消息失败"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

func ListMessages(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.First(&consultation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if userRole == "grassroot" && consultation.GrassrootDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看"})
		return
	}
	if userRole == "expert" {
		if consultation.ExpertDoctorID == nil || *consultation.ExpertDoctorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看"})
			return
		}
	}

	var messages []models.Message
	config.DB.Where("consultation_id = ?", id).Order("created_at ASC").Find(&messages)
	c.JSON(http.StatusOK, messages)
}
