package tenant

import (
	"database/sql"
	"multitenant/internal/database"
	"time"
)

type Tenant struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	PlanID        string    `json:"plan_id"`
	PendingPlanID string    `json:"pending_plan_id"`
	CycleStartAt  time.Time `json:"cycle_start_at"`
	CycleEndAt    time.Time `json:"cycle_end_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(t *Tenant) error {
	now := database.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	if t.CycleStartAt.IsZero() {
		t.CycleStartAt = now
	}
	if t.CycleEndAt.IsZero() {
		t.CycleEndAt = t.CycleStartAt.AddDate(0, 1, 0)
	}
	stmt := `INSERT INTO tenants (id, name, plan_id, pending_plan_id, cycle_start_at, cycle_end_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(stmt, t.ID, t.Name, t.PlanID, t.PendingPlanID, t.CycleStartAt, t.CycleEndAt, t.CreatedAt, t.UpdatedAt)
	return err
}

func (s *Service) Get(id string) (*Tenant, error) {
	var t Tenant
	var pendingPlanID sql.NullString
	stmt := `SELECT id, name, plan_id, pending_plan_id, cycle_start_at, cycle_end_at, created_at, updated_at FROM tenants WHERE id = ?`
	err := s.db.QueryRow(stmt, id).Scan(&t.ID, &t.Name, &t.PlanID, &pendingPlanID, &t.CycleStartAt, &t.CycleEndAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if pendingPlanID.Valid {
		t.PendingPlanID = pendingPlanID.String
	}
	return &t, nil
}

func (s *Service) List() ([]*Tenant, error) {
	rows, err := s.db.Query(`SELECT id, name, plan_id, pending_plan_id, cycle_start_at, cycle_end_at, created_at, updated_at FROM tenants`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tenants := []*Tenant{}
	for rows.Next() {
		var t Tenant
		var pendingPlanID sql.NullString
		err := rows.Scan(&t.ID, &t.Name, &t.PlanID, &pendingPlanID, &t.CycleStartAt, &t.CycleEndAt, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if pendingPlanID.Valid {
			t.PendingPlanID = pendingPlanID.String
		}
		tenants = append(tenants, &t)
	}
	return tenants, nil
}

func (s *Service) Update(id string, updates map[string]interface{}) (*Tenant, error) {
	updates["updated_at"] = database.Now()
	query := "UPDATE tenants SET "
	args := []interface{}{}
	first := true
	for k, v := range updates {
		if k == "id" || k == "created_at" {
			continue
		}
		if !first {
			query += ", "
		}
		query += k + " = ?"
		args = append(args, v)
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
	_, err := s.db.Exec(`DELETE FROM tenants WHERE id = ?`, id)
	return err
}

func (s *Service) ChangePlan(tenantID, newPlanID string) error {
	now := database.Now()
	query := `UPDATE tenants SET pending_plan_id = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, newPlanID, now, tenantID)
	return err
}

func (s *Service) ApplyPendingPlans() (int, error) {
	now := database.Now()
	rows, err := s.db.Query(`SELECT id, pending_plan_id, cycle_end_at FROM tenants WHERE pending_plan_id IS NOT NULL`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	for rows.Next() {
		var tenantID, pendingPlanID string
		var cycleEndAt time.Time
		if err := rows.Scan(&tenantID, &pendingPlanID, &cycleEndAt); err != nil {
			return 0, err
		}
		if now.After(cycleEndAt) {
			newCycleStart := cycleEndAt
			newCycleEnd := newCycleStart.AddDate(0, 1, 0)
			_, err := tx.Exec(`UPDATE tenants SET plan_id = ?, pending_plan_id = NULL, cycle_start_at = ?, cycle_end_at = ?, updated_at = ? WHERE id = ?`,
				pendingPlanID, newCycleStart, newCycleEnd, now, tenantID)
			if err != nil {
				return 0, err
			}
			count++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Service) Exists(id string) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tenants WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}
