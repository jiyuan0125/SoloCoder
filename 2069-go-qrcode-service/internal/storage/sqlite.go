package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"qrcode-service/internal/types"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	s := &DB{db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	
	return s, nil
}

func (db *DB) initSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	schema := `
	CREATE TABLE IF NOT EXISTS resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_id INTEGER,
		op_type TEXT NOT NULL,
		format TEXT,
		count INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (resource_id) REFERENCES resources(id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_ops_resource ON operations(resource_id);
	CREATE INDEX IF NOT EXISTS idx_ops_time ON operations(created_at);
	`
	
	_, err := db.ExecContext(ctx, schema)
	return err
}

func (db *DB) CreateResource(res *types.Resource) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	result, err := db.ExecContext(ctx,
		`INSERT INTO resources(type, name, description) VALUES(?, ?, ?)`,
		res.Type, res.Name, res.Description,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetResource(id int64) (*types.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	var res types.Resource
	err := db.QueryRowContext(ctx,
		`SELECT id, type, name, description, created_at FROM resources WHERE id = ?`, id,
	).Scan(&res.ID, &res.Type, &res.Name, &res.Description, &res.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &res, err
}

func (db *DB) ListResources() ([]types.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	rows, err := db.QueryContext(ctx,
		`SELECT id, type, name, description, created_at FROM resources ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var resources []types.Resource
	for rows.Next() {
		var res types.Resource
		if err := rows.Scan(&res.ID, &res.Type, &res.Name, &res.Description, &res.CreatedAt); err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	
	return resources, rows.Err()
}

func (db *DB) LogOperation(resourceID int64, opType, format string, count int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if resourceID <= 0 {
		_, err := db.ExecContext(ctx,
			`INSERT INTO operations(op_type, format, count) VALUES(?, ?, ?)`,
			opType, format, count,
		)
		return err
	}
	
	_, err := db.ExecContext(ctx,
		`INSERT INTO operations(resource_id, op_type, format, count) VALUES(?, ?, ?, ?)`,
		resourceID, opType, format, count,
	)
	return err
}

func (db *DB) GetResourceSummary(resourceID int64) (*types.ResourceSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	var summary types.ResourceSummary
	query := `
		SELECT 
			r.id, r.name, r.type,
			COUNT(o.id) as total_ops,
			MAX(o.created_at) as last_op_time
		FROM resources r
		LEFT JOIN operations o ON r.id = o.resource_id
		WHERE r.id = ?
		GROUP BY r.id
	`
	
	err := db.QueryRowContext(ctx, query, resourceID).Scan(
		&summary.ResourceID, &summary.ResourceName, &summary.ResourceType,
		&summary.TotalOps, &summary.LastOpTime,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("resource not found")
	}
	return &summary, err
}

func (db *DB) ListAllSummaries() ([]types.ResourceSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	query := `
		SELECT 
			r.id, r.name, r.type,
			COUNT(o.id) as total_ops,
			IFNULL(MAX(o.created_at), '') as last_op_time
		FROM resources r
		LEFT JOIN operations o ON r.id = o.resource_id
		GROUP BY r.id
		ORDER BY r.id DESC
	`
	
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var summaries []types.ResourceSummary
	for rows.Next() {
		var s types.ResourceSummary
		if err := rows.Scan(&s.ResourceID, &s.ResourceName, &s.ResourceType, &s.TotalOps, &s.LastOpTime); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	
	return summaries, rows.Err()
}

func (db *DB) GetOperations(resourceID int64, limit int) ([]types.Operation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if limit <= 0 {
		limit = 100
	}
	
	rows, err := db.QueryContext(ctx,
		`SELECT o.id, o.resource_id, o.op_type, o.format, o.count, o.created_at, r.name 
		 FROM operations o 
		 LEFT JOIN resources r ON o.resource_id = r.id 
		 WHERE o.resource_id = ? 
		 ORDER BY o.id DESC 
		 LIMIT ?`, resourceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ops []types.Operation
	for rows.Next() {
		var op types.Operation
		var name sql.NullString
		if err := rows.Scan(&op.ID, &op.ResourceID, &op.OpType, &op.Format, &op.Count, &op.CreatedAt, &name); err != nil {
			return nil, err
		}
		if name.Valid {
			op.ResourceName = name.String
		}
		ops = append(ops, op)
	}
	
	return ops, rows.Err()
}
