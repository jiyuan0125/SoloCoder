package dao

import (
	"coupon-service/model"
	"database/sql"
	"fmt"
	"time"
)

type CouponDAO struct {
	db *sql.DB
}

func NewCouponDAO() *CouponDAO {
	return &CouponDAO{db: DB}
}

func (d *CouponDAO) CreateCoupon(coupon *model.Coupon, tx *sql.Tx) (int64, error) {
	query := `
	INSERT INTO coupons (batch_id, user_id, code, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	var err error
	var result sql.Result

	if tx != nil {
		result, err = tx.Exec(query,
			coupon.BatchID,
			coupon.UserID,
			coupon.Code,
			coupon.Status,
			now,
			now,
		)
	} else {
		result, err = d.db.Exec(query,
			coupon.BatchID,
			coupon.UserID,
			coupon.Code,
			coupon.Status,
			now,
			now,
		)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to create coupon: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (d *CouponDAO) GetCouponByCode(code string) (*model.Coupon, error) {
	query := `
	SELECT id, batch_id, user_id, code, status, redeemed_at, created_at, updated_at
	FROM coupons
	WHERE code = ?
	`

	coupon := &model.Coupon{}
	var redeemedAt, createdAt, updatedAt sql.NullString

	err := d.db.QueryRow(query, code).Scan(
		&coupon.ID,
		&coupon.BatchID,
		&coupon.UserID,
		&coupon.Code,
		&coupon.Status,
		&redeemedAt,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon by code: %w", err)
	}

	if redeemedAt.Valid {
		coupon.RedeemedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", redeemedAt.String)
		if coupon.RedeemedAt.IsZero() {
			coupon.RedeemedAt, _ = time.Parse("2006-01-02T15:04:05Z", redeemedAt.String)
		}
	}
	if createdAt.Valid {
		coupon.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt.String)
		if coupon.CreatedAt.IsZero() {
			coupon.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt.String)
		}
	}
	if updatedAt.Valid {
		coupon.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt.String)
		if coupon.UpdatedAt.IsZero() {
			coupon.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt.String)
		}
	}

	return coupon, nil
}

func (d *CouponDAO) GetCouponByID(id int64) (*model.Coupon, error) {
	query := `
	SELECT id, batch_id, user_id, code, status, redeemed_at, created_at, updated_at
	FROM coupons
	WHERE id = ?
	`

	coupon := &model.Coupon{}
	var redeemedAt, createdAt, updatedAt sql.NullString

	err := d.db.QueryRow(query, id).Scan(
		&coupon.ID,
		&coupon.BatchID,
		&coupon.UserID,
		&coupon.Code,
		&coupon.Status,
		&redeemedAt,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon by id: %w", err)
	}

	if redeemedAt.Valid {
		coupon.RedeemedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", redeemedAt.String)
		if coupon.RedeemedAt.IsZero() {
			coupon.RedeemedAt, _ = time.Parse("2006-01-02T15:04:05Z", redeemedAt.String)
		}
	}
	if createdAt.Valid {
		coupon.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt.String)
		if coupon.CreatedAt.IsZero() {
			coupon.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt.String)
		}
	}
	if updatedAt.Valid {
		coupon.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt.String)
		if coupon.UpdatedAt.IsZero() {
			coupon.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt.String)
		}
	}

	return coupon, nil
}

func (d *CouponDAO) CountUserCouponsByBatch(userID string, batchID int64) (int, error) {
	query := `
	SELECT COUNT(*) FROM coupons
	WHERE user_id = ? AND batch_id = ? AND status IN ('issued', 'redeemed')
	`

	var count int
	err := d.db.QueryRow(query, userID, batchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user coupons: %w", err)
	}

	return count, nil
}

func (d *CouponDAO) RedeemCoupon(couponID int64, tx *sql.Tx) error {
	query := `
	UPDATE coupons
	SET status = ?, redeemed_at = ?, updated_at = ?
	WHERE id = ? AND status = ?
	`

	now := time.Now()
	var err error

	if tx != nil {
		_, err = tx.Exec(query,
			model.CouponStatusRedeemed,
			now,
			now,
			couponID,
			model.CouponStatusIssued,
		)
	} else {
		_, err = d.db.Exec(query,
			model.CouponStatusRedeemed,
			now,
			now,
			couponID,
			model.CouponStatusIssued,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to redeem coupon: %w", err)
	}

	return nil
}

func (d *CouponDAO) BeginTx() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *CouponDAO) GetUserAvailableCoupons(userID string) ([]*model.Coupon, error) {
	query := `
	SELECT c.id, c.batch_id, c.user_id, c.code, c.status, c.redeemed_at, c.created_at, c.updated_at
	FROM coupons c
	JOIN coupon_batches b ON c.batch_id = b.id
	WHERE c.user_id = ? AND c.status = ?
	  AND datetime('now') >= datetime(b.valid_start)
	  AND datetime('now') <= datetime(b.valid_end, '+23 hours', '+59 minutes', '+59 seconds')
	ORDER BY b.valid_end ASC
	`

	rows, err := d.db.Query(query, userID, model.CouponStatusIssued)
	if err != nil {
		return nil, fmt.Errorf("failed to get user available coupons: %w", err)
	}
	defer rows.Close()

	var coupons []*model.Coupon
	for rows.Next() {
		coupon := &model.Coupon{}
		var redeemedAt, createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&coupon.ID,
			&coupon.BatchID,
			&coupon.UserID,
			&coupon.Code,
			&coupon.Status,
			&redeemedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coupon: %w", err)
		}

		if redeemedAt.Valid {
			coupon.RedeemedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", redeemedAt.String)
			if coupon.RedeemedAt.IsZero() {
				coupon.RedeemedAt, _ = time.Parse("2006-01-02T15:04:05Z", redeemedAt.String)
			}
		}
		if createdAt.Valid {
			coupon.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt.String)
			if coupon.CreatedAt.IsZero() {
				coupon.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt.String)
			}
		}
		if updatedAt.Valid {
			coupon.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt.String)
			if coupon.UpdatedAt.IsZero() {
				coupon.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt.String)
			}
		}

		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

func (d *CouponDAO) GetUserIssuedCouponsByBatch(userID string, batchID int64) ([]*model.Coupon, error) {
	query := `
	SELECT id, batch_id, user_id, code, status, redeemed_at, created_at, updated_at
	FROM coupons
	WHERE user_id = ? AND batch_id = ? AND status = ?
	`

	rows, err := d.db.Query(query, userID, batchID, model.CouponStatusIssued)
	if err != nil {
		return nil, fmt.Errorf("failed to get user issued coupons: %w", err)
	}
	defer rows.Close()

	var coupons []*model.Coupon
	for rows.Next() {
		coupon := &model.Coupon{}
		var redeemedAt, createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&coupon.ID,
			&coupon.BatchID,
			&coupon.UserID,
			&coupon.Code,
			&coupon.Status,
			&redeemedAt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coupon: %w", err)
		}

		if redeemedAt.Valid {
			coupon.RedeemedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", redeemedAt.String)
			if coupon.RedeemedAt.IsZero() {
				coupon.RedeemedAt, _ = time.Parse("2006-01-02T15:04:05Z", redeemedAt.String)
			}
		}
		if createdAt.Valid {
			coupon.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAt.String)
			if coupon.CreatedAt.IsZero() {
				coupon.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt.String)
			}
		}
		if updatedAt.Valid {
			coupon.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", updatedAt.String)
			if coupon.UpdatedAt.IsZero() {
				coupon.UpdatedAt, _ = time.Parse("2006-01-02T15:04:05Z", updatedAt.String)
			}
		}

		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

func (d *CouponDAO) ReturnCoupon(couponID int64, tx *sql.Tx) error {
	query := `
	UPDATE coupons
	SET status = ?, updated_at = ?
	WHERE id = ? AND status = ?
	`

	now := time.Now()
	var err error

	if tx != nil {
		_, err = tx.Exec(query,
			model.CouponStatusReturned,
			now,
			couponID,
			model.CouponStatusIssued,
		)
	} else {
		_, err = d.db.Exec(query,
			model.CouponStatusReturned,
			now,
			couponID,
			model.CouponStatusIssued,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to return coupon: %w", err)
	}

	return nil
}
