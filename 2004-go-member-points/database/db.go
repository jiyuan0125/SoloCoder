package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			level TEXT NOT NULL DEFAULT 'normal',
			points INTEGER NOT NULL DEFAULT 0,
			yearly_points INTEGER NOT NULL DEFAULT 0,
			year INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS consumptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			amount INTEGER NOT NULL,
			points_earned INTEGER NOT NULL,
			multiplier REAL NOT NULL DEFAULT 1.0,
			processed BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (member_id) REFERENCES members(id)
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			points_cost INTEGER NOT NULL,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS exchanges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			points_used INTEGER NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (member_id) REFERENCES members(id),
			FOREIGN KEY (product_id) REFERENCES products(id),
			UNIQUE(member_id, product_id)
		)`,
		`CREATE TABLE IF NOT EXISTS year_end_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			year INTEGER NOT NULL,
			total_points INTEGER NOT NULL,
			old_level TEXT NOT NULL,
			new_level TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (member_id) REFERENCES members(id),
			UNIQUE(member_id, year)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_consumptions_member_id ON consumptions(member_id)`,
		`CREATE INDEX IF NOT EXISTS idx_exchanges_member_product ON exchanges(member_id, product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_exchanges_member_id ON exchanges(member_id)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %v, error: %v", query, err)
		}
	}

	return nil
}
