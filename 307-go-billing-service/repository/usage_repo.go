package repository

import (
	"billing-service/model"
	"database/sql"
	"time"
)

type UsageRepository struct {
	db *Database
}

func NewUsageRepository(db *Database) *UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) GetOrCreate(customerID uint, year int, month int) (*model.Usage, error) {
	query := `
		SELECT id, customer_id, year, month, sms_used, storage_used, created_at, updated_at
		FROM usages
		WHERE customer_id = ? AND year = ? AND month = ?
	`

	var usage model.Usage
	err := r.db.QueryRow(query, customerID, year, month).Scan(
		&usage.ID,
		&usage.CustomerID,
		&usage.Year,
		&usage.Month,
		&usage.SmsUsed,
		&usage.StorageUsed,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		now := time.Now()
		usage = model.Usage{
			CustomerID:  customerID,
			Year:        year,
			Month:       month,
			SmsUsed:     0,
			StorageUsed: 0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		result, err := r.db.Exec(`
			INSERT INTO usages (customer_id, year, month, sms_used, storage_used, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, usage.CustomerID, usage.Year, usage.Month, usage.SmsUsed, usage.StorageUsed, usage.CreatedAt, usage.UpdatedAt)
		if err != nil {
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		usage.ID = uint(id)

		return &usage, nil
	}

	if err != nil {
		return nil, err
	}

	return &usage, nil
}

func (r *UsageRepository) UpdateSmsUsage(customerID uint, year int, month int, smsCount int) error {
	query := `
		UPDATE usages
		SET sms_used = sms_used + ?, updated_at = ?
		WHERE customer_id = ? AND year = ? AND month = ?
	`

	_, err := r.db.Exec(query, smsCount, time.Now(), customerID, year, month)
	return err
}

func (r *UsageRepository) UpdateStorageUsage(customerID uint, year int, month int, storageGB float64) error {
	query := `
		UPDATE usages
		SET storage_used = ?, updated_at = ?
		WHERE customer_id = ? AND year = ? AND month = ?
	`

	_, err := r.db.Exec(query, storageGB, time.Now(), customerID, year, month)
	return err
}

func (r *UsageRepository) GetByCustomerAndMonth(customerID uint, year int, month int) (*model.Usage, error) {
	query := `
		SELECT id, customer_id, year, month, sms_used, storage_used, created_at, updated_at
		FROM usages
		WHERE customer_id = ? AND year = ? AND month = ?
	`

	var usage model.Usage
	err := r.db.QueryRow(query, customerID, year, month).Scan(
		&usage.ID,
		&usage.CustomerID,
		&usage.Year,
		&usage.Month,
		&usage.SmsUsed,
		&usage.StorageUsed,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &usage, nil
}
