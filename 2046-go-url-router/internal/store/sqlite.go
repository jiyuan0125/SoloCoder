package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"router/internal/model"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		method TEXT NOT NULL,
		target TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(path, method)
	);
	CREATE INDEX IF NOT EXISTS idx_routes_path ON routes(path);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) Add(ctx context.Context, path, method, target string) (*model.Route, error) {
	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO routes (path, method, target, created_at) VALUES (?, ?, ?, ?)`,
		path, method, target, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &model.Route{
		ID:        id,
		Path:      path,
		Method:    method,
		Target:    target,
		CreatedAt: now,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM routes WHERE id = ?`, id)
	return err
}

func (s *Store) GetAll(ctx context.Context) ([]*model.Route, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, path, method, target, created_at FROM routes ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []*model.Route
	for rows.Next() {
		r := &model.Route{}
		if err := rows.Scan(&r.ID, &r.Path, &r.Method, &r.Target, &r.CreatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}
	return routes, rows.Err()
}
