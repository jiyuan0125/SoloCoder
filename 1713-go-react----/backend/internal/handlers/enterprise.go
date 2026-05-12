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

type EnterpriseHandler struct{}

func (h *EnterpriseHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var enterprises []models.Enterprise
	var total int64

	query := db.GetDB().Model(&models.Enterprise{})
	query.Count(&total)

	offset := (page - 1) * size
	query.Preload("HazardFactors").Offset(offset).Limit(size).Find(&enterprises)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       enterprises,
		"page":       page,
		"size":       size,
		"total":      total,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (h *EnterpriseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ent models.Enterprise
	if err := json.NewDecoder(r.Body).Decode(&ent); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	ent.UnifiedSocialCode = strings.ToUpper(strings.TrimSpace(ent.UnifiedSocialCode))
	if !validation.IsValidUSCC(ent.UnifiedSocialCode) {
		http.Error(w, `{"error":"统一社会信用代码格式不正确"}`, http.StatusBadRequest)
		return
	}

	var existing models.Enterprise
	if err := db.GetDB().Where("name = ?", ent.Name).First(&existing).Error; err == nil {
		http.Error(w, `{"error":"企业名称已存在"}`, http.StatusConflict)
		return
	}

	if err := db.GetDB().Where("unified_social_code = ?", ent.UnifiedSocialCode).First(&existing).Error; err == nil {
		http.Error(w, `{"error":"统一社会信用代码已存在"}`, http.StatusConflict)
		return
	}

	ent.Status = "登记"
	ent.CreatedAt = time.Now()
	ent.UpdatedAt = time.Now()

	if err := db.GetDB().Create(&ent).Error; err != nil {
		http.Error(w, `{"error":"创建企业失败"}`, http.StatusInternalServerError)
		return
	}

	h.createTodoForEnterprise(ent.ID, ent.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ent)
}

func (h *EnterpriseHandler) createTodoForEnterprise(entID uint, entName string) {
	todo := models.TodoItem{
		Type:        "危害因素识别",
		Title:       "企业 " + entName + " 危害因素识别",
		Description: "请完成该企业的危害因素识别和登记",
		Assignee:    "职业卫生医师",
		DueDate:     time.Now().AddDate(0, 0, 7),
		Status:      "待处理",
		RelatedID:   entID,
		RelatedType: "enterprise",
		Priority:    "高",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.GetDB().Create(&todo)
}

func (h *EnterpriseHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/api/enterprises/"), 10, 64)
	if err != nil || strings.Contains(r.URL.Path, "/sub") {
		return
	}

	var ent models.Enterprise
	if err := db.GetDB().Preload("HazardFactors").Preload("Workers").First(&ent, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"企业不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ent)
}

func (h *EnterpriseHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/enterprises/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var ent models.Enterprise
	if err := db.GetDB().First(&ent, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"企业不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	if ent.Status == "归档" {
		http.Error(w, `{"error":"归档后不可修改"}`, http.StatusForbidden)
		return
	}

	var updateData models.Enterprise
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if updateData.UnifiedSocialCode != "" {
		updateData.UnifiedSocialCode = strings.ToUpper(strings.TrimSpace(updateData.UnifiedSocialCode))
		if !validation.IsValidUSCC(updateData.UnifiedSocialCode) {
			http.Error(w, `{"error":"统一社会信用代码格式不正确"}`, http.StatusBadRequest)
			return
		}
	}

	ent.Name = updateData.Name
	if updateData.UnifiedSocialCode != "" {
		ent.UnifiedSocialCode = updateData.UnifiedSocialCode
	}
	ent.Industry = updateData.Industry
	ent.Region = updateData.Region
	ent.ContactPerson = updateData.ContactPerson
	ent.ContactPhone = updateData.ContactPhone
	ent.Address = updateData.Address
	if updateData.Status != "" {
		ent.Status = updateData.Status
	}
	ent.UpdatedAt = time.Now()

	if err := db.GetDB().Save(&ent).Error; err != nil {
		http.Error(w, `{"error":"更新失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ent)
}

func (h *EnterpriseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/enterprises/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	if err := db.GetDB().Delete(&models.Enterprise{}, id).Error; err != nil {
		http.Error(w, `{"error":"删除失败"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EnterpriseHandler) GetSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/enterprises/"), "/")
	if len(parts) < 2 {
		http.Error(w, `{"error":"无效的路径"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	subType := parts[1]
	switch subType {
	case "hazard-factors":
		var factors []models.HazardFactor
		db.GetDB().Where("enterprise_id = ?", id).Find(&factors)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": factors})
	case "workers":
		var workers []models.Worker
		db.GetDB().Where("enterprise_id = ?", id).Preload("ExposedFactors").Find(&workers)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": workers})
	case "examinations":
		var exams []models.Examination
		db.GetDB().Where("enterprise_id = ?", id).Preload("Worker").Find(&exams)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": exams})
	case "reports":
		var reports []models.EnterpriseReport
		db.GetDB().Where("enterprise_id = ?", id).Preload("Enterprise").Find(&reports)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": reports})
	default:
		http.Error(w, `{"error":"未知的子资源"}`, http.StatusNotFound)
	}
}
