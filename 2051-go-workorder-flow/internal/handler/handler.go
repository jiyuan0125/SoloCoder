package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"workorder-flow/internal/database"
	"workorder-flow/internal/duty"
	"workorder-flow/internal/engine"
	"workorder-flow/internal/models"
	"workorder-flow/internal/notification"
)

type Handler struct {
	db          *database.DB
	flowEngine  *engine.FlowEngine
	scheduler   *duty.Scheduler
	notifier    *notification.NotificationService
}

func NewHandler(db *database.DB, flowEngine *engine.FlowEngine, scheduler *duty.Scheduler, notifier *notification.NotificationService) *Handler {
	return &Handler{
		db:         db,
		flowEngine: flowEngine,
		scheduler:  scheduler,
		notifier:   notifier,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type createWorkOrderRequest struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	SubmitterID int64    `json:"submitter_id"`
	ResourceIDs []int64  `json:"resource_ids"`
}

func (h *Handler) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if req.SubmitterID == 0 {
		writeError(w, http.StatusBadRequest, "submitter_id is required")
		return
	}

	var woType models.WorkOrderType
	typ := strings.TrimSpace(req.Type)
	switch typ {
	case "故障报修":
		woType = models.WorkOrderTypeFaultReport
	case "服务请求":
		woType = models.WorkOrderTypeServiceRequest
	case "投诉":
		woType = models.WorkOrderTypeComplaint
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid work order type: %s", req.Type))
		return
	}

	createReq := &engine.CreateWorkOrderRequest{
		Type:        woType,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		SubmitterID: req.SubmitterID,
		ResourceIDs: req.ResourceIDs,
	}

	wo, err := h.flowEngine.CreateWorkOrder(ctx, createReq)
	if err != nil {
		if errors.Is(err, duty.ErrNoDutyUser) {
			writeError(w, http.StatusServiceUnavailable, "当前无可用值班人员，请稍后重试")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if wo.CurrentHandlerID != nil {
		h.notifier.NotifyAssigned(ctx, wo.ID, *wo.CurrentHandlerID)
	}

	writeJSON(w, http.StatusCreated, wo)
}

func (h *Handler) GetWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	wo, err := h.db.GetWorkOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resourceIDs, err := h.db.GetWorkOrderResources(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type workOrderDetail struct {
		*models.WorkOrder
		ResourceIDs []int64 `json:"resource_ids"`
	}

	writeJSON(w, http.StatusOK, workOrderDetail{WorkOrder: wo, ResourceIDs: resourceIDs})
}

func (h *Handler) ListWorkOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := h.db.ListWorkOrders(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

type claimRequest struct {
	HandlerID int64 `json:"handler_id"`
}

func (h *Handler) ClaimWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/claim")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req claimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.HandlerID == 0 {
		writeError(w, http.StatusBadRequest, "handler_id is required")
		return
	}

	wo, err := h.flowEngine.ClaimWorkOrder(ctx, id, req.HandlerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrAlreadyAssigned) {
			parts := strings.Split(err.Error(), ": current handler is ")
			handlerID := parts[len(parts)-1]
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":              "work order already assigned",
				"current_handler_id": handlerID,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

type advanceRequest struct {
	OperatorID int64 `json:"operator_id"`
}

func (h *Handler) AdvanceWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/advance")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req advanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OperatorID == 0 {
		writeError(w, http.StatusBadRequest, "operator_id is required")
		return
	}

	wo, err := h.flowEngine.AdvanceWorkOrderStatus(ctx, id, req.OperatorID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrNotCurrentHandler) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, engine.ErrEscalatedCannotOperate) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

type confirmRequest struct {
	SubmitterID int64  `json:"submitter_id"`
	Rating      *int   `json:"rating,omitempty"`
}

func (h *Handler) ConfirmWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/confirm")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req confirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SubmitterID == 0 {
		writeError(w, http.StatusBadRequest, "submitter_id is required")
		return
	}

	wo, err := h.flowEngine.ConfirmAndClose(ctx, id, req.SubmitterID, req.Rating)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

type reassignRequest struct {
	CurrentHandlerID int64  `json:"current_handler_id"`
	NewHandlerID     int64  `json:"new_handler_id"`
	Reason           string `json:"reason"`
}

func (h *Handler) ReassignWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/reassign")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req reassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentHandlerID == 0 {
		writeError(w, http.StatusBadRequest, "current_handler_id is required")
		return
	}

	if req.NewHandlerID == 0 {
		writeError(w, http.StatusBadRequest, "new_handler_id is required")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	wo, err := h.flowEngine.ReassignWorkOrder(ctx, id, req.CurrentHandlerID, req.NewHandlerID, req.Reason)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrReasonRequired) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrNotCurrentHandler) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidHandler) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.notifier.NotifyAssigned(ctx, id, req.NewHandlerID)

	writeJSON(w, http.StatusOK, wo)
}

type escalateRequest struct {
	OperatorID int64  `json:"operator_id"`
	Reason     string `json:"reason"`
}

func (h *Handler) EscalateWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/escalate")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req escalateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	wo, err := h.flowEngine.EscalateWorkOrder(ctx, id, req.OperatorID, req.Reason)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrReasonRequired) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrAlreadyEscalated) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if wo.CurrentHandlerID != nil {
		h.notifier.NotifyEscalated(ctx, id, 0, *wo.CurrentHandlerID, req.Reason)
	}

	writeJSON(w, http.StatusOK, wo)
}

type processRequest struct {
	OperatorID int64  `json:"operator_id"`
	Comment    string `json:"comment,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (h *Handler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/submit-for-review")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OperatorID == 0 {
		writeError(w, http.StatusBadRequest, "operator_id is required")
		return
	}

	wo, err := h.flowEngine.SubmitForReview(ctx, id, req.OperatorID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

func (h *Handler) ApproveWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/approve")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OperatorID == 0 {
		writeError(w, http.StatusBadRequest, "operator_id is required")
		return
	}

	wo, err := h.flowEngine.ApproveWorkOrder(ctx, id, req.OperatorID, req.Comment)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

func (h *Handler) RejectWorkOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/reject")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OperatorID == 0 {
		writeError(w, http.StatusBadRequest, "operator_id is required")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	wo, err := h.flowEngine.RejectWorkOrder(ctx, id, req.OperatorID, req.Reason)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrReasonRequired) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

func (h *Handler) StartExecution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/start-execution")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OperatorID == 0 {
		writeError(w, http.StatusBadRequest, "operator_id is required")
		return
	}

	wo, err := h.flowEngine.StartExecution(ctx, id, req.OperatorID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrNotCurrentHandler) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, engine.ErrEscalatedCannotOperate) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

func (h *Handler) CompleteExecution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/complete-execution")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OperatorID == 0 {
		writeError(w, http.StatusBadRequest, "operator_id is required")
		return
	}

	wo, err := h.flowEngine.CompleteExecution(ctx, id, req.OperatorID, req.Comment)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, engine.ErrWorkOrderClosed) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrInvalidStateTransition) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, engine.ErrNotCurrentHandler) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, engine.ErrEscalatedCannotOperate) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, wo)
}

func (h *Handler) GetWorkOrderHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/history")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	_, err = h.db.GetWorkOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	history, err := h.db.GetWorkOrderHistory(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, history)
}

type communicationRequest struct {
	SenderID int64  `json:"sender_id"`
	Content  string `json:"content"`
}

func (h *Handler) AddCommunication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/communicate")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	var req communicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SenderID == 0 {
		writeError(w, http.StatusBadRequest, "sender_id is required")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	if err := h.flowEngine.AddCommunication(ctx, id, req.SenderID, req.Content); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) GetWorkOrderCommunications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := strings.TrimPrefix(r.URL.Path, "/workorders/")
	idStr = strings.TrimSuffix(idStr, "/communications")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order id")
		return
	}

	_, err = h.db.GetWorkOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	records, err := h.db.GetWorkOrderCommunications(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, records)
}

func (h *Handler) ListResourceTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rts, err := h.db.ListResourceTypes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, rts)
}

func (h *Handler) ListResourcesByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rtIDStr := strings.TrimPrefix(r.URL.Path, "/resource-types/")
	rtIDStr = strings.TrimSuffix(rtIDStr, "/resources")
	rtID, err := strconv.ParseInt(rtIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid resource type id")
		return
	}

	_, err = h.db.GetResourceTypeByID(ctx, rtID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusBadRequest, "resource type not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resources, err := h.db.ListResourcesByType(ctx, rtID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resources)
}

func (h *Handler) GetResourceWorkOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resourceIDStr := strings.TrimPrefix(r.URL.Path, "/resources/")
	resourceIDStr = strings.TrimSuffix(resourceIDStr, "/workorders")
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid resource id")
		return
	}

	_, err = h.db.GetResourceByID(ctx, resourceID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusBadRequest, "resource not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	orders, err := h.db.ListWorkOrdersByResource(ctx, resourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type resourceSummary struct {
		ResourceID   int64                `json:"resource_id"`
		WorkOrderIDs []int64              `json:"work_order_ids"`
		WorkOrders   []models.WorkOrder   `json:"work_orders"`
		Total        int                  `json:"total"`
	}

	ids := make([]int64, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
	}

	writeJSON(w, http.StatusOK, resourceSummary{
		ResourceID:   resourceID,
		WorkOrderIDs: ids,
		WorkOrders:   orders,
		Total:        len(orders),
	})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"note": "see seeded users in database"})
}

func (h *Handler) StartOverdueChecker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.flowEngine.ProcessOverdueWorkOrders(ctx); err != nil {
				fmt.Printf("error processing overdue work orders: %v\n", err)
			}
		}
	}
}
