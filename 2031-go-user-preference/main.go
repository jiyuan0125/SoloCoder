package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 数据库连接
var db *sql.DB

// 客户端类型
const (
	ClientWeb     = "web"
	ClientApp     = "app"
	ClientMiniApp = "miniapp"
)

// 类别
const (
	CategoryNotification = "notification"
	CategoryDisplay      = "display"
	CategoryPrivacy      = "privacy"
)

// 偏好项定义
type PreferenceDefinition struct {
	Key          string      `json:"key"`
	Category     string      `json:"category"`
	DefaultValue interface{} `json:"default_value"`
	ValueType    string      `json:"value_type"`
	Description  string      `json:"description"`
}

// 所有偏好项定义
var preferenceDefinitions = []PreferenceDefinition{
	// 通知设置
	{"email_notification", CategoryNotification, true, "boolean", "是否接收邮件通知"},
	{"push_notification", CategoryNotification, true, "boolean", "是否接收推送通知"},
	{"sms_notification", CategoryNotification, false, "boolean", "是否接收短信通知"},
	// 显示设置
	{"theme", CategoryDisplay, "light", "string", "主题颜色（light/dark）"},
	{"language", CategoryDisplay, "zh-CN", "string", "语言设置"},
	{"font_size", CategoryDisplay, "medium", "string", "字体大小（small/medium/large）"},
	// 隐私设置
	{"show_online_status", CategoryPrivacy, true, "boolean", "是否显示在线状态"},
	{"allow_data_collection", CategoryPrivacy, true, "boolean", "是否允许数据收集"},
	{"show_last_seen", CategoryPrivacy, false, "boolean", "是否显示最后在线时间"},
}

// 用户设置
type UserPreference struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// 客户端覆盖
type ClientOverride struct {
	Client string `json:"client"`
	Key    string `json:"key"`
	Value  interface{} `json:"value"`
}

// 生效的偏好值
type EffectivePreference struct {
	Key         string      `json:"key"`
	Category    string      `json:"category"`
	Value       interface{} `json:"value"`
	Source      string      `json:"source"`
	Description string      `json:"description"`
}

// 历史记录
type HistoryRecord struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	Client    string          `json:"client"`
	Action    string          `json:"action"`
	Timestamp time.Time       `json:"timestamp"`
	Changes   json.RawMessage `json:"changes"`
}

// 响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./user_preference.db")
	if err != nil {
		log.Fatal("无法打开数据库:", err)
	}
	defer db.Close()

	if err = initDatabase(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	if err = insertDefaultPreferences(); err != nil {
		log.Fatal("插入默认偏好失败:", err)
	}

	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUser)
	http.HandleFunc("/preferences", handlePreferences)
	http.HandleFunc("/preferences/defaults", handleDefaults)
	http.HandleFunc("/history", handleHistory)
	http.HandleFunc("/restore", handleRestore)

	log.Println("服务器启动在端口 8700...")
	log.Fatal(http.ListenAndServe(":9901", nil))
}

func initDatabase() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS preference_definitions (
			key TEXT PRIMARY KEY,
			category TEXT NOT NULL,
			default_value TEXT NOT NULL,
			value_type TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_preferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (key) REFERENCES preference_definitions(key),
			UNIQUE(user_id, key)
		)`,
		`CREATE TABLE IF NOT EXISTS client_overrides (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			client TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (key) REFERENCES preference_definitions(key),
			UNIQUE(user_id, client, key)
		)`,
		`CREATE TABLE IF NOT EXISTS preference_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			client TEXT DEFAULT '',
			action TEXT NOT NULL,
			changes TEXT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return err
		}
	}
	return nil
}

func insertDefaultPreferences() error {
	for _, def := range preferenceDefinitions {
		defaultValue, _ := json.Marshal(def.DefaultValue)
		_, err := db.Exec(`INSERT OR IGNORE INTO preference_definitions (key, category, default_value, value_type, description)
			VALUES (?, ?, ?, ?, ?)`,
			def.Key, def.Category, string(defaultValue), def.ValueType, def.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "无效的请求格式")
			return
		}

		if req.Username == "" {
			writeError(w, http.StatusBadRequest, "用户名不能为空")
			return
		}

		result, err := db.Exec(`INSERT INTO users (username) VALUES (?)`, req.Username)
		if err != nil {
			writeError(w, http.StatusConflict, "用户已存在")
			return
		}

		id, _ := result.LastInsertId()
		writeSuccess(w, map[string]interface{}{
			"id":       id,
			"username": req.Username,
		})
		return
	}

	if r.Method == http.MethodGet {
		rows, err := db.Query(`SELECT id, username FROM users`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "查询失败")
			return
		}
		defer rows.Close()

		users := []map[string]interface{}{}
		for rows.Next() {
			var id int64
			var username string
			rows.Scan(&id, &username)
			users = append(users, map[string]interface{}{
				"id":       id,
				"username": username,
			})
		}
		writeSuccess(w, users)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "无效的路径")
		return
	}

	userID := parts[0]
	if userID == "" {
		writeError(w, http.StatusNotFound, "用户ID不能为空")
		return
	}

	// 检查用户是否存在
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, userID).Scan(&exists)
	if err != nil || !exists {
		writeError(w, http.StatusNotFound, "用户不存在")
		return
	}

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			getUserInfo(w, r, userID)
		} else if r.Method == http.MethodDelete {
			deleteUser(w, r, userID)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		}
		return
	}

	// /users/{id}/preferences
	if len(parts) >= 2 && parts[1] == "preferences" {
		client := r.URL.Query().Get("client")
		if len(parts) == 3 {
			// /users/{id}/preferences/{key}
			key := parts[2]
			if r.Method == http.MethodGet {
				getUserPreference(w, r, userID, key, client)
			} else if r.Method == http.MethodPut {
				updateUserPreference(w, r, userID, key, client)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
			}
		} else {
			if r.Method == http.MethodGet {
				getUserPreferences(w, r, userID, client)
			} else if r.Method == http.MethodPatch {
				batchUpdatePreferences(w, r, userID, client)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
			}
		}
		return
	}

	// /users/{id}/history
	if len(parts) >= 2 && parts[1] == "history" {
		if r.Method == http.MethodGet {
			getUserHistory(w, r, userID)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		}
		return
	}

	writeError(w, http.StatusNotFound, "无效的路径")
}

func getUserInfo(w http.ResponseWriter, r *http.Request, userID string) {
	var username string
	var createdAt string
	err := db.QueryRow(`SELECT username, created_at FROM users WHERE id = ?`, userID).Scan(&username, &createdAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "用户不存在")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"id":         userID,
		"username":   username,
		"created_at": createdAt,
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request, userID string) {
	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM client_overrides WHERE user_id = ?`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	_, err = tx.Exec(`DELETE FROM user_preferences WHERE user_id = ?`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	_, err = tx.Exec(`DELETE FROM preference_history WHERE user_id = ?`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	_, err = tx.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}

	writeSuccess(w, map[string]string{"message": "用户已删除"})
}

func getUserPreferences(w http.ResponseWriter, r *http.Request, userID string, client string) {
	preferences, err := calculateEffectivePreferences(userID, client)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取偏好失败")
		return
	}

	writeSuccess(w, preferences)
}

func getUserPreference(w http.ResponseWriter, r *http.Request, userID string, key string, client string) {
	preferences, err := calculateEffectivePreferences(userID, client)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取偏好失败")
		return
	}

	for _, p := range preferences {
		if p.Key == key {
			writeSuccess(w, p)
			return
		}
	}

	writeError(w, http.StatusNotFound, "偏好项不存在")
}

func updateUserPreference(w http.ResponseWriter, r *http.Request, userID string, key string, client string) {
	var req struct {
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	def, err := getPreferenceDefinition(key)
	if err != nil {
		writeError(w, http.StatusNotFound, "偏好项不存在")
		return
	}

	if err := validateValueType(req.Value, def.ValueType); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	valueJSON, _ := json.Marshal(req.Value)
	var changes []UserPreference
	changes = append(changes, UserPreference{Key: key, Value: req.Value})

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}
	defer tx.Rollback()

	var table string
	if client != "" {
		table = "client_overrides"
		_, err = tx.Exec(`INSERT OR REPLACE INTO client_overrides (user_id, client, key, value) VALUES (?, ?, ?, ?)`,
			userID, client, key, string(valueJSON))
	} else {
		table = "user_preferences"
		_, err = tx.Exec(`INSERT OR REPLACE INTO user_preferences (user_id, key, value) VALUES (?, ?, ?)`,
			userID, key, string(valueJSON))
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	changesJSON, _ := json.Marshal(changes)
	action := "update"
	if client != "" {
		action = fmt.Sprintf("update_override_%s", client)
	}
	_, err = tx.Exec(`INSERT INTO preference_history (user_id, client, action, changes) VALUES (?, ?, ?, ?)`,
		userID, client, action, string(changesJSON))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"key":    key,
		"value":  req.Value,
		"source": getSourceForUpdate(table, client),
	})
}

func batchUpdatePreferences(w http.ResponseWriter, r *http.Request, userID string, client string) {
	var req struct {
		Preferences []UserPreference `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	if len(req.Preferences) == 0 {
		writeError(w, http.StatusBadRequest, "未提供任何偏好项")
		return
	}

	for _, p := range req.Preferences {
		def, err := getPreferenceDefinition(p.Key)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("偏好项 %s 不存在", p.Key))
			return
		}

		if err := validateValueType(p.Value, def.ValueType); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}
	defer tx.Rollback()

	for _, p := range req.Preferences {
		valueJSON, _ := json.Marshal(p.Value)
		if client != "" {
			_, err = tx.Exec(`INSERT OR REPLACE INTO client_overrides (user_id, client, key, value) VALUES (?, ?, ?, ?)`,
				userID, client, p.Key, string(valueJSON))
		} else {
			_, err = tx.Exec(`INSERT OR REPLACE INTO user_preferences (user_id, key, value) VALUES (?, ?, ?)`,
				userID, p.Key, string(valueJSON))
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "更新失败")
			return
		}
	}

	changesJSON, _ := json.Marshal(req.Preferences)
	action := "batch_update"
	if client != "" {
		action = fmt.Sprintf("batch_update_override_%s", client)
	}
	_, err = tx.Exec(`INSERT INTO preference_history (user_id, client, action, changes) VALUES (?, ?, ?, ?)`,
		userID, client, action, string(changesJSON))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"updated": len(req.Preferences),
		"client":  client,
	})
}

func handlePreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		defs := []map[string]interface{}{}
		for _, def := range preferenceDefinitions {
			defs = append(defs, map[string]interface{}{
				"key":           def.Key,
				"category":      def.Category,
				"default_value": def.DefaultValue,
				"value_type":    def.ValueType,
				"description":   def.Description,
			})
		}
		writeSuccess(w, defs)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
}

func handleDefaults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		defaults := map[string]interface{}{}
		for _, def := range preferenceDefinitions {
			defaults[def.Key] = map[string]interface{}{
				"value":       def.DefaultValue,
				"category":    def.Category,
				"description": def.Description,
			}
		}
		writeSuccess(w, defaults)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeError(w, http.StatusMethodNotAllowed, "请使用 /users/{id}/history 访问")
}

func getUserHistory(w http.ResponseWriter, r *http.Request, userID string) {
	client := r.URL.Query().Get("client")

	query := `SELECT id, user_id, client, action, timestamp, changes FROM preference_history WHERE user_id = ?`
	var args []interface{}
	args = append(args, userID)

	if client != "" {
		query += ` AND client = ?`
		args = append(args, client)
	}
	query += ` ORDER BY timestamp DESC LIMIT 100`

	rows, err := db.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询历史失败")
		return
	}
	defer rows.Close()

	records := []HistoryRecord{}
	for rows.Next() {
		var record HistoryRecord
		var clientVal sql.NullString
		rows.Scan(&record.ID, &record.UserID, &clientVal, &record.Action, &record.Timestamp, &record.Changes)
		if clientVal.Valid {
			record.Client = clientVal.String
		}
		records = append(records, record)
	}

	writeSuccess(w, records)
}

func handleRestore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req struct {
		UserID    string `json:"user_id"`
		HistoryID int64  `json:"history_id"`
		Client    string `json:"client,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	if req.UserID == "" || req.HistoryID == 0 {
		writeError(w, http.StatusBadRequest, "缺少必要参数")
		return
	}

	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, req.UserID).Scan(&exists)
	if err != nil || !exists {
		writeError(w, http.StatusNotFound, "用户不存在")
		return
	}

	var action string
	var changesJSON string
	var timestamp time.Time
	err = db.QueryRow(`SELECT action, changes, timestamp FROM preference_history WHERE id = ? AND user_id = ?`,
		req.HistoryID, req.UserID).Scan(&action, &changesJSON, &timestamp)
	if err != nil {
		writeError(w, http.StatusNotFound, "历史记录不存在")
		return
	}

	var preferences []UserPreference
	if err := json.Unmarshal([]byte(changesJSON), &preferences); err != nil {
		writeError(w, http.StatusInternalServerError, "恢复失败")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "恢复失败")
		return
	}
	defer tx.Rollback()

	for _, p := range preferences {
		def, err := getPreferenceDefinition(p.Key)
		if err != nil {
			continue
		}
		if err := validateValueType(p.Value, def.ValueType); err != nil {
			continue
		}

		valueJSON, _ := json.Marshal(p.Value)
		if req.Client != "" {
			_, err = tx.Exec(`INSERT OR REPLACE INTO client_overrides (user_id, client, key, value) VALUES (?, ?, ?, ?)`,
				req.UserID, req.Client, p.Key, string(valueJSON))
		} else {
			_, err = tx.Exec(`INSERT OR REPLACE INTO user_preferences (user_id, key, value) VALUES (?, ?, ?)`,
				req.UserID, p.Key, string(valueJSON))
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "恢复失败")
			return
		}
	}

	restoreAction := fmt.Sprintf("restore_to_%s", timestamp.Format("2006-01-02_15:04:05"))
	_, err = tx.Exec(`INSERT INTO preference_history (user_id, client, action, changes) VALUES (?, ?, ?, ?)`,
		req.UserID, req.Client, restoreAction, changesJSON)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "恢复失败")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "恢复失败")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"message":    fmt.Sprintf("已恢复到 %s 的设置", timestamp.Format("2006-01-02 15:04:05")),
		"restored":   len(preferences),
		"history_id": req.HistoryID,
	})
}

func calculateEffectivePreferences(userID string, client string) ([]EffectivePreference, error) {
	userPrefs := make(map[string]string)
	clientPrefs := make(map[string]string)

	rows, err := db.Query(`SELECT key, value FROM user_preferences WHERE user_id = ?`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var key, value string
			rows.Scan(&key, &value)
			userPrefs[key] = value
		}
	}

	if client != "" {
		rows, err = db.Query(`SELECT key, value FROM client_overrides WHERE user_id = ? AND client = ?`, userID, client)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var key, value string
				rows.Scan(&key, &value)
				clientPrefs[key] = value
			}
		}
	}

	result := []EffectivePreference{}
	for _, def := range preferenceDefinitions {
		var value interface{}
		var source string

		if clientValue, ok := clientPrefs[def.Key]; ok {
			json.Unmarshal([]byte(clientValue), &value)
			source = fmt.Sprintf("client_%s", client)
		} else if userValue, ok := userPrefs[def.Key]; ok {
			json.Unmarshal([]byte(userValue), &value)
			source = "user"
		} else {
			value = def.DefaultValue
			source = "default"
		}

		result = append(result, EffectivePreference{
			Key:         def.Key,
			Category:    def.Category,
			Value:       value,
			Source:      source,
			Description: def.Description,
		})
	}

	return result, nil
}

func getPreferenceDefinition(key string) (*PreferenceDefinition, error) {
	for _, def := range preferenceDefinitions {
		if def.Key == key {
			return &def, nil
		}
	}
	return nil, fmt.Errorf("偏好项不存在")
}

func validateValueType(value interface{}, expectedType string) error {
	switch expectedType {
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("值类型错误，期望 boolean")
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("值类型错误，期望 string")
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("值类型错误，期望 number")
		}
	}
	return nil
}

func getSourceForUpdate(table string, client string) string {
	if client != "" {
		return fmt.Sprintf("client_%s", client)
	}
	return "user"
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}
