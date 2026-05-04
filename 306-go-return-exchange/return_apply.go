package main

import (
	"net/http"
)

type ReturnApplyHandler struct {
	storage *FileStorage
}

func (h *ReturnApplyHandler) CreateReturn(w http.ResponseWriter, r *http.Request) {
	var req CreateReturnRequest
	if err := ParseJSONBody(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OrderID == "" || req.UserID == "" || req.Reason == "" {
		JSONError(w, http.StatusBadRequest, "order_id, user_id and reason are required")
		return
	}

	order, exists := h.storage.GetOrderByID(req.OrderID)
	if !exists {
		JSONError(w, http.StatusNotFound, "order not found")
		return
	}

	if order.UserID != req.UserID {
		JSONError(w, http.StatusForbidden, "order does not belong to user")
		return
	}

	if order.Status != OrderStatusDelivered && order.Status != OrderStatusCompleted {
		JSONError(w, http.StatusBadRequest, "order is not eligible for return")
		return
	}

	existingApps := h.storage.GetApplicationsByOrderID(req.OrderID)
	for _, app := range existingApps {
		if app.Status == StatusCompleted || app.Status == StatusApproved || app.Status == StatusProcessing {
			JSONError(w, http.StatusBadRequest, "order already has an active or completed return/exchange application")
			return
		}
	}

	if req.Reason == ReturnReasonQualityIssue || req.Reason == ReturnReasonDamaged {
		if len(req.Vouchers) == 0 {
			JSONError(w, http.StatusBadRequest, "voucher images are required for quality issues or damaged goods")
			return
		}
	}

	app := &ReturnExchangeApplication{
		ApplicationID: h.storage.GenerateApplicationID(),
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Type:          TypeReturn,
		Reason:        req.Reason,
		ReasonDetail:  req.ReasonDetail,
		Status:        StatusPending,
		CreatedAt:     GetNowTime(),
		UpdatedAt:     GetNowTime(),
	}

	if err := h.storage.CreateApplication(app); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to create application")
		return
	}

	for i, voucherBase64 := range req.Vouchers {
		if voucherBase64 == "" {
			continue
		}
		voucher := &Voucher{
			VoucherID:     h.storage.GenerateVoucherID(),
			ApplicationID: app.ApplicationID,
			ImageBase64:   voucherBase64,
			CreatedAt:     GetNowTime(),
		}
		_ = h.storage.CreateVoucher(voucher)
		_ = i
	}

	JSONResponse(w, http.StatusCreated, app)
}

func (h *ReturnApplyHandler) CreateExchange(w http.ResponseWriter, r *http.Request) {
	var req CreateExchangeRequest
	if err := ParseJSONBody(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OrderID == "" || req.UserID == "" || req.Reason == "" || req.NewSKU == "" || req.NewSpec == "" {
		JSONError(w, http.StatusBadRequest, "order_id, user_id, reason, new_sku and new_spec are required")
		return
	}

	order, exists := h.storage.GetOrderByID(req.OrderID)
	if !exists {
		JSONError(w, http.StatusNotFound, "order not found")
		return
	}

	if order.UserID != req.UserID {
		JSONError(w, http.StatusForbidden, "order does not belong to user")
		return
	}

	if order.Status != OrderStatusDelivered && order.Status != OrderStatusCompleted {
		JSONError(w, http.StatusBadRequest, "order is not eligible for exchange")
		return
	}

	existingApps := h.storage.GetApplicationsByOrderID(req.OrderID)
	for _, app := range existingApps {
		if app.Status == StatusCompleted || app.Status == StatusApproved || app.Status == StatusProcessing {
			JSONError(w, http.StatusBadRequest, "order already has an active or completed return/exchange application")
			return
		}
	}

	_, exists = h.storage.GetSKUPrice(req.NewSKU, req.NewSpec)
	if !exists {
		JSONError(w, http.StatusBadRequest, "new SKU and spec combination not found")
		return
	}

	if req.Reason == ReturnReasonQualityIssue || req.Reason == ReturnReasonDamaged {
		if len(req.Vouchers) == 0 {
			JSONError(w, http.StatusBadRequest, "voucher images are required for quality issues or damaged goods")
			return
		}
	}

	app := &ReturnExchangeApplication{
		ApplicationID: h.storage.GenerateApplicationID(),
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Type:          TypeExchange,
		Reason:        req.Reason,
		ReasonDetail:  req.ReasonDetail,
		NewSKU:        req.NewSKU,
		NewSpec:       req.NewSpec,
		Status:        StatusPending,
		CreatedAt:     GetNowTime(),
		UpdatedAt:     GetNowTime(),
	}

	if err := h.storage.CreateApplication(app); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to create application")
		return
	}

	for _, voucherBase64 := range req.Vouchers {
		if voucherBase64 == "" {
			continue
		}
		voucher := &Voucher{
			VoucherID:     h.storage.GenerateVoucherID(),
			ApplicationID: app.ApplicationID,
			ImageBase64:   voucherBase64,
			CreatedAt:     GetNowTime(),
		}
		_ = h.storage.CreateVoucher(voucher)
	}

	JSONResponse(w, http.StatusCreated, app)
}

func (h *ReturnApplyHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	appID, ok := GetIDFromPath(r, "/api/applications/")
	if !ok || appID == "" {
		JSONError(w, http.StatusBadRequest, "invalid application ID")
		return
	}

	app, exists := h.storage.GetApplicationByID(appID)
	if !exists {
		JSONError(w, http.StatusNotFound, "application not found")
		return
	}

	vouchers := h.storage.GetVouchersByApplicationID(appID)

	response := map[string]interface{}{
		"application": app,
		"vouchers":    vouchers,
	}

	if app.Type == TypeReturn {
		refund, exists := h.storage.GetRefundRecordByApplicationID(appID)
		if exists {
			response["refund"] = refund
		}
	} else {
		shipping, exists := h.storage.GetShippingOrderByApplicationID(appID)
		if exists {
			response["shipping"] = shipping
		}
	}

	JSONResponse(w, http.StatusOK, response)
}

func (h *ReturnApplyHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	orderID := r.URL.Query().Get("order_id")

	var apps []*ReturnExchangeApplication

	if userID != "" {
		apps = h.storage.GetApplicationsByUserID(userID)
	} else if orderID != "" {
		apps = h.storage.GetApplicationsByOrderID(orderID)
	} else {
		JSONError(w, http.StatusBadRequest, "user_id or order_id is required")
		return
	}

	result := make([]map[string]interface{}, 0, len(apps))
	for _, app := range apps {
		item := map[string]interface{}{
			"application_id": app.ApplicationID,
			"order_id":       app.OrderID,
			"user_id":        app.UserID,
			"type":           app.Type,
			"reason":         app.Reason,
			"status":         app.Status,
			"created_at":     app.CreatedAt,
			"updated_at":     app.UpdatedAt,
		}
		result = append(result, item)
	}

	JSONResponse(w, http.StatusOK, result)
}
