package handlers

import (
	"net/http"
	"time"

	"hospital-infection/database"
	"hospital-infection/models"

	"github.com/gin-gonic/gin"
)

type CreateInfectionCaseRequest struct {
	PatientID       string                `json:"PatientID" binding:"required"`
	PatientName     string                `json:"PatientName" binding:"required"`
	Gender          string                `json:"Gender"`
	Age             int                   `json:"Age"`
	DepartmentID    uint                  `json:"DepartmentID" binding:"required"`
	AdmissionDate   time.Time             `json:"AdmissionDate" binding:"required"`
	InfectionDate   time.Time             `json:"InfectionDate" binding:"required"`
	InfectionSite   models.InfectionSite  `json:"InfectionSite" binding:"required"`
	Pathogen        string                `json:"Pathogen" binding:"required"`
	DrugSensitivity bool                  `json:"DrugSensitivity"`
	InfectionType   string                `json:"InfectionType"`
}

func CreateInfectionCase(c *gin.Context) {
	var req CreateInfectionCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PatientID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "患者住院号不能为空"})
		return
	}

	if req.InfectionDate.Before(req.AdmissionDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "感染日期不能早于入院日期"})
		return
	}

	if req.InfectionSite == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "感染部位不能为空"})
		return
	}

	if req.Pathogen == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "病原体不能为空"})
		return
	}

	if req.InfectionType != "" {
		automaticType := calculateInfectionType(req.AdmissionDate, req.InfectionDate)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "感染类型由系统自动判断，不能手动指定",
			"automatic_type": automaticType,
		})
		return
	}

	var department models.Department
	if err := database.DB.First(&department, req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "科室不存在"})
		return
	}

	caseItem := models.InfectionCase{
		PatientID:       req.PatientID,
		PatientName:     req.PatientName,
		Gender:          req.Gender,
		Age:             req.Age,
		DepartmentID:    req.DepartmentID,
		AdmissionDate:   req.AdmissionDate,
		InfectionDate:   req.InfectionDate,
		InfectionSite:   req.InfectionSite,
		InfectionType:   calculateInfectionType(req.AdmissionDate, req.InfectionDate),
		Pathogen:        req.Pathogen,
		DrugSensitivity: req.DrugSensitivity,
		Status:          "active",
	}

	if err := database.DB.Create(&caseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建感染病例失败"})
		return
	}

	database.DB.Preload("Department").First(&caseItem, caseItem.ID)
	c.JSON(http.StatusCreated, caseItem)
}

func GetInfectionCases(c *gin.Context) {
	var cases []models.InfectionCase
	query := database.DB.Preload("Department")

	departmentID := c.Query("DepartmentID")
	if departmentID == "" {
		departmentID = c.Query("department_id")
	}
	if departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}

	infectionType := c.Query("Type")
	if infectionType == "" {
		infectionType = c.Query("type")
	}
	if infectionType != "" {
		query = query.Where("infection_type = ?", infectionType)
	}

	if err := query.Order("created_at desc").Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取感染病例失败"})
		return
	}

	c.JSON(http.StatusOK, cases)
}

func GetInfectionCase(c *gin.Context) {
	id := c.Param("id")
	var caseItem models.InfectionCase
	if err := database.DB.Preload("Department").First(&caseItem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "感染病例不存在"})
		return
	}
	c.JSON(http.StatusOK, caseItem)
}

func UpdateInfectionCase(c *gin.Context) {
	id := c.Param("id")
	var caseItem models.InfectionCase
	if err := database.DB.First(&caseItem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "感染病例不存在"})
		return
	}

	var req CreateInfectionCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PatientID != "" {
		caseItem.PatientID = req.PatientID
	}
	if req.PatientName != "" {
		caseItem.PatientName = req.PatientName
	}
	caseItem.Gender = req.Gender
	caseItem.Age = req.Age
	if req.DepartmentID != 0 {
		caseItem.DepartmentID = req.DepartmentID
	}
	if !req.AdmissionDate.IsZero() {
		caseItem.AdmissionDate = req.AdmissionDate
	}
	if !req.InfectionDate.IsZero() {
		if req.InfectionDate.Before(caseItem.AdmissionDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "感染日期不能早于入院日期"})
			return
		}
		caseItem.InfectionDate = req.InfectionDate
		caseItem.InfectionType = calculateInfectionType(caseItem.AdmissionDate, caseItem.InfectionDate)
	}
	if req.InfectionSite != "" {
		caseItem.InfectionSite = req.InfectionSite
	}
	if req.Pathogen != "" {
		caseItem.Pathogen = req.Pathogen
	}
	caseItem.DrugSensitivity = req.DrugSensitivity

	if err := database.DB.Save(&caseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新感染病例失败"})
		return
	}

	database.DB.Preload("Department").First(&caseItem, caseItem.ID)
	c.JSON(http.StatusOK, caseItem)
}

func DeleteInfectionCase(c *gin.Context) {
	id := c.Param("id")
	var caseItem models.InfectionCase
	if err := database.DB.First(&caseItem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "感染病例不存在"})
		return
	}

	database.DB.Where("infection_case_id = ?", id).Delete(&models.PreventionMeasure{})

	if err := database.DB.Delete(&caseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除感染病例失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func calculateInfectionType(admissionDate, infectionDate time.Time) models.InfectionType {
	diff := infectionDate.Sub(admissionDate)
	if diff.Hours() >= 48 {
		return models.InfectionTypeNosocomial
	}
	return models.InfectionTypeCommunity
}
