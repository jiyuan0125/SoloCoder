package database

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init() error {
	dbPath := filepath.Join(".", "shift_scheduler.db")

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func runMigrations() error {
	schema := `
	CREATE TABLE IF NOT EXISTS departments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		department_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (department_id) REFERENCES departments(id)
	);

	CREATE TABLE IF NOT EXISTS holidays (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS shift_masters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee_id INTEGER NOT NULL,
		shift_date TEXT NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		position TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		scheduled_hours REAL NOT NULL DEFAULT 0,
		actual_hours REAL NOT NULL DEFAULT 0,
		is_holiday INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (employee_id) REFERENCES employees(id),
		UNIQUE(employee_id, shift_date)
	);

	CREATE TABLE IF NOT EXISTS shift_details (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		master_id INTEGER NOT NULL,
		employee_id INTEGER NOT NULL,
		shift_date TEXT NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		position TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (master_id) REFERENCES shift_masters(id),
		FOREIGN KEY (employee_id) REFERENCES employees(id)
	);

	CREATE TABLE IF NOT EXISTS operation_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		master_id INTEGER NOT NULL,
		operator_id INTEGER NOT NULL,
		action TEXT NOT NULL,
		from_status TEXT NOT NULL,
		to_status TEXT NOT NULL,
		note TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (master_id) REFERENCES shift_masters(id)
	);

	CREATE TABLE IF NOT EXISTS swap_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		requester_shift_id INTEGER NOT NULL,
		responder_shift_id INTEGER NOT NULL,
		requester_id INTEGER NOT NULL,
		responder_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		confirmed_at DATETIME,
		FOREIGN KEY (requester_shift_id) REFERENCES shift_masters(id),
		FOREIGN KEY (responder_shift_id) REFERENCES shift_masters(id),
		FOREIGN KEY (requester_id) REFERENCES employees(id),
		FOREIGN KEY (responder_id) REFERENCES employees(id)
	);

	CREATE INDEX IF NOT EXISTS idx_shift_masters_employee_date ON shift_masters(employee_id, shift_date);
	CREATE INDEX IF NOT EXISTS idx_operation_history_master ON operation_history(master_id);
	CREATE INDEX IF NOT EXISTS idx_swap_requests_status ON swap_requests(status);
	`

	_, err := DB.Exec(schema)
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
