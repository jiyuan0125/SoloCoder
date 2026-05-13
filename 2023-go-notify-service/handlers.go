package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type CreateDepartmentRequest struct {
	Name string `json:"name"`
}

func CreateDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "Department name is required")
		return
	}

	result, err := db.Exec(`INSERT INTO departments (name) VALUES (?)`, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeError(w, http.StatusConflict, "Department already exists")
			return
		}
		log.Printf("Failed to create department: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create department")
		return
	}

	id, _ := result.LastInsertId()
	dept := Department{ID: id, Name: req.Name}
	writeJSON(w, http.StatusCreated, dept)
}

func ListDepartmentsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, created_at FROM departments ORDER BY id`)
	if err != nil {
		log.Printf("Failed to list departments: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to list departments")
		return
	}
	defer rows.Close()

	var depts []Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.Name, &d.CreatedAt); err != nil {
			log.Printf("Failed to scan department: %v", err)
			writeError(w, http.StatusInternalServerError, "Failed to list departments")
			return
		}
		depts = append(depts, d)
	}

	writeJSON(w, http.StatusOK, depts)
}

type CreateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	DepartmentID *int64 `json:"department_id,omitempty"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Phone) == "" {
		writeError(w, http.StatusBadRequest, "Username, email, and phone are required")
		return
	}

	if req.DepartmentID != nil {
		exists, err := departmentExists(*req.DepartmentID)
		if err != nil {
			log.Printf("Failed to check department: %v", err)
			writeError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
		if !exists {
			writeError(w, http.StatusBadRequest, "Department does not exist")
			return
		}
	}

	result, err := db.Exec(
		`INSERT INTO users (username, email, phone, department_id) VALUES (?, ?, ?, ?)`,
		req.Username, req.Email, req.Phone, req.DepartmentID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeError(w, http.StatusConflict, "Username already exists")
			return
		}
		log.Printf("Failed to create user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	id, _ := result.LastInsertId()
	user := User{
		ID:           id,
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		DepartmentID: req.DepartmentID,
	}
	writeJSON(w, http.StatusCreated, user)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := getUserByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Failed to get user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	deptIDStr := r.URL.Query().Get("department_id")
	var rows *sql.Rows
	var err error

	if deptIDStr != "" {
		deptID, parseErr := strconv.ParseInt(deptIDStr, 10, 64)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "Invalid department ID")
			return
		}
		rows, err = db.Query(
			`SELECT id, username, email, phone, department_id, created_at FROM users WHERE department_id = ? ORDER BY id`,
			deptID,
		)
	} else {
		rows, err = db.Query(`SELECT id, username, email, phone, department_id, created_at FROM users ORDER BY id`)
	}

	if err != nil {
		log.Printf("Failed to list users: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var deptID sql.NullInt64
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &deptID, &u.CreatedAt); err != nil {
			log.Printf("Failed to scan user: %v", err)
			writeError(w, http.StatusInternalServerError, "Failed to list users")
			return
		}
		if deptID.Valid {
			id := deptID.Int64
			u.DepartmentID = &id
		}
		users = append(users, u)
	}

	writeJSON(w, http.StatusOK, users)
}

type SetPreferenceRequest struct {
	UserID      int64  `json:"user_id"`
	MessageType string `json:"message_type"`
	Channel     string `json:"channel"`
}

func SetPreferenceHandler(w http.ResponseWriter, r *http.Request) {
	var req SetPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !ValidChannel(req.Channel) {
		writeError(w, http.StatusBadRequest, "Unsupported channel type")
		return
	}

	if strings.TrimSpace(req.MessageType) == "" {
		writeError(w, http.StatusBadRequest, "Message type is required")
		return
	}

	exists, err := userExists(req.UserID)
	if err != nil {
		log.Printf("Failed to check user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to set preference")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	_, err = db.Exec(
		`INSERT INTO notification_preferences (user_id, message_type, channel)
		 VALUES (?, ?, ?)
		 ON CONFLICT(user_id, message_type) DO UPDATE SET channel = excluded.channel`,
		req.UserID, req.MessageType, req.Channel,
	)
	if err != nil {
		log.Printf("Failed to set preference: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to set preference")
		return
	}

	pref := NotificationPreference{
		UserID:      req.UserID,
		MessageType: req.MessageType,
		Channel:     req.Channel,
	}
	writeJSON(w, http.StatusOK, pref)
}

func GetPreferenceHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userIDStr := query.Get("user_id")
	messageType := query.Get("message_type")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if strings.TrimSpace(messageType) == "" {
		writeError(w, http.StatusBadRequest, "Message type is required")
		return
	}

	exists, err := userExists(userID)
	if err != nil {
		log.Printf("Failed to check user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to get preference")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	var pref NotificationPreference
	err = db.QueryRow(
		`SELECT id, user_id, message_type, channel FROM notification_preferences
		 WHERE user_id = ? AND message_type = ?`,
		userID, messageType,
	).Scan(&pref.ID, &pref.UserID, &pref.MessageType, &pref.Channel)

	if err == sql.ErrNoRows {
		pref = NotificationPreference{
			UserID:      userID,
			MessageType: messageType,
			Channel:     ChannelSite.String(),
		}
	} else if err != nil {
		log.Printf("Failed to get preference: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to get preference")
		return
	}

	writeJSON(w, http.StatusOK, pref)
}

func SendNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var req SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID <= 0 {
		writeError(w, http.StatusBadRequest, "Valid user ID is required")
		return
	}

	if strings.TrimSpace(req.MessageType) == "" || strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "Message type and content are required")
		return
	}

	exists, err := userExists(req.UserID)
	if err != nil {
		log.Printf("Failed to check user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to send notification")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	record, err := SendNotification(req.UserID, req.MessageType, req.Content)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to send notification")
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func BatchNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var req BatchNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DepartmentID <= 0 {
		writeError(w, http.StatusBadRequest, "Valid department ID is required")
		return
	}

	if strings.TrimSpace(req.MessageType) == "" || strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "Message type and content are required")
		return
	}

	exists, err := departmentExists(req.DepartmentID)
	if err != nil {
		log.Printf("Failed to check department: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to send batch notification")
		return
	}
	if !exists {
		writeError(w, http.StatusBadRequest, "Department does not exist")
		return
	}

	users, err := getUsersByDepartment(req.DepartmentID)
	if err != nil {
		log.Printf("Failed to get users by department: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to send batch notification")
		return
	}

	var records []*NotificationRecord
	for _, user := range users {
		record, err := SendNotification(user.ID, req.MessageType, req.Content)
		if err != nil {
			log.Printf("Failed to send to user %d: %v", user.ID, err)
			continue
		}
		records = append(records, record)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sent_count":   len(records),
		"total_users":  len(users),
		"notifications": records,
	})
}

func ListNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	statusFilter := r.URL.Query().Get("status")

	var query string
	var args []interface{}

	query = `SELECT id, user_id, message_type, content, channel, status, retry_count, 
	              next_retry_at, created_at, updated_at, degraded_from
	         FROM notifications WHERE 1=1`

	if userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if statusFilter != "" {
		query += " AND status = ?"
		args = append(args, statusFilter)
	}

	query += " ORDER BY id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Failed to list notifications: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to list notifications")
		return
	}
	defer rows.Close()

	var records []NotificationRecord
	for rows.Next() {
		var r NotificationRecord
		var nextRetry, degradedFrom sql.NullString
		if err := rows.Scan(&r.ID, &r.UserID, &r.MessageType, &r.Content, &r.Channel,
			&r.Status, &r.RetryCount, &nextRetry, &r.CreatedAt, &r.UpdatedAt, &degradedFrom); err != nil {
			log.Printf("Failed to scan notification: %v", err)
			writeError(w, http.StatusInternalServerError, "Failed to list notifications")
			return
		}
		if nextRetry.Valid {
			val := nextRetry.String
			r.NextRetryAt = &val
		}
		if degradedFrom.Valid {
			val := degradedFrom.String
			r.DegradedFrom = &val
		}
		records = append(records, r)
	}

	writeJSON(w, http.StatusOK, records)
}

func ManualRetryHandler(w http.ResponseWriter, r *http.Request) {
	if err := RetryPendingNotifications(); err != nil {
		log.Printf("Manual retry failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Retry failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Retry triggered"})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/departments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateDepartmentHandler(w, r)
		case http.MethodGet:
			ListDepartmentsHandler(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateUserHandler(w, r)
		case http.MethodGet:
			ListUsersHandler(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetUserHandler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/preferences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			SetPreferenceHandler(w, r)
		case http.MethodGet:
			GetPreferenceHandler(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/notifications/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			SendNotificationHandler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/notifications/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			BatchNotificationHandler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ListNotificationsHandler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
	mux.HandleFunc("/notifications/retry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ManualRetryHandler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("REQUEST: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("RESPONSE: %s %s took %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func StartServer(port int) error {
	mux := http.NewServeMux()
	SetupRoutes(mux)

	handler := loggingMiddleware(mux)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, handler)
}
