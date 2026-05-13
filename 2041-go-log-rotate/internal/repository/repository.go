package repository

import (
	"database/sql"
	"fmt"
	"log"
	"log-rotate/internal/config"
	"log-rotate/internal/model"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_fk=1")
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	repo := &Repository{db: db}
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("init schema failed: %w", err)
	}

	if err := repo.initDefaultConfig(); err != nil {
		return nil, fmt.Errorf("init default config failed: %w", err)
	}

	return repo, nil
}

func (r *Repository) initSchema() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS log_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			max_file_size_mb INTEGER NOT NULL DEFAULT 100,
			rotate_hour INTEGER NOT NULL DEFAULT 3,
			rotate_minute INTEGER NOT NULL DEFAULT 0,
			max_archive_files INTEGER NOT NULL DEFAULT 30,
			log_dir TEXT NOT NULL,
			db_path TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS log_archives (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT NOT NULL UNIQUE,
			file_size INTEGER NOT NULL DEFAULT 0,
			compressed INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_archives_created ON log_archives(created_at)`,
	}

	for _, s := range schemas {
		if _, err := r.db.Exec(s); err != nil {
			return fmt.Errorf("exec schema failed: %w", err)
		}
	}
	return nil
}

func (r *Repository) initDefaultConfig() error {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM log_config").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = r.db.Exec(
			`INSERT INTO log_config (max_file_size_mb, rotate_hour, rotate_minute, max_archive_files, log_dir, db_path)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			config.DefaultMaxFileSizeMB,
			config.DefaultRotateHour,
			config.DefaultRotateMinute,
			config.DefaultMaxArchiveFiles,
			config.DefaultLogDir,
			config.DefaultDBPath,
		)
	}
	return err
}

func (r *Repository) GetConfig() (*model.LogConfig, error) {
	var cfg model.LogConfig
	err := r.db.QueryRow(
		`SELECT id, max_file_size_mb, rotate_hour, rotate_minute, max_archive_files, log_dir, db_path, updated_at
		 FROM log_config ORDER BY id DESC LIMIT 1`,
	).Scan(
		&cfg.ID,
		&cfg.MaxFileSizeMB,
		&cfg.RotateHour,
		&cfg.RotateMinute,
		&cfg.MaxArchiveFiles,
		&cfg.LogDir,
		&cfg.DBPath,
		&cfg.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no config found")
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *Repository) UpdateConfig(cfg *model.LogConfig) error {
	result, err := r.db.Exec(
		`UPDATE log_config 
		 SET max_file_size_mb = ?, rotate_hour = ?, rotate_minute = ?, max_archive_files = ?, 
		     log_dir = ?, db_path = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		cfg.MaxFileSizeMB,
		cfg.RotateHour,
		cfg.RotateMinute,
		cfg.MaxArchiveFiles,
		cfg.LogDir,
		cfg.DBPath,
		cfg.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		_, err = r.db.Exec(
			`INSERT INTO log_config (max_file_size_mb, rotate_hour, rotate_minute, max_archive_files, log_dir, db_path)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			cfg.MaxFileSizeMB,
			cfg.RotateHour,
			cfg.RotateMinute,
			cfg.MaxArchiveFiles,
			cfg.LogDir,
			cfg.DBPath,
		)
	}
	return err
}

func (r *Repository) AddArchive(fileName string, fileSize int64, compressed bool) error {
	_, err := r.db.Exec(
		`INSERT INTO log_archives (file_name, file_size, compressed, created_at)
		 VALUES (?, ?, ?, ?)`,
		fileName,
		fileSize,
		compressed,
		time.Now(),
	)
	return err
}

func (r *Repository) GetArchives(limit int) ([]*model.LogArchive, error) {
	rows, err := r.db.Query(
		`SELECT id, file_name, file_size, compressed, created_at, deleted
		 FROM log_archives 
		 WHERE deleted = 0
		 ORDER BY created_at DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var archives []*model.LogArchive
	for rows.Next() {
		var a model.LogArchive
		var compressedInt, deletedInt int
		if err := rows.Scan(&a.ID, &a.FileName, &a.FileSize, &compressedInt, &a.CreatedAt, &deletedInt); err != nil {
			log.Printf("scan archive failed: %v", err)
			continue
		}
		a.Compressed = compressedInt == 1
		a.Deleted = deletedInt == 1
		archives = append(archives, &a)
	}
	return archives, nil
}

func (r *Repository) GetArchiveByName(fileName string) (*model.LogArchive, error) {
	var a model.LogArchive
	var compressedInt, deletedInt int
	err := r.db.QueryRow(
		`SELECT id, file_name, file_size, compressed, created_at, deleted
		 FROM log_archives WHERE file_name = ? AND deleted = 0`,
		fileName,
	).Scan(&a.ID, &a.FileName, &a.FileSize, &compressedInt, &a.CreatedAt, &deletedInt)
	if err != nil {
		return nil, err
	}
	a.Compressed = compressedInt == 1
	a.Deleted = deletedInt == 1
	return &a, nil
}

func (r *Repository) MarkArchiveDeleted(id int64) error {
	_, err := r.db.Exec(
		`UPDATE log_archives SET deleted = 1 WHERE id = ?`,
		id,
	)
	return err
}

func (r *Repository) UpdateArchiveCompressed(fileName string) error {
	_, err := r.db.Exec(
		`UPDATE log_archives SET compressed = 1 WHERE file_name = ?`,
		fileName,
	)
	return err
}

func (r *Repository) Close() error {
	return r.db.Close()
}
