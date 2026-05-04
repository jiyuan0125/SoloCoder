package database

import (
	"database/sql"
	"fmt"
	"time"
)

// OrderRepository 订单数据访问
type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

func (r *OrderRepository) GetByOrderNo(orderNo string) (*Order, error) {
	query := `
		SELECT id, order_no, user_id, sku, specification, original_price, actual_payment, shipping_fee, created_at
		FROM orders WHERE order_no = ?
	`
	
	var order Order
	err := DB.QueryRow(query, orderNo).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.SKU, &order.Specification,
		&order.OriginalPrice, &order.ActualPayment, &order.ShippingFee, &order.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &order, nil
}

func (r *OrderRepository) Create(order *Order) (int64, error) {
	query := `
		INSERT INTO orders (order_no, user_id, sku, specification, original_price, actual_payment, shipping_fee)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := DB.Exec(query, 
		order.OrderNo, order.UserID, order.SKU, order.Specification,
		order.OriginalPrice, order.ActualPayment, order.ShippingFee,
	)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

// ReturnExchangeRepository 退换货申请数据访问
type ReturnExchangeRepository struct{}

func NewReturnExchangeRepository() *ReturnExchangeRepository {
	return &ReturnExchangeRepository{}
}

func (r *ReturnExchangeRepository) Create(re *ReturnExchange) (int64, error) {
	query := `
		INSERT INTO return_exchanges (
			order_no, user_id, type, reason, reason_detail, status,
			original_sku, original_specification, original_price, actual_payment, shipping_fee, shipping_bearer,
			new_sku, new_specification, new_price, price_difference
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := DB.Exec(query,
		re.OrderNo, re.UserID, re.Type, re.Reason, re.ReasonDetail, re.Status,
		re.OriginalSKU, re.OriginalSpecification, re.OriginalPrice, re.ActualPayment, re.ShippingFee, re.ShippingBearer,
		re.NewSKU, re.NewSpecification, re.NewPrice, re.PriceDifference,
	)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func (r *ReturnExchangeRepository) GetByID(id int64) (*ReturnExchange, error) {
	query := `
		SELECT id, order_no, user_id, type, reason, reason_detail, status,
			original_sku, original_specification, original_price, actual_payment, shipping_fee, shipping_bearer,
			new_sku, new_specification, new_price, price_difference,
			admin_id, review_comment, reviewed_at, created_at, updated_at
		FROM return_exchanges WHERE id = ?
	`
	
	var re ReturnExchange
	var adminID sql.NullInt64
	var reviewComment sql.NullString
	var reviewedAt sql.NullTime
	
	err := DB.QueryRow(query, id).Scan(
		&re.ID, &re.OrderNo, &re.UserID, &re.Type, &re.Reason, &re.ReasonDetail, &re.Status,
		&re.OriginalSKU, &re.OriginalSpecification, &re.OriginalPrice, &re.ActualPayment, &re.ShippingFee, &re.ShippingBearer,
		&re.NewSKU, &re.NewSpecification, &re.NewPrice, &re.PriceDifference,
		&adminID, &reviewComment, &reviewedAt, &re.CreatedAt, &re.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if adminID.Valid {
		re.AdminID = adminID.Int64
	}
	if reviewComment.Valid {
		re.ReviewComment = reviewComment.String
	}
	if reviewedAt.Valid {
		re.ReviewedAt = reviewedAt.Time
	}
	
	return &re, nil
}

func (r *ReturnExchangeRepository) GetByOrderNo(orderNo string) (*ReturnExchange, error) {
	query := `
		SELECT id, order_no, user_id, type, reason, reason_detail, status,
			original_sku, original_specification, original_price, actual_payment, shipping_fee, shipping_bearer,
			new_sku, new_specification, new_price, price_difference,
			admin_id, review_comment, reviewed_at, created_at, updated_at
		FROM return_exchanges WHERE order_no = ?
	`
	
	var re ReturnExchange
	var adminID sql.NullInt64
	var reviewComment sql.NullString
	var reviewedAt sql.NullTime
	
	err := DB.QueryRow(query, orderNo).Scan(
		&re.ID, &re.OrderNo, &re.UserID, &re.Type, &re.Reason, &re.ReasonDetail, &re.Status,
		&re.OriginalSKU, &re.OriginalSpecification, &re.OriginalPrice, &re.ActualPayment, &re.ShippingFee, &re.ShippingBearer,
		&re.NewSKU, &re.NewSpecification, &re.NewPrice, &re.PriceDifference,
		&adminID, &reviewComment, &reviewedAt, &re.CreatedAt, &re.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if adminID.Valid {
		re.AdminID = adminID.Int64
	}
	if reviewComment.Valid {
		re.ReviewComment = reviewComment.String
	}
	if reviewedAt.Valid {
		re.ReviewedAt = reviewedAt.Time
	}
	
	return &re, nil
}

func (r *ReturnExchangeRepository) GetByUserID(userID int64) ([]*ReturnExchange, error) {
	query := `
		SELECT id, order_no, user_id, type, reason, reason_detail, status,
			original_sku, original_specification, original_price, actual_payment, shipping_fee, shipping_bearer,
			new_sku, new_specification, new_price, price_difference,
			admin_id, review_comment, reviewed_at, created_at, updated_at
		FROM return_exchanges WHERE user_id = ? ORDER BY created_at DESC
	`
	
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var list []*ReturnExchange
	for rows.Next() {
		var re ReturnExchange
		var adminID sql.NullInt64
		var reviewComment sql.NullString
		var reviewedAt sql.NullTime
		
		err := rows.Scan(
			&re.ID, &re.OrderNo, &re.UserID, &re.Type, &re.Reason, &re.ReasonDetail, &re.Status,
			&re.OriginalSKU, &re.OriginalSpecification, &re.OriginalPrice, &re.ActualPayment, &re.ShippingFee, &re.ShippingBearer,
			&re.NewSKU, &re.NewSpecification, &re.NewPrice, &re.PriceDifference,
			&adminID, &reviewComment, &reviewedAt, &re.CreatedAt, &re.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if adminID.Valid {
			re.AdminID = adminID.Int64
		}
		if reviewComment.Valid {
			re.ReviewComment = reviewComment.String
		}
		if reviewedAt.Valid {
			re.ReviewedAt = reviewedAt.Time
		}
		
		list = append(list, &re)
	}
	
	return list, nil
}

func (r *ReturnExchangeRepository) UpdateStatus(id int64, status string, adminID int64, comment string) error {
	query := `
		UPDATE return_exchanges 
		SET status = ?, admin_id = ?, review_comment = ?, reviewed_at = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := DB.Exec(query, status, adminID, comment, now, now, id)
	return err
}

func (r *ReturnExchangeRepository) HasCompletedReturn(orderNo string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM return_exchanges 
		WHERE order_no = ? AND status = ?
	`
	
	var count int
	err := DB.QueryRow(query, orderNo, StatusCompleted).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// EvidenceRepository 凭证图片数据访问
type EvidenceRepository struct{}

func NewEvidenceRepository() *EvidenceRepository {
	return &EvidenceRepository{}
}

func (r *EvidenceRepository) Create(evidence *Evidence) (int64, error) {
	query := `
		INSERT INTO evidences (return_exchange_id, image_base64)
		VALUES (?, ?)
	`
	
	result, err := DB.Exec(query, evidence.ReturnExchangeID, evidence.ImageBase64)
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func (r *EvidenceRepository) GetByReturnExchangeID(returnExchangeID int64) ([]*Evidence, error) {
	query := `
		SELECT id, return_exchange_id, image_base64, created_at
		FROM evidences WHERE return_exchange_id = ?
	`
	
	rows, err := DB.Query(query, returnExchangeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var list []*Evidence
	for rows.Next() {
		var evidence Evidence
		err := rows.Scan(&evidence.ID, &evidence.ReturnExchangeID, &evidence.ImageBase64, &evidence.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &evidence)
	}
	
	return list, nil
}

// RefundRepository 退款记录数据访问
type RefundRepository struct{}

func NewRefundRepository() *RefundRepository {
	return &RefundRepository{}
}

func (r *RefundRepository) Create(refund *Refund) (int64, error) {
	query := `
		INSERT INTO refunds (return_exchange_id, refund_no, refund_amount, shipping_fee_refund, total_refund, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	result, err := DB.Exec(query, 
		refund.ReturnExchangeID, refund.RefundNo, refund.RefundAmount, 
		refund.ShippingFeeRefund, refund.TotalRefund, refund.Status,
	)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func (r *RefundRepository) GetByReturnExchangeID(returnExchangeID int64) (*Refund, error) {
	query := `
		SELECT id, return_exchange_id, refund_no, refund_amount, shipping_fee_refund, total_refund, status, created_at, completed_at
		FROM refunds WHERE return_exchange_id = ?
	`
	
	var refund Refund
	var completedAt sql.NullTime
	
	err := DB.QueryRow(query, returnExchangeID).Scan(
		&refund.ID, &refund.ReturnExchangeID, &refund.RefundNo, &refund.RefundAmount,
		&refund.ShippingFeeRefund, &refund.TotalRefund, &refund.Status, &refund.CreatedAt, &completedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if completedAt.Valid {
		refund.CompletedAt = completedAt.Time
	}
	
	return &refund, nil
}

func (r *RefundRepository) GenerateRefundNo() (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	
	query := `
		SELECT COUNT(*) FROM refunds WHERE refund_no LIKE ?
	`
	
	likePattern := fmt.Sprintf("RF%s%%", dateStr)
	var count int
	err := DB.QueryRow(query, likePattern).Scan(&count)
	if err != nil {
		return "", err
	}
	
	refundNo := fmt.Sprintf("RF%s%06d", dateStr, count+1)
	return refundNo, nil
}

// ShipmentRepository 发货单数据访问
type ShipmentRepository struct{}

func NewShipmentRepository() *ShipmentRepository {
	return &ShipmentRepository{}
}

func (r *ShipmentRepository) Create(shipment *Shipment) (int64, error) {
	query := `
		INSERT INTO shipments (return_exchange_id, shipment_no, sku, specification, quantity, shipping_address, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := DB.Exec(query, 
		shipment.ReturnExchangeID, shipment.ShipmentNo, shipment.SKU, 
		shipment.Specification, shipment.Quantity, shipment.ShippingAddress, shipment.Status,
	)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func (r *ShipmentRepository) GetByReturnExchangeID(returnExchangeID int64) (*Shipment, error) {
	query := `
		SELECT id, return_exchange_id, shipment_no, sku, specification, quantity, shipping_address, status, shipped_at, delivered_at, created_at
		FROM shipments WHERE return_exchange_id = ?
	`
	
	var shipment Shipment
	var shippedAt sql.NullTime
	var deliveredAt sql.NullTime
	
	err := DB.QueryRow(query, returnExchangeID).Scan(
		&shipment.ID, &shipment.ReturnExchangeID, &shipment.ShipmentNo, &shipment.SKU,
		&shipment.Specification, &shipment.Quantity, &shipment.ShippingAddress, &shipment.Status,
		&shippedAt, &deliveredAt, &shipment.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if shippedAt.Valid {
		shipment.ShippedAt = shippedAt.Time
	}
	if deliveredAt.Valid {
		shipment.DeliveredAt = deliveredAt.Time
	}
	
	return &shipment, nil
}

func (r *ShipmentRepository) GenerateShipmentNo() (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	
	query := `
		SELECT COUNT(*) FROM shipments WHERE shipment_no LIKE ?
	`
	
	likePattern := fmt.Sprintf("SH%s%%", dateStr)
	var count int
	err := DB.QueryRow(query, likePattern).Scan(&count)
	if err != nil {
		return "", err
	}
	
	shipmentNo := fmt.Sprintf("SH%s%06d", dateStr, count+1)
	return shipmentNo, nil
}
