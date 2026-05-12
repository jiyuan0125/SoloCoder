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

type EnrollSubjectRequest struct {
	ProtocolID        string    `json:"protocol_id" binding:"required"`
	SiteID            string    `json:"site_id"`
	SiteCode          string    `json:"site_code"`
	ScreeningNumber   string    `json:"screening_number" binding:"required"`
	NameInitials      string    `json:"name_initials" binding:"required"`
	Gender            string    `json:"gender" binding:"required"`
	BirthDate         time.Time `json:"birth_date" binding:"required"`
	EnrollmentDate    *time.Time `json:"enrollment_date"`
	InitialStatus     string    `json:"initial_status"`
}

type UpdateSubjectStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func CreateSubject(c *gin.Context) {
	var req EnrollSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	protocolID, err := uuid.Parse(req.ProtocolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方案ID"})
		return
	}

	protocol, err := services.GetProtocolByID(protocolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var siteID uuid.UUID
	if req.SiteID != "" {
		siteID, err = uuid.Parse(req.SiteID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的中心ID"})
			return
		}
	} else if req.SiteCode != "" {
		site, err := services.GetSiteByCode(protocolID, req.SiteCode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "中心不存在"})
			return
		}
		siteID = site.ID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要提供中心ID或中心编码"})
		return
	}

	site, err := services.GetSiteByID(siteID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "中心不存在"})
		return
	}

	seqNum, err := services.GetNextSeqNumber(siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	randomID := utils.GenerateRandomizationID(site.SiteCode, seqNum)

	if services.IsDuplicateRandomID(randomID) {
		c.JSON(http.StatusConflict, gin.H{"error": "受试者编号已存在"})
		return
	}

	count, _ := services.CountSubjectsByProtocol(protocolID)
	group := utils.GenerateNextGroup(protocol.GroupRatio, int(count))

	status := models.SubjectScreening
	if req.InitialStatus != "" {
		status = models.SubjectStatus(req.InitialStatus)
	}

	enrollmentDate := req.EnrollmentDate
	if enrollmentDate == nil && (status == models.SubjectEnrolled || status == models.SubjectTreating || status == models.SubjectFollowup) {
		now := time.Now()
		enrollmentDate = &now
	}

	subject := &models.Subject{
		ProtocolID:      protocolID,
		SiteID:          siteID,
		ScreeningNumber: req.ScreeningNumber,
		RandomizationID: randomID,
		NameInitials:    req.NameInitials,
		Gender:          req.Gender,
		BirthDate:       req.BirthDate,
		EnrollmentDate:  enrollmentDate,
		GroupAssignment: group,
		Status:          status,
		SeqNumber:       seqNum,
	}

	if err := services.CreateSubject(subject); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subject)
}

func ListSubjects(c *gin.Context) {
	protocolID := c.Query("protocol_id")
	siteID := c.Query("site_id")

	var subjects []models.Subject
	var err error

	if protocolID != "" {
		pID, err := uuid.Parse(protocolID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方案ID"})
			return
		}
		var sID uuid.UUID
		if siteID != "" {
			sID, _ = uuid.Parse(siteID)
		}
		subjects, err = services.GetSubjectsWithProtocolSite(pID, sID)
	} else {
		subjects, err = services.GetAllSubjectsWithDetails()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subjects)
}

func GetSubject(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	subject, err := services.GetSubjectByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "受试者不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subject)
}

func UpdateSubjectStatus(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	subject, err := services.GetSubjectByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "受试者不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req UpdateSubjectStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !utils.CanTransitionStatus(string(subject.Status), req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态流转: " + string(subject.Status) + " -> " + req.Status})
		return
	}

	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	if req.Status == string(models.SubjectEnrolled) && subject.EnrollmentDate == nil {
		now := time.Now()
		updates["enrollment_date"] = &now
	}

	if err := services.UpdateSubject(uuidID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, _ := services.GetSubjectByID(uuidID)
	c.JSON(http.StatusOK, updated)
}

func WithdrawSubject(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	subject, err := services.GetSubjectByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "受试者不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subject.Status == models.SubjectCompleted || subject.Status == models.SubjectWithdrawn {
		c.JSON(http.StatusBadRequest, gin.H{"error": "受试者已完成或已退出"})
		return
	}

	updates := map[string]interface{}{
		"status":     models.SubjectWithdrawn,
		"updated_at": time.Now(),
	}

	if err := services.UpdateSubject(uuidID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "受试者已退出试验", "randomization_id": subject.RandomizationID})
}
