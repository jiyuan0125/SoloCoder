package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
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
	createCouponsTable := `
	CREATE TABLE IF NOT EXISTS coupons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id TEXT NOT NULL UNIQUE,
		denomination INTEGER NOT NULL,
		threshold INTEGER NOT NULL,
		total_count INTEGER NOT NULL,
		claimed_count INTEGER NOT NULL DEFAULT 0,
		used_count INTEGER NOT NULL DEFAULT 0,
		limit_per_user INTEGER NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	createUserCouponsTable := `
	CREATE TABLE IF NOT EXISTS user_coupons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		coupon_id INTEGER NOT NULL,
		batch_id TEXT NOT NULL,
		status TEXT NOT NULL,
		claimed_at DATETIME NOT NULL,
		redeemed_at DATETIME,
		order_id TEXT,
		redeemed_amount INTEGER,
		FOREIGN KEY (coupon_id) REFERENCES coupons(id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_coupons_user_id ON user_coupons(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_coupons_coupon_id ON user_coupons(coupon_id);
	CREATE INDEX IF NOT EXISTS idx_user_coupons_batch_id ON user_coupons(batch_id);`

	createRedemptionRecordsTable := `
	CREATE TABLE IF NOT EXISTS redemption_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_coupon_id INTEGER NOT NULL,
		coupon_id INTEGER NOT NULL,
		batch_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		order_id TEXT NOT NULL,
		order_amount INTEGER NOT NULL,
		redeemed_amount INTEGER NOT NULL,
		redeemed_at DATETIME NOT NULL,
		is_refunded BOOLEAN NOT NULL DEFAULT 0,
		FOREIGN KEY (user_coupon_id) REFERENCES user_coupons(id),
		FOREIGN KEY (coupon_id) REFERENCES coupons(id)
	);
	CREATE INDEX IF NOT EXISTS idx_redemption_batch_id ON redemption_records(batch_id);
	CREATE INDEX IF NOT EXISTS idx_redemption_order_id ON redemption_records(order_id);`

	createStatsTable := `
	CREATE TABLE IF NOT EXISTS coupon_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		batch_id TEXT NOT NULL UNIQUE,
		total_count INTEGER NOT NULL,
		claimed_count INTEGER NOT NULL,
		used_count INTEGER NOT NULL,
		usage_rate REAL NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_stats_batch_id ON coupon_stats(batch_id);`

	tables := []string{createCouponsTable, createUserCouponsTable, createRedemptionRecordsTable, createStatsTable}

	for _, tableSQL := range tables {
		_, err := DB.Exec(tableSQL)
		if err != nil {
			return err
		}
	}

	return nil
}
