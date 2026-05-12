package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	return createTables()
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			instructor TEXT NOT NULL,
			course_type TEXT NOT NULL,
			price INTEGER NOT NULL DEFAULT 0,
			subject TEXT,
			start_time DATETIME,
			duration INTEGER,
			max_online INTEGER DEFAULT 500,
			live_status TEXT,
			video_duration INTEGER,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nickname TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			grade TEXT NOT NULL,
			balance INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS recharge_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_number TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			amount INTEGER NOT NULL,
			is_used INTEGER NOT NULL DEFAULT 0,
			used_by INTEGER,
			used_at DATETIME,
			purchase_at DATETIME NOT NULL,
			expire_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS enrollments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			course_id INTEGER NOT NULL,
			price_paid INTEGER NOT NULL DEFAULT 0,
			enrolled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(student_id, course_id)
		)`,
		`CREATE TABLE IF NOT EXISTS watch_progress (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			course_id INTEGER NOT NULL,
			watched_seconds INTEGER NOT NULL DEFAULT 0,
			is_completed INTEGER NOT NULL DEFAULT 0,
			completed_at DATETIME,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(student_id, course_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			course_id INTEGER NOT NULL,
			rating INTEGER NOT NULL,
			comment TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(student_id, course_id)
		)`,
		`CREATE TABLE IF NOT EXISTS danmakus (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS votes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			end_time DATETIME,
			is_active INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS vote_options (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vote_id INTEGER NOT NULL,
			text TEXT NOT NULL,
			votes INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS vote_participants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vote_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			option_id INTEGER NOT NULL,
			voted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(vote_id, student_id)
		)`,
		`CREATE TABLE IF NOT EXISTS live_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL UNIQUE,
			peak_online INTEGER NOT NULL DEFAULT 0,
			average_online_time REAL NOT NULL DEFAULT 0,
			danmaku_count INTEGER NOT NULL DEFAULT 0,
			vote_participation REAL NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS live_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			join_time DATETIME NOT NULL,
			leave_time DATETIME,
			duration_seconds INTEGER DEFAULT 0
		)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %v, error: %v", query, err)
		}
	}

	return insertSeedData()
}

func insertSeedData() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM students").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO students (nickname, phone, grade, balance) VALUES 
		('小明', '13800138001', 'primary1', 10000),
		('小红', '13800138002', 'junior2', 5000),
		('小李', '13800138003', 'university', 20000)`)
	if err != nil {
		return err
	}

	now := time.Now()
	startTime1 := now.Add(1 * time.Hour)
	startTime2 := now.Add(24 * time.Hour)

	_, err = tx.Exec(`INSERT INTO courses (name, instructor, course_type, price, subject, start_time, duration, max_online, live_status, video_duration, description) VALUES 
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"高等数学精讲", "张教授", "live", 9900, "math", startTime1.Format(time.RFC3339), 90, 500, "not_started", nil, nil,
		"英语听力训练", "李老师", "live", 0, "english", startTime2.Format(time.RFC3339), 60, 300, "not_started", nil, nil,
		"语文作文技巧", "王老师", "record", 5900, "chinese", nil, nil, nil, nil, 3600, "全面提升语文写作能力，适合初中生",
		"物理实验基础", "刘教授", "record", 7900, "physics", nil, nil, nil, nil, 5400, "从基础到进阶的物理实验课程",
		"化学方程式", "赵老师", "record", 0, "chemistry", nil, nil, nil, nil, 1800, "免费学习化学方程式基础")
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO recharge_cards (card_number, password, amount, purchase_at, expire_at) VALUES 
		(?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?)`,
		"CARD001", "PASS123", 5000, now.Format(time.RFC3339), now.AddDate(0, 0, 90).Format(time.RFC3339),
		"CARD002", "PASS456", 10000, now.Format(time.RFC3339), now.AddDate(0, 0, 90).Format(time.RFC3339),
		"CARD003", "PASS789", 20000, now.Format(time.RFC3339), now.AddDate(0, 0, 90).Format(time.RFC3339))
	if err != nil {
		return err
	}

	return tx.Commit()
}
