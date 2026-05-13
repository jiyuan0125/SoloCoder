package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	"json-validator/config"
	"json-validator/validator"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type ValidationRecord struct {
	ID        int64
	Schema    string
	Data      string
	Result    string
	Valid     bool
	CreatedAt time.Time
}

var store *Store

func InitDB() error {
	db, err := sql.Open("sqlite3", config.DBPath)
	if err != nil {
		return err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS validations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		schema TEXT NOT NULL,
		data TEXT NOT NULL,
		result TEXT NOT NULL,
		valid INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	store = &Store{db: db}
	return nil
}

func GetStore() *Store {
	return store
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) SaveValidation(schema, data string, result *validator.ValidationResult) (int64, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return 0, err
	}

	res, err := s.db.Exec(`
		INSERT INTO validations (schema, data, result, valid, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, schema, data, string(resultJSON), result.Valid, time.Now())
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) GetValidation(id int64) (*ValidationRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, schema, data, result, valid, created_at
		FROM validations
		WHERE id = ?
	`, id)

	record := &ValidationRecord{}
	var validInt int
	err := row.Scan(&record.ID, &record.Schema, &record.Data, &record.Result, &validInt, &record.CreatedAt)
	if err != nil {
		return nil, err
	}
	record.Valid = validInt == 1

	return record, nil
}

func (s *Store) GetRecentValidations(limit int) ([]*ValidationRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, schema, data, result, valid, created_at
		FROM validations
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*ValidationRecord, 0)
	for rows.Next() {
		record := &ValidationRecord{}
		var validInt int
		err := rows.Scan(&record.ID, &record.Schema, &record.Data, &record.Result, &validInt, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		record.Valid = validInt == 1
		records = append(records, record)
	}

	return records, nil
}
