package store

import (
	"database/sql"
	"time"

	"filedownload/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		stored_path TEXT NOT NULL,
		size INTEGER NOT NULL,
		upload_time DATETIME NOT NULL,
		expire_time DATETIME NOT NULL,
		max_downloads INTEGER NOT NULL,
		current_download INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
	CREATE INDEX IF NOT EXISTS idx_files_filename ON files(filename);
	
	CREATE TABLE IF NOT EXISTS download_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		ip TEXT NOT NULL,
		user_agent TEXT NOT NULL,
		time DATETIME NOT NULL,
		FOREIGN KEY(file_id) REFERENCES files(id)
	);
	CREATE INDEX IF NOT EXISTS idx_records_file_id ON download_records(file_id);
	CREATE INDEX IF NOT EXISTS idx_records_user_id ON download_records(user_id);
	CREATE INDEX IF NOT EXISTS idx_records_time ON download_records(time);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) SaveFile(file *model.File) error {
	query := `INSERT INTO files (id, user_id, filename, stored_path, size, upload_time, expire_time, max_downloads, current_download) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, file.ID, file.UserID, file.Filename, file.StoredPath, file.Size, file.UploadTime, file.ExpireTime, file.MaxDownloads, file.CurrentDownload)
	return err
}

func (s *SQLiteStore) GetFileByID(id string) (*model.File, error) {
	query := `SELECT id, user_id, filename, stored_path, size, upload_time, expire_time, max_downloads, current_download FROM files WHERE id = ?`
	row := s.db.QueryRow(query, id)
	var file model.File
	err := row.Scan(&file.ID, &file.UserID, &file.Filename, &file.StoredPath, &file.Size, &file.UploadTime, &file.ExpireTime, &file.MaxDownloads, &file.CurrentDownload)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *SQLiteStore) GetFileByUserAndPath(userID, filename string) (*model.File, error) {
	query := `SELECT id, user_id, filename, stored_path, size, upload_time, expire_time, max_downloads, current_download FROM files WHERE user_id = ? AND filename = ? ORDER BY upload_time DESC LIMIT 1`
	row := s.db.QueryRow(query, userID, filename)
	var file model.File
	err := row.Scan(&file.ID, &file.UserID, &file.Filename, &file.StoredPath, &file.Size, &file.UploadTime, &file.ExpireTime, &file.MaxDownloads, &file.CurrentDownload)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *SQLiteStore) UpdateDownloadCount(fileID string) error {
	query := `UPDATE files SET current_download = current_download + 1 WHERE id = ?`
	_, err := s.db.Exec(query, fileID)
	return err
}

func (s *SQLiteStore) SaveDownloadRecord(record *model.DownloadRecord) error {
	query := `INSERT INTO download_records (file_id, user_id, filename, ip, user_agent, time) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, record.FileID, record.UserID, record.Filename, record.IP, record.UserAgent, record.Time)
	return err
}

func (s *SQLiteStore) ListUserFiles(userID string) ([]*model.File, error) {
	query := `SELECT id, user_id, filename, stored_path, size, upload_time, expire_time, max_downloads, current_download FROM files WHERE user_id = ? ORDER BY upload_time DESC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var file model.File
		if err := rows.Scan(&file.ID, &file.UserID, &file.Filename, &file.StoredPath, &file.Size, &file.UploadTime, &file.ExpireTime, &file.MaxDownloads, &file.CurrentDownload); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, nil
}

func (s *SQLiteStore) SearchFiles(userID, keyword string) ([]*model.File, error) {
	query := `SELECT id, user_id, filename, stored_path, size, upload_time, expire_time, max_downloads, current_download FROM files WHERE user_id = ? AND filename LIKE ? ORDER BY upload_time DESC`
	rows, err := s.db.Query(query, userID, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var file model.File
		if err := rows.Scan(&file.ID, &file.UserID, &file.Filename, &file.StoredPath, &file.Size, &file.UploadTime, &file.ExpireTime, &file.MaxDownloads, &file.CurrentDownload); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, nil
}

func (s *SQLiteStore) ListAllFiles() ([]*model.File, error) {
	query := `SELECT id, user_id, filename, stored_path, size, upload_time, expire_time, max_downloads, current_download FROM files ORDER BY upload_time DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var file model.File
		if err := rows.Scan(&file.ID, &file.UserID, &file.Filename, &file.StoredPath, &file.Size, &file.UploadTime, &file.ExpireTime, &file.MaxDownloads, &file.CurrentDownload); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	return files, nil
}

func (s *SQLiteStore) ListAllDownloadRecords() ([]*model.DownloadRecord, error) {
	query := `SELECT id, file_id, user_id, filename, ip, user_agent, time FROM download_records ORDER BY time DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.DownloadRecord
	for rows.Next() {
		var record model.DownloadRecord
		if err := rows.Scan(&record.ID, &record.FileID, &record.UserID, &record.Filename, &record.IP, &record.UserAgent, &record.Time); err != nil {
			return nil, err
		}
		records = append(records, &record)
	}
	return records, nil
}

func (s *SQLiteStore) IsLinkExpired(file *model.File) bool {
	if !file.ExpireTime.IsZero() && file.ExpireTime.Before(time.Now()) {
		return true
	}
	if file.MaxDownloads > 0 && file.CurrentDownload >= file.MaxDownloads {
		return true
	}
	return false
}
