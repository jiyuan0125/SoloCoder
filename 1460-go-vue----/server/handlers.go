package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"supplier-portal/common"
	"supplier-portal/core"
)

type Server struct {
	store *core.Store
}

func NewServer(store *core.Store) *Server {
	return &Server{store: store}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, common.ErrorResponse{Error: err.Error()})
}

func parseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func toInquiryResponse(inq *core.Inquiry) common.InquiryResponse {
	return common.InquiryResponse{
		ID:               inq.ID,
		Title:            inq.Title,
		BuyerID:          inq.BuyerID,
		Materials:        common.ToMaterialDTOs(inq.Materials),
		SupplierIDs:      inq.SupplierIDs,
		Deadline:         formatTime(inq.Deadline),
		CreatedAt:        formatTime(inq.CreatedAt),
		UpdatedAt:        formatTime(inq.UpdatedAt),
		Status:           string(inq.Status),
		MaterialsVersion: inq.MaterialsVersion,
	}
}

func toQuotationResponse(q *core.Quotation) common.QuotationResponse {
	items := make([]common.QuotationItemDTO, len(q.Items))
	for i, item := range q.Items {
		items[i] = common.QuotationItemDTO{
			MaterialID:   item.MaterialID,
			UnitPrice:    item.UnitPrice,
			DeliveryDays: item.DeliveryDays,
		}
	}

	return common.QuotationResponse{
		ID:               q.ID,
		InquiryID:        q.InquiryID,
		SupplierID:       q.SupplierID,
		Items:            items,
		TotalPrice:       q.TotalPrice,
		Remark:           q.Remark,
		SubmittedAt:      formatTime(q.SubmittedAt),
		Status:           string(q.Status),
		MaterialsVersion: q.MaterialsVersion,
	}
}

func toComparisonResponse(cr *core.ComparisonResult) common.ComparisonResponse {
	items := make([]common.ComparisonItemResponse, len(cr.Items))
	for i, item := range cr.Items {
		items[i] = common.ComparisonItemResponse{
			MaterialID:     item.MaterialID,
			MaterialName:   item.MaterialName,
			Spec:           item.Spec,
			Quantity:       item.Quantity,
			Unit:           item.Unit,
			SupplierPrices: item.SupplierPrices,
			HasNoPrice:     item.HasNoPrice,
		}
	}

	return common.ComparisonResponse{
		InquiryID:   cr.InquiryID,
		Items:       items,
		GeneratedAt: formatTime(cr.GeneratedAt),
	}
}

func toPurchaseOrderResponse(po *core.PurchaseOrder) common.PurchaseOrderResponse {
	items := make([]common.PurchaseOrderItemDTO, len(po.Items))
	for i, item := range po.Items {
		items[i] = common.PurchaseOrderItemDTO{
			MaterialID:   item.MaterialID,
			Name:         item.Name,
			Spec:         item.Spec,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			UnitPrice:    item.UnitPrice,
			Amount:       item.Amount,
			DeliveryDays: item.DeliveryDays,
		}
	}

	return common.PurchaseOrderResponse{
		ID:          po.ID,
		InquiryID:   po.InquiryID,
		SupplierID:  po.SupplierID,
		BuyerID:     po.BuyerID,
		Items:       items,
		TotalAmount: po.TotalAmount,
		CreatedAt:   formatTime(po.CreatedAt),
		Remark:      po.Remark,
	}
}

func toTodoReminderResponse(todo *core.TodoReminder) common.TodoReminderResponse {
	return common.TodoReminderResponse{
		ID:           todo.ID,
		SupplierID:   todo.SupplierID,
		InquiryID:    todo.InquiryID,
		InquiryTitle: todo.InquiryTitle,
		Deadline:     formatTime(todo.Deadline),
		CreatedAt:    formatTime(todo.CreatedAt),
	}
}

func (s *Server) CreateInquiryHandler(w http.ResponseWriter, r *http.Request) {
	var req common.CreateInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	deadline, err := parseTime(req.Deadline)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	inq := &core.Inquiry{
		Title:       req.Title,
		BuyerID:     req.BuyerID,
		Materials:   common.FromMaterialDTOs(req.Materials),
		SupplierIDs: req.SupplierIDs,
		Deadline:    deadline,
	}

	if err := s.store.CreateInquiry(inq); err != nil {
		status := http.StatusInternalServerError
		if err == core.ErrDeadlineTooEarly {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusCreated, toInquiryResponse(inq))
}

func (s *Server) GetInquiryHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing id parameter"))
		return
	}

	inq, err := s.store.GetInquiry(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, toInquiryResponse(inq))
}

func (s *Server) ListInquiriesHandler(w http.ResponseWriter, r *http.Request) {
	inqs := s.store.ListInquiries()
	responses := make([]common.InquiryResponse, len(inqs))
	for i, inq := range inqs {
		responses[i] = toInquiryResponse(inq)
	}
	writeJSON(w, http.StatusOK, responses)
}

func (s *Server) UpdateInquiryMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	var req common.UpdateInquiryMaterialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	materials := common.FromMaterialDTOs(req.Materials)
	if err := s.store.UpdateInquiryMaterials(req.InquiryID, materials); err != nil {
		status := http.StatusInternalServerError
		if err == core.ErrInquiryNotFound {
			status = http.StatusNotFound
		} else if err == core.ErrAlreadyClosed {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}

	inq, _ := s.store.GetInquiry(req.InquiryID)
	writeJSON(w, http.StatusOK, toInquiryResponse(inq))
}

func (s *Server) GetTodoRemindersHandler(w http.ResponseWriter, r *http.Request) {
	supplierID := r.URL.Query().Get("supplier_id")
	if supplierID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing supplier_id parameter"))
		return
	}

	todos := s.store.GetTodoReminders(supplierID)
	responses := make([]common.TodoReminderResponse, len(todos))
	for i, todo := range todos {
		responses[i] = toTodoReminderResponse(todo)
	}
	writeJSON(w, http.StatusOK, responses)
}

func (s *Server) SubmitQuotationHandler(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitQuotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	inq, err := s.store.GetInquiry(req.InquiryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	items := make([]core.QuotationItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = core.QuotationItem{
			MaterialID:   item.MaterialID,
			UnitPrice:    item.UnitPrice,
			DeliveryDays: item.DeliveryDays,
		}
	}

	q := &core.Quotation{
		InquiryID:  req.InquiryID,
		SupplierID: req.SupplierID,
		Items:      items,
		Remark:     req.Remark,
	}

	if err := s.store.SubmitQuotation(q, inq); err != nil {
		status := http.StatusInternalServerError
		if err == core.ErrInquiryExpired || err == core.ErrNotInvitedSupplier || err == core.ErrInvalidMaterials {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusCreated, toQuotationResponse(q))
}

func (s *Server) GetQuotationsHandler(w http.ResponseWriter, r *http.Request) {
	inquiryID := r.URL.Query().Get("inquiry_id")
	if inquiryID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing inquiry_id parameter"))
		return
	}

	quotations, err := s.store.GetQuotationsByInquiry(inquiryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	responses := make([]common.QuotationResponse, len(quotations))
	for i, q := range quotations {
		responses[i] = toQuotationResponse(q)
	}
	writeJSON(w, http.StatusOK, responses)
}

func (s *Server) GenerateComparisonHandler(w http.ResponseWriter, r *http.Request) {
	var req common.ComparisonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	inq, err := s.store.GetInquiry(req.InquiryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	result, err := s.store.GenerateComparison(inq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, toComparisonResponse(result))
}

func (s *Server) AwardAndCreateOrdersHandler(w http.ResponseWriter, r *http.Request) {
	var req common.AwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	inq, err := s.store.GetInquiry(req.InquiryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	if err := s.store.CloseInquiry(req.InquiryID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	awards := make([]core.AwardedMaterial, len(req.Awards))
	for i, a := range req.Awards {
		awards[i] = core.AwardedMaterial{
			MaterialID: a.MaterialID,
			SupplierID: a.SupplierID,
		}
	}

	orders, err := s.store.CreatePurchaseOrders(inq, awards)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	responses := make([]common.PurchaseOrderResponse, len(orders))
	for i, po := range orders {
		responses[i] = toPurchaseOrderResponse(po)
	}

	writeJSON(w, http.StatusCreated, responses)
}

func (s *Server) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if inquiryID := r.URL.Query().Get("inquiry_id"); inquiryID != "" {
		orders := s.store.GetPurchaseOrdersByInquiry(inquiryID)
		responses := make([]common.PurchaseOrderResponse, len(orders))
		for i, po := range orders {
			responses[i] = toPurchaseOrderResponse(po)
		}
		writeJSON(w, http.StatusOK, responses)
		return
	}

	if supplierID := r.URL.Query().Get("supplier_id"); supplierID != "" {
		orders := s.store.GetPurchaseOrdersBySupplier(supplierID)
		responses := make([]common.PurchaseOrderResponse, len(orders))
		for i, po := range orders {
			responses[i] = toPurchaseOrderResponse(po)
		}
		writeJSON(w, http.StatusOK, responses)
		return
	}

	writeError(w, http.StatusBadRequest, fmt.Errorf("missing inquiry_id or supplier_id parameter"))
}
