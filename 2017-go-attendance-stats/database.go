package main

import (
	"database/sql"
	"time"
)

var db *sql.DB

func initDB(dbPath string) {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS attendance_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee_id INTEGER NOT NULL,
		punch_time DATETIME NOT NULL,
		punch_type TEXT NOT NULL CHECK(punch_type IN ('in', 'out')),
		status TEXT NOT NULL DEFAULT 'normal',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (employee_id) REFERENCES employees(id)
	);

	CREATE INDEX IF NOT EXISTS idx_attendance_employee_date ON attendance_records(employee_id, date(punch_time), punch_type);

	CREATE TABLE IF NOT EXISTS makeup_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		punch_type TEXT NOT NULL CHECK(punch_type IN ('in', 'out')),
		reason TEXT,
		status TEXT NOT NULL DEFAULT 'draft',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (employee_id) REFERENCES employees(id)
	);

	CREATE INDEX IF NOT EXISTS idx_makeup_employee_date ON makeup_requests(employee_id, date, status);
	`

	if _, err := db.Exec(schema); err != nil {
		panic(err)
	}
}

func initTestData() {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM employees").Scan(&count)
	if count > 0 {
		return
	}

	employees := []string{"张三", "李四", "王五", "赵六"}
	for _, name := range employees {
		db.Exec("INSERT INTO employees (name) VALUES (?)", name)
	}
}

func getEmployeeByID(id int) (*Employee, error) {
	emp := &Employee{}
	err := db.QueryRow(
		"SELECT id, name, created_at FROM employees WHERE id = ?", id,
	).Scan(&emp.ID, &emp.Name, &emp.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return emp, err
}

func getAllEmployees() ([]Employee, error) {
	rows, err := db.Query("SELECT id, name, created_at FROM employees ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.CreatedAt); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}

func hasPunchedToday(employeeID int, punchType string, date time.Time) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM attendance_records 
		 WHERE employee_id = ? AND punch_type = ? AND date(punch_time) = date(?)`,
		employeeID, punchType, date,
	).Scan(&count)
	return count > 0, err
}

func insertAttendanceRecord(record *AttendanceRecord) error {
	result, err := db.Exec(
		`INSERT INTO attendance_records (employee_id, punch_time, punch_type, status)
		VALUES (?, ?, ?, ?)`,
		record.EmployeeID, record.PunchTime, record.PunchType, record.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	record.ID = int(id)
	return nil
}

func getAttendanceByEmployeeAndDate(employeeID int, date time.Time) ([]AttendanceRecord, error) {
	rows, err := db.Query(
		`SELECT id, employee_id, punch_time, punch_type, status, created_at 
		 FROM attendance_records 
		 WHERE employee_id = ? AND date(punch_time) = date(?)
		 ORDER BY punch_time`,
		employeeID, date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AttendanceRecord
	for rows.Next() {
		var r AttendanceRecord
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.PunchTime, &r.PunchType, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func getAllAttendanceRecords() ([]AttendanceRecord, error) {
	rows, err := db.Query(
		`SELECT id, employee_id, punch_time, punch_type, status, created_at 
		 FROM attendance_records 
		 ORDER BY punch_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AttendanceRecord
	for rows.Next() {
		var r AttendanceRecord
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.PunchTime, &r.PunchType, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func getAttendanceRecordByID(id int) (*AttendanceRecord, error) {
	r := &AttendanceRecord{}
	err := db.QueryRow(
		`SELECT id, employee_id, punch_time, punch_type, status, created_at 
		 FROM attendance_records WHERE id = ?`, id,
	).Scan(&r.ID, &r.EmployeeID, &r.PunchTime, &r.PunchType, &r.Status, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func getMakeupRequestByID(id int) (*MakeupRequest, error) {
	m := &MakeupRequest{}
	err := db.QueryRow(
		`SELECT id, employee_id, date, punch_type, reason, status, created_at, updated_at 
		 FROM makeup_requests WHERE id = ?`, id,
	).Scan(&m.ID, &m.EmployeeID, &m.Date, &m.PunchType, &m.Reason, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func getMakeupRequests() ([]MakeupRequest, error) {
	rows, err := db.Query(
		`SELECT id, employee_id, date, punch_type, reason, status, created_at, updated_at 
		 FROM makeup_requests ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []MakeupRequest
	for rows.Next() {
		var r MakeupRequest
		if err := rows.Scan(&r.ID, &r.EmployeeID, &r.Date, &r.PunchType, &r.Reason, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, nil
}

func hasMakeupToday(employeeID int, date string) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM makeup_requests 
		 WHERE employee_id = ? AND date = ? AND status IN ('draft', 'approved', 'executed', 'confirmed')`,
		employeeID, date,
	).Scan(&count)
	return count > 0, err
}

func insertMakeupRequest(req *MakeupRequest) error {
	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now
	result, err := db.Exec(
		`INSERT INTO makeup_requests (employee_id, date, punch_type, reason, status)
		VALUES (?, ?, ?, ?, ?)`,
		req.EmployeeID, req.Date, req.PunchType, req.Reason, req.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	req.ID = int(id)
	return nil
}

func updateMakeupStatus(id int, status string) error {
	_, err := db.Exec(
		`UPDATE makeup_requests SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	return err
}
