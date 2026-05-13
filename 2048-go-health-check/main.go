package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/glebarez/sqlite"
)

type CheckType string

const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
	CheckTypeDB   CheckType = "db"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusSubHealth HealthStatus = "sub_health"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type TargetConfig struct {
	ID           string      `json:"id"`
	Type         CheckType   `json:"type"`
	Name         string      `json:"name"`
	Address      string      `json:"address"`
	Interval     int         `json:"interval,omitempty"`
	Timeout      int         `json:"timeout,omitempty"`
	ResponseTime int         `json:"response_time,omitempty"`
	Method       string      `json:"method,omitempty"`
	Headers      http.Header `json:"headers,omitempty"`
	ExpectedCode int         `json:"expected_code,omitempty"`
	DBType       string      `json:"db_type,omitempty"`
}

type CheckResult struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         CheckType    `json:"type"`
	Status       HealthStatus `json:"status"`
	ResponseTime int64        `json:"response_time_ms"`
	Error        string       `json:"error,omitempty"`
	Timestamp    time.Time    `json:"timestamp"`
}

type TargetState struct {
	Target         *TargetConfig
	CurrentStatus  HealthStatus
	CheckHistory   []CheckResult
	SuccessStreak  int
	FailureStreak  int
	LastCheck      time.Time
	mu             sync.RWMutex
}

type HealthCheckService struct {
	db             *sql.DB
	targets        map[string]*TargetState
	defaultInterval time.Duration
	defaultTimeout  time.Duration
	mu             sync.RWMutex
}

func NewHealthCheckService(dbPath string) (*HealthCheckService, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	service := &HealthCheckService{
		db:              db,
		targets:         make(map[string]*TargetState),
		defaultInterval: 30 * time.Second,
		defaultTimeout:  5 * time.Second,
	}

	if err := service.initDB(); err != nil {
		return nil, err
	}

	if err := service.loadTargets(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *HealthCheckService) initDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS targets (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		address TEXT NOT NULL,
		interval INTEGER,
		timeout INTEGER,
		response_time INTEGER,
		method TEXT,
		expected_code INTEGER,
		db_type TEXT,
		headers TEXT
	);

	CREATE TABLE IF NOT EXISTS check_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id TEXT NOT NULL,
		status TEXT NOT NULL,
		response_time INTEGER NOT NULL,
		error TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (target_id) REFERENCES targets(id)
	);

	CREATE INDEX IF NOT EXISTS idx_check_results_target_id ON check_results(target_id);
	CREATE INDEX IF NOT EXISTS idx_check_results_timestamp ON check_results(timestamp);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *HealthCheckService) loadTargets() error {
	rows, err := s.db.Query(`SELECT id, type, name, address, interval, timeout, response_time, method, expected_code, db_type, headers FROM targets`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var config TargetConfig
		var headersStr sql.NullString
		err := rows.Scan(
			&config.ID, &config.Type, &config.Name, &config.Address,
			&config.Interval, &config.Timeout, &config.ResponseTime,
			&config.Method, &config.ExpectedCode, &config.DBType, &headersStr,
		)
		if err != nil {
			return err
		}

		if headersStr.Valid {
			var headers map[string][]string
			if err := json.Unmarshal([]byte(headersStr.String), &headers); err == nil {
				config.Headers = headers
			}
		}

		state := &TargetState{
			Target:        &config,
			CurrentStatus: StatusHealthy,
			CheckHistory:  make([]CheckResult, 0),
		}

		s.targets[config.ID] = state
	}

	return nil
}

func (s *HealthCheckService) validateConfig(config *TargetConfig) error {
	if config.ID == "" {
		return fmt.Errorf("id is required")
	}

	if config.Name == "" {
		return fmt.Errorf("name is required")
	}

	switch config.Type {
	case CheckTypeHTTP:
		if config.Address == "" {
			return fmt.Errorf("address is required for http check")
		}
		if _, err := url.Parse(config.Address); err != nil {
			return fmt.Errorf("invalid address: %v", err)
		}
		if config.Method == "" {
			config.Method = "GET"
		}
		if config.ExpectedCode == 0 {
			config.ExpectedCode = 200
		}

	case CheckTypeTCP:
		if config.Address == "" {
			return fmt.Errorf("address is required for tcp check")
		}
		host, portStr, err := net.SplitHostPort(config.Address)
		if err != nil {
			return fmt.Errorf("invalid address format, expected host:port: %v", err)
		}
		if host == "" {
			return fmt.Errorf("host is required in address")
		}
		if port, err := strconv.Atoi(portStr); err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("invalid port number")
		}

	case CheckTypeDB:
		if config.DBType == "" {
			return fmt.Errorf("db_type is required for database check")
		}
		if config.Address == "" {
			return fmt.Errorf("address is required for database check")
		}

	default:
		return fmt.Errorf("invalid check type: %s (must be http, tcp, or db)", config.Type)
	}

	if config.Interval < 0 {
		return fmt.Errorf("interval cannot be negative")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if config.ResponseTime < 0 {
		return fmt.Errorf("response_time cannot be negative")
	}

	return nil
}

func (s *HealthCheckService) AddTarget(config *TargetConfig) error {
	if err := s.validateConfig(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.targets[config.ID]; exists {
		return fmt.Errorf("target with id %s already exists", config.ID)
	}

	var headersJSON []byte
	if config.Headers != nil {
		var err error
		headersJSON, err = json.Marshal(config.Headers)
		if err != nil {
			return err
		}
	}

	_, err := s.db.Exec(`
		INSERT INTO targets (id, type, name, address, interval, timeout, response_time, method, expected_code, db_type, headers)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, config.ID, config.Type, config.Name, config.Address, config.Interval, config.Timeout,
		config.ResponseTime, config.Method, config.ExpectedCode, config.DBType, string(headersJSON))
	if err != nil {
		return err
	}

	state := &TargetState{
		Target:        config,
		CurrentStatus: StatusHealthy,
		CheckHistory:  make([]CheckResult, 0),
	}

	s.targets[config.ID] = state
	go s.startCheckLoop(state)
	return nil
}

func (s *HealthCheckService) UpdateTarget(config *TargetConfig) error {
	if err := s.validateConfig(config); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.targets[config.ID]; !exists {
		return fmt.Errorf("target with id %s not found", config.ID)
	}

	var headersJSON []byte
	if config.Headers != nil {
		var err error
		headersJSON, err = json.Marshal(config.Headers)
		if err != nil {
			return err
		}
	}

	_, err := s.db.Exec(`
		UPDATE targets SET type=?, name=?, address=?, interval=?, timeout=?, response_time=?, 
		method=?, expected_code=?, db_type=?, headers=? WHERE id=?
	`, config.Type, config.Name, config.Address, config.Interval, config.Timeout,
		config.ResponseTime, config.Method, config.ExpectedCode, config.DBType, string(headersJSON), config.ID)
	if err != nil {
		return err
	}

	s.targets[config.ID].Target = config
	return nil
}

func (s *HealthCheckService) DeleteTarget(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.targets[id]; !exists {
		return fmt.Errorf("target with id %s not found", id)
	}

	_, err := s.db.Exec("DELETE FROM check_results WHERE target_id=?", id)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM targets WHERE id=?", id)
	if err != nil {
		return err
	}

	delete(s.targets, id)
	return nil
}

func (s *HealthCheckService) GetTarget(id string) (*TargetConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.targets[id]
	if !exists {
		return nil, fmt.Errorf("target with id %s not found", id)
	}

	return state.Target, nil
}

func (s *HealthCheckService) GetAllTargets() []*TargetConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	targets := make([]*TargetConfig, 0, len(s.targets))
	for _, state := range s.targets {
		targets = append(targets, state.Target)
	}

	return targets
}

func (s *HealthCheckService) checkHTTP(ctx context.Context, target *TargetConfig) (HealthStatus, int64, string) {
	timeout := s.defaultTimeout
	if target.Timeout > 0 {
		timeout = time.Duration(target.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, target.Method, target.Address, nil)
	if err != nil {
		return StatusUnhealthy, 0, fmt.Sprintf("failed to create request: %v", err)
	}

	if target.Headers != nil {
		for key, values := range target.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	client := &http.Client{}
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		if ctx.Err() != nil {
			return StatusUnhealthy, elapsed.Milliseconds(), "request timeout"
		}
		return StatusUnhealthy, elapsed.Milliseconds(), fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close()

	responseTime := 2000
	if target.ResponseTime > 0 {
		responseTime = target.ResponseTime
	}

	if resp.StatusCode != target.ExpectedCode {
		return StatusUnhealthy, elapsed.Milliseconds(), fmt.Sprintf("unexpected status code: %d, expected: %d", resp.StatusCode, target.ExpectedCode)
	}

	if elapsed > time.Duration(responseTime)*time.Millisecond {
		return StatusSubHealth, elapsed.Milliseconds(), ""
	}

	return StatusHealthy, elapsed.Milliseconds(), ""
}

func (s *HealthCheckService) checkTCP(ctx context.Context, target *TargetConfig) (HealthStatus, int64, string) {
	timeout := s.defaultTimeout
	if target.Timeout > 0 {
		timeout = time.Duration(target.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := &net.Dialer{}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", target.Address)
	elapsed := time.Since(start)

	if err != nil {
		if ctx.Err() != nil {
			return StatusUnhealthy, elapsed.Milliseconds(), "connection timeout"
		}
		return StatusUnhealthy, elapsed.Milliseconds(), fmt.Sprintf("connection failed: %v", err)
	}
	defer conn.Close()

	responseTime := 2000
	if target.ResponseTime > 0 {
		responseTime = target.ResponseTime
	}

	if elapsed > time.Duration(responseTime)*time.Millisecond {
		return StatusSubHealth, elapsed.Milliseconds(), ""
	}

	return StatusHealthy, elapsed.Milliseconds(), ""
}

func (s *HealthCheckService) checkDB(ctx context.Context, target *TargetConfig) (HealthStatus, int64, string) {
	timeout := s.defaultTimeout
	if target.Timeout > 0 {
		timeout = time.Duration(target.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var driverName string
	switch strings.ToLower(target.DBType) {
	case "mysql":
		driverName = "mysql"
	case "postgres", "postgresql":
		driverName = "postgres"
	case "sqlite", "sqlite3":
		driverName = "sqlite"
	default:
		return StatusUnhealthy, 0, fmt.Sprintf("unsupported database type: %s", target.DBType)
	}

	start := time.Now()
	db, err := sql.Open(driverName, target.Address)
	if err != nil {
		return StatusUnhealthy, time.Since(start).Milliseconds(), fmt.Sprintf("failed to open connection: %v", err)
	}
	defer db.Close()

	err = db.PingContext(ctx)
	elapsed := time.Since(start)

	if err != nil {
		if ctx.Err() != nil {
			return StatusUnhealthy, elapsed.Milliseconds(), "database connection timeout"
		}
		return StatusUnhealthy, elapsed.Milliseconds(), fmt.Sprintf("database ping failed: %v", err)
	}

	responseTime := 2000
	if target.ResponseTime > 0 {
		responseTime = target.ResponseTime
	}

	if elapsed > time.Duration(responseTime)*time.Millisecond {
		return StatusSubHealth, elapsed.Milliseconds(), ""
	}

	return StatusHealthy, elapsed.Milliseconds(), ""
}

func (s *HealthCheckService) performCheck(ctx context.Context, target *TargetConfig) (HealthStatus, int64, string) {
	switch target.Type {
	case CheckTypeHTTP:
		return s.checkHTTP(ctx, target)
	case CheckTypeTCP:
		return s.checkTCP(ctx, target)
	case CheckTypeDB:
		return s.checkDB(ctx, target)
	default:
		return StatusUnhealthy, 0, fmt.Sprintf("unknown check type: %s", target.Type)
	}
}

func (s *HealthCheckService) updateState(state *TargetState, status HealthStatus, responseTime int64, errMsg string) {
	state.mu.Lock()
	defer state.mu.Unlock()

	result := CheckResult{
		ID:           state.Target.ID,
		Name:         state.Target.Name,
		Type:         state.Target.Type,
		Status:       status,
		ResponseTime: responseTime,
		Error:        errMsg,
		Timestamp:    time.Now(),
	}

	state.CheckHistory = append(state.CheckHistory, result)
	state.LastCheck = result.Timestamp

	if len(state.CheckHistory) > 100 {
		state.CheckHistory = state.CheckHistory[len(state.CheckHistory)-100:]
	}

	if status == StatusHealthy || status == StatusSubHealth {
		state.SuccessStreak++
		state.FailureStreak = 0

		if state.CurrentStatus == StatusUnhealthy && state.SuccessStreak >= 3 {
			state.CurrentStatus = status
		} else if state.CurrentStatus != StatusUnhealthy {
			state.CurrentStatus = status
		}
	} else {
		state.FailureStreak++
		state.SuccessStreak = 0

		if state.FailureStreak >= 3 {
			state.CurrentStatus = StatusUnhealthy
		}
	}

	s.saveCheckResult(&result)
}

func (s *HealthCheckService) saveCheckResult(result *CheckResult) {
	_, err := s.db.Exec(`
		INSERT INTO check_results (target_id, status, response_time, error, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, result.ID, result.Status, result.ResponseTime, result.Error, result.Timestamp.Format("2006-01-02 15:04:05"))
	if err != nil {
		fmt.Printf("Failed to save check result: %v\n", err)
	}
}

func (s *HealthCheckService) cleanupOldResults() {
	cutoff := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	_, err := s.db.Exec("DELETE FROM check_results WHERE timestamp < ?", cutoff)
	if err != nil {
		fmt.Printf("Failed to cleanup old results: %v\n", err)
	}
}

func (s *HealthCheckService) Start() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			s.cleanupOldResults()
		}
	}()

	s.mu.RLock()
	targets := make([]*TargetState, 0, len(s.targets))
	for _, state := range s.targets {
		targets = append(targets, state)
	}
	s.mu.RUnlock()

	for _, state := range targets {
		go s.startCheckLoop(state)
	}
}

func (s *HealthCheckService) startCheckLoop(state *TargetState) {
	interval := s.defaultInterval
	if state.Target.Interval > 0 {
		interval = time.Duration(state.Target.Interval) * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := context.Background()

	for range ticker.C {
		status, responseTime, errMsg := s.performCheck(ctx, state.Target)
		s.updateState(state, status, responseTime, errMsg)
	}
}

func (s *HealthCheckService) GetStatus(id string) (*CheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.targets[id]
	if !exists {
		return nil, fmt.Errorf("target with id %s not found", id)
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	if len(state.CheckHistory) == 0 {
		return nil, fmt.Errorf("no check results yet")
	}

	latest := state.CheckHistory[len(state.CheckHistory)-1]
	latest.Status = state.CurrentStatus
	return &latest, nil
}

func (s *HealthCheckService) GetAllStatuses() []CheckResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]CheckResult, 0, len(s.targets))
	for _, state := range s.targets {
		state.mu.RLock()
		if len(state.CheckHistory) > 0 {
			latest := state.CheckHistory[len(state.CheckHistory)-1]
			latest.Status = state.CurrentStatus
			results = append(results, latest)
		} else {
			results = append(results, CheckResult{
				ID:           state.Target.ID,
				Name:         state.Target.Name,
				Type:         state.Target.Type,
				Status:       state.CurrentStatus,
				ResponseTime: 0,
				Timestamp:    state.LastCheck,
			})
		}
		state.mu.RUnlock()
	}

	return results
}

func (s *HealthCheckService) GetHistory(id string, limit int) ([]CheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	target, exists := s.targets[id]
	if !exists {
		return nil, fmt.Errorf("target with id %s not found", id)
	}

	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`
		SELECT target_id, status, response_time, error, timestamp 
		FROM check_results 
		WHERE target_id = ? 
		ORDER BY timestamp DESC 
		LIMIT ?
	`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]CheckResult, 0)
	for rows.Next() {
		var result CheckResult
		var timestampStr string
		err := rows.Scan(&result.ID, &result.Status, &result.ResponseTime, &result.Error, &timestampStr)
		if err != nil {
			return nil, err
		}

		result.Name = target.Target.Name
		result.Type = target.Target.Type
		result.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestampStr)
		results = append(results, result)
	}

	return results, nil
}

type HourlyStats struct {
	Hour          string  `json:"hour"`
	TotalChecks   int     `json:"total_checks"`
	SuccessChecks int     `json:"success_checks"`
	Availability  float64 `json:"availability"`
}

func (s *HealthCheckService) GetAvailabilityStats(id string, hours int) ([]HourlyStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.targets[id]; !exists {
		return nil, fmt.Errorf("target with id %s not found", id)
	}

	if hours <= 0 || hours > 24 {
		hours = 24
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	rows, err := s.db.Query(`
		SELECT 
			strftime('%Y-%m-%d %H:00:00', timestamp) as hour,
			COUNT(*) as total,
			SUM(CASE WHEN status IN ('healthy', 'sub_health') THEN 1 ELSE 0 END) as success
		FROM check_results
		WHERE target_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY strftime('%Y-%m-%d %H:00:00', timestamp)
		ORDER BY hour
	`, id, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsMap := make(map[string]HourlyStats)
	for rows.Next() {
		var hour string
		var total, success int
		err := rows.Scan(&hour, &total, &success)
		if err != nil {
			return nil, err
		}

		availability := 0.0
		if total > 0 {
			availability = float64(success) / float64(total) * 100
		}

		statsMap[hour] = HourlyStats{
			Hour:          hour,
			TotalChecks:   total,
			SuccessChecks: success,
			Availability:  availability,
		}
	}

	stats := make([]HourlyStats, 0)
	for i := hours - 1; i >= 0; i-- {
		hourTime := endTime.Add(-time.Duration(i) * time.Hour).Truncate(time.Hour)
		hourStr := hourTime.Format("2006-01-02 15:00:00")

		if stat, exists := statsMap[hourStr]; exists {
			stats = append(stats, stat)
		} else {
			stats = append(stats, HourlyStats{
				Hour:          hourStr,
				TotalChecks:   0,
				SuccessChecks: 0,
				Availability:  0.0,
			})
		}
	}

	return stats, nil
}

type APIHandler struct {
	service *HealthCheckService
}

func NewAPIHandler(service *HealthCheckService) *APIHandler {
	return &APIHandler{service: service}
}

func (h *APIHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *APIHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *APIHandler) HandleTargets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		targets := h.service.GetAllTargets()
		if targets == nil {
			targets = make([]*TargetConfig, 0)
		}
		h.writeJSON(w, http.StatusOK, targets)

	case http.MethodPost:
		var config TargetConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if err := h.service.AddTarget(&config); err != nil {
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		h.writeJSON(w, http.StatusCreated, config)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *APIHandler) HandleTarget(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/targets/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "target id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		target, err := h.service.GetTarget(id)
		if err != nil {
			h.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		h.writeJSON(w, http.StatusOK, target)

	case http.MethodPut:
		var config TargetConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		config.ID = id

		if err := h.service.UpdateTarget(&config); err != nil {
			if err.Error() == fmt.Sprintf("target with id %s not found", id) {
				h.writeError(w, http.StatusNotFound, err.Error())
			} else {
				h.writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		h.writeJSON(w, http.StatusOK, config)

	case http.MethodDelete:
		if err := h.service.DeleteTarget(id); err != nil {
			h.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		h.writeJSON(w, http.StatusNoContent, nil)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *APIHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		statuses := h.service.GetAllStatuses()
		if statuses == nil {
			statuses = make([]CheckResult, 0)
		}
		h.writeJSON(w, http.StatusOK, statuses)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *APIHandler) HandleTargetStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/status/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "target id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		status, err := h.service.GetStatus(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				h.writeError(w, http.StatusNotFound, err.Error())
			} else {
				h.writeError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		h.writeJSON(w, http.StatusOK, status)

	default:
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *APIHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/history/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "target id is required")
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.service.GetHistory(id, limit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if history == nil {
		history = make([]CheckResult, 0)
	}

	h.writeJSON(w, http.StatusOK, history)
}

func (h *APIHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/stats/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "target id is required")
		return
	}

	hours := 24
	if hoursStr := r.URL.Query().Get("hours"); hoursStr != "" {
		if parsed, err := strconv.Atoi(hoursStr); err == nil && parsed > 0 && parsed <= 24 {
			hours = parsed
		}
	}

	stats, err := h.service.GetAvailabilityStats(id, hours)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, http.StatusNotFound, err.Error())
		} else {
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if stats == nil {
		stats = make([]HourlyStats, 0)
	}

	h.writeJSON(w, http.StatusOK, stats)
}

func main() {
	service, err := NewHealthCheckService("./health_check.db")
	if err != nil {
		fmt.Printf("Failed to create service: %v\n", err)
		return
	}

	handler := NewAPIHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/targets", handler.HandleTargets)
	mux.HandleFunc("/api/targets/", handler.HandleTarget)
	mux.HandleFunc("/api/status", handler.HandleStatus)
	mux.HandleFunc("/api/status/", handler.HandleTargetStatus)
	mux.HandleFunc("/api/history/", handler.HandleHistory)
	mux.HandleFunc("/api/stats/", handler.HandleStats)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			handler.writeError(w, http.StatusNotFound, "endpoint not found")
			return
		}
		w.Write([]byte("Health Check Service API"))
	})

	service.Start()

	fmt.Println("Health Check Service starting on port 9000...")
	if err := http.ListenAndServe(":8200", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
