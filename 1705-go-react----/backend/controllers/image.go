package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"telemedicine/config"
	"telemedicine/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ImageDataRequest struct {
	TemplateID  uint            `json:"template_id" binding:"required"`
	ImageType   string          `json:"image_type" binding:"required"`
	FieldValues json.RawMessage `json:"field_values" binding:"required"`
}

func ListImageTemplates(c *gin.Context) {
	var templates []models.ImageTemplate
	config.DB.Where("is_active = ?", true).Find(&templates)
	c.JSON(http.StatusOK, templates)
}

func AddImageData(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	consultationID, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var req ImageDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.First(&consultation, consultationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if userRole == "grassroot" && consultation.GrassrootDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
		return
	}

	var template models.ImageTemplate
	if err := config.DB.First(&template, req.TemplateID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "影像模板不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询模板失败"})
		return
	}

	var fieldValues map[string]interface{}
	if err := json.Unmarshal(req.FieldValues, &fieldValues); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "字段值格式错误"})
		return
	}

	templateFields := strings.Split(template.Fields, ",")
	for field := range fieldValues {
		found := false
		for _, tf := range templateFields {
			if strings.TrimSpace(tf) == field {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未知的模板字段: " + field})
			return
		}
	}

	imageData := models.ImageData{
		ConsultationID: consultationID,
		TemplateID:     req.TemplateID,
		ImageType:      req.ImageType,
		FieldValues:    string(req.FieldValues),
	}

	if err := config.DB.Create(&imageData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存影像数据失败"})
		return
	}

	config.DB.Preload("Template").First(&imageData, imageData.ID)
	c.JSON(http.StatusCreated, imageData)
}
