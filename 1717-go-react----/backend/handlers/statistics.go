package handlers

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"hospital-infection/database"
	"hospital-infection/models"

	"github.com/gin-gonic/gin"
)

type MonthlyRate struct {
	DepartmentID   uint    `json:"DepartmentID"`
	DepartmentName string  `json:"DepartmentName"`
	Month          string  `json:"Month"`
	InfectionCount int     `json:"InfectionCount"`
	DischargeCount int     `json:"DischargeCount"`
	InfectionRate  float64 `json:"InfectionRate"`
	Threshold      float64 `json:"Threshold"`
	Exceeded       bool    `json:"Exceeded"`
}

type SiteDistribution struct {
	Site  models.InfectionSite `json:"Site"`
	Count int                   `json:"Count"`
	Ratio float64               `json:"Ratio"`
}

type PathogenDistribution struct {
	Pathogen string  `json:"Pathogen"`
	Count    int     `json:"Count"`
	Ratio    float64 `json:"Ratio"`
}

type StatisticsResponse struct {
	HospitalRate         float64               `json:"HospitalRate"`
	DepartmentRates      []MonthlyRate         `json:"DepartmentRates"`
	SiteDistribution     []SiteDistribution    `json:"SiteDistribution"`
	PathogenDistribution []PathogenDistribution `json:"PathogenDistribution"`
	Month                string                `json:"Month"`
}

func getDefaultThreshold() float64 {
	return 2.0
}

func CalculateMonthlyRates(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		month = c.Query("Month")
	}
	if month == "" {
		now := time.Now()
		month = fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	}

	threshold := getDefaultThreshold()

	var departments []models.Department
	database.DB.Find(&departments)

	monthStart, _ := time.Parse("2006-01", month)
	monthEnd := monthStart.AddDate(0, 1, 0)

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
		} else {
			rate = 0
		}

		rate = math.Round(rate*100) / 100

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

		if rate > threshold {
			generateAlert(dept.ID, month, rate, threshold)
		}
	}

	sort.Slice(departmentRates, func(i, j int) bool {
		return departmentRates[i].InfectionRate > departmentRates[j].InfectionRate
	})

	siteDist := calculateSiteDistribution(monthStart, monthEnd)
	pathogenDist := calculatePathogenDistribution(monthStart, monthEnd)

	var hospitalRate float64
	if totalDischarges > 0 {
		hospitalRate = float64(totalInfections) / float64(totalDischarges) * 100
	}
	hospitalRate = math.Round(hospitalRate*100) / 100

	c.JSON(http.StatusOK, StatisticsResponse{
		HospitalRate:         hospitalRate,
		DepartmentRates:      departmentRates,
		SiteDistribution:     siteDist,
		PathogenDistribution: pathogenDist,
		Month:                month,
	})
}

func Get12MonthTrend(c *gin.Context) {
	type TrendPoint struct {
		Month         string  `json:"Month"`
		InfectionRate float64 `json:"InfectionRate"`
	}

	var trend []TrendPoint
	now := time.Now()

	for i := 11; i >= 0; i-- {
		monthDate := now.AddDate(0, -i, 0)
		month := fmt.Sprintf("%d-%02d", monthDate.Year(), monthDate.Month())

		monthStart := time.Date(monthDate.Year(), monthDate.Month(), 1, 0, 0, 0, 0, time.Local)
		monthEnd := monthStart.AddDate(0, 1, 0)

		var infectionCount int64
		database.DB.Model(&models.InfectionCase{}).
			Where("infection_date >= ? AND infection_date < ? AND infection_type = ?",
				monthStart, monthEnd, models.InfectionTypeNosocomial).
			Count(&infectionCount)

		dischargeCount := estimateHospitalDischargeCount(month)

		var rate float64
		if dischargeCount > 0 {
			rate = float64(infectionCount) / float64(dischargeCount) * 100
		}
		rate = math.Round(rate*100) / 100

		trend = append(trend, TrendPoint{
			Month:         month,
			InfectionRate: rate,
		})
	}

	c.JSON(http.StatusOK, trend)
}

func estimateDischargeCount(departmentID uint, month string) int {
	var count int64
	database.DB.Model(&models.InfectionCase{}).
		Where("department_id = ?", departmentID).
		Count(&count)
	base := int(count) * 10
	if base < 50 {
		base = 50
	}
	return base
}

func estimateHospitalDischargeCount(month string) int {
	var departments []models.Department
	database.DB.Find(&departments)

	total := 0
	for _, dept := range departments {
		total += estimateDischargeCount(dept.ID, month)
	}
	return total
}

func calculateSiteDistribution(start, end time.Time) []SiteDistribution {
	sites := []models.InfectionSite{
		models.SiteRespiratory,
		models.SiteSurgicalIncision,
		models.SiteUrinaryTract,
		models.SiteBloodstream,
		models.SiteDigestive,
		models.SiteSkinSoftTissue,
	}

	var total int64
	database.DB.Model(&models.InfectionCase{}).
		Where("infection_date >= ? AND infection_date < ?", start, end).
		Count(&total)

	var results []SiteDistribution
	for _, site := range sites {
		var count int64
		database.DB.Model(&models.InfectionCase{}).
			Where("infection_site = ? AND infection_date >= ? AND infection_date < ?", site, start, end).
			Count(&count)

		var ratio float64
		if total > 0 {
			ratio = float64(count) / float64(total) * 100
		}
		ratio = math.Round(ratio*100) / 100

		results = append(results, SiteDistribution{
			Site:  site,
			Count: int(count),
			Ratio: ratio,
		})
	}

	return results
}

func calculatePathogenDistribution(start, end time.Time) []PathogenDistribution {
	var cases []models.InfectionCase
	database.DB.Where("infection_date >= ? AND infection_date < ?", start, end).Find(&cases)

	pathogenMap := make(map[string]int)
	total := 0
	for _, c := range cases {
		if c.Pathogen != "" {
			pathogenMap[c.Pathogen]++
			total++
		}
	}

	var results []PathogenDistribution
	for pathogen, count := range pathogenMap {
		var ratio float64
		if total > 0 {
			ratio = float64(count) / float64(total) * 100
		}
		ratio = math.Round(ratio*100) / 100

		results = append(results, PathogenDistribution{
			Pathogen: pathogen,
			Count:    count,
			Ratio:    ratio,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	return results
}

func generateAlert(departmentID uint, month string, rate, threshold float64) {
	var existingAlert models.Alert
	database.DB.Where("department_id = ? AND month = ?", departmentID, month).First(&existingAlert)

	if existingAlert.ID > 0 {
		return
	}

	alert := models.Alert{
		DepartmentID:  departmentID,
		AlertType:     models.AlertRateExceeded,
		Month:         month,
		InfectionRate: rate,
		Threshold:     threshold,
		Message:       fmt.Sprintf("科室月度感染率 %.2f%% 超过预警阈值 %.2f%%", rate, threshold),
		Status:        "active",
		Notified:      true,
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&alert)

	checkConsecutiveMonths(departmentID, month)
}

func checkConsecutiveMonths(departmentID uint, currentMonth string) {
	currentDate, _ := time.Parse("2006-01", currentMonth)
	prevMonth := currentDate.AddDate(0, -1, 0)
	prevMonthStr := fmt.Sprintf("%d-%02d", prevMonth.Year(), prevMonth.Month())

	var prevAlert models.Alert
	database.DB.Where("department_id = ? AND month = ?", departmentID, prevMonthStr).First(&prevAlert)

	if prevAlert.ID > 0 {
		var existingIntervention models.Alert
		database.DB.Where("department_id = ? AND month = ? AND alert_type = ?",
			departmentID, currentMonth, models.AlertNeedIntervention).First(&existingIntervention)

		if existingIntervention.ID == 0 {
			intervention := models.Alert{
				DepartmentID: departmentID,
				AlertType:    models.AlertNeedIntervention,
				Month:        currentMonth,
				Message:      "连续两个月感染率超标，需要进行干预",
				Status:       "active",
				Notified:     true,
				CreatedAt:    time.Now(),
			}
			database.DB.Create(&intervention)
		}
	}
}

func GetAlerts(c *gin.Context) {
	var alerts []models.Alert
	query := database.DB.Preload("Department")

	status := c.Query("status")
	if status == "" {
		status = c.Query("Status")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	alertType := c.Query("type")
	if alertType == "" {
		alertType = c.Query("AlertType")
	}
	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}

	query.Order("created_at desc").Find(&alerts)
	c.JSON(http.StatusOK, alerts)
}
