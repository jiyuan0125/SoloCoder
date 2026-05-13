package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"connpool/internal/models"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initTables(); err != nil {
		return nil, fmt.Errorf("init tables: %w", err)
	}

	return storage, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) initTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pools (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		max_connections INTEGER NOT NULL,
		idle_timeout INTEGER NOT NULL,
		max_lifetime INTEGER NOT NULL,
		backend_address TEXT NOT NULL,
		wait_timeout INTEGER NOT NULL,
		health_check_url TEXT,
		quota INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS statistics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		total_active INTEGER NOT NULL,
		total_idle INTEGER NOT NULL,
		total_waiting INTEGER NOT NULL,
		total_max INTEGER NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_pools_name ON pools(name);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStorage) CreatePool(ctx context.Context, pool *models.PoolConfig) error {
	query := `
	INSERT INTO pools (id, name, max_connections, idle_timeout, max_lifetime, 
		backend_address, wait_timeout, health_check_url, quota, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		pool.ID,
		pool.Name,
		pool.MaxConnections,
		int64(pool.IdleTimeout),
		int64(pool.MaxLifetime),
		pool.BackendAddress,
		int64(pool.WaitTimeout),
		pool.HealthCheckURL,
		pool.Quota,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert pool: %w", err)
	}

	pool.CreatedAt = now
	pool.UpdatedAt = now
	return nil
}

func (s *SQLiteStorage) GetPool(ctx context.Context, id string) (*models.PoolConfig, error) {
	query := `
	SELECT id, name, max_connections, idle_timeout, max_lifetime, 
		backend_address, wait_timeout, health_check_url, quota, created_at, updated_at
	FROM pools WHERE id = ?
	`

	pool := &models.PoolConfig{}
	var idleTimeout, maxLifetime, waitTimeout int64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&pool.ID,
		&pool.Name,
		&pool.MaxConnections,
		&idleTimeout,
		&maxLifetime,
		&pool.BackendAddress,
		&waitTimeout,
		&pool.HealthCheckURL,
		&pool.Quota,
		&pool.CreatedAt,
		&pool.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query pool: %w", err)
	}

	pool.IdleTimeout = time.Duration(idleTimeout)
	pool.MaxLifetime = time.Duration(maxLifetime)
	pool.WaitTimeout = time.Duration(waitTimeout)

	return pool, nil
}

func (s *SQLiteStorage) ListPools(ctx context.Context) ([]*models.PoolConfig, error) {
	query := `
	SELECT id, name, max_connections, idle_timeout, max_lifetime, 
		backend_address, wait_timeout, health_check_url, quota, created_at, updated_at
	FROM pools ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list pools: %w", err)
	}
	defer rows.Close()

	var pools []*models.PoolConfig
	for rows.Next() {
		pool := &models.PoolConfig{}
		var idleTimeout, maxLifetime, waitTimeout int64

		err := rows.Scan(
			&pool.ID,
			&pool.Name,
			&pool.MaxConnections,
			&idleTimeout,
			&maxLifetime,
			&pool.BackendAddress,
			&waitTimeout,
			&pool.HealthCheckURL,
			&pool.Quota,
			&pool.CreatedAt,
			&pool.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pool: %w", err)
		}

		pool.IdleTimeout = time.Duration(idleTimeout)
		pool.MaxLifetime = time.Duration(maxLifetime)
		pool.WaitTimeout = time.Duration(waitTimeout)
		pools = append(pools, pool)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return pools, nil
}

func (s *SQLiteStorage) UpdatePool(ctx context.Context, pool *models.PoolConfig) error {
	query := `
	UPDATE pools SET 
		name = ?, max_connections = ?, idle_timeout = ?, max_lifetime = ?,
		backend_address = ?, wait_timeout = ?, health_check_url = ?, quota = ?, updated_at = ?
	WHERE id = ?
	`

	now := time.Now()
	result, err := s.db.ExecContext(ctx, query,
		pool.Name,
		pool.MaxConnections,
		int64(pool.IdleTimeout),
		int64(pool.MaxLifetime),
		pool.BackendAddress,
		int64(pool.WaitTimeout),
		pool.HealthCheckURL,
		pool.Quota,
		now,
		pool.ID,
	)
	if err != nil {
		return fmt.Errorf("update pool: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}

	pool.UpdatedAt = now
	return nil
}

func (s *SQLiteStorage) DeletePool(ctx context.Context, id string) error {
	query := `DELETE FROM pools WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete pool: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) SaveStatistics(ctx context.Context, stats *models.Statistics) error {
	query := `
	INSERT INTO statistics (total_active, total_idle, total_waiting, total_max, updated_at)
	VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		stats.TotalActive,
		stats.TotalIdle,
		stats.TotalWaiting,
		stats.TotalMax,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert statistics: %w", err)
	}

	stats.UpdatedAt = now
	return nil
}

func (s *SQLiteStorage) GetLatestStatistics(ctx context.Context) (*models.Statistics, error) {
	query := `
	SELECT total_active, total_idle, total_waiting, total_max, updated_at
	FROM statistics ORDER BY id DESC LIMIT 1
	`

	stats := &models.Statistics{}
	err := s.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalActive,
		&stats.TotalIdle,
		&stats.TotalWaiting,
		&stats.TotalMax,
		&stats.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &models.Statistics{UpdatedAt: time.Now()}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query statistics: %w", err)
	}

	return stats, nil
}
