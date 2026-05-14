package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	Port              = ":8101"
	BatchSize         = 1000
	SQLiteDBPath      = "data_migrate.db"
	StatusPending     = "pending"
	StatusRunning     = "running"
	StatusPaused      = "paused"
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
)

type ConnectionConfig struct {
	Type     string            `json:"type"`
	DSN      string            `json:"dsn"`
	TimeZone string            `json:"timezone,omitempty"`
}

type FieldMapping struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type TableMapping struct {
	SourceTable      string            `json:"source_table"`
	TargetTable      string            `json:"target_table"`
	PrimaryKey       string            `json:"primary_key"`
	TimestampField   string            `json:"timestamp_field,omitempty"`
	Fields           []FieldMapping    `json:"fields"`
}

type MigrationConfig struct {
	Name        string            `json:"name"`
	Source      ConnectionConfig  `json:"source"`
	Target      ConnectionConfig  `json:"target"`
	Tables      []TableMapping    `json:"tables"`
}

type MigrationProgress struct {
	TaskID          string    `json:"task_id"`
	ConfigID        string    `json:"config_id"`
	CurrentTable    string    `json:"current_table"`
	LastPrimaryKey  string    `json:"last_primary_key"`
	LastTimestamp   string    `json:"last_timestamp"`
	RecordsMigrated int64    `json:"records_migrated"`
	RecordsFailed   int64    `json:"records_failed"`
	TotalRecords    int64    `json:"total_records"`
}

type MigrationTask struct {
	ID              string               `json:"id"`
	ConfigID        string               `json:"config_id"`
	Status          string               `json:"status"`
	Progress        *MigrationProgress   `json:"progress"`
	StartTime       *time.Time           `json:"start_time,omitempty"`
	EndTime         *time.Time           `json:"end_time,omitempty"`
	ErrorMessage    string               `json:"error_message,omitempty"`
	mutex           sync.Mutex
	pauseRequested  bool
	stopRequested   bool
}

type ErrorRecord struct {
	ID        int64     `json:"id"`
	TaskID    string    `json:"task_id"`
	TableName string    `json:"table_name"`
	RecordKey string    `json:"record_key"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	db             *sql.DB
	configs        map[string]*MigrationConfig
	tasks          map[string]*MigrationTask
	globalMutex    sync.RWMutex
)

func main() {
	var err error
	db, err = sql.Open("sqlite", SQLiteDBPath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping SQLite database: %v", err)
	}

	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	configs = make(map[string]*MigrationConfig)
	tasks = make(map[string]*MigrationTask)

	http.HandleFunc("/configs", handleConfigs)
	http.HandleFunc("/configs/", handleConfigByID)
	http.HandleFunc("/tasks", handleTasks)
	http.HandleFunc("/tasks/", handleTaskByID)
	http.HandleFunc("/errors/", handleErrorsByTask)

	fmt.Printf("Data Migration Server running on port %s\n", Port)
	if err := http.ListenAndServe(Port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initDatabase() error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS configs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			config_json TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			config_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			progress_json TEXT,
			start_time TIMESTAMP,
			end_time TIMESTAMP,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(config_id) REFERENCES configs(id)
		)`,
		`CREATE TABLE IF NOT EXISTS error_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			table_name TEXT NOT NULL,
			record_key TEXT,
			reason TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(task_id) REFERENCES tasks(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_config_id ON tasks(config_id)`,
		`CREATE INDEX IF NOT EXISTS idx_errors_task_id ON error_records(task_id)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute schema: %v", err)
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func validateConfig(config *MigrationConfig) error {
	if config.Name == "" {
		return fmt.Errorf("config name is required")
	}
	if config.Source.Type == "" {
		return fmt.Errorf("source database type is required")
	}
	if config.Source.DSN == "" {
		return fmt.Errorf("source DSN is required")
	}
	if config.Target.Type == "" {
		return fmt.Errorf("target database type is required")
	}
	if config.Target.DSN == "" {
		return fmt.Errorf("target DSN is required")
	}
	if len(config.Tables) == 0 {
		return fmt.Errorf("at least one table mapping is required")
	}
	
	for _, table := range config.Tables {
		if table.SourceTable == "" {
			return fmt.Errorf("source table name is required")
		}
		if table.TargetTable == "" {
			return fmt.Errorf("target table name is required")
		}
		if table.PrimaryKey == "" {
			return fmt.Errorf("primary key is required for table %s", table.SourceTable)
		}
		if len(table.Fields) == 0 {
			return fmt.Errorf("at least one field mapping is required for table %s", table.SourceTable)
		}
	}
	
	return nil
}

func handleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listConfigs(w, r)
	case http.MethodPost:
		createConfig(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleConfigByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/configs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Config ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		getConfig(w, r, id)
	case http.MethodPut:
		updateConfig(w, r, id)
	case http.MethodDelete:
		deleteConfig(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func listConfigs(w http.ResponseWriter, r *http.Request) {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	rows, err := db.Query("SELECT id, name, config_json FROM configs ORDER BY created_at DESC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type ConfigListItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Config *MigrationConfig `json:"config"`
	}
	
	configList := make([]ConfigListItem, 0)
	for rows.Next() {
		var id, name, configJSON string
		if err := rows.Scan(&id, &name, &configJSON); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		var config MigrationConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			continue
		}
		
		configList = append(configList, ConfigListItem{
			ID:     id,
			Name:   name,
			Config: &config,
		})
	}

	writeJSON(w, http.StatusOK, configList)
}

func createConfig(w http.ResponseWriter, r *http.Request) {
	var config MigrationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if err := validateConfig(&config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := connectToDB(&config.Source); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to connect to source database: "+err.Error())
		return
	}
	if _, err := connectToDB(&config.Target); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to connect to target database: "+err.Error())
		return
	}

	globalMutex.Lock()
	defer globalMutex.Unlock()

	id := generateID()
	configJSON, _ := json.Marshal(config)

	_, err := db.Exec(
		"INSERT INTO configs (id, name, config_json) VALUES (?, ?, ?)",
		id, config.Name, string(configJSON),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	configs[id] = &config
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":     id,
		"config": config,
	})
}

func getConfig(w http.ResponseWriter, r *http.Request, id string) {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	var configJSON string
	err := db.QueryRow("SELECT config_json FROM configs WHERE id = ?", id).Scan(&configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var config MigrationConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, config)
}

func updateConfig(w http.ResponseWriter, r *http.Request, id string) {
	var config MigrationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if err := validateConfig(&config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	globalMutex.Lock()
	defer globalMutex.Unlock()

	var existingID string
	err := db.QueryRow("SELECT id FROM configs WHERE id = ?", id).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	configJSON, _ := json.Marshal(config)
	_, err = db.Exec(
		"UPDATE configs SET name = ?, config_json = ? WHERE id = ?",
		config.Name, string(configJSON), id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	configs[id] = &config
	writeJSON(w, http.StatusOK, config)
}

func deleteConfig(w http.ResponseWriter, r *http.Request, id string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	_, err := db.Exec("DELETE FROM configs WHERE id = ?", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	delete(configs, id)
	w.WriteHeader(http.StatusNoContent)
}

func connectToDB(cfg *ConnectionConfig) (*sql.DB, error) {
	var driverName string
	switch strings.ToLower(cfg.Type) {
	case "sqlite", "sqlite3":
		driverName = "sqlite"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	conn, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}

func getColumns(conn *sql.DB, tableName string) ([]string, error) {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]string, 0)
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt_value interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}

	return columns, nil
}

func convertValue(srcValue interface{}, targetType string) (interface{}, error) {
	if srcValue == nil {
		return nil, nil
	}

	v := reflect.ValueOf(srcValue)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch strings.ToLower(targetType) {
	case "text", "varchar", "char":
		switch v.Kind() {
		case reflect.String:
			return v.String(), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(v.Int(), 10), nil
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
		case reflect.Bool:
			return strconv.FormatBool(v.Bool()), nil
		default:
			if t, ok := srcValue.(time.Time); ok {
				return t.Format(time.RFC3339), nil
			}
			return fmt.Sprintf("%v", srcValue), nil
		}

	case "integer", "int", "bigint":
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return v.Int(), nil
		case reflect.Float32, reflect.Float64:
			return int64(v.Float()), nil
		case reflect.String:
			i, err := strconv.ParseInt(v.String(), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert string to int: %v", err)
			}
			return i, nil
		case reflect.Bool:
			if v.Bool() {
				return int64(1), nil
			}
			return int64(0), nil
		}

	case "real", "float", "double":
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(v.Int()), nil
		case reflect.Float32, reflect.Float64:
			return v.Float(), nil
		case reflect.String:
			f, err := strconv.ParseFloat(v.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert string to float: %v", err)
			}
			return f, nil
		}

	case "boolean", "bool":
		switch v.Kind() {
		case reflect.Bool:
			return v.Bool(), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return v.Int() != 0, nil
		case reflect.String:
			b, err := strconv.ParseBool(v.String())
			if err != nil {
				return nil, fmt.Errorf("failed to convert string to bool: %v", err)
			}
			return b, nil
		}

	case "datetime", "date", "time":
		switch tv := srcValue.(type) {
		case time.Time:
			return tv, nil
		case string:
			t, err := time.Parse(time.RFC3339, tv)
			if err != nil {
				t, err = time.Parse("2006-01-02 15:04:05", tv)
				if err != nil {
					return nil, fmt.Errorf("failed to convert string to time: %v", err)
				}
			}
			return t, nil
		}
	}

	return srcValue, nil
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listTasks(w, r)
	case http.MethodPost:
		createTask(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleTaskByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}
	
	taskID := parts[0]
	
	if len(parts) > 1 {
		switch parts[1] {
		case "pause":
			if r.Method == http.MethodPost {
				pauseTask(w, r, taskID)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		case "resume":
			if r.Method == http.MethodPost {
				resumeTask(w, r, taskID)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		default:
			writeError(w, http.StatusNotFound, "Endpoint not found")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTask(w, r, taskID)
	case http.MethodDelete:
		deleteTask(w, r, taskID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func listTasks(w http.ResponseWriter, r *http.Request) {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	rows, err := db.Query(`SELECT id, config_id, status, progress_json, start_time, end_time, error_message 
		FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	taskList := make([]*MigrationTask, 0)
	for rows.Next() {
		var id, configID, status string
		var progressJSON, errorMessage sql.NullString
		var startTime, endTime sql.NullTime
		
		if err := rows.Scan(&id, &configID, &status, &progressJSON, &startTime, &endTime, &errorMessage); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		task := &MigrationTask{
			ID:           id,
			ConfigID:     configID,
			Status:       status,
			ErrorMessage: errorMessage.String,
		}
		
		if progressJSON.Valid {
			var progress MigrationProgress
			if err := json.Unmarshal([]byte(progressJSON.String), &progress); err == nil {
				task.Progress = &progress
			}
		}
		
		if startTime.Valid {
			t := startTime.Time
			task.StartTime = &t
		}
		if endTime.Valid {
			t := endTime.Time
			task.EndTime = &t
		}
		
		taskList = append(taskList, task)
	}

	writeJSON(w, http.StatusOK, taskList)
}

func getLastSuccessfulProgress(configID string) (*MigrationProgress, error) {
	var progressJSON sql.NullString
	err := db.QueryRow(`SELECT progress_json FROM tasks 
		WHERE config_id = ? AND status = ? 
		ORDER BY created_at DESC LIMIT 1`, configID, StatusCompleted).Scan(&progressJSON)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if !progressJSON.Valid {
		return nil, nil
	}
	
	var progress MigrationProgress
	if err := json.Unmarshal([]byte(progressJSON.String), &progress); err != nil {
		return nil, err
	}
	
	return &progress, nil
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfigID string `json:"config_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.ConfigID == "" {
		writeError(w, http.StatusBadRequest, "config_id is required")
		return
	}

	globalMutex.Lock()
	defer globalMutex.Unlock()

	var configJSON string
	err := db.QueryRow("SELECT config_json FROM configs WHERE id = ?", req.ConfigID).Scan(&configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var config MigrationConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	lastProgress, err := getLastSuccessfulProgress(req.ConfigID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	taskID := generateID()
	task := &MigrationTask{
		ID:       taskID,
		ConfigID: req.ConfigID,
		Status:   StatusPending,
		Progress: &MigrationProgress{
			TaskID:          taskID,
			ConfigID:        req.ConfigID,
			RecordsMigrated: 0,
			RecordsFailed:   0,
			TotalRecords:    0,
		},
	}

	if lastProgress != nil {
		task.Progress.LastTimestamp = lastProgress.LastTimestamp
		task.Progress.LastPrimaryKey = lastProgress.LastPrimaryKey
	}

	progressJSON, _ := json.Marshal(task.Progress)
	_, err = db.Exec(
		"INSERT INTO tasks (id, config_id, status, progress_json) VALUES (?, ?, ?, ?)",
		taskID, req.ConfigID, StatusPending, string(progressJSON),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	configs[req.ConfigID] = &config
	tasks[taskID] = task

	isIncremental := lastProgress != nil && lastProgress.LastTimestamp != ""
	go executeMigration(taskID, &config, isIncremental)

	writeJSON(w, http.StatusCreated, task)
}

func getTask(w http.ResponseWriter, r *http.Request, taskID string) {
	globalMutex.RLock()
	defer globalMutex.RUnlock()

	var configID, status string
	var progressJSON, errorMessage sql.NullString
	var startTime, endTime sql.NullTime
	
	err := db.QueryRow(`SELECT config_id, status, progress_json, start_time, end_time, error_message 
		FROM tasks WHERE id = ?`, taskID).Scan(&configID, &status, &progressJSON, &startTime, &endTime, &errorMessage)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	task := &MigrationTask{
		ID:           taskID,
		ConfigID:     configID,
		Status:       status,
		ErrorMessage: errorMessage.String,
	}
	
	if progressJSON.Valid {
		var progress MigrationProgress
		if err := json.Unmarshal([]byte(progressJSON.String), &progress); err == nil {
			task.Progress = &progress
		}
	}
	
	if startTime.Valid {
		t := startTime.Time
		task.StartTime = &t
	}
	if endTime.Valid {
		t := endTime.Time
		task.EndTime = &t
	}

	writeJSON(w, http.StatusOK, task)
}

func deleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if task, exists := tasks[taskID]; exists {
		task.stopRequested = true
	}

	_, err := db.Exec("DELETE FROM error_records WHERE task_id = ?", taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	delete(tasks, taskID)
	w.WriteHeader(http.StatusNoContent)
}

func pauseTask(w http.ResponseWriter, r *http.Request, taskID string) {
	globalMutex.Lock()
	task, exists := tasks[taskID]
	if !exists {
		globalMutex.Unlock()
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	globalMutex.Unlock()

	task.mutex.Lock()
	defer task.mutex.Unlock()

	if task.Status != StatusRunning {
		writeError(w, http.StatusBadRequest, "Task is not running")
		return
	}

	task.pauseRequested = true
	writeJSON(w, http.StatusOK, map[string]string{"message": "Pause requested"})
}

func resumeTask(w http.ResponseWriter, r *http.Request, taskID string) {
	globalMutex.Lock()
	
	var configID, status string
	var progressJSON sql.NullString
	
	err := db.QueryRow(`SELECT config_id, status, progress_json FROM tasks WHERE id = ?`, taskID).Scan(
		&configID, &status, &progressJSON)
	if err != nil {
		globalMutex.Unlock()
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if status != StatusPaused {
		globalMutex.Unlock()
		writeError(w, http.StatusBadRequest, "Task is not paused")
		return
	}

	var configJSON string
	err = db.QueryRow("SELECT config_json FROM configs WHERE id = ?", configID).Scan(&configJSON)
	if err != nil {
		globalMutex.Unlock()
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var config MigrationConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		globalMutex.Unlock()
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.Exec("UPDATE tasks SET status = ? WHERE id = ?", StatusRunning, taskID)
	if err != nil {
		globalMutex.Unlock()
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var progress MigrationProgress
	if progressJSON.Valid {
		json.Unmarshal([]byte(progressJSON.String), &progress)
	}

	task := &MigrationTask{
		ID:       taskID,
		ConfigID: configID,
		Status:   StatusRunning,
		Progress: &progress,
	}

	configs[configID] = &config
	tasks[taskID] = task
	globalMutex.Unlock()

	go executeMigration(taskID, &config, true)

	writeJSON(w, http.StatusOK, map[string]string{"message": "Resume requested"})
}

func handleTaskPauseResume(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "Use /tasks/{id}/pause or /tasks/{id}/resume")
}

func handleErrorsByTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	taskID := strings.TrimPrefix(r.URL.Path, "/errors/")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	rows, err := db.Query(`SELECT id, task_id, table_name, record_key, reason, created_at 
		FROM error_records WHERE task_id = ? ORDER BY id`, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	errors := make([]ErrorRecord, 0)
	for rows.Next() {
		var rec ErrorRecord
		var createdAt sql.NullTime
		if err := rows.Scan(&rec.ID, &rec.TaskID, &rec.TableName, &rec.RecordKey, &rec.Reason, &createdAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if createdAt.Valid {
			rec.CreatedAt = createdAt.Time
		}
		errors = append(errors, rec)
	}

	writeJSON(w, http.StatusOK, errors)
}

func logErrorRecord(taskID, tableName, recordKey, reason string) {
	db.Exec(
		"INSERT INTO error_records (task_id, table_name, record_key, reason) VALUES (?, ?, ?, ?)",
		taskID, tableName, recordKey, reason,
	)
}

func saveProgress(task *MigrationTask) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	progressJSON, _ := json.Marshal(task.Progress)
	
	var err error
	if task.ErrorMessage != "" {
		_, err = db.Exec(
			"UPDATE tasks SET status = ?, progress_json = ?, error_message = ? WHERE id = ?",
			task.Status, string(progressJSON), task.ErrorMessage, task.ID,
		)
	} else {
		_, err = db.Exec(
			"UPDATE tasks SET status = ?, progress_json = ? WHERE id = ?",
			task.Status, string(progressJSON), task.ID,
		)
	}
	
	if err != nil {
		log.Printf("Failed to save progress: %v", err)
	}
}

func executeMigration(taskID string, config *MigrationConfig, isResume bool) {
	globalMutex.RLock()
	task, exists := tasks[taskID]
	globalMutex.RUnlock()
	
	if !exists {
		log.Printf("Task %s not found", taskID)
		return
	}

	task.mutex.Lock()
	if task.stopRequested {
		task.mutex.Unlock()
		return
	}
	task.Status = StatusRunning
	now := time.Now()
	task.StartTime = &now
	task.pauseRequested = false
	task.stopRequested = false
	task.mutex.Unlock()

	saveProgress(task)

	sourceConn, err := connectToDB(&config.Source)
	if err != nil {
		task.mutex.Lock()
		task.Status = StatusFailed
		task.ErrorMessage = fmt.Sprintf("Failed to connect to source database: %v", err)
		endTime := time.Now()
		task.EndTime = &endTime
		task.mutex.Unlock()
		saveProgress(task)
		return
	}
	defer sourceConn.Close()

	targetConn, err := connectToDB(&config.Target)
	if err != nil {
		task.mutex.Lock()
		task.Status = StatusFailed
		task.ErrorMessage = fmt.Sprintf("Failed to connect to target database: %v", err)
		endTime := time.Now()
		task.EndTime = &endTime
		task.mutex.Unlock()
		saveProgress(task)
		return
	}
	defer targetConn.Close()

	var startTableIndex int = 0
	if isResume && task.Progress != nil && task.Progress.CurrentTable != "" {
		for i, table := range config.Tables {
			if table.SourceTable == task.Progress.CurrentTable {
				startTableIndex = i
				break
			}
		}
	}

	for i := startTableIndex; i < len(config.Tables); i++ {
		tableMapping := config.Tables[i]
		task.mutex.Lock()
		task.Progress.CurrentTable = tableMapping.SourceTable
		task.mutex.Unlock()
		saveProgress(task)

		err := migrateTable(task, sourceConn, targetConn, &tableMapping, isResume && i == startTableIndex)
		if err != nil {
			if err.Error() == "paused" {
				return
			}
			task.mutex.Lock()
			task.Status = StatusFailed
			task.ErrorMessage = fmt.Sprintf("Error migrating table %s: %v", tableMapping.SourceTable, err)
			endTime := time.Now()
			task.EndTime = &endTime
			task.mutex.Unlock()
			saveProgress(task)
			return
		}

		task.mutex.Lock()
		if task.pauseRequested {
			task.Status = StatusPaused
			task.mutex.Unlock()
			saveProgress(task)
			return
		}
		task.mutex.Unlock()
	}

	task.mutex.Lock()
	task.Status = StatusCompleted
	endTime := time.Now()
	task.EndTime = &endTime
	task.mutex.Unlock()
	saveProgress(task)
}

func checkPauseOrStop(task *MigrationTask) (shouldPause, shouldStop bool) {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	if task.stopRequested {
		return false, true
	}
	if task.pauseRequested {
		task.pauseRequested = false
		return true, false
	}
	return false, false
}

func migrateTable(task *MigrationTask, sourceConn, targetConn *sql.DB, 
	tableMapping *TableMapping, isResume bool) error {

	sourceCols, err := getColumns(sourceConn, tableMapping.SourceTable)
	if err != nil {
		return fmt.Errorf("failed to get source columns: %v", err)
	}

	targetCols, err := getColumns(targetConn, tableMapping.TargetTable)
	if err != nil {
		return fmt.Errorf("failed to get target columns: %v", err)
	}

	sourceFieldMap := make(map[string]bool)
	for _, col := range sourceCols {
		sourceFieldMap[col] = true
	}

	targetFieldMap := make(map[string]bool)
	for _, col := range targetCols {
		targetFieldMap[col] = true
	}

	for _, fm := range tableMapping.Fields {
		if !sourceFieldMap[fm.Source] {
			return fmt.Errorf("source field %s not found in table %s", fm.Source, tableMapping.SourceTable)
		}
		if !targetFieldMap[fm.Target] {
			return fmt.Errorf("target field %s not found in table %s", fm.Target, tableMapping.TargetTable)
		}
	}

	sourceFields := make([]string, 0, len(tableMapping.Fields))
	targetFields := make([]string, 0, len(tableMapping.Fields))
	for _, fm := range tableMapping.Fields {
		sourceFields = append(sourceFields, fm.Source)
		targetFields = append(targetFields, fm.Target)
	}

	totalRows, err := getTotalRows(sourceConn, tableMapping.SourceTable, tableMapping, task.Progress)
	if err != nil {
		return fmt.Errorf("failed to count rows: %v", err)
	}
	task.mutex.Lock()
	task.Progress.TotalRecords = totalRows
	task.mutex.Unlock()
	saveProgress(task)

	var lastPK interface{}
	var lastTimestamp time.Time
	
	if isResume && task.Progress != nil {
		if task.Progress.LastPrimaryKey != "" {
			lastPK = task.Progress.LastPrimaryKey
		}
		if task.Progress.LastTimestamp != "" {
			lastTimestamp, _ = time.Parse(time.RFC3339, task.Progress.LastTimestamp)
		}
	}

	for {
		shouldPause, shouldStop := checkPauseOrStop(task)
		if shouldStop {
			return nil
		}
		if shouldPause {
			return fmt.Errorf("paused")
		}

		rows, newLastPK, newLastTimestamp, err := fetchBatch(sourceConn, tableMapping, 
			sourceFields, lastPK, lastTimestamp)
		if err != nil {
			return fmt.Errorf("failed to fetch batch: %v", err)
		}

		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			shouldPause, shouldStop := checkPauseOrStop(task)
			if shouldStop {
				return nil
			}
			if shouldPause {
				return fmt.Errorf("paused")
			}

			recordKey := fmt.Sprintf("%v", row[tableMapping.PrimaryKey])
			
			err := insertRecord(targetConn, tableMapping, targetFields, row)
			if err != nil {
				logErrorRecord(task.ID, tableMapping.SourceTable, recordKey, err.Error())
				task.mutex.Lock()
				task.Progress.RecordsFailed++
				task.mutex.Unlock()
				saveProgress(task)
				continue
			}

			task.mutex.Lock()
			task.Progress.RecordsMigrated++
			task.mutex.Unlock()
			saveProgress(task)
		}

		if newLastPK != nil {
			lastPK = newLastPK
			task.mutex.Lock()
			task.Progress.LastPrimaryKey = fmt.Sprintf("%v", newLastPK)
			if !newLastTimestamp.IsZero() {
				task.Progress.LastTimestamp = newLastTimestamp.Format(time.RFC3339)
			}
			task.mutex.Unlock()
			saveProgress(task)
		}

		if len(rows) < BatchSize {
			break
		}
	}

	return nil
}

func getTotalRows(conn *sql.DB, tableName string, mapping *TableMapping, progress *MigrationProgress) (int64, error) {
	var query string
	var args []interface{}

	if mapping.TimestampField != "" && progress != nil && progress.LastTimestamp != "" {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s > ?", tableName, mapping.TimestampField)
		args = append(args, progress.LastTimestamp)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	}

	var count int64
	err := conn.QueryRow(query, args...).Scan(&count)
	return count, err
}

func fetchBatch(conn *sql.DB, mapping *TableMapping, fields []string, 
	lastPK interface{}, lastTimestamp time.Time) ([]map[string]interface{}, interface{}, time.Time, error) {

	fieldList := strings.Join(fields, ", ")
	var query string
	var args []interface{}

	query = fmt.Sprintf("SELECT %s FROM %s", fieldList, mapping.SourceTable)

	conditions := make([]string, 0)
	
	if mapping.TimestampField != "" && !lastTimestamp.IsZero() {
		conditions = append(conditions, fmt.Sprintf("%s > ?", mapping.TimestampField))
		args = append(args, lastTimestamp)
	}

	if lastPK != nil {
		conditions = append(conditions, fmt.Sprintf("%s > ?", mapping.PrimaryKey))
		args = append(args, lastPK)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := fmt.Sprintf(" ORDER BY %s", mapping.PrimaryKey)
	if mapping.TimestampField != "" {
		orderBy = fmt.Sprintf(" ORDER BY %s, %s", mapping.TimestampField, mapping.PrimaryKey)
	}
	query += orderBy + fmt.Sprintf(" LIMIT %d", BatchSize)

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	result := make([]map[string]interface{}, 0)
	var newLastPK interface{}
	var newLastTimestamp time.Time

	for rows.Next() {
		values := make([]interface{}, len(columnTypes))
		valuePtrs := make([]interface{}, len(columnTypes))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, time.Time{}, err
		}

		row := make(map[string]interface{})
		for i, ct := range columnTypes {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[ct.Name()] = string(b)
			} else {
				row[ct.Name()] = val
			}

			if ct.Name() == mapping.PrimaryKey {
				newLastPK = row[ct.Name()]
			}
			if mapping.TimestampField != "" && ct.Name() == mapping.TimestampField {
				if t, ok := row[ct.Name()].(time.Time); ok {
					newLastTimestamp = t
				}
			}
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, time.Time{}, err
	}

	return result, newLastPK, newLastTimestamp, nil
}

func insertRecord(conn *sql.DB, mapping *TableMapping, targetFields []string, 
	row map[string]interface{}) error {

	values := make([]interface{}, 0, len(mapping.Fields))
	placeholders := make([]string, 0, len(mapping.Fields))

	for _, fm := range mapping.Fields {
		srcVal := row[fm.Source]
		
		convertedVal, err := convertValue(srcVal, "TEXT")
		if err != nil {
			return fmt.Errorf("field %s conversion failed: %v", fm.Source, err)
		}
		
		values = append(values, convertedVal)
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		mapping.TargetTable,
		strings.Join(targetFields, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := conn.Exec(sql, values...)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}

	return nil
}

