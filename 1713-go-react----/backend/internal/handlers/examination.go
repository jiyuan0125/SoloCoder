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

type ExaminationHandler struct{}

func (h *ExaminationHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	workerID := r.URL.Query().Get("worker_id")
	entID := r.URL.Query().Get("enterprise_id")
	hasAbnormal := r.URL.Query().Get("has_abnormal")
	isSuspected := r.URL.Query().Get("is_suspected")

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var exams []models.Examination
	var total int64

	query := db.GetDB().Model(&models.Examination{})
	if workerID != "" {
		query = query.Where("worker_id = ?", workerID)
	}
	if entID != "" {
		query = query.Where("enterprise_id = ?", entID)
	}
	if hasAbnormal == "true" {
		query = query.Where("has_abnormal = ?", true)
	}
	if isSuspected == "true" {
		query = query.Where("is_suspected = ?", true)
	}
	query.Count(&total)

	offset := (page - 1) * size
	query.Preload("Worker").Preload("Enterprise").Preload("ExamItems").
		Order("created_at desc").Offset(offset).Limit(size).Find(&exams)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       exams,
		"page":       page,
		"size":       size,
		"total":      total,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (h *ExaminationHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var data struct {
		WorkerID      uint   `json:"worker_id"`
		EnterpriseID  uint   `json:"enterprise_id"`
		ExamType      string `json:"exam_type"`
		ScheduledDate string `json:"scheduled_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	var worker models.Worker
	if err := db.GetDB().Preload("ExposedFactors").First(&worker, data.WorkerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"劳动者不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	categories := make([]string, 0)
	factorNames := make([]string, 0)
	for _, ef := range worker.ExposedFactors {
		categories = append(categories, ef.Category)
		factorNames = append(factorNames, ef.FactorName)
	}

	if len(factorNames) == 0 {
		var factors []models.HazardFactor
		db.GetDB().Where("enterprise_id = ? AND post_name = ?", data.EnterpriseID, worker.PostName).Find(&factors)
		for _, f := range factors {
			categories = append(categories, f.Category)
			factorNames = append(factorNames, f.FactorName)
		}
	}

	cycleMonths := validation.CalculateExamCycleMonths(categories, factorNames)

	var scheduledDate time.Time
	if data.ScheduledDate != "" {
		if t, err := time.Parse("2006-01-02", data.ScheduledDate); err == nil {
			scheduledDate = t
		} else {
			scheduledDate = time.Now()
		}
	} else {
		scheduledDate = time.Now()
	}

	exam := models.Examination{
		WorkerID:        data.WorkerID,
		EnterpriseID:    data.EnterpriseID,
		ExamType:        data.ExamType,
		ScheduledDate:   scheduledDate,
		Status:          "已安排",
		FlowStatus:      "等待",
		ExamCycleMonths: cycleMonths,
		NextExamDate:    scheduledDate.AddDate(0, cycleMonths, 0),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := db.GetDB().Create(&exam).Error; err != nil {
		http.Error(w, `{"error":"安排体检失败"}`, http.StatusInternalServerError)
		return
	}

	items := validation.GetExamItemsForFactors(factorNames)
	for _, item := range items {
		result := models.ExamItemResult{
			ExaminationID: exam.ID,
			ItemCode:      item.Code,
			ItemName:      item.Name,
			Result:        "",
			IsAbnormal:    false,
			IsSuspected:   false,
		}
		db.GetDB().Create(&result)
	}

	exam.ExamItems = make([]models.ExamItemResult, 0)
	db.GetDB().Where("examination_id = ?", exam.ID).Find(&exam.ExamItems)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exam)
}

func (h *ExaminationHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/examinations/")
	if strings.Contains(path, "/sub") {
		return
	}
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		return
	}

	var exam models.Examination
	if err := db.GetDB().Preload("Worker").Preload("Enterprise").Preload("ExamItems").
		First(&exam, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"体检记录不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exam)
}

func (h *ExaminationHandler) RecordResult(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/examinations/")
	path = strings.TrimSuffix(path, "/results")
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var exam models.Examination
	if err := db.GetDB().Preload("Worker").Preload("ExamItems").First(&exam, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"体检记录不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	if exam.FlowStatus == "归档" {
		http.Error(w, `{"error":"归档后不可修改"}`, http.StatusForbidden)
		return
	}

	var data struct {
		ExamItems []models.ExamItemResult `json:"exam_items"`
		ExamDate  string                  `json:"exam_date"`
		Remarks   string                  `json:"remarks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	var worker models.Worker
	db.GetDB().Preload("ExposedFactors").First(&worker, exam.WorkerID)
	factorNames := make([]string, 0)
	for _, ef := range worker.ExposedFactors {
		factorNames = append(factorNames, ef.FactorName)
	}

	if len(factorNames) == 0 {
		var factors []models.HazardFactor
		db.GetDB().Where("enterprise_id = ? AND post_name = ?", exam.EnterpriseID, worker.PostName).Find(&factors)
		for _, f := range factors {
			factorNames = append(factorNames, f.FactorName)
		}
	}

	for _, item := range data.ExamItems {
		if item.ItemCode != "" && !validation.IsValidExamItemCode(item.ItemCode, factorNames) {
			http.Error(w, `{"error":"体检项目 "+item.ItemCode+" 不在允许的列表中"}`, http.StatusBadRequest)
			return
		}
	}

	hasAbnormal := false
	isSuspected := false
	for _, item := range data.ExamItems {
		for i := range exam.ExamItems {
			if exam.ExamItems[i].ItemCode == item.ItemCode {
				exam.ExamItems[i].Result = item.Result
				exam.ExamItems[i].IsAbnormal = item.IsAbnormal
				exam.ExamItems[i].IsSuspected = item.IsSuspected
				exam.ExamItems[i].Remarks = item.Remarks
				if item.IsAbnormal {
					hasAbnormal = true
				}
				if item.IsSuspected {
					isSuspected = true
				}
			}
		}
	}

	exam.HasAbnormal = hasAbnormal
	exam.IsSuspected = isSuspected
	exam.Status = "已完成"
	exam.FlowStatus = "检查"
	if data.ExamDate != "" {
		if t, err := time.Parse("2006-01-02", data.ExamDate); err == nil {
			exam.ExamDate = t
		}
	} else {
		exam.ExamDate = time.Now()
	}
	exam.Remarks = data.Remarks
	exam.UpdatedAt = time.Now()

	tx := db.GetDB().Begin()

	if err := tx.Save(&exam).Error; err != nil {
		tx.Rollback()
		http.Error(w, `{"error":"保存结果失败"}`, http.StatusInternalServerError)
		return
	}

	for i := range exam.ExamItems {
		if err := tx.Save(&exam.ExamItems[i]).Error; err != nil {
			tx.Rollback()
			http.Error(w, `{"error":"保存项目结果失败"}`, http.StatusInternalServerError)
			return
		}
	}

	if hasAbnormal && !isSuspected {
		todo := models.TodoItem{
			Type:        "复查确认",
			Title:       "劳动者 " + worker.Name + " 体检异常复查确认",
			Description: "体检发现异常，请安排复查确认",
			Assignee:    "职业健康检查医师",
			DueDate:     time.Now().AddDate(0, 0, 14),
			Status:      "待处理",
			RelatedID:   exam.ID,
			RelatedType: "examination",
			Priority:    "中",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.Create(&todo).Error; err != nil {
			tx.Rollback()
			http.Error(w, `{"error":"创建待办失败"}`, http.StatusInternalServerError)
			return
		}
	}

	if isSuspected {
		todo := models.TodoItem{
			Type:        "职业病诊断",
			Title:       "劳动者 " + worker.Name + " 疑似职业病诊断",
			Description: "体检发现疑似职业病，请职业病诊断医师处理",
			Assignee:    "职业病诊断医师",
			DueDate:     time.Now().AddDate(0, 1, 0),
			Status:      "待处理",
			RelatedID:   exam.ID,
			RelatedType: "examination",
			Priority:    "高",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.Create(&todo).Error; err != nil {
			tx.Rollback()
			http.Error(w, `{"error":"创建待办失败"}`, http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exam)
}

func (h *ExaminationHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/examinations/")
	path = strings.TrimSuffix(path, "/flow")
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var data struct {
		FlowStatus string `json:"flow_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	var exam models.Examination
	if err := db.GetDB().First(&exam, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"体检记录不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	validTransitions := map[string][]string{
		"登记":    {"等待"},
		"等待":    {"进行中"},
		"进行中":  {"检查"},
		"检查":    {"进行中", "完成"},
		"完成":    {"归档"},
		"归档":    {},
	}

	valid, ok := validTransitions[exam.FlowStatus]
	if !ok {
		http.Error(w, `{"error":"无效的状态"}`, http.StatusBadRequest)
		return
	}

	canTransition := false
	for _, s := range valid {
		if s == data.FlowStatus {
			canTransition = true
			break
		}
	}

	if !canTransition && data.FlowStatus != exam.FlowStatus {
		http.Error(w, `{"error":"不能从当前状态 "+exam.FlowStatus+" 转换到 "+data.FlowStatus}`, http.StatusBadRequest)
		return
	}

	exam.FlowStatus = data.FlowStatus
	exam.UpdatedAt = time.Now()
	db.GetDB().Save(&exam)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exam)
}

func (h *ExaminationHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/examinations/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var exam models.Examination
	if err := db.GetDB().First(&exam, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"体检记录不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	if exam.FlowStatus == "归档" {
		http.Error(w, `{"error":"归档后不可修改"}`, http.StatusForbidden)
		return
	}

	var updateData models.Examination
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if updateData.ScheduledDate != exam.ScheduledDate {
		exam.ScheduledDate = updateData.ScheduledDate
	}
	if updateData.ExamType != "" {
		exam.ExamType = updateData.ExamType
	}
	if updateData.Status != "" {
		exam.Status = updateData.Status
	}
	exam.UpdatedAt = time.Now()

	db.GetDB().Save(&exam)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exam)
}

func (h *ExaminationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/examinations/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	tx := db.GetDB().Begin()
	tx.Where("examination_id = ?", id).Delete(&models.ExamItemResult{})
	tx.Delete(&models.Examination{}, id)
	tx.Commit()

	w.WriteHeader(http.StatusNoContent)
}
