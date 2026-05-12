package handlers

import (
	"math"
	"net/http"

	"hospital-infection/database"
	"hospital-infection/models"

	"github.com/gin-gonic/gin"
)

type CreateTargetMonitoringRequest struct {
	DepartmentID        uint   `json:"DepartmentID" binding:"required"`
	Month               string `json:"Month" binding:"required"`
	HospitalizationDays int    `json:"HospitalizationDays"`
	VentilatorDays      int    `json:"VentilatorDays"`
	VAPCases            int    `json:"VAPCases"`
	CentralLineDays     int    `json:"CentralLineDays"`
	CLABSICases         int    `json:"CLABSICases"`
	CatheterDays        int    `json:"CatheterDays"`
	CAUTICases          int    `json:"CAUTICases"`
}

type TargetMonitoringIndicators struct {
	ID                   uint    `json:"ID"`
	DepartmentID         uint    `json:"DepartmentID"`
	DepartmentName       string  `json:"DepartmentName"`
	Month                string  `json:"Month"`
	HospitalizationDays  int     `json:"HospitalizationDays"`
	VentilatorDays       int     `json:"VentilatorDays"`
	VAPCases             int     `json:"VAPCases"`
	CentralLineDays      int     `json:"CentralLineDays"`
	CLABSICases          int     `json:"CLABSICases"`
	CatheterDays         int     `json:"CatheterDays"`
	CAUTICases           int     `json:"CAUTICases"`
	VentilatorUsageRate  float64 `json:"VentilatorUsageRate"`
	CentralLineUsageRate float64 `json:"CentralLineUsageRate"`
	CatheterUsageRate    float64 `json:"CatheterUsageRate"`
	VAPRate              float64 `json:"VAPRate"`
	CLABSIRate           float64 `json:"CLABSIRate"`
	CAUTIRate            float64 `json:"CAUTIRate"`
}

func CreateTargetMonitoring(c *gin.Context) {
	var req CreateTargetMonitoringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.VentilatorDays < 0 || req.CentralLineDays < 0 || req.CatheterDays < 0 || req.HospitalizationDays < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "器械使用天数和住院天数不能为负"})
		return
	}

	var department models.Department
	if err := database.DB.First(&department, req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "科室不存在"})
		return
	}

	var existing models.TargetMonitoring
	database.DB.Where("department_id = ? AND month = ?", req.DepartmentID, req.Month).First(&existing)

	monitoring := models.TargetMonitoring{
		DepartmentID:        req.DepartmentID,
		Month:               req.Month,
		HospitalizationDays: req.HospitalizationDays,
		VentilatorDays:      req.VentilatorDays,
		VAPCases:            req.VAPCases,
		CentralLineDays:     req.CentralLineDays,
		CLABSICases:         req.CLABSICases,
		CatheterDays:        req.CatheterDays,
		CAUTICases:          req.CAUTICases,
	}

	if existing.ID > 0 {
		existing.HospitalizationDays = req.HospitalizationDays
		existing.VentilatorDays = req.VentilatorDays
		existing.VAPCases = req.VAPCases
		existing.CentralLineDays = req.CentralLineDays
		existing.CLABSICases = req.CLABSICases
		existing.CatheterDays = req.CatheterDays
		existing.CAUTICases = req.CAUTICases
		database.DB.Save(&existing)
		monitoring = existing
	} else {
		if err := database.DB.Create(&monitoring).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目标性监测失败"})
			return
		}
	}

	database.DB.Preload("Department").First(&monitoring, monitoring.ID)
	c.JSON(http.StatusCreated, calculateIndicators(monitoring))
}

func GetTargetMonitorings(c *gin.Context) {
	var monitorings []models.TargetMonitoring
	query := database.DB.Preload("Department")

	departmentID := c.Query("department_id")
	if departmentID == "" {
		departmentID = c.Query("DepartmentID")
	}
	if departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}

	month := c.Query("month")
	if month == "" {
		month = c.Query("Month")
	}
	if month != "" {
		query = query.Where("month = ?", month)
	}

	query.Order("month desc").Find(&monitorings)

	var results []TargetMonitoringIndicators
	for _, m := range monitorings {
		results = append(results, calculateIndicators(m))
	}

	c.JSON(http.StatusOK, results)
}

func GetTargetMonitoring(c *gin.Context) {
	id := c.Param("id")
	var monitoring models.TargetMonitoring
	if err := database.DB.Preload("Department").First(&monitoring, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标性监测记录不存在"})
		return
	}
	c.JSON(http.StatusOK, calculateIndicators(monitoring))
}

func DeleteTargetMonitoring(c *gin.Context) {
	id := c.Param("id")
	var monitoring models.TargetMonitoring
	if err := database.DB.First(&monitoring, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标性监测记录不存在"})
		return
	}

	if err := database.DB.Delete(&monitoring).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除目标性监测失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func calculateIndicators(m models.TargetMonitoring) TargetMonitoringIndicators {
	result := TargetMonitoringIndicators{
		ID:                  m.ID,
		DepartmentID:        m.DepartmentID,
		Month:               m.Month,
		HospitalizationDays: m.HospitalizationDays,
		VentilatorDays:      m.VentilatorDays,
		VAPCases:            m.VAPCases,
		CentralLineDays:     m.CentralLineDays,
		CLABSICases:         m.CLABSICases,
		CatheterDays:        m.CatheterDays,
		CAUTICases:          m.CAUTICases,
	}

	if m.Department.Name != "" {
		result.DepartmentName = m.Department.Name
	}

	if m.HospitalizationDays > 0 {
		result.VentilatorUsageRate = math.Round(float64(m.VentilatorDays)/float64(m.HospitalizationDays)*10000) / 100
		result.CentralLineUsageRate = math.Round(float64(m.CentralLineDays)/float64(m.HospitalizationDays)*10000) / 100
		result.CatheterUsageRate = math.Round(float64(m.CatheterDays)/float64(m.HospitalizationDays)*10000) / 100
	}

	if m.VentilatorDays > 0 {
		result.VAPRate = math.Round(float64(m.VAPCases)/float64(m.VentilatorDays)*100000) / 100
	}
	if m.CentralLineDays > 0 {
		result.CLABSIRate = math.Round(float64(m.CLABSICases)/float64(m.CentralLineDays)*100000) / 100
	}
	if m.CatheterDays > 0 {
		result.CAUTIRate = math.Round(float64(m.CAUTICases)/float64(m.CatheterDays)*100000) / 100
	}

	return result
}

func GetTargetDepartments(c *gin.Context) {
	targetDepts := []string{"ICU", "新生儿科", "烧伤科", "血液科"}
	var departments []models.Department
	database.DB.Where("name IN ?", targetDepts).Find(&departments)
	c.JSON(http.StatusOK, departments)
}
