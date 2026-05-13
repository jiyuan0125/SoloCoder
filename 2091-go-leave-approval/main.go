package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type LeaveType string

const (
	LeaveTypeAnnual LeaveType = "annual"
	LeaveTypeSick   LeaveType = "sick"
	LeaveTypePersonal LeaveType = "personal"
	LeaveTypeCompensatory LeaveType = "compensatory"
)

type LeaveStatus string

const (
	LeaveStatusPending    LeaveStatus = "pending"
	LeaveStatusApproved   LeaveStatus = "approved"
	LeaveStatusRejected   LeaveStatus = "rejected"
	LeaveStatusRevoked    LeaveStatus = "revoked"
	LeaveStatusClosed     LeaveStatus = "closed"
)

type User struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Department    string `json:"department"`
	HireDate      string `json:"hire_date"`
	ManagerID     *int64 `json:"manager_id,omitempty"`
	HRID          *int64 `json:"hr_id,omitempty"`
	DeptManagerID *int64 `json:"dept_manager_id,omitempty"`
}

type LeaveRequest struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	LeaveType       LeaveType   `json:"leave_type"`
	StartDate       string      `json:"start_date"`
	EndDate         string      `json:"end_date"`
	ActualStartDate string      `json:"actual_start_date,omitempty"`
	ActualEndDate   string      `json:"actual_end_date,omitempty"`
	Days            int         `json:"days"`
	ActualDays      int         `json:"actual_days,omitempty"`
	Reason          string      `json:"reason"`
	SickNote        string      `json:"sick_note,omitempty"`
	Status          LeaveStatus `json:"status"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
}

type ApprovalAction struct {
	ID            int64       `json:"id"`
	LeaveID       int64       `json:"leave_id"`
	ApproverID    int64       `json:"approver_id"`
	ApproverName  string      `json:"approver_name,omitempty"`
	Action        string      `json:"action"`
	Reason        string      `json:"reason,omitempty"`
	CreatedAt     string      `json:"created_at"`
}

type ActionComment struct {
	ID         int64  `json:"id"`
	ActionID   int64  `json:"action_id"`
	CommenterID int64 `json:"commenter_id"`
	Comment    string `json:"comment"`
	CreatedAt  string `json:"created_at"`
}

type OvertimeRecord struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Date      string `json:"date"`
	Hours     int    `json:"hours"`
	Used      bool   `json:"used"`
}

type LeaveBalance struct {
	UserID          int64 `json:"user_id"`
	AnnualLeave     int   `json:"annual_leave"`
	CompensatoryHours int `json:"compensatory_hours"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "./leave_approval.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = initDatabase(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/users", handleUsers)
	mux.HandleFunc("/users/", handleUserByID)

	mux.HandleFunc("/leave-requests", handleLeaveRequests)
	mux.HandleFunc("/leave-requests/", handleLeaveRequestByID)

	mux.HandleFunc("/leave-requests/{id}/actions", handleLeaveActions)
	mux.HandleFunc("/leave-requests/{id}/actions/{actionId}", handleActionDetail)
	mux.HandleFunc("/leave-requests/{id}/actions/{actionId}/comments", handleActionComments)

	mux.HandleFunc("/leave-requests/{id}/approve", handleApproveLeave)
	mux.HandleFunc("/leave-requests/{id}/reject", handleRejectLeave)
	mux.HandleFunc("/leave-requests/{id}/close", handleCloseLeave)

	mux.HandleFunc("/overtime-records", handleOvertimeRecords)
	mux.HandleFunc("/leave-balances", handleLeaveBalances)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func initDatabase() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		role TEXT NOT NULL,
		department TEXT,
		hire_date TEXT NOT NULL,
		manager_id INTEGER,
		hr_id INTEGER,
		dept_manager_id INTEGER,
		FOREIGN KEY (manager_id) REFERENCES users(id),
		FOREIGN KEY (hr_id) REFERENCES users(id),
		FOREIGN KEY (dept_manager_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS overtime_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		hours INTEGER NOT NULL,
		used INTEGER DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS leave_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		leave_type TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		actual_start_date TEXT,
		actual_end_date TEXT,
		days INTEGER NOT NULL,
		actual_days INTEGER,
		reason TEXT,
		sick_note TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS approval_actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		leave_id INTEGER NOT NULL,
		approver_id INTEGER NOT NULL,
		approver_name TEXT,
		action TEXT NOT NULL,
		reason TEXT,
		created_at TEXT NOT NULL,
		FOREIGN KEY (leave_id) REFERENCES leave_requests(id),
		FOREIGN KEY (approver_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS action_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action_id INTEGER NOT NULL,
		commenter_id INTEGER NOT NULL,
		comment TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (action_id) REFERENCES approval_actions(id),
		FOREIGN KEY (commenter_id) REFERENCES users(id)
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	return seedData()
}

func seedData() error {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Format("2006-01-02 15:04:05")

	_, err = tx.Exec(`INSERT INTO users (name, role, department, hire_date) VALUES (?, ?, ?, ?)`,
		"HR Manager", "hr", "HR", "2020-01-01")
	hrID := int64(1)

	_, err = tx.Exec(`INSERT INTO users (name, role, department, hire_date, hr_id) VALUES (?, ?, ?, ?, ?)`,
		"Tech Manager", "manager", "Tech", "2019-01-01", hrID)
	techMgrID := int64(2)

	_, err = tx.Exec(`INSERT INTO users (name, role, department, hire_date, manager_id, hr_id, dept_manager_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"Employee A", "employee", "Tech", "2022-06-01", techMgrID, hrID, techMgrID)
	empAID := int64(3)

	_, err = tx.Exec(`INSERT INTO users (name, role, department, hire_date, manager_id, hr_id, dept_manager_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"Employee B", "employee", "Tech", "2021-03-15", techMgrID, hrID, techMgrID)

	_, err = tx.Exec(`INSERT INTO overtime_records (user_id, date, hours, used) VALUES (?, ?, ?, ?)`,
		empAID, "2024-12-20", 8, 0)
	_, err = tx.Exec(`INSERT INTO overtime_records (user_id, date, hours, used) VALUES (?, ?, ?, ?)`,
		empAID, "2024-12-21", 4, 0)

	_ = now
	return tx.Commit()
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Missing user ID")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if r.Method == http.MethodGet {
		getUser(w, id)
	} else {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, role, department, hire_date, manager_id, hr_id, dept_manager_id FROM users`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Role, &u.Department, &u.HireDate, &u.ManagerID, &u.HRID, &u.DeptManagerID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		users = append(users, u)
	}

	respondJSON(w, http.StatusOK, users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if u.Name == "" || u.Role == "" || u.HireDate == "" {
		respondError(w, http.StatusBadRequest, "Name, role, and hire_date are required")
		return
	}

	result, err := db.Exec(`INSERT INTO users (name, role, department, hire_date, manager_id, hr_id, dept_manager_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.Name, u.Role, u.Department, u.HireDate, u.ManagerID, u.HRID, u.DeptManagerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	u.ID = id
	respondJSON(w, http.StatusCreated, u)
}

func getUser(w http.ResponseWriter, id int64) {
	var u User
	err := db.QueryRow(`SELECT id, name, role, department, hire_date, manager_id, hr_id, dept_manager_id FROM users WHERE id = ?`, id).Scan(
		&u.ID, &u.Name, &u.Role, &u.Department, &u.HireDate, &u.ManagerID, &u.HRID, &u.DeptManagerID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, u)
}

func calculateAnnualLeaveBalance(userID int64) (int, error) {
	var hireDate string
	err := db.QueryRow(`SELECT hire_date FROM users WHERE id = ?`, userID).Scan(&hireDate)
	if err != nil {
		return 0, err
	}

	hd, err := time.Parse("2006-01-02", hireDate)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	years := now.Year() - hd.Year()
	if now.YearDay() < hd.YearDay() {
		years--
	}

	var balance int
	if years < 1 {
		balance = 5
	} else if years < 10 {
		balance = 10
	} else {
		balance = 15
	}

	var used int
	db.QueryRow(`SELECT COALESCE(SUM(actual_days), 0) FROM leave_requests WHERE user_id = ? AND leave_type = ? AND status IN (?, ?, ?)`,
		userID, LeaveTypeAnnual, LeaveStatusApproved, LeaveStatusRevoked, LeaveStatusClosed).Scan(&used)

	return balance - used, nil
}

func hasUnusedOvertime(userID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM overtime_records WHERE user_id = ? AND used = 0`, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func isWeekendOrHoliday(date time.Time) bool {
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return true
	}

	holidays := map[string]bool{
		"01-01": true,
		"01-28": true,
		"01-29": true,
		"01-30": true,
		"01-31": true,
		"02-01": true,
		"02-02": true,
		"02-03": true,
		"02-04": true,
		"04-04": true,
		"04-05": true,
		"04-06": true,
		"05-01": true,
		"05-02": true,
		"05-03": true,
		"05-04": true,
		"05-05": true,
		"06-10": true,
		"09-17": true,
		"09-18": true,
		"09-19": true,
		"10-01": true,
		"10-02": true,
		"10-03": true,
		"10-04": true,
		"10-05": true,
		"10-06": true,
		"10-07": true,
	}

	dateStr := date.Format("01-02")
	return holidays[dateStr]
}

func calculateActualDates(startDate, endDate string) (string, string, int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return "", "", 0, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return "", "", 0, err
	}

	actualStart := start
	for isWeekendOrHoliday(actualStart) {
		actualStart = actualStart.AddDate(0, 0, 1)
	}

	actualEnd := actualStart
	workDays := 0
	for workDays <= int(end.Sub(start).Hours()/24) {
		if !isWeekendOrHoliday(actualEnd) {
			workDays++
			if workDays > int(end.Sub(start).Hours()/24) {
				break
			}
		}
		if workDays <= int(end.Sub(start).Hours()/24) {
			actualEnd = actualEnd.AddDate(0, 0, 1)
		}
	}

	return actualStart.Format("2006-01-02"), actualEnd.Format("2006-01-02"), workDays, nil
}

func checkOverlap(userID int64, startDate, endDate string, excludeID *int64) (bool, string, string, error) {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	query := `SELECT start_date, end_date FROM leave_requests WHERE user_id = ? AND status IN (?, ?, ?)`
	args := []interface{}{userID, LeaveStatusPending, LeaveStatusApproved}
	if excludeID != nil {
		query += ` AND id != ?`
		args = append(args, *excludeID)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return false, "", "", err
	}
	defer rows.Close()

	for rows.Next() {
		var s, e string
		rows.Scan(&s, &e)
		existingStart, _ := time.Parse("2006-01-02", s)
		existingEnd, _ := time.Parse("2006-01-02", e)

		if !(end.Before(existingStart) || start.After(existingEnd)) {
			return true, s, e, nil
		}
	}

	return false, "", "", nil
}

func handleLeaveRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listLeaveRequests(w, r)
	case http.MethodPost:
		createLeaveRequest(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleLeaveRequestByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/leave-requests/")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Missing leave request ID")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		getLeaveRequest(w, id)
	case http.MethodPut:
		updateLeaveRequest(w, r, id)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func listLeaveRequests(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")

	query := `SELECT id, user_id, leave_type, start_date, end_date, actual_start_date, actual_end_date, days, actual_days, reason, sick_note, status, created_at, updated_at FROM leave_requests WHERE 1=1`
	args := []interface{}{}

	if userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			query += ` AND user_id = ?`
			args = append(args, userID)
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var leaves []LeaveRequest
	for rows.Next() {
		var l LeaveRequest
		var actualStart, actualEnd, sickNote sql.NullString
		var actualDays sql.NullInt64

		err := rows.Scan(&l.ID, &l.UserID, &l.LeaveType, &l.StartDate, &l.EndDate,
			&actualStart, &actualEnd, &l.Days, &actualDays, &l.Reason, &sickNote,
			&l.Status, &l.CreatedAt, &l.UpdatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if actualStart.Valid {
			l.ActualStartDate = actualStart.String
		}
		if actualEnd.Valid {
			l.ActualEndDate = actualEnd.String
		}
		if sickNote.Valid {
			l.SickNote = sickNote.String
		}
		if actualDays.Valid {
			l.ActualDays = int(actualDays.Int64)
		}

		leaves = append(leaves, l)
	}

	respondJSON(w, http.StatusOK, leaves)
}

func createLeaveRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    int64     `json:"user_id"`
		LeaveType LeaveType `json:"leave_type"`
		StartDate string    `json:"start_date"`
		EndDate   string    `json:"end_date"`
		Reason    string    `json:"reason"`
		SickNote  string    `json:"sick_note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.UserID == 0 || req.LeaveType == "" || req.StartDate == "" || req.EndDate == "" {
		respondError(w, http.StatusBadRequest, "user_id, leave_type, start_date, end_date are required")
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid start_date format, use YYYY-MM-DD")
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid end_date format, use YYYY-MM-DD")
		return
	}
	if end.Before(start) {
		respondError(w, http.StatusBadRequest, "end_date cannot be before start_date")
		return
	}
	days := int(end.Sub(start).Hours()/24) + 1

	if req.LeaveType == LeaveTypeAnnual {
		balance, err := calculateAnnualLeaveBalance(req.UserID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if balance < days {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":             "Insufficient annual leave balance",
				"remaining_balance": balance,
			})
			return
		}
	}

	if req.LeaveType == LeaveTypeSick && days > 3 && req.SickNote == "" {
		respondError(w, http.StatusBadRequest, "Sick leave over 3 days requires sick note")
		return
	}

	if req.LeaveType == LeaveTypeCompensatory {
		hasOT, err := hasUnusedOvertime(req.UserID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !hasOT {
			respondError(w, http.StatusBadRequest, "No unused overtime records for compensatory leave")
			return
		}
	}

	overlap, conflictStart, conflictEnd, err := checkOverlap(req.UserID, req.StartDate, req.EndDate, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if overlap {
		respondJSON(w, http.StatusConflict, map[string]interface{}{
			"error":           "Leave time overlaps with existing request",
			"conflict_period": map[string]string{"start": conflictStart, "end": conflictEnd},
		})
		return
	}

	actualStart, actualEnd, actualDays, err := calculateActualDates(req.StartDate, req.EndDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.Exec(`INSERT INTO leave_requests (user_id, leave_type, start_date, end_date, actual_start_date, actual_end_date, days, actual_days, reason, sick_note, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.UserID, req.LeaveType, req.StartDate, req.EndDate, actualStart, actualEnd, days, actualDays, req.Reason, req.SickNote, LeaveStatusPending, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	leave := LeaveRequest{
		ID:              id,
		UserID:          req.UserID,
		LeaveType:       req.LeaveType,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		ActualStartDate: actualStart,
		ActualEndDate:   actualEnd,
		Days:            days,
		ActualDays:      actualDays,
		Reason:          req.Reason,
		SickNote:        req.SickNote,
		Status:          LeaveStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	respondJSON(w, http.StatusCreated, leave)
}

func getLeaveRequest(w http.ResponseWriter, id int64) {
	var l LeaveRequest
	var actualStart, actualEnd, sickNote sql.NullString
	var actualDays sql.NullInt64

	err := db.QueryRow(`SELECT id, user_id, leave_type, start_date, end_date, actual_start_date, actual_end_date, days, actual_days, reason, sick_note, status, created_at, updated_at FROM leave_requests WHERE id = ?`, id).Scan(
		&l.ID, &l.UserID, &l.LeaveType, &l.StartDate, &l.EndDate,
		&actualStart, &actualEnd, &l.Days, &actualDays, &l.Reason, &sickNote,
		&l.Status, &l.CreatedAt, &l.UpdatedAt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Leave request not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if actualStart.Valid {
		l.ActualStartDate = actualStart.String
	}
	if actualEnd.Valid {
		l.ActualEndDate = actualEnd.String
	}
	if sickNote.Valid {
		l.SickNote = sickNote.String
	}
	if actualDays.Valid {
		l.ActualDays = int(actualDays.Int64)
	}

	respondJSON(w, http.StatusOK, l)
}

func updateLeaveRequest(w http.ResponseWriter, r *http.Request, id int64) {
	var leave LeaveRequest
	var currentStatus LeaveStatus
	var currentUserID int64

	err := db.QueryRow(`SELECT status, user_id FROM leave_requests WHERE id = ?`, id).Scan(&currentStatus, &currentUserID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Leave request not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if currentStatus != LeaveStatusRejected {
		respondError(w, http.StatusBadRequest, "Only rejected leave requests can be modified")
		return
	}

	var req struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Reason    string `json:"reason"`
		SickNote  string `json:"sick_note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	overlap, conflictStart, conflictEnd, err := checkOverlap(currentUserID, req.StartDate, req.EndDate, &id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if overlap {
		respondJSON(w, http.StatusConflict, map[string]interface{}{
			"error":           "Leave time overlaps with existing request",
			"conflict_period": map[string]string{"start": conflictStart, "end": conflictEnd},
		})
		return
	}

	actualStart, actualEnd, actualDays, err := calculateActualDates(req.StartDate, req.EndDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	days := int(end.Sub(start).Hours()/24) + 1
	now := time.Now().Format("2006-01-02 15:04:05")

	_, err = db.Exec(`UPDATE leave_requests SET start_date = ?, end_date = ?, actual_start_date = ?, actual_end_date = ?, days = ?, actual_days = ?, reason = ?, sick_note = ?, status = ?, updated_at = ? WHERE id = ?`,
		req.StartDate, req.EndDate, actualStart, actualEnd, days, actualDays, req.Reason, req.SickNote, LeaveStatusPending, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	leave.ID = id
	leave.UserID = currentUserID
	leave.StartDate = req.StartDate
	leave.EndDate = req.EndDate
	leave.ActualStartDate = actualStart
	leave.ActualEndDate = actualEnd
	leave.Days = days
	leave.ActualDays = actualDays
	leave.Reason = req.Reason
	leave.SickNote = req.SickNote
	leave.Status = LeaveStatusPending
	leave.UpdatedAt = now

	respondJSON(w, http.StatusOK, leave)
}

func getApprovalChain(userID int64, days int) ([]int64, error) {
	var user User
	err := db.QueryRow(`SELECT id, manager_id, hr_id, dept_manager_id FROM users WHERE id = ?`, userID).Scan(
		&user.ID, &user.ManagerID, &user.HRID, &user.DeptManagerID)
	if err != nil {
		return nil, err
	}

	var chain []int64

	if user.ManagerID != nil {
		chain = append(chain, *user.ManagerID)
	}

	if days >= 3 && days <= 5 {
		if user.HRID != nil {
			chain = append(chain, *user.HRID)
		}
	} else if days > 5 {
		if user.DeptManagerID != nil && *user.DeptManagerID != *user.ManagerID {
			chain = append(chain, *user.DeptManagerID)
		}
		if user.HRID != nil {
			chain = append(chain, *user.HRID)
		}
	}

	return chain, nil
}

func isInApprovalChain(approverID, leaveID int64) (bool, error) {
	var userID int64
	var days int
	err := db.QueryRow(`SELECT user_id, actual_days FROM leave_requests WHERE id = ?`, leaveID).Scan(&userID, &days)
	if err != nil {
		return false, err
	}

	chain, err := getApprovalChain(userID, days)
	if err != nil {
		return false, err
	}

	for _, id := range chain {
		if id == approverID {
			return true, nil
		}
	}
	return false, nil
}

func handleApproveLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/leave-requests/")
	idStr := strings.TrimSuffix(path, "/approve")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	var req struct {
		ApproverID int64  `json:"approver_id"`
		Reason     string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	var currentStatus LeaveStatus
	var userID int64
	err = db.QueryRow(`SELECT status, user_id FROM leave_requests WHERE id = ?`, id).Scan(&currentStatus, &userID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Leave request not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if currentStatus != LeaveStatusPending {
		respondError(w, http.StatusBadRequest, "Only pending leave requests can be approved")
		return
	}

	if req.ApproverID == userID {
		respondError(w, http.StatusForbidden, "Cannot approve your own leave request")
		return
	}

	isApprover, err := isInApprovalChain(req.ApproverID, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isApprover {
		respondError(w, http.StatusForbidden, "You are not authorized to approve this leave request")
		return
	}

	var approverName string
	db.QueryRow(`SELECT name FROM users WHERE id = ?`, req.ApproverID).Scan(&approverName)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec(`INSERT INTO approval_actions (leave_id, approver_id, approver_name, action, reason, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, req.ApproverID, approverName, "approved", req.Reason, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.Exec(`UPDATE leave_requests SET status = ?, updated_at = ? WHERE id = ?`,
		LeaveStatusApproved, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func handleRejectLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/leave-requests/")
	idStr := strings.TrimSuffix(path, "/reject")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	var req struct {
		ApproverID int64  `json:"approver_id"`
		Reason     string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "Rejection reason is required")
		return
	}

	var currentStatus LeaveStatus
	var userID int64
	err = db.QueryRow(`SELECT status, user_id FROM leave_requests WHERE id = ?`, id).Scan(&currentStatus, &userID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Leave request not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if currentStatus != LeaveStatusPending {
		respondError(w, http.StatusBadRequest, "Only pending leave requests can be rejected")
		return
	}

	if req.ApproverID == userID {
		respondError(w, http.StatusForbidden, "Cannot reject your own leave request")
		return
	}

	isApprover, err := isInApprovalChain(req.ApproverID, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isApprover {
		respondError(w, http.StatusForbidden, "You are not authorized to reject this leave request")
		return
	}

	var approverName string
	db.QueryRow(`SELECT name FROM users WHERE id = ?`, req.ApproverID).Scan(&approverName)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec(`INSERT INTO approval_actions (leave_id, approver_id, approver_name, action, reason, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, req.ApproverID, approverName, "rejected", req.Reason, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.Exec(`UPDATE leave_requests SET status = ?, updated_at = ? WHERE id = ?`,
		LeaveStatusRejected, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func handleCloseLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/leave-requests/")
	idStr := strings.TrimSuffix(path, "/close")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	var currentStatus LeaveStatus
	var userID int64
	var leaveType LeaveType
	var actualDays int

	err = db.QueryRow(`SELECT status, user_id, leave_type, actual_days FROM leave_requests WHERE id = ?`, id).Scan(&currentStatus, &userID, &leaveType, &actualDays)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Leave request not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if currentStatus != LeaveStatusApproved {
		respondError(w, http.StatusBadRequest, "Only approved leave requests can be closed")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	if leaveType == LeaveTypeCompensatory {
		hoursNeeded := actualDays * 8
		rows, err := tx.Query(`SELECT id, hours FROM overtime_records WHERE user_id = ? AND used = 0 ORDER BY date ASC`, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for rows.Next() && hoursNeeded > 0 {
			var otID int64
			var otHours int
			rows.Scan(&otID, &otHours)
			if otHours <= hoursNeeded {
				_, _ = tx.Exec(`UPDATE overtime_records SET used = 1 WHERE id = ?`, otID)
				hoursNeeded -= otHours
			} else {
				hoursNeeded = 0
			}
		}
		rows.Close()
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = tx.Exec(`UPDATE leave_requests SET status = ?, updated_at = ? WHERE id = ?`,
		LeaveStatusClosed, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func handleLeaveActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/leave-requests/")
	idStr := strings.TrimSuffix(path, "/actions")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	var exists int
	err = db.QueryRow(`SELECT COUNT(*) FROM leave_requests WHERE id = ?`, id).Scan(&exists)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exists == 0 {
		respondError(w, http.StatusNotFound, "Leave request not found")
		return
	}

	rows, err := db.Query(`SELECT id, leave_id, approver_id, approver_name, action, reason, created_at FROM approval_actions WHERE leave_id = ? ORDER BY created_at ASC`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var actions []ApprovalAction
	for rows.Next() {
		var a ApprovalAction
		var reason sql.NullString
		err := rows.Scan(&a.ID, &a.LeaveID, &a.ApproverID, &a.ApproverName, &a.Action, &reason, &a.CreatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if reason.Valid {
			a.Reason = reason.String
		}
		actions = append(actions, a)
	}

	respondJSON(w, http.StatusOK, actions)
}

func handleActionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	leaveID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid leave request ID")
		return
	}

	actionID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid action ID")
		return
	}

	var a ApprovalAction
	var reason sql.NullString
	err = db.QueryRow(`SELECT id, leave_id, approver_id, approver_name, action, reason, created_at FROM approval_actions WHERE id = ? AND leave_id = ?`,
		actionID, leaveID).Scan(&a.ID, &a.LeaveID, &a.ApproverID, &a.ApproverName, &a.Action, &reason, &a.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Action not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if reason.Valid {
		a.Reason = reason.String
	}

	respondJSON(w, http.StatusOK, a)
}

func handleActionComments(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		respondError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	actionID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid action ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		listActionComments(w, actionID)
	case http.MethodPost:
		addActionComment(w, r, actionID)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func listActionComments(w http.ResponseWriter, actionID int64) {
	rows, err := db.Query(`SELECT id, action_id, commenter_id, comment, created_at FROM action_comments WHERE action_id = ? ORDER BY created_at ASC`, actionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var comments []ActionComment
	for rows.Next() {
		var c ActionComment
		err := rows.Scan(&c.ID, &c.ActionID, &c.CommenterID, &c.Comment, &c.CreatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		comments = append(comments, c)
	}

	respondJSON(w, http.StatusOK, comments)
}

func addActionComment(w http.ResponseWriter, r *http.Request, actionID int64) {
	var req struct {
		CommenterID int64  `json:"commenter_id"`
		Comment     string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Comment == "" {
		respondError(w, http.StatusBadRequest, "Comment is required")
		return
	}

	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM approval_actions WHERE id = ?`, actionID).Scan(&exists)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exists == 0 {
		respondError(w, http.StatusNotFound, "Action not found")
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.Exec(`INSERT INTO action_comments (action_id, commenter_id, comment, created_at) VALUES (?, ?, ?, ?)`,
		actionID, req.CommenterID, req.Comment, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	comment := ActionComment{
		ID:          id,
		ActionID:    actionID,
		CommenterID: req.CommenterID,
		Comment:     req.Comment,
		CreatedAt:   now,
	}

	respondJSON(w, http.StatusCreated, comment)
}

func handleOvertimeRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listOvertimeRecords(w, r)
	case http.MethodPost:
		createOvertimeRecord(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func listOvertimeRecords(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	query := `SELECT id, user_id, date, hours, used FROM overtime_records WHERE 1=1`
	args := []interface{}{}

	if userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			query += ` AND user_id = ?`
			args = append(args, userID)
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var records []OvertimeRecord
	for rows.Next() {
		var o OvertimeRecord
		var usedInt int
		err := rows.Scan(&o.ID, &o.UserID, &o.Date, &o.Hours, &usedInt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		o.Used = usedInt > 0
		records = append(records, o)
	}

	respondJSON(w, http.StatusOK, records)
}

func createOvertimeRecord(w http.ResponseWriter, r *http.Request) {
	var o OvertimeRecord
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if o.UserID == 0 || o.Date == "" || o.Hours <= 0 {
		respondError(w, http.StatusBadRequest, "user_id, date, and hours are required")
		return
	}

	result, err := db.Exec(`INSERT INTO overtime_records (user_id, date, hours, used) VALUES (?, ?, ?, 0)`,
		o.UserID, o.Date, o.Hours)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	o.ID = id
	o.Used = false
	respondJSON(w, http.StatusCreated, o)
}

func handleLeaveBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	annualBalance, err := calculateAnnualLeaveBalance(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var compHours int
	db.QueryRow(`SELECT COALESCE(SUM(hours), 0) FROM overtime_records WHERE user_id = ? AND used = 0`, userID).Scan(&compHours)

	balance := LeaveBalance{
		UserID:            userID,
		AnnualLeave:       annualBalance,
		CompensatoryHours: compHours,
	}

	respondJSON(w, http.StatusOK, balance)
}

var _ = io.EOF
