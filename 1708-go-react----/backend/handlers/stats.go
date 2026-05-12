package handlers

import (
	"blood-management-system/config"
	"blood-management-system/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatsHandler struct {
	db *gorm.DB
}

func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

type MonthlyCollectionItem struct {
	BloodType string  `json:"blood_type"`
	Count     int64   `json:"count"`
	Volume    float64 `json:"volume_ml"`
}

func (h *StatsHandler) MonthlyCollection(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", fmt.Sprintf("%d", time.Now().Month())))

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	result := make([]MonthlyCollectionItem, 0)

	for _, bt := range config.BloodTypeCombinations {
		var count int64
		var totalVolume float64

		h.db.Model(&models.BloodCollection{}).
			Joins("LEFT JOIN donors ON blood_collections.donor_id = donors.id").
			Where("donors.blood_type = ? AND blood_collections.collection_time >= ? AND blood_collections.collection_time < ?", bt, startDate, endDate).
			Count(&count)

		h.db.Model(&models.BloodCollection{}).
			Joins("LEFT JOIN donors ON blood_collections.donor_id = donors.id").
			Where("donors.blood_type = ? AND blood_collections.collection_time >= ? AND blood_collections.collection_time < ?", bt, startDate, endDate).
			Select("COALESCE(SUM(volume_ml), 0)").Scan(&totalVolume)

		if count > 0 {
			result = append(result, MonthlyCollectionItem{
				BloodType: bt,
				Count:     count,
				Volume:    totalVolume,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"year":  year,
		"month": month,
		"data":  result,
	})
}

type MonthlyScrapItem struct {
	BloodType      string  `json:"blood_type"`
	TotalTested    int64   `json:"total_tested"`
	Scrapped       int64   `json:"scrapped"`
	ScrapRate      string  `json:"scrap_rate"`
}

func (h *StatsHandler) MonthlyScrap(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", fmt.Sprintf("%d", time.Now().Month())))

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	result := make([]MonthlyScrapItem, 0)

	for _, bt := range config.BloodTypeCombinations {
		var totalTested int64
		var scrapped int64

		h.db.Model(&models.BloodCollection{}).
			Joins("LEFT JOIN donors ON blood_collections.donor_id = donors.id").
			Where("donors.blood_type = ? AND blood_collections.test_progress >= 1 AND blood_collections.created_at >= ? AND blood_collections.created_at < ?", bt, startDate, endDate).
			Count(&totalTested)

		h.db.Model(&models.BloodCollection{}).
			Joins("LEFT JOIN donors ON blood_collections.donor_id = donors.id").
			Where("donors.blood_type = ? AND blood_collections.status IN ? AND blood_collections.created_at >= ? AND blood_collections.created_at < ?", bt, []string{config.StatusScrapped, config.StatusDisqualified}, startDate, endDate).
			Count(&scrapped)

		scrapRate := "0.00%"
		if totalTested > 0 {
			rate := float64(scrapped) / float64(totalTested) * 100
			scrapRate = fmt.Sprintf("%.2f%%", rate)
		}

		result = append(result, MonthlyScrapItem{
			BloodType:   bt,
			TotalTested: totalTested,
			Scrapped:    scrapped,
			ScrapRate:   scrapRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":  year,
		"month": month,
		"data":  result,
	})
}

type HospitalRankingItem struct {
	HospitalName string  `json:"hospital_name"`
	TotalVolume  float64 `json:"total_volume_ml"`
	TotalBags    int64   `json:"total_bags"`
}

func (h *StatsHandler) HospitalRanking(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	startDate := time.Now().AddDate(0, 0, -days)

	type RawResult struct {
		HospitalName string
		TotalVolume  float64
		TotalBags    int64
	}

	var rawResults []RawResult
	h.db.Model(&models.BloodRequest{}).
		Select("hospital_name, COUNT(*) as total_bags, SUM(quantity * 200) as total_volume").
		Where("status = ? AND created_at >= ?", config.ReqStatusIssued, startDate).
		Group("hospital_name").
		Order("total_bags DESC").
		Scan(&rawResults)

	result := make([]HospitalRankingItem, 0)
	for i, r := range rawResults {
		result = append(result, HospitalRankingItem{
			HospitalName: fmt.Sprintf("%d. %s", i+1, r.HospitalName),
			TotalVolume:  r.TotalVolume,
			TotalBags:    r.TotalBags,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"days": days,
		"data": result,
	})
}

type InventoryTurnoverItem struct {
	BloodType   string  `json:"blood_type"`
	ProductType string  `json:"product_type"`
	AvgStock    float64 `json:"avg_stock"`
	DailyIssued float64 `json:"daily_issued"`
	TurnoverDays float64 `json:"turnover_days"`
}

func (h *StatsHandler) InventoryTurnover(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	startDate := time.Now().AddDate(0, 0, -days)

	result := make([]InventoryTurnoverItem, 0)

	for _, bt := range config.BloodTypeCombinations {
		for _, pt := range config.ProductTypes {
			var totalIssued int64
			var totalVolume float64

			h.db.Model(&models.Inventory{}).
				Where("blood_type = ? AND product_type = ? AND status = ? AND issued_date >= ?", bt, pt, config.StatusIssued, startDate).
				Count(&totalIssued)

			h.db.Model(&models.Inventory{}).
				Where("blood_type = ? AND product_type = ? AND status = ? AND issued_date >= ?", bt, pt, config.StatusIssued, startDate).
				Select("COALESCE(SUM(volume_ml), 0)").Scan(&totalVolume)

			var currentStock int64
			h.db.Model(&models.Inventory{}).
				Where("blood_type = ? AND product_type = ? AND status IN ?", bt, pt, []string{config.StatusInStock, config.StatusFrozen}).
				Count(&currentStock)

			dailyIssued := float64(totalIssued) / float64(days)
			avgStock := float64(currentStock)

			turnoverDays := 0.0
			if dailyIssued > 0 {
				turnoverDays = avgStock / dailyIssued
			}

			if currentStock > 0 || totalIssued > 0 {
				result = append(result, InventoryTurnoverItem{
					BloodType:    bt,
					ProductType:  pt,
					AvgStock:     avgStock,
					DailyIssued:  dailyIssued,
					TurnoverDays: turnoverDays,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"days": days,
		"data": result,
	})
}

func (h *StatsHandler) Dashboard(c *gin.Context) {
	var totalDonors int64
	h.db.Model(&models.Donor{}).Count(&totalDonors)

	var pendingTests int64
	h.db.Model(&models.BloodCollection{}).Where("status IN ?", []string{config.StatusPendingTest, config.StatusTesting}).Count(&pendingTests)

	var totalInStock int64
	h.db.Model(&models.Inventory{}).Where("status IN ?", []string{config.StatusInStock, config.StatusFrozen}).Count(&totalInStock)

	var pendingRequests int64
	h.db.Model(&models.BloodRequest{}).Where("status IN ?", []string{config.ReqStatusPending, config.ReqStatusQueued}).Count(&pendingRequests)

	lowStockAlerts := make([]map[string]interface{}, 0)
	for _, bt := range config.BloodTypeCombinations {
		for _, pt := range config.ProductTypes {
			var count int64
			h.db.Model(&models.Inventory{}).
				Where("blood_type = ? AND product_type = ? AND status IN ?", bt, pt, []string{config.StatusInStock, config.StatusFrozen}).
				Count(&count)

			var safetyStock models.SafetyStock
			h.db.Where("blood_type = ? AND product_type = ?", bt, pt).First(&safetyStock)

			if count < int64(safetyStock.MinQuantity) {
				lowStockAlerts = append(lowStockAlerts, map[string]interface{}{
					"blood_type":   bt,
					"product_type": pt,
					"current":      count,
					"min_required": safetyStock.MinQuantity,
					"deficit":      safetyStock.MinQuantity - int(count),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_donors":      totalDonors,
		"pending_tests":     pendingTests,
		"total_in_stock":    totalInStock,
		"pending_requests":  pendingRequests,
		"low_stock_alerts":  lowStockAlerts,
	})
}
