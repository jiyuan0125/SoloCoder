package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type MetricReport struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Tags  Tags    `json:"tags,omitempty"`
}

type Handler struct {
	store *MetricStore
}

func NewHandler(store *MetricStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/metrics" {
		h.handleReport(w, r)
		return
	}
	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/metrics/") {
		h.handleQuery(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request) {
	var report MetricReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if !IsValidMetricName(report.Name) {
		http.Error(w, "Invalid metric name", http.StatusBadRequest)
		return
	}

	var metricType MetricType
	switch strings.ToLower(report.Type) {
	case "counter":
		metricType = Counter
	case "gauge":
		metricType = Gauge
	case "timer":
		metricType = Timer
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	m := h.store.GetOrCreate(report.Name, metricType)
	m.Record(report.Tags, report.Value, metricType)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleQuery(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/metrics/")
	if name == "" {
		http.Error(w, "Metric name required", http.StatusBadRequest)
		return
	}

	m := h.store.Get(name)
	if m == nil {
		http.NotFound(w, r)
		return
	}

	filter := make(QueryFilter)
	for k, vals := range r.URL.Query() {
		if k == "group_by" {
			continue
		}
		if len(vals) > 0 {
			filter[k] = vals[0]
		}
	}

	groupBy := r.URL.Query().Get("group_by")

	result := m.Query(filter, groupBy)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
