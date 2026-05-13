package usage

import (
	"database/sql"
	"errors"
	"fmt"
	"multitenant/internal/database"
	"time"
)

const (
	ResourceUser      = "users"
	ResourceStorage   = "storage_mb"
	ResourceAPICall   = "api_calls"
	WarningThreshold  = 0.80
)

var (
	ErrQuotaExceeded   = errors.New("quota exceeded")
	ErrQuotaWarning    = errors.New("quota warning: usage exceeds 80%")
)

type UsageRecord struct {
	ID          int64     `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ResourceType string   `json:"resource_type"`
	Amount      int       `json:"amount"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Notification struct {
	ID        int64     `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) getOrCreateUsage(tenantID, resourceType string, periodStart, periodEnd time.Time) (*UsageRecord, error) {
	var ur UsageRecord
	err := s.db.QueryRow(`SELECT id, tenant_id, resource_type, amount, period_start, period_end, created_at, updated_at FROM usage_stats WHERE tenant_id = ? AND resource_type = ? AND period_start = ?`,
		tenantID, resourceType, periodStart).Scan(&ur.ID, &ur.TenantID, &ur.ResourceType, &ur.Amount, &ur.PeriodStart, &ur.PeriodEnd, &ur.CreatedAt, &ur.UpdatedAt)
	if err == nil {
		return &ur, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	now := database.Now()
	res, err := s.db.Exec(`INSERT INTO usage_stats (tenant_id, resource_type, amount, period_start, period_end, created_at, updated_at) VALUES (?, ?, 0, ?, ?, ?, ?)`,
		tenantID, resourceType, periodStart, periodEnd, now, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &UsageRecord{
		ID:          id,
		TenantID:    tenantID,
		ResourceType: resourceType,
		Amount:      0,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Service) Add(tenantID, resourceType string, amount int, maxLimit int, periodStart, periodEnd time.Time) (int, int, error) {
	now := database.Now()

	var urID int64
	var currentAmount int
	err := s.db.QueryRow(`SELECT id, amount FROM usage_stats WHERE tenant_id = ? AND resource_type = ? AND period_start = ?`,
		tenantID, resourceType, periodStart).Scan(&urID, &currentAmount)
	if err == sql.ErrNoRows {
		res, err := s.db.Exec(`INSERT INTO usage_stats (tenant_id, resource_type, amount, period_start, period_end, created_at, updated_at) VALUES (?, ?, 0, ?, ?, ?, ?)`,
			tenantID, resourceType, periodStart, periodEnd, now, now)
		if err != nil {
			return 0, 0, err
		}
		urID, _ = res.LastInsertId()
		currentAmount = 0
	} else if err != nil {
		return 0, 0, err
	}

	newAmount := currentAmount + amount
	if maxLimit > 0 && newAmount > maxLimit {
		return currentAmount, maxLimit, ErrQuotaExceeded
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE usage_stats SET amount = ?, updated_at = ? WHERE id = ?`, newAmount, now, urID)
	if err != nil {
		return 0, 0, err
	}

	if maxLimit > 0 {
		ratio := float64(newAmount) / float64(maxLimit)
		if ratio > WarningThreshold {
			msg := fmt.Sprintf("资源用量警告: %s 已使用 %d, 限额 %d", resourceType, newAmount, maxLimit)
			_, err = tx.Exec(`INSERT INTO notifications (tenant_id, message, level, read, created_at) VALUES (?, ?, ?, 0, ?)`, tenantID, msg, "warning", now)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	remaining := maxLimit - newAmount
	if maxLimit <= 0 {
		remaining = -1
	}
	return newAmount, remaining, nil
}

func (s *Service) Get(tenantID, resourceType string, periodStart time.Time) (*UsageRecord, error) {
	var ur UsageRecord
	err := s.db.QueryRow(`SELECT id, tenant_id, resource_type, amount, period_start, period_end, created_at, updated_at FROM usage_stats WHERE tenant_id = ? AND resource_type = ? AND period_start = ?`,
		tenantID, resourceType, periodStart).Scan(&ur.ID, &ur.TenantID, &ur.ResourceType, &ur.Amount, &ur.PeriodStart, &ur.PeriodEnd, &ur.CreatedAt, &ur.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ur, nil
}

func (s *Service) List(tenantID string) ([]*UsageRecord, error) {
	rows, err := s.db.Query(`SELECT id, tenant_id, resource_type, amount, period_start, period_end, created_at, updated_at FROM usage_stats WHERE tenant_id = ?`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*UsageRecord{}
	for rows.Next() {
		var ur UsageRecord
		err := rows.Scan(&ur.ID, &ur.TenantID, &ur.ResourceType, &ur.Amount, &ur.PeriodStart, &ur.PeriodEnd, &ur.CreatedAt, &ur.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &ur)
	}
	return list, nil
}

func (s *Service) ResetCycle(tenantID, resourceType string, newPeriodStart, newPeriodEnd time.Time) error {
	_, err := s.db.Exec(`UPDATE usage_stats SET amount = 0, period_start = ?, period_end = ?, updated_at = ? WHERE tenant_id = ? AND resource_type = ?`,
		newPeriodStart, newPeriodEnd, database.Now(), tenantID, resourceType)
	return err
}

func (s *Service) addNotification(tx *sql.Tx, tenantID, message, level string) error {
	now := database.Now()
	_, err := tx.Exec(`INSERT INTO notifications (tenant_id, message, level, read, created_at) VALUES (?, ?, ?, 0, ?)`, tenantID, message, level, now)
	return err
}

func (s *Service) ListNotifications(tenantID string, unreadOnly bool) ([]*Notification, error) {
	query := `SELECT id, tenant_id, message, level, read, created_at FROM notifications WHERE tenant_id = ?`
	args := []interface{}{tenantID}
	if unreadOnly {
		query += ` AND read = 0`
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*Notification{}
	for rows.Next() {
		var n Notification
		var readInt int
		err := rows.Scan(&n.ID, &n.TenantID, &n.Message, &n.Level, &readInt, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		n.Read = readInt != 0
		list = append(list, &n)
	}
	return list, nil
}

func (s *Service) MarkNotificationRead(id int64) error {
	_, err := s.db.Exec(`UPDATE notifications SET read = 1 WHERE id = ?`, id)
	return err
}
