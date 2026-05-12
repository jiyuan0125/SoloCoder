package handlers

import (
	"encoding/json"
	"net/http"
	"ohims/internal/models"
	"ohims/pkg/db"
	"ohims/pkg/validation"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type HazardHandler struct{}

func (h *HazardHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	entID := r.URL.Query().Get("enterprise_id")

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var factors []models.HazardFactor
	var total int64

	query := db.GetDB().Model(&models.HazardFactor{})
	if entID != "" {
		query = query.Where("enterprise_id = ?", entID)
	}
	query.Count(&total)

	offset := (page - 1) * size
	query.Offset(offset).Limit(size).Find(&factors)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       factors,
		"page":       page,
		"size":       size,
		"total":      total,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (h *HazardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var factor models.HazardFactor
	if err := json.NewDecoder(r.Body).Decode(&factor); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if !validation.IsValidHazardLevel(factor.HazardLevel) {
		http.Error(w, `{"error":"危害等级不在范围内，应为：一般、较重、严重"}`, http.StatusBadRequest)
		return
	}

	if !validation.IsValidCategory(factor.Category) {
		http.Error(w, `{"error":"危害因素类别无效"}`, http.StatusBadRequest)
		return
	}

	if factor.LastMonitorDate.IsZero() {
		factor.LastMonitorDate = time.Now()
	}

	limit, unit := validation.GetExposureLimit(factor.FactorName)
	if limit > 0 {
		factor.ExposureLimit = limit
		if unit != "" {
			factor.MonitorUnit = unit
		}
		factor.ExceedLimit = validation.IsExceedingLimit(factor.FactorName, factor.LastMonitorValue)
	}

	factor.CreatedAt = time.Now()
	factor.UpdatedAt = time.Now()

	if err := db.GetDB().Create(&factor).Error; err != nil {
		http.Error(w, `{"error":"创建危害因素失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(factor)
}

func (h *HazardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/hazard-factors/"), 10, 64)
	if err != nil || strings.Contains(r.URL.Path, "/sub") {
		return
	}

	var factor models.HazardFactor
	if err := db.GetDB().First(&factor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"危害因素不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(factor)
}

func (h *HazardHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/hazard-factors/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var factor models.HazardFactor
	if err := db.GetDB().First(&factor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"危害因素不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	var updateData models.HazardFactor
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if updateData.HazardLevel != "" && !validation.IsValidHazardLevel(updateData.HazardLevel) {
		http.Error(w, `{"error":"危害等级不在范围内"}`, http.StatusBadRequest)
		return
	}

	factor.Workshop = updateData.Workshop
	factor.PostName = updateData.PostName
	if updateData.Category != "" {
		factor.Category = updateData.Category
	}
	if updateData.FactorName != "" {
		factor.FactorName = updateData.FactorName
	}
	if updateData.HazardLevel != "" {
		factor.HazardLevel = updateData.HazardLevel
	}
	factor.ProtectiveMeasures = updateData.ProtectiveMeasures
	factor.LastMonitorValue = updateData.LastMonitorValue
	if !updateData.LastMonitorDate.IsZero() {
		factor.LastMonitorDate = updateData.LastMonitorDate
	}

	limit, _ := validation.GetExposureLimit(factor.FactorName)
	if limit > 0 {
		factor.ExposureLimit = limit
		factor.ExceedLimit = validation.IsExceedingLimit(factor.FactorName, factor.LastMonitorValue)
	}

	factor.UpdatedAt = time.Now()

	if err := db.GetDB().Save(&factor).Error; err != nil {
		http.Error(w, `{"error":"更新失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(factor)
}

func (h *HazardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/hazard-factors/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	if err := db.GetDB().Delete(&models.HazardFactor{}, id).Error; err != nil {
		http.Error(w, `{"error":"删除失败"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HazardHandler) UpdateMonitor(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/hazard-factors/")
	idStr = strings.TrimSuffix(idStr, "/monitor")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var data struct {
		MonitorValue float64 `json:"monitor_value"`
		MonitorDate  string  `json:"monitor_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	var factor models.HazardFactor
	if err := db.GetDB().First(&factor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"危害因素不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	factor.LastMonitorValue = data.MonitorValue
	if data.MonitorDate != "" {
		if t, err := time.Parse("2006-01-02", data.MonitorDate); err == nil {
			factor.LastMonitorDate = t
		} else {
			factor.LastMonitorDate = time.Now()
		}
	} else {
		factor.LastMonitorDate = time.Now()
	}

	limit, _ := validation.GetExposureLimit(factor.FactorName)
	if limit > 0 {
		factor.ExposureLimit = limit
		factor.ExceedLimit = validation.IsExceedingLimit(factor.FactorName, factor.LastMonitorValue)
	}

	factor.UpdatedAt = time.Now()
	db.GetDB().Save(&factor)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(factor)
}
