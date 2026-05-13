package repository

import (
	"database/sql"

	"flashsale/internal/model"
)

type CacheRepository struct {
	db *sql.DB
}

func NewCacheRepository(db *sql.DB) *CacheRepository {
	return &CacheRepository{db: db}
}

func (r *CacheRepository) CreateNotification(notification *model.CacheNotification) error {
	_, err := r.db.Exec(
		`INSERT INTO cache_notifications (module_name, event_type, activity_id, payload)
		VALUES (?, ?, ?, ?)`,
		notification.ModuleName, notification.EventType,
		notification.ActivityID, notification.Payload,
	)
	return err
}

func (r *CacheRepository) CreateQuotaAllocation(quota *model.QuotaAllocation) error {
	_, err := r.db.Exec(
		`INSERT INTO quota_allocations (activity_id, segment_name, total_quota, allocated_quota, ratio)
		VALUES (?, ?, ?, ?, ?)`,
		quota.ActivityID, quota.SegmentName, quota.TotalQuota, quota.AllocatedQuota, quota.Ratio,
	)
	return err
}

func (r *CacheRepository) UpdateQuotaAllocation(quota *model.QuotaAllocation) error {
	_, err := r.db.Exec(
		`UPDATE quota_allocations SET total_quota = ?, allocated_quota = ?, ratio = ?
		WHERE activity_id = ? AND segment_name = ?`,
		quota.TotalQuota, quota.AllocatedQuota, quota.Ratio,
		quota.ActivityID, quota.SegmentName,
	)
	return err
}

func (r *CacheRepository) GetQuotaAllocations(activityID int64) ([]*model.QuotaAllocation, error) {
	rows, err := r.db.Query(
		`SELECT id, activity_id, segment_name, total_quota, allocated_quota, ratio
		FROM quota_allocations WHERE activity_id = ?`, activityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotas := make([]*model.QuotaAllocation, 0)
	for rows.Next() {
		quota := &model.QuotaAllocation{}
		err := rows.Scan(
			&quota.ID, &quota.ActivityID, &quota.SegmentName,
			&quota.TotalQuota, &quota.AllocatedQuota, &quota.Ratio,
		)
		if err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, nil
}

func (r *CacheRepository) DeleteQuotaAllocations(activityID int64) error {
	_, err := r.db.Exec(
		`DELETE FROM quota_allocations WHERE activity_id = ?`, activityID,
	)
	return err
}
