package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = initTables(); err != nil {
		return fmt.Errorf("failed to init tables: %w", err)
	}

	return nil
}

func initTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			manager_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (manager_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS approval_chains (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS approval_nodes (
			id TEXT PRIMARY KEY,
			chain_id TEXT NOT NULL,
			level INTEGER NOT NULL,
			node_type TEXT NOT NULL,
			approver_ids TEXT NOT NULL,
			condition TEXT,
			is_sign_all INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (chain_id) REFERENCES approval_chains(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS applications (
			id TEXT PRIMARY KEY,
			chain_id TEXT NOT NULL,
			applicant_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			data TEXT NOT NULL,
			current_level INTEGER DEFAULT 0,
			current_approver_ids TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			reject_reason TEXT,
			submission_count INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (chain_id) REFERENCES approval_chains(id),
			FOREIGN KEY (applicant_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS approval_operations (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL,
			operator_id TEXT NOT NULL,
			level INTEGER NOT NULL,
			operation TEXT NOT NULL,
			reason TEXT,
			target_user_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (application_id) REFERENCES applications(id),
			FOREIGN KEY (operator_id) REFERENCES users(id),
			FOREIGN KEY (target_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			application_id TEXT NOT NULL,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			read INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (application_id) REFERENCES applications(id)
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id TEXT PRIMARY KEY,
			report_date TEXT NOT NULL,
			chain_id TEXT NOT NULL,
			total_amount REAL DEFAULT 0,
			approved_count INTEGER DEFAULT 0,
			rejected_count INTEGER DEFAULT 0,
			pending_count INTEGER DEFAULT 0,
			detail_stats TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS timeouts (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL,
			level INTEGER NOT NULL,
			reminded_48h INTEGER DEFAULT 0,
			transferred_72h INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (application_id) REFERENCES applications(id)
		)`,
	}

	for _, stmt := range tables {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute table creation: %w", err)
		}
	}

	return nil
}
