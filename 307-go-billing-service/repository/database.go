package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

func InitDB(dbPath string) (*Database, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := database.initDefaultData(); err != nil {
		return nil, fmt.Errorf("failed to init default data: %w", err)
	}

	return database, nil
}

func (db *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS packages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			price_monthly REAL NOT NULL,
			sms_quota INTEGER NOT NULL,
			storage_quota INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			current_package TEXT NOT NULL,
			package_start_date TIMESTAMP NOT NULL,
			package_end_date TIMESTAMP,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS package_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			old_package TEXT NOT NULL,
			new_package TEXT NOT NULL,
			change_date TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
		`CREATE TABLE IF NOT EXISTS usages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL,
			sms_used INTEGER NOT NULL DEFAULT 0,
			storage_used REAL NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id),
			UNIQUE(customer_id, year, month)
		)`,
		`CREATE TABLE IF NOT EXISTS bills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			bill_year INTEGER NOT NULL,
			bill_month INTEGER NOT NULL,
			package_fee REAL NOT NULL,
			sms_fee REAL NOT NULL,
			storage_fee REAL NOT NULL,
			total_amount REAL NOT NULL,
			status TEXT NOT NULL,
			due_date TIMESTAMP NOT NULL,
			paid_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers(id),
			UNIQUE(customer_id, bill_year, bill_month)
		)`,
		`CREATE TABLE IF NOT EXISTS bill_payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bill_id INTEGER NOT NULL,
			operator TEXT NOT NULL,
			remark TEXT,
			paid_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (bill_id) REFERENCES bills(id)
		)`,
		`CREATE TABLE IF NOT EXISTS configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bills_customer ON bills(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bills_status ON bills(status)`,
		`CREATE INDEX IF NOT EXISTS idx_usages_customer_year_month ON usages(customer_id, year, month)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (db *Database) initDefaultData() error {
	defaultPackages := []struct {
		pkgType     string
		name        string
		price       float64
		smsQuota    int
		storageQuota int
	}{
		{"basic", "基础版", 99.0, 1000, 10},
		{"pro", "专业版", 299.0, 5000, 50},
		{"enterprise", "企业版", 899.0, 20000, 200},
	}

	for _, pkg := range defaultPackages {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO packages (type, name, price_monthly, sms_quota, storage_quota)
			VALUES (?, ?, ?, ?, ?)
		`, pkg.pkgType, pkg.name, pkg.price, pkg.smsQuota, pkg.storageQuota)
		if err != nil {
			return err
		}
	}

	defaultConfigs := []struct {
		key         string
		value       string
		description string
	}{
		{"sms.unit_price", "0.1", "短信单价（元/条）"},
		{"storage.unit_price", "5.0", "云存储单价（元/GB/月）"},
		{"payment.due_days", "15", "账单付款期限（天）"},
		{"overdue.severe_days", "60", "严重逾期天数（天）"},
	}

	for _, cfg := range defaultConfigs {
		now := time.Now()
		_, err := db.Exec(`
			INSERT OR IGNORE INTO configs (key, value, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, cfg.key, cfg.value, cfg.description, now, now)
		if err != nil {
			return err
		}
	}

	return nil
}
