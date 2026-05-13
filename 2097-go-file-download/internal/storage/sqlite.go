package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"filedownload/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

var GlobalStorage *Storage

func New(dbPath string) (*Storage, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.initSchema(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		file_name TEXT NOT NULL,
		original_name TEXT NOT NULL,
		size INTEGER NOT NULL,
		uploaded_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS download_links (
		id TEXT PRIMARY KEY,
		file_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		max_downloads INTEGER NOT NULL,
		downloaded INTEGER NOT NULL DEFAULT 0,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (file_id) REFERENCES files(id)
	);

	CREATE TABLE IF NOT EXISTS download_records (
		id TEXT PRIMARY KEY,
		link_id TEXT NOT NULL,
		file_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		downloaded_at DATETIME NOT NULL,
		ip_address TEXT NOT NULL,
		FOREIGN KEY (link_id) REFERENCES download_links(id),
		FOREIGN KEY (file_id) REFERENCES files(id)
	);

	CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
	CREATE INDEX IF NOT EXISTS idx_files_original_name ON files(original_name);
	CREATE INDEX IF NOT EXISTS idx_download_links_file_id ON download_links(file_id);
	CREATE INDEX IF NOT EXISTS idx_download_records_file_id ON download_records(file_id);
	CREATE INDEX IF NOT EXISTS idx_download_records_user_id ON download_records(user_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) CreateFile(file *models.File) error {
	query := `
	INSERT INTO files (id, user_id, file_name, original_name, size, uploaded_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, file.ID, file.UserID, file.FileName, file.OriginalName, file.Size, file.UploadedAt)
	return err
}

func (s *Storage) GetFileByID(id string) (*models.File, error) {
	query := `
	SELECT id, user_id, file_name, original_name, size, uploaded_at
	FROM files WHERE id = ?
	`
	file := &models.File{}
	err := s.db.QueryRow(query, id).Scan(&file.ID, &file.UserID, &file.FileName, &file.OriginalName, &file.Size, &file.UploadedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return file, err
}

func (s *Storage) GetFilesByUserID(userID string) ([]*models.File, error) {
	query := `
	SELECT id, user_id, file_name, original_name, size, uploaded_at
	FROM files WHERE user_id = ? ORDER BY uploaded_at DESC
	`
	return s.queryFiles(query, userID)
}

func (s *Storage) SearchFiles(userID, keyword string) ([]*models.File, error) {
	query := `
	SELECT id, user_id, file_name, original_name, size, uploaded_at
	FROM files WHERE user_id = ? AND original_name LIKE ? ORDER BY uploaded_at DESC
	`
	return s.queryFiles(query, userID, "%"+keyword+"%")
}

func (s *Storage) GetAllFiles() ([]*models.File, error) {
	query := `
	SELECT id, user_id, file_name, original_name, size, uploaded_at
	FROM files ORDER BY uploaded_at DESC
	`
	return s.queryFiles(query)
}

func (s *Storage) queryFiles(query string, args ...interface{}) ([]*models.File, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		file := &models.File{}
		err := rows.Scan(&file.ID, &file.UserID, &file.FileName, &file.OriginalName, &file.Size, &file.UploadedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (s *Storage) CreateDownloadLink(link *models.DownloadLink) error {
	query := `
	INSERT INTO download_links (id, file_id, user_id, max_downloads, downloaded, expires_at, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, link.ID, link.FileID, link.UserID, link.MaxDownloads, link.Downloaded, link.ExpiresAt, link.CreatedAt)
	return err
}

func (s *Storage) GetDownloadLink(id string) (*models.DownloadLink, error) {
	query := `
	SELECT id, file_id, user_id, max_downloads, downloaded, expires_at, created_at
	FROM download_links WHERE id = ?
	`
	link := &models.DownloadLink{}
	err := s.db.QueryRow(query, id).Scan(&link.ID, &link.FileID, &link.UserID, &link.MaxDownloads, &link.Downloaded, &link.ExpiresAt, &link.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return link, err
}

func (s *Storage) IncrementDownloads(linkID string) error {
	query := `UPDATE download_links SET downloaded = downloaded + 1 WHERE id = ?`
	_, err := s.db.Exec(query, linkID)
	return err
}

func (s *Storage) CreateDownloadRecord(record *models.DownloadRecord) error {
	query := `
	INSERT INTO download_records (id, link_id, file_id, user_id, downloaded_at, ip_address)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, record.ID, record.LinkID, record.FileID, record.UserID, record.DownloadedAt, record.IPAddress)
	return err
}

func (s *Storage) GetAllDownloadRecords() ([]*models.DownloadRecord, error) {
	query := `
	SELECT id, link_id, file_id, user_id, downloaded_at, ip_address
	FROM download_records ORDER BY downloaded_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.DownloadRecord
	for rows.Next() {
		record := &models.DownloadRecord{}
		err := rows.Scan(&record.ID, &record.LinkID, &record.FileID, &record.UserID, &record.DownloadedAt, &record.IPAddress)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Storage) GetFileByUserAndName(userID, originalName string) (*models.File, error) {
	query := `
	SELECT id, user_id, file_name, original_name, size, uploaded_at
	FROM files WHERE user_id = ? AND original_name = ?
	ORDER BY uploaded_at DESC LIMIT 1
	`
	file := &models.File{}
	err := s.db.QueryRow(query, userID, originalName).Scan(&file.ID, &file.UserID, &file.FileName, &file.OriginalName, &file.Size, &file.UploadedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return file, err
}

func (s *Storage) GetDownloadRecordsByUserID(userID string) ([]*models.DownloadRecord, error) {
	query := `
	SELECT id, link_id, file_id, user_id, downloaded_at, ip_address
	FROM download_records WHERE user_id = ? ORDER BY downloaded_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.DownloadRecord
	for rows.Next() {
		record := &models.DownloadRecord{}
		err := rows.Scan(&record.ID, &record.LinkID, &record.FileID, &record.UserID, &record.DownloadedAt, &record.IPAddress)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
