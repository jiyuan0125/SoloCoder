package handlers

import (
	"encoding/json"
	"net/http"
	"ohims/internal/models"
	"ohims/pkg/db"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type WorkerHandler struct{}

func (h *WorkerHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	entID := r.URL.Query().Get("enterprise_id")

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var workers []models.Worker
	var total int64

	query := db.GetDB().Model(&models.Worker{})
	if entID != "" {
		query = query.Where("enterprise_id = ?", entID)
	}
	query.Count(&total)

	offset := (page - 1) * size
	query.Preload("ExposedFactors").Offset(offset).Limit(size).Find(&workers)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       workers,
		"page":       page,
		"size":       size,
		"total":      total,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (h *WorkerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var worker models.Worker
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if worker.BirthDate.IsZero() {
		worker.BirthDate = time.Now().AddDate(-30, 0, 0)
	}
	if worker.EntryDate.IsZero() {
		worker.EntryDate = time.Now()
	}
	worker.Status = "在职"
	worker.CreatedAt = time.Now()
	worker.UpdatedAt = time.Now()

	if err := db.GetDB().Create(&worker).Error; err != nil {
		http.Error(w, `{"error":"创建劳动者失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(worker)
}

func (h *WorkerHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/workers/")
	if strings.Contains(path, "/sub") {
		return
	}
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		return
	}

	var worker models.Worker
	if err := db.GetDB().Preload("ExposedFactors").First(&worker, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"劳动者不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worker)
}

func (h *WorkerHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/workers/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var worker models.Worker
	if err := db.GetDB().First(&worker, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"劳动者不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	var updateData models.Worker
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	worker.Name = updateData.Name
	worker.IDCard = updateData.IDCard
	worker.Gender = updateData.Gender
	if !updateData.BirthDate.IsZero() {
		worker.BirthDate = updateData.BirthDate
	}
	worker.PostName = updateData.PostName
	worker.Workshop = updateData.Workshop
	if !updateData.EntryDate.IsZero() {
		worker.EntryDate = updateData.EntryDate
	}
	if updateData.Status != "" {
		worker.Status = updateData.Status
	}
	worker.UpdatedAt = time.Now()

	db.GetDB().Save(&worker)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worker)
}

func (h *WorkerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/workers/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	tx := db.GetDB().Begin()
	tx.Where("worker_id = ?", id).Delete(&models.WorkerHazardFactor{})
	tx.Delete(&models.Worker{}, id)
	tx.Commit()

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkerHandler) AddExposedFactors(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/workers/")
	path = strings.TrimSuffix(path, "/exposed-factors")
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var factors []models.WorkerHazardFactor
	if err := json.NewDecoder(r.Body).Decode(&factors); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	tx := db.GetDB().Begin()
	tx.Where("worker_id = ?", id).Delete(&models.WorkerHazardFactor{})
	for _, f := range factors {
		f.WorkerID = uint(id)
		tx.Create(&f)
	}
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
