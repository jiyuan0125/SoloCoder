package common

import "return-exchange/pkg/models"

type SubmitReturnRequest struct {
	OrderID        string                `json:"order_id"`
	UserID         string                `json:"user_id"`
	Reason         models.ReturnReason   `json:"reason"`
	EvidenceImages []string              `json:"evidence_images"`
}

type SubmitExchangeRequest struct {
	OrderID          string                `json:"order_id"`
	UserID           string                `json:"user_id"`
	Reason           models.ReturnReason   `json:"reason"`
	EvidenceImages   []string              `json:"evidence_images"`
	NewSKU           string                `json:"new_sku"`
	NewSpecification string                `json:"new_specification"`
	NewSKUPrice      float64               `json:"new_sku_price"`
}

type ReviewRequest struct {
	ApplicationID string                    `json:"application_id"`
	Approved      bool                      `json:"approved"`
	ShippingFee   float64                   `json:"shipping_fee,omitempty"`
}

type CommonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SubmitResponse struct {
	CommonResponse
	ApplicationID string `json:"application_id,omitempty"`
}

type ApplicationDetailResponse struct {
	CommonResponse
	Application *models.ReturnExchangeApplication `json:"application,omitempty"`
}

type ApplicationListResponse struct {
	CommonResponse
	Applications []models.ReturnExchangeApplication `json:"applications,omitempty"`
}

type ProcessRefundRequest struct {
	ApplicationID string `json:"application_id"`
}

type ProcessRefundResponse struct {
	CommonResponse
	RefundID     string  `json:"refund_id,omitempty"`
	TotalRefund  float64 `json:"total_refund,omitempty"`
}

type CreateShippingOrderRequest struct {
	ApplicationID string `json:"application_id"`
}

type CreateShippingOrderResponse struct {
	CommonResponse
	ShippingOrderID string `json:"shipping_order_id,omitempty"`
}
