package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ohims/internal/models"
	"ohims/pkg/db"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ReportHandler struct{}

func (h *ReportHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	entID := r.URL.Query().Get("enterprise_id")

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var reports []models.EnterpriseReport
	var total int64

	query := db.GetDB().Model(&models.EnterpriseReport{})
	if entID != "" {
		query = query.Where("enterprise_id = ?", entID)
	}
	query.Count(&total)

	offset := (page - 1) * size
	query.Preload("Enterprise").Order("year desc").Offset(offset).Limit(size).Find(&reports)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       reports,
		"page":       page,
		"size":       size,
		"total":      total,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (h *ReportHandler) GenerateEnterpriseReport(w http.ResponseWriter, r *http.Request) {
	var data struct {
		EnterpriseID uint `json:"enterprise_id"`
		Year         int  `json:"year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	var ent models.Enterprise
	if err := db.GetDB().First(&ent, data.EnterpriseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"企业不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	var factors []models.HazardFactor
	db.GetDB().Where("enterprise_id = ?", data.EnterpriseID).Find(&factors)

	var workers []models.Worker
	db.GetDB().Where("enterprise_id = ?", data.EnterpriseID).Find(&workers)

	startDate := time.Date(data.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(data.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	var exams []models.Examination
	db.GetDB().Where("enterprise_id = ? AND exam_date >= ? AND exam_date < ?",
		data.EnterpriseID, startDate, endDate).Find(&exams)

	totalWorkers := len(workers)
	totalExamined := 0
	abnormalCount := 0
	suspectedCount := 0

	for _, e := range exams {
		if e.Status == "已完成" {
			totalExamined++
			if e.HasAbnormal {
				abnormalCount++
			}
			if e.IsSuspected {
				suspectedCount++
			}
		}
	}

	content := fmt.Sprintf(`
企业职业健康监护总结报告
========================
企业名称: %s
统一社会信用代码: %s
报告年度: %d

一、企业基本情况
----------------
行业: %s
地区: %s
地址: %s
联系人: %s
联系电话: %s

二、危害因素分布
----------------
`, ent.Name, ent.UnifiedSocialCode, data.Year, ent.Industry, ent.Region, ent.Address, ent.ContactPerson, ent.ContactPhone)

	for _, f := range factors {
		exceedStatus := "正常"
		if f.ExceedLimit {
			exceedStatus = "超标"
		}
		content += fmt.Sprintf("- %s/%s: %s (等级: %s, 监测值: %.2f%s/%s, 限值: %.2f, 状态: %s)\n",
			f.Workshop, f.PostName, f.FactorName, f.HazardLevel,
			f.LastMonitorValue, f.MonitorUnit, f.LastMonitorDate.Format("2006-01-02"),
			f.ExposureLimit, exceedStatus)
	}

	coverage := 0.0
	if totalWorkers > 0 {
		coverage = float64(totalExamined) / float64(totalWorkers) * 100
	}

	content += fmt.Sprintf(`
三、职业健康检查情况
--------------------
接触危害因素人数: %d
应体检人数: %d
已体检人数: %d
体检覆盖率: %.1f%%
体检异常人数: %d
疑似职业病人数: %d
`, totalWorkers, totalWorkers, totalExamined, coverage, abnormalCount, suspectedCount)

	report := models.EnterpriseReport{
		EnterpriseID: data.EnterpriseID,
		Year:         data.Year,
		Content:      content,
		GeneratedAt:  time.Now(),
	}

	var existing models.EnterpriseReport
	if err := db.GetDB().Where("enterprise_id = ? AND year = ?", data.EnterpriseID, data.Year).First(&existing).Error; err == nil {
		existing.Content = content
		existing.GeneratedAt = time.Now()
		db.GetDB().Save(&existing)
		report = existing
	} else {
		db.GetDB().Create(&report)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ReportHandler) GetRegulatoryStats(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	if year == 0 {
		year = time.Now().Year()
	}

	var enterprises []models.Enterprise
	entQuery := db.GetDB()
	if region != "" {
		entQuery = entQuery.Where("region = ?", region)
	}
	entQuery.Find(&enterprises)

	entIDs := make([]uint, 0, len(enterprises))
	for _, e := range enterprises {
		entIDs = append(entIDs, e.ID)
	}

	var workers []models.Worker
	if len(entIDs) > 0 {
		db.GetDB().Where("enterprise_id IN ?", entIDs).Find(&workers)
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	var exams []models.Examination
	examQuery := db.GetDB().Where("exam_date >= ? AND exam_date < ?", startDate, endDate)
	if len(entIDs) > 0 {
		examQuery = examQuery.Where("enterprise_id IN ?", entIDs)
	}
	examQuery.Find(&exams)

	totalEnterprises := len(enterprises)
	totalExposed := len(workers)

	examinedSet := make(map[uint]bool)
	abnormalCount := 0
	suspectedCount := 0

	for _, e := range exams {
		if e.Status == "已完成" {
			examinedSet[e.WorkerID] = true
			if e.HasAbnormal {
				abnormalCount++
			}
			if e.IsSuspected {
				suspectedCount++
			}
		}
	}

	totalExamined := len(examinedSet)
	coverage := 0.0
	if totalExposed > 0 {
		coverage = float64(totalExamined) / float64(totalExposed) * 100
	}

	result := map[string]interface{}{
		"region":              region,
		"year":                year,
		"total_enterprises":   totalEnterprises,
		"exposed_workers":     totalExposed,
		"examined_workers":    totalExamined,
		"should_examine":      totalExposed,
		"coverage_rate":       fmt.Sprintf("%.1f%%", coverage),
		"coverage_rate_value": coverage,
		"abnormal_count":      abnormalCount,
		"suspected_count":     suspectedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *ReportHandler) GetAggregatedMetrics(w http.ResponseWriter, r *http.Request) {
	industryCoverage := h.calculateIndustryCoverage()
	factorExceedRates := h.calculateFactorExceedRates()
	diseaseDetectionRates := h.calculateDiseaseDetectionRates()

	result := map[string]interface{}{
		"industry_coverage":       industryCoverage,
		"factor_exceed_rates":     factorExceedRates,
		"disease_detection_rates": diseaseDetectionRates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *ReportHandler) calculateIndustryCoverage() []map[string]interface{} {
	type industryStat struct {
		Industry string
	}

	rows, err := db.GetDB().Raw(`
		SELECT 
			e.industry,
			COUNT(DISTINCT w.id) as total_workers,
			COUNT(DISTINCT CASE WHEN ex.status = '已完成' THEN w.id END) as examined_workers
		FROM enterprises e
		LEFT JOIN workers w ON w.enterprise_id = e.id
		LEFT JOIN examinations ex ON ex.worker_id = w.id
		GROUP BY e.industry
	`).Rows()
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		var industry string
		var total, examined int
		rows.Scan(&industry, &total, &examined)
		coverage := 0.0
		if total > 0 {
			coverage = float64(examined) / float64(total) * 100
		}
		if industry == "" {
			industry = "其他"
		}
		result = append(result, map[string]interface{}{
			"industry":    industry,
			"total":       total,
			"examined":    examined,
			"coverage":    coverage,
			"coverage_str": fmt.Sprintf("%.1f%%", coverage),
		})
	}

	return result
}

func (h *ReportHandler) calculateFactorExceedRates() []map[string]interface{} {
	rows, err := db.GetDB().Raw(`
		SELECT 
			factor_name,
			COUNT(*) as total,
			SUM(CASE WHEN exceed_limit = 1 THEN 1 ELSE 0 END) as exceeded
		FROM hazard_factors
		GROUP BY factor_name
	`).Rows()
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		var factor string
		var total, exceeded int
		rows.Scan(&factor, &total, &exceeded)
		rate := 0.0
		if total > 0 {
			rate = float64(exceeded) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"factor":       factor,
			"total":        total,
			"exceeded":     exceeded,
			"rate":         rate,
			"rate_str":     fmt.Sprintf("%.1f%%", rate),
		})
	}

	return result
}

func (h *ReportHandler) calculateDiseaseDetectionRates() []map[string]interface{} {
	rows, err := db.GetDB().Raw(`
		SELECT 
			ex.exam_type,
			COUNT(*) as total_exams,
			SUM(CASE WHEN ex.is_suspected = 1 THEN 1 ELSE 0 END) as suspected
		FROM examinations ex
		WHERE ex.status = '已完成'
		GROUP BY ex.exam_type
	`).Rows()
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		var examType string
		var total, suspected int
		rows.Scan(&examType, &total, &suspected)
		rate := 0.0
		if total > 0 {
			rate = float64(suspected) / float64(total) * 100
		}
		if examType == "" {
			examType = "其他"
		}
		result = append(result, map[string]interface{}{
			"exam_type":   examType,
			"total":       total,
			"suspected":   suspected,
			"rate":        rate,
			"rate_str":    fmt.Sprintf("%.1f%%", rate),
		})
	}

	return result
}

func (h *ReportHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/reports/")
	if strings.Contains(idStr, "/sub") {
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return
	}

	var report models.EnterpriseReport
	if err := db.GetDB().Preload("Enterprise").First(&report, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"报告不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ReportHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/reports/")
	idStr = strings.TrimSuffix(idStr, "/export")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var report models.EnterpriseReport
	if err := db.GetDB().Preload("Enterprise").First(&report, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"报告不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="report-%s-%d.txt"`, report.Enterprise.Name, report.Year))
	w.Write([]byte(report.Content))
}
