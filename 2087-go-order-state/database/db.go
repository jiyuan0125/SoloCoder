package database

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	err = migrate(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no TEXT UNIQUE NOT NULL,
			resource_id INTEGER,
			amount REAL NOT NULL,
			current_status TEXT NOT NULL,
			is_cancelled BOOLEAN DEFAULT 0,
			cancellation_type TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (resource_id) REFERENCES resources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS order_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			operator TEXT NOT NULL,
			reason TEXT NOT NULL,
			resource_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (resource_id) REFERENCES resources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS aftersales (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			resource_id INTEGER,
			type TEXT NOT NULL,
			current_status TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (resource_id) REFERENCES resources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS aftersales_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			aftersales_id INTEGER NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			operator TEXT NOT NULL,
			reason TEXT NOT NULL,
			resource_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (aftersales_id) REFERENCES aftersales(id),
			FOREIGN KEY (resource_id) REFERENCES resources(id)
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}
