package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createTables(); err != nil {
		return err
	}

	return initDefaultPosition()
}

func createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			total_quota INTEGER NOT NULL DEFAULT 10,
			used_quota INTEGER NOT NULL DEFAULT 0,
			pass_score REAL NOT NULL DEFAULT 60,
			tech_threshold REAL NOT NULL DEFAULT 70,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS candidates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			phone TEXT,
			position_id INTEGER NOT NULL,
			current_stage TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			last_rejected_at TIMESTAMP,
			FOREIGN KEY (position_id) REFERENCES positions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS stage_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			candidate_id INTEGER NOT NULL,
			stage TEXT NOT NULL,
			owner TEXT NOT NULL,
			due_date TIMESTAMP NOT NULL,
			score REAL,
			status TEXT,
			remark TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (candidate_id) REFERENCES candidates(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tech_interviewer_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			candidate_id INTEGER NOT NULL,
			interviewer TEXT NOT NULL,
			score REAL,
			submitted INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (candidate_id) REFERENCES candidates(id)
		)`,
		`CREATE TABLE IF NOT EXISTS offers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			candidate_id INTEGER NOT NULL,
			position_id INTEGER NOT NULL,
			valid_until TIMESTAMP NOT NULL,
			accepted INTEGER NOT NULL DEFAULT 0,
			cancelled INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (candidate_id) REFERENCES candidates(id),
			FOREIGN KEY (position_id) REFERENCES positions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS history_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			candidate_id INTEGER NOT NULL,
			stage TEXT NOT NULL,
			action TEXT NOT NULL,
			operator TEXT NOT NULL,
			remark TEXT,
			score REAL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (candidate_id) REFERENCES candidates(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_candidates_position ON candidates(position_id, current_stage)`,
		`CREATE INDEX IF NOT EXISTS idx_candidates_email ON candidates(email, position_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stage_records_candidate ON stage_records(candidate_id, stage)`,
		`CREATE INDEX IF NOT EXISTS idx_history_candidate ON history_records(candidate_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_offers_candidate ON offers(candidate_id)`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return nil
}

func initDefaultPosition() error {
	now := time.Now()
	result, err := DB.Exec(`INSERT OR IGNORE INTO positions (id, name, total_quota, used_quota, pass_score, tech_threshold, created_at, updated_at) VALUES (1, ?, 10, 0, 60, 70, ?, ?)`,
		"Software Engineer", now, now)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		fmt.Println("Created default position: Software Engineer")
	}
	return nil
}
