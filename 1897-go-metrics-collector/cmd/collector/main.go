package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"metrics-collector/internal/metrics"
)

type Server struct {
	registry  *metrics.Registry
	store     *metrics.Store
	qps       *metrics.QPSCounter
}

func NewServer() *Server {
	return &Server{
		registry: metrics.NewRegistry(),
		store:    metrics.NewStore(),
		qps:      metrics.NewQPSCounter(),
	}
}

type RegisterRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ReportRequest struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
}

func (s *Server) parseMetricType(typ string) (metrics.MetricType, error) {
	switch strings.ToLower(typ) {
	case "counter", "c":
		return metrics.TypeCounter, nil
	case "gauge", "g":
		return metrics.TypeGauge, nil
	case "histogram", "h":
		return metrics.TypeHistogram, nil
	default:
		return "", fmt.Errorf("invalid metric type: %s", typ)
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	metricType, err := s.parseMetricType(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.registry.Register(req.Name, metricType); err != nil {
		if errors.Is(err, metrics.ErrMetricTypeMismatch) {
			http.Error(w, "metric type mismatch: already registered with different type", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("registered"))
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.qps.Increment()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req ReportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	labelsKey := metrics.LabelsKey(req.Labels)

	desc, exists := s.registry.Get(req.Name)
	var metricType metrics.MetricType

	if !exists {
		metricType = metrics.TypeGauge
		s.registry.Register(req.Name, metricType)
		desc, _ = s.registry.Get(req.Name)
	} else {
		metricType = desc.Type
	}

	if req.Type != "" {
		explicitType, err := s.parseMetricType(req.Type)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if explicitType != metricType {
			http.Error(w, "metric type mismatch: already registered with different type", http.StatusConflict)
			return
		}
	}

	if err := s.registry.AddLabelSeries(req.Name, labelsKey); err != nil {
		if errors.Is(err, metrics.ErrLabelSeriesLimit) {
			http.Error(w, "too many label combinations for this metric", http.StatusTooManyRequests)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch metricType {
	case metrics.TypeCounter:
		if err := s.store.IncrementCounter(req.Name, labelsKey, req.Value); err != nil {
			if errors.Is(err, metrics.ErrInvalidValue) {
				http.Error(w, "counter value must be non-negative", http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case metrics.TypeGauge:
		s.store.SetGauge(req.Name, labelsKey, req.Value)
	case metrics.TypeHistogram:
		s.store.ObserveHistogram(req.Name, labelsKey, req.Value)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("reported"))
}

type StatsResponse struct {
	MetricsCount int     `json:"metrics_count"`
	SeriesCount  int     `json:"series_count"`
	QPS          float64 `json:"qps_last_minute"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := StatsResponse{
		MetricsCount: s.registry.MetricCount(),
		SeriesCount:  s.registry.SeriesCount(),
		QPS:          s.qps.QPS(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

type HistogramBucketResponse struct {
	UpperBound float64 `json:"upper_bound"`
	Count      int     `json:"count"`
}

type HistogramResponse struct {
	Count   int                         `json:"count"`
	Sum     float64                     `json:"sum"`
	P50     float64                     `json:"p50"`
	P90     float64                     `json:"p90"`
	P99     float64                     `json:"p99"`
	Buckets []HistogramBucketResponse   `json:"buckets"`
}

type MetricSeriesResponse struct {
	Labels    map[string]string   `json:"labels"`
	Value     *float64            `json:"value,omitempty"`
	Counter   *float64            `json:"counter,omitempty"`
	Histogram *HistogramResponse  `json:"histogram,omitempty"`
}

type QueryResponse struct {
	Name   string                  `json:"name"`
	Type   string                  `json:"type"`
	Series []MetricSeriesResponse  `json:"series"`
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name parameter is required", http.StatusBadRequest)
		return
	}

	values, exists := metrics.GetMetricValues(s.registry, s.store, name)
	if !exists {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	if len(values) == 0 {
		desc, _ := s.registry.Get(name)
		resp := QueryResponse{
			Name:   name,
			Type:   string(desc.Type),
			Series: []MetricSeriesResponse{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := QueryResponse{
		Name:   name,
		Type:   string(values[0].Type),
		Series: make([]MetricSeriesResponse, 0, len(values)),
	}

	for _, v := range values {
		series := MetricSeriesResponse{
			Labels: v.Labels,
		}

		switch v.Type {
		case metrics.TypeCounter:
			val := v.Counter.Value
			series.Counter = &val
		case metrics.TypeGauge:
			val := v.Gauge.Value
			series.Value = &val
		case metrics.TypeHistogram:
			buckets := make([]HistogramBucketResponse, 0, len(v.Histogram.Buckets))
			for bound, count := range v.Histogram.Buckets {
				buckets = append(buckets, HistogramBucketResponse{
					UpperBound: bound,
					Count:      count,
				})
			}
			sort.Slice(buckets, func(i, j int) bool {
				return buckets[i].UpperBound < buckets[j].UpperBound
			})
			series.Histogram = &HistogramResponse{
				Count:   v.Histogram.Count,
				Sum:     v.Histogram.Sum,
				P50:     v.Histogram.P50,
				P90:     v.Histogram.P90,
				P99:     v.Histogram.P99,
				Buckets: buckets,
			}
		}

		resp.Series = append(resp.Series, series)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	output := metrics.ToPrometheusFormat(s.registry, s.store)
	w.Write([]byte(output))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("invalid PORT: %s", port)
	}

	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/register", server.handleRegister)
	mux.HandleFunc("/report", server.handleReport)
	mux.HandleFunc("/query", server.handleQuery)
	mux.HandleFunc("/stats", server.handleStats)
	mux.HandleFunc("/metrics", server.handleMetrics)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("metrics collector listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
