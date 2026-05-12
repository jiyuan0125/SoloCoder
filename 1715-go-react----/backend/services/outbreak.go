package services

import (
	"encoding/json"
	"fmt"
	"time"

	"epidemic-management/database"
	"epidemic-management/models"
)

func CreateOutbreakReport(report *models.OutbreakReport) error {
	var disease models.Disease
	if err := database.DB.Where("name = ?", report.DiseaseName).First(&disease).Error; err != nil {
		return fmt.Errorf("未找到该传染病信息")
	}

	report.IsDelayed = checkDelayed(report.OnsetTime, report.ReportTime, disease.ReportDeadline)
	report.Status = "pending"

	if err := database.DB.Create(report).Error; err != nil {
		return err
	}

	if err := checkClustering(report); err != nil {
		fmt.Println("检查聚集性预警失败:", err)
	}

	CreateTodo(&models.Todo{
		Title:       "启动流调",
		Description: fmt.Sprintf("新疫情报告: %s - %s", report.PatientName, report.DiseaseName),
		Priority:    models.PriorityHigh,
		Status:      models.TodoPending,
		RelatedType: "outbreak",
		RelatedID:   report.ID,
	})

	if disease.Class == models.ClassA {
		CreateTodo(&models.Todo{
			Title:       "启动应急响应",
			Description: fmt.Sprintf("甲类传染病报告: %s - %s", report.PatientName, report.DiseaseName),
			Priority:    models.PriorityHighest,
			Status:      models.TodoPending,
			RelatedType: "emergency",
			RelatedID:   report.ID,
		})
	}

	return nil
}

func checkDelayed(onset, report time.Time, deadlineHours int) bool {
	diff := report.Sub(onset)
	return diff.Hours() > float64(deadlineHours)
}

func checkClustering(report *models.OutbreakReport) error {
	startTime := report.ReportTime.Add(-24 * time.Hour)
	var count int64
	database.DB.Model(&models.OutbreakReport{}).
		Where("disease_name = ? AND district = ? AND report_time >= ? AND id != ?",
			report.DiseaseName, report.District, startTime, report.ID).
		Count(&count)

	if count+1 >= 3 {
		report.HasClustering = true
		database.DB.Save(report)

		var existingAlert models.Alert
		err := database.DB.Where("type = ? AND disease_name = ? AND district = ? AND is_active = ?",
			"clustering", report.DiseaseName, report.District, true).First(&existingAlert).Error

		if err != nil {
			alert := &models.Alert{
				Type:        "clustering",
				Message:     fmt.Sprintf("聚集性疫情预警: %s在%s出现%d例", report.DiseaseName, report.District, count+1),
				DiseaseName: report.DiseaseName,
				District:    report.District,
				Count:       int(count + 1),
				IsActive:    true,
			}
			database.DB.Create(alert)
		} else {
			existingAlert.Count = int(count + 1)
			existingAlert.Message = fmt.Sprintf("聚集性疫情预警: %s在%s出现%d例", report.DiseaseName, report.District, count+1)
			database.DB.Save(&existingAlert)
		}
	}
	return nil
}

func ListOutbreakReports() ([]models.OutbreakReport, error) {
	var reports []models.OutbreakReport
	err := database.DB.Order("created_at desc").Find(&reports).Error
	return reports, err
}

func GetOutbreakReport(id uint) (*models.OutbreakReport, error) {
	var report models.OutbreakReport
	err := database.DB.First(&report, id).Error
	return &report, err
}

func GetActiveAlerts() ([]models.Alert, error) {
	var alerts []models.Alert
	err := database.DB.Where("is_active = ?", true).Find(&alerts).Error
	return alerts, err
}

func GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCases int64
	database.DB.Model(&models.OutbreakReport{}).Count(&totalCases)
	stats["totalCases"] = totalCases

	var delayedCount int64
	database.DB.Model(&models.OutbreakReport{}).Where("is_delayed = ?", true).Count(&delayedCount)
	stats["delayedCount"] = delayedCount

	var activeContacts int64
	database.DB.Model(&models.Contact{}).Where("status IN ?", []string{"正常", "出现症状"}).Count(&activeContacts)
	stats["activeContacts"] = activeContacts

	var confirmedContacts int64
	database.DB.Model(&models.Contact{}).Where("status = ?", "确诊").Count(&confirmedContacts)
	stats["confirmedContacts"] = confirmedContacts

	var excludedContacts int64
	database.DB.Model(&models.Contact{}).Where("status = ?", "排除").Count(&excludedContacts)
	stats["excludedContacts"] = excludedContacts

	var pendingTodos int64
	database.DB.Model(&models.Todo{}).Where("status = ?", "待办").Count(&pendingTodos)
	stats["pendingTodos"] = pendingTodos

	var totalVaccinations int64
	database.DB.Model(&models.VaccinationRecord{}).Count(&totalVaccinations)
	stats["totalVaccinations"] = totalVaccinations

	diseaseDist := make(map[string]int64)
	var results []struct {
		DiseaseName string
		Count       int64
	}
	database.DB.Model(&models.OutbreakReport{}).Select("disease_name, count(*) as count").
		Group("disease_name").Scan(&results)
	for _, r := range results {
		diseaseDist[r.DiseaseName] = r.Count
	}
	distJSON, _ := json.Marshal(diseaseDist)
	stats["diseaseDistribution"] = string(distJSON)

	regionDist := make(map[string]int64)
	var regionResults []struct {
		District string
		Count    int64
	}
	database.DB.Model(&models.OutbreakReport{}).Select("district, count(*) as count").
		Group("district").Scan(&regionResults)
	for _, r := range regionResults {
		regionDist[r.District] = r.Count
	}
	regionJSON, _ := json.Marshal(regionDist)
	stats["regionDistribution"] = string(regionJSON)

	var last30Days []struct {
		Date  string
		Count int64
	}
	database.DB.Model(&models.OutbreakReport{}).
		Select("date(created_at) as date, count(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -30)).
		Group("date(created_at)").Order("date asc").Scan(&last30Days)
	trendJSON, _ := json.Marshal(last30Days)
	stats["trend"] = string(trendJSON)

	return stats, nil
}
