package report

import (
	"fmt"
	"time"

	"warehouse-mgmt/internal/models"
)

type CountItem struct {
	ProductID    string
	LocationCode string
	Quantity     int
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateCountReport(db *models.Database, actualItems []CountItem) *models.StockCountReport {
	report := &models.StockCountReport{
		ID:        models.GenerateID("RPT"),
		CreatedAt: time.Now(),
		Status:    "generated",
	}

	for _, actual := range actualItems {
		systemQty := 0
		sr := db.GetStockRecord(actual.ProductID, actual.LocationCode)
		if sr != nil {
			systemQty = sr.Quantity
		}

		diff := actual.Quantity - systemQty
		diffPct := 0.0
		if systemQty > 0 {
			diffPct = float64(abs(diff)) / float64(systemQty) * 100
		}

		item := models.StockCountItem{
			ProductID:       actual.ProductID,
			LocationCode:    actual.LocationCode,
			SystemQuantity:  systemQty,
			ActualQuantity:  actual.Quantity,
			Difference:      diff,
			DifferencePct:   diffPct,
			Highlight:       diffPct > 5.0,
		}
		report.Items = append(report.Items, item)
	}

	for _, sr := range db.StockRecords {
		found := false
		for _, actual := range actualItems {
			if actual.ProductID == sr.ProductID && actual.LocationCode == sr.LocationCode {
				found = true
				break
			}
		}
		if !found {
			diff := -sr.Quantity
			diffPct := 100.0
			item := models.StockCountItem{
				ProductID:       sr.ProductID,
				LocationCode:    sr.LocationCode,
				SystemQuantity:  sr.Quantity,
				ActualQuantity:  0,
				Difference:      diff,
				DifferencePct:   diffPct,
				Highlight:       true,
			}
			report.Items = append(report.Items, item)
		}
	}

	db.Reports = append(db.Reports, *report)
	return report
}

func (s *Service) PrintReport(report *models.StockCountReport) {
	fmt.Printf("Stock Count Report %s\n", report.ID)
	fmt.Println("========================================================")
	fmt.Printf("%-10s %-10s %-10s %-10s %-10s %-10s\n",
		"Product", "Location", "System", "Actual", "Diff", "Diff%")
	fmt.Println("--------------------------------------------------------")
	for _, item := range report.Items {
		prefix := ""
		suffix := ""
		if item.Highlight {
			prefix = "**"
			suffix = "**"
		}
		fmt.Printf("%s%-10s %-10s %-10d %-10d %-10d %-10.2f%%%s\n",
			prefix, item.ProductID, item.LocationCode,
			item.SystemQuantity, item.ActualQuantity,
			item.Difference, item.DifferencePct, suffix)
	}
	fmt.Println("========================================================")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
