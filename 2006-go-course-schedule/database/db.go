package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Store interface {
	GetDB() *sql.DB
}

func InitDB(dataSource string) error {
	dir := filepath.Dir(dataSource)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	db, err := sql.Open("sqlite3", dataSource+"?_foreign_keys=on")
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return err
	}

	DB = db

	if err := createTables(); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			instructor TEXT NOT NULL,
			classroom TEXT NOT NULL,
			capacity INTEGER NOT NULL DEFAULT 0,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			series_id TEXT,
			start_time DATETIME NOT NULL,
			duration_hours INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS enrollments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			student_name TEXT NOT NULL,
			enrolled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
			UNIQUE(course_id, student_name)
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			student_name TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_course ON schedules(course_id)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_series ON schedules(series_id)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_start ON schedules(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_enrollments_course ON enrollments(course_id)`,
		`CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_name)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_student ON notifications(student_name)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_course ON notifications(course_id)`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func Now() time.Time {
	return time.Now().UTC()
}

func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
