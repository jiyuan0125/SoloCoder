package service

import (
	"errors"
	"fmt"

	"return-exchange/database"
)

// 需要凭证的退货原因
var reasonsRequiringEvidence = map[string]bool{
	"质量问题":    true,
	"商品破损":    true,
}

// 平台承担运费的原因
var reasonsPlatformBearsShipping = map[string]bool{
	"质量问题":    true,
	"商品破损":    true,
	"与描述不符":   true,
}

// 买家承担运费的原因
var reasonsBuyerBearsShipping = map[string]bool{
	"不想要了":    true,
	"拍错":      true,
}

// ReturnExchangeService 退换货服务
type ReturnExchangeService struct {
	orderRepo          *database.OrderRepository
	returnExchangeRepo *database.ReturnExchangeRepository
	evidenceRepo       *database.EvidenceRepository
	refundRepo         *database.RefundRepository
	shipmentRepo       *database.ShipmentRepository
}

func NewReturnExchangeService() *ReturnExchangeService {
	return &ReturnExchangeService{
		orderRepo:          database.NewOrderRepository(),
		returnExchangeRepo: database.NewReturnExchangeRepository(),
		evidenceRepo:       database.NewEvidenceRepository(),
		refundRepo:         database.NewRefundRepository(),
		shipmentRepo:       database.NewShipmentRepository(),
	}
}

// CreateReturnRequest 创建退货申请
func (s *ReturnExchangeService) CreateReturnRequest(
	userID int64,
	orderNo string,
	reason string,
	reasonDetail string,
	evidenceImages []string,
) (*database.ReturnExchange, error) {
	// 1. 检查订单是否存在
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	// 2. 检查订单是否已经完成退货退款
	hasCompleted, err := s.returnExchangeRepo.HasCompletedReturn(orderNo)
	if err != nil {
		return nil, fmt.Errorf("检查订单状态失败: %w", err)
	}
	if hasCompleted {
		return nil, errors.New("该订单已完成退货退款，不能再次提交申请")
	}

	// 3. 验证原因是否需要凭证
	if reasonsRequiringEvidence[reason] && len(evidenceImages) == 0 {
		return nil, errors.New("选择该原因需要上传凭证照片")
	}

	// 4. 确定运费承担方
	shippingBearer := database.BearerBuyer
	if reasonsPlatformBearsShipping[reason] {
		shippingBearer = database.BearerPlatform
	}

	// 5. 创建退换货申请
	re := &database.ReturnExchange{
		OrderNo:              orderNo,
		UserID:               userID,
		Type:                 database.TypeReturn,
		Reason:               reason,
		ReasonDetail:         reasonDetail,
		Status:               database.StatusPending,
		OriginalSKU:          order.SKU,
		OriginalSpecification: order.Specification,
		OriginalPrice:        order.OriginalPrice,
		ActualPayment:        order.ActualPayment,
		ShippingFee:          order.ShippingFee,
		ShippingBearer:       shippingBearer,
	}

	id, err := s.returnExchangeRepo.Create(re)
	if err != nil {
		return nil, fmt.Errorf("创建退换货申请失败: %w", err)
	}

	// 6. 保存凭证图片
	if len(evidenceImages) > 0 {
		for _, imgBase64 := range evidenceImages {
			evidence := &database.Evidence{
				ReturnExchangeID: id,
				ImageBase64:      imgBase64,
			}
			_, err := s.evidenceRepo.Create(evidence)
			if err != nil {
				return nil, fmt.Errorf("保存凭证图片失败: %w", err)
			}
		}
	}

	// 7. 获取完整的申请信息
	createdRE, err := s.returnExchangeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取申请信息失败: %w", err)
	}

	return createdRE, nil
}

// CreateExchangeRequest 创建换货申请
func (s *ReturnExchangeService) CreateExchangeRequest(
	userID int64,
	orderNo string,
	reason string,
	reasonDetail string,
	evidenceImages []string,
	newSKU string,
	newSpecification string,
	newPrice float64,
	shippingAddress string,
) (*database.ReturnExchange, error) {
	// 1. 检查订单是否存在
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	// 2. 检查订单是否已经完成退货退款
	hasCompleted, err := s.returnExchangeRepo.HasCompletedReturn(orderNo)
	if err != nil {
		return nil, fmt.Errorf("检查订单状态失败: %w", err)
	}
	if hasCompleted {
		return nil, errors.New("该订单已完成退货退款，不能再次提交申请")
	}

	// 3. 验证换货必填参数
	if newSKU == "" {
		return nil, errors.New("换货需要选择新的SKU")
	}
	if shippingAddress == "" {
		return nil, errors.New("换货需要提供收货地址")
	}

	// 4. 验证原因是否需要凭证
	if reasonsRequiringEvidence[reason] && len(evidenceImages) == 0 {
		return nil, errors.New("选择该原因需要上传凭证照片")
	}

	// 5. 确定运费承担方
	shippingBearer := database.BearerBuyer
	if reasonsPlatformBearsShipping[reason] {
		shippingBearer = database.BearerPlatform
	}

	// 6. 计算差价
	priceDifference := newPrice - order.ActualPayment

	// 7. 创建退换货申请
	re := &database.ReturnExchange{
		OrderNo:              orderNo,
		UserID:               userID,
		Type:                 database.TypeExchange,
		Reason:               reason,
		ReasonDetail:         reasonDetail,
		Status:               database.StatusPending,
		OriginalSKU:          order.SKU,
		OriginalSpecification: order.Specification,
		OriginalPrice:        order.OriginalPrice,
		ActualPayment:        order.ActualPayment,
		ShippingFee:          order.ShippingFee,
		ShippingBearer:       shippingBearer,
		NewSKU:               newSKU,
		NewSpecification:     newSpecification,
		NewPrice:             newPrice,
		PriceDifference:      priceDifference,
	}

	id, err := s.returnExchangeRepo.Create(re)
	if err != nil {
		return nil, fmt.Errorf("创建换货申请失败: %w", err)
	}

	// 8. 保存凭证图片
	if len(evidenceImages) > 0 {
		for _, imgBase64 := range evidenceImages {
			evidence := &database.Evidence{
				ReturnExchangeID: id,
				ImageBase64:      imgBase64,
			}
			_, err := s.evidenceRepo.Create(evidence)
			if err != nil {
				return nil, fmt.Errorf("保存凭证图片失败: %w", err)
			}
		}
	}

	// 9. 获取完整的申请信息
	createdRE, err := s.returnExchangeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取申请信息失败: %w", err)
	}

	return createdRE, nil
}

// ApproveRequest 审核通过申请
func (s *ReturnExchangeService) ApproveRequest(
	id int64,
	adminID int64,
	comment string,
	shippingAddress string,
) error {
	// 1. 获取申请信息
	re, err := s.returnExchangeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("查询申请失败: %w", err)
	}
	if re == nil {
		return errors.New("申请不存在")
	}

	// 2. 检查状态是否可以审核
	if re.Status != database.StatusPending {
		return errors.New("只有待审核状态的申请才能审核")
	}

	// 3. 更新申请状态为已同意
	err = s.returnExchangeRepo.UpdateStatus(id, database.StatusApproved, adminID, comment)
	if err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}

	// 4. 根据类型处理后续逻辑
	if re.Type == database.TypeReturn {
		// 退货：生成退款记录
		err = s.createRefundForReturn(re)
		if err != nil {
			return fmt.Errorf("创建退款记录失败: %w", err)
		}
	} else if re.Type == database.TypeExchange {
		// 换货：创建发货单
		err = s.createShipmentForExchange(re, shippingAddress)
		if err != nil {
			return fmt.Errorf("创建发货单失败: %w", err)
		}

		// 如果新SKU更便宜，需要退差价
		if re.PriceDifference < 0 {
			err = s.createRefundForPriceDifference(re, -re.PriceDifference)
			if err != nil {
				return fmt.Errorf("创建差价退款失败: %w", err)
			}
		}
	}

	return nil
}

// RejectRequest 审核拒绝申请
func (s *ReturnExchangeService) RejectRequest(
	id int64,
	adminID int64,
	comment string,
) error {
	// 1. 获取申请信息
	re, err := s.returnExchangeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("查询申请失败: %w", err)
	}
	if re == nil {
		return errors.New("申请不存在")
	}

	// 2. 检查状态是否可以审核
	if re.Status != database.StatusPending {
		return errors.New("只有待审核状态的申请才能审核")
	}

	// 3. 更新申请状态为已拒绝
	err = s.returnExchangeRepo.UpdateStatus(id, database.StatusRejected, adminID, comment)
	if err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}

	return nil
}

// GetRequestByID 获取申请详情
func (s *ReturnExchangeService) GetRequestByID(id int64) (*database.ReturnExchange, error) {
	re, err := s.returnExchangeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("查询申请失败: %w", err)
	}
	return re, nil
}

// GetRequestsByUserID 获取用户的所有申请
func (s *ReturnExchangeService) GetRequestsByUserID(userID int64) ([]*database.ReturnExchange, error) {
	list, err := s.returnExchangeRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户申请失败: %w", err)
	}
	return list, nil
}

// GetRefundByRequestID 获取退款记录
func (s *ReturnExchangeService) GetRefundByRequestID(returnExchangeID int64) (*database.Refund, error) {
	refund, err := s.refundRepo.GetByReturnExchangeID(returnExchangeID)
	if err != nil {
		return nil, fmt.Errorf("查询退款记录失败: %w", err)
	}
	return refund, nil
}

// GetShipmentByRequestID 获取发货单
func (s *ReturnExchangeService) GetShipmentByRequestID(returnExchangeID int64) (*database.Shipment, error) {
	shipment, err := s.shipmentRepo.GetByReturnExchangeID(returnExchangeID)
	if err != nil {
		return nil, fmt.Errorf("查询发货单失败: %w", err)
	}
	return shipment, nil
}

// createRefundForReturn 为退货创建退款记录
func (s *ReturnExchangeService) createRefundForReturn(re *database.ReturnExchange) error {
	// 生成退款单号
	refundNo, err := s.refundRepo.GenerateRefundNo()
	if err != nil {
		return err
	}

	// 计算退款金额
	refundAmount := re.ActualPayment
	shippingFeeRefund := 0.0

	// 如果平台承担运费，加上运费
	if re.ShippingBearer == database.BearerPlatform {
		shippingFeeRefund = re.ShippingFee
	}

	totalRefund := refundAmount + shippingFeeRefund

	// 确保退款金额不超过原单金额（实际支付+运费）
	maxRefund := re.ActualPayment + re.ShippingFee
	if totalRefund > maxRefund {
		totalRefund = maxRefund
	}

	// 创建退款记录
	refund := &database.Refund{
		ReturnExchangeID:  re.ID,
		RefundNo:          refundNo,
		RefundAmount:      refundAmount,
		ShippingFeeRefund: shippingFeeRefund,
		TotalRefund:       totalRefund,
		Status:            database.StatusProcessing,
	}

	_, err = s.refundRepo.Create(refund)
	return err
}

// createShipmentForExchange 为换货创建发货单
func (s *ReturnExchangeService) createShipmentForExchange(re *database.ReturnExchange, shippingAddress string) error {
	// 生成发货单号
	shipmentNo, err := s.shipmentRepo.GenerateShipmentNo()
	if err != nil {
		return err
	}

	// 创建发货单
	shipment := &database.Shipment{
		ReturnExchangeID: re.ID,
		ShipmentNo:       shipmentNo,
		SKU:              re.NewSKU,
		Specification:    re.NewSpecification,
		Quantity:         1,
		ShippingAddress:  shippingAddress,
		Status:           database.StatusPending,
	}

	_, err = s.shipmentRepo.Create(shipment)
	return err
}

// createRefundForPriceDifference 为差价创建退款记录
func (s *ReturnExchangeService) createRefundForPriceDifference(re *database.ReturnExchange, difference float64) error {
	// 生成退款单号
	refundNo, err := s.refundRepo.GenerateRefundNo()
	if err != nil {
		return err
	}

	// 创建差价退款记录
	refund := &database.Refund{
		ReturnExchangeID:  re.ID,
		RefundNo:          refundNo,
		RefundAmount:      difference,
		ShippingFeeRefund: 0,
		TotalRefund:       difference,
		Status:            database.StatusProcessing,
	}

	_, err = s.refundRepo.Create(refund)
	return err
}

// GetOrderByNo 获取订单信息
func (s *ReturnExchangeService) GetOrderByNo(orderNo string) (*database.Order, error) {
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}
	return order, nil
}
