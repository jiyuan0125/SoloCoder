package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go-batch-scheduler/models"

	_ "modernc.org/sqlite"
)

type Database struct {
	*sql.DB
	mu sync.Mutex
}

func New(dbPath string) (*Database, error) {
	os.MkdirAll("data", 0755)

	d, err := sql.Open("sqlite", dbPath+"?_journal=WAL&_fk=1")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	d.SetMaxOpenConns(1)

	db := &Database{DB: d}
	if err := db.init(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) init() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS resources (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS resource_relations (
			from_id TEXT NOT NULL,
			to_id TEXT NOT NULL,
			relation TEXT NOT NULL,
			PRIMARY KEY (from_id, to_id),
			FOREIGN KEY (from_id) REFERENCES resources(id),
			FOREIGN KEY (to_id) REFERENCES resources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			priority INTEGER NOT NULL,
			timeout INTEGER NOT NULL,
			max_retries INTEGER NOT NULL DEFAULT 0,
			retry_count INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			flow_status TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			started_at DATETIME,
			completed_at DATETIME,
			failed_reason TEXT,
			is_final_failure INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (resource_id) REFERENCES resources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS dependencies (
			task_id TEXT NOT NULL,
			depends_on_id TEXT NOT NULL,
			PRIMARY KEY (task_id, depends_on_id),
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (depends_on_id) REFERENCES tasks(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_resource ON tasks(resource_id)`,
	}

	for _, q := range queries {
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	return tx.Commit()
}

func (db *Database) CreateResource(r *models.Resource) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.Exec(`INSERT OR REPLACE INTO resources (id, name, type) VALUES (?, ?, ?)`, r.ID, r.Name, r.Type)
	return err
}

func (db *Database) GetResource(id string) (*models.Resource, error) {
	r := &models.Resource{}
	err := db.QueryRow(`SELECT id, name, type FROM resources WHERE id = ?`, id).Scan(&r.ID, &r.Name, &r.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (db *Database) ListResources() ([]*models.Resource, error) {
	rows, err := db.Query(`SELECT id, name, type FROM resources ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs []*models.Resource
	for rows.Next() {
		r := &models.Resource{}
		if err := rows.Scan(&r.ID, &r.Name, &r.Type); err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, rows.Err()
}

func (db *Database) CreateResourceRelation(fromID, toID, relation string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.Exec(`INSERT OR REPLACE INTO resource_relations (from_id, to_id, relation) VALUES (?, ?, ?)`, fromID, toID, relation)
	return err
}

func (db *Database) GetResourceRelations(resourceID string) ([]*models.ResourceRelation, error) {
	rows, err := db.Query(`SELECT from_id, to_id, relation FROM resource_relations WHERE from_id = ? OR to_id = ?`, resourceID, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs []*models.ResourceRelation
	for rows.Next() {
		r := &models.ResourceRelation{}
		if err := rows.Scan(&r.FromID, &r.ToID, &r.Relation); err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, rows.Err()
}

func (db *Database) CreateTask(ctx context.Context, task *models.Task) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.ExecContext(ctx, `INSERT INTO tasks 
		(id, name, type, priority, timeout, max_retries, retry_count, status, flow_status, resource_id, created_at, is_final_failure) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		task.ID, task.Name, task.Type, int(task.Priority),
		int64(task.Timeout/time.Second), task.MaxRetries, task.RetryCount,
		string(task.Status), string(task.FlowStatus), task.ResourceID, task.CreatedAt)
	return err
}

func (db *Database) GetTask(id string) (*models.Task, error) {
	t := &models.Task{}
	var startedAt, completedAt sql.NullTime
	var failedReason sql.NullString
	var priority int

	err := db.QueryRow(`SELECT id, name, type, priority, timeout, max_retries, retry_count, status, flow_status, 
		resource_id, created_at, started_at, completed_at, failed_reason, is_final_failure 
		FROM tasks WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Type, &priority, &t.Timeout,
		&t.MaxRetries, &t.RetryCount, &t.Status, &t.FlowStatus,
		&t.ResourceID, &t.CreatedAt, &startedAt, &completedAt,
		&failedReason, &t.IsFinalFailure,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	t.Priority = models.Priority(priority)
	if startedAt.Valid {
		t.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if failedReason.Valid {
		t.FailedReason = failedReason.String
	}
	return t, nil
}

func (db *Database) UpdateTask(ctx context.Context, task *models.Task) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.ExecContext(ctx, `UPDATE tasks SET 
		name=?, type=?, priority=?, timeout=?, max_retries=?, retry_count=?, status=?, flow_status=?, 
		resource_id=?, started_at=?, completed_at=?, failed_reason=?, is_final_failure=? WHERE id=?`,
		task.Name, task.Type, int(task.Priority),
		int64(task.Timeout/time.Second), task.MaxRetries, task.RetryCount,
		string(task.Status), string(task.FlowStatus), task.ResourceID,
		task.StartedAt, task.CompletedAt, task.FailedReason, task.IsFinalFailure, task.ID)
	return err
}

func (db *Database) AddDependency(taskID, dependsOnID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.Exec(`INSERT OR IGNORE INTO dependencies (task_id, depends_on_id) VALUES (?, ?)`, taskID, dependsOnID)
	return err
}

func (db *Database) GetDependencies(taskID string) ([]string, error) {
	rows, err := db.Query(`SELECT depends_on_id FROM dependencies WHERE task_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *Database) GetDependents(taskID string) ([]string, error) {
	rows, err := db.Query(`SELECT task_id FROM dependencies WHERE depends_on_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *Database) AreAllDependenciesSuccessful(taskID string) (bool, error) {
	deps, err := db.GetDependencies(taskID)
	if err != nil {
		return false, err
	}
	if len(deps) == 0 {
		return true, nil
	}

	for _, depID := range deps {
		t, err := db.GetTask(depID)
		if err != nil {
			return false, err
		}
		if t == nil || t.Status != models.StatusSuccess {
			return false, nil
		}
	}
	return true, nil
}

func (db *Database) HasUnfinishedDependents(taskID string) (bool, error) {
	deps, err := db.GetDependents(taskID)
	if err != nil {
		return false, err
	}
	for _, depID := range deps {
		t, err := db.GetTask(depID)
		if err != nil {
			return false, err
		}
		if t != nil && t.Status != models.StatusSuccess && t.Status != models.StatusFailed {
			return true, nil
		}
	}
	return false, nil
}

func (db *Database) GetPendingTasks(ctx context.Context) ([]*models.Task, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, type, priority, timeout, max_retries, retry_count, status, flow_status, 
		resource_id, created_at, started_at, completed_at, failed_reason, is_final_failure 
		FROM tasks WHERE status = ? AND flow_status = ?
		ORDER BY priority DESC, created_at ASC`, models.StatusQueued, models.FlowApproved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		t := &models.Task{}
		var startedAt, completedAt sql.NullTime
		var failedReason sql.NullString
		var priority int

		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &priority, &t.Timeout,
			&t.MaxRetries, &t.RetryCount, &t.Status, &t.FlowStatus,
			&t.ResourceID, &t.CreatedAt, &startedAt, &completedAt,
			&failedReason, &t.IsFinalFailure); err != nil {
			return nil, err
		}

		t.Priority = models.Priority(priority)
		if startedAt.Valid {
			t.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if failedReason.Valid {
			t.FailedReason = failedReason.String
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (db *Database) GetStats() (*models.Stats, error) {
	s := &models.Stats{}

	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&s.TotalTasks); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = ?`, models.StatusQueued).Scan(&s.QueuedCount); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = ?`, models.StatusRunning).Scan(&s.RunningCount); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = ?`, models.StatusSuccess).Scan(&s.SuccessCount); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status = ?`, models.StatusFailed).Scan(&s.FailedCount); err != nil {
		return nil, err
	}

	var totalMs sql.NullInt64
	var count int64
	if err := db.QueryRow(`SELECT SUM(strftime('%s', completed_at) - strftime('%s', started_at)) * 1000, COUNT(*) 
		FROM tasks WHERE started_at IS NOT NULL AND completed_at IS NOT NULL`).Scan(&totalMs, &count); err != nil {
		return nil, err
	}
	if count > 0 && totalMs.Valid {
		s.AvgExecTime = time.Duration(totalMs.Int64/count) * time.Millisecond
	}
	return s, nil
}

func (db *Database) GetResourceSummaries() ([]*models.ResourceSummary, error) {
	resources, err := db.ListResources()
	if err != nil {
		return nil, err
	}

	var summaries []*models.ResourceSummary
	for _, r := range resources {
		summary := &models.ResourceSummary{
			ResourceID:   r.ID,
			ResourceName: r.Name,
			ResourceType: r.Type,
		}

		if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE resource_id = ?`, r.ID).Scan(&summary.TotalTasks); err != nil {
			log.Printf("error getting total for %s: %v", r.ID, err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE resource_id = ? AND status = ?`, r.ID, models.StatusQueued).Scan(&summary.QueuedCount); err != nil {
			log.Printf("error getting queued for %s: %v", r.ID, err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE resource_id = ? AND status = ?`, r.ID, models.StatusRunning).Scan(&summary.RunningCount); err != nil {
			log.Printf("error getting running for %s: %v", r.ID, err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE resource_id = ? AND status = ?`, r.ID, models.StatusSuccess).Scan(&summary.SuccessCount); err != nil {
			log.Printf("error getting success for %s: %v", r.ID, err)
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE resource_id = ? AND status = ?`, r.ID, models.StatusFailed).Scan(&summary.FailedCount); err != nil {
			log.Printf("error getting failed for %s: %v", r.ID, err)
		}

		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (db *Database) CheckCyclicDependency(taskID string, dependencies []string) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(id string) bool
	dfs = func(id string) bool {
		visited[id] = true
		recStack[id] = true

		deps, err := db.GetDependencies(id)
		if err != nil {
			return false
		}

		for _, dep := range deps {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		if id == taskID {
			for _, newDep := range dependencies {
				if !visited[newDep] {
					if dfs(newDep) {
						return true
					}
				} else if recStack[newDep] {
					return true
				}
			}
		}

		recStack[id] = false
		return false
	}

	// Add temporary dependencies for check
	for _, dep := range dependencies {
		if dep == taskID {
			return true
		}
	}

	return dfs(taskID)
}

func (db *Database) GetAllTaskIDs() ([]string, error) {
	rows, err := db.Query(`SELECT id FROM tasks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
