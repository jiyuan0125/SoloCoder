package main

import (
	"net/http"
)

type ProcessHandler struct {
	storage *FileStorage
}

func (h *ProcessHandler) ReviewApplication(w http.ResponseWriter, r *http.Request) {
	var req ReviewRequest
	if err := ParseJSONBody(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ApplicationID == "" {
		JSONError(w, http.StatusBadRequest, "application_id is required")
		return
	}

	app, exists := h.storage.GetApplicationByID(req.ApplicationID)
	if !exists {
		JSONError(w, http.StatusNotFound, "application not found")
		return
	}

	if app.Status != StatusPending {
		JSONError(w, http.StatusBadRequest, "application is not in pending status")
		return
	}

	if !req.Approved {
		if req.RejectReason == "" {
			JSONError(w, http.StatusBadRequest, "reject_reason is required when rejecting")
			return
		}
		app.Status = StatusRejected
		app.RejectReason = req.RejectReason
		if err := h.storage.UpdateApplication(app); err != nil {
			JSONError(w, http.StatusInternalServerError, "failed to update application")
			return
		}
		JSONResponse(w, http.StatusOK, app)
		return
	}

	app.Status = StatusApproved
	if err := h.storage.UpdateApplication(app); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to update application")
		return
	}

	if app.Type == TypeReturn {
		refund, err := h.processReturnRefund(app)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		app.Status = StatusCompleted
		h.storage.UpdateApplication(app)
		JSONResponse(w, http.StatusOK, map[string]interface{}{
			"application": app,
			"refund":      refund,
		})
	} else {
		shipping, err := h.processExchangeShipping(app)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		app.Status = StatusProcessing
		h.storage.UpdateApplication(app)
		JSONResponse(w, http.StatusOK, map[string]interface{}{
			"application": app,
			"shipping":    shipping,
		})
	}
}

func (h *ProcessHandler) processReturnRefund(app *ReturnExchangeApplication) (*RefundRecord, error) {
	order, exists := h.storage.GetOrderByID(app.OrderID)
	if !exists {
		return nil, &ProcessError{message: "order not found"}
	}

	refundAmount := order.PaidAmount
	shippingFee := 0.0

	if IsPlatformPaysShipping(app.Reason) {
		shippingFee = order.ShippingFee
	}

	totalRefund := refundAmount + shippingFee

	if totalRefund > order.TotalAmount {
		totalRefund = order.TotalAmount
		refundAmount = totalRefund - shippingFee
		if refundAmount < 0 {
			refundAmount = 0
			shippingFee = totalRefund
		}
	}

	refund := &RefundRecord{
		RefundID:      h.storage.GenerateRefundID(),
		ApplicationID: app.ApplicationID,
		OrderID:       app.OrderID,
		RefundAmount:  refundAmount,
		ShippingFee:   shippingFee,
		TotalRefund:   totalRefund,
		CreatedAt:     GetNowTime(),
	}

	if err := h.storage.CreateRefundRecord(refund); err != nil {
		return nil, &ProcessError{message: "failed to create refund record"}
	}

	return refund, nil
}

func (h *ProcessHandler) processExchangeShipping(app *ReturnExchangeApplication) (*ShippingOrder, error) {
	order, exists := h.storage.GetOrderByID(app.OrderID)
	if !exists {
		return nil, &ProcessError{message: "order not found"}
	}

	newSKUPrice, exists := h.storage.GetSKUPrice(app.NewSKU, app.NewSpec)
	if !exists {
		return nil, &ProcessError{message: "new SKU and spec not found"}
	}

	oldTotalPrice := order.UnitPrice * float64(order.Quantity)
	newTotalPrice := newSKUPrice.Price * float64(order.Quantity)

	priceDifference := newTotalPrice - oldTotalPrice

	needsTopUp := false
	if priceDifference > 0 {
		needsTopUp = true
	}

	if priceDifference < 0 {
		refundAmount := -priceDifference
		var shippingFee float64 = 0
		if IsPlatformPaysShipping(app.Reason) {
			shippingFee = order.ShippingFee
		}

		refund := &RefundRecord{
			RefundID:      h.storage.GenerateRefundID(),
			ApplicationID: app.ApplicationID,
			OrderID:       app.OrderID,
			RefundAmount:  refundAmount,
			ShippingFee:   shippingFee,
			TotalRefund:   refundAmount + shippingFee,
			CreatedAt:     GetNowTime(),
		}
		h.storage.CreateRefundRecord(refund)
	}

	shipping := &ShippingOrder{
		ShippingID:      h.storage.GenerateShippingID(),
		ApplicationID:   app.ApplicationID,
		OrderID:         app.OrderID,
		SKU:             app.NewSKU,
		Spec:            app.NewSpec,
		Quantity:        order.Quantity,
		PriceDifference: priceDifference,
		NeedsTopUp:      needsTopUp,
		Status:          "pending",
		CreatedAt:       GetNowTime(),
	}

	if err := h.storage.CreateShippingOrder(shipping); err != nil {
		return nil, &ProcessError{message: "failed to create shipping order"}
	}

	return shipping, nil
}

func (h *ProcessHandler) GetRefund(w http.ResponseWriter, r *http.Request) {
	refundID, ok := GetIDFromPath(r, "/api/refunds/")
	if !ok || refundID == "" {
		JSONError(w, http.StatusBadRequest, "invalid refund ID")
		return
	}

	refund, exists := h.storage.GetRefundRecordByID(refundID)
	if !exists {
		JSONError(w, http.StatusNotFound, "refund not found")
		return
	}

	JSONResponse(w, http.StatusOK, refund)
}

func (h *ProcessHandler) GetShipping(w http.ResponseWriter, r *http.Request) {
	shippingID, ok := GetIDFromPath(r, "/api/shippings/")
	if !ok || shippingID == "" {
		JSONError(w, http.StatusBadRequest, "invalid shipping ID")
		return
	}

	shipping, exists := h.storage.GetShippingOrderByID(shippingID)
	if !exists {
		JSONError(w, http.StatusNotFound, "shipping order not found")
		return
	}

	JSONResponse(w, http.StatusOK, shipping)
}

func IsPlatformPaysShipping(reason ReturnReason) bool {
	return reason == ReturnReasonQualityIssue ||
		reason == ReturnReasonDamaged ||
		reason == ReturnReasonNotMatch
}

type ProcessError struct {
	message string
}

func (e *ProcessError) Error() string {
	return e.message
}
