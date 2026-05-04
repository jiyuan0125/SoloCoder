package dao

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

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

	return nil
}

func createTables() error {
	createCouponBatchTable := `
	CREATE TABLE IF NOT EXISTS coupon_batches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		discount_amount INTEGER NOT NULL,
		threshold_amount INTEGER NOT NULL,
		total_quantity INTEGER NOT NULL,
		issued_quantity INTEGER NOT NULL DEFAULT 0,
		redeemed_quantity INTEGER NOT NULL DEFAULT 0,
		valid_start DATETIME NOT NULL,
		valid_end DATETIME NOT NULL,
		limit_per_user INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	createCouponTable := `
	CREATE TABLE IF NOT EXISTS coupons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		code TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL DEFAULT 'issued',
		redeemed_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (batch_id) REFERENCES coupon_batches(id)
	);`

	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_coupons_batch_id ON coupons(batch_id);
	CREATE INDEX IF NOT EXISTS idx_coupons_user_id ON coupons(user_id);
	CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
	CREATE INDEX IF NOT EXISTS idx_coupons_status ON coupons(status);
	CREATE INDEX IF NOT EXISTS idx_coupon_batches_name ON coupon_batches(name);
	`

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(createCouponBatchTable); err != nil {
		return fmt.Errorf("failed to create coupon_batches table: %w", err)
	}

	if _, err = tx.Exec(createCouponTable); err != nil {
		return fmt.Errorf("failed to create coupons table: %w", err)
	}

	if _, err = tx.Exec(createIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
