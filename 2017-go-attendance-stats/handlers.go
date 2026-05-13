package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func handleEmployees(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		employees, err := getAllEmployees()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if employees == nil {
			employees = []Employee{}
		}
		writeJSON(w, http.StatusOK, employees)
	case http.MethodPost:
		var emp Employee
		if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if emp.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		result, err := db.Exec("INSERT INTO employees (name) VALUES (?)", emp.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		id, _ := result.LastInsertId()
		emp.ID = int(id)
		writeJSON(w, http.StatusCreated, emp)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleEmployeeDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/employees/")
	if strings.Contains(path, "/attendance") {
		handleEmployeeAttendance(w, r, path)
		return
	}
	if strings.Contains(path, "/makeup") {
		handleEmployeeMakeup(w, r, path)
		return
	}

	idStr := strings.TrimSuffix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		emp, err := getEmployeeByID(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if emp == nil {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeJSON(w, http.StatusOK, emp)
	case http.MethodPut:
		var emp Employee
		if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		_, err := db.Exec("UPDATE employees SET name = ? WHERE id = ?", emp.Name, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		emp.ID = id
		writeJSON(w, http.StatusOK, emp)
	case http.MethodDelete:
		_, err := db.Exec("DELETE FROM employees WHERE id = ?", id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleEmployeeAttendance(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	empID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee id")
		return
	}

	emp, err := getEmployeeByID(empID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if emp == nil {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
			return
		}
	} else {
		date = time.Now()
	}

	records, err := getAttendanceByEmployeeAndDate(empID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if records == nil {
		records = []AttendanceRecord{}
	}
	writeJSON(w, http.StatusOK, records)
}

func handleEmployeeMakeup(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	empID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee id")
		return
	}

	emp, err := getEmployeeByID(empID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if emp == nil {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}

	rows, err := db.Query(
		`SELECT id, employee_id, date, punch_type, reason, status, created_at, updated_at 
		 FROM makeup_requests WHERE employee_id = ? ORDER BY created_at DESC`,
		empID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var requests []MakeupRequest
	for rows.Next() {
		var r MakeupRequest
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.Date, &r.PunchType, &r.Reason, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		requests = append(requests, r)
	}
	if requests == nil {
		requests = []MakeupRequest{}
	}
	writeJSON(w, http.StatusOK, requests)
}

func handleAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	records, err := getAllAttendanceRecords()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if records == nil {
		records = []AttendanceRecord{}
	}
	writeJSON(w, http.StatusOK, records)
}

func handleAttendanceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/attendance/")
	id, err := strconv.Atoi(strings.TrimSuffix(path, "/"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	record, err := getAttendanceRecordByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if record == nil {
		writeError(w, http.StatusNotFound, "record not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func handlePunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		EmployeeID int    `json:"employee_id"`
		PunchType  string `json:"punch_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PunchType != "in" && req.PunchType != "out" {
		writeError(w, http.StatusBadRequest, "punch_type must be 'in' or 'out'")
		return
	}

	emp, err := getEmployeeByID(req.EmployeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if emp == nil {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	}

	now := time.Now()

	has, err := hasPunchedToday(req.EmployeeID, req.PunchType, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if has {
		writeJSON(w, http.StatusOK, map[string]string{"message": "already punched today for this type"})
		return
	}

	status := calculateStatus(now, req.PunchType)

	record := &AttendanceRecord{
		EmployeeID: req.EmployeeID,
		PunchTime:  now,
		PunchType:  req.PunchType,
		Status:     status,
	}

	if err := insertAttendanceRecord(record); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, record)
}

func calculateStatus(punchTime time.Time, punchType string) string {
	hour := punchTime.Hour()
	minute := punchTime.Minute()
	totalMinutes := hour*60 + minute

	if punchType == "in" {
		if totalMinutes <= 9*60 {
			return "normal"
		} else if totalMinutes <= 9*60+15 {
			return "late_minor"
		} else if totalMinutes <= 10*60 {
			return "late"
		} else {
			return "absent_half"
		}
	} else {
		if totalMinutes >= 18*60 {
			return "normal"
		} else if totalMinutes >= 17*60+30 {
			return "early_minor"
		} else if totalMinutes >= 16*60 {
			return "early"
		} else {
			return "absent_half"
		}
	}
}

func handleMakeup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		requests, err := getMakeupRequests()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if requests == nil {
			requests = []MakeupRequest{}
		}
		writeJSON(w, http.StatusOK, requests)
	case http.MethodPost:
		var req MakeupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.EmployeeID == 0 {
			writeError(w, http.StatusBadRequest, "employee_id is required")
			return
		}
		if req.Date == "" {
			writeError(w, http.StatusBadRequest, "date is required")
			return
		}
		if req.PunchType != "in" && req.PunchType != "out" {
			writeError(w, http.StatusBadRequest, "punch_type must be 'in' or 'out'")
			return
		}
		if req.Reason == "" {
			writeError(w, http.StatusBadRequest, "reason is required")
			return
		}

		emp, err := getEmployeeByID(req.EmployeeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if emp == nil {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}

		has, err := hasMakeupToday(req.EmployeeID, req.Date)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if has {
			writeError(w, http.StatusConflict, "今日已补签")
			return
		}

		req.Status = "draft"
		if err := insertMakeupRequest(&req); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, req)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleMakeupDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/makeup/")

	if strings.Contains(path, "/") {
		handleMakeupAction(w, r, path)
		return
	}

	id, err := strconv.Atoi(strings.TrimSuffix(path, "/"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := getMakeupRequestByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if req == nil {
		writeError(w, http.StatusNotFound, "makeup request not found")
		return
	}
	writeJSON(w, http.StatusOK, req)
}

var statusFlow = map[string]string{
	"draft":     "approved",
	"approved":  "executed",
	"executed":  "confirmed",
	"confirmed": "closed",
}

func handleMakeupAction(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	action := parts[1]

	req, err := getMakeupRequestByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if req == nil {
		writeError(w, http.StatusNotFound, "makeup request not found")
		return
	}

	var nextStatus string
	switch action {
	case "approve":
		if req.Status != "draft" {
			writeError(w, http.StatusBadRequest, "can only approve draft requests")
			return
		}
		nextStatus = "approved"
	case "execute":
		if req.Status != "approved" {
			writeError(w, http.StatusBadRequest, "can only execute approved requests")
			return
		}
		nextStatus = "executed"

		date, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			has, _ := hasPunchedToday(req.EmployeeID, req.PunchType, date)
			if !has {
				punchTime := date
				if req.PunchType == "in" {
					punchTime = time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())
				} else {
					punchTime = time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())
				}
				status := calculateStatus(punchTime, req.PunchType)
				insertAttendanceRecord(&AttendanceRecord{
					EmployeeID: req.EmployeeID,
					PunchTime:  punchTime,
					PunchType:  req.PunchType,
					Status:     status,
				})
			}
		}
	case "confirm":
		if req.Status != "executed" {
			writeError(w, http.StatusBadRequest, "can only confirm executed requests")
			return
		}
		nextStatus = "confirmed"
	case "close":
		if req.Status != "confirmed" {
			writeError(w, http.StatusBadRequest, "can only close confirmed requests")
			return
		}
		nextStatus = "closed"
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid action: %s", action))
		return
	}

	if err := updateMakeupStatus(id, nextStatus); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	req.Status = nextStatus
	writeJSON(w, http.StatusOK, req)
}

func handleMonthlyStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	empIDStr := r.URL.Query().Get("employee_id")

	var year, month int
	var empID *int
	var err error

	now := time.Now()
	if yearStr == "" {
		year = now.Year()
	} else {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}
	}
	if monthStr == "" {
		month = int(now.Month())
	} else {
		month, err = strconv.Atoi(monthStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}
	}
	if empIDStr != "" {
		id, err := strconv.Atoi(empIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		empID = &id
	}

	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endDate := fmt.Sprintf("%d-%02d-01", year, month+1)

	query := `
		SELECT 
			ar.employee_id,
			COUNT(DISTINCT date(ar.punch_time)) as attendance_days,
			SUM(CASE WHEN ar.status IN ('late', 'late_minor') THEN 1 ELSE 0 END) as late_count,
			SUM(CASE WHEN ar.status IN ('early', 'early_minor') THEN 1 ELSE 0 END) as early_count
		FROM attendance_records ar
		WHERE ar.punch_time >= ? AND ar.punch_time < ?
	`
	args := []interface{}{startDate, endDate}

	if empID != nil {
		query += " AND ar.employee_id = ?"
		args = append(args, *empID)
	}
	query += " GROUP BY ar.employee_id"

	rows, err := db.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	statsMap := make(map[int]*MonthlyStats)
	for rows.Next() {
		var empIDVal int
		var attendanceDays, lateCount, earlyCount int
		if err := rows.Scan(&empIDVal, &attendanceDays, &lateCount, &earlyCount); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		statsMap[empIDVal] = &MonthlyStats{
			EmployeeID:     empIDVal,
			Year:           year,
			Month:          month,
			AttendanceDays: attendanceDays,
			LateCount:      lateCount,
			EarlyLeaveCount: earlyCount,
		}
	}

	overtimeQuery := `
		SELECT 
			ar.employee_id,
			date(ar.punch_time) as punch_date,
			MIN(ar.punch_time) as in_time,
			MAX(ar.punch_time) as out_time
		FROM attendance_records ar
		WHERE ar.punch_time >= ? AND ar.punch_time < ?
	`
	args2 := []interface{}{startDate, endDate}
	if empID != nil {
		overtimeQuery += " AND ar.employee_id = ?"
		args2 = append(args2, *empID)
	}
	overtimeQuery += " GROUP BY ar.employee_id, date(ar.punch_time) HAVING COUNT(CASE WHEN ar.punch_type = 'in' THEN 1 END) > 0 AND COUNT(CASE WHEN ar.punch_type = 'out' THEN 1 END) > 0"

	otRows, err := db.Query(overtimeQuery, args2...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer otRows.Close()

	for otRows.Next() {
		var empIDVal int
		var punchDate, inTimeStr, outTimeStr string
		if err := otRows.Scan(&empIDVal, &punchDate, &inTimeStr, &outTimeStr); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		outTime, _ := time.Parse("2006-01-02 15:04:05", outTimeStr)
		_ = inTimeStr

		endTime := time.Date(outTime.Year(), outTime.Month(), outTime.Day(), 18, 0, 0, 0, outTime.Location())

		if outTime.After(endTime.Add(30 * time.Minute)) {
			overtime := outTime.Sub(endTime.Add(30 * time.Minute))
			hours := int(overtime.Hours())
			if hours > 0 {
				if _, ok := statsMap[empIDVal]; !ok {
					statsMap[empIDVal] = &MonthlyStats{EmployeeID: empIDVal, Year: year, Month: month}
				}
				statsMap[empIDVal].OvertimeHours += hours
			}
		}
	}

	statsList := make([]MonthlyStats, 0, len(statsMap))
	for _, s := range statsMap {
		statsList = append(statsList, *s)
	}
	if len(statsList) == 0 {
		statsList = []MonthlyStats{}
	}

	writeJSON(w, http.StatusOK, statsList)
}
