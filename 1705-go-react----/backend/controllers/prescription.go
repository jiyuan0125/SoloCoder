package controllers

import (
	"fmt"
	"net/http"
	"telemedicine/config"
	"telemedicine/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PrescriptionItemRequest struct {
	DrugName      string  `json:"drug_name" binding:"required"`
	Specification string  `json:"specification" binding:"required"`
	Usage         string  `json:"usage" binding:"required"`
	Dosage        string  `json:"dosage" binding:"required"`
	Days          int     `json:"days" binding:"required"`
	UnitPrice     float64 `json:"unit_price"`
	Quantity      int     `json:"quantity"`
}

type CreatePrescriptionRequest struct {
	ConsultationID uint                      `json:"consultation_id" binding:"required"`
	Items          []PrescriptionItemRequest `json:"items" binding:"required"`
}

func generatePrescriptionNo() string {
	now := time.Now()
	return fmt.Sprintf("PR%s%s", now.Format("20060102150405"), fmt.Sprintf("%04d", now.Nanosecond()/1000000))
}

func CreatePrescription(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	if userRole != "expert" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有专家可以开具处方"})
		return
	}

	var req CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "处方至少包含一种药品"})
		return
	}

	if len(req.Items) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每个处方的药品总数不能超过5种"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.First(&consultation, req.ConsultationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if consultation.ExpertDoctorID == nil || *consultation.ExpertDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限为此问诊开处方"})
		return
	}

	if consultation.Status != "in_progress" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能在问诊中开具处方"})
		return
	}

	for _, item := range req.Items {
		var forbidden models.ForbiddenDrug
		err := config.DB.Where("name = ?", item.DrugName).First(&forbidden).Error
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("药品'%s'是禁止药品，原因：%s", item.DrugName, forbidden.Reason),
			})
			return
		}
	}

	var totalAmount float64
	for _, item := range req.Items {
		if item.UnitPrice > 0 && item.Quantity > 0 {
			totalAmount += item.UnitPrice * float64(item.Quantity)
		}
	}

	if totalAmount > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "远程问诊处方金额不得超过200元，请调整"})
		return
	}

	tx := config.DB.Begin()

	prescription := models.Prescription{
		ConsultationID: consultation.ID,
		PrescriptionNo: generatePrescriptionNo(),
		TotalAmount:    totalAmount,
		Status:         "active",
		ExpiredAt:      time.Now().AddDate(0, 0, 3),
	}

	if err := tx.Create(&prescription).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建处方失败"})
		return
	}

	for _, item := range req.Items {
		amount := 0.0
		if item.UnitPrice > 0 && item.Quantity > 0 {
			amount = item.UnitPrice * float64(item.Quantity)
		}
		pItem := models.PrescriptionItem{
			PrescriptionID: prescription.ID,
			DrugName:       item.DrugName,
			Specification:  item.Specification,
			Usage:          item.Usage,
			Dosage:         item.Dosage,
			Days:           item.Days,
			UnitPrice:      item.UnitPrice,
			Quantity:       item.Quantity,
			Amount:         amount,
		}
		if err := tx.Create(&pItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加处方药品失败"})
			return
		}
	}

	tx.Commit()

	config.DB.Preload("Items").First(&prescription, prescription.ID)
	c.JSON(http.StatusCreated, prescription)
}

func GetPrescription(c *gin.Context) {
	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处方ID"})
		return
	}

	var prescription models.Prescription
	if err := config.DB.Preload("Items").First(&prescription, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, prescription)
}

func GetPrescriptionsByConsultation(c *gin.Context) {
	consultationID, err := parseInt(c, "consultationID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var prescriptions []models.Prescription
	config.DB.Preload("Items").Where("consultation_id = ?", consultationID).Find(&prescriptions)
	c.JSON(http.StatusOK, prescriptions)
}

func ListForbiddenDrugs(c *gin.Context) {
	var drugs []models.ForbiddenDrug
	config.DB.Find(&drugs)
	c.JSON(http.StatusOK, drugs)
}
