package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"dns-resolver/pkg/database"
	"dns-resolver/pkg/dns"
	"dns-resolver/pkg/healthcheck"
	"dns-resolver/pkg/whois"
)

const (
	MaxBatchDomains = 50
	DBPath          = "./dns_resolver.db"
)

type Server struct {
	resolver      *dns.Resolver
	db            *database.DB
	healthChecker *healthcheck.HealthChecker
	whoisService  *whois.WHOISService
}

type SingleQueryRequest struct {
	Domain     string `json:"domain"`
	RecordType string `json:"record_type"`
	DNSServer  string `json:"dns_server,omitempty"`
}

type BatchQueryRequest struct {
	Queries []SingleQueryRequest `json:"queries"`
}

type BatchQueryResponse struct {
	Results []*dns.DNSQueryResult `json:"results"`
	Errors  []QueryError          `json:"errors,omitempty"`
}

type QueryError struct {
	Domain     string `json:"domain"`
	RecordType string `json:"record_type"`
	Error      string `json:"error"`
}

type ReverseQueryRequest struct {
	IP        string `json:"ip"`
	DNSServer string `json:"dns_server,omitempty"`
}

type HealthCheckRequest struct {
	Domain          string `json:"domain"`
	IntervalMinutes int    `json:"interval_minutes"`
	Enabled         bool   `json:"enabled"`
}

type ErrorResponse struct {
	Error      string   `json:"error"`
	StatusCode int      `json:"status_code"`
	Supported  []string `json:"supported_types,omitempty"`
}

func main() {
	resolver := dns.NewResolver()

	db, err := database.NewDB(DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	hc, err := healthcheck.NewHealthChecker(DBPath, &DNSQuerierAdapter{resolver: resolver})
	if err != nil {
		log.Fatalf("Failed to initialize health checker: %v", err)
	}

	whoisSvc, err := whois.NewWHOISService(DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize WHOIS service: %v", err)
	}

	server := &Server{
		resolver:      resolver,
		db:            db,
		healthChecker: hc,
		whoisService:  whoisSvc,
	}

	go func() {
		for notification := range hc.Notifications() {
			log.Printf("HEALTH NOTIFICATION: %s", notification.Message)
		}
	}()

	http.HandleFunc("/api/dns/query", server.handleDNSQuery)
	http.HandleFunc("/api/dns/batch", server.handleBatchQuery)
	http.HandleFunc("/api/dns/reverse", server.handleReverseQuery)
	http.HandleFunc("/api/dns/history", server.handleQueryHistory)
	http.HandleFunc("/api/healthcheck", server.handleHealthCheck)
	http.HandleFunc("/api/healthcheck/status", server.handleHealthCheckStatus)
	http.HandleFunc("/api/whois", server.handleWHOISQuery)

	log.Printf("DNS Resolver Server starting on port 9502...")
	if err := http.ListenAndServe(":9502", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type DNSQuerierAdapter struct {
	resolver *dns.Resolver
}

func (d *DNSQuerierAdapter) QueryARecords(ctx context.Context, domain string) ([]string, error) {
	result, err := d.resolver.Query(ctx, domain, "A", "")
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, record := range result.Records {
		ips = append(ips, record.RecordValue)
	}
	return ips, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string, supportedTypes []string) {
	response := ErrorResponse{
		Error:      message,
		StatusCode: status,
		Supported:  supportedTypes,
	}
	writeJSON(w, status, response)
}

func (s *Server) handleDNSQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req SingleQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.Domain == "" {
		writeError(w, http.StatusBadRequest, "Domain is required", nil)
		return
	}

	if req.RecordType == "" {
		writeError(w, http.StatusBadRequest, "Record type is required", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := s.resolver.Query(ctx, req.Domain, req.RecordType, req.DNSServer)
	if err != nil {
		s.handleDNSQueryError(w, err)
		return
	}

	if !result.FromCache {
		s.db.SaveQueryHistory(req.Domain, req.RecordType, req.DNSServer, result, result.QueryTime)
	}

	syncData := s.syncStatus(result)
	if syncData != nil {
		result = syncData
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBatchQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req BatchQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if len(req.Queries) == 0 {
		writeError(w, http.StatusBadRequest, "No queries provided", nil)
		return
	}

	if len(req.Queries) > MaxBatchDomains {
		writeError(w, http.StatusBadRequest, "Batch limit exceeded. Maximum 50 queries allowed", nil)
		return
	}

	response := BatchQueryResponse{
		Results: make([]*dns.DNSQueryResult, 0, len(req.Queries)),
		Errors:  make([]QueryError, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, q := range req.Queries {
		wg.Add(1)
		go func(query SingleQueryRequest) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			result, err := s.resolver.Query(ctx, query.Domain, query.RecordType, query.DNSServer)
			if err != nil {
				mu.Lock()
				response.Errors = append(response.Errors, QueryError{
					Domain:     query.Domain,
					RecordType: query.RecordType,
					Error:      err.Error(),
				})
				mu.Unlock()
				return
			}

			if !result.FromCache {
				s.db.SaveQueryHistory(query.Domain, query.RecordType, query.DNSServer, result, result.QueryTime)
			}

			mu.Lock()
			response.Results = append(response.Results, result)
			mu.Unlock()
		}(q)
	}

	wg.Wait()

	for i, result := range response.Results {
		if synced := s.syncStatus(result); synced != nil {
			response.Results[i] = synced
		}
	}

	s.reallocateResources(&response)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleReverseQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req ReverseQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.IP == "" {
		writeError(w, http.StatusBadRequest, "IP address is required", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := s.resolver.ReverseQuery(ctx, req.IP, req.DNSServer)
	if err != nil {
		s.handleDNSQueryError(w, err)
		return
	}

	if !result.FromCache {
		s.db.SaveQueryHistory(req.IP, "PTR", req.DNSServer, result, result.QueryTime)
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleQueryHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	limit := 100
	history, err := s.db.GetQueryHistory(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve query history", nil)
		return
	}

	writeJSON(w, http.StatusOK, history)
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleAddHealthCheck(w, r)
	case http.MethodDelete:
		s.handleRemoveHealthCheck(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

func (s *Server) handleAddHealthCheck(w http.ResponseWriter, r *http.Request) {
	var req HealthCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.Domain == "" {
		writeError(w, http.StatusBadRequest, "Domain is required", nil)
		return
	}

	if !dns.IsValidDomain(req.Domain) {
		writeError(w, http.StatusBadRequest, "Invalid domain format", nil)
		return
	}

	config := healthcheck.HealthCheckConfig{
		Domain:          req.Domain,
		IntervalMinutes: req.IntervalMinutes,
		Enabled:         req.Enabled,
	}

	if err := s.healthChecker.AddCheck(config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"status":  "success",
		"message": "Health check added for " + req.Domain,
	})
}

func (s *Server) handleRemoveHealthCheck(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeError(w, http.StatusBadRequest, "Domain parameter is required", nil)
		return
	}

	if err := s.healthChecker.RemoveCheck(domain); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Health check removed for " + domain,
	})
}

func (s *Server) handleHealthCheckStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain != "" {
		status, err := s.healthChecker.GetStatus(domain)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		if status == nil {
			writeError(w, http.StatusNotFound, "Health check not found for domain", nil)
			return
		}
		writeJSON(w, http.StatusOK, status)
		return
	}

	statuses, err := s.healthChecker.GetAllStatuses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, statuses)
}

func (s *Server) handleWHOISQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.Domain == "" {
		writeError(w, http.StatusBadRequest, "Domain is required", nil)
		return
	}

	if !dns.IsValidDomain(req.Domain) {
		writeError(w, http.StatusBadRequest, "Invalid domain format", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.whoisService.Query(ctx, req.Domain)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "WHOIS query failed: "+err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDNSQueryError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *dns.ValidationError:
		writeError(w, http.StatusBadRequest, e.Error(), nil)
	case *dns.ServerError:
		writeError(w, http.StatusBadGateway, e.Error(), nil)
	case *dns.RecordTypeError:
		writeError(w, http.StatusBadRequest, e.Error(), e.Supported)
	default:
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
	}
}

func (s *Server) syncStatus(result *dns.DNSQueryResult) *dns.DNSQueryResult {
	if result == nil || len(result.Records) == 0 {
		return nil
	}

	syncedResult := *result
	syncedResult.Records = make([]dns.RecordResult, len(result.Records))
	copy(syncedResult.Records, result.Records)

	return &syncedResult
}

func (s *Server) reallocateResources(response *BatchQueryResponse) {
	if response == nil || len(response.Results) == 0 {
		return
	}

	totalRecords := 0
	for _, result := range response.Results {
		totalRecords += len(result.Records)
	}

	if totalRecords == 0 {
		return
	}

	for _, result := range response.Results {
		_ = result
	}
}
