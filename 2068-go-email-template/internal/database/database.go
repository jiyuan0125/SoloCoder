package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Template struct {
	ID        int64
	Name      string
	Subject   string
	HTML      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SendRecord struct {
	ID             string
	TemplateID     int64
	TemplateName   string
	Recipient      string
	Subject        string
	Variables      string
	Status         string
	Attempts       int
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	NextRetryAt    *time.Time
	SendResult     string
	BatchID        string
}

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		subject TEXT NOT NULL,
		html TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS send_records (
		id TEXT PRIMARY KEY,
		template_id INTEGER NOT NULL,
		template_name TEXT NOT NULL,
		recipient TEXT NOT NULL,
		subject TEXT NOT NULL,
		variables TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		last_error TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		next_retry_at DATETIME,
		send_result TEXT,
		batch_id TEXT,
		FOREIGN KEY (template_id) REFERENCES templates(id)
	);

	CREATE INDEX IF NOT EXISTS idx_send_records_status ON send_records(status);
	CREATE INDEX IF NOT EXISTS idx_send_records_batch_id ON send_records(batch_id);
	CREATE INDEX IF NOT EXISTS idx_send_records_next_retry_at ON send_records(next_retry_at);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) CreateTemplate(ctx context.Context, tpl *Template) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO templates (name, subject, html) VALUES (?, ?, ?)`,
		tpl.Name, tpl.Subject, tpl.HTML,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) GetTemplateByName(ctx context.Context, name string) (*Template, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, name, subject, html, created_at, updated_at FROM templates WHERE name = ?`,
		name,
	)
	tpl := &Template{}
	err := row.Scan(&tpl.ID, &tpl.Name, &tpl.Subject, &tpl.HTML, &tpl.CreatedAt, &tpl.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tpl, err
}

func (db *DB) UpdateTemplate(ctx context.Context, tpl *Template) error {
	_, err := db.ExecContext(ctx,
		`UPDATE templates SET subject = ?, html = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		tpl.Subject, tpl.HTML, tpl.ID,
	)
	return err
}

func (db *DB) DeleteTemplate(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM templates WHERE id = ?`, id)
	return err
}

func (db *DB) ListTemplates(ctx context.Context) ([]*Template, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, subject, html, created_at, updated_at FROM templates ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*Template
	for rows.Next() {
		tpl := &Template{}
		if err := rows.Scan(&tpl.ID, &tpl.Name, &tpl.Subject, &tpl.HTML, &tpl.CreatedAt, &tpl.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, tpl)
	}
	return templates, nil
}

func (db *DB) CreateSendRecord(ctx context.Context, record *SendRecord) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO send_records (id, template_id, template_name, recipient, subject, variables, status, attempts, last_error, created_at, updated_at, next_retry_at, send_result, batch_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.TemplateID, record.TemplateName, record.Recipient, record.Subject,
		record.Variables, record.Status, record.Attempts, record.LastError,
		record.CreatedAt, record.UpdatedAt, record.NextRetryAt, record.SendResult, record.BatchID,
	)
	return err
}

func (db *DB) GetSendRecord(ctx context.Context, id string) (*SendRecord, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, template_id, template_name, recipient, subject, variables, status, attempts, last_error, created_at, updated_at, next_retry_at, send_result, batch_id
		 FROM send_records WHERE id = ?`,
		id,
	)
	record := &SendRecord{}
	err := row.Scan(&record.ID, &record.TemplateID, &record.TemplateName, &record.Recipient, &record.Subject,
		&record.Variables, &record.Status, &record.Attempts, &record.LastError,
		&record.CreatedAt, &record.UpdatedAt, &record.NextRetryAt, &record.SendResult, &record.BatchID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return record, err
}

func (db *DB) UpdateSendRecord(ctx context.Context, record *SendRecord) error {
	_, err := db.ExecContext(ctx,
		`UPDATE send_records SET status = ?, attempts = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP, next_retry_at = ?, send_result = ?
		 WHERE id = ?`,
		record.Status, record.Attempts, record.LastError, record.NextRetryAt, record.SendResult, record.ID,
	)
	return err
}

func (db *DB) GetPendingRecordsForRetry(ctx context.Context, limit int) ([]*SendRecord, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, template_id, template_name, recipient, subject, variables, status, attempts, last_error, created_at, updated_at, next_retry_at, send_result, batch_id
		 FROM send_records 
		 WHERE status IN ('pending', 'retry') AND (next_retry_at IS NULL OR next_retry_at <= ?) AND attempts < 3
		 ORDER BY next_retry_at ASC
		 LIMIT ?`,
		time.Now(), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*SendRecord
	for rows.Next() {
		record := &SendRecord{}
		if err := rows.Scan(&record.ID, &record.TemplateID, &record.TemplateName, &record.Recipient, &record.Subject,
			&record.Variables, &record.Status, &record.Attempts, &record.LastError,
			&record.CreatedAt, &record.UpdatedAt, &record.NextRetryAt, &record.SendResult, &record.BatchID); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (db *DB) GetRecordsByBatch(ctx context.Context, batchID string) ([]*SendRecord, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, template_id, template_name, recipient, subject, variables, status, attempts, last_error, created_at, updated_at, next_retry_at, send_result, batch_id
		 FROM send_records 
		 WHERE batch_id = ?
		 ORDER BY created_at ASC`,
		batchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*SendRecord
	for rows.Next() {
		record := &SendRecord{}
		if err := rows.Scan(&record.ID, &record.TemplateID, &record.TemplateName, &record.Recipient, &record.Subject,
			&record.Variables, &record.Status, &record.Attempts, &record.LastError,
			&record.CreatedAt, &record.UpdatedAt, &record.NextRetryAt, &record.SendResult, &record.BatchID); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
