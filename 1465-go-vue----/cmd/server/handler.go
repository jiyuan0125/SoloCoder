package main

import (
	"encoding/json"
	"net/http"
	"time"

	"envmonitor/internal/common"
	"envmonitor/internal/core"
)

type Handler struct {
	service *core.Service
	mux     *http.ServeMux
}

func NewHandler(service *core.Service) *Handler {
	h := &Handler{
		service: service,
		mux:     http.NewServeMux(),
	}
	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/api/v1/gas/outlets", h.registerGasOutlet)
	h.mux.HandleFunc("/api/v1/gas/reports", h.handleGasReports)
	h.mux.HandleFunc("/api/v1/wastewater/outlets", h.registerWastewaterOutlet)
	h.mux.HandleFunc("/api/v1/wastewater/reports", h.handleWastewaterReports)
	h.mux.HandleFunc("/api/v1/wastewater/daily", h.getDailyReport)
	h.mux.HandleFunc("/api/v1/solid-waste", h.handleSolidWaste)
	h.mux.HandleFunc("/api/v1/alarms", h.handleAlarms)
	h.mux.HandleFunc("/api/v1/alarms/resolve", h.resolveAlarm)
}

func (h *Handler) Start(addr string) error {
	return http.ListenAndServe(addr, h.mux)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) registerGasOutlet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
		return
	}

	var req common.RegisterGasOutletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "invalid request body"))
		return
	}

	if req.ID == "" {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "id is required"))
		return
	}

	outlet := h.service.RegisterGasOutlet(req.ID, req.Location)
	h.writeJSON(w, http.StatusOK, common.Success(convertGasOutlet(outlet)))
}

func (h *Handler) handleGasReports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.submitGasReport(w, r)
	case http.MethodGet:
		h.getGasReports(w, r)
	default:
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
	}
}

func (h *Handler) submitGasReport(w http.ResponseWriter, r *http.Request) {
	var req common.GasReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "invalid request body"))
		return
	}

	reportedAt := req.ReportedAt
	if reportedAt.IsZero() {
		reportedAt = time.Now()
	}

	report, err := h.service.ProcessGasReport(req.OutletID, reportedAt, req.Measurements)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, err.Error()))
		return
	}

	h.writeJSON(w, http.StatusOK, common.Success(convertGasReport(report)))
}

func (h *Handler) getGasReports(w http.ResponseWriter, r *http.Request) {
	outletID := r.URL.Query().Get("outlet_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time
	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -7)
	}
	if to.IsZero() {
		to = time.Now()
	}

	reports := h.service.GetGasReports(outletID, from, to)
	responses := make([]*common.GasReportResponse, 0, len(reports))
	for _, report := range reports {
		responses = append(responses, convertGasReport(report))
	}

	h.writeJSON(w, http.StatusOK, common.Success(responses))
}

func (h *Handler) registerWastewaterOutlet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
		return
	}

	var req common.RegisterWastewaterOutletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "invalid request body"))
		return
	}

	if req.ID == "" {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "id is required"))
		return
	}

	h.service.RegisterWastewaterOutlet(req.ID)
	h.writeJSON(w, http.StatusOK, common.Success(map[string]string{"id": req.ID}))
}

func (h *Handler) handleWastewaterReports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.submitWastewaterReport(w, r)
	case http.MethodGet:
		h.getWastewaterReports(w, r)
	default:
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
	}
}

func (h *Handler) submitWastewaterReport(w http.ResponseWriter, r *http.Request) {
	var req common.WastewaterReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "invalid request body"))
		return
	}

	reportedAt := req.ReportedAt
	if reportedAt.IsZero() {
		reportedAt = time.Now()
	}

	report, err := h.service.ProcessWastewaterReport(req.OutletID, reportedAt, req.Measurements)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, err.Error()))
		return
	}

	h.writeJSON(w, http.StatusOK, common.Success(convertWastewaterReport(report)))
}

func (h *Handler) getWastewaterReports(w http.ResponseWriter, r *http.Request) {
	outletID := r.URL.Query().Get("outlet_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time
	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -7)
	}
	if to.IsZero() {
		to = time.Now()
	}

	reports := h.service.GetWastewaterReports(outletID, from, to)
	responses := make([]*common.WastewaterReportResponse, 0, len(reports))
	for _, report := range reports {
		responses = append(responses, convertWastewaterReport(report))
	}

	h.writeJSON(w, http.StatusOK, common.Success(responses))
}

func (h *Handler) getDailyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
		return
	}

	outletID := r.URL.Query().Get("outlet_id")
	dateStr := r.URL.Query().Get("date")

	if outletID == "" {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "outlet_id is required"))
		return
	}

	var date time.Time
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = t
		}
	}
	if date.IsZero() {
		date = time.Now()
	}

	report, exists := h.service.GetWastewaterDailyReport(outletID, date)
	if !exists {
		h.writeJSON(w, http.StatusNotFound, common.Error(404, "daily report not found"))
		return
	}

	h.writeJSON(w, http.StatusOK, common.Success(convertDailyReport(report)))
}

func (h *Handler) handleSolidWaste(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.addSolidWasteRecord(w, r)
	case http.MethodGet:
		h.getSolidWasteRecords(w, r)
	default:
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
	}
}

func (h *Handler) addSolidWasteRecord(w http.ResponseWriter, r *http.Request) {
	var req common.AddSolidWasteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "invalid request body"))
		return
	}

	generatedAt := req.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}

	record, err := h.service.AddSolidWasteRecord(
		req.Name,
		core.WasteCategory(req.Category),
		req.Amount,
		req.StorageLocation,
		generatedAt,
	)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, err.Error()))
		return
	}

	h.writeJSON(w, http.StatusOK, common.Success(convertSolidWasteRecord(record)))
}

func (h *Handler) getSolidWasteRecords(w http.ResponseWriter, r *http.Request) {
	categoryStr := r.URL.Query().Get("category")
	statusStr := r.URL.Query().Get("status")

	var category *core.WasteCategory
	var status *core.WasteStatus

	if categoryStr != "" {
		if c, err := strconvToInt(categoryStr); err == nil {
			cat := core.WasteCategory(c)
			category = &cat
		}
	}

	if statusStr != "" {
		if s, err := strconvToInt(statusStr); err == nil {
			st := core.WasteStatus(s)
			status = &st
		}
	}

	records := h.service.GetSolidWasteRecords(category, status)
	responses := make([]*common.SolidWasteRecordResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, convertSolidWasteRecord(record))
	}

	h.writeJSON(w, http.StatusOK, common.Success(responses))
}

func (h *Handler) handleAlarms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
		return
	}

	entityTypeStr := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	levelStr := r.URL.Query().Get("level")
	resolvedStr := r.URL.Query().Get("resolved")

	var entityType *core.EntityType
	var level *core.AlarmLevel
	var resolved *bool

	if entityTypeStr != "" {
		if et, err := strconvToInt(entityTypeStr); err == nil {
			e := core.EntityType(et)
			entityType = &e
		}
	}

	if levelStr != "" {
		if l, err := strconvToInt(levelStr); err == nil {
			al := core.AlarmLevel(l)
			level = &al
		}
	}

	if resolvedStr != "" {
		if resolvedStr == "true" {
			r := true
			resolved = &r
		} else if resolvedStr == "false" {
			r := false
			resolved = &r
		}
	}

	alarms := h.service.GetAlarms(entityType, entityID, level, resolved)
	responses := make([]*common.AlarmResponse, 0, len(alarms))
	for _, alarm := range alarms {
		responses = append(responses, convertAlarm(alarm))
	}

	h.writeJSON(w, http.StatusOK, common.Success(responses))
}

func (h *Handler) resolveAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, common.Error(405, "method not allowed"))
		return
	}

	var req common.ResolveAlarmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "invalid request body"))
		return
	}

	if req.ID == "" {
		h.writeJSON(w, http.StatusBadRequest, common.Error(400, "id is required"))
		return
	}

	alarm, resolved := h.service.ResolveAlarm(req.ID)
	if !resolved {
		h.writeJSON(w, http.StatusNotFound, common.Error(404, "alarm not found or already resolved"))
		return
	}

	h.writeJSON(w, http.StatusOK, common.Success(convertAlarm(alarm)))
}
