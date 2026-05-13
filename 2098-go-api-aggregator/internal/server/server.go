package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"api-aggregator/internal/aggregator"
	"api-aggregator/internal/config"
	"api-aggregator/internal/db"
	"api-aggregator/internal/validator"
)

type Server struct {
	mux *http.ServeMux
}

func New() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /aggregate", s.handleAggregate)
	s.mux.HandleFunc("POST /aggregate", s.handleAggregate)
	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.HandleFunc("GET /endpoints", s.handleListEndpoints)
	s.mux.HandleFunc("POST /endpoints", s.handleCreateEndpoint)
	s.mux.HandleFunc("GET /endpoints/{id}", s.handleGetEndpoint)
	s.mux.HandleFunc("PUT /endpoints/{id}", s.handleUpdateEndpoint)
	s.mux.HandleFunc("DELETE /endpoints/{id}", s.handleDeleteEndpoint)

	s.mux.HandleFunc("GET /configs", s.handleListConfigs)
	s.mux.HandleFunc("POST /configs", s.handleCreateConfig)
	s.mux.HandleFunc("GET /configs/{id}", s.handleGetConfig)
	s.mux.HandleFunc("POST /configs/{id}/aggregate", s.handleAggregateByConfig)

	s.mux.HandleFunc("POST /validate/{id}", s.handleValidate)
	s.mux.HandleFunc("POST /confirm/{id}", s.handleConfirm)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, map[string]interface{}{
		"service": "API Aggregator",
		"endpoints": map[string]interface{}{
			"GET /aggregate":          "Aggregate external APIs via query params",
			"POST /aggregate":         "Aggregate external APIs via JSON body",
			"GET /health":             "Health check",
			"REST /endpoints":         "Endpoint CRUD",
			"REST /configs":           "Aggregation config CRUD",
			"POST /configs/{id}/aggregate": "Aggregate via stored config",
			"POST /validate/{id}":     "Validate stored data",
			"POST /confirm/{id}":      "Confirm and trigger validation",
		},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) handleAggregate(w http.ResponseWriter, r *http.Request) {
	var spec *config.AggregationSpec
	var query url.Values

	if r.Method == "POST" {
		var bodySpec config.AggregationSpec
		if err := json.NewDecoder(r.Body).Decode(&bodySpec); err != nil && err != io.EOF {
			writeJSON(w, 400, map[string]string{"error": "invalid JSON body: " + err.Error()})
			return
		}
		spec = &bodySpec
		query = r.URL.Query()
	} else {
		var err error
		spec, err = config.ParseFromQuery(r.URL.Query())
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		query = r.URL.Query()
	}

	if err := config.ValidateSpec(spec); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	result, err := aggregator.Aggregate(spec, query)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.StatusCode)
	w.Write(result.Body)
}

func (s *Server) handleListEndpoints(w http.ResponseWriter, r *http.Request) {
	eps, err := db.ListEndpoints()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, eps)
}

func (s *Server) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var ep db.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	if err := validateEndpoint(&ep); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	if err := db.CreateEndpoint(&ep); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 201, ep)
}

func (s *Server) handleGetEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	ep, err := db.GetEndpoint(id)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if ep == nil {
		writeJSON(w, 404, map[string]string{"error": "endpoint not found"})
		return
	}

	writeJSON(w, 200, ep)
}

func (s *Server) handleUpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	existing, err := db.GetEndpoint(id)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if existing == nil {
		writeJSON(w, 404, map[string]string{"error": "endpoint not found"})
		return
	}

	var ep db.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	ep.ID = id
	if err := validateEndpoint(&ep); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	if err := db.UpdateEndpoint(&ep); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, ep)
}

func (s *Server) handleDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	if err := db.DeleteEndpoint(id); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 204, nil)
}

func (s *Server) handleListConfigs(w http.ResponseWriter, r *http.Request) {
	cfgs, err := db.ListAggregationConfigs()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, cfgs)
}

func (s *Server) handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg db.AggregationConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	if len(cfg.EndpointIDs) == 0 {
		writeJSON(w, 400, map[string]string{"error": "at least one endpoint id required"})
		return
	}

	if err := db.CreateAggregationConfig(&cfg); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 201, cfg)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	cfg, err := db.GetAggregationConfig(id)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if cfg == nil {
		writeJSON(w, 404, map[string]string{"error": "config not found"})
		return
	}

	writeJSON(w, 200, cfg)
}

func (s *Server) handleAggregateByConfig(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	result, err := aggregator.AggregateWithStoredConfig(id, r.URL.Query())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.StatusCode)
	w.Write(result.Body)
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	result, err := validator.ValidateAndProportionallyUpdate(id, r.URL.Query(), false)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, result)
}

func (s *Server) handleConfirm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	result, err := validator.ValidateAndProportionallyUpdate(id, r.URL.Query(), true)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, result)
}

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", s)
	}
	return id, nil
}

func validateEndpoint(ep *db.Endpoint) error {
	if ep.Name == "" {
		return fmt.Errorf("name is required")
	}

	u, err := url.Parse(ep.URL)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url must be http or https")
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url missing host")
	}

	if isInternal(host, u.Port()) {
		return fmt.Errorf("internal addresses not allowed (SSRF protection)")
	}

	if ep.Timeout < 1 {
		ep.Timeout = 1
	}

	if ep.Retries > 3 {
		ep.Retries = 3
	}
	if ep.Retries < 0 {
		ep.Retries = 0
	}

	if ep.Method == "" {
		ep.Method = "GET"
	}
	ep.Method = strings.ToUpper(ep.Method)

	return nil
}

func isInternal(host string, port string) bool {
	if port == "8080" {
		return true
	}

	lower := strings.ToLower(host)
	if lower == "localhost" || lower == "127.0.0.1" || lower == "::1" ||
		lower == "0.0.0.0" {
		return true
	}

	if strings.HasPrefix(lower, "127.") {
		return true
	}

	if strings.HasSuffix(lower, ".local") || strings.HasSuffix(lower, ".internal") {
		return true
	}

	if strings.HasPrefix(lower, "192.168.") ||
		strings.HasPrefix(lower, "10.") ||
		strings.HasPrefix(lower, "172.") {
		return true
	}

	return false
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
