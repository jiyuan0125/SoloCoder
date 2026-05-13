package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = createTables(); err != nil {
		return err
	}

	return initSeedData()
}

func createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			code TEXT UNIQUE NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'normal',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS check_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			max_score INTEGER NOT NULL DEFAULT 100
		)`,
		`CREATE TABLE IF NOT EXISTS inspection_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			order_index INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (device_id) REFERENCES devices(id)
		)`,
		`CREATE TABLE IF NOT EXISTS inspection_plans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			frequency TEXT NOT NULL,
			start_date DATETIME NOT NULL,
			last_generate_at DATETIME,
			active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS plan_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plan_id INTEGER NOT NULL,
			point_id INTEGER NOT NULL,
			order_index INTEGER NOT NULL,
			FOREIGN KEY (plan_id) REFERENCES inspection_plans(id),
			FOREIGN KEY (point_id) REFERENCES inspection_points(id)
		)`,
		`CREATE TABLE IF NOT EXISTS inspection_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plan_id INTEGER NOT NULL,
			code TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			inspector_id INTEGER,
			assigned_at DATETIME,
			started_at DATETIME,
			completed_at DATETIME,
			reviewer_id INTEGER,
			reviewed_at DATETIME,
			review_comment TEXT,
			due_date DATETIME NOT NULL,
			total_score INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (plan_id) REFERENCES inspection_plans(id),
			FOREIGN KEY (inspector_id) REFERENCES users(id),
			FOREIGN KEY (reviewer_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS task_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			point_id INTEGER NOT NULL,
			order_index INTEGER NOT NULL,
			checked INTEGER DEFAULT 0,
			checked_at DATETIME,
			checked_by INTEGER,
			FOREIGN KEY (task_id) REFERENCES inspection_tasks(id),
			FOREIGN KEY (point_id) REFERENCES inspection_points(id)
		)`,
		`CREATE TABLE IF NOT EXISTS inspection_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_point_id INTEGER NOT NULL,
			check_item_id INTEGER NOT NULL,
			score INTEGER NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_point_id) REFERENCES task_points(id),
			FOREIGN KEY (check_item_id) REFERENCES check_items(id)
		)`,
		`CREATE TABLE IF NOT EXISTS repair_orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			device_id INTEGER NOT NULL,
			repairer_id INTEGER,
			status TEXT NOT NULL DEFAULT 'pending',
			description TEXT,
			assigned_at DATETIME,
			completed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES inspection_tasks(id),
			FOREIGN KEY (device_id) REFERENCES devices(id),
			FOREIGN KEY (repairer_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS statistics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATE UNIQUE NOT NULL,
			total_tasks INTEGER DEFAULT 0,
			pending_tasks INTEGER DEFAULT 0,
			completed_tasks INTEGER DEFAULT 0,
			abnormal_tasks INTEGER DEFAULT 0,
			normal_tasks INTEGER DEFAULT 0,
			review_pass_tasks INTEGER DEFAULT 0,
			review_fail_tasks INTEGER DEFAULT 0,
			closed_tasks INTEGER DEFAULT 0,
			overdue_tasks INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON inspection_tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_inspector ON inspection_tasks(inspector_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_due ON inspection_tasks(due_date)`,
		`CREATE INDEX IF NOT EXISTS idx_task_points_task ON task_points(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_repair_device ON repair_orders(device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_plans_active ON inspection_plans(active)`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func initSeedData() error {
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	users := []struct {
		username string
		password string
		role     string
	}{
		{"admin", "admin123", "admin"},
		{"inspector1", "pass123", "inspector"},
		{"inspector2", "pass123", "inspector"},
		{"repair1", "pass123", "repair"},
		{"reviewer1", "pass123", "reviewer"},
	}

	for _, u := range users {
		_, err := DB.Exec(
			"INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			u.username, u.password, u.role,
		)
		if err != nil {
			return err
		}
	}

	devices := []struct {
		name        string
		code        string
		description string
	}{
		{"电机A", "DEV-001", "主电机"},
		{"传送带A", "DEV-002", "1号传送带"},
		{"空压机A", "DEV-003", "主空压机"},
	}

	for _, d := range devices {
		_, err := DB.Exec(
			"INSERT INTO devices (name, code, description) VALUES (?, ?, ?)",
			d.name, d.code, d.description,
		)
		if err != nil {
			return err
		}
	}

	checkItems := []struct {
		name        string
		description string
		maxScore    int
	}{
		{"温度", "设备运行温度检查", 100},
		{"外观", "设备外观完整性检查", 100},
		{"声音", "设备运行声音检查", 100},
		{"振动", "设备振动检查", 100},
	}

	for _, ci := range checkItems {
		_, err := DB.Exec(
			"INSERT INTO check_items (name, description, max_score) VALUES (?, ?, ?)",
			ci.name, ci.description, ci.maxScore,
		)
		if err != nil {
			return err
		}
	}

	points := []struct {
		deviceID   int
		name       string
		orderIndex int
	}{
		{1, "电机A-前端", 1},
		{1, "电机A-后端", 2},
		{2, "传送带A-入口", 3},
		{2, "传送带A-出口", 4},
		{3, "空压机A-主机", 5},
	}

	for _, p := range points {
		_, err := DB.Exec(
			"INSERT INTO inspection_points (device_id, name, order_index) VALUES (?, ?, ?)",
			p.deviceID, p.name, p.orderIndex,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
