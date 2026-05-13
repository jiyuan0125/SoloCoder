package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Resource struct {
	ID          int64
	Name        string
	Type        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ConversionStatus string

const (
	StatusPending   ConversionStatus = "pending"
	StatusConverting ConversionStatus = "converting"
	StatusCompleted  ConversionStatus = "completed"
	StatusFailed     ConversionStatus = "failed"
)

type ConversionJob struct {
	ID              int64
	ResourceID      sql.NullInt64
	SourceFormat    string
	TargetFormat    string
	Status          ConversionStatus
	OriginalFile    string
	ConvertedFile   string
	ProcessedCount  int
	SkippedCount    int
	TotalCount      int
	ErrorMessage    string
	CurrentStep     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ResourceRelation struct {
	ID          int64
	SourceID    int64
	TargetID    int64
	RelationType string
	CreatedAt   time.Time
}

var DB *sql.DB

func Init(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	if err = DB.Ping(); err != nil {
		return err
	}
	return createTables()
}

func createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS conversion_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			resource_id INTEGER,
			source_format TEXT NOT NULL,
			target_format TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			original_file TEXT,
			converted_file TEXT,
			processed_count INTEGER DEFAULT 0,
			skipped_count INTEGER DEFAULT 0,
			total_count INTEGER DEFAULT 0,
			error_message TEXT,
			current_step TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (resource_id) REFERENCES resources(id)
		)`,
		`CREATE TABLE IF NOT EXISTS resource_relations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER NOT NULL,
			target_id INTEGER NOT NULL,
			relation_type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (source_id) REFERENCES resources(id),
			FOREIGN KEY (target_id) REFERENCES resources(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_resource ON conversion_jobs(resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status ON conversion_jobs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_relations_source ON resource_relations(source_id)`,
		`CREATE INDEX IF NOT EXISTS idx_relations_target ON resource_relations(target_id)`,
	}

	for _, stmt := range statements {
		_, err := DB.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateResource(r *Resource) error {
	result, err := DB.Exec(
		`INSERT INTO resources (name, type, description) VALUES (?, ?, ?)`,
		r.Name, r.Type, r.Description,
	)
	if err != nil {
		return err
	}
	r.ID, err = result.LastInsertId()
	return err
}

func GetResource(id int64) (*Resource, error) {
	r := &Resource{}
	err := DB.QueryRow(
		`SELECT id, name, type, description, created_at, updated_at FROM resources WHERE id = ?`,
		id,
	).Scan(&r.ID, &r.Name, &r.Type, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func ListResources() ([]*Resource, error) {
	rows, err := DB.Query(`SELECT id, name, type, description, created_at, updated_at FROM resources ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*Resource
	for rows.Next() {
		r := &Resource{}
		err := rows.Scan(&r.ID, &r.Name, &r.Type, &r.Description, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func CreateJob(j *ConversionJob) error {
	result, err := DB.Exec(
		`INSERT INTO conversion_jobs (resource_id, source_format, target_format, status, original_file) 
		 VALUES (?, ?, ?, ?, ?)`,
		j.ResourceID, j.SourceFormat, j.TargetFormat, StatusPending, j.OriginalFile,
	)
	if err != nil {
		return err
	}
	j.ID, err = result.LastInsertId()
	return err
}

func GetJob(id int64) (*ConversionJob, error) {
	j := &ConversionJob{}
	err := DB.QueryRow(`
		SELECT id, resource_id, source_format, target_format, status, 
		       original_file, converted_file, processed_count, skipped_count, 
		       total_count, error_message, current_step, created_at, updated_at 
		FROM conversion_jobs WHERE id = ?`,
		id,
	).Scan(&j.ID, &j.ResourceID, &j.SourceFormat, &j.TargetFormat, &j.Status,
		&j.OriginalFile, &j.ConvertedFile, &j.ProcessedCount, &j.SkippedCount,
		&j.TotalCount, &j.ErrorMessage, &j.CurrentStep, &j.CreatedAt, &j.UpdatedAt)
	return j, err
}

func UpdateJob(j *ConversionJob) error {
	_, err := DB.Exec(`
		UPDATE conversion_jobs SET 
			status = ?, converted_file = ?, processed_count = ?, 
			skipped_count = ?, total_count = ?, error_message = ?, 
			current_step = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		j.Status, j.ConvertedFile, j.ProcessedCount, j.SkippedCount,
		j.TotalCount, j.ErrorMessage, j.CurrentStep, j.ID,
	)
	return err
}

func UpdateJobStatus(id int64, status ConversionStatus, currentStep string) error {
	_, err := DB.Exec(`
		UPDATE conversion_jobs SET status = ?, current_step = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		status, currentStep, id,
	)
	return err
}

func ListJobsByResource(resourceID int64) ([]*ConversionJob, error) {
	rows, err := DB.Query(`
		SELECT id, resource_id, source_format, target_format, status, 
		       original_file, converted_file, processed_count, skipped_count, 
		       total_count, error_message, current_step, created_at, updated_at 
		FROM conversion_jobs WHERE resource_id = ? ORDER BY created_at DESC`,
		resourceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*ConversionJob
	for rows.Next() {
		j := &ConversionJob{}
		err := rows.Scan(&j.ID, &j.ResourceID, &j.SourceFormat, &j.TargetFormat, &j.Status,
			&j.OriginalFile, &j.ConvertedFile, &j.ProcessedCount, &j.SkippedCount,
			&j.TotalCount, &j.ErrorMessage, &j.CurrentStep, &j.CreatedAt, &j.UpdatedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func CreateRelation(r *ResourceRelation) error {
	result, err := DB.Exec(
		`INSERT INTO resource_relations (source_id, target_id, relation_type) VALUES (?, ?, ?)`,
		r.SourceID, r.TargetID, r.RelationType,
	)
	if err != nil {
		return err
	}
	r.ID, err = result.LastInsertId()
	return err
}

func GetResourceRelations(resourceID int64) ([]*ResourceRelation, error) {
	rows, err := DB.Query(`
		SELECT id, source_id, target_id, relation_type, created_at 
		FROM resource_relations WHERE source_id = ? OR target_id = ? ORDER BY created_at DESC`,
		resourceID, resourceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []*ResourceRelation
	for rows.Next() {
		r := &ResourceRelation{}
		err := rows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		relations = append(relations, r)
	}
	return relations, nil
}
