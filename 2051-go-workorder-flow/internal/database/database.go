package database

import (
	"database/sql"
	"fmt"
	"time"

	"workorder-flow/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		leader_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		phone TEXT,
		team_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS duty_schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		duty_date TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(team_id, duty_date)
	);

	CREATE TABLE IF NOT EXISTS resource_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_type_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS work_orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		content TEXT,
		current_handler_id INTEGER,
		current_team_id INTEGER NOT NULL,
		process_status TEXT NOT NULL,
		work_order_status TEXT NOT NULL,
		submitter_id INTEGER NOT NULL,
		escalated INTEGER NOT NULL DEFAULT 0,
		rating INTEGER,
		escalated_at DATETIME,
		deadline_at DATETIME,
		process_start_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS work_order_resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		work_order_id INTEGER NOT NULL,
		resource_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS history_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		work_order_id INTEGER NOT NULL,
		operation_type TEXT NOT NULL,
		operator_id INTEGER NOT NULL,
		old_process_status TEXT,
		new_process_status TEXT,
		old_work_order_status TEXT,
		new_work_order_status TEXT,
		assigned_from INTEGER,
		assigned_to INTEGER,
		reason TEXT,
		comment TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS communication_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		work_order_id INTEGER NOT NULL,
		sender_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_work_orders_type ON work_orders(type);
	CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(work_order_status);
	CREATE INDEX IF NOT EXISTS idx_work_orders_process_status ON work_orders(process_status);
	CREATE INDEX IF NOT EXISTS idx_work_orders_submitter ON work_orders(submitter_id);
	CREATE INDEX IF NOT EXISTS idx_work_orders_handler ON work_orders(current_handler_id);
	CREATE INDEX IF NOT EXISTS idx_history_work_order ON history_records(work_order_id);
	CREATE INDEX IF NOT EXISTS idx_communication_work_order ON communication_records(work_order_id);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) SeedInitialData() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		return err
	}
	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	rows.Close()

	if count > 0 {
		return tx.Commit()
	}

	users := []struct {
		username string
		email    string
		phone    string
	}{
		{"admin", "admin@example.com", "13800000001"},
		{"tech1", "tech1@example.com", "13800000002"},
		{"tech2", "tech2@example.com", "13800000003"},
		{"service1", "service1@example.com", "13800000004"},
		{"service2", "service2@example.com", "13800000005"},
		{"complaint1", "complaint1@example.com", "13800000006"},
		{"complaint2", "complaint2@example.com", "13800000007"},
		{"submitter", "submitter@example.com", "13800000008"},
	}

	userIDs := make(map[string]int64)
	for _, u := range users {
		res, err := tx.Exec("INSERT INTO users (username, email, phone) VALUES (?, ?, ?)", u.username, u.email, u.phone)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		userIDs[u.username] = id
	}

	teams := []struct {
		name   string
		typ    models.WorkOrderType
		leader string
	}{
		{"故障报修团队", models.WorkOrderTypeFaultReport, "tech1"},
		{"服务请求团队", models.WorkOrderTypeServiceRequest, "service1"},
		{"投诉处理团队", models.WorkOrderTypeComplaint, "complaint1"},
	}

	teamIDs := make(map[string]int64)
	for _, t := range teams {
		res, err := tx.Exec("INSERT INTO teams (name, type, leader_id) VALUES (?, ?, ?)", t.name, string(t.typ), userIDs[t.leader])
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		teamIDs[t.name] = id
	}

	updateUserTeam := func(username, teamName string) error {
		if teamName == "" {
			_, err := tx.Exec("UPDATE users SET team_id = NULL WHERE id = ?", userIDs[username])
			return err
		}
		_, err := tx.Exec("UPDATE users SET team_id = ? WHERE id = ?", teamIDs[teamName], userIDs[username])
		return err
	}

	updateUserTeam("tech1", "故障报修团队")
	updateUserTeam("tech2", "故障报修团队")
	updateUserTeam("service1", "服务请求团队")
	updateUserTeam("service2", "服务请求团队")
	updateUserTeam("complaint1", "投诉处理团队")
	updateUserTeam("complaint2", "投诉处理团队")

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	schedules := []struct {
		teamName string
		username string
		date     string
	}{
		{"故障报修团队", "tech1", today},
		{"服务请求团队", "service1", today},
		{"投诉处理团队", "complaint1", today},
		{"故障报修团队", "tech2", tomorrow},
		{"服务请求团队", "service2", tomorrow},
		{"投诉处理团队", "complaint2", tomorrow},
	}

	for _, s := range schedules {
		_, err := tx.Exec("INSERT INTO duty_schedules (team_id, user_id, duty_date) VALUES (?, ?, ?)",
			teamIDs[s.teamName], userIDs[s.username], s.date)
		if err != nil {
			return err
		}
	}

	resourceTypes := []struct {
		name string
		desc string
	}{
		{"服务器", "服务器资源"},
		{"网络设备", "网络设备资源"},
		{"软件系统", "软件系统资源"},
	}

	rtIDs := make(map[string]int64)
	for _, rt := range resourceTypes {
		res, err := tx.Exec("INSERT INTO resource_types (name, description) VALUES (?, ?)", rt.name, rt.desc)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		rtIDs[rt.name] = id
	}

	resources := []struct {
		rtName string
		name   string
		desc   string
	}{
		{"服务器", "应用服务器1", "主应用服务器"},
		{"服务器", "数据库服务器1", "主数据库服务器"},
		{"网络设备", "核心交换机", "机房核心交换机"},
		{"软件系统", "OA系统", "办公自动化系统"},
		{"软件系统", "CRM系统", "客户关系管理系统"},
	}

	for _, r := range resources {
		_, err := tx.Exec("INSERT INTO resources (resource_type_id, name, description) VALUES (?, ?, ?)",
			rtIDs[r.rtName], r.name, r.desc)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
