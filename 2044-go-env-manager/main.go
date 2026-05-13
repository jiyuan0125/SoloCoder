package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type EnvVar struct {
	Project    string    `json:"project"`
	Environment string   `json:"environment"`
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ChangeHistory struct {
	ID         int64     `json:"id"`
	Project    string    `json:"project"`
	Environment string   `json:"environment"`
	Key        string    `json:"key"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	ChangedAt  time.Time `json:"changed_at"`
	ChangedBy  string    `json:"changed_by"`
}

type EnvDiff struct {
	OnlyInA   map[string]string `json:"only_in_a"`
	OnlyInB   map[string]string `json:"only_in_b"`
	Different map[string]struct {
		ValueA string `json:"value_a"`
		ValueB string `json:"value_b"`
	} `json:"different"`
}

type ImportResult struct {
	Imported    int   `json:"imported"`
	Skipped     []int `json:"skipped"`
	SkippedLine []string `json:"skipped_lines"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "./env_manager.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	if err = initDatabase(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/projects/", projectHandler)
	http.HandleFunc("/variables/", variableHandler)
	http.HandleFunc("/import/", importHandler)
	http.HandleFunc("/export/", exportHandler)
	http.HandleFunc("/diff/", diffHandler)
	http.HandleFunc("/history/", historyHandler)

	fmt.Println("Server starting on :8104")
	if err := http.ListenAndServe(":8104", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func initDatabase() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		name TEXT PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS env_vars (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		environment TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project) REFERENCES projects(name),
		UNIQUE(project, environment, key)
	);

	CREATE TABLE IF NOT EXISTS change_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		environment TEXT NOT NULL,
		key TEXT NOT NULL,
		old_value TEXT NOT NULL,
		new_value TEXT NOT NULL,
		changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		changed_by TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_env_vars_project ON env_vars(project);
	CREATE INDEX IF NOT EXISTS idx_env_vars_project_env ON env_vars(project, environment);
	CREATE INDEX IF NOT EXISTS idx_history_project_env ON change_history(project, environment);
	`

	_, err := db.Exec(schema)
	return err
}

func projectExists(project string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE name = ?)", project).Scan(&exists)
	return exists, err
}

func projectHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	segments := strings.Split(path, "/")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			listProjects(w, r)
		case http.MethodPost:
			createProject(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	projectName := segments[0]
	exists, err := projectExists(projectName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getProject(w, r, projectName)
	case http.MethodDelete:
		deleteProject(w, r, projectName)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT name, created_at FROM projects ORDER BY name")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing projects: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Project struct {
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
	}

	projects := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.Name, &p.CreatedAt); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning project: %v", err), http.StatusInternalServerError)
			return
		}
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	exists, err := projectExists(req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Project already exists", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO projects (name) VALUES (?)", req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating project: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"name": req.Name})
}

func getProject(w http.ResponseWriter, r *http.Request, name string) {
	row := db.QueryRow("SELECT name, created_at FROM projects WHERE name = ?", name)
	var project struct {
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
	}

	if err := row.Scan(&project.Name, &project.CreatedAt); err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func deleteProject(w http.ResponseWriter, r *http.Request, name string) {
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec("DELETE FROM change_history WHERE project = ?", name); err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error deleting history: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec("DELETE FROM env_vars WHERE project = ?", name); err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error deleting variables: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec("DELETE FROM projects WHERE name = ?", name); err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error deleting project: %v", err), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("Error committing transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validEnvironment(env string) bool {
	return env == "dev" || env == "staging" || env == "prod"
}

func variableHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/variables/")
	segments := strings.Split(path, "/")

	if len(segments) < 2 {
		http.Error(w, "Invalid path. Use /variables/{project}/{environment}", http.StatusBadRequest)
		return
	}

	project := segments[0]
	environment := segments[1]
	key := ""
	if len(segments) >= 3 {
		key = segments[2]
	}

	exists, err := projectExists(project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !validEnvironment(environment) {
		http.Error(w, "Invalid environment. Must be dev, staging, or prod", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if key == "" {
			listVariables(w, r, project, environment)
		} else {
			getVariable(w, r, project, environment, key)
		}
	case http.MethodPost, http.MethodPut:
		setVariable(w, r, project, environment)
	case http.MethodDelete:
		if key == "" {
			http.Error(w, "Key is required for delete", http.StatusBadRequest)
			return
		}
		deleteVariable(w, r, project, environment, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func listVariables(w http.ResponseWriter, r *http.Request, project, environment string) {
	rows, err := db.Query(`
		SELECT key, value, updated_at 
		FROM env_vars 
		WHERE project = ? AND environment = ?
		ORDER BY key
	`, project, environment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing variables: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	vars := []EnvVar{}
	for rows.Next() {
		var v EnvVar
		v.Project = project
		v.Environment = environment
		if err := rows.Scan(&v.Key, &v.Value, &v.UpdatedAt); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning variable: %v", err), http.StatusInternalServerError)
			return
		}
		vars = append(vars, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars)
}

func getVariable(w http.ResponseWriter, r *http.Request, project, environment, key string) {
	row := db.QueryRow(`
		SELECT value, updated_at 
		FROM env_vars 
		WHERE project = ? AND environment = ? AND key = ?
	`, project, environment, key)

	var v EnvVar
	v.Project = project
	v.Environment = environment
	v.Key = key

	if err := row.Scan(&v.Value, &v.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Variable not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error getting variable: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func setVariable(w http.ResponseWriter, r *http.Request, project, environment string) {
	var req struct {
		Key       string `json:"key"`
		Value     string `json:"value"`
		ChangedBy string `json:"changed_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	var oldValue string
	err = tx.QueryRow(`
		SELECT value FROM env_vars 
		WHERE project = ? AND environment = ? AND key = ?
	`, project, environment, req.Key).Scan(&oldValue)

	isUpdate := err == nil
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error checking existing value: %v", err), http.StatusInternalServerError)
		return
	}

	if isUpdate {
		if oldValue == req.Value {
			tx.Rollback()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "No changes needed, value is the same",
			})
			return
		}

		_, err = tx.Exec(`
			UPDATE env_vars 
			SET value = ?, updated_at = CURRENT_TIMESTAMP 
			WHERE project = ? AND environment = ? AND key = ?
		`, req.Value, project, environment, req.Key)
	} else {
		_, err = tx.Exec(`
			INSERT INTO env_vars (project, environment, key, value) 
			VALUES (?, ?, ?, ?)
		`, project, environment, req.Key, req.Value)
	}

	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error setting variable: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		INSERT INTO change_history (project, environment, key, old_value, new_value, changed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, project, environment, req.Key, oldValue, req.Value, req.ChangedBy)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error recording history: %v", err), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("Error committing transaction: %v", err), http.StatusInternalServerError)
		return
	}

	if isUpdate {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(EnvVar{
		Project:     project,
		Environment: environment,
		Key:         req.Key,
		Value:       req.Value,
		UpdatedAt:   time.Now(),
	})
}

func deleteVariable(w http.ResponseWriter, r *http.Request, project, environment, key string) {
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	var oldValue string
	err = tx.QueryRow(`
		SELECT value FROM env_vars 
		WHERE project = ? AND environment = ? AND key = ?
	`, project, environment, key).Scan(&oldValue)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			http.Error(w, "Variable not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error getting variable: %v", err), http.StatusInternalServerError)
		}
		return
	}

	_, err = tx.Exec(`
		DELETE FROM env_vars 
		WHERE project = ? AND environment = ? AND key = ?
	`, project, environment, key)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error deleting variable: %v", err), http.StatusInternalServerError)
		return
	}

	var changedBy string
	if ua := r.Header.Get("X-Changed-By"); ua != "" {
		changedBy = ua
	}

	_, err = tx.Exec(`
		INSERT INTO change_history (project, environment, key, old_value, new_value, changed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, project, environment, key, oldValue, "", changedBy)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Error recording history: %v", err), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("Error committing transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/export/")
	segments := strings.Split(path, "/")

	if len(segments) < 2 {
		http.Error(w, "Invalid path. Use /export/{project}/{environment}", http.StatusBadRequest)
		return
	}

	project := segments[0]
	environment := segments[1]

	exists, err := projectExists(project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !validEnvironment(environment) {
		http.Error(w, "Invalid environment. Must be dev, staging, or prod", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
		SELECT key, value 
		FROM env_vars 
		WHERE project = ? AND environment = ?
		ORDER BY key
	`, project, environment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing variables: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Exported for project: %s, environment: %s\n", project, environment))
	sb.WriteString(fmt.Sprintf("# Generated at: %s\n", time.Now().Format(time.RFC3339)))

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning variable: %v", err), http.StatusInternalServerError)
			return
		}
		sb.WriteString(formatEnvLine(key, value))
		sb.WriteString("\n")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.env", project, environment))
	w.Write([]byte(sb.String()))
}

func formatEnvLine(key, value string) string {
	needsQuoting := false
	for _, r := range value {
		if r == '\n' || r == '"' || r == '\'' || r == '$' || r == '=' || r == ' ' {
			needsQuoting = true
			break
		}
	}

	if !needsQuoting {
		return fmt.Sprintf("%s=%s", key, value)
	}

	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")

	return fmt.Sprintf("%s=\"%s\"", key, escaped)
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/import/")
	segments := strings.Split(path, "/")

	if len(segments) < 2 {
		http.Error(w, "Invalid path. Use /import/{project}/{environment}", http.StatusBadRequest)
		return
	}

	project := segments[0]
	environment := segments[1]

	exists, err := projectExists(project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !validEnvironment(environment) {
		http.Error(w, "Invalid environment. Must be dev, staging, or prod", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusInternalServerError)
		return
	}

	if !utf8.Valid(body) {
		body = []byte(strings.ToValidUTF8(string(body), "\uFFFD"))
	}

	changedBy := r.Header.Get("X-Changed-By")
	result := parseEnvAndImport(string(body), project, environment, changedBy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func parseEnvAndImport(content, project, environment, changedBy string) ImportResult {
	result := ImportResult{
		Skipped:     []int{},
		SkippedLine: []string{},
	}

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		kv := parseEnvLine(trimmed)
		if kv == nil {
			result.Skipped = append(result.Skipped, lineNum)
			result.SkippedLine = append(result.SkippedLine, line)
			continue
		}

		err := importOrUpdateVariable(project, environment, kv.Key, kv.Value, changedBy)
		if err != nil {
			result.Skipped = append(result.Skipped, lineNum)
			result.SkippedLine = append(result.SkippedLine, line)
			continue
		}

		result.Imported++
	}

	return result
}

type KeyValue struct {
	Key   string
	Value string
}

func parseEnvLine(line string) *KeyValue {
	eqIdx := strings.Index(line, "=")
	if eqIdx == -1 {
		return nil
	}

	key := strings.TrimSpace(line[:eqIdx])
	if key == "" {
		return nil
	}

	valueRaw := strings.TrimSpace(line[eqIdx+1:])
	var value string

	if len(valueRaw) >= 2 {
		first := valueRaw[0]
		last := valueRaw[len(valueRaw)-1]

		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			quoted := valueRaw[1 : len(valueRaw)-1]
			if first == '"' {
				var err error
				value, err = unescapeQuotedValue(quoted)
				if err != nil {
					return nil
				}
			} else {
				value = quoted
			}
		} else {
			commentIdx := strings.Index(valueRaw, "#")
			if commentIdx != -1 {
				value = strings.TrimSpace(valueRaw[:commentIdx])
			} else {
				value = valueRaw
			}
		}
	} else {
		commentIdx := strings.Index(valueRaw, "#")
		if commentIdx != -1 {
			value = strings.TrimSpace(valueRaw[:commentIdx])
		} else {
			value = valueRaw
		}
	}

	return &KeyValue{Key: key, Value: value}
}

func unescapeQuotedValue(s string) (string, error) {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
				i += 2
			case 'r':
				result.WriteByte('\r')
				i += 2
			case 't':
				result.WriteByte('\t')
				i += 2
			case '\\', '"', '\'':
				result.WriteByte(s[i+1])
				i += 2
			default:
				result.WriteByte(s[i])
				i++
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String(), nil
}

func importOrUpdateVariable(project, environment, key, value, changedBy string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var oldValue string
	err = tx.QueryRow(`
		SELECT value FROM env_vars 
		WHERE project = ? AND environment = ? AND key = ?
	`, project, environment, key).Scan(&oldValue)

	isUpdate := err == nil
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	if isUpdate {
		if oldValue == value {
			tx.Rollback()
			return nil
		}

		_, err = tx.Exec(`
			UPDATE env_vars 
			SET value = ?, updated_at = CURRENT_TIMESTAMP 
			WHERE project = ? AND environment = ? AND key = ?
		`, value, project, environment, key)
	} else {
		_, err = tx.Exec(`
			INSERT INTO env_vars (project, environment, key, value) 
			VALUES (?, ?, ?, ?)
		`, project, environment, key, value)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO change_history (project, environment, key, old_value, new_value, changed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, project, environment, key, oldValue, value, changedBy)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func diffHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/diff/")
	segments := strings.Split(path, "/")

	if len(segments) < 3 {
		http.Error(w, "Invalid path. Use /diff/{project}/{envA}/{envB}", http.StatusBadRequest)
		return
	}

	project := segments[0]
	envA := segments[1]
	envB := segments[2]

	exists, err := projectExists(project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !validEnvironment(envA) || !validEnvironment(envB) {
		http.Error(w, "Invalid environment. Must be dev, staging, or prod", http.StatusBadRequest)
		return
	}

	varsA, err := getVarsMap(project, envA)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting variables for %s: %v", envA, err), http.StatusInternalServerError)
		return
	}

	varsB, err := getVarsMap(project, envB)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting variables for %s: %v", envB, err), http.StatusInternalServerError)
		return
	}

	diff := EnvDiff{
		OnlyInA:   make(map[string]string),
		OnlyInB:   make(map[string]string),
		Different: make(map[string]struct {
			ValueA string `json:"value_a"`
			ValueB string `json:"value_b"`
		}),
	}

	keys := make(map[string]bool)
	for k := range varsA {
		keys[k] = true
	}
	for k := range varsB {
		keys[k] = true
	}

	for k := range keys {
		valA, inA := varsA[k]
		valB, inB := varsB[k]

		if inA && !inB {
			diff.OnlyInA[k] = valA
		} else if !inA && inB {
			diff.OnlyInB[k] = valB
		} else if valA != valB {
			diff.Different[k] = struct {
				ValueA string `json:"value_a"`
				ValueB string `json:"value_b"`
			}{ValueA: valA, ValueB: valB}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diff)
}

func getVarsMap(project, environment string) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT key, value 
		FROM env_vars 
		WHERE project = ? AND environment = ?
	`, project, environment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vars := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		vars[k] = v
	}

	return vars, nil
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/history/")
	segments := strings.Split(path, "/")

	if len(segments) < 2 {
		http.Error(w, "Invalid path. Use /history/{project}/{environment}[/{key}]", http.StatusBadRequest)
		return
	}

	project := segments[0]
	environment := segments[1]
	key := ""
	if len(segments) >= 3 {
		key = segments[2]
	}

	exists, err := projectExists(project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking project: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	if !validEnvironment(environment) {
		http.Error(w, "Invalid environment. Must be dev, staging, or prod", http.StatusBadRequest)
		return
	}

	var rows *sql.Rows
	if key == "" {
		rows, err = db.Query(`
			SELECT id, project, environment, key, old_value, new_value, changed_at, changed_by
			FROM change_history
			WHERE project = ? AND environment = ?
			ORDER BY changed_at DESC
		`, project, environment)
	} else {
		rows, err = db.Query(`
			SELECT id, project, environment, key, old_value, new_value, changed_at, changed_by
			FROM change_history
			WHERE project = ? AND environment = ? AND key = ?
			ORDER BY changed_at DESC
		`, project, environment, key)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting history: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := []ChangeHistory{}
	for rows.Next() {
		var h ChangeHistory
		var changedBy sql.NullString
		if err := rows.Scan(&h.ID, &h.Project, &h.Environment, &h.Key, &h.OldValue, &h.NewValue, &h.ChangedAt, &changedBy); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning history: %v", err), http.StatusInternalServerError)
			return
		}
		if changedBy.Valid {
			h.ChangedBy = changedBy.String
		}
		history = append(history, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
