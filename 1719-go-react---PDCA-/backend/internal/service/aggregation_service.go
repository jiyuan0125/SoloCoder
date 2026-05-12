package service

import (
	"fmt"
	"sort"
	"time"

	"medical-quality-system/internal/model"
	"medical-quality-system/pkg/db"
)

type AggregationService struct{}

func NewAggregationService() *AggregationService {
	return &AggregationService{}
}

type DepartmentAggregation struct {
	Department   string                   `json:"department"`
	Indicators   []IndicatorDepartmentStat `json:"indicators"`
	TotalCount   int                      `json:"total_count"`
	MetCount     int                      `json:"met_count"`
	ComplianceRate float64               `json:"compliance_rate"`
}

type IndicatorDepartmentStat struct {
	IndicatorID   uint    `json:"indicator_id"`
	IndicatorCode string  `json:"indicator_code"`
	IndicatorName string  `json:"indicator_name"`
	Value         float64 `json:"value"`
	TargetValue   float64 `json:"target_value"`
	IsTargetMet   bool    `json:"is_target_met"`
	Month         string  `json:"month"`
}

func (s *AggregationService) ByDepartment(month string) ([]DepartmentAggregation, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	var indicators []model.Indicator
	if err := db.DB.Find(&indicators).Error; err != nil {
		return nil, err
	}

	deptMap := make(map[string][]IndicatorDepartmentStat)
	for _, indicator := range indicators {
		var data model.IndicatorData
		err := db.DB.Where("indicator_id = ? AND month = ?", indicator.ID, month).First(&data).Error
		if err != nil {
			continue
		}

		stat := IndicatorDepartmentStat{
			IndicatorID:   indicator.ID,
			IndicatorCode: indicator.Code,
			IndicatorName: indicator.Name,
			Value:         data.Value,
			TargetValue:   indicator.TargetValue,
			IsTargetMet:   data.IsTargetMet,
			Month:         month,
		}

		dept := indicator.SourceDept
		if dept == "" {
			dept = "未指定科室"
		}
		deptMap[dept] = append(deptMap[dept], stat)
	}

	result := make([]DepartmentAggregation, 0, len(deptMap))
	for dept, stats := range deptMap {
		metCount := 0
		for _, stat := range stats {
			if stat.IsTargetMet {
				metCount++
			}
		}

		complianceRate := 0.0
		if len(stats) > 0 {
			complianceRate = float64(metCount) / float64(len(stats)) * 100
		}

		result = append(result, DepartmentAggregation{
			Department:     dept,
			Indicators:     stats,
			TotalCount:     len(stats),
			MetCount:       metCount,
			ComplianceRate: complianceRate,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ComplianceRate < result[j].ComplianceRate
	})

	return result, nil
}

type CategoryAggregation struct {
	Category       model.IndicatorCategory `json:"category"`
	CategoryName   string                  `json:"category_name"`
	TotalIndicators int                     `json:"total_indicators"`
	DataCount      int                     `json:"data_count"`
	MetCount       int                     `json:"met_count"`
	ComplianceRate float64                 `json:"compliance_rate"`
	AverageValue   float64                 `json:"average_value"`
}

var categoryNames = map[model.IndicatorCategory]string{
	model.CategorySafety:     "医疗安全指标",
	model.CategoryEfficiency: "医疗效率指标",
	model.CategoryQuality:    "医疗质量指标",
	model.CategoryExperience: "患者体验指标",
}

func (s *AggregationService) ByCategory(fromMonth, toMonth string) ([]CategoryAggregation, error) {
	if fromMonth == "" {
		fromMonth = time.Now().AddDate(0, -11, 0).Format("2006-01")
	}
	if toMonth == "" {
		toMonth = time.Now().Format("2006-01")
	}

	var indicators []model.Indicator
	if err := db.DB.Find(&indicators).Error; err != nil {
		return nil, err
	}

	categoryMap := make(map[model.IndicatorCategory]CategoryAggregation)
	for _, indicator := range indicators {
		cat := indicator.Category
		if cat == "" {
			cat = model.CategoryQuality
		}

		agg, exists := categoryMap[cat]
		if !exists {
			agg = CategoryAggregation{
				Category:     cat,
				CategoryName: categoryNames[cat],
			}
		}
		agg.TotalIndicators++

		var dataList []model.IndicatorData
		db.DB.Where("indicator_id = ? AND month >= ? AND month <= ?", indicator.ID, fromMonth, toMonth).Find(&dataList)

		for _, data := range dataList {
			agg.DataCount++
			if data.IsTargetMet {
				agg.MetCount++
			}
			agg.AverageValue += data.Value
		}

		categoryMap[cat] = agg
	}

	result := make([]CategoryAggregation, 0, len(categoryMap))
	for _, agg := range categoryMap {
		if agg.DataCount > 0 {
			agg.ComplianceRate = float64(agg.MetCount) / float64(agg.DataCount) * 100
			agg.AverageValue = agg.AverageValue / float64(agg.DataCount)
		}
		result = append(result, agg)
	}

	return result, nil
}

type TimeAggregation struct {
	Period        string                  `json:"period"`
	PeriodType    string                  `json:"period_type"`
	TotalCount    int                     `json:"total_count"`
	MetCount      int                     `json:"met_count"`
	ComplianceRate float64                `json:"compliance_rate"`
}

func (s *AggregationService) ByTime(periodType string, fromTime, toTime string) ([]TimeAggregation, error) {
	now := time.Now()
	if periodType == "" {
		periodType = "month"
	}

	var periods []string
	var periodFormat string
	var step int

	switch periodType {
	case "quarter":
		periodFormat = "2006-Q1"
		step = 3
		if fromTime == "" {
			fromTime = now.AddDate(0, -11, 0).Format("2006-01")
		}
		if toTime == "" {
			toTime = now.Format("2006-01")
		}
		periods = generateQuarterPeriods(fromTime, toTime)
	case "year":
		periodFormat = "2006"
		step = 12
		if fromTime == "" {
			fromTime = now.AddDate(-2, 0, 0).Format("2006")
		}
		if toTime == "" {
			toTime = now.Format("2006")
		}
		periods = generateYearPeriods(fromTime, toTime)
	default:
		periodFormat = "2006-01"
		step = 1
		if fromTime == "" {
			fromTime = now.AddDate(0, -11, 0).Format("2006-01")
		}
		if toTime == "" {
			toTime = now.Format("2006-01")
		}
		periods = generateMonthPeriods(fromTime, toTime)
		_ = periodFormat
		_ = step
	}

	result := make([]TimeAggregation, 0, len(periods))
	for _, period := range periods {
		var dataList []model.IndicatorData
		var err error

		switch periodType {
		case "quarter":
			dataList, err = s.getDataByQuarter(period)
		case "year":
			dataList, err = s.getDataByYear(period)
		default:
			dataList, err = s.getDataByMonth(period)
		}

		if err != nil {
			return nil, err
		}

		metCount := 0
		for _, data := range dataList {
			if data.IsTargetMet {
				metCount++
			}
		}

		complianceRate := 0.0
		if len(dataList) > 0 {
			complianceRate = float64(metCount) / float64(len(dataList)) * 100
		}

		result = append(result, TimeAggregation{
			Period:         period,
			PeriodType:     periodType,
			TotalCount:     len(dataList),
			MetCount:       metCount,
			ComplianceRate: complianceRate,
		})
	}

	return result, nil
}

func (s *AggregationService) getDataByMonth(month string) ([]model.IndicatorData, error) {
	var dataList []model.IndicatorData
	err := db.DB.Where("month = ?", month).Find(&dataList).Error
	return dataList, err
}

func (s *AggregationService) getDataByQuarter(quarter string) ([]model.IndicatorData, error) {
	months := getQuarterMonths(quarter)
	var dataList []model.IndicatorData
	err := db.DB.Where("month IN ?", months).Find(&dataList).Error
	return dataList, err
}

func (s *AggregationService) getDataByYear(year string) ([]model.IndicatorData, error) {
	var dataList []model.IndicatorData
	err := db.DB.Where("month LIKE ?", year+"-%").Find(&dataList).Error
	return dataList, err
}

func generateMonthPeriods(from, to string) []string {
	var periods []string
	start, _ := time.Parse("2006-01", from)
	end, _ := time.Parse("2006-01", to)

	for current := start; !current.After(end); current = current.AddDate(0, 1, 0) {
		periods = append(periods, current.Format("2006-01"))
	}
	return periods
}

func generateQuarterPeriods(from, to string) []string {
	var periods []string
	startYear, _ := time.Parse("2006-01", from)
	endYear, _ := time.Parse("2006-01", to)

	for y := startYear.Year(); y <= endYear.Year(); y++ {
		for q := 1; q <= 4; q++ {
			period := fmt.Sprintf("%d-Q%d", y, q)
			periods = append(periods, period)
		}
	}
	return periods
}

func generateYearPeriods(from, to string) []string {
	var periods []string
	startYear, _ := time.Parse("2006", from)
	endYear, _ := time.Parse("2006", to)

	for y := startYear.Year(); y <= endYear.Year(); y++ {
		periods = append(periods, fmt.Sprintf("%d", y))
	}
	return periods
}

func getQuarterMonths(quarter string) []string {
	// Format: YYYY-Qn
	var year int
	var q int
	fmt.Sscanf(quarter, "%d-Q%d", &year, &q)

	startMonth := (q-1)*3 + 1
	months := make([]string, 3)
	for i := 0; i < 3; i++ {
		months[i] = fmt.Sprintf("%04d-%02d", year, startMonth+i)
	}
	return months
}
