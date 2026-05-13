package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./performance.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	createTables()
}

func createTables() {
	createEmployeeTable := `
	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		team_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createQuarterTable := `
	CREATE TABLE IF NOT EXISTS quarters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		quarter INTEGER NOT NULL CHECK (quarter IN (1, 2, 3, 4)),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(year, quarter)
	);`

	createReviewTable := `
	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee_id INTEGER NOT NULL,
		quarter_id INTEGER NOT NULL,
		quality REAL NOT NULL,
		efficiency REAL NOT NULL,
		collaboration REAL NOT NULL,
		innovation REAL NOT NULL,
		total_score REAL,
		level TEXT,
		final_level TEXT,
		status TEXT DEFAULT 'draft',
		need_improvement INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (employee_id) REFERENCES employees(id),
		FOREIGN KEY (quarter_id) REFERENCES quarters(id),
		UNIQUE(employee_id, quarter_id)
	);`

	createAnnualReviewTable := `
	CREATE TABLE IF NOT EXISTS annual_reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		employee_id INTEGER NOT NULL,
		year INTEGER NOT NULL,
		average_score REAL,
		level TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (employee_id) REFERENCES employees(id),
		UNIQUE(employee_id, year)
	);`

	statements := []string{createEmployeeTable, createQuarterTable, createReviewTable, createAnnualReviewTable}

	for _, stmt := range statements {
		_, err := DB.Exec(stmt)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}
