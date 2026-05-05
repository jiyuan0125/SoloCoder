package server

import (
	"encoding/json"
	"metric-aggregator/pkg/common"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /metrics", h.CreateMetric)
	mux.HandleFunc("DELETE /metrics/{name}", h.DeleteMetric)
	mux.HandleFunc("GET /metrics", h.ListMetrics)
	mux.HandleFunc("GET /metrics/{name}", h.GetMetricInfo)
	mux.HandleFunc("POST /report", h.Report)
	mux.HandleFunc("POST /query", h.Query)
	mux.HandleFunc("GET /health", h.Health)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code common.ErrorCode) {
	resp := common.ErrorResponse{
		Code:    code,
		Message: code.Message(),
	}
	h.writeJSON(w, status, resp)
}

func (h *Handler) writeErrorWithMessage(w http.ResponseWriter, status int, code common.ErrorCode, message string) {
	resp := common.ErrorResponse{
		Code:    code,
		Message: message,
	}
	h.writeJSON(w, status, resp)
}

func (h *Handler) CreateMetric(w http.ResponseWriter, r *http.Request) {
	var req common.CreateMetricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, common.ErrInvalidValue)
		return
	}
	
	if req.Metric == "" {
		h.writeErrorWithMessage(w, http.StatusBadRequest, common.ErrInvalidValue, "指标名称不能为空")
		return
	}
	
	if h.store.MetricExists(req.Metric) {
		h.writeErrorWithMessage(w, http.StatusConflict, common.ErrMetricExists, req.Metric)
		return
	}
	
	if err := h.store.CreateMetric(req.Metric); err != nil {
		if me, ok := err.(*MetricError); ok {
			h.writeErrorWithMessage(w, http.StatusConflict, me.Code, me.Error())
			return
		}
		h.writeError(w, http.StatusInternalServerError, common.ErrInternal)
		return
	}
	
	h.writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "metric": req.Metric})
}

func (h *Handler) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, common.ErrInvalidValue)
		return
	}
	
	if err := h.store.DeleteMetric(name); err != nil {
		if me, ok := err.(*MetricError); ok {
			if me.Code == common.ErrMetricNotFound {
				h.writeErrorWithMessage(w, http.StatusNotFound, me.Code, me.Error())
				return
			}
		}
		h.writeError(w, http.StatusInternalServerError, common.ErrInternal)
		return
	}
	
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "metric": name})
}

func (h *Handler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.store.ListMetrics()
	h.writeJSON(w, http.StatusOK, map[string][]string{"metrics": metrics})
}

func (h *Handler) GetMetricInfo(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, common.ErrInvalidValue)
		return
	}
	
	info, err := h.store.GetMetricInfo(name)
	if err != nil {
		if me, ok := err.(*MetricError); ok {
			if me.Code == common.ErrMetricNotFound {
				h.writeErrorWithMessage(w, http.StatusNotFound, me.Code, me.Error())
				return
			}
		}
		h.writeError(w, http.StatusInternalServerError, common.ErrInternal)
		return
	}
	
	h.writeJSON(w, http.StatusOK, info)
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	var req common.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, common.ErrInvalidValue)
		return
	}
	
	successCount := 0
	outlierCount := 0
	timestamps := make([]int64, 0, len(req.Points))
	
	for _, point := range req.Points {
		if point.Metric == "" {
			continue
		}
		
		success, ts := h.store.ReportPoint(point)
		if success {
			if !common.ValidValue(point.Value) {
				outlierCount++
			} else {
				successCount++
			}
			timestamps = append(timestamps, ts)
		}
	}
	
	resp := common.ReportResponse{
		Success:    successCount,
		Outliers:   outlierCount,
		Timestamps: timestamps,
	}
	
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	var req common.QueryRequest
	
	if r.Method == http.MethodGet {
		if err := h.parseQueryParams(r, &req); err != nil {
			h.writeErrorWithMessage(w, http.StatusBadRequest, common.ErrInvalidValue, err.Error())
			return
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeError(w, http.StatusBadRequest, common.ErrInvalidValue)
			return
		}
	}
	
	results, err := h.store.Query(req)
	if err != nil {
		if me, ok := err.(*MetricError); ok {
			if me.Code == common.ErrMetricNotFound {
				h.writeErrorWithMessage(w, http.StatusNotFound, me.Code, me.Error())
				return
			}
			h.writeErrorWithMessage(w, http.StatusBadRequest, me.Code, me.Error())
			return
		}
		h.writeError(w, http.StatusInternalServerError, common.ErrInternal)
		return
	}
	
	resp := common.QueryResponse{Results: results}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) parseQueryParams(r *http.Request, req *common.QueryRequest) error {
	metricsStr := r.URL.Query().Get("metrics")
	if metricsStr == "" {
		return &MetricError{Code: common.ErrEmptyMetrics}
	}
	req.Metrics = strings.Split(metricsStr, ",")
	
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		return &MetricError{Code: common.ErrInvalidTimestamp}
	}
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return &MetricError{Code: common.ErrInvalidTimestamp}
	}
	req.Start = start
	
	endStr := r.URL.Query().Get("end")
	if endStr == "" {
		return &MetricError{Code: common.ErrInvalidTimestamp}
	}
	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return &MetricError{Code: common.ErrInvalidTimestamp}
	}
	req.End = end
	
	req.Granularity = common.Granularity(r.URL.Query().Get("granularity"))
	if req.Granularity == "" {
		req.Granularity = common.Granularity1Min
	}
	
	req.Aggregation = common.AggregationType(r.URL.Query().Get("aggregation"))
	if req.Aggregation == "" {
		req.Aggregation = common.AggregationAvg
	}
	
	return nil
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
