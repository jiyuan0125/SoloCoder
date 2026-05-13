package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const dbPath = "./tar_archiver.db"

type ArchiveOperation struct {
	ID          int64
	Operation   string
	ArchivePath string
	SourcePath  string
	StartTime   time.Time
	EndTime     *time.Time
	Status      string
	ErrorMsg    string
}

type CleanupLog struct {
	ID          int64
	OperationID int64
	CleanupTime time.Time
	Details     string
}

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err = createTables(); err != nil {
		return err
	}

	return nil
}

func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS archive_operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operation TEXT NOT NULL,
		archive_path TEXT NOT NULL,
		source_path TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		status TEXT NOT NULL,
		error_msg TEXT
	);

	CREATE TABLE IF NOT EXISTS cleanup_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operation_id INTEGER NOT NULL,
		cleanup_time DATETIME NOT NULL,
		details TEXT NOT NULL,
		FOREIGN KEY (operation_id) REFERENCES archive_operations(id)
	);
	`

	_, err := DB.ExecContext(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	return nil
}

func CreateArchiveOperation(op *ArchiveOperation) error {
	query := `
	INSERT INTO archive_operations (operation, archive_path, source_path, start_time, status)
	VALUES (?, ?, ?, ?, ?)
	`

	result, err := DB.ExecContext(
		context.Background(),
		query,
		op.Operation,
		op.ArchivePath,
		op.SourcePath,
		op.StartTime,
		op.Status,
	)
	if err != nil {
		return err
	}

	op.ID, err = result.LastInsertId()
	return err
}

func UpdateArchiveOperation(op *ArchiveOperation) error {
	query := `
	UPDATE archive_operations 
	SET end_time = ?, status = ?, error_msg = ?
	WHERE id = ?
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		op.EndTime,
		op.Status,
		op.ErrorMsg,
		op.ID,
	)
	return err
}

func CreateCleanupLog(log *CleanupLog) error {
	query := `
	INSERT INTO cleanup_logs (operation_id, cleanup_time, details)
	VALUES (?, ?, ?)
	`

	result, err := DB.ExecContext(
		context.Background(),
		query,
		log.OperationID,
		log.CleanupTime,
		log.Details,
	)
	if err != nil {
		return err
	}

	log.ID, err = result.LastInsertId()
	return err
}

func RemoveDB() error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(dbPath)
}
