package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

const DBPath = "reimburse.db"

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite3", DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	createReimbursementTable := `
	CREATE TABLE IF NOT EXISTS reimbursements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee_id TEXT NOT NULL,
		amount_cent INTEGER NOT NULL,
		expense_type TEXT NOT NULL,
		occurred_date TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		current_approver TEXT NOT NULL,
		modify_count INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	`

	createHistoryTable := `
	CREATE TABLE IF NOT EXISTS status_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reimbursement_id INTEGER NOT NULL,
		from_status TEXT,
		to_status TEXT NOT NULL,
		changed_by TEXT NOT NULL,
		changed_at TEXT NOT NULL,
		note TEXT,
		FOREIGN KEY(reimbursement_id) REFERENCES reimbursements(id)
	);
	`

	createIndex := `
	CREATE INDEX IF NOT EXISTS idx_employee_expense ON reimbursements(employee_id, occurred_date, amount_cent, expense_type);
	CREATE INDEX IF NOT EXISTS idx_status ON reimbursements(status);
	CREATE INDEX IF NOT EXISTS idx_created_at ON reimbursements(created_at);
	`

	if _, err := DB.Exec(createReimbursementTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createHistoryTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createIndex); err != nil {
		return err
	}

	return nil
}

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
