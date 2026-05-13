package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"serviceregistry/internal/model"
)

type SQLiteStorage struct {
	db   *sql.DB
	lock sync.RWMutex
	path string
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	s := &SQLiteStorage{path: path}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	_, err = db.Exec(`PRAGMA journal_mode=WAL`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	s.db = db

	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) initSchema() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	schema := `
	CREATE TABLE IF NOT EXISTS instances (
		id TEXT PRIMARY KEY,
		service_name TEXT NOT NULL,
		address TEXT NOT NULL,
		port INTEGER NOT NULL,
		version TEXT NOT NULL,
		weight INTEGER NOT NULL DEFAULT 1,
		environment TEXT NOT NULL DEFAULT '',
		metadata TEXT,
		status TEXT NOT NULL,
		last_heartbeat INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		UNIQUE(service_name, address, port)
	);

	CREATE INDEX IF NOT EXISTS idx_instances_service ON instances(service_name);

	CREATE TABLE IF NOT EXISTS resources (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		UNIQUE(type, name)
	);

	CREATE TABLE IF NOT EXISTS resource_associations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		target_type TEXT NOT NULL,
		operation TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_ra_resource ON resource_associations(resource_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStorage) SaveInstance(inst *model.ServiceInstance) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.saveInstanceLocked(inst)
}

func (s *SQLiteStorage) saveInstanceLocked(inst *model.ServiceInstance) error {
	metadata, err := json.Marshal(inst.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
	INSERT OR REPLACE INTO instances 
	(id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		inst.ID, inst.ServiceName, inst.Address, inst.Port,
		inst.Version, inst.Weight, inst.Environment, string(metadata),
		string(inst.Status), inst.LastHeartbeat.Unix(), inst.CreatedAt.Unix(),
	)
	return err
}

func (s *SQLiteStorage) SaveInstanceTx(tx *sql.Tx, inst *model.ServiceInstance) error {
	metadata, err := json.Marshal(inst.Metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	INSERT OR REPLACE INTO instances 
	(id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		inst.ID, inst.ServiceName, inst.Address, inst.Port,
		inst.Version, inst.Weight, inst.Environment, string(metadata),
		string(inst.Status), inst.LastHeartbeat.Unix(), inst.CreatedAt.Unix(),
	)
	return err
}

func (s *SQLiteStorage) DeleteInstance(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.deleteInstanceLocked(id)
}

func (s *SQLiteStorage) deleteInstanceLocked(id string) error {
	_, err := s.db.Exec(`DELETE FROM instances WHERE id = ?`, id)
	return err
}

func (s *SQLiteStorage) DeleteInstanceTx(tx *sql.Tx, id string) error {
	_, err := tx.Exec(`DELETE FROM instances WHERE id = ?`, id)
	return err
}

func (s *SQLiteStorage) UpdateInstanceStatus(id string, status model.InstanceStatus) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, err := s.db.Exec(`UPDATE instances SET status = ? WHERE id = ?`, string(status), id)
	return err
}

func (s *SQLiteStorage) UpdateHeartbeat(id string, t time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, err := s.db.Exec(`UPDATE instances SET last_heartbeat = ? WHERE id = ?`, t.Unix(), id)
	return err
}

func (s *SQLiteStorage) GetInstance(id string) (*model.ServiceInstance, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.getInstanceLocked(id)
}

func (s *SQLiteStorage) getInstanceLocked(id string) (*model.ServiceInstance, error) {
	row := s.db.QueryRow(`
	SELECT id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at
	FROM instances WHERE id = ?
	`, id)
	return s.scanInstance(row)
}

func (s *SQLiteStorage) GetInstanceTx(tx *sql.Tx, id string) (*model.ServiceInstance, error) {
	row := tx.QueryRow(`
	SELECT id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at
	FROM instances WHERE id = ?
	`, id)
	return s.scanInstance(row)
}

func (s *SQLiteStorage) GetInstancesByService(serviceName string) ([]*model.ServiceInstance, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rows, err := s.db.Query(`
	SELECT id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at
	FROM instances WHERE service_name = ?
	`, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanInstances(rows)
}

func (s *SQLiteStorage) GetAllInstances() ([]*model.ServiceInstance, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rows, err := s.db.Query(`
	SELECT id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at
	FROM instances
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanInstances(rows)
}

func (s *SQLiteStorage) GetInstancesNeedingHeartbeatCheck(now time.Time, unhealthyThreshold time.Duration) ([]*model.ServiceInstance, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	deadline := now.Add(-unhealthyThreshold).Unix()

	rows, err := s.db.Query(`
	SELECT id, service_name, address, port, version, weight, environment, metadata, status, last_heartbeat, created_at
	FROM instances WHERE last_heartbeat < ?
	`, deadline)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanInstances(rows)
}

func (s *SQLiteStorage) scanInstances(rows *sql.Rows) ([]*model.ServiceInstance, error) {
	var instances []*model.ServiceInstance
	for rows.Next() {
		inst, err := s.scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

func (s *SQLiteStorage) scanInstance(row interface {
	Scan(dest ...interface{}) error
}) (*model.ServiceInstance, error) {
	var (
		metadataJSON  string
		lastHeartbeat int64
		createdAt     int64
		inst          = &model.ServiceInstance{Metadata: make(map[string]string)}
	)

	err := row.Scan(
		&inst.ID, &inst.ServiceName, &inst.Address, &inst.Port,
		&inst.Version, &inst.Weight, &inst.Environment, &metadataJSON,
		&inst.Status, &lastHeartbeat, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &inst.Metadata); err != nil {
			return nil, err
		}
	}

	inst.LastHeartbeat = time.Unix(lastHeartbeat, 0)
	inst.CreatedAt = time.Unix(createdAt, 0)

	return inst, nil
}

func (s *SQLiteStorage) SaveResource(r *model.Resource) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.saveResourceLocked(r)
}

func (s *SQLiteStorage) saveResourceLocked(r *model.Resource) error {
	_, err := s.db.Exec(`
	INSERT OR REPLACE INTO resources (id, type, name) VALUES (?, ?, ?)
	`, r.ID, string(r.Type), r.Name)
	return err
}

func (s *SQLiteStorage) SaveResourceTx(tx *sql.Tx, r *model.Resource) error {
	_, err := tx.Exec(`
	INSERT OR REPLACE INTO resources (id, type, name) VALUES (?, ?, ?)
	`, r.ID, string(r.Type), r.Name)
	return err
}

func (s *SQLiteStorage) GetResource(id string) (*model.Resource, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	row := s.db.QueryRow(`SELECT id, type, name FROM resources WHERE id = ?`, id)
	var r model.Resource
	var t string
	err := row.Scan(&r.ID, &t, &r.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	r.Type = model.ResourceType(t)
	return &r, err
}

func (s *SQLiteStorage) AddAssociation(assoc *model.ResourceAssociation) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.addAssociationLocked(assoc)
}

func (s *SQLiteStorage) addAssociationLocked(assoc *model.ResourceAssociation) error {
	_, err := s.db.Exec(`
	INSERT INTO resource_associations 
	(resource_id, target_id, target_type, operation, timestamp)
	VALUES (?, ?, ?, ?, ?)
	`,
		assoc.ResourceID, assoc.TargetID, string(assoc.TargetType),
		assoc.Operation, assoc.Timestamp.Unix(),
	)
	return err
}

func (s *SQLiteStorage) AddAssociationTx(tx *sql.Tx, assoc *model.ResourceAssociation) error {
	_, err := tx.Exec(`
	INSERT INTO resource_associations 
	(resource_id, target_id, target_type, operation, timestamp)
	VALUES (?, ?, ?, ?, ?)
	`,
		assoc.ResourceID, assoc.TargetID, string(assoc.TargetType),
		assoc.Operation, assoc.Timestamp.Unix(),
	)
	return err
}

func (s *SQLiteStorage) GetResourceSummary(resourceID string) ([]*model.ResourceAssociation, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rows, err := s.db.Query(`
	SELECT resource_id, target_id, target_type, operation, timestamp
	FROM resource_associations WHERE resource_id = ?
	ORDER BY timestamp DESC
	`, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var associations []*model.ResourceAssociation
	for rows.Next() {
		var (
			a          = &model.ResourceAssociation{}
			targetType string
			ts         int64
		)
		if err := rows.Scan(&a.ResourceID, &a.TargetID, &targetType, &a.Operation, &ts); err != nil {
			return nil, err
		}
		a.TargetType = model.ResourceType(targetType)
		a.Timestamp = time.Unix(ts, 0)
		associations = append(associations, a)
	}

	return associations, rows.Err()
}

func (s *SQLiteStorage) Transactional(fn func(tx *sql.Tx) error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}
