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
)

type CreateProtocolRequest struct {
	ProtocolNumber    string    `json:"protocol_number" binding:"required"`
	DrugName          string    `json:"drug_name" binding:"required"`
	Indication        string    `json:"indication" binding:"required"`
	TrialPhase        string    `json:"trial_phase" binding:"required"`
	PlannedEnrollment int       `json:"planned_enrollment" binding:"required"`
	StartDate         time.Time `json:"start_date" binding:"required"`
	EndDate           time.Time `json:"end_date" binding:"required"`
	InclusionCriteria string    `json:"inclusion_criteria"`
	ExclusionCriteria string    `json:"exclusion_criteria"`
	Status            string    `json:"status"`
	GroupRatio        string    `json:"group_ratio"`
	Sites             []struct {
		SiteCode              string `json:"site_code" binding:"required"`
		SiteName              string `json:"site_name" binding:"required"`
		PrincipalInvestigator string `json:"principal_investigator" binding:"required"`
		PlannedEnrollment     int    `json:"planned_enrollment" binding:"required"`
	} `json:"sites"`
	Visits []struct {
		VisitName       string `json:"visit_name" binding:"required"`
		VisitOrder      int    `json:"visit_order" binding:"required"`
		WindowDays      int    `json:"window_days"`
		WindowTolerance int    `json:"window_tolerance"`
	} `json:"visits"`
}

type UpdateProtocolRequest struct {
	DrugName          string    `json:"drug_name"`
	Indication        string    `json:"indication"`
	TrialPhase        string    `json:"trial_phase"`
	PlannedEnrollment int       `json:"planned_enrollment"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	InclusionCriteria string    `json:"inclusion_criteria"`
	ExclusionCriteria string    `json:"exclusion_criteria"`
	Status            string    `json:"status"`
	GroupRatio        string    `json:"group_ratio"`
}

func CreateProtocol(c *gin.Context) {
	var req CreateProtocolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if services.IsDuplicateProtocol(req.ProtocolNumber) {
		c.JSON(http.StatusConflict, gin.H{"error": "方案编号已存在"})
		return
	}

	status := models.ProtocolPreparing
	if req.Status != "" {
		status = models.ProtocolStatus(req.Status)
	}

	groupRatio := "1:1"
	if req.GroupRatio != "" {
		groupRatio = req.GroupRatio
	}

	protocol := &models.Protocol{
		ProtocolNumber:    req.ProtocolNumber,
		DrugName:          req.DrugName,
		Indication:        req.Indication,
		TrialPhase:        models.TrialPhase(req.TrialPhase),
		PlannedEnrollment: req.PlannedEnrollment,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		InclusionCriteria: req.InclusionCriteria,
		ExclusionCriteria: req.ExclusionCriteria,
		Status:            status,
		GroupRatio:        groupRatio,
	}

	if err := services.CreateProtocol(protocol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, s := range req.Sites {
		site := &models.Site{
			ProtocolID:            protocol.ID,
			SiteCode:              s.SiteCode,
			SiteName:              s.SiteName,
			PrincipalInvestigator: s.PrincipalInvestigator,
			PlannedEnrollment:     s.PlannedEnrollment,
		}
		services.CreateSite(site)
	}

	for _, v := range req.Visits {
		visit := &models.Visit{
			ProtocolID:      protocol.ID,
			VisitName:       v.VisitName,
			VisitOrder:      v.VisitOrder,
			WindowDays:      v.WindowDays,
			WindowTolerance: v.WindowTolerance,
		}
		services.CreateVisit(visit)
	}

	created, _ := services.GetProtocolByID(protocol.ID)
	c.JSON(http.StatusCreated, created)
}

func ListProtocols(c *gin.Context) {
	protocols, err := services.ListProtocols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, protocols)
}

func GetProtocol(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	protocol, err := services.GetProtocolByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, protocol)
}

func UpdateProtocol(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !services.ProtocolExists(uuidID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
		return
	}

	var req UpdateProtocolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.DrugName != "" {
		updates["drug_name"] = req.DrugName
	}
	if req.Indication != "" {
		updates["indication"] = req.Indication
	}
	if req.TrialPhase != "" {
		updates["trial_phase"] = req.TrialPhase
	}
	if req.PlannedEnrollment > 0 {
		updates["planned_enrollment"] = req.PlannedEnrollment
	}
	if !req.StartDate.IsZero() {
		updates["start_date"] = req.StartDate
	}
	if !req.EndDate.IsZero() {
		updates["end_date"] = req.EndDate
	}
	if req.InclusionCriteria != "" {
		updates["inclusion_criteria"] = req.InclusionCriteria
	}
	if req.ExclusionCriteria != "" {
		updates["exclusion_criteria"] = req.ExclusionCriteria
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.GroupRatio != "" {
		updates["group_ratio"] = req.GroupRatio
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := services.UpdateProtocol(uuidID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	protocol, _ := services.GetProtocolByID(uuidID)
	c.JSON(http.StatusOK, protocol)
}

func AddSiteToProtocol(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !services.ProtocolExists(uuidID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
		return
	}

	var siteReq struct {
		SiteCode              string `json:"site_code" binding:"required"`
		SiteName              string `json:"site_name" binding:"required"`
		PrincipalInvestigator string `json:"principal_investigator" binding:"required"`
		PlannedEnrollment     int    `json:"planned_enrollment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&siteReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	site := &models.Site{
		ProtocolID:            uuidID,
		SiteCode:              siteReq.SiteCode,
		SiteName:              siteReq.SiteName,
		PrincipalInvestigator: siteReq.PrincipalInvestigator,
		PlannedEnrollment:     siteReq.PlannedEnrollment,
	}

	if err := services.CreateSite(site); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, site)
}

func AddVisitToProtocol(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if !services.ProtocolExists(uuidID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
		return
	}

	var visitReq struct {
		VisitName       string `json:"visit_name" binding:"required"`
		VisitOrder      int    `json:"visit_order" binding:"required"`
		WindowDays      int    `json:"window_days"`
		WindowTolerance int    `json:"window_tolerance"`
	}
	if err := c.ShouldBindJSON(&visitReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	visit := &models.Visit{
		ProtocolID:      uuidID,
		VisitName:       visitReq.VisitName,
		VisitOrder:      visitReq.VisitOrder,
		WindowDays:      visitReq.WindowDays,
		WindowTolerance: visitReq.WindowTolerance,
	}

	if err := services.CreateVisit(visit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, visit)
}
