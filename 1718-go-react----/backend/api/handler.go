package api

import (
	"clinical-path-backend/models"
	"clinical-path-backend/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{service: s}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	} else {
		errorMsg = http.StatusText(status)
	}
	json.NewEncoder(w).Encode(models.APIResponse{Success: false, Error: errorMsg})
}

func (h *Handler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (h *Handler) HandlePaths(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		paths := h.service.ListPaths()
		writeJSON(w, http.StatusOK, paths)
		return
	}

	if r.Method == http.MethodPost {
		var path models.ClinicalPath
		if err := json.NewDecoder(r.Body).Decode(&path); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		err := h.service.CreatePath(&path)
		if err != nil {
			if strings.Contains(err.Error(), "已存在") {
				writeError(w, http.StatusConflict, err)
			} else if strings.Contains(err.Error(), "不能大于") {
				writeError(w, http.StatusBadRequest, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
		writeJSON(w, http.StatusCreated, path)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandlePathByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/r/")
	idx := strings.Index(idStr, "/")
	var id int64
	var err error
	if idx > 0 {
		id, err = strconv.ParseInt(idStr[:idx], 10, 64)
	} else {
		id, err = strconv.ParseInt(idStr, 10, 64)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if idx > 0 {
		subPath := idStr[idx:]
		if strings.HasPrefix(subPath, "/items") {
			h.HandlePathItems(w, r, id)
			return
		}
		if strings.HasPrefix(subPath, "/stages") {
			h.HandlePathStages(w, r, id)
			return
		}
		if strings.HasPrefix(subPath, "/actions/") {
			h.HandlePathActions(w, r, id, subPath[len("/actions/"):])
			return
		}
		writeError(w, http.StatusNotFound, nil)
		return
	}

	if r.Method == http.MethodGet {
		path, err := h.service.GetPath(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, path)
		return
	}

	if r.Method == http.MethodPut {
		var path models.ClinicalPath
		if err := json.NewDecoder(r.Body).Decode(&path); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		path.ID = id
		err := h.service.UpdatePath(&path)
		if err != nil {
			if strings.Contains(err.Error(), "已存在") {
				writeError(w, http.StatusConflict, err)
			} else if strings.Contains(err.Error(), "不能大于") {
				writeError(w, http.StatusBadRequest, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
		writeJSON(w, http.StatusOK, path)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandlePathStages(w http.ResponseWriter, r *http.Request, pathID int64) {
	if r.Method == http.MethodGet {
		stages := h.service.ListStages(pathID)
		writeJSON(w, http.StatusOK, stages)
		return
	}

	if r.Method == http.MethodPost {
		var stage models.PathStage
		if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		stage.PathID = pathID
		err := h.service.CreateStage(&stage)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, stage)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandlePathItems(w http.ResponseWriter, r *http.Request, pathID int64) {
	if r.Method == http.MethodGet {
		items := h.service.ListItems(pathID)
		writeJSON(w, http.StatusOK, items)
		return
	}

	if r.Method == http.MethodPost {
		var item models.OrderItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.PathID = pathID
		err := h.service.CreateItem(&item)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandlePathActions(w http.ResponseWriter, r *http.Request, pathID int64, action string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	switch action {
	case "approve":
		writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
	case "reject":
		writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
	case "cancel":
		writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
	default:
		writeError(w, http.StatusNotFound, nil)
	}
}

func (h *Handler) HandlePatients(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		patients := h.service.ListPatients()
		writeJSON(w, http.StatusOK, patients)
		return
	}

	if r.Method == http.MethodPost {
		var patient models.Patient
		if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		err := h.service.CreatePatient(&patient)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, patient)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandleEnrollments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		enrollments := h.service.ListEnrollments()
		writeJSON(w, http.StatusOK, enrollments)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			PatientID int64 `json:"patient_id"`
			PathID    int64 `json:"path_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		enrollment, err := h.service.EnrollPatient(req.PatientID, req.PathID)
		if err != nil {
			if strings.Contains(err.Error(), "已在其他路径") {
				writeError(w, http.StatusConflict, err)
			} else if strings.Contains(err.Error(), "不在路径适用范围") {
				writeError(w, http.StatusBadRequest, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
		writeJSON(w, http.StatusCreated, enrollment)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandleEnrollmentByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/enrollments/")
	idx := strings.Index(idStr, "/")
	var id int64
	var err error
	if idx > 0 {
		id, err = strconv.ParseInt(idStr[:idx], 10, 64)
	} else {
		id, err = strconv.ParseInt(idStr, 10, 64)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if idx > 0 {
		subPath := idStr[idx:]
		if strings.HasPrefix(subPath, "/orders") {
			h.HandleDailyOrders(w, r, id)
			return
		}
		if strings.HasPrefix(subPath, "/variations") {
			h.HandleVariations(w, r, id)
			return
		}
		if strings.HasPrefix(subPath, "/exit") {
			h.HandleEnrollmentExit(w, r, id)
			return
		}
		if strings.HasPrefix(subPath, "/complete") {
			h.HandleEnrollmentComplete(w, r, id)
			return
		}
		writeError(w, http.StatusNotFound, nil)
		return
	}

	if r.Method == http.MethodGet {
		enrollment, err := h.service.GetEnrollment(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, enrollment)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandleDailyOrders(w http.ResponseWriter, r *http.Request, enrollmentID int64) {
	if r.Method == http.MethodGet {
		orders := h.service.ListDailyOrders(enrollmentID)
		writeJSON(w, http.StatusOK, orders)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandleOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		Executed bool `json:"executed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	order, err := h.service.ExecuteOrder(id, req.Executed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) HandleVariations(w http.ResponseWriter, r *http.Request, enrollmentID int64) {
	if r.Method == http.MethodGet {
		variations := h.service.ListVariations(enrollmentID)
		writeJSON(w, http.StatusOK, variations)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Date    string `json:"date"`
			Content string `json:"content"`
			Reason  string `json:"reason"`
			Type    string `json:"type"`
			Action  string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			date = time.Now()
		}

		variation, err := h.service.RecordVariation(enrollmentID, date, req.Content, req.Reason, req.Type, req.Action)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, variation)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, nil)
}

func (h *Handler) HandleEnrollmentExit(w http.ResponseWriter, r *http.Request, enrollmentID int64) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req struct {
		ExitReason string `json:"exit_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.ExitEnrollment(enrollmentID, req.ExitReason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "exited"})
}

func (h *Handler) HandleEnrollmentComplete(w http.ResponseWriter, r *http.Request, enrollmentID int64) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req struct {
		ActualDays int     `json:"actual_days"`
		ActualCost float64 `json:"actual_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.CompleteEnrollment(enrollmentID, req.ActualDays, req.ActualCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *Handler) HandleQualityMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, nil)
		return
	}

	now := time.Now()
	year := now.Year()
	month := now.Month()

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}
	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = time.Month(m)
		}
	}

	metrics, err := h.service.CalculateQualityMetrics(year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}
