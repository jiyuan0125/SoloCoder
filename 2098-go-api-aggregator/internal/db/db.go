package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Endpoint struct {
	ID        int64
	Name      string
	URL       string
	Method    string
	Timeout   int
	Retries   int
	Headers   map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CacheEntry struct {
	ID        int64
	Key       string
	Data      []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}

type AggregationConfig struct {
	ID             int64
	Name           string
	EndpointIDs    []int64
	FormatType     string
	FormatConfig   map[string]interface{}
	CacheTTL       int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Record struct {
	ID        int64
	ConfigID  int64
	QueryHash string
	Total     float64
	ItemsJSON []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	dbInstance *sql.DB
	dbMutex    sync.Mutex
)

func Init(dbPath string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbInstance != nil {
		return nil
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping db: %w", err)
	}

	dbInstance = db

	if err := migrate(); err != nil {
		db.Close()
		dbInstance = nil
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func Close() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbInstance == nil {
		return nil
	}

	err := dbInstance.Close()
	dbInstance = nil
	return err
}

func getDB() *sql.DB {
	return dbInstance
}

func migrate() error {
	db := getDB()

	schema := `
	CREATE TABLE IF NOT EXISTS endpoints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		method TEXT NOT NULL DEFAULT 'GET',
		timeout INTEGER NOT NULL DEFAULT 5,
		retries INTEGER NOT NULL DEFAULT 0,
		headers TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS cache_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cache_key TEXT NOT NULL UNIQUE,
		data BLOB NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_cache_expires ON cache_entries(expires_at);

	CREATE TABLE IF NOT EXISTS aggregation_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		endpoint_ids TEXT NOT NULL,
		format_type TEXT NOT NULL DEFAULT 'merge',
		format_config TEXT,
		cache_ttl INTEGER NOT NULL DEFAULT 300,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_id INTEGER NOT NULL,
		query_hash TEXT NOT NULL,
		total REAL NOT NULL DEFAULT 0,
		items_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (config_id) REFERENCES aggregation_configs(id)
	);

	CREATE INDEX IF NOT EXISTS idx_records_config_query ON records(config_id, query_hash);
	`

	_, err := db.Exec(schema)
	return err
}

func CreateEndpoint(ep *Endpoint) error {
	db := getDB()
	headersJSON, _ := json.Marshal(ep.Headers)
	if ep.Headers == nil {
		headersJSON = []byte("{}")
	}

	result, err := db.Exec(`
		INSERT INTO endpoints (name, url, method, timeout, retries, headers)
		VALUES (?, ?, ?, ?, ?, ?)
	`, ep.Name, ep.URL, ep.Method, ep.Timeout, ep.Retries, string(headersJSON))
	if err != nil {
		return err
	}

	ep.ID, _ = result.LastInsertId()
	ep.CreatedAt = time.Now()
	ep.UpdatedAt = time.Now()
	return nil
}

func GetEndpoint(id int64) (*Endpoint, error) {
	db := getDB()
	row := db.QueryRow(`
		SELECT id, name, url, method, timeout, retries, headers, created_at, updated_at
		FROM endpoints WHERE id = ?
	`, id)

	return scanEndpoint(row)
}

func ListEndpoints() ([]*Endpoint, error) {
	db := getDB()
	rows, err := db.Query(`
		SELECT id, name, url, method, timeout, retries, headers, created_at, updated_at
		FROM endpoints ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eps []*Endpoint
	for rows.Next() {
		ep, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		eps = append(eps, ep)
	}
	return eps, rows.Err()
}

func UpdateEndpoint(ep *Endpoint) error {
	db := getDB()
	headersJSON, _ := json.Marshal(ep.Headers)
	if ep.Headers == nil {
		headersJSON = []byte("{}")
	}

	_, err := db.Exec(`
		UPDATE endpoints SET name=?, url=?, method=?, timeout=?, retries=?, headers=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, ep.Name, ep.URL, ep.Method, ep.Timeout, ep.Retries, string(headersJSON), ep.ID)
	return err
}

func DeleteEndpoint(id int64) error {
	db := getDB()
	_, err := db.Exec("DELETE FROM endpoints WHERE id=?", id)
	return err
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanEndpoint(sc rowScanner) (*Endpoint, error) {
	var headersStr string
	ep := &Endpoint{}
	err := sc.Scan(&ep.ID, &ep.Name, &ep.URL, &ep.Method, &ep.Timeout, &ep.Retries, &headersStr, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		return nil, err
	}

	ep.Headers = make(map[string]string)
	if headersStr != "" && headersStr != "null" {
		json.Unmarshal([]byte(headersStr), &ep.Headers)
	}
	return ep, nil
}

func GetCacheEntry(key string) (*CacheEntry, error) {
	db := getDB()
	row := db.QueryRow(`
		SELECT id, cache_key, data, expires_at, created_at
		FROM cache_entries WHERE cache_key = ? AND expires_at > CURRENT_TIMESTAMP
	`, key)

	ce := &CacheEntry{}
	err := row.Scan(&ce.ID, &ce.Key, &ce.Data, &ce.ExpiresAt, &ce.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ce, err
}

func SetCacheEntry(key string, data []byte, ttlSeconds int) error {
	db := getDB()
	expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)

	_, err := db.Exec(`
		INSERT OR REPLACE INTO cache_entries (cache_key, data, expires_at, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, key, data, expiresAt)
	return err
}

func CleanExpiredCache() error {
	db := getDB()
	_, err := db.Exec("DELETE FROM cache_entries WHERE expires_at < CURRENT_TIMESTAMP")
	return err
}

func CreateAggregationConfig(cfg *AggregationConfig) error {
	db := getDB()
	idsJSON, _ := json.Marshal(cfg.EndpointIDs)
	formatJSON, _ := json.Marshal(cfg.FormatConfig)
	if cfg.FormatConfig == nil {
		formatJSON = []byte("{}")
	}

	result, err := db.Exec(`
		INSERT INTO aggregation_configs (name, endpoint_ids, format_type, format_config, cache_ttl)
		VALUES (?, ?, ?, ?, ?)
	`, cfg.Name, string(idsJSON), cfg.FormatType, string(formatJSON), cfg.CacheTTL)
	if err != nil {
		return err
	}

	cfg.ID, _ = result.LastInsertId()
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()
	return nil
}

func GetAggregationConfig(id int64) (*AggregationConfig, error) {
	db := getDB()
	row := db.QueryRow(`
		SELECT id, name, endpoint_ids, format_type, format_config, cache_ttl, created_at, updated_at
		FROM aggregation_configs WHERE id = ?
	`, id)

	return scanAggregationConfig(row)
}

func ListAggregationConfigs() ([]*AggregationConfig, error) {
	db := getDB()
	rows, err := db.Query(`
		SELECT id, name, endpoint_ids, format_type, format_config, cache_ttl, created_at, updated_at
		FROM aggregation_configs ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cfgs []*AggregationConfig
	for rows.Next() {
		cfg, err := scanAggregationConfig(rows)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs, rows.Err()
}

func scanAggregationConfig(sc rowScanner) (*AggregationConfig, error) {
	var idsStr, formatStr string
	cfg := &AggregationConfig{}
	err := sc.Scan(&cfg.ID, &cfg.Name, &idsStr, &cfg.FormatType, &formatStr, &cfg.CacheTTL, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, err
	}

	cfg.EndpointIDs = []int64{}
	if idsStr != "" && idsStr != "null" {
		json.Unmarshal([]byte(idsStr), &cfg.EndpointIDs)
	}

	cfg.FormatConfig = make(map[string]interface{})
	if formatStr != "" && formatStr != "null" {
		json.Unmarshal([]byte(formatStr), &cfg.FormatConfig)
	}
	return cfg, nil
}

func CreateRecord(rec *Record) error {
	db := getDB()
	itemsStr := string(rec.ItemsJSON)
	if itemsStr == "" {
		itemsStr = "[]"
	}

	result, err := db.Exec(`
		INSERT INTO records (config_id, query_hash, total, items_json)
		VALUES (?, ?, ?, ?)
	`, rec.ConfigID, rec.QueryHash, rec.Total, itemsStr)
	if err != nil {
		return err
	}

	rec.ID, _ = result.LastInsertId()
	rec.CreatedAt = time.Now()
	rec.UpdatedAt = time.Now()
	return nil
}

func GetRecord(id int64) (*Record, error) {
	db := getDB()
	row := db.QueryRow(`
		SELECT id, config_id, query_hash, total, items_json, created_at, updated_at
		FROM records WHERE id = ?
	`, id)

	return scanRecord(row)
}

func GetRecordByQuery(configID int64, queryHash string) (*Record, error) {
	db := getDB()
	row := db.QueryRow(`
		SELECT id, config_id, query_hash, total, items_json, created_at, updated_at
		FROM records WHERE config_id = ? AND query_hash = ?
		ORDER BY id DESC LIMIT 1
	`, configID, queryHash)

	return scanRecord(row)
}

func UpdateRecord(rec *Record) error {
	db := getDB()
	itemsStr := string(rec.ItemsJSON)
	if itemsStr == "" {
		itemsStr = "[]"
	}

	_, err := db.Exec(`
		UPDATE records SET total=?, items_json=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, rec.Total, itemsStr, rec.ID)
	return err
}

func scanRecord(sc rowScanner) (*Record, error) {
	var itemsStr string
	rec := &Record{}
	err := sc.Scan(&rec.ID, &rec.ConfigID, &rec.QueryHash, &rec.Total, &itemsStr, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rec.ItemsJSON = []byte(itemsStr)
	return rec, nil
}
