package database

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"
)

type QueryHistory struct {
	ID         int64     `json:"id"`
	Domain     string    `json:"domain"`
	RecordType string    `json:"record_type"`
	DNSServer  string    `json:"dns_server"`
	Result     string    `json:"result"`
	QueryTime  int64     `json:"query_time_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

type DB struct {
	db *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &DB{db: db}
	if err := database.init(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS query_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL,
		record_type TEXT NOT NULL,
		dns_server TEXT,
		result TEXT,
		query_time_ms INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS health_checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL UNIQUE,
		last_ip TEXT,
		last_check_at DATETIME,
		next_check_at DATETIME,
		interval_seconds INTEGER DEFAULT 60,
		enabled INTEGER DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS whois_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL UNIQUE,
		registrar TEXT,
		creation_date TEXT,
		expiration_date TEXT,
		updated_date TEXT,
		name_servers TEXT,
		raw_result TEXT,
		queried_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_query_history_domain ON query_history(domain);
	CREATE INDEX IF NOT EXISTS idx_query_history_created_at ON query_history(created_at);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *DB) SaveQueryHistory(domain, recordType, dnsServer string, result interface{}, queryTime time.Duration) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO query_history (domain, record_type, dns_server, result, query_time_ms, created_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = d.db.Exec(
		query,
		domain,
		recordType,
		dnsServer,
		string(resultJSON),
		queryTime.Milliseconds(),
		time.Now(),
	)
	return err
}

func (d *DB) GetQueryHistory(limit int) ([]QueryHistory, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
	SELECT id, domain, record_type, dns_server, result, query_time_ms, created_at
	FROM query_history
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []QueryHistory
	for rows.Next() {
		var h QueryHistory
		err := rows.Scan(
			&h.ID,
			&h.Domain,
			&h.RecordType,
			&h.DNSServer,
			&h.Result,
			&h.QueryTime,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}

	return histories, nil
}
