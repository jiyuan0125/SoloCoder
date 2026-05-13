package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"config-reload/config"
)

type ConfigHistory struct {
	ID          int64
	Timestamp   time.Time
	Hash        string
	OldHash     string
	ChangeType  string
	ServiceName string
	ConfigData  string
	Message     string
	Success     bool
}

type Storage struct {
	mu sync.Mutex
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.initTables(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS config_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			hash TEXT NOT NULL,
			old_hash TEXT,
			change_type TEXT NOT NULL,
			service_name TEXT,
			config_data TEXT,
			message TEXT,
			success INTEGER DEFAULT 1
		)`,
		`CREATE INDEX IF NOT EXISTS idx_config_history_hash ON config_history(hash)`,
		`CREATE INDEX IF NOT EXISTS idx_config_history_timestamp ON config_history(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_config_history_success ON config_history(success)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	return nil
}

func (s *Storage) RecordChange(change *config.ConfigChange, configData string, message string, success bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(change.Added) > 0 {
		for _, name := range change.Added {
			if err := s.insertRecord(tx, change, "added", name, configData, message, success); err != nil {
				return err
			}
		}
	}
	if len(change.Removed) > 0 {
		for _, name := range change.Removed {
			if err := s.insertRecord(tx, change, "removed", name, configData, message, success); err != nil {
				return err
			}
		}
	}
	if len(change.Modified) > 0 {
		for _, name := range change.Modified {
			if err := s.insertRecord(tx, change, "modified", name, configData, message, success); err != nil {
				return err
			}
		}
	}

	if len(change.Added)+len(change.Removed)+len(change.Modified) == 0 {
		if err := s.insertRecord(tx, change, "full_reload", "", configData, message, success); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) RecordFailure(hash, oldHash, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO config_history (timestamp, hash, old_hash, change_type, message, success)
		 VALUES (?, ?, ?, ?, ?, 0)`,
		time.Now(),
		hash,
		oldHash,
		"failed",
		message,
	)
	return err
}

func (s *Storage) insertRecord(tx *sql.Tx, change *config.ConfigChange, changeType, serviceName, configData, message string, success bool) error {
	_, err := tx.Exec(
		`INSERT INTO config_history (timestamp, hash, old_hash, change_type, service_name, config_data, message, success)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		change.Timestamp,
		change.Hash,
		change.OldHash,
		changeType,
		serviceName,
		configData,
		message,
		boolToInt(success),
	)
	return err
}

func (s *Storage) GetRecentHistory(limit int) ([]ConfigHistory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(
		`SELECT id, timestamp, hash, old_hash, change_type, service_name, config_data, message, success
		 FROM config_history
		 ORDER BY timestamp DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ConfigHistory
	for rows.Next() {
		var rec ConfigHistory
		var successInt int
		err := rows.Scan(
			&rec.ID,
			&rec.Timestamp,
			&rec.Hash,
			&rec.OldHash,
			&rec.ChangeType,
			&rec.ServiceName,
			&rec.ConfigData,
			&rec.Message,
			&successInt,
		)
		if err != nil {
			return nil, err
		}
		rec.Success = successInt == 1
		records = append(records, rec)
	}
	return records, nil
}

func (s *Storage) LogInfo(message string) {
	log.Printf("[INFO] %s", message)
}

func (s *Storage) LogError(message string) {
	log.Printf("[ERROR] %s", message)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ConfigToJSON(c *config.AppConfig) string {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return string(data)
}
