package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, Response{Success: false, Error: message})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{Success: true, Data: data})
}

type CreateDeviceRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Location string `json:"location"`
}

func (s *Server) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var req CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "Name and code are required")
		return
	}

	result, err := s.db.db.Exec(`
		INSERT INTO devices (name, code, location) VALUES (?, ?, ?)
	`, req.Name, req.Code, req.Location)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create device")
		return
	}

	id, _ := result.LastInsertId()
	logOperation(fmt.Sprintf("Created device: %s (ID: %d)", req.Name, id))
	s.db.UpdateStatistics()
	writeSuccess(w, map[string]int64{"id": id})
}

func (s *Server) ListDevices(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, name, code, location, status, created_at FROM devices ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list devices")
		return
	}
	defer rows.Close()

	type Device struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Code      string `json:"code"`
		Location  string `json:"location"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}

	var devices []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.Location, &d.Status, &d.CreatedAt); err != nil {
			continue
		}
		devices = append(devices, d)
	}

	writeSuccess(w, devices)
}

type CreatePointRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (s *Server) CreateInspectionPoint(w http.ResponseWriter, r *http.Request) {
	var req CreatePointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "Name and code are required")
		return
	}

	result, err := s.db.db.Exec(`
		INSERT INTO inspection_points (name, code, description) VALUES (?, ?, ?)
	`, req.Name, req.Code, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create inspection point")
		return
	}

	id, _ := result.LastInsertId()
	logOperation(fmt.Sprintf("Created inspection point: %s (ID: %d)", req.Name, id))
	writeSuccess(w, map[string]int64{"id": id})
}

func (s *Server) ListInspectionPoints(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, name, code, description, created_at FROM inspection_points ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list points")
		return
	}
	defer rows.Close()

	type Point struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
	}

	var points []Point
	for rows.Next() {
		var p Point
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.Description, &p.CreatedAt); err != nil {
			continue
		}
		points = append(points, p)
	}

	writeSuccess(w, points)
}

type RoutePointItem struct {
	PointID  int64 `json:"point_id"`
	DeviceID int64 `json:"device_id,omitempty"`
}

type CreateRouteRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Points      []RoutePointItem  `json:"points"`
}

func (s *Server) CreateRoute(w http.ResponseWriter, r *http.Request) {
	var req CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO routes (name, description) VALUES (?, ?)
	`, req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create route")
		return
	}

	routeID, _ := result.LastInsertId()

	for i, point := range req.Points {
		var deviceID interface{}
		if point.DeviceID > 0 {
			deviceID = point.DeviceID
		}
		_, err := tx.Exec(`
			INSERT INTO route_points (route_id, point_id, order_num, device_id)
			VALUES (?, ?, ?, ?)
		`, routeID, point.PointID, i+1, deviceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to add route points")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	logOperation(fmt.Sprintf("Created route: %s (ID: %d) with %d points", req.Name, routeID, len(req.Points)))
	writeSuccess(w, map[string]int64{"id": routeID})
}

func (s *Server) ListRoutes(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, name, description, created_at FROM routes ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list routes")
		return
	}
	defer rows.Close()

	type RoutePointDetail struct {
		ID        int64  `json:"id"`
		OrderNum  int    `json:"order_num"`
		PointID   int64  `json:"point_id"`
		PointName string `json:"point_name"`
		DeviceID  int64  `json:"device_id,omitempty"`
		DeviceName string `json:"device_name,omitempty"`
	}

	type Route struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		Points      []RoutePointDetail `json:"points,omitempty"`
	}

	var routes []Route
	for rows.Next() {
		var route Route
		if err := rows.Scan(&route.ID, &route.Name, &route.Description, &route.CreatedAt); err != nil {
			continue
		}

		pointRows, err := s.db.db.Query(`
			SELECT rp.id, rp.order_num, rp.point_id, ip.name, rp.device_id, d.name
			FROM route_points rp
			LEFT JOIN inspection_points ip ON rp.point_id = ip.id
			LEFT JOIN devices d ON rp.device_id = d.id
			WHERE rp.route_id = ?
			ORDER BY rp.order_num
		`, route.ID)
		if err != nil {
			continue
		}

		for pointRows.Next() {
			var p RoutePointDetail
			var devID sql.NullInt64
			var devName sql.NullString
			if err := pointRows.Scan(&p.ID, &p.OrderNum, &p.PointID, &p.PointName, &devID, &devName); err != nil {
				continue
			}
			if devID.Valid {
				p.DeviceID = devID.Int64
				p.DeviceName = devName.String
			}
			route.Points = append(route.Points, p)
		}
		pointRows.Close()

		routes = append(routes, route)
	}

	writeSuccess(w, routes)
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Username == "" || req.Name == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "Username, name, and role are required")
		return
	}

	validRoles := map[string]bool{"inspector": true, "maintainer": true, "reviewer": true}
	if !validRoles[req.Role] {
		writeError(w, http.StatusBadRequest, "Invalid role. Must be: inspector, maintainer, reviewer")
		return
	}

	result, err := s.db.db.Exec(`
		INSERT INTO users (username, name, role) VALUES (?, ?, ?)
	`, req.Username, req.Name, req.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	id, _ := result.LastInsertId()
	logOperation(fmt.Sprintf("Created user: %s (ID: %d, Role: %s)", req.Username, id, req.Role))
	writeSuccess(w, map[string]int64{"id": id})
}

func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, username, name, role, created_at FROM users ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}
	defer rows.Close()

	type User struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}

	writeSuccess(w, users)
}

type CreateInspectionItemRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	StandardValue string `json:"standard_value"`
	MaxScore     int    `json:"max_score"`
}

func (s *Server) CreateInspectionItem(w http.ResponseWriter, r *http.Request) {
	var req CreateInspectionItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.MaxScore <= 0 {
		req.MaxScore = 100
	}

	result, err := s.db.db.Exec(`
		INSERT INTO inspection_items (name, description, standard_value, max_score)
		VALUES (?, ?, ?, ?)
	`, req.Name, req.Description, req.StandardValue, req.MaxScore)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create inspection item")
		return
	}

	id, _ := result.LastInsertId()
	logOperation(fmt.Sprintf("Created inspection item: %s (ID: %d)", req.Name, id))
	writeSuccess(w, map[string]int64{"id": id})
}

func (s *Server) ListInspectionItems(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, name, description, standard_value, max_score, created_at
		FROM inspection_items ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list items")
		return
	}
	defer rows.Close()

	type Item struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		StandardValue string `json:"standard_value"`
		MaxScore      int    `json:"max_score"`
		CreatedAt     string `json:"created_at"`
	}

	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.StandardValue, &i.MaxScore, &i.CreatedAt); err != nil {
			continue
		}
		items = append(items, i)
	}

	writeSuccess(w, items)
}

type CreateScheduleRequest struct {
	Name        string `json:"name"`
	RouteID     int64  `json:"route_id"`
	Frequency   string `json:"frequency"`
	StartDate   string `json:"start_date"`
	QuotaTotal  int    `json:"quota_total"`
	ItemIDs     []int64 `json:"item_ids"`
}

func (s *Server) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validFreq := map[string]bool{ScheduleFrequencyDaily: true, ScheduleFrequencyWeekly: true, ScheduleFrequencyMonthly: true}
	if !validFreq[req.Frequency] {
		writeError(w, http.StatusBadRequest, "Invalid frequency")
		return
	}

	if req.QuotaTotal <= 0 {
		req.QuotaTotal = 100
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var startDate time.Time
	if req.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid start_date format (use YYYY-MM-DD)")
			return
		}
	} else {
		startDate = time.Now()
	}

	result, err := tx.Exec(`
		INSERT INTO schedules (name, route_id, frequency, start_date, next_run_date, quota_total)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Name, req.RouteID, req.Frequency, startDate.Format("2006-01-02"), startDate.Format("2006-01-02"), req.QuotaTotal)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create schedule")
		return
	}

	scheduleID, _ := result.LastInsertId()

	if len(req.ItemIDs) > 0 {
		baseQuota := req.QuotaTotal / len(req.ItemIDs)
		remainder := req.QuotaTotal % len(req.ItemIDs)

		for i, itemID := range req.ItemIDs {
			quota := baseQuota
			if i < remainder {
				quota++
			}
			_, err := tx.Exec(`
				INSERT INTO schedule_item_quotas (schedule_id, item_id, quota)
				VALUES (?, ?, ?)
			`, scheduleID, itemID, quota)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to add item quotas")
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Created schedule: %s (ID: %d, Frequency: %s)", req.Name, scheduleID, req.Frequency))
	writeSuccess(w, map[string]int64{"id": scheduleID})
}

func (s *Server) ListSchedules(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, name, route_id, frequency, start_date, next_run_date, active, quota_total, created_at
		FROM schedules ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list schedules")
		return
	}
	defer rows.Close()

	type ScheduleItemQuota struct {
		ItemID   int64  `json:"item_id"`
		ItemName string `json:"item_name"`
		Quota    int    `json:"quota"`
	}

	type Schedule struct {
		ID          int64                `json:"id"`
		Name        string               `json:"name"`
		RouteID     int64                `json:"route_id"`
		Frequency   string               `json:"frequency"`
		StartDate   string               `json:"start_date"`
		NextRunDate string               `json:"next_run_date"`
		Active      bool                 `json:"active"`
		QuotaTotal  int                  `json:"quota_total"`
		CreatedAt   string               `json:"created_at"`
		Quotas      []ScheduleItemQuota  `json:"quotas,omitempty"`
	}

	var schedules []Schedule
	for rows.Next() {
		var sch Schedule
		if err := rows.Scan(&sch.ID, &sch.Name, &sch.RouteID, &sch.Frequency, &sch.StartDate,
			&sch.NextRunDate, &sch.Active, &sch.QuotaTotal, &sch.CreatedAt); err != nil {
			continue
		}

		quotaRows, err := s.db.db.Query(`
			SELECT siq.item_id, ii.name, siq.quota
			FROM schedule_item_quotas siq
			JOIN inspection_items ii ON siq.item_id = ii.id
			WHERE siq.schedule_id = ?
		`, sch.ID)
		if err == nil {
			for quotaRows.Next() {
				var q ScheduleItemQuota
				if err := quotaRows.Scan(&q.ItemID, &q.ItemName, &q.Quota); err != nil {
					continue
				}
				sch.Quotas = append(sch.Quotas, q)
			}
			quotaRows.Close()
		}

		schedules = append(schedules, sch)
	}

	writeSuccess(w, schedules)
}

type UpdateScheduleQuotaRequest struct {
	QuotaTotal int `json:"quota_total"`
}

func (s *Server) UpdateScheduleQuota(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	scheduleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	var req UpdateScheduleQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.QuotaTotal <= 0 {
		writeError(w, http.StatusBadRequest, "Quota total must be positive")
		return
	}

	_, err = s.db.db.Exec(`
		UPDATE schedules SET quota_total = ? WHERE id = ?
	`, req.QuotaTotal, scheduleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update quota")
		return
	}

	if err := s.db.RecalculateQuotas(scheduleID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to recalculate quotas")
		return
	}

	logOperation(fmt.Sprintf("Updated schedule quota: Schedule ID %d, New Total: %d", scheduleID, req.QuotaTotal))
	s.db.UpdateStatistics()
	writeSuccess(w, map[string]string{"message": "Quota updated and redistributed"})
}

type ClaimTaskRequest struct {
	InspectorID int64 `json:"inspector_id"`
}

func (s *Server) ClaimTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req ClaimTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var currentStatus string
	var existingInspector sql.NullInt64
	err = tx.QueryRow(`
		SELECT status, inspector_id FROM tasks WHERE id = ?
	`, taskID).Scan(&currentStatus, &existingInspector)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check task")
		return
	}

	if existingInspector.Valid {
		if existingInspector.Int64 == req.InspectorID {
			writeSuccess(w, map[string]string{"message": "Task already claimed by you"})
			return
		}
		writeError(w, http.StatusConflict, "Task already claimed by another inspector")
		return
	}

	if currentStatus != TaskStatusPending {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Task is not in pending state. Current status: %s", currentStatus))
		return
	}

	_, err = tx.Exec(`
		UPDATE tasks SET inspector_id = ?, status = ? WHERE id = ?
	`, req.InspectorID, TaskStatusInProgress, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to claim task")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Task %d claimed by inspector %d", taskID, req.InspectorID))
	s.db.UpdateStatistics()
	writeSuccess(w, map[string]string{"message": "Task claimed successfully"})
}

type CheckinRequest struct {
	RoutePointID int64 `json:"route_point_id"`
}

func (s *Server) Checkin(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req CheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var currentStatus string
	var routeID int64
	err = tx.QueryRow(`
		SELECT status, route_id FROM tasks WHERE id = ?
	`, taskID).Scan(&currentStatus, &routeID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check task")
		return
	}

	if currentStatus != TaskStatusInProgress {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Task is not in progress. Current status: %s", currentStatus))
		return
	}

	var targetOrder int
	err = tx.QueryRow(`
		SELECT order_num FROM route_points WHERE id = ? AND route_id = ?
	`, req.RoutePointID, routeID).Scan(&targetOrder)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusBadRequest, "Invalid route point for this task")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check route point")
		return
	}

	rows, err := tx.Query(`
		SELECT rp.order_num
		FROM task_checkins tc
		JOIN route_points rp ON tc.route_point_id = rp.id
		WHERE tc.task_id = ?
		ORDER BY rp.order_num DESC
	`, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check existing checkins")
		return
	}
	defer rows.Close()

	var lastOrder int
	if rows.Next() {
		rows.Scan(&lastOrder)
	}

	if targetOrder != lastOrder+1 {
		expectedPoint := lastOrder + 1
		var expectedPointID int64
		var expectedPointName string
		err = tx.QueryRow(`
			SELECT rp.id, ip.name
			FROM route_points rp
			JOIN inspection_points ip ON rp.point_id = ip.id
			WHERE rp.route_id = ? AND rp.order_num = ?
		`, routeID, expectedPoint).Scan(&expectedPointID, &expectedPointName)
		if err == nil {
			writeError(w, http.StatusBadRequest,
				fmt.Sprintf("Must checkin in order. Next expected: Order %d (Point: %s, ID: %d)",
					expectedPoint, expectedPointName, expectedPointID))
		} else {
			writeError(w, http.StatusBadRequest,
				fmt.Sprintf("Must checkin in order. Next expected: Order %d", expectedPoint))
		}
		return
	}

	_, err = tx.Exec(`
		INSERT INTO task_checkins (task_id, route_point_id) VALUES (?, ?)
	`, taskID, req.RoutePointID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to checkin")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Checkin successful for task %d, route point %d", taskID, req.RoutePointID))
	writeSuccess(w, map[string]string{"message": "Checkin successful"})
}

type TaskResultItem struct {
	ItemID      int64  `json:"item_id"`
	Score       int    `json:"score"`
	AbnormalDesc string `json:"abnormal_desc,omitempty"`
}

type SubmitTaskRequest struct {
	Results []TaskResultItem `json:"results"`
}

func (s *Server) SubmitTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var currentStatus string
	var routeID int64
	var scheduleID sql.NullInt64
	err = tx.QueryRow(`
		SELECT status, route_id, schedule_id FROM tasks WHERE id = ?
	`, taskID).Scan(&currentStatus, &routeID, &scheduleID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check task")
		return
	}

	if currentStatus != TaskStatusInProgress {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Task is not in progress. Current status: %s", currentStatus))
		return
	}

	var totalPoints int
	var checkedPoints int
	tx.QueryRow(`SELECT COUNT(*) FROM route_points WHERE route_id = ?`, routeID).Scan(&totalPoints)
	tx.QueryRow(`SELECT COUNT(*) FROM task_checkins WHERE task_id = ?`, taskID).Scan(&checkedPoints)
	if checkedPoints < totalPoints {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Must complete all checkins first. Completed: %d/%d", checkedPoints, totalPoints))
		return
	}

	totalScore := 0
	hasAbnormal := false
	for _, result := range req.Results {
		if result.Score < 0 || result.Score > 100 {
			writeError(w, http.StatusBadRequest, "Score must be between 0 and 100")
			return
		}
		totalScore += result.Score
		if result.AbnormalDesc != "" || result.Score < 80 {
			hasAbnormal = true
		}
	}

	averageScore := 0
	if len(req.Results) > 0 {
		averageScore = totalScore / len(req.Results)
	}

	isAbnormal := hasAbnormal || averageScore < 80

	for _, result := range req.Results {
		_, err := tx.Exec(`
			INSERT INTO task_results (task_id, item_id, score, abnormal_desc)
			VALUES (?, ?, ?, ?)
		`, taskID, result.ItemID, result.Score, result.AbnormalDesc)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save results")
			return
		}
	}

	newStatus := TaskStatusNormal
	if isAbnormal {
		newStatus = TaskStatusAbnormal
	}

	_, err = tx.Exec(`
		UPDATE tasks SET status = ?, total_score = ?, is_abnormal = ?, completed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, newStatus, averageScore, isAbnormal, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	if isAbnormal {
		deviceRows, err := tx.Query(`
			SELECT DISTINCT rp.device_id FROM route_points rp
			JOIN task_checkins tc ON rp.id = tc.route_point_id
			WHERE tc.task_id = ? AND rp.device_id IS NOT NULL
		`, taskID)
		if err == nil {
			defer deviceRows.Close()
			for deviceRows.Next() {
				var deviceID int64
				if err := deviceRows.Scan(&deviceID); err != nil {
					continue
				}
				_, err := tx.Exec(`
					INSERT INTO maintenance_records (task_id, device_id, status, issue_desc)
					VALUES (?, ?, ?, ?)
				`, taskID, deviceID, MaintenanceStatusPending, "Auto-created from abnormal inspection")
				if err != nil {
					continue
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Task %d submitted. Status: %s, Score: %d, Abnormal: %v",
		taskID, newStatus, averageScore, isAbnormal))
	s.db.UpdateStatistics()

	writeSuccess(w, map[string]interface{}{
		"message": "Task submitted successfully",
		"status":  newStatus,
		"score":   averageScore,
	})
}

type TransitionTaskRequest struct {
	TargetStatus string `json:"target_status"`
}

func (s *Server) TransitionTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req TransitionTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.QueryRow(`SELECT status FROM tasks WHERE id = ?`, taskID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check task")
		return
	}

	if !s.db.canTransition(currentStatus, req.TargetStatus) {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid transition. Current status: %s, cannot transition to %s",
				currentStatus, req.TargetStatus))
		return
	}

	_, err = tx.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, req.TargetStatus, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update status")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Task %d transitioned: %s -> %s", taskID, currentStatus, req.TargetStatus))
	s.db.UpdateStatistics()

	writeSuccess(w, map[string]string{
		"message":        "Status updated successfully",
		"current_status": req.TargetStatus,
	})
}

func (s *Server) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	type Task struct {
		ID          int64  `json:"id"`
		ScheduleID  int64  `json:"schedule_id,omitempty"`
		RouteID     int64  `json:"route_id"`
		Status      string `json:"status"`
		InspectorID int64  `json:"inspector_id,omitempty"`
		TotalScore  int    `json:"total_score"`
		IsAbnormal  bool   `json:"is_abnormal"`
		DueDate     string `json:"due_date,omitempty"`
		CompletedAt string `json:"completed_at,omitempty"`
		CreatedAt   string `json:"created_at"`
	}

	var task Task
	var schedID sql.NullInt64
	var inspID sql.NullInt64
	var dueDate sql.NullString
	var completedAt sql.NullString

	err = s.db.db.QueryRow(`
		SELECT id, schedule_id, route_id, status, inspector_id, total_score, is_abnormal,
		       due_date, completed_at, created_at
		FROM tasks WHERE id = ?
	`, taskID).Scan(&task.ID, &schedID, &task.RouteID, &task.Status, &inspID,
		&task.TotalScore, &task.IsAbnormal, &dueDate, &completedAt, &task.CreatedAt)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}

	if schedID.Valid {
		task.ScheduleID = schedID.Int64
	}
	if inspID.Valid {
		task.InspectorID = inspID.Int64
	}
	if dueDate.Valid {
		task.DueDate = dueDate.String
	}
	if completedAt.Valid {
		task.CompletedAt = completedAt.String
	}

	writeSuccess(w, task)
}

func (s *Server) ListTasks(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")

	query := `
		SELECT id, schedule_id, route_id, status, inspector_id, total_score, is_abnormal,
		       due_date, completed_at, created_at
		FROM tasks
	`
	args := []interface{}{}

	if statusFilter != "" {
		query += " WHERE status = ?"
		args = append(args, statusFilter)
	}
	query += " ORDER BY id DESC"

	rows, err := s.db.db.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list tasks")
		return
	}
	defer rows.Close()

	type Task struct {
		ID          int64  `json:"id"`
		ScheduleID  int64  `json:"schedule_id,omitempty"`
		RouteID     int64  `json:"route_id"`
		Status      string `json:"status"`
		InspectorID int64  `json:"inspector_id,omitempty"`
		TotalScore  int    `json:"total_score"`
		IsAbnormal  bool   `json:"is_abnormal"`
		DueDate     string `json:"due_date,omitempty"`
		CompletedAt string `json:"completed_at,omitempty"`
		CreatedAt   string `json:"created_at"`
	}

	var tasks []Task
	for rows.Next() {
		var task Task
		var schedID sql.NullInt64
		var inspID sql.NullInt64
		var dueDate sql.NullString
		var completedAt sql.NullString

		if err := rows.Scan(&task.ID, &schedID, &task.RouteID, &task.Status, &inspID,
			&task.TotalScore, &task.IsAbnormal, &dueDate, &completedAt, &task.CreatedAt); err != nil {
			continue
		}

		if schedID.Valid {
			task.ScheduleID = schedID.Int64
		}
		if inspID.Valid {
			task.InspectorID = inspID.Int64
		}
		if dueDate.Valid {
			task.DueDate = dueDate.String
		}
		if completedAt.Valid {
			task.CompletedAt = completedAt.String
		}

		tasks = append(tasks, task)
	}

	writeSuccess(w, tasks)
}

type AssignMaintenanceRequest struct {
	MaintainerID int64 `json:"maintainer_id"`
}

func (s *Server) AssignMaintenance(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	maintenanceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid maintenance ID")
		return
	}

	var req AssignMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err = s.db.db.Exec(`
		UPDATE maintenance_records SET maintainer_id = ?, status = ? WHERE id = ?
	`, req.MaintainerID, MaintenanceStatusInProgress, maintenanceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to assign maintenance")
		return
	}

	logOperation(fmt.Sprintf("Maintenance %d assigned to maintainer %d", maintenanceID, req.MaintainerID))
	writeSuccess(w, map[string]string{"message": "Maintenance assigned successfully"})
}

type CompleteMaintenanceRequest struct {
	Solution string `json:"solution"`
}

func (s *Server) CompleteMaintenance(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	maintenanceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid maintenance ID")
		return
	}

	var req CompleteMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var taskID int64
	err = tx.QueryRow(`
		UPDATE maintenance_records SET status = ?, solution = ?, completed_at = CURRENT_TIMESTAMP
		WHERE id = ? RETURNING task_id
	`, MaintenanceStatusCompleted, req.Solution, maintenanceID).Scan(&taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to complete maintenance")
		return
	}

	_, err = tx.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, TaskStatusInProgress, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update task status")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Maintenance %d completed for task %d", maintenanceID, taskID))
	s.db.UpdateStatistics()

	writeSuccess(w, map[string]string{"message": "Maintenance completed, task reopened for re-inspection"})
}

func (s *Server) ListMaintenanceRecords(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, task_id, device_id, maintainer_id, status, issue_desc, solution,
		       completed_at, created_at
		FROM maintenance_records ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list maintenance records")
		return
	}
	defer rows.Close()

	type MaintenanceRecord struct {
		ID           int64  `json:"id"`
		TaskID       int64  `json:"task_id"`
		DeviceID     int64  `json:"device_id"`
		MaintainerID int64  `json:"maintainer_id,omitempty"`
		Status       string `json:"status"`
		IssueDesc    string `json:"issue_desc"`
		Solution     string `json:"solution,omitempty"`
		CompletedAt  string `json:"completed_at,omitempty"`
		CreatedAt    string `json:"created_at"`
	}

	var records []MaintenanceRecord
	for rows.Next() {
		var r MaintenanceRecord
		var maintID sql.NullInt64
		var solution sql.NullString
		var completedAt sql.NullString

		if err := rows.Scan(&r.ID, &r.TaskID, &r.DeviceID, &maintID, &r.Status,
			&r.IssueDesc, &solution, &completedAt, &r.CreatedAt); err != nil {
			continue
		}

		if maintID.Valid {
			r.MaintainerID = maintID.Int64
		}
		if solution.Valid {
			r.Solution = solution.String
		}
		if completedAt.Valid {
			r.CompletedAt = completedAt.String
		}

		records = append(records, r)
	}

	writeSuccess(w, records)
}

type ReviewTaskRequest struct {
	ReviewerID int64  `json:"reviewer_id"`
	Result     string `json:"result"`
	Reason     string `json:"reason,omitempty"`
}

func (s *Server) ReviewTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req ReviewTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Result != "pass" && req.Result != "fail" {
		writeError(w, http.StatusBadRequest, "Result must be 'pass' or 'fail'")
		return
	}

	if req.Result == "fail" && req.Reason == "" {
		writeError(w, http.StatusBadRequest, "Reason is required when review fails")
		return
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.QueryRow(`SELECT status FROM tasks WHERE id = ?`, taskID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check task")
		return
	}

	if currentStatus != TaskStatusUnderReview {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Task is not under review. Current status: %s", currentStatus))
		return
	}

	_, err = tx.Exec(`
		INSERT INTO review_records (task_id, reviewer_id, result, reason)
		VALUES (?, ?, ?, ?)
	`, taskID, req.ReviewerID, req.Result, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save review")
		return
	}

	var newStatus string
	if req.Result == "pass" {
		newStatus = TaskStatusReviewPass
	} else {
		newStatus = TaskStatusReviewFail
	}

	_, err = tx.Exec(`
		UPDATE tasks SET status = ?, reviewer_id = ? WHERE id = ?
	`, newStatus, req.ReviewerID, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit")
		return
	}

	logOperation(fmt.Sprintf("Task %d reviewed: %s -> %s", taskID, currentStatus, newStatus))
	s.db.UpdateStatistics()

	writeSuccess(w, map[string]string{
		"message":        "Review completed",
		"current_status": newStatus,
	})
}

func (s *Server) ListReviewRecords(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT id, task_id, reviewer_id, result, reason, created_at
		FROM review_records ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list review records")
		return
	}
	defer rows.Close()

	type ReviewRecord struct {
		ID         int64  `json:"id"`
		TaskID     int64  `json:"task_id"`
		ReviewerID int64  `json:"reviewer_id"`
		Result     string `json:"result"`
		Reason     string `json:"reason,omitempty"`
		CreatedAt  string `json:"created_at"`
	}

	var records []ReviewRecord
	for rows.Next() {
		var r ReviewRecord
		if err := rows.Scan(&r.ID, &r.TaskID, &r.ReviewerID, &r.Result, &r.Reason, &r.CreatedAt); err != nil {
			continue
		}
		records = append(records, r)
	}

	writeSuccess(w, records)
}

func (s *Server) GetStatistics(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.db.Query(`
		SELECT date, total_tasks, completed_tasks, abnormal_tasks, maintenance_tasks, created_at
		FROM statistics ORDER BY date DESC LIMIT 30
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}
	defer rows.Close()

	type Stat struct {
		Date            string `json:"date"`
		TotalTasks      int    `json:"total_tasks"`
		CompletedTasks  int    `json:"completed_tasks"`
		AbnormalTasks   int    `json:"abnormal_tasks"`
		MaintenanceTasks int   `json:"maintenance_tasks"`
		CreatedAt       string `json:"created_at"`
	}

	var stats []Stat
	for rows.Next() {
		var s Stat
		if err := rows.Scan(&s.Date, &s.TotalTasks, &s.CompletedTasks,
			&s.AbnormalTasks, &s.MaintenanceTasks, &s.CreatedAt); err != nil {
			continue
		}
		stats = append(stats, s)
	}

	writeSuccess(w, stats)
}

func (s *Server) RunScheduler(w http.ResponseWriter, r *http.Request) {
	err := s.RunScheduledTasks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to run scheduler")
		return
	}
	writeSuccess(w, map[string]string{"message": "Scheduler executed successfully"})
}

func (s *Server) CheckReminders(w http.ResponseWriter, r *http.Request) {
	reminders, err := s.CheckOverdueTasks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check reminders")
		return
	}
	writeSuccess(w, reminders)
}
