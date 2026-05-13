package feature

import (
	"database/sql"
	"multitenant/internal/database"
	"time"
)

type FeatureSwitch struct {
	ID         int64     `json:"id"`
	TenantID   string    `json:"tenant_id"`
	FeatureKey string    `json:"feature_key"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Set(tenantID, featureKey string, enabled bool) error {
	now := database.Now()
	var id int64
	err := s.db.QueryRow(`SELECT id FROM feature_switches WHERE tenant_id = ? AND feature_key = ?`, tenantID, featureKey).Scan(&id)
	if err == sql.ErrNoRows {
		_, err := s.db.Exec(`INSERT INTO feature_switches (tenant_id, feature_key, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
			tenantID, featureKey, boolToInt(enabled), now, now)
		return err
	}
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE feature_switches SET enabled = ?, updated_at = ? WHERE id = ?`, boolToInt(enabled), now, id)
	return err
}

func (s *Service) Get(tenantID, featureKey string) (*FeatureSwitch, error) {
	var fs FeatureSwitch
	var enabledInt int
	err := s.db.QueryRow(`SELECT id, tenant_id, feature_key, enabled, created_at, updated_at FROM feature_switches WHERE tenant_id = ? AND feature_key = ?`,
		tenantID, featureKey).Scan(&fs.ID, &fs.TenantID, &fs.FeatureKey, &enabledInt, &fs.CreatedAt, &fs.UpdatedAt)
	if err != nil {
		return nil, err
	}
	fs.Enabled = enabledInt != 0
	return &fs, nil
}

func (s *Service) List(tenantID string) ([]*FeatureSwitch, error) {
	rows, err := s.db.Query(`SELECT id, tenant_id, feature_key, enabled, created_at, updated_at FROM feature_switches WHERE tenant_id = ?`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*FeatureSwitch{}
	for rows.Next() {
		var fs FeatureSwitch
		var enabledInt int
		err := rows.Scan(&fs.ID, &fs.TenantID, &fs.FeatureKey, &enabledInt, &fs.CreatedAt, &fs.UpdatedAt)
		if err != nil {
			return nil, err
		}
		fs.Enabled = enabledInt != 0
		list = append(list, &fs)
	}
	return list, nil
}

func (s *Service) IsEnabled(tenantID, featureKey string, defaultEnabled bool) (bool, error) {
	var enabledInt int
	err := s.db.QueryRow(`SELECT enabled FROM feature_switches WHERE tenant_id = ? AND feature_key = ?`, tenantID, featureKey).Scan(&enabledInt)
	if err == sql.ErrNoRows {
		return defaultEnabled, nil
	}
	if err != nil {
		return false, err
	}
	return enabledInt != 0, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
