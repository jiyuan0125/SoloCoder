package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"epidemic-management/database"
	"epidemic-management/models"
)

func StartScheduler() {
	c := cron.New()

	c.AddFunc("0 0 * * *", func() {
		CheckAndExpireContacts()
		fmt.Println("[Scheduler] 已执行每日密接者观察到期检查")
	})

	c.AddFunc("0 0 * * 1", func() {
		GenerateWeeklyReport()
		fmt.Println("[Scheduler] 已生成疫情周报")
	})

	c.Start()
}

func GenerateWeeklyReport() (*models.WeeklyReport, error) {
	now := time.Now()
	weekEnd := now
	weekStart := weekEnd.AddDate(0, 0, -7)

	var newCases int64
	database.DB.Model(&models.OutbreakReport{}).
		Where("created_at >= ? AND created_at < ?", weekStart, weekEnd).Count(&newCases)

	diseaseDist := make(map[string]int64)
	var diseaseResults []struct {
		DiseaseName string
		Count       int64
	}
	database.DB.Model(&models.OutbreakReport{}).
		Select("disease_name, count(*) as count").
		Where("created_at >= ? AND created_at < ?", weekStart, weekEnd).
		Group("disease_name").Scan(&diseaseResults)
	for _, r := range diseaseResults {
		diseaseDist[r.DiseaseName] = r.Count
	}

	regionDist := make(map[string]int64)
	var regionResults []struct {
		District string
		Count    int64
	}
	database.DB.Model(&models.OutbreakReport{}).
		Select("district, count(*) as count").
		Where("created_at >= ? AND created_at < ?", weekStart, weekEnd).
		Group("district").Scan(&regionResults)
	for _, r := range regionResults {
		regionDist[r.District] = r.Count
	}

	contactStats := make(map[string]interface{})
	var active, confirmed, excluded int64
	database.DB.Model(&models.Contact{}).Where("status IN ?", []string{"正常", "出现症状"}).Count(&active)
	database.DB.Model(&models.Contact{}).Where("status = ?", "确诊").Count(&confirmed)
	database.DB.Model(&models.Contact{}).Where("status = ?", "排除").Count(&excluded)
	contactStats["active"] = active
	contactStats["confirmed"] = confirmed
	contactStats["excluded"] = excluded

	distJSON, _ := json.Marshal(diseaseDist)
	regionJSON, _ := json.Marshal(regionDist)
	contactJSON, _ := json.Marshal(contactStats)

	report := &models.WeeklyReport{
		WeekStart:     weekStart,
		WeekEnd:       weekEnd,
		NewCases:      int(newCases),
		DiseaseDist:   string(distJSON),
		RegionDist:    string(regionJSON),
		ContactStats:  string(contactJSON),
		ReportContent: fmt.Sprintf("本周新增%d例病例", newCases),
	}

	if err := database.DB.Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func ListWeeklyReports() ([]models.WeeklyReport, error) {
	var reports []models.WeeklyReport
	err := database.DB.Order("created_at desc").Find(&reports).Error
	return reports, err
}

func ListDiseases() ([]models.Disease, error) {
	var diseases []models.Disease
	err := database.DB.Find(&diseases).Error
	return diseases, err
}
