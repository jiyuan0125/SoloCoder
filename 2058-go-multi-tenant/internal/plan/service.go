package plan

import (
	"database/sql"
	"encoding/json"
	"multitenant/internal/database"
	"time"
)

type Plan struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	MaxUsers     int      `json:"max_users"`
	MaxStorageMB int      `json:"max_storage_mb"`
	MaxAPICalls  int      `json:"max_api_calls"`
	Features     []string `json:"features"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(p *Plan) error {
	now := database.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	features, err := json.Marshal(p.Features)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO plans (id, name, description, max_users, max_storage_mb, max_api_calls, features, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(stmt, p.ID, p.Name, p.Description, p.MaxUsers, p.MaxStorageMB, p.MaxAPICalls, string(features), p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *Service) Get(id string) (*Plan, error) {
	var featuresStr string
	var p Plan
	stmt := `SELECT id, name, description, max_users, max_storage_mb, max_api_calls, features, created_at, updated_at FROM plans WHERE id = ?`
	err := s.db.QueryRow(stmt, id).Scan(&p.ID, &p.Name, &p.Description, &p.MaxUsers, &p.MaxStorageMB, &p.MaxAPICalls, &featuresStr, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(featuresStr), &p.Features)
	return &p, nil
}

func (s *Service) List() ([]*Plan, error) {
	rows, err := s.db.Query(`SELECT id, name, description, max_users, max_storage_mb, max_api_calls, features, created_at, updated_at FROM plans`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := []*Plan{}
	for rows.Next() {
		var featuresStr string
		var p Plan
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.MaxUsers, &p.MaxStorageMB, &p.MaxAPICalls, &featuresStr, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(featuresStr), &p.Features)
		plans = append(plans, &p)
	}
	return plans, nil
}

func (s *Service) Update(id string, updates map[string]interface{}) (*Plan, error) {
	updates["updated_at"] = database.Now()
	query := "UPDATE plans SET "
	args := []interface{}{}
	first := true
	for k, v := range updates {
		if k == "id" || k == "created_at" {
			continue
		}
		if !first {
			query += ", "
		}
		if k == "features" {
			features, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			query += k + " = ?"
			args = append(args, string(features))
		} else {
			query += k + " = ?"
			args = append(args, v)
		}
		first = false
	}
	query += " WHERE id = ?"
	args = append(args, id)
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *Service) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM plans WHERE id = ?`, id)
	return err
}

func (s *Service) HasFeature(planID, featureKey string) (bool, error) {
	var featuresStr string
	err := s.db.QueryRow(`SELECT features FROM plans WHERE id = ?`, planID).Scan(&featuresStr)
	if err != nil {
		return false, err
	}
	var features []string
	json.Unmarshal([]byte(featuresStr), &features)
	for _, f := range features {
		if f == featureKey {
			return true, nil
		}
	}
	return false, nil
}
