package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dataDir, "cicd.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pipelines (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		total_amount REAL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS pipeline_triggers (
		pipeline_id TEXT NOT NULL,
		trigger_type TEXT NOT NULL,
		PRIMARY KEY (pipeline_id, trigger_type),
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS phases (
		id TEXT PRIMARY KEY,
		pipeline_id TEXT NOT NULL,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		task_mode TEXT NOT NULL,
		planned_amount REAL DEFAULT 0,
		order_index INTEGER NOT NULL,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		phase_id TEXT NOT NULL,
		name TEXT NOT NULL,
		script TEXT NOT NULL,
		timeout_sec INTEGER NOT NULL DEFAULT 60,
		failure_strategy TEXT NOT NULL DEFAULT 'stop',
		order_index INTEGER NOT NULL,
		FOREIGN KEY (phase_id) REFERENCES phases(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS executions (
		id TEXT PRIMARY KEY,
		pipeline_id TEXT NOT NULL,
		trigger_type TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		target_env TEXT,
		target_phase TEXT,
		created_at DATETIME NOT NULL,
		is_rollback INTEGER NOT NULL DEFAULT 0,
		rollback_from TEXT,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
	);

	CREATE TABLE IF NOT EXISTS phase_results (
		id TEXT PRIMARY KEY,
		execution_id TEXT NOT NULL,
		phase_type TEXT NOT NULL,
		phase_name TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		order_index INTEGER NOT NULL,
		FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS task_results (
		id TEXT PRIMARY KEY,
		phase_result_id TEXT NOT NULL,
		task_id TEXT NOT NULL,
		task_name TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		log_path TEXT,
		order_index INTEGER NOT NULL,
		FOREIGN KEY (phase_result_id) REFERENCES phase_results(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS artifacts (
		id TEXT PRIMARY KEY,
		execution_id TEXT NOT NULL,
		task_result_id TEXT,
		name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		size INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS cleanup_records (
		id TEXT PRIMARY KEY,
		execution_id TEXT NOT NULL,
		action TEXT NOT NULL,
		status TEXT NOT NULL,
		result TEXT,
		created_at DATETIME NOT NULL,
		finished_at DATETIME,
		FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_executions_pipeline ON executions(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);
	CREATE INDEX IF NOT EXISTS idx_phase_results_execution ON phase_results(execution_id);
	CREATE INDEX IF NOT EXISTS idx_task_results_phase ON task_results(phase_result_id);
	CREATE INDEX IF NOT EXISTS idx_artifacts_execution ON artifacts(execution_id);
	CREATE INDEX IF NOT EXISTS idx_cleanup_execution ON cleanup_records(execution_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func Now() time.Time {
	return time.Now().UTC()
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func TimePtrNil() *time.Time {
	return nil
}

func ScanTime(rows *sql.Rows, dest ...interface{}) error {
	return rows.Scan(dest...)
}

func NullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func NullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}

func NullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

func ScanNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func ScanNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		t := nt.Time
		return &t
	}
	return nil
}

func FormatError(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
