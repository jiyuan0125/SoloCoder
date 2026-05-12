package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hospital-infection/database"
	"hospital-infection/models"

	"github.com/gin-gonic/gin"
)

type GenerateReportRequest struct {
	ReportType string `json:"report_type" binding:"required"`
	Month      string `json:"month"`
	Year       int    `json:"year"`
}

type ReportApprovalRequest struct {
	Operator string `json:"operator" binding:"required"`
	Comment  string `json:"comment"`
}

const (
	SmallAmountThreshold = 10000
)

func GenerateReport(c *gin.Context) {
	var req GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ReportType == "monthly" && req.Month == "" {
		now := time.Now()
		req.Month = fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	}

	if req.ReportType == "yearly" && req.Year == 0 {
		req.Year = time.Now().Year()
	}

	var report models.Report

	if req.ReportType == "monthly" {
		report = generateMonthlyReport(req.Month)
	} else {
		report = generateYearlyReport(req.Year)
	}

	if err := database.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建报告失败"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

func generateMonthlyReport(month string) models.Report {
	monthStart, _ := time.Parse("2006-01", month)
	year, _, _ := monthStart.Date()

	stats := calculateMonthlyStats(month)

	deptRatesJSON, _ := json.Marshal(stats.DepartmentRates)
	siteDistJSON, _ := json.Marshal(stats.SiteDistribution)
	pathogenDistJSON, _ := json.Marshal(stats.PathogenDistribution)

	return models.Report{
		ReportType:           "monthly",
		Month:                month,
		Year:                 year,
		HospitalRate:         stats.HospitalRate,
		DepartmentRates:      string(deptRatesJSON),
		SiteDistribution:     string(siteDistJSON),
		PathogenDistribution: string(pathogenDistJSON),
		AntibioticUsage:      "{}",
		ApprovalStatus:       models.ApprovalStatusSubmitted,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func generateYearlyReport(year int) models.Report {
	var monthlyReports []models.Report
	database.DB.Where("report_type = ? AND year = ?", "monthly", year).Find(&monthlyReports)

	var totalRate float64
	for _, r := range monthlyReports {
		totalRate += r.HospitalRate
	}

	var avgRate float64
	if len(monthlyReports) > 0 {
		avgRate = totalRate / float64(len(monthlyReports))
	}

	return models.Report{
		ReportType:           "yearly",
		Year:                 year,
		HospitalRate:         avgRate,
		DepartmentRates:      "[]",
		SiteDistribution:     "[]",
		PathogenDistribution: "[]",
		AntibioticUsage:      "{}",
		ApprovalStatus:       models.ApprovalStatusSubmitted,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func calculateMonthlyStats(month string) StatisticsResponse {
	monthStart, _ := time.Parse("2006-01", month)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var departments []models.Department
	database.DB.Find(&departments)

	threshold := getDefaultThreshold()

	var departmentRates []MonthlyRate
	totalInfections := 0
	totalDischarges := 0

	for _, dept := range departments {
		var infectionCount int64
		database.DB.Model(&models.InfectionCase{}).
			Where("department_id = ? AND infection_date >= ? AND infection_date < ? AND infection_type = ?",
				dept.ID, monthStart, monthEnd, models.InfectionTypeNosocomial).
			Count(&infectionCount)

		dischargeCount := estimateDischargeCount(dept.ID, month)

		var rate float64
		if dischargeCount > 0 {
			rate = float64(infectionCount) / float64(dischargeCount) * 100
		}

		departmentRates = append(departmentRates, MonthlyRate{
			DepartmentID:   dept.ID,
			DepartmentName: dept.Name,
			Month:          month,
			InfectionCount: int(infectionCount),
			DischargeCount: dischargeCount,
			InfectionRate:  rate,
			Threshold:      threshold,
			Exceeded:       rate > threshold,
		})

		totalInfections += int(infectionCount)
		totalDischarges += dischargeCount
	}

	siteDist := calculateSiteDistribution(monthStart, monthEnd)
	pathogenDist := calculatePathogenDistribution(monthStart, monthEnd)

	var hospitalRate float64
	if totalDischarges > 0 {
		hospitalRate = float64(totalInfections) / float64(totalDischarges) * 100
	}

	return StatisticsResponse{
		HospitalRate:       hospitalRate,
		DepartmentRates:    departmentRates,
		SiteDistribution:   siteDist,
		PathogenDistribution: pathogenDist,
		Month:              month,
	}
}

func GetReports(c *gin.Context) {
	var reports []models.Report
	query := database.DB

	reportType := c.Query("type")
	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}

	status := c.Query("status")
	if status != "" {
		query = query.Where("approval_status = ?", status)
	}

	query.Order("created_at desc").Find(&reports)
	c.JSON(http.StatusOK, reports)
}

func GetReport(c *gin.Context) {
	id := c.Param("id")
	var report models.Report
	if err := database.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "报告不存在"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func ApproveReport(c *gin.Context) {
	id := c.Param("id")
	action := c.Param("action")

	var report models.Report
	if err := database.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "报告不存在"})
		return
	}

	var req ReportApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if report.ApprovalStatus == models.ApprovalStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "报告已终审通过，不可修改"})
		return
	}

	var newStatus models.ApprovalStatus
	isSmallAmount := true

	switch action {
	case "first-review":
		if report.ApprovalStatus != models.ApprovalStatusSubmitted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许一审"})
			return
		}
		if isSmallAmount {
			newStatus = models.ApprovalStatusFinalReview
		} else {
			newStatus = models.ApprovalStatusFirstReview
		}

	case "second-review":
		if report.ApprovalStatus != models.ApprovalStatusFirstReview {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许二审"})
			return
		}
		newStatus = models.ApprovalStatusSecondReview

	case "final-review":
		if report.ApprovalStatus != models.ApprovalStatusSecondReview &&
			report.ApprovalStatus != models.ApprovalStatusFinalReview &&
			report.ApprovalStatus != models.ApprovalStatusSubmitted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许终审"})
			return
		}
		newStatus = models.ApprovalStatusApproved
		now := time.Now()
		report.ApprovedAt = &now
		report.ApprovedBy = req.Operator

	case "reject":
		newStatus = models.ApprovalStatusRejected

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审批操作"})
		return
	}

	report.ApprovalStatus = newStatus
	report.UpdatedAt = time.Now()

	if err := database.DB.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新报告失败"})
		return
	}

	history := models.ApprovalHistory{
		ReportID:  report.ID,
		Status:    newStatus,
		Operator:  req.Operator,
		Comment:   req.Comment,
		CreatedAt: time.Now(),
	}
	database.DB.Create(&history)

	c.JSON(http.StatusOK, report)
}

func GetReportApprovalHistory(c *gin.Context) {
	reportID := c.Param("id")
	var history []models.ApprovalHistory
	database.DB.Where("report_id = ?", reportID).Order("created_at desc").Find(&history)
	c.JSON(http.StatusOK, history)
}

func DeleteReport(c *gin.Context) {
	id := c.Param("id")
	var report models.Report
	if err := database.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "报告不存在"})
		return
	}

	if report.ApprovalStatus == models.ApprovalStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已终审的报告不可删除"})
		return
	}

	database.DB.Where("report_id = ?", id).Delete(&models.ApprovalHistory{})

	if err := database.DB.Delete(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除报告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetDepartments(c *gin.Context) {
	var departments []models.Department
	database.DB.Find(&departments)
	c.JSON(http.StatusOK, departments)
}
