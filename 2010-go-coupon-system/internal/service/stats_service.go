package service

import (
	"database/sql"
	"time"

	"coupon-system/internal/database"
	"coupon-system/internal/model"
)

func UpdateStats(tx *sql.Tx, batchID string) error {
	var totalCount, claimedCount, usedCount int64
	var usageRate float64

	err := tx.QueryRow(`
		SELECT total_count, claimed_count, used_count 
		FROM coupons 
		WHERE batch_id = ?
	`, batchID).Scan(&totalCount, &claimedCount, &usedCount)
	if err != nil {
		return err
	}

	if totalCount > 0 {
		usageRate = float64(claimedCount) / float64(totalCount)
	} else {
		usageRate = 0
	}

	result, err := tx.Exec(`
		UPDATE coupon_stats 
		SET total_count = ?, claimed_count = ?, used_count = ?, usage_rate = ?, updated_at = ?
		WHERE batch_id = ?
	`, totalCount, claimedCount, usedCount, usageRate, time.Now(), batchID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		_, err = tx.Exec(`
			INSERT INTO coupon_stats (batch_id, total_count, claimed_count, used_count, usage_rate, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, batchID, totalCount, claimedCount, usedCount, usageRate, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}

func GetStatsByBatch(batchID string) (*model.CouponStats, error) {
	var stats model.CouponStats
	err := database.DB.QueryRow(`
		SELECT id, batch_id, total_count, claimed_count, used_count, usage_rate, updated_at
		FROM coupon_stats
		WHERE batch_id = ?
	`, batchID).Scan(
		&stats.ID, &stats.BatchID, &stats.TotalCount, &stats.ClaimedCount, 
		&stats.UsedCount, &stats.UsageRate, &stats.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return &model.CouponStats{
			BatchID: batchID,
			TotalCount: 0,
			ClaimedCount: 0,
			UsedCount: 0,
			UsageRate: 0,
			UpdatedAt: time.Now(),
		}, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}

func VerifyCounts(tx *sql.Tx, batchID string) error {
	var dbClaimed, dbUsed int64
	err := tx.QueryRow(`
		SELECT claimed_count, used_count FROM coupons WHERE batch_id = ?
	`, batchID).Scan(&dbClaimed, &dbUsed)
	if err != nil {
		return err
	}

	var actualClaimed int64
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM user_coupons WHERE batch_id = ?
	`, batchID).Scan(&actualClaimed)
	if err != nil {
		return err
	}

	var actualUsed int64
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM user_coupons WHERE batch_id = ? AND status = ?
	`, batchID, model.UserCouponStatusRedeemed).Scan(&actualUsed)
	if err != nil {
		return err
	}

	if dbClaimed != actualClaimed || dbUsed != actualUsed {
		_, err = tx.Exec(`
			UPDATE coupons SET claimed_count = ?, used_count = ? WHERE batch_id = ?
		`, actualClaimed, actualUsed, batchID)
		if err != nil {
			return err
		}
	}

	return nil
}
