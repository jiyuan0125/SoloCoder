package repositories

import (
	"database/sql"
	"image-batch/config"
	"image-batch/models"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository() (*TaskRepository, error) {
	err := os.MkdirAll(filepath.Dir(config.DatabasePath), 0755)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		return nil, err
	}

	repo := &TaskRepository{db: db}
	err = repo.initSchema()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *TaskRepository) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS process_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT NOT NULL,
		total INTEGER NOT NULL,
		success INTEGER NOT NULL,
		failed INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		finished_at DATETIME,
		zip_path TEXT,
		report TEXT
	);`

	_, err := r.db.Exec(query)
	return err
}

func (r *TaskRepository) CreateTask(total int) (*models.ProcessTask, error) {
	task := &models.ProcessTask{
		Status:    "processing",
		Total:     total,
		Success:   0,
		Failed:    0,
		CreatedAt: time.Now(),
	}

	result, err := r.db.Exec(`
		INSERT INTO process_tasks (status, total, success, failed, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		task.Status, task.Total, task.Success, task.Failed, task.CreatedAt)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	task.ID = id
	return task, nil
}

func (r *TaskRepository) UpdateTask(task *models.ProcessTask) error {
	_, err := r.db.Exec(`
		UPDATE process_tasks 
		SET status = ?, success = ?, failed = ?, finished_at = ?, zip_path = ?, report = ?
		WHERE id = ?`,
		task.Status, task.Success, task.Failed, task.FinishedAt, task.ZipPath, task.Report, task.ID)
	return err
}

func (r *TaskRepository) GetTaskByID(id int64) (*models.ProcessTask, error) {
	task := &models.ProcessTask{}
	var finishedAt sql.NullTime
	var zipPath, report sql.NullString

	err := r.db.QueryRow(`
		SELECT id, status, total, success, failed, created_at, finished_at, zip_path, report
		FROM process_tasks WHERE id = ?`, id).Scan(
		&task.ID, &task.Status, &task.Total, &task.Success, &task.Failed,
		&task.CreatedAt, &finishedAt, &zipPath, &report)

	if err != nil {
		return nil, err
	}

	if finishedAt.Valid {
		task.FinishedAt = finishedAt.Time
	}
	if zipPath.Valid {
		task.ZipPath = zipPath.String
	}
	if report.Valid {
		task.Report = report.String
	}

	return task, nil
}

func (r *TaskRepository) Close() error {
	return r.db.Close()
}
