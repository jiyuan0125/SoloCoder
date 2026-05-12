package handlers

import (
	"net/http"
	"time"

	"health-supervision-system/internal/models"
	"health-supervision-system/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IssueOpinionRequest struct {
	InspectionID      uint   `json:"inspection_id" binding:"required"`
	IssueDate         string `json:"issue_date"`
	RectificationDays int    `json:"rectification_days" binding:"required"`
	Remarks           string `json:"remarks"`
}

type ReviewOpinionRequest struct {
	IsPassed bool   `json:"is_passed" binding:"required"`
	Remarks  string `json:"remarks"`
}

type IssuePenaltyRequest struct {
	OpinionID    uint     `json:"opinion_id" binding:"required"`
	PenaltyTypes string   `json:"penalty_types" binding:"required"`
	FineAmount   float64  `json:"fine_amount"`
	FineReason   string   `json:"fine_reason"`
	IssueDate    string   `json:"issue_date"`
	Remarks      string   `json:"remarks"`
}

func ListOpinions(c *gin.Context) {
	unitID := c.Query("unit_id")
	status := c.Query("status")

	var opinions []models.HealthOpinion
	query := storage.DB.Preload("InspectionRecord").Preload("InspectionRecord.Unit")

	if unitID != "" {
		query = query.Joins("JOIN inspection_records ON inspection_records.id = health_opinions.inspection_id").
			Where("inspection_records.unit_id = ?", unitID)
	}
	if status != "" {
		query = query.Where("health_opinions.status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&opinions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list opinions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": opinions})
}

func GetOpinion(c *gin.Context) {
	id := c.Param("id")
	var opinion models.HealthOpinion

	if err := storage.DB.Preload("InspectionRecord").Preload("InspectionRecord.Unit").First(&opinion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "opinion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get opinion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": opinion})
}

func IssueOpinion(c *gin.Context) {
	var req IssueOpinionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RectificationDays != 7 && req.RectificationDays != 15 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "整改期限必须是7天（一般问题）或15天（严重问题）"})
		return
	}

	var inspection models.InspectionRecord
	if err := storage.DB.Preload("Unit").First(&inspection, req.InspectionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "inspection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inspection"})
		return
	}

	if inspection.IsQualified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "检查合格的单位不需要下达整改意见书"})
		return
	}

	var existingOpinion models.HealthOpinion
	if err := storage.DB.Where("inspection_id = ?", req.InspectionID).First(&existingOpinion).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该检查记录已下达意见书"})
		return
	}

	issueDate := time.Now()
	if req.IssueDate != "" {
		parsed, err := time.Parse("2006-01-02", req.IssueDate)
		if err == nil {
			issueDate = parsed
		}
	}

	deadline := issueDate.AddDate(0, 0, req.RectificationDays)

	opinion := models.HealthOpinion{
		InspectionID:      req.InspectionID,
		IssueDate:         issueDate,
		RectificationDays: req.RectificationDays,
		Deadline:          deadline,
		Status:            models.RectificationPending,
		Remarks:           req.Remarks,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := storage.DB.Create(&opinion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create opinion"})
		return
	}

	opinionID := opinion.ID
	storage.CreateAuditLog(0, "system", "下达", "卫生监督意见书", &opinionID, "下达卫生监督意见书，单位: "+inspection.Unit.Name, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"data": opinion})
}

func ReviewOpinion(c *gin.Context) {
	id := c.Param("id")
	var opinion models.HealthOpinion

	if err := storage.DB.Preload("InspectionRecord").Preload("InspectionRecord.Unit").First(&opinion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "opinion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get opinion"})
		return
	}

	var req ReviewOpinionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := storage.DB.Begin()

	if req.IsPassed {
		opinion.Status = models.RectificationPassed
		unit := opinion.InspectionRecord.Unit
		unit.Status = models.StatusNormal
		if err := tx.Save(&unit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update unit status"})
			return
		}
	} else {
		opinion.Status = models.RectificationFailed
	}

	opinion.UpdatedAt = time.Now()

	if err := tx.Save(&opinion).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update opinion"})
		return
	}

	tx.Commit()

	opinionID := opinion.ID
	statusDesc := "整改合格"
	if !req.IsPassed {
		statusDesc = "整改不合格，进入行政处罚程序"
	}
	storage.CreateAuditLog(0, "system", "复查", "卫生监督意见书", &opinionID, "复查意见书结果: "+statusDesc, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"data": opinion})
}

func ListPenalties(c *gin.Context) {
	unitID := c.Query("unit_id")

	var penalties []models.Penalty
	query := storage.DB.Preload("Opinion").Preload("Opinion.InspectionRecord").Preload("Opinion.InspectionRecord.Unit")

	if unitID != "" {
		query = query.Where("unit_id = ?", unitID)
	}

	if err := query.Order("created_at DESC").Find(&penalties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list penalties"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": penalties})
}

func GetPenalty(c *gin.Context) {
	id := c.Param("id")
	var penalty models.Penalty

	if err := storage.DB.Preload("Opinion").Preload("Opinion.InspectionRecord").Preload("Opinion.InspectionRecord.Unit").First(&penalty, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "penalty not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get penalty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": penalty})
}

func IssuePenalty(c *gin.Context) {
	var req IssuePenaltyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var opinion models.HealthOpinion
	if err := storage.DB.Preload("InspectionRecord").Preload("InspectionRecord.Unit").First(&opinion, req.OpinionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "opinion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get opinion"})
		return
	}

	if opinion.Status != models.RectificationFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有整改失败的单位才能进行行政处罚"})
		return
	}

	if req.FineAmount > 0 {
		min, max, ok := models.GetFineRange(req.FineReason)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的罚款原因"})
			return
		}
		if req.FineAmount < min || req.FineAmount > max {
			c.JSON(http.StatusBadRequest, gin.H{"error": "罚款金额不在法定范围内"})
			return
		}
	}

	issueDate := time.Now()
	if req.IssueDate != "" {
		parsed, err := time.Parse("2006-01-02", req.IssueDate)
		if err == nil {
			issueDate = parsed
		}
	}

	tx := storage.DB.Begin()

	penalty := models.Penalty{
		OpinionID:    req.OpinionID,
		UnitID:       opinion.InspectionRecord.UnitID,
		PenaltyTypes: req.PenaltyTypes,
		FineAmount:   req.FineAmount,
		FineReason:   req.FineReason,
		IssueDate:    issueDate,
		Remarks:      req.Remarks,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := tx.Create(&penalty).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create penalty"})
		return
	}

	hasSuspend := false
	hasRevoke := false
	for _, pt := range []string{"停业整顿", "吊销许可证"} {
		if len(req.PenaltyTypes) > 0 && containsString(req.PenaltyTypes, pt) {
			if pt == "停业整顿" {
				hasSuspend = true
			}
			if pt == "吊销许可证" {
				hasRevoke = true
			}
		}
	}

	unit := opinion.InspectionRecord.Unit
	if hasRevoke {
		unit.Status = models.StatusCancelled
	} else if hasSuspend {
		unit.Status = models.StatusSuspended
	}

	if err := tx.Save(&unit).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update unit status"})
		return
	}

	tx.Commit()

	penaltyID := penalty.ID
	storage.CreateAuditLog(0, "system", "下达", "行政处罚", &penaltyID, "下达行政处罚，单位: "+unit.Name+"，处罚类型: "+req.PenaltyTypes, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"data": penalty})
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func GetFineRanges(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": models.PenaltyFineRanges})
}
