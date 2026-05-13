package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

func Init(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_name TEXT NOT NULL,
			original_price DECIMAL(10,2) NOT NULL,
			flash_price DECIMAL(10,2) NOT NULL,
			total_stock INTEGER NOT NULL,
			start_time DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activities_start_time ON activities(start_time)`,

		`CREATE TABLE IF NOT EXISTS stocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL UNIQUE,
			available_stock INTEGER NOT NULL,
			version INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no TEXT NOT NULL UNIQUE,
			activity_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			paid_at DATETIME,
			cancelled_at DATETIME,
			expire_at DATETIME NOT NULL,
			FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
			UNIQUE(activity_id, user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_activity_user ON orders(activity_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_expire_at ON orders(expire_at)`,

		`CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL UNIQUE,
			total_orders INTEGER NOT NULL DEFAULT 0,
			paid_orders INTEGER NOT NULL DEFAULT 0,
			unpaid_orders INTEGER NOT NULL DEFAULT 0,
			cancelled_orders INTEGER NOT NULL DEFAULT 0,
			total_revenue DECIMAL(10,2) NOT NULL DEFAULT 0,
			generated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			verified_at DATETIME,
			FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS cache_notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			module_name TEXT NOT NULL,
			event_type TEXT NOT NULL,
			activity_id INTEGER,
			payload TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			processed_at DATETIME
		)`,

		`CREATE TABLE IF NOT EXISTS quota_allocations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			segment_name TEXT NOT NULL,
			total_quota INTEGER NOT NULL,
			allocated_quota INTEGER NOT NULL DEFAULT 0,
			ratio DECIMAL(5,4) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
			UNIQUE(activity_id, segment_name)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
