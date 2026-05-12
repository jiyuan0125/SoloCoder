package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"trial-management-system/internal/models"
	"trial-management-system/internal/services"
	"trial-management-system/internal/utils"
)

type RecordVisitRequest struct {
	ProtocolID string    `json:"protocol_id" binding:"required"`
	SubjectID  string    `json:"subject_id" binding:"required"`
	VisitID    string    `json:"visit_id" binding:"required"`
	ActualDate time.Time `json:"actual_date" binding:"required"`
	VitalSigns string    `json:"vital_signs"`
	LabTests   string    `json:"lab_tests"`
	OtherData  string    `json:"other_data"`
}

func RecordVisit(c *gin.Context) {
	var req RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	protocolID, err := uuid.Parse(req.ProtocolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方案ID"})
		return
	}

	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的受试者ID"})
		return
	}

	visitID, err := uuid.Parse(req.VisitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的访视ID"})
		return
	}

	if !services.ProtocolExists(protocolID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
		return
	}

	if !services.SubjectExists(subjectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "受试者不存在"})
		return
	}

	subject, err := services.GetSubjectByID(subjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "受试者不存在"})
		return
	}

	visit, err := services.GetVisitByID(visitID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "访视不存在"})
		return
	}

	if subject.ProtocolID != protocolID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "受试者不属于该方案"})
		return
	}

	if visit.ProtocolID != protocolID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "访视不属于该方案"})
		return
	}

	if services.IsDuplicateVisitRecord(subjectID, visitID) {
		c.JSON(http.StatusConflict, gin.H{"error": "该访视已录入"})
		return
	}

	isOutOfWindow := false
	if subject.EnrollmentDate != nil && !subject.EnrollmentDate.IsZero() {
		isOutOfWindow = utils.IsOutOfWindow(req.ActualDate, visit.WindowDays, visit.WindowTolerance, *subject.EnrollmentDate)
	}

	record := &models.VisitRecord{
		SubjectID:     subjectID,
		VisitID:       visitID,
		ActualDate:    req.ActualDate,
		IsOutOfWindow: isOutOfWindow,
		VitalSigns:    req.VitalSigns,
		LabTests:      req.LabTests,
		OtherData:     req.OtherData,
	}

	if err := services.CreateVisitRecord(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func ListVisitRecords(c *gin.Context) {
	subjectIDStr := c.Query("subject_id")
	visitIDStr := c.Query("visit_id")

	var records []models.VisitRecord
	db := services.GetDB()

	if subjectIDStr != "" {
		subjectID, err := uuid.Parse(subjectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的受试者ID"})
			return
		}
		db = db.Where("subject_id = ?", subjectID)
	}

	if visitIDStr != "" {
		visitID, err := uuid.Parse(visitIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的访视ID"})
			return
		}
		db = db.Where("visit_id = ?", visitID)
	}

	if err := db.Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func GetVisitRecord(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var record models.VisitRecord
	err = services.GetDB().Where("id = ?", uuidID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "访视记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}
