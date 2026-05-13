package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Status string

const (
	StatusDeveloping   Status = "developing"
	StatusTesting      Status = "testing"
	StatusPending      Status = "pending_release"
	StatusCanary       Status = "canary"
	StatusFullRelease  Status = "full_release"
	StatusCompleted    Status = "completed"
)

var validTransitions = map[Status][]Status{
	StatusDeveloping:  {StatusTesting},
	StatusTesting:     {StatusDeveloping, StatusPending},
	StatusPending:     {StatusCanary},
	StatusCanary:      {StatusFullRelease},
	StatusFullRelease: {StatusCompleted},
}

type Release struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Status       Status    `json:"status"`
	CodeFrozen   bool      `json:"code_frozen"`
	CanaryRatio  *int      `json:"canary_ratio,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StatusLog struct {
	ID          int64     `json:"id"`
	ReleaseID   int64     `json:"release_id"`
	FromStatus  Status    `json:"from_status"`
	ToStatus    Status    `json:"to_status"`
	Operator    string    `json:"operator"`
	Reason      string    `json:"reason"`
	CanaryRatio *int      `json:"canary_ratio,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type RollbackRecord struct {
	ID            int64     `json:"id"`
	ReleaseID     int64     `json:"release_id"`
	FromVersion   string    `json:"from_version"`
	ToVersion     string    `json:"to_version"`
	Operator      string    `json:"operator"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

var db *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS releases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		version TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL,
		code_frozen INTEGER NOT NULL DEFAULT 0,
		canary_ratio INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS status_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		from_status TEXT NOT NULL,
		to_status TEXT NOT NULL,
		operator TEXT NOT NULL,
		reason TEXT NOT NULL,
		canary_ratio INTEGER,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (release_id) REFERENCES releases(id)
	);

	CREATE TABLE IF NOT EXISTS rollback_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		from_version TEXT NOT NULL,
		to_version TEXT NOT NULL,
		operator TEXT NOT NULL,
		reason TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (release_id) REFERENCES releases(id)
	);

	CREATE INDEX IF NOT EXISTS idx_status_logs_release ON status_logs(release_id);
	CREATE INDEX IF NOT EXISTS idx_rollback_records_release ON rollback_records(release_id);
	`
	_, err := db.Exec(query)
	return err
}

func GetDB() *sql.DB {
	return db
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func IsValidCanaryRatio(ratio int) bool {
	return ratio == 1 || ratio == 5 || ratio == 10 || ratio == 50
}

func CanTransition(from, to Status) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (r *Release) ValidateStatusTransition(to Status) error {
	if r.Status == StatusCompleted {
		return errors.New("已完成的发布单不能回退")
	}

	if !CanTransition(r.Status, to) {
		return fmt.Errorf("状态流转无效: %s -> %s", r.Status, to)
	}

	return nil
}

func (r *Release) CanAddCommit() bool {
	if r.Status == StatusDeveloping && !r.CodeFrozen {
		return true
	}
	return false
}
