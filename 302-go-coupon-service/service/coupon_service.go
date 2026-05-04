package service

import (
	"coupon-service/dao"
	"coupon-service/model"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrCouponNotFound       = errors.New("优惠券不存在")
	ErrCouponAlreadyRedeemed = errors.New("优惠券已被使用")
	ErrCouponExpired        = errors.New("优惠券已过期")
	ErrCouponUserMismatch   = errors.New("优惠券不属于当前用户")
	ErrCouponLimitExceeded  = errors.New("已达到该批次每人限领数量")
	ErrCouponNoStock        = errors.New("该批次优惠券已发放完毕")
	ErrCouponReturned       = errors.New("优惠券已被退回")
	ErrNoCouponsToReturn    = errors.New("没有可退回的优惠券")
)

type CouponService struct {
	couponDAO *dao.CouponDAO
	batchDAO  *dao.BatchDAO
	mu        sync.Map
}

func NewCouponService() *CouponService {
	return &CouponService{
		couponDAO: dao.NewCouponDAO(),
		batchDAO:  dao.NewBatchDAO(),
	}
}

func (s *CouponService) generateCouponCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("CP%s", hex.EncodeToString(b))
}

func (s *CouponService) getLockKey(batchID int64, userID string) string {
	return fmt.Sprintf("%d-%s", batchID, userID)
}

func (s *CouponService) ClaimCoupon(userID string, batchID int64) (*model.Coupon, error) {
	batch, err := s.batchDAO.GetBatchByID(batchID)
	if err != nil {
		return nil, fmt.Errorf("获取批次信息失败: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}

	if batch.IssuedQuantity >= batch.TotalQuantity {
		return nil, ErrCouponNoStock
	}

	lockKey := s.getLockKey(batchID, userID)
	lock, _ := s.mu.LoadOrStore(lockKey, &sync.Mutex{})
	mtx := lock.(*sync.Mutex)
	mtx.Lock()
	defer mtx.Unlock()

	count, err := s.couponDAO.CountUserCouponsByBatch(userID, batchID)
	if err != nil {
		return nil, fmt.Errorf("检查用户领券记录失败: %w", err)
	}
	if count >= batch.LimitPerUser {
		return nil, ErrCouponLimitExceeded
	}

	tx, err := s.couponDAO.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	coupon := &model.Coupon{
		BatchID: batchID,
		UserID:  userID,
		Code:    s.generateCouponCode(),
		Status:  model.CouponStatusIssued,
	}

	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		coupon.Code = s.generateCouponCode()
		existing, err := s.couponDAO.GetCouponByCode(coupon.Code)
		if err != nil {
			continue
		}
		if existing == nil {
			break
		}
	}

	couponID, err := s.couponDAO.CreateCoupon(coupon, tx)
	if err != nil {
		return nil, fmt.Errorf("创建优惠券失败: %w", err)
	}

	if err := s.batchDAO.IncrementIssuedQuantity(batchID, tx); err != nil {
		return nil, fmt.Errorf("更新发放数量失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	coupon.ID = couponID
	coupon.CreatedAt = time.Now()
	coupon.UpdatedAt = time.Now()
	return coupon, nil
}

func (s *CouponService) RedeemCoupon(userID string, couponCode string, orderAmount int) (*model.RedeemCouponResponse, error) {
	coupon, err := s.couponDAO.GetCouponByCode(couponCode)
	if err != nil {
		return nil, fmt.Errorf("获取优惠券信息失败: %w", err)
	}
	if coupon == nil {
		return nil, ErrCouponNotFound
	}

	if coupon.Status != model.CouponStatusIssued {
		if coupon.Status == model.CouponStatusRedeemed {
			return nil, ErrCouponAlreadyRedeemed
		}
		if coupon.Status == model.CouponStatusReturned {
			return nil, ErrCouponReturned
		}
		return nil, errors.New("优惠券状态异常")
	}

	if coupon.UserID != userID {
		return nil, ErrCouponUserMismatch
	}

	batch, err := s.batchDAO.GetBatchByID(coupon.BatchID)
	if err != nil {
		return nil, fmt.Errorf("获取批次信息失败: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}

	now := time.Now()
	validEnd := time.Date(
		batch.ValidEnd.Year(),
		batch.ValidEnd.Month(),
		batch.ValidEnd.Day(),
		23, 59, 59, 0,
		batch.ValidEnd.Location(),
	)

	if now.Before(batch.ValidStart) {
		return nil, errors.New("优惠券尚未开始生效")
	}
	if now.After(validEnd) {
		return nil, ErrCouponExpired
	}

	if orderAmount < batch.ThresholdAmount {
		return nil, errors.New(fmt.Sprintf("订单金额不满足使用门槛，需要满%d元", batch.ThresholdAmount))
	}

	tx, err := s.couponDAO.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	if err := s.couponDAO.RedeemCoupon(coupon.ID, tx); err != nil {
		return nil, fmt.Errorf("核销优惠券失败: %w", err)
	}

	if err := s.batchDAO.IncrementRedeemedQuantity(coupon.BatchID, tx); err != nil {
		return nil, fmt.Errorf("更新核销数量失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &model.RedeemCouponResponse{
		CouponID:    coupon.ID,
		BatchID:     coupon.BatchID,
		Code:        coupon.Code,
		RedeemedAt:  time.Now(),
		DiscountAmt: batch.DiscountAmount,
	}, nil
}

func (s *CouponService) GetUserAvailableCoupons(userID string) ([]*model.Coupon, error) {
	coupons, err := s.couponDAO.GetUserAvailableCoupons(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户可用优惠券失败: %w", err)
	}
	return coupons, nil
}

func (s *CouponService) ReturnCoupons(userID string, batchID int64) (int, error) {
	batch, err := s.batchDAO.GetBatchByID(batchID)
	if err != nil {
		return 0, fmt.Errorf("获取批次信息失败: %w", err)
	}
	if batch == nil {
		return 0, ErrBatchNotFound
	}

	coupons, err := s.couponDAO.GetUserIssuedCouponsByBatch(userID, batchID)
	if err != nil {
		return 0, fmt.Errorf("获取用户优惠券失败: %w", err)
	}

	if len(coupons) == 0 {
		return 0, ErrNoCouponsToReturn
	}

	tx, err := s.couponDAO.BeginTx()
	if err != nil {
		return 0, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	returnedCount := 0
	for _, coupon := range coupons {
		if coupon.Status == model.CouponStatusIssued {
			if err := s.couponDAO.ReturnCoupon(coupon.ID, tx); err != nil {
				return 0, fmt.Errorf("退回优惠券失败: %w", err)
			}
			returnedCount++
		}
	}

	if returnedCount > 0 {
		if err := s.batchDAO.DecrementIssuedQuantity(batchID, returnedCount, tx); err != nil {
			return 0, fmt.Errorf("更新发放数量失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return returnedCount, nil
}

func (s *CouponService) GetCouponByCode(code string) (*model.Coupon, error) {
	coupon, err := s.couponDAO.GetCouponByCode(code)
	if err != nil {
		return nil, fmt.Errorf("获取优惠券信息失败: %w", err)
	}
	return coupon, nil
}
