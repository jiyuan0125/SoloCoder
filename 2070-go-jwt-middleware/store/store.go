package store

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"jwt-middleware/config"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db   *sql.DB
	mu   sync.RWMutex
}

var instance *Store
var once sync.Once

func NewStore(dbPath string) (*Store, error) {
	var err error
	once.Do(func() {
		var db *sql.DB
		db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
		if err != nil {
			return
		}
		if err = db.Ping(); err != nil {
			return
		}
		store := &Store{db: db}
		if err = store.init(); err != nil {
			return
		}
		instance = store
	})
	return instance, err
}

func (s *Store) init() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS blacklist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		jti TEXT NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		expires_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL
	)`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		jti TEXT NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		role TEXT NOT NULL,
		expires_at INTEGER NOT NULL,
		is_used INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL
	)`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		total_tokens_issued INTEGER NOT NULL DEFAULT 0,
		total_tokens_revoked INTEGER NOT NULL DEFAULT 0,
		total_refreshes INTEGER NOT NULL DEFAULT 0,
		total_users INTEGER NOT NULL DEFAULT 0,
		active_tokens INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL
	)`)
	if err != nil {
		return err
	}
	row := s.db.QueryRow(`SELECT COUNT(*) FROM stats`)
	var count int
	row.Scan(&count)
	if count == 0 {
		_, err = s.db.Exec(`
		INSERT INTO stats (total_tokens_issued, total_tokens_revoked, total_refreshes, total_users, active_tokens, updated_at)
		VALUES (0, 0, 0, 0, 0, ?)`, time.Now().Unix())
	}
	return err
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) AddToBlacklist(jti, userID string, expiresAt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
	INSERT OR IGNORE INTO blacklist (jti, user_id, expires_at, created_at)
	VALUES (?, ?, ?, ?)
	`, jti, userID, expiresAt, time.Now().Unix())
	return err
}

func (s *Store) IsBlacklisted(jti string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var count int
	err := s.db.QueryRow(`
	SELECT COUNT(*) FROM blacklist 
	WHERE jti = ? AND expires_at > ?
	`, jti, time.Now().Unix()).Scan(&count)
	return count > 0, err
}

func (s *Store) RevokeAllTokens(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`
	SELECT jti, expires_at FROM refresh_tokens 
	WHERE user_id = ? AND is_used = 0 AND expires_at > ?
	`, userID, time.Now().Unix())
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var jti string
		var exp int64
		if err := rows.Scan(&jti, &exp); err != nil {
			return err
		}
		_, err = tx.Exec(`
		INSERT OR IGNORE INTO blacklist (jti, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
		`, jti, userID, exp, time.Now().Unix())
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`
	UPDATE refresh_tokens SET is_used = 1 
	WHERE user_id = ? AND is_used = 0
	`, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) StoreRefreshToken(jti, userID, role string, expiresAt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
	INSERT OR REPLACE INTO refresh_tokens (jti, user_id, role, expires_at, is_used, created_at)
	VALUES (?, ?, ?, ?, 0, ?)
	`, jti, userID, role, expiresAt, time.Now().Unix())
	return err
}

func (s *Store) UseRefreshToken(jti, userID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var isUsed int
	var storedUserID string
	err = tx.QueryRow(`
	SELECT is_used, user_id FROM refresh_tokens WHERE jti = ?
	`, jti).Scan(&isUsed, &storedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if storedUserID != userID {
		return false, nil
	}
	if isUsed == 1 {
		return false, nil
	}
	_, err = tx.Exec(`
	UPDATE refresh_tokens SET is_used = 1 WHERE jti = ?
	`, jti)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	return err == nil, err
}

func (s *Store) IncrementTokensIssued(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	UPDATE stats 
	SET total_tokens_issued = total_tokens_issued + 1,
	    active_tokens = active_tokens + 1,
	    updated_at = ?
	`, time.Now().Unix())
	if err != nil {
		return err
	}
	var userExists int
	err = tx.QueryRow(`
	SELECT COUNT(DISTINCT user_id) FROM refresh_tokens WHERE user_id = ?
	`, userID).Scan(&userExists)
	if err != nil {
		return err
	}
	if userExists == 0 {
		_, err = tx.Exec(`
		UPDATE stats 
		SET total_users = total_users + 1,
		    updated_at = ?
		`, time.Now().Unix())
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) IncrementTokensRevoked() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
	UPDATE stats 
	SET total_tokens_revoked = total_tokens_revoked + 1,
	    active_tokens = active_tokens - 1,
	    updated_at = ?
	`, time.Now().Unix())
	return err
}

func (s *Store) IncrementRefreshes() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
	UPDATE stats 
	SET total_refreshes = total_refreshes + 1,
	    updated_at = ?
	`, time.Now().Unix())
	return err
}

func (s *Store) GetStats() (map[string]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var totalIssued, totalRevoked, totalRefreshes, totalUsers, activeTokens int64
	err := s.db.QueryRow(`
	SELECT total_tokens_issued, total_tokens_revoked, total_refreshes, total_users, active_tokens
	FROM stats ORDER BY id DESC LIMIT 1
	`).Scan(&totalIssued, &totalRevoked, &totalRefreshes, &totalUsers, &activeTokens)
	if err != nil {
		return nil, err
	}
	realActive, err := s.countActiveTokens()
	if err == nil {
		if realActive != activeTokens {
			log.Printf("Stats mismatch: stored=%d, actual=%d, correcting", activeTokens, realActive)
			s.db.Exec(`UPDATE stats SET active_tokens = ?`, realActive)
			activeTokens = realActive
		}
	}
	return map[string]int64{
		"total_tokens_issued": totalIssued,
		"total_tokens_revoked": totalRevoked,
		"total_refreshes": totalRefreshes,
		"total_users": totalUsers,
		"active_tokens": activeTokens,
	}, nil
}

func (s *Store) countActiveTokens() (int64, error) {
	var count int64
	now := time.Now().Unix()
	row := s.db.QueryRow(`
	SELECT COUNT(*) FROM (
	    SELECT jti FROM refresh_tokens WHERE is_used = 0 AND expires_at > ?
	    UNION
	    SELECT jti FROM blacklist WHERE expires_at > ?
	) combined
	`, now, now)
	err := row.Scan(&count)
	return count, err
}

func (s *Store) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	_, err := s.db.Exec(`DELETE FROM blacklist WHERE expires_at < ?`, now)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < ?`, now)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func init() {
	_, err := NewStore(config.DBPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}
}
