package dao

import (
	"coupon-service/model"
	"database/sql"
	"fmt"
	"time"
)

type BatchDAO struct {
	db *sql.DB
}

func NewBatchDAO() *BatchDAO {
	return &BatchDAO{db: DB}
}

func (d *BatchDAO) CreateBatch(batch *model.CouponBatch) (int64, error) {
	query := `
	INSERT INTO coupon_batches (
		name, discount_amount, threshold_amount, total_quantity,
		issued_quantity, redeemed_quantity, valid_start, valid_end,
		limit_per_user, created_at, updated_at
	) VALUES (?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := d.db.Exec(query,
		batch.Name,
		batch.DiscountAmount,
		batch.ThresholdAmount,
		batch.TotalQuantity,
		batch.ValidStart,
		batch.ValidEnd,
		batch.LimitPerUser,
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create batch: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (d *BatchDAO) GetBatchByID(id int64) (*model.CouponBatch, error) {
	query := `
	SELECT id, name, discount_amount, threshold_amount, total_quantity,
		   issued_quantity, redeemed_quantity, valid_start, valid_end,
		   limit_per_user, created_at, updated_at
	FROM coupon_batches
	WHERE id = ?
	`

	batch := &model.CouponBatch{}
	var validStart, validEnd, createdAt, updatedAt string

	err := d.db.QueryRow(query, id).Scan(
		&batch.ID,
		&batch.Name,
		&batch.DiscountAmount,
		&batch.ThresholdAmount,
		&batch.TotalQuantity,
		&batch.IssuedQuantity,
		&batch.RedeemedQuantity,
		&validStart,
		&validEnd,
		&batch.LimitPerUser,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get batch by id: %w", err)
	}

	batch.ValidStart, _ = time.Parse("2006-01-02 15:04:05-07:00", validStart)
	if batch.ValidStart.IsZero() {
		batch.ValidStart, _ = time.Parse("2006-01-02T15:04:05Z", validStart)
	}
	batch.ValidEnd, _ = time.Parse("2006-01-02 15:04:05-07:00", validEnd)
	if batch.ValidEnd.IsZero() {
		batch.ValidEnd, _ = time.Parse("2006-01-02T15:04:05Z", validEnd)
	}
	batch.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt)
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt)
	}
	batch.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt)
	if batch.UpdatedAt.IsZero() {
		batch.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt)
	}

	return batch, nil
}

func (d *BatchDAO) GetBatchByName(name string) (*model.CouponBatch, error) {
	query := `
	SELECT id, name, discount_amount, threshold_amount, total_quantity,
		   issued_quantity, redeemed_quantity, valid_start, valid_end,
		   limit_per_user, created_at, updated_at
	FROM coupon_batches
	WHERE name = ?
	`

	batch := &model.CouponBatch{}
	var validStart, validEnd, createdAt, updatedAt string

	err := d.db.QueryRow(query, name).Scan(
		&batch.ID,
		&batch.Name,
		&batch.DiscountAmount,
		&batch.ThresholdAmount,
		&batch.TotalQuantity,
		&batch.IssuedQuantity,
		&batch.RedeemedQuantity,
		&validStart,
		&validEnd,
		&batch.LimitPerUser,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get batch by name: %w", err)
	}

	batch.ValidStart, _ = time.Parse("2006-01-02 15:04:05-07:00", validStart)
	if batch.ValidStart.IsZero() {
		batch.ValidStart, _ = time.Parse("2006-01-02T15:04:05Z", validStart)
	}
	batch.ValidEnd, _ = time.Parse("2006-01-02 15:04:05-07:00", validEnd)
	if batch.ValidEnd.IsZero() {
		batch.ValidEnd, _ = time.Parse("2006-01-02T15:04:05Z", validEnd)
	}
	batch.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt)
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt)
	}
	batch.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt)
	if batch.UpdatedAt.IsZero() {
		batch.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt)
	}

	return batch, nil
}

func (d *BatchDAO) UpdateBatch(batch *model.CouponBatch) error {
	query := `
	UPDATE coupon_batches
	SET name = ?, discount_amount = ?, threshold_amount = ?,
		valid_start = ?, valid_end = ?, limit_per_user = ?,
		updated_at = ?
	WHERE id = ?
	`

	now := time.Now()
	_, err := d.db.Exec(query,
		batch.Name,
		batch.DiscountAmount,
		batch.ThresholdAmount,
		batch.ValidStart,
		batch.ValidEnd,
		batch.LimitPerUser,
		now,
		batch.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update batch: %w", err)
	}

	return nil
}

func (d *BatchDAO) IncrementIssuedQuantity(batchID int64, tx *sql.Tx) error {
	query := `
	UPDATE coupon_batches
	SET issued_quantity = issued_quantity + 1,
		updated_at = ?
	WHERE id = ? AND issued_quantity < total_quantity
	`

	var err error
	now := time.Now()
	if tx != nil {
		_, err = tx.Exec(query, now, batchID)
	} else {
		_, err = d.db.Exec(query, now, batchID)
	}

	if err != nil {
		return fmt.Errorf("failed to increment issued quantity: %w", err)
	}

	return nil
}

func (d *BatchDAO) IncrementRedeemedQuantity(batchID int64, tx *sql.Tx) error {
	query := `
	UPDATE coupon_batches
	SET redeemed_quantity = redeemed_quantity + 1,
		updated_at = ?
	WHERE id = ?
	`

	var err error
	now := time.Now()
	if tx != nil {
		_, err = tx.Exec(query, now, batchID)
	} else {
		_, err = d.db.Exec(query, now, batchID)
	}

	if err != nil {
		return fmt.Errorf("failed to increment redeemed quantity: %w", err)
	}

	return nil
}

func (d *BatchDAO) DecrementIssuedQuantity(batchID int64, count int, tx *sql.Tx) error {
	query := `
	UPDATE coupon_batches
	SET issued_quantity = issued_quantity - ?,
		updated_at = ?
	WHERE id = ?
	`

	var err error
	now := time.Now()
	if tx != nil {
		_, err = tx.Exec(query, count, now, batchID)
	} else {
		_, err = d.db.Exec(query, count, now, batchID)
	}

	if err != nil {
		return fmt.Errorf("failed to decrement issued quantity: %w", err)
	}

	return nil
}

func (d *BatchDAO) GetBatchProgress(batchID int64) (*model.BatchProgress, error) {
	query := `
	SELECT id, total_quantity, issued_quantity, redeemed_quantity
	FROM coupon_batches
	WHERE id = ?
	`

	progress := &model.BatchProgress{}
	err := d.db.QueryRow(query, batchID).Scan(
		&progress.BatchID,
		&progress.TotalQuantity,
		&progress.IssuedQuantity,
		&progress.RedeemedQuantity,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get batch progress: %w", err)
	}

	progress.RemainingQuantity = progress.TotalQuantity - progress.IssuedQuantity

	return progress, nil
}

func (d *BatchDAO) ListBatches() ([]*model.CouponBatch, error) {
	query := `
	SELECT id, name, discount_amount, threshold_amount, total_quantity,
		   issued_quantity, redeemed_quantity, valid_start, valid_end,
		   limit_per_user, created_at, updated_at
	FROM coupon_batches
	ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list batches: %w", err)
	}
	defer rows.Close()

	var batches []*model.CouponBatch
	for rows.Next() {
		batch := &model.CouponBatch{}
		var validStart, validEnd, createdAt, updatedAt string

		err := rows.Scan(
			&batch.ID,
			&batch.Name,
			&batch.DiscountAmount,
			&batch.ThresholdAmount,
			&batch.TotalQuantity,
			&batch.IssuedQuantity,
			&batch.RedeemedQuantity,
			&validStart,
			&validEnd,
			&batch.LimitPerUser,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch: %w", err)
		}

		batch.ValidStart, _ = time.Parse("2006-01-02 15:04:05-07:00", validStart)
		if batch.ValidStart.IsZero() {
			batch.ValidStart, _ = time.Parse("2006-01-02T15:04:05Z", validStart)
		}
		batch.ValidEnd, _ = time.Parse("2006-01-02 15:04:05-07:00", validEnd)
		if batch.ValidEnd.IsZero() {
			batch.ValidEnd, _ = time.Parse("2006-01-02T15:04:05Z", validEnd)
		}
		batch.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt)
		if batch.CreatedAt.IsZero() {
			batch.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt)
		}
		batch.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt)
		if batch.UpdatedAt.IsZero() {
			batch.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt)
		}

		batches = append(batches, batch)
	}

	return batches, nil
}
