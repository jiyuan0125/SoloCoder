package handler

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"shift-scheduler/internal/models"
	"shift-scheduler/internal/repository"
	"shift-scheduler/internal/service"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, err error) {
	if ve, ok := err.(*service.ValidationError); ok {
		writeJSON(w, ve.Code, map[string]string{"error": ve.Message})
		return
	}
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "resource not found"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

func parseID(r *http.Request, paramName string) (int64, error) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, part := range parts {
		if part == paramName && i+1 < len(parts) {
			return strconv.ParseInt(parts[i+1], 10, 64)
		}
	}
	return 0, fmt.Errorf("id not found")
}

func DepartmentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		depts, err := repository.ListDepartments()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, depts)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		dept, err := repository.CreateDepartment(req.Name)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, dept)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func EmployeeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		emps, err := repository.ListEmployees()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, emps)
	case http.MethodPost:
		var req struct {
			Name         string `json:"name"`
			DepartmentID int64  `json:"department_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		emp, err := repository.CreateEmployee(req.Name, req.DepartmentID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, emp)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func HolidayHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		holidays, err := repository.ListHolidays()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, holidays)
	case http.MethodPost:
		var req struct {
			Date string `json:"date"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			writeError(w, &service.ValidationError{Code: 400, Message: "invalid date format, use YYYY-MM-DD"})
			return
		}
		holiday, err := repository.AddHoliday(date, req.Name)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, holiday)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func ShiftListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		shifts, err := repository.ListShifts()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, shifts)
	case http.MethodPost:
		var req struct {
			EmployeeID int64  `json:"employee_id"`
			ShiftDate  string `json:"shift_date"`
			StartTime  string `json:"start_time"`
			EndTime    string `json:"end_time"`
			Position   string `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}

		shiftDate, err := time.Parse("2006-01-02", req.ShiftDate)
		if err != nil {
			writeError(w, &service.ValidationError{Code: 400, Message: "invalid shift_date format, use YYYY-MM-DD"})
			return
		}

		startTime, err := time.Parse("15:04", req.StartTime)
		if err != nil {
			writeError(w, &service.ValidationError{Code: 400, Message: "invalid start_time format, use HH:MM"})
			return
		}

		endTime, err := time.Parse("15:04", req.EndTime)
		if err != nil {
			writeError(w, &service.ValidationError{Code: 400, Message: "invalid end_time format, use HH:MM"})
			return
		}

		if !endTime.After(startTime) {
			writeError(w, &service.ValidationError{Code: 400, Message: "end_time must be after start_time"})
			return
		}

		shift, err := service.CreateShift(req.EmployeeID, shiftDate, startTime, endTime, req.Position)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, shift)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func ShiftDetailHandler(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseID(r, "shifts")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid shift id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		shift, err := repository.GetShiftByID(shiftID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, shift)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func ShiftHistoryHandler(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseID(r, "shifts")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid shift id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		history, err := repository.GetOperationHistory(shiftID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, history)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func HistoryNoteHandler(w http.ResponseWriter, r *http.Request) {
	historyID, err := parseID(r, "history")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid history id"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req struct {
			Note string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		if err := repository.AddNoteToHistory(historyID, req.Note); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "note added"})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func ShiftAdvanceHandler(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseID(r, "shifts")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid shift id"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req struct {
			OperatorID int64  `json:"operator_id"`
			Note       string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		if err := service.AdvanceStatus(shiftID, req.OperatorID, req.Note); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "advanced"})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func ShiftRejectHandler(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseID(r, "shifts")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid shift id"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req struct {
			OperatorID int64  `json:"operator_id"`
			Note       string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		if err := service.RejectShift(shiftID, req.OperatorID, req.Note); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func SwapRequestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			RequesterShiftID int64 `json:"requester_shift_id"`
			ResponderShiftID int64 `json:"responder_shift_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, err)
			return
		}
		swap, err := service.RequestSwap(req.RequesterShiftID, req.ResponderShiftID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, swap)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, nil)
	}
}

func SwapActionHandler(w http.ResponseWriter, r *http.Request) {
	swapID, err := parseID(r, "swaps")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid swap id"})
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	action := ""
	for i, part := range parts {
		if part == "swaps" && i+2 < len(parts) {
			action = parts[i+2]
			break
		}
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, nil)
		return
	}

	switch action {
	case "confirm":
		if err := service.ConfirmSwap(swapID); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
	case "cancel":
		if err := service.CancelSwap(swapID); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
	case "reject":
		if err := service.RejectSwap(swapID); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invalid action"})
	}
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, nil)
		return
	}

	query := r.URL.Query()
	yearStr := query.Get("year")
	monthStr := query.Get("month")
	deptIDStr := query.Get("department_id")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		year = time.Now().Year()
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		month = int(time.Now().Month())
	}

	var stats []models.MonthlyStats
	if deptIDStr != "" {
		deptID, err := strconv.Atoi(deptIDStr)
		if err != nil {
			writeError(w, &service.ValidationError{Code: 400, Message: "invalid department_id"})
			return
		}
		stats, err = service.GetDepartmentStats(deptID, year, month)
		if err != nil {
			writeError(w, err)
			return
		}
	} else {
		stats, err = service.GetMonthlyStats(year, month)
		if err != nil {
			writeError(w, err)
			return
		}
	}

	if query.Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=stats_%04d_%02d.csv", year, month))
		cw := csv.NewWriter(w)
		cw.Write([]string{"EmployeeID", "EmployeeName", "DepartmentID", "DepartmentName", "Month", "ScheduledHours", "ActualHours", "Difference"})
		for _, s := range stats {
			cw.Write([]string{
				strconv.FormatInt(s.EmployeeID, 10),
				s.EmployeeName,
				strconv.FormatInt(s.DepartmentID, 10),
				s.DepartmentName,
				s.Month,
				fmt.Sprintf("%.1f", s.ScheduledHours),
				fmt.Sprintf("%.1f", s.ActualHours),
				fmt.Sprintf("%.1f", s.Difference),
			})
		}
		cw.Flush()
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func UpdateActualHoursHandler(w http.ResponseWriter, r *http.Request) {
	shiftID, err := parseID(r, "shifts")
	if err != nil {
		writeError(w, &service.ValidationError{Code: 400, Message: "invalid shift id"})
		return
	}

	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req struct {
		ActualHours float64 `json:"actual_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
		return
	}

	if err := repository.UpdateShiftActualHours(shiftID, req.ActualHours); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/departments", DepartmentHandler)
	mux.HandleFunc("/api/employees", EmployeeHandler)
	mux.HandleFunc("/api/holidays", HolidayHandler)
	mux.HandleFunc("/api/shifts", ShiftListHandler)
	mux.HandleFunc("/api/shifts/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/shifts/")
		if strings.Contains(path, "/history") {
			if strings.Contains(path, "/notes") {
				HistoryNoteHandler(w, r)
			} else {
				ShiftHistoryHandler(w, r)
			}
		} else if strings.Contains(path, "/advance") {
			ShiftAdvanceHandler(w, r)
		} else if strings.Contains(path, "/reject") {
			ShiftRejectHandler(w, r)
		} else if strings.Contains(path, "/actual-hours") {
			UpdateActualHoursHandler(w, r)
		} else {
			ShiftDetailHandler(w, r)
		}
	})
	mux.HandleFunc("/api/swaps", SwapRequestHandler)
	mux.HandleFunc("/api/swaps/", SwapActionHandler)
	mux.HandleFunc("/api/stats", StatsHandler)
}
