package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func InitSchema(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS plans (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		max_users INTEGER NOT NULL DEFAULT 0,
		max_storage_mb INTEGER NOT NULL DEFAULT 0,
		max_api_calls INTEGER NOT NULL DEFAULT 0,
		features TEXT NOT NULL DEFAULT '[]',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tenants (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		plan_id TEXT NOT NULL,
		pending_plan_id TEXT,
		cycle_start_at TIMESTAMP NOT NULL,
		cycle_end_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES plans(id)
	);

	CREATE TABLE IF NOT EXISTS feature_switches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id TEXT NOT NULL,
		feature_key TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		UNIQUE(tenant_id, feature_key),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS custom_fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id TEXT NOT NULL,
		field_name TEXT NOT NULL,
		field_type TEXT NOT NULL,
		options TEXT,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		UNIQUE(tenant_id, field_name),
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS custom_field_values (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		field_id INTEGER NOT NULL,
		record_id TEXT NOT NULL,
		value TEXT,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (field_id) REFERENCES custom_fields(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS usage_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id TEXT NOT NULL,
		resource_type TEXT NOT NULL,
		amount INTEGER NOT NULL DEFAULT 0,
		period_start TIMESTAMP NOT NULL,
		period_end TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		UNIQUE(tenant_id, resource_type, period_start)
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id TEXT NOT NULL,
		message TEXT NOT NULL,
		level TEXT NOT NULL,
		read INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL
	);
	`
	_, err := db.Exec(stmt)
	return err
}

func Now() time.Time {
	return time.Now().UTC()
}
