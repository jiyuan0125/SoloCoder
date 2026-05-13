package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

const DBFile = "./data/audit_log.db"

func InitDB() error {
	dbDir := filepath.Dir(DBFile)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", DBFile+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operator TEXT NOT NULL,
		resource_type TEXT NOT NULL,
		operation TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		before_value TEXT,
		after_value TEXT,
		timestamp DATETIME NOT NULL,
		archived BOOLEAN DEFAULT 0,
		archive_file TEXT
	);
	
	CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_operator ON audit_logs(operator);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_archived ON audit_logs(archived);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_composite ON audit_logs(timestamp, operator, resource_type);
	`

	_, err := DB.Exec(createTableSQL)
	return err
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
