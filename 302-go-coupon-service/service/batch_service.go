package service

import (
	"coupon-service/dao"
	"coupon-service/model"
	"errors"
	"fmt"
	"time"
)

var (
	ErrBatchNameEmpty       = errors.New("优惠券批次名称不能为空")
	ErrDiscountAmountZero   = errors.New("优惠券面额必须大于零")
	ErrInvalidDateRange     = errors.New("有效期起始日期必须早于结束日期")
	ErrTotalQuantityZero    = errors.New("发放总量必须大于零")
	ErrLimitPerUserZero     = errors.New("每人限领数量必须大于零")
	ErrBatchNameDuplicate   = errors.New("优惠券批次名称已存在")
	ErrBatchNotFound        = errors.New("优惠券批次不存在")
	ErrBatchAlreadyIssued   = errors.New("该批次已发放完毕，不能修改参数")
)

type BatchService struct {
	batchDAO *dao.BatchDAO
}

func NewBatchService() *BatchService {
	return &BatchService{
		batchDAO: dao.NewBatchDAO(),
	}
}

func (s *BatchService) CreateBatch(req *model.CreateBatchRequest) (*model.CouponBatch, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.batchDAO.GetBatchByName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("检查批次名称时出错: %w", err)
	}
	if existing != nil {
		return nil, ErrBatchNameDuplicate
	}

	batch := &model.CouponBatch{
		Name:            req.Name,
		DiscountAmount:  req.DiscountAmount,
		ThresholdAmount: req.ThresholdAmount,
		TotalQuantity:   req.TotalQuantity,
		ValidStart:      req.ValidStart,
		ValidEnd:        req.ValidEnd,
		LimitPerUser:    req.LimitPerUser,
	}

	id, err := s.batchDAO.CreateBatch(batch)
	if err != nil {
		return nil, fmt.Errorf("创建批次失败: %w", err)
	}

	batch.ID = id
	return batch, nil
}

func (s *BatchService) validateCreateRequest(req *model.CreateBatchRequest) error {
	if req.Name == "" {
		return ErrBatchNameEmpty
	}
	if req.DiscountAmount <= 0 {
		return ErrDiscountAmountZero
	}
	if req.TotalQuantity <= 0 {
		return ErrTotalQuantityZero
	}
	if req.LimitPerUser <= 0 {
		return ErrLimitPerUserZero
	}
	if !req.ValidStart.Before(req.ValidEnd) {
		return ErrInvalidDateRange
	}
	return nil
}

func (s *BatchService) UpdateBatch(batchID int64, req *model.UpdateBatchRequest) (*model.CouponBatch, error) {
	batch, err := s.batchDAO.GetBatchByID(batchID)
	if err != nil {
		return nil, fmt.Errorf("获取批次信息失败: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}

	if batch.IssuedQuantity >= batch.TotalQuantity {
		return nil, ErrBatchAlreadyIssued
	}

	if req.Name != nil && *req.Name != batch.Name {
		existing, err := s.batchDAO.GetBatchByName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("检查批次名称时出错: %w", err)
		}
		if existing != nil && existing.ID != batchID {
			return nil, ErrBatchNameDuplicate
		}
		if *req.Name == "" {
			return nil, ErrBatchNameEmpty
		}
	}

	if req.Name != nil {
		batch.Name = *req.Name
	}
	if req.DiscountAmount != nil {
		if *req.DiscountAmount <= 0 {
			return nil, ErrDiscountAmountZero
		}
		batch.DiscountAmount = *req.DiscountAmount
	}
	if req.ThresholdAmount != nil {
		batch.ThresholdAmount = *req.ThresholdAmount
	}
	if req.ValidStart != nil {
		batch.ValidStart = *req.ValidStart
	}
	if req.ValidEnd != nil {
		batch.ValidEnd = *req.ValidEnd
	}
	if req.LimitPerUser != nil {
		if *req.LimitPerUser <= 0 {
			return nil, ErrLimitPerUserZero
		}
		batch.LimitPerUser = *req.LimitPerUser
	}

	if !batch.ValidStart.Before(batch.ValidEnd) {
		return nil, ErrInvalidDateRange
	}

	if err := s.batchDAO.UpdateBatch(batch); err != nil {
		return nil, fmt.Errorf("更新批次失败: %w", err)
	}

	return batch, nil
}

func (s *BatchService) GetBatchByID(batchID int64) (*model.CouponBatch, error) {
	batch, err := s.batchDAO.GetBatchByID(batchID)
	if err != nil {
		return nil, fmt.Errorf("获取批次信息失败: %w", err)
	}
	if batch == nil {
		return nil, ErrBatchNotFound
	}
	return batch, nil
}

func (s *BatchService) GetBatchProgress(batchID int64) (*model.BatchProgress, error) {
	progress, err := s.batchDAO.GetBatchProgress(batchID)
	if err != nil {
		return nil, fmt.Errorf("获取批次进度失败: %w", err)
	}
	if progress == nil {
		return nil, ErrBatchNotFound
	}
	return progress, nil
}

func (s *BatchService) ListBatches() ([]*model.CouponBatch, error) {
	batches, err := s.batchDAO.ListBatches()
	if err != nil {
		return nil, fmt.Errorf("获取批次列表失败: %w", err)
	}
	return batches, nil
}

func (s *BatchService) IsBatchFullyIssued(batchID int64) (bool, error) {
	batch, err := s.batchDAO.GetBatchByID(batchID)
	if err != nil {
		return false, fmt.Errorf("获取批次信息失败: %w", err)
	}
	if batch == nil {
		return false, ErrBatchNotFound
	}
	return batch.IssuedQuantity >= batch.TotalQuantity, nil
}

func (s *BatchService) IsCouponValidForRedeem(batch *model.CouponBatch, orderAmount int) error {
	now := time.Now()

	validEnd := time.Date(
		batch.ValidEnd.Year(),
		batch.ValidEnd.Month(),
		batch.ValidEnd.Day(),
		23, 59, 59, 0,
		batch.ValidEnd.Location(),
	)

	if now.Before(batch.ValidStart) {
		return errors.New("优惠券尚未开始生效")
	}
	if now.After(validEnd) {
		return errors.New("优惠券已过期")
	}

	if orderAmount < batch.ThresholdAmount {
		return errors.New(fmt.Sprintf("订单金额不满足使用门槛，需要满%d元", batch.ThresholdAmount))
	}

	return nil
}
