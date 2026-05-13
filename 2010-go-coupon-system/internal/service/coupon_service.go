package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"coupon-system/internal/database"
	"coupon-system/internal/model"
)

type CreateCouponRequest struct {
	BatchID      string    `json:"batch_id"`
	Denomination int64     `json:"denomination"`
	Threshold    int64     `json:"threshold"`
	TotalCount   int64     `json:"total_count"`
	LimitPerUser int64     `json:"limit_per_user"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

type ClaimCouponRequest struct {
	UserID  string `json:"user_id"`
	BatchID string `json:"batch_id"`
}

type RedeemCouponRequest struct {
	UserCouponID int64  `json:"user_coupon_id"`
	OrderID      string `json:"order_id"`
	OrderAmount  int64  `json:"order_amount"`
}

type RefundRequest struct {
	OrderID     string `json:"order_id"`
	RefundAmount int64  `json:"refund_amount"`
}

func CreateCoupon(req *CreateCouponRequest) (*model.Coupon, error) {
	if req.Denomination <= 0 {
		return nil, errors.New("面额必须为正数")
	}
	if req.Threshold < req.Denomination {
		return nil, errors.New("使用门槛不能低于面额")
	}
	if req.TotalCount <= 0 {
		return nil, errors.New("总发行量必须为正数")
	}
	if req.LimitPerUser <= 0 {
		return nil, errors.New("每人限领数量必须为正数")
	}
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}
	if req.EndTime.Before(time.Now()) {
		return nil, errors.New("结束时间不能早于当前时间")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRow(`SELECT id FROM coupons WHERE batch_id = ?`, req.BatchID).Scan(&existingID)
	if err == nil {
		return nil, errors.New("批次ID已存在")
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	result, err := tx.Exec(`
		INSERT INTO coupons (batch_id, denomination, threshold, total_count, claimed_count, used_count, limit_per_user, start_time, end_time, status, created_at)
		VALUES (?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?)
	`, req.BatchID, req.Denomination, req.Threshold, req.TotalCount, req.LimitPerUser, req.StartTime, req.EndTime, model.CouponStatusActive, time.Now())
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err = VerifyCounts(tx, req.BatchID); err != nil {
		return nil, err
	}

	if err = UpdateStats(tx, req.BatchID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &model.Coupon{
		ID:           id,
		BatchID:      req.BatchID,
		Denomination: req.Denomination,
		Threshold:    req.Threshold,
		TotalCount:   req.TotalCount,
		ClaimedCount: 0,
		UsedCount:    0,
		LimitPerUser: req.LimitPerUser,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Status:       model.CouponStatusActive,
		CreatedAt:    time.Now(),
	}, nil
}

func ClaimCoupon(req *ClaimCouponRequest) (*model.UserCoupon, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var coupon model.Coupon
	err = tx.QueryRow(`
		SELECT id, batch_id, denomination, threshold, total_count, claimed_count, used_count, limit_per_user, start_time, end_time, status, created_at
		FROM coupons WHERE batch_id = ?
	`, req.BatchID).Scan(
		&coupon.ID, &coupon.BatchID, &coupon.Denomination, &coupon.Threshold,
		&coupon.TotalCount, &coupon.ClaimedCount, &coupon.UsedCount, &coupon.LimitPerUser,
		&coupon.StartTime, &coupon.EndTime, &coupon.Status, &coupon.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("优惠券不存在")
	}
	if err != nil {
		return nil, err
	}

	if coupon.Status != model.CouponStatusActive {
		return nil, errors.New("优惠券已过期或已领完")
	}

	if coupon.ClaimedCount >= coupon.TotalCount {
		return nil, fmt.Errorf("已领完")
	}

	var userClaimCount int64
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM user_coupons WHERE user_id = ? AND batch_id = ?
	`, req.UserID, req.BatchID).Scan(&userClaimCount)
	if err != nil {
		return nil, err
	}

	if userClaimCount >= coupon.LimitPerUser {
		return nil, fmt.Errorf("已达领取上限")
	}

	result, err := tx.Exec(`
		INSERT INTO user_coupons (user_id, coupon_id, batch_id, status, claimed_at)
		VALUES (?, ?, ?, ?, ?)
	`, req.UserID, coupon.ID, req.BatchID, model.UserCouponStatusClaimed, time.Now())
	if err != nil {
		return nil, err
	}

	userCouponID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE coupons SET claimed_count = claimed_count + 1 WHERE id = ?
	`, coupon.ID)
	if err != nil {
		return nil, err
	}

	if err = VerifyCounts(tx, req.BatchID); err != nil {
		return nil, err
	}

	if err = UpdateStats(tx, req.BatchID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &model.UserCoupon{
		ID:        userCouponID,
		UserID:    req.UserID,
		CouponID:  coupon.ID,
		BatchID:   req.BatchID,
		Status:    model.UserCouponStatusClaimed,
		ClaimedAt: time.Now(),
	}, nil
}

func RedeemCoupon(req *RedeemCouponRequest) (*model.RedemptionRecord, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userCoupon model.UserCoupon
	err = tx.QueryRow(`
		SELECT id, user_id, coupon_id, batch_id, status, claimed_at, redeemed_at, order_id, redeemed_amount
		FROM user_coupons WHERE id = ?
	`, req.UserCouponID).Scan(
		&userCoupon.ID, &userCoupon.UserID, &userCoupon.CouponID, &userCoupon.BatchID,
		&userCoupon.Status, &userCoupon.ClaimedAt, &userCoupon.RedeemedAt, &userCoupon.OrderID, &userCoupon.RedeemedAmount,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("用户优惠券不存在")
	}
	if err != nil {
		return nil, err
	}

	if userCoupon.Status == model.UserCouponStatusRedeemed {
		return nil, fmt.Errorf("已核销")
	}

	if userCoupon.Status != model.UserCouponStatusClaimed {
		return nil, errors.New("优惠券状态无效")
	}

	var coupon model.Coupon
	err = tx.QueryRow(`
		SELECT id, batch_id, denomination, threshold, total_count, claimed_count, used_count, limit_per_user, start_time, end_time, status, created_at
		FROM coupons WHERE id = ?
	`, userCoupon.CouponID).Scan(
		&coupon.ID, &coupon.BatchID, &coupon.Denomination, &coupon.Threshold,
		&coupon.TotalCount, &coupon.ClaimedCount, &coupon.UsedCount, &coupon.LimitPerUser,
		&coupon.StartTime, &coupon.EndTime, &coupon.Status, &coupon.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if now.Before(coupon.StartTime) || now.After(coupon.EndTime) {
		return nil, fmt.Errorf("不在有效期内，有效期为 %s 至 %s", 
			coupon.StartTime.Format("2006-01-02 15:04:05"), 
			coupon.EndTime.Format("2006-01-02 15:04:05"))
	}

	if req.OrderAmount < coupon.Threshold {
		return nil, fmt.Errorf("不满足使用门槛，门槛为 %d 分，当前订单金额为 %d 分", coupon.Threshold, req.OrderAmount)
	}

	redeemedAmount := coupon.Denomination
	redeemedAt := time.Now()

	_, err = tx.Exec(`
		UPDATE user_coupons 
		SET status = ?, redeemed_at = ?, order_id = ?, redeemed_amount = ?
		WHERE id = ?
	`, model.UserCouponStatusRedeemed, redeemedAt, req.OrderID, redeemedAmount, userCoupon.ID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE coupons SET used_count = used_count + 1 WHERE id = ?
	`, coupon.ID)
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(`
		INSERT INTO redemption_records (user_coupon_id, coupon_id, batch_id, user_id, order_id, order_amount, redeemed_amount, redeemed_at, is_refunded)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, userCoupon.ID, coupon.ID, coupon.BatchID, userCoupon.UserID, req.OrderID, req.OrderAmount, redeemedAmount, redeemedAt)
	if err != nil {
		return nil, err
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err = VerifyCounts(tx, coupon.BatchID); err != nil {
		return nil, err
	}

	if err = UpdateStats(tx, coupon.BatchID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &model.RedemptionRecord{
		ID:             recordID,
		UserCouponID:   userCoupon.ID,
		CouponID:       coupon.ID,
		BatchID:        coupon.BatchID,
		UserID:         userCoupon.UserID,
		OrderID:        req.OrderID,
		OrderAmount:    req.OrderAmount,
		RedeemedAmount: redeemedAmount,
		RedeemedAt:     redeemedAt,
		IsRefunded:     false,
	}, nil
}

func PartialRefund(req *RefundRequest) (int64, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var record model.RedemptionRecord
	err = tx.QueryRow(`
		SELECT id, user_coupon_id, coupon_id, batch_id, user_id, order_id, order_amount, redeemed_amount, redeemed_at, is_refunded
		FROM redemption_records WHERE order_id = ?
	`, req.OrderID).Scan(
		&record.ID, &record.UserCouponID, &record.CouponID, &record.BatchID,
		&record.UserID, &record.OrderID, &record.OrderAmount, &record.RedeemedAmount,
		&record.RedeemedAt, &record.IsRefunded,
	)
	if err == sql.ErrNoRows {
		return 0, errors.New("核销记录不存在")
	}
	if err != nil {
		return 0, err
	}

	if record.IsRefunded {
		return 0, errors.New("该订单已退款")
	}

	if req.RefundAmount <= 0 || req.RefundAmount > record.OrderAmount {
		return 0, errors.New("退款金额无效")
	}

	refundDiscount := int64(float64(record.RedeemedAmount) * float64(req.RefundAmount) / float64(record.OrderAmount))

	if refundDiscount > 0 {
		_, err = tx.Exec(`
			UPDATE redemption_records SET is_refunded = 1 WHERE id = ?
		`, record.ID)
		if err != nil {
			return 0, err
		}

		_, err = tx.Exec(`
			UPDATE user_coupons SET status = ? WHERE id = ?
		`, model.UserCouponStatusClaimed, record.UserCouponID)
		if err != nil {
			return 0, err
		}

		_, err = tx.Exec(`
			UPDATE coupons SET used_count = used_count - 1 WHERE id = ?
		`, record.CouponID)
		if err != nil {
			return 0, err
		}

		if err = VerifyCounts(tx, record.BatchID); err != nil {
			return 0, err
		}

		if err = UpdateStats(tx, record.BatchID); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return refundDiscount, nil
}

func GetRedemptionRecords(batchID string) ([]*model.RedemptionRecord, error) {
	rows, err := database.DB.Query(`
		SELECT id, user_coupon_id, coupon_id, batch_id, user_id, order_id, order_amount, redeemed_amount, redeemed_at, is_refunded
		FROM redemption_records WHERE batch_id = ?
		ORDER BY redeemed_at DESC
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*model.RedemptionRecord, 0)
	for rows.Next() {
		var r model.RedemptionRecord
		err := rows.Scan(
			&r.ID, &r.UserCouponID, &r.CouponID, &r.BatchID, &r.UserID,
			&r.OrderID, &r.OrderAmount, &r.RedeemedAmount, &r.RedeemedAt, &r.IsRefunded,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, &r)
	}

	return records, nil
}

func ExpireCoupons() error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	rows, err := tx.Query(`
		SELECT id, batch_id FROM coupons 
		WHERE status = ? AND end_time < ?
	`, model.CouponStatusActive, now)
	if err != nil {
		return err
	}
	defer rows.Close()

	expiredBatches := make(map[string]bool)
	for rows.Next() {
		var id int64
		var batchID string
		if err := rows.Scan(&id, &batchID); err != nil {
			return err
		}
		expiredBatches[batchID] = true

		_, err = tx.Exec(`UPDATE coupons SET status = ? WHERE id = ?`, model.CouponStatusExpired, id)
		if err != nil {
			return err
		}
	}

	for batchID := range expiredBatches {
		_, err = tx.Exec(`
			UPDATE user_coupons SET status = ? 
			WHERE batch_id = ? AND status = ?
		`, model.UserCouponStatusExpired, batchID, model.UserCouponStatusClaimed)
		if err != nil {
			return err
		}

		if err = VerifyCounts(tx, batchID); err != nil {
			return err
		}

		if err = UpdateStats(tx, batchID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
