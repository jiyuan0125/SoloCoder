package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	dbPath         = "./config_center.db"
	defaultPort    = "9801"
	envDevelopment = "dev"
	envTesting     = "test"
	envProduction  = "prod"
)

var envHierarchy = map[string][]string{
	envProduction:  {envProduction, envTesting, envDevelopment},
	envTesting:     {envTesting, envDevelopment},
	envDevelopment: {envDevelopment},
}

type ConfigItem struct {
	ID        int64     `json:"id"`
	Service   string    `json:"service"`
	Env       string    `json:"env"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	ValueType string    `json:"value_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ConfigHistory struct {
	ID        int64     `json:"id"`
	ConfigID  int64     `json:"config_id"`
	Service   string    `json:"service"`
	Env       string    `json:"env"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	ValueType string    `json:"value_type"`
	Version   int       `json:"version"`
	Action    string    `json:"action"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type MasterEntity struct {
	ID        int64     `json:"id"`
	Service   string    `json:"service"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DetailRecord struct {
	ID         int64     `json:"id"`
	MasterID   int64     `json:"master_id"`
	Service    string    `json:"service"`
	Content    string    `json:"content"`
	Version    int       `json:"version"`
	ChangeNote string    `json:"change_note"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConfigRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type RollbackRequest struct {
	Version int    `json:"version"`
	Comment string `json:"comment"`
}

var db *sql.DB

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/services/", serviceHandler)
	http.HandleFunc("/configs/", configHandler)
	http.HandleFunc("/configs/history/", historyHandler)
	http.HandleFunc("/configs/rollback/", rollbackHandler)
	http.HandleFunc("/masters/", masterDetailHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("配置中心服务启动，监听端口 %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS config_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service TEXT NOT NULL,
		env TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		value_type TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(service, env, key)
	);

	CREATE TABLE IF NOT EXISTS config_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_id INTEGER NOT NULL,
		service TEXT NOT NULL,
		env TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		value_type TEXT NOT NULL,
		version INTEGER NOT NULL,
		action TEXT NOT NULL,
		comment TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(config_id) REFERENCES config_items(id)
	);

	CREATE TABLE IF NOT EXISTS master_entities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(service, name)
	);

	CREATE TABLE IF NOT EXISTS detail_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		master_id INTEGER NOT NULL,
		service TEXT NOT NULL,
		content TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		change_note TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(master_id) REFERENCES master_entities(id)
	);

	CREATE INDEX IF NOT EXISTS idx_config_service_env ON config_items(service, env);
	CREATE INDEX IF NOT EXISTS idx_history_config_id ON config_history(config_id);
	CREATE INDEX IF NOT EXISTS idx_detail_master_id ON detail_records(master_id);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	return nil
}

func detectValueType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}, []interface{}:
		return "object"
	default:
		return "string"
	}
}

func getValueType(v string) string {
	if v == "true" || v == "false" {
		return "boolean"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return "number"
	}
	var obj interface{}
	if err := json.Unmarshal([]byte(v), &obj); err == nil {
		if _, isMap := obj.(map[string]interface{}); isMap {
			return "object"
		}
		if _, isArr := obj.([]interface{}); isArr {
			return "object"
		}
	}
	return "string"
}

func valueToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(val), nil
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return "", fmt.Errorf("不支持的类型: %T", v)
	}
}

func parseValueByType(v string, valueType string) (interface{}, error) {
	switch valueType {
	case "string":
		return v, nil
	case "number":
		return strconv.ParseFloat(v, 64)
	case "boolean":
		return strconv.ParseBool(v)
	case "object":
		var obj interface{}
		if err := json.Unmarshal([]byte(v), &obj); err != nil {
			return nil, err
		}
		return obj, nil
	default:
		return v, nil
	}
}

func checkTypeMatch(newValue interface{}, expectedType string) bool {
	actualType := detectValueType(newValue)
	return actualType == expectedType
}

func serviceExists(service string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT (SELECT COUNT(*) FROM config_items WHERE service = ?) + 
		       (SELECT COUNT(*) FROM master_entities WHERE service = ?)
	`, service, service).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT DISTINCT service FROM config_items UNION SELECT DISTINCT service FROM master_entities")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	services := make([]string, 0)
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			continue
		}
		services = append(services, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"services": services})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/configs/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "需要服务名和环境", http.StatusBadRequest)
		return
	}

	service := parts[0]
	env := parts[1]

	exists, err := serviceExists(service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists && r.Method == http.MethodGet {
		http.Error(w, "服务不存在", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetConfig(w, service, env)
	case http.MethodPost, http.MethodPut:
		handleSetConfig(w, r, service, env)
	case http.MethodDelete:
		handleDeleteConfig(w, r, service, env)
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func handleGetConfig(w http.ResponseWriter, service, env string) {
	envs, ok := envHierarchy[env]
	if !ok {
		envs = []string{env}
	}

	configs := make(map[string]interface{})
	seenKeys := make(map[string]bool)

	for _, e := range envs {
		rows, err := db.Query("SELECT key, value, value_type FROM config_items WHERE service = ? AND env = ?", service, e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var key, value, valueType string
			if err := rows.Scan(&key, &value, &valueType); err != nil {
				continue
			}
			if !seenKeys[key] {
				parsedValue, err := parseValueByType(value, valueType)
				if err != nil {
					configs[key] = value
				} else {
					configs[key] = parsedValue
				}
				seenKeys[key] = true
			}
		}
	}

	if len(configs) == 0 {
		http.Error(w, "服务不存在或无配置", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": service,
		"env":     env,
		"configs": configs,
	})
}

func handleSetConfig(w http.ResponseWriter, r *http.Request, service, env string) {
	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}
	if req.Key == "" {
		http.Error(w, "配置键不能为空", http.StatusBadRequest)
		return
	}

	newType := detectValueType(req.Value)
	newValueStr, err := valueToString(req.Value)
	if err != nil {
		http.Error(w, "值转换失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var existingID int64
	var existingType string
	var existingVersion int

	err = tx.QueryRow("SELECT id, value_type FROM config_items WHERE service = ? AND env = ? AND key = ?", service, env, req.Key).Scan(&existingID, &existingType)

	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO config_items (service, env, key, value, value_type) VALUES (?, ?, ?, ?, ?)", service, env, req.Key, newValueStr, newType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		existingID, _ = result.LastInsertId()
		existingVersion = 0
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		if !checkTypeMatch(req.Value, existingType) {
			http.Error(w, fmt.Sprintf("类型不匹配: 期望 %s, 实际 %s", existingType, newType), http.StatusBadRequest)
			return
		}
		_, err = tx.Exec("UPDATE config_items SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", newValueStr, existingID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tx.QueryRow("SELECT COALESCE(MAX(version), 0) FROM config_history WHERE config_id = ?", existingID).Scan(&existingVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	_, err = tx.Exec("INSERT INTO config_history (config_id, service, env, key, value, value_type, version, action, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		existingID, service, env, req.Key, newValueStr, newType, existingVersion+1, "update", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置保存成功",
	})
}

func handleDeleteConfig(w http.ResponseWriter, r *http.Request, service, env string) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "需要提供 key 参数", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var id int64
	var value, valueType string
	err = tx.QueryRow("SELECT id, value, value_type FROM config_items WHERE service = ? AND env = ? AND key = ?", service, env, key).Scan(&id, &value, &valueType)
	if err == sql.ErrNoRows {
		http.Error(w, "配置不存在", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxVersion int
	err = tx.QueryRow("SELECT COALESCE(MAX(version), 0) FROM config_history WHERE config_id = ?", id).Scan(&maxVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO config_history (config_id, service, env, key, value, value_type, version, action, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, service, env, key, value, valueType, maxVersion+1, "delete", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM config_items WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置删除成功",
	})
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/configs/history/")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 3 {
		http.Error(w, "需要服务名、环境和配置键", http.StatusBadRequest)
		return
	}

	service := parts[0]
	env := parts[1]
	key := parts[2]

	rows, err := db.Query("SELECT id, config_id, service, env, key, value, value_type, version, action, comment, created_at FROM config_history WHERE service = ? AND env = ? AND key = ? ORDER BY version DESC", service, env, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	histories := make([]ConfigHistory, 0)
	for rows.Next() {
		var h ConfigHistory
		if err := rows.Scan(&h.ID, &h.ConfigID, &h.Service, &h.Env, &h.Key, &h.Value, &h.ValueType, &h.Version, &h.Action, &h.Comment, &h.CreatedAt); err != nil {
			continue
		}
		histories = append(histories, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   service,
		"env":       env,
		"key":       key,
		"histories": histories,
	})
}

func rollbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/configs/rollback/")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 3 {
		http.Error(w, "需要服务名、环境和配置键", http.StatusBadRequest)
		return
	}

	service := parts[0]
	env := parts[1]
	key := parts[2]

	var req RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}
	if req.Version <= 0 {
		http.Error(w, "版本号必须大于 0", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var targetValue, targetValueType string
	err = tx.QueryRow("SELECT value, value_type FROM config_history WHERE service = ? AND env = ? AND key = ? AND version = ?", service, env, key, req.Version).Scan(&targetValue, &targetValueType)
	if err == sql.ErrNoRows {
		http.Error(w, "指定版本不存在", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var configID int64
	err = tx.QueryRow("SELECT id FROM config_items WHERE service = ? AND env = ? AND key = ?", service, env, key).Scan(&configID)

	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO config_items (service, env, key, value, value_type) VALUES (?, ?, ?, ?, ?)", service, env, key, targetValue, targetValueType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		configID, _ = result.LastInsertId()
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		_, err = tx.Exec("UPDATE config_items SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", targetValue, configID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var maxVersion int
	err = tx.QueryRow("SELECT COALESCE(MAX(version), 0) FROM config_history WHERE service = ? AND env = ? AND key = ?", service, env, key).Scan(&maxVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	comment := fmt.Sprintf("回滚到版本 %d", req.Version)
	if req.Comment != "" {
		comment = fmt.Sprintf("%s: %s", comment, req.Comment)
	}

	_, err = tx.Exec("INSERT INTO config_history (config_id, service, env, key, value, value_type, version, action, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		configID, service, env, key, targetValue, targetValueType, maxVersion+1, "rollback", comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("已回滚到版本 %d", req.Version),
	})
}

func masterDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/masters/")
	parts := strings.SplitN(path, "/", 4)

	if len(parts) >= 3 && parts[1] == "details" {
		if r.Method != http.MethodGet {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			return
		}
		handleGetDetails(w, parts[0], parts[2])
		return
	}

	service := parts[0]

	switch r.Method {
	case http.MethodGet:
		handleGetMasters(w, service)
	case http.MethodPost:
		handleCreateMaster(w, r, service)
	case http.MethodPut:
		if len(parts) < 2 {
			http.Error(w, "需要主实体 ID", http.StatusBadRequest)
			return
		}
		handleUpdateMaster(w, r, service, parts[1])
	case http.MethodDelete:
		if len(parts) < 2 {
			http.Error(w, "需要主实体 ID", http.StatusBadRequest)
			return
		}
		handleDeleteMaster(w, service, parts[1])
	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func handleGetDetails(w http.ResponseWriter, service, masterIDStr string) {
	masterID, err := strconv.ParseInt(masterIDStr, 10, 64)
	if err != nil {
		http.Error(w, "无效的主实体 ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT id, master_id, service, content, version, change_note, created_at FROM detail_records WHERE master_id = ? AND service = ? ORDER BY version DESC", masterID, service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	details := make([]DetailRecord, 0)
	for rows.Next() {
		var d DetailRecord
		if err := rows.Scan(&d.ID, &d.MasterID, &d.Service, &d.Content, &d.Version, &d.ChangeNote, &d.CreatedAt); err != nil {
			continue
		}
		details = append(details, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"master_id": masterID,
		"service":   service,
		"details":   details,
	})
}

func handleGetMasters(w http.ResponseWriter, service string) {
	rows, err := db.Query("SELECT id, service, name, created_at, updated_at FROM master_entities WHERE service = ? ORDER BY id", service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	masters := make([]MasterEntity, 0)
	for rows.Next() {
		var m MasterEntity
		if err := rows.Scan(&m.ID, &m.Service, &m.Name, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		masters = append(masters, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": service,
		"masters": masters,
	})
}

func handleCreateMaster(w http.ResponseWriter, r *http.Request, service string) {
	var req struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "名称不能为空", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec("INSERT INTO master_entities (service, name) VALUES (?, ?)", service, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	masterID, _ := result.LastInsertId()

	if req.Content != "" {
		_, err = tx.Exec("INSERT INTO detail_records (master_id, service, content, version, change_note) VALUES (?, ?, ?, ?, ?)",
			masterID, service, req.Content, 1, "初始创建")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"master_id":  masterID,
	})
}

func handleUpdateMaster(w http.ResponseWriter, r *http.Request, service, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name       string `json:"name"`
		Content    string `json:"content"`
		ChangeNote string `json:"change_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if req.Name != "" {
		_, err = tx.Exec("UPDATE master_entities SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND service = ?", req.Name, id, service)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.Content != "" {
		var maxVersion int
		err = tx.QueryRow("SELECT COALESCE(MAX(version), 0) FROM detail_records WHERE master_id = ?", id).Scan(&maxVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		changeNote := req.ChangeNote
		if changeNote == "" {
			changeNote = fmt.Sprintf("更新到版本 %d", maxVersion+1)
		}

		_, err = tx.Exec("INSERT INTO detail_records (master_id, service, content, version, change_note) VALUES (?, ?, ?, ?, ?)",
			id, service, req.Content, maxVersion+1, changeNote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "更新成功",
	})
}

func handleDeleteMaster(w http.ResponseWriter, service, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM detail_records WHERE master_id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM master_entities WHERE id = ? AND service = ?", id, service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "删除成功",
	})
}


