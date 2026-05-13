package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS players (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			player_id TEXT NOT NULL,
			dimension TEXT NOT NULL,
			dimension_key TEXT NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(player_id, dimension, dimension_key)
		)`,
		`CREATE TABLE IF NOT EXISTS history_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			player_id TEXT NOT NULL,
			highest_score INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(player_id)
		)`,
		`CREATE TABLE IF NOT EXISTS plans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			total_amount REAL NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS plan_periods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plan_id INTEGER NOT NULL,
			period_index INTEGER NOT NULL,
			planned_amount REAL NOT NULL DEFAULT 0,
			settled_amount REAL NOT NULL DEFAULT 0,
			is_settled INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(plan_id) REFERENCES plans(id),
			UNIQUE(plan_id, period_index)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_scores_dimension ON scores(dimension, dimension_key)`,
		`CREATE INDEX IF NOT EXISTS idx_scores_score ON scores(score DESC, submitted_at ASC)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", query, err)
		}
	}

	return nil
}
