package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

const (
	TaskStatusPending      = "pending"
	TaskStatusInProgress   = "in_progress"
	TaskStatusNormal       = "normal"
	TaskStatusAbnormal     = "abnormal"
	TaskStatusUnderReview  = "under_review"
	TaskStatusReviewPass   = "review_pass"
	TaskStatusReviewFail   = "review_fail"
	TaskStatusClosed       = "closed"
	TaskStatusInMaintenance = "in_maintenance"

	MaintenanceStatusPending   = "pending"
	MaintenanceStatusInProgress = "in_progress"
	MaintenanceStatusCompleted  = "completed"

	ScheduleFrequencyDaily   = "daily"
	ScheduleFrequencyWeekly  = "weekly"
	ScheduleFrequencyMonthly = "monthly"
)

var TaskStatusOrder = []string{
	TaskStatusPending,
	TaskStatusInProgress,
	TaskStatusNormal,
	TaskStatusAbnormal,
	TaskStatusUnderReview,
	TaskStatusReviewPass,
	TaskStatusReviewFail,
	TaskStatusClosed,
	TaskStatusInMaintenance,
}

var ValidStatusTransitions = map[string][]string{
	TaskStatusPending:      {TaskStatusInProgress},
	TaskStatusInProgress:   {TaskStatusNormal, TaskStatusAbnormal},
	TaskStatusNormal:       {TaskStatusUnderReview},
	TaskStatusAbnormal:     {TaskStatusUnderReview, TaskStatusInMaintenance},
	TaskStatusUnderReview:  {TaskStatusReviewPass, TaskStatusReviewFail},
	TaskStatusReviewPass:   {TaskStatusClosed},
	TaskStatusReviewFail:   {TaskStatusInProgress},
	TaskStatusInMaintenance: {TaskStatusInProgress},
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.InitSchema(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT UNIQUE NOT NULL,
		location TEXT,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS inspection_points (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS route_points (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_id INTEGER NOT NULL,
		point_id INTEGER NOT NULL,
		order_num INTEGER NOT NULL,
		device_id INTEGER,
		FOREIGN KEY (route_id) REFERENCES routes(id) ON DELETE CASCADE,
		FOREIGN KEY (point_id) REFERENCES inspection_points(id),
		FOREIGN KEY (device_id) REFERENCES devices(id)
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL CHECK (role IN ('inspector', 'maintainer', 'reviewer')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS inspection_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		standard_value TEXT,
		max_score INTEGER DEFAULT 100,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		route_id INTEGER NOT NULL,
		frequency TEXT NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly')),
		start_date DATE NOT NULL,
		next_run_date DATE,
		active BOOLEAN DEFAULT 1,
		quota_total INTEGER DEFAULT 100,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (route_id) REFERENCES routes(id)
	);

	CREATE TABLE IF NOT EXISTS schedule_item_quotas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		schedule_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		quota INTEGER NOT NULL,
		FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES inspection_items(id)
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		schedule_id INTEGER,
		route_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		inspector_id INTEGER,
		reviewer_id INTEGER,
		total_score INTEGER DEFAULT 0,
		is_abnormal BOOLEAN DEFAULT 0,
		due_date DATETIME,
		completed_at DATETIME,
		reminder_sent BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (schedule_id) REFERENCES schedules(id),
		FOREIGN KEY (route_id) REFERENCES routes(id),
		FOREIGN KEY (inspector_id) REFERENCES users(id),
		FOREIGN KEY (reviewer_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS task_checkins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		route_point_id INTEGER NOT NULL,
		checked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (route_point_id) REFERENCES route_points(id)
	);

	CREATE TABLE IF NOT EXISTS task_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		score INTEGER,
		abnormal_desc TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES inspection_items(id)
	);

	CREATE TABLE IF NOT EXISTS maintenance_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		device_id INTEGER NOT NULL,
		maintainer_id INTEGER,
		status TEXT NOT NULL DEFAULT 'pending',
		issue_desc TEXT,
		solution TEXT,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id),
		FOREIGN KEY (device_id) REFERENCES devices(id),
		FOREIGN KEY (maintainer_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS review_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		reviewer_id INTEGER NOT NULL,
		result TEXT NOT NULL CHECK (result IN ('pass', 'fail')),
		reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id),
		FOREIGN KEY (reviewer_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS statistics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date DATE UNIQUE,
		total_tasks INTEGER DEFAULT 0,
		completed_tasks INTEGER DEFAULT 0,
		abnormal_tasks INTEGER DEFAULT 0,
		maintenance_tasks INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_inspector ON tasks(inspector_id);
	CREATE INDEX IF NOT EXISTS idx_route_points_route ON route_points(route_id);
	CREATE INDEX IF NOT EXISTS idx_schedules_active ON schedules(active);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) getStatusIndex(status string) int {
	for i, s := range TaskStatusOrder {
		if s == status {
			return i
		}
	}
	return -1
}

func (d *Database) canTransition(from, to string) bool {
	valid, exists := ValidStatusTransitions[from]
	if !exists {
		return false
	}
	for _, v := range valid {
		if v == to {
			return true
		}
	}
	return false
}

func (d *Database) UpdateStatistics() error {
	_, err := d.db.Exec(`
	INSERT INTO statistics (date, total_tasks, completed_tasks, abnormal_tasks, maintenance_tasks)
	SELECT DATE('now'), 
		COUNT(*) as total_tasks,
		SUM(CASE WHEN status IN ('closed', 'review_pass') THEN 1 ELSE 0 END) as completed_tasks,
		SUM(CASE WHEN status IN ('abnormal', 'in_maintenance') THEN 1 ELSE 0 END) as abnormal_tasks,
		SUM(CASE WHEN status = 'in_maintenance' THEN 1 ELSE 0 END) as maintenance_tasks
	FROM tasks
	ON CONFLICT(date) DO UPDATE SET
		total_tasks = excluded.total_tasks,
		completed_tasks = excluded.completed_tasks,
		abnormal_tasks = excluded.abnormal_tasks,
		maintenance_tasks = excluded.maintenance_tasks
	`)
	if err != nil {
		log.Printf("Failed to update statistics: %v", err)
		return err
	}
	return nil
}

func (d *Database) RecalculateQuotas(scheduleID int64) error {
	var totalQuota int
	err := d.db.QueryRow("SELECT quota_total FROM schedules WHERE id = ?", scheduleID).Scan(&totalQuota)
	if err != nil {
		return err
	}

	var itemCount int
	err = d.db.QueryRow(`
		SELECT COUNT(*) FROM schedule_item_quotas WHERE schedule_id = ?
	`, scheduleID).Scan(&itemCount)
	if err != nil || itemCount == 0 {
		return err
	}

	baseQuota := totalQuota / itemCount
	remainder := totalQuota % itemCount

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT id FROM schedule_item_quotas WHERE schedule_id = ? ORDER BY id
	`, scheduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	for i, id := range ids {
		quota := baseQuota
		if i < remainder {
			quota++
		}
		_, err := tx.Exec(`
			UPDATE schedule_item_quotas SET quota = ? WHERE id = ?
		`, quota, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func logOperation(message string) {
	fmt.Printf("[LOG] %s\n", message)
}
