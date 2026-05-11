package core

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"time"
	"recycling/api"
)

func GenerateCSV(records []api.Record, categoryID string) []byte {
	filtered := records
	if categoryID != "" {
		filtered = make([]api.Record, 0)
		for _, r := range records {
			for _, item := range r.Items {
				if item.CategoryID == categoryID {
					filtered = append(filtered, r)
					break
				}
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString("\uFEFF")

	writer := csv.NewWriter(&buf)

	headers := []string{
		"记录ID",
		"日期时间",
		"客户ID",
		"客户名称",
		"客户类型",
		"结算方式",
		"分类ID",
		"分类名称",
		"重量(kg)",
		"单价(元/kg)",
		"金额(元)",
		"总金额(元)",
	}
	writer.Write(headers)

	for _, r := range filtered {
		for _, item := range r.Items {
			if categoryID != "" && item.CategoryID != categoryID {
				continue
			}

			row := []string{
				r.ID,
				r.DateTime.Format("2006-01-02 15:04:05"),
				r.CustomerID,
				r.CustomerName,
				r.CustomerType,
				r.Settlement,
				item.CategoryID,
				item.CategoryName,
				fmt.Sprintf("%.2f", item.Weight),
				fmt.Sprintf("%.2f", item.Price),
				fmt.Sprintf("%.2f", item.Amount),
				fmt.Sprintf("%.2f", r.TotalAmount),
			}
			writer.Write(row)
		}
	}

	writer.Flush()
	return buf.Bytes()
}

func CalculateSummary(records []api.Record, prevMonthRecords []api.Record) api.SummaryResponse {
	categoryStats := make(map[string]*categoryStat)

	for _, r := range records {
		for _, item := range r.Items {
			stat, exists := categoryStats[item.CategoryID]
			if !exists {
				stat = &categoryStat{
					categoryID:   item.CategoryID,
					categoryName: item.CategoryName,
				}
				categoryStats[item.CategoryID] = stat
			}
			stat.totalWeight += item.Weight
			stat.totalAmount += item.Amount
		}
	}

	prevStats := make(map[string]*categoryStat)
	for _, r := range prevMonthRecords {
		for _, item := range r.Items {
			stat, exists := prevStats[item.CategoryID]
			if !exists {
				stat = &categoryStat{
					categoryID: item.CategoryID,
				}
				prevStats[item.CategoryID] = stat
			}
			stat.totalWeight += item.Weight
			stat.totalAmount += item.Amount
		}
	}

	items := make([]api.SummaryItem, 0, len(categoryStats))
	for _, stat := range categoryStats {
		prevStat, hasPrev := prevStats[stat.categoryID]
		var weightChange, amountChange float64
		if hasPrev && prevStat.totalWeight > 0 {
			weightChange = ((stat.totalWeight - prevStat.totalWeight) / prevStat.totalWeight) * 100
		}
		if hasPrev && prevStat.totalAmount > 0 {
			amountChange = ((stat.totalAmount - prevStat.totalAmount) / prevStat.totalAmount) * 100
		}

		items = append(items, api.SummaryItem{
			CategoryID:    stat.categoryID,
			CategoryName:  stat.categoryName,
			TotalWeight:   roundToCents(stat.totalWeight),
			TotalAmount:   roundToCents(stat.totalAmount),
			WeightChange:  roundToCents(weightChange),
			AmountChange:  roundToCents(amountChange),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CategoryID < items[j].CategoryID
	})

	return api.SummaryResponse{Items: items}
}

func GenerateSummaryCSV(summary api.SummaryResponse, year, month int) []byte {
	var buf bytes.Buffer
	buf.WriteString("\uFEFF")

	writer := csv.NewWriter(&buf)

	headers := []string{
		"分类ID",
		"分类名称",
		"总重量(kg)",
		"总金额(元)",
		"重量环比(%)",
		"金额环比(%)",
	}
	writer.Write(headers)

	for _, item := range summary.Items {
		row := []string{
			item.CategoryID,
			item.CategoryName,
			fmt.Sprintf("%.2f", item.TotalWeight),
			fmt.Sprintf("%.2f", item.TotalAmount),
			fmt.Sprintf("%.2f", item.WeightChange),
			fmt.Sprintf("%.2f", item.AmountChange),
		}
		writer.Write(row)
	}

	writer.Flush()
	return buf.Bytes()
}

type categoryStat struct {
	categoryID   string
	categoryName string
	totalWeight  float64
	totalAmount  float64
}

func GetPrevMonth(year, month int) (int, int) {
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	prev := t.AddDate(0, -1, 0)
	return prev.Year(), int(prev.Month())
}
