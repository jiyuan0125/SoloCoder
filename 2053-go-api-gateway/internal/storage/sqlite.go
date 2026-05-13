package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/example/apigateway/internal/model"
)

type SQLiteStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

var storage *SQLiteStorage
var once sync.Once

func GetStorage() *SQLiteStorage {
	once.Do(func() {
		dbPath := os.Getenv("SQLITE_DB_PATH")
		if dbPath == "" {
			dbPath = "./apigateway.db"
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Fatalf("failed to open sqlite: %v", err)
		}
		db.SetMaxOpenConns(1)

		s := &SQLiteStorage{db: db}
		if err := s.init(); err != nil {
			log.Fatalf("failed to init db: %v", err)
		}
		storage = s
	})
	return storage
}

func (s *SQLiteStorage) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS route_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL,
		match_type TEXT NOT NULL,
		headers TEXT NOT NULL DEFAULT '{}',
		target_url TEXT NOT NULL,
		timeout INTEGER NOT NULL DEFAULT 30,
		auth_type TEXT NOT NULL DEFAULT 'none',
		rate_limit_ip INTEGER NOT NULL DEFAULT 0,
		rate_limit_key INTEGER NOT NULL DEFAULT 0,
		healthy INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS route_changes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_id INTEGER,
		operation TEXT NOT NULL,
		old_config TEXT,
		new_config TEXT,
		status TEXT NOT NULL DEFAULT 'draft',
		applicant TEXT,
		approver1 TEXT,
		approver2 TEXT,
		final_approver TEXT,
		reason TEXT,
		applied_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		expires_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS jwt_secrets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		secret TEXT NOT NULL,
		issuer TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS access_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_path TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		duration INTEGER NOT NULL,
		client_id TEXT,
		client_ip TEXT,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS rate_limits (
		key TEXT PRIMARY KEY,
		count INTEGER NOT NULL DEFAULT 0,
		window DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_route_rules_path ON route_rules(path);
	CREATE INDEX IF NOT EXISTS idx_route_changes_route_id ON route_changes(route_id);
	CREATE INDEX IF NOT EXISTS idx_access_logs_timestamp ON access_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_rate_limits_window ON rate_limits(window);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) CreateRoute(rule *model.RouteRule) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	headersJSON, _ := json.Marshal(rule.Headers)
	if rule.Headers == nil {
		headersJSON = []byte("{}")
	}

	res, err := s.db.Exec(`
		INSERT INTO route_rules (name, path, match_type, headers, target_url, timeout, auth_type, rate_limit_ip, rate_limit_key, healthy)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rule.Name, rule.Path, rule.MatchType, string(headersJSON), rule.TargetURL, int(rule.Timeout.Seconds()), rule.AuthType, rule.RateLimitIP, rule.RateLimitKey, rule.Healthy)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStorage) UpdateRoute(rule *model.RouteRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	headersJSON, _ := json.Marshal(rule.Headers)
	if rule.Headers == nil {
		headersJSON = []byte("{}")
	}

	_, err := s.db.Exec(`
		UPDATE route_rules SET name=?, path=?, match_type=?, headers=?, target_url=?, timeout=?, auth_type=?, rate_limit_ip=?, rate_limit_key=?, healthy=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, rule.Name, rule.Path, rule.MatchType, string(headersJSON), rule.TargetURL, int(rule.Timeout.Seconds()), rule.AuthType, rule.RateLimitIP, rule.RateLimitKey, rule.Healthy, rule.ID)
	return err
}

func (s *SQLiteStorage) DeleteRoute(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec("DELETE FROM route_rules WHERE id=?", id)
	return err
}

func (s *SQLiteStorage) GetRoute(id int64) (*model.RouteRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rule model.RouteRule
	var headersStr string
	var timeoutSec int
	var healthyInt int

	err := s.db.QueryRow(`
		SELECT id, name, path, match_type, headers, target_url, timeout, auth_type, rate_limit_ip, rate_limit_key, healthy, created_at, updated_at
		FROM route_rules WHERE id=?
	`, id).Scan(&rule.ID, &rule.Name, &rule.Path, &rule.MatchType, &headersStr, &rule.TargetURL, &timeoutSec, &rule.AuthType, &rule.RateLimitIP, &rule.RateLimitKey, &healthyInt, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rule.Timeout = time.Duration(timeoutSec) * time.Second
	rule.Healthy = healthyInt != 0
	json.Unmarshal([]byte(headersStr), &rule.Headers)
	return &rule, nil
}

func (s *SQLiteStorage) ListRoutes() ([]*model.RouteRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, name, path, match_type, headers, target_url, timeout, auth_type, rate_limit_ip, rate_limit_key, healthy, created_at, updated_at
		FROM route_rules ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*model.RouteRule
	for rows.Next() {
		var rule model.RouteRule
		var headersStr string
		var timeoutSec int
		var healthyInt int

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Path, &rule.MatchType, &headersStr, &rule.TargetURL, &timeoutSec, &rule.AuthType, &rule.RateLimitIP, &rule.RateLimitKey, &healthyInt, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}

		rule.Timeout = time.Duration(timeoutSec) * time.Second
		rule.Healthy = healthyInt != 0
		json.Unmarshal([]byte(headersStr), &rule.Headers)
		rules = append(rules, &rule)
	}
	return rules, nil
}

func (s *SQLiteStorage) CreateRouteChange(change *model.RouteChange) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec(`
		INSERT INTO route_changes (route_id, operation, old_config, new_config, status, applicant, approver1, approver2, final_approver, reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, change.RouteID, change.Operation, change.OldConfig, change.NewConfig, change.Status, change.Applicant, change.Approver1, change.Approver2, change.FinalApprover, change.Reason)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStorage) UpdateRouteChange(change *model.RouteChange) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		UPDATE route_changes SET route_id=?, operation=?, old_config=?, new_config=?, status=?, applicant=?, approver1=?, approver2=?, final_approver=?, reason=?, applied_at=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, change.RouteID, change.Operation, change.OldConfig, change.NewConfig, change.Status, change.Applicant, change.Approver1, change.Approver2, change.FinalApprover, change.Reason, change.AppliedAt, change.ID)
	return err
}

func (s *SQLiteStorage) GetRouteChange(id int64) (*model.RouteChange, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var change model.RouteChange
	var appliedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, route_id, operation, old_config, new_config, status, applicant, approver1, approver2, final_approver, reason, applied_at, created_at, updated_at
		FROM route_changes WHERE id=?
	`, id).Scan(&change.ID, &change.RouteID, &change.Operation, &change.OldConfig, &change.NewConfig, &change.Status, &change.Applicant, &change.Approver1, &change.Approver2, &change.FinalApprover, &change.Reason, &appliedAt, &change.CreatedAt, &change.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if appliedAt.Valid {
		t := appliedAt.Time
		change.AppliedAt = &t
	}
	return &change, nil
}

func (s *SQLiteStorage) ListRouteChanges() ([]*model.RouteChange, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT id, route_id, operation, old_config, new_config, status, applicant, approver1, approver2, final_approver, reason, applied_at, created_at, updated_at
		FROM route_changes ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*model.RouteChange
	for rows.Next() {
		var change model.RouteChange
		var appliedAt sql.NullTime

		if err := rows.Scan(&change.ID, &change.RouteID, &change.Operation, &change.OldConfig, &change.NewConfig, &change.Status, &change.Applicant, &change.Approver1, &change.Approver2, &change.FinalApprover, &change.Reason, &appliedAt, &change.CreatedAt, &change.UpdatedAt); err != nil {
			return nil, err
		}

		if appliedAt.Valid {
			t := appliedAt.Time
			change.AppliedAt = &t
		}
		changes = append(changes, &change)
	}
	return changes, nil
}

func (s *SQLiteStorage) CreateAPIKey(key *model.APIKey) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec(`
		INSERT INTO api_keys (key, user_id, expires_at) VALUES (?, ?, ?)
	`, key.Key, key.UserID, key.ExpiresAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStorage) GetAPIKey(keyStr string) (*model.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var key model.APIKey
	var expiresAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, key, user_id, expires_at, created_at FROM api_keys WHERE key=?
	`, keyStr).Scan(&key.ID, &key.Key, &key.UserID, &expiresAt, &key.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if expiresAt.Valid {
		t := expiresAt.Time
		key.ExpiresAt = &t
	}
	return &key, nil
}

func (s *SQLiteStorage) ListAPIKeys() ([]*model.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, key, user_id, expires_at, created_at FROM api_keys ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*model.APIKey
	for rows.Next() {
		var key model.APIKey
		var expiresAt sql.NullTime

		if err := rows.Scan(&key.ID, &key.Key, &key.UserID, &expiresAt, &key.CreatedAt); err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			t := expiresAt.Time
			key.ExpiresAt = &t
		}
		keys = append(keys, &key)
	}
	return keys, nil
}

func (s *SQLiteStorage) DeleteAPIKey(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec("DELETE FROM api_keys WHERE id=?", id)
	return err
}

func (s *SQLiteStorage) CreateJWTSecret(secret *model.JWTSecret) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec(`
		INSERT INTO jwt_secrets (secret, issuer) VALUES (?, ?)
	`, secret.Secret, secret.Issuer)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStorage) GetJWTSecret() (*model.JWTSecret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var secret model.JWTSecret
	err := s.db.QueryRow("SELECT id, secret, issuer, created_at FROM jwt_secrets ORDER BY id DESC LIMIT 1").Scan(
		&secret.ID, &secret.Secret, &secret.Issuer, &secret.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &secret, nil
}

func (s *SQLiteStorage) InsertAccessLog(logEntry *model.AccessLog) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO access_logs (request_path, status_code, duration, client_id, client_ip)
		VALUES (?, ?, ?, ?, ?)
	`, logEntry.RequestPath, logEntry.StatusCode, logEntry.Duration.Milliseconds(), logEntry.ClientID, logEntry.ClientIP)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to insert access log: %v\n", err)
	}
}

func (s *SQLiteStorage) GetAndIncrementRateLimit(key string, window time.Time, limit int) (int, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow("SELECT count FROM rate_limits WHERE key=? AND window=?", key, window).Scan(&count)

	if err == sql.ErrNoRows {
		count = 1
		_, err = tx.Exec("INSERT INTO rate_limits (key, count, window) VALUES (?, 1, ?)", key, window)
		if err != nil {
			return 0, false, err
		}
	} else if err != nil {
		return 0, false, err
	} else {
		count++
		if count > limit {
			return count, false, tx.Commit()
		}
		_, err = tx.Exec("UPDATE rate_limits SET count=? WHERE key=? AND window=?", count, key, window)
		if err != nil {
			return 0, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return count, count <= limit, nil
}

func (s *SQLiteStorage) CleanupOldRateLimits(cutoff time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec("DELETE FROM rate_limits WHERE window < ?", cutoff)
	return err
}
