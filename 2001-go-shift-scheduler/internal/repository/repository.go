package repository

import (
	"database/sql"
	"fmt"
	"time"

	"shift-scheduler/internal/database"
	"shift-scheduler/internal/models"
)

func CreateDepartment(name string) (*models.Department, error) {
	result, err := database.DB.Exec("INSERT INTO departments (name) VALUES (?)", name)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Department{ID: id, Name: name}, nil
}

func GetDepartmentByID(id int64) (*models.Department, error) {
	var dept models.Department
	err := database.DB.QueryRow("SELECT id, name FROM departments WHERE id = ?", id).Scan(&dept.ID, &dept.Name)
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func ListDepartments() ([]models.Department, error) {
	rows, err := database.DB.Query("SELECT id, name FROM departments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depts []models.Department
	for rows.Next() {
		var d models.Department
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		depts = append(depts, d)
	}
	return depts, nil
}

func CreateEmployee(name string, deptID int64) (*models.Employee, error) {
	result, err := database.DB.Exec("INSERT INTO employees (name, department_id) VALUES (?, ?)", name, deptID)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Employee{
		ID:           id,
		Name:         name,
		DepartmentID: deptID,
		CreatedAt:    time.Now(),
	}, nil
}

func GetEmployeeByID(id int64) (*models.Employee, error) {
	var emp models.Employee
	var createdAt string
	err := database.DB.QueryRow("SELECT id, name, department_id, created_at FROM employees WHERE id = ?", id).
		Scan(&emp.ID, &emp.Name, &emp.DepartmentID, &createdAt)
	if err != nil {
		return nil, err
	}
	emp.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &emp, nil
}

func ListEmployees() ([]models.Employee, error) {
	rows, err := database.DB.Query("SELECT id, name, department_id, created_at FROM employees")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []models.Employee
	for rows.Next() {
		var e models.Employee
		var createdAt string
		if err := rows.Scan(&e.ID, &e.Name, &e.DepartmentID, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		emps = append(emps, e)
	}
	return emps, nil
}

func AddHoliday(date time.Time, name string) (*models.Holiday, error) {
	result, err := database.DB.Exec("INSERT INTO holidays (date, name) VALUES (?, ?)",
		date.Format("2006-01-02"), name)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Holiday{ID: id, Date: date, Name: name}, nil
}

func IsHoliday(date time.Time) (bool, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM holidays WHERE date = ?", date.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func ListHolidays() ([]models.Holiday, error) {
	rows, err := database.DB.Query("SELECT id, date, name FROM holidays ORDER BY date")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holidays []models.Holiday
	for rows.Next() {
		var h models.Holiday
		var dateStr string
		if err := rows.Scan(&h.ID, &dateStr, &h.Name); err != nil {
			return nil, err
		}
		h.Date, _ = time.Parse("2006-01-02", dateStr)
		holidays = append(holidays, h)
	}
	return holidays, nil
}

func CreateShift(empID int64, shiftDate, startTime, endTime time.Time, position string, isHoliday bool) (*models.ShiftMaster, error) {
	hours := endTime.Sub(startTime).Hours()

	result, err := database.DB.Exec(`
		INSERT INTO shift_masters (employee_id, shift_date, start_time, end_time, position, status, scheduled_hours, is_holiday)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, empID, shiftDate.Format("2006-01-02"), startTime.Format("15:04"), endTime.Format("15:04"), position, models.StatusDraft, hours, isHoliday)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = database.DB.Exec(`
		INSERT INTO shift_details (master_id, employee_id, shift_date, start_time, end_time, position, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, empID, shiftDate.Format("2006-01-02"), startTime.Format("15:04"), endTime.Format("15:04"), position, models.StatusDraft)
	if err != nil {
		return nil, err
	}

	return &models.ShiftMaster{
		ID:             id,
		EmployeeID:     empID,
		ShiftDate:      shiftDate,
		StartTime:      startTime,
		EndTime:        endTime,
		Position:       position,
		Status:         models.StatusDraft,
		ScheduledHours: hours,
		IsHoliday:      isHoliday,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func GetShiftByID(id int64) (*models.ShiftMaster, error) {
	var shift models.ShiftMaster
	var shiftDateStr, startTimeStr, endTimeStr, createdAtStr, updatedAtStr string
	var isHolidayInt int

	err := database.DB.QueryRow(`
		SELECT id, employee_id, shift_date, start_time, end_time, position, status, 
		       scheduled_hours, actual_hours, is_holiday, created_at, updated_at
		FROM shift_masters WHERE id = ?
	`, id).Scan(
		&shift.ID, &shift.EmployeeID, &shiftDateStr, &startTimeStr, &endTimeStr,
		&shift.Position, &shift.Status, &shift.ScheduledHours, &shift.ActualHours,
		&isHolidayInt, &createdAtStr, &updatedAtStr,
	)
	if err != nil {
		return nil, err
	}

	shift.ShiftDate, _ = time.Parse("2006-01-02", shiftDateStr)
	shift.StartTime, _ = time.Parse("15:04", startTimeStr)
	shift.EndTime, _ = time.Parse("15:04", endTimeStr)
	shift.IsHoliday = isHolidayInt == 1
	shift.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	shift.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return &shift, nil
}

func ListShifts() ([]models.ShiftMaster, error) {
	rows, err := database.DB.Query(`
		SELECT id, employee_id, shift_date, start_time, end_time, position, status,
		       scheduled_hours, actual_hours, is_holiday, created_at, updated_at
		FROM shift_masters ORDER BY shift_date DESC, employee_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.ShiftMaster
	for rows.Next() {
		var s models.ShiftMaster
		var shiftDateStr, startTimeStr, endTimeStr, createdAtStr, updatedAtStr string
		var isHolidayInt int

		if err := rows.Scan(
			&s.ID, &s.EmployeeID, &shiftDateStr, &startTimeStr, &endTimeStr,
			&s.Position, &s.Status, &s.ScheduledHours, &s.ActualHours,
			&isHolidayInt, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, err
		}

		s.ShiftDate, _ = time.Parse("2006-01-02", shiftDateStr)
		s.StartTime, _ = time.Parse("15:04", startTimeStr)
		s.EndTime, _ = time.Parse("15:04", endTimeStr)
		s.IsHoliday = isHolidayInt == 1
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		shifts = append(shifts, s)
	}
	return shifts, nil
}

func GetShiftByEmployeeAndDate(empID int64, date time.Time) (*models.ShiftMaster, error) {
	var shift models.ShiftMaster
	var shiftDateStr, startTimeStr, endTimeStr, createdAtStr, updatedAtStr string
	var isHolidayInt int

	err := database.DB.QueryRow(`
		SELECT id, employee_id, shift_date, start_time, end_time, position, status,
		       scheduled_hours, actual_hours, is_holiday, created_at, updated_at
		FROM shift_masters WHERE employee_id = ? AND shift_date = ?
	`, empID, date.Format("2006-01-02")).Scan(
		&shift.ID, &shift.EmployeeID, &shiftDateStr, &startTimeStr, &endTimeStr,
		&shift.Position, &shift.Status, &shift.ScheduledHours, &shift.ActualHours,
		&isHolidayInt, &createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	shift.ShiftDate, _ = time.Parse("2006-01-02", shiftDateStr)
	shift.StartTime, _ = time.Parse("15:04", startTimeStr)
	shift.EndTime, _ = time.Parse("15:04", endTimeStr)
	shift.IsHoliday = isHolidayInt == 1
	shift.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	shift.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return &shift, nil
}

func GetLastShiftBefore(empID int64, date time.Time) (*models.ShiftMaster, error) {
	var shift models.ShiftMaster
	var shiftDateStr, startTimeStr, endTimeStr, createdAtStr, updatedAtStr string
	var isHolidayInt int

	err := database.DB.QueryRow(`
		SELECT id, employee_id, shift_date, start_time, end_time, position, status,
		       scheduled_hours, actual_hours, is_holiday, created_at, updated_at
		FROM shift_masters WHERE employee_id = ? AND shift_date < ?
		ORDER BY shift_date DESC, end_time DESC LIMIT 1
	`, empID, date.Format("2006-01-02")).Scan(
		&shift.ID, &shift.EmployeeID, &shiftDateStr, &startTimeStr, &endTimeStr,
		&shift.Position, &shift.Status, &shift.ScheduledHours, &shift.ActualHours,
		&isHolidayInt, &createdAtStr, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	shift.ShiftDate, _ = time.Parse("2006-01-02", shiftDateStr)
	shift.StartTime, _ = time.Parse("15:04", startTimeStr)
	shift.EndTime, _ = time.Parse("15:04", endTimeStr)
	shift.IsHoliday = isHolidayInt == 1
	shift.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	shift.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return &shift, nil
}

func GetWeeklyHours(empID int64, weekStart, weekEnd time.Time) (float64, error) {
	var totalHours float64
	err := database.DB.QueryRow(`
		SELECT COALESCE(SUM(scheduled_hours), 0) FROM shift_masters
		WHERE employee_id = ? AND shift_date >= ? AND shift_date <= ?
	`, empID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02")).Scan(&totalHours)
	if err != nil {
		return 0, err
	}
	return totalHours, nil
}

func UpdateShiftStatus(shiftID int64, newStatus models.ShiftStatus, operatorID int64, note string) error {
	shift, err := GetShiftByID(shiftID)
	if err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE shift_masters SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, newStatus, shiftID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO shift_details (master_id, employee_id, shift_date, start_time, end_time, position, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, shiftID, shift.EmployeeID, shift.ShiftDate.Format("2006-01-02"), shift.StartTime.Format("15:04"), shift.EndTime.Format("15:04"), shift.Position, newStatus)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO operation_history (master_id, operator_id, action, from_status, to_status, note)
		VALUES (?, ?, ?, ?, ?, ?)
	`, shiftID, operatorID, "status_change", string(shift.Status), string(newStatus), note)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func UpdateShiftActualHours(shiftID int64, actualHours float64) error {
	_, err := database.DB.Exec(`
		UPDATE shift_masters SET actual_hours = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, actualHours, shiftID)
	return err
}

func GetOperationHistory(masterID int64) ([]models.OperationHistory, error) {
	rows, err := database.DB.Query(`
		SELECT id, master_id, operator_id, action, from_status, to_status, note, created_at
		FROM operation_history WHERE master_id = ? ORDER BY created_at DESC
	`, masterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []models.OperationHistory
	for rows.Next() {
		var h models.OperationHistory
		var createdAtStr string
		if err := rows.Scan(&h.ID, &h.MasterID, &h.OperatorID, &h.Action, &h.FromStatus, &h.ToStatus, &h.Note, &createdAtStr); err != nil {
			return nil, err
		}
		h.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		histories = append(histories, h)
	}
	return histories, nil
}

func AddNoteToHistory(historyID int64, note string) error {
	_, err := database.DB.Exec(`
		UPDATE operation_history SET note = COALESCE(note, '') || CHAR(10) || ? WHERE id = ?
	`, note, historyID)
	return err
}

func CreateSwapRequest(requesterShiftID, responderShiftID int64) (*models.SwapRequest, error) {
	requesterShift, err := GetShiftByID(requesterShiftID)
	if err != nil {
		return nil, fmt.Errorf("requester shift not found: %w", err)
	}

	responderShift, err := GetShiftByID(responderShiftID)
	if err != nil {
		return nil, fmt.Errorf("responder shift not found: %w", err)
	}

	result, err := database.DB.Exec(`
		INSERT INTO swap_requests (requester_shift_id, responder_shift_id, requester_id, responder_id, status)
		VALUES (?, ?, ?, ?, 'pending')
	`, requesterShiftID, responderShiftID, requesterShift.EmployeeID, responderShift.EmployeeID)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.SwapRequest{
		ID:               id,
		RequesterShiftID: requesterShiftID,
		ResponderShiftID: responderShiftID,
		RequesterID:      requesterShift.EmployeeID,
		ResponderID:      responderShift.EmployeeID,
		Status:           "pending",
		CreatedAt:        time.Now(),
	}, nil
}

func GetSwapRequestByID(id int64) (*models.SwapRequest, error) {
	var swap models.SwapRequest
	var createdAtStr string
	var confirmedAt sql.NullString

	err := database.DB.QueryRow(`
		SELECT id, requester_shift_id, responder_shift_id, requester_id, responder_id, status, created_at, confirmed_at
		FROM swap_requests WHERE id = ?
	`, id).Scan(
		&swap.ID, &swap.RequesterShiftID, &swap.ResponderShiftID,
		&swap.RequesterID, &swap.ResponderID, &swap.Status,
		&createdAtStr, &confirmedAt,
	)
	if err != nil {
		return nil, err
	}

	swap.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if confirmedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", confirmedAt.String)
		swap.ConfirmedAt = &t
	}

	return &swap, nil
}

func UpdateSwapRequestStatus(id int64, status string) error {
	if status == "confirmed" {
		_, err := database.DB.Exec(`
			UPDATE swap_requests SET status = ?, confirmed_at = CURRENT_TIMESTAMP WHERE id = ?
		`, status, id)
		return err
	}
	_, err := database.DB.Exec(`
		UPDATE swap_requests SET status = ? WHERE id = ?
	`, status, id)
	return err
}

func ExecuteSwap(swapID int64) error {
	swap, err := GetSwapRequestByID(swapID)
	if err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	reqShift, err := GetShiftByID(swap.RequesterShiftID)
	if err != nil {
		return err
	}

	respShift, err := GetShiftByID(swap.ResponderShiftID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE shift_masters SET employee_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, swap.ResponderID, swap.RequesterShiftID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE shift_masters SET employee_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, swap.RequesterID, swap.ResponderShiftID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO shift_details (master_id, employee_id, shift_date, start_time, end_time, position, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, swap.RequesterShiftID, swap.ResponderID, reqShift.ShiftDate.Format("2006-01-02"), reqShift.StartTime.Format("15:04"), reqShift.EndTime.Format("15:04"), reqShift.Position, reqShift.Status)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO shift_details (master_id, employee_id, shift_date, start_time, end_time, position, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, swap.ResponderShiftID, swap.RequesterID, respShift.ShiftDate.Format("2006-01-02"), respShift.StartTime.Format("15:04"), respShift.EndTime.Format("15:04"), respShift.Position, respShift.Status)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetMonthlyStats(year, month int) ([]models.MonthlyStats, error) {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, -1)

	rows, err := database.DB.Query(`
		SELECT 
			e.id,
			e.name,
			d.id,
			d.name,
			COALESCE(SUM(sm.scheduled_hours), 0),
			COALESCE(SUM(sm.actual_hours), 0)
		FROM employees e
		JOIN departments d ON e.department_id = d.id
		LEFT JOIN shift_masters sm ON e.id = sm.employee_id 
			AND sm.shift_date BETWEEN ? AND ?
		GROUP BY e.id, e.name, d.id, d.name
		ORDER BY d.id, e.id
	`, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.MonthlyStats
	for rows.Next() {
		var s models.MonthlyStats
		if err := rows.Scan(&s.EmployeeID, &s.EmployeeName, &s.DepartmentID, &s.DepartmentName, &s.ScheduledHours, &s.ActualHours); err != nil {
			return nil, err
		}
		s.Month = fmt.Sprintf("%04d-%02d", year, month)
		s.Difference = s.ActualHours - s.ScheduledHours
		stats = append(stats, s)
	}
	return stats, nil
}

func GetDepartmentMonthlyStats(deptID, year, month int) ([]models.MonthlyStats, error) {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, -1)

	rows, err := database.DB.Query(`
		SELECT 
			e.id,
			e.name,
			d.id,
			d.name,
			COALESCE(SUM(sm.scheduled_hours), 0),
			COALESCE(SUM(sm.actual_hours), 0)
		FROM employees e
		JOIN departments d ON e.department_id = d.id
		LEFT JOIN shift_masters sm ON e.id = sm.employee_id 
			AND sm.shift_date BETWEEN ? AND ?
		WHERE d.id = ?
		GROUP BY e.id, e.name, d.id, d.name
		ORDER BY e.id
	`, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"), deptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.MonthlyStats
	for rows.Next() {
		var s models.MonthlyStats
		if err := rows.Scan(&s.EmployeeID, &s.EmployeeName, &s.DepartmentID, &s.DepartmentName, &s.ScheduledHours, &s.ActualHours); err != nil {
			return nil, err
		}
		s.Month = fmt.Sprintf("%04d-%02d", year, month)
		s.Difference = s.ActualHours - s.ScheduledHours
		stats = append(stats, s)
	}
	return stats, nil
}
