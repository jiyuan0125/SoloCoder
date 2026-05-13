package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"aftersale-ticket/db"
	"aftersale-ticket/logistics"
	"aftersale-ticket/models"
	"aftersale-ticket/qc"
	"aftersale-ticket/review"
)

type Handler struct {
	store          *db.Store
	reviewEngine   *review.Engine
	logisticsSvc   *logistics.Service
	qcSvc          *qc.Service
}

func NewHandler(store *db.Store) *Handler {
	return &Handler{
		store:        store,
		reviewEngine: review.NewEngine(store),
		logisticsSvc: logistics.NewService(store),
		qcSvc:        qc.NewService(store),
	}
}

type ErrorResponse struct {
	Error       string `json:"error"`
	CurrentStatus string `json:"current_status,omitempty"`
	OverdueDays int    `json:"overdue_days,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, err string, extra ...interface{}) {
	resp := ErrorResponse{Error: err}
	if len(extra) >= 1 {
		if s, ok := extra[0].(string); ok {
			resp.CurrentStatus = s
		}
	}
	if len(extra) >= 2 {
		if d, ok := extra[1].(int); ok {
			resp.OverdueDays = d
		}
	}
	writeJSON(w, status, resp)
}

func generateTicketID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "TK" + hex.EncodeToString(b)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	switch {
	case path == "/tickets" && r.Method == http.MethodPost:
		h.createTicket(w, r)
	case path == "/tickets" && r.Method == http.MethodGet:
		h.listTickets(w, r)
	case strings.HasPrefix(path, "/tickets/"):
		parts := strings.Split(path, "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		ticketID := parts[2]

		switch {
		case len(parts) == 3:
			switch r.Method {
			case http.MethodGet:
				h.getTicket(w, r, ticketID)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case len(parts) >= 4:
			action := parts[3]
			switch {
			case action == "claim" && len(parts) == 4 && r.Method == http.MethodPost:
				h.claimTicket(w, r, ticketID)
			case action == "review" && len(parts) == 4 && r.Method == http.MethodPost:
				h.reviewTicket(w, r, ticketID)
			case action == "logistics" && len(parts) == 4 && r.Method == http.MethodGet:
				h.getLogistics(w, r, ticketID)
			case action == "qc" && len(parts) == 4 && r.Method == http.MethodPost:
				h.performQC(w, r, ticketID)
			case action == "repair" && len(parts) == 4 && r.Method == http.MethodPost:
				h.addRepairRecord(w, r, ticketID)
			case action == "repair" && len(parts) == 4 && r.Method == http.MethodGet:
				h.listRepairRecords(w, r, ticketID)
			case action == "status" && len(parts) == 4 && r.Method == http.MethodPost:
				h.updateStatus(w, r, ticketID)
			default:
				http.NotFound(w, r)
			}
		}
	default:
		http.NotFound(w, r)
	}
}

type CreateTicketRequest struct {
	Type        models.TicketType `json:"type"`
	OrderID     string            `json:"order_id"`
	Reason      string            `json:"reason"`
	Description string            `json:"description,omitempty"`
	CustomerID  string            `json:"customer_id"`
}

func (h *Handler) createTicket(w http.ResponseWriter, r *http.Request) {
	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := models.ValidateOrderID(req.OrderID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.store.GetOrder(req.OrderID)
	if err != nil {
		if err.Error() == "order not found" {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if req.Type != models.TicketTypeReturn && req.Type != models.TicketTypeExchange && req.Type != models.TicketTypeRepair {
		writeError(w, http.StatusBadRequest, "invalid ticket type")
		return
	}

	ticket := &models.Ticket{
		ID:          generateTicketID(),
		Type:        req.Type,
		OrderID:     req.OrderID,
		Status:      models.StatusPendingReview,
		Reason:      req.Reason,
		Description: req.Description,
		CustomerID:  req.CustomerID,
	}

	if err := h.store.CreateTicket(ticket); err != nil {
		if err.Error() == "order not found" {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (h *Handler) listTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.store.ListTickets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (h *Handler) getTicket(w http.ResponseWriter, r *http.Request, ticketID string) {
	ticket, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

type ClaimRequest struct {
	AgentID string `json:"agent_id"`
}

func (h *Handler) claimTicket(w http.ResponseWriter, r *http.Request, ticketID string) {
	var req ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "agent_id is required")
		return
	}

	success, err := h.store.ClaimTicket(ticketID, req.AgentID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !success {
		ticket, _ := h.store.GetTicket(ticketID)
		writeError(w, http.StatusConflict, "ticket already claimed by another agent", string(ticket.Status))
		return
	}

	ticket, _ := h.store.GetTicket(ticketID)
	writeJSON(w, http.StatusOK, ticket)
}

type ReviewRequest struct {
	Approved bool `json:"approved"`
}

func (h *Handler) reviewTicket(w http.ResponseWriter, r *http.Request, ticketID string) {
	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if ticket.Status != models.StatusPendingReview {
		writeError(w, http.StatusBadRequest, "invalid status transition", string(ticket.Status))
		return
	}

	result, err := h.reviewEngine.Review(ticket)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !result.Approved {
		writeError(w, http.StatusBadRequest, result.ErrorMessage, string(ticket.Status), result.OverdueDays)
		return
	}

	if ticket.Type == models.TicketTypeExchange {
		invResult, err := h.reviewEngine.CheckExchangeInventory(ticket)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !invResult.Approved {
			writeError(w, http.StatusConflict, invResult.ErrorMessage)
			return
		}
	}

	if !req.Approved {
		now := ticket.CreatedAt
		_ = now
		h.store.UpdateTicketStatus(ticketID, models.StatusClosed, "review rejected")
		ticket, _ = h.store.GetTicket(ticketID)
		writeJSON(w, http.StatusOK, ticket)
		return
	}

	if ticket.Type == models.TicketTypeReturn {
		_, err := h.logisticsSvc.CreateReturnOrder(ticketID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if ticket.Type == models.TicketTypeExchange {
		order, _ := h.store.GetOrder(ticket.OrderID)
		h.store.DeductInventory(order.ProductSKU, order.Quantity)
		_, err := h.logisticsSvc.CreateExchangeOrder(ticketID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := h.store.UpdateTicketStatus(ticketID, models.StatusProcessing, ""); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ticket, _ = h.store.GetTicket(ticketID)
	writeJSON(w, http.StatusOK, ticket)
}

func (h *Handler) getLogistics(w http.ResponseWriter, r *http.Request, ticketID string) {
	_, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	lo, err := h.logisticsSvc.GetByTicket(ticketID)
	if err != nil {
		if err.Error() == "logistics order not found" {
			writeError(w, http.StatusNotFound, "logistics order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, lo)
}

type QCRequest struct {
	Passed    bool   `json:"passed"`
	Reason    string `json:"reason,omitempty"`
	Inspector string `json:"inspector"`
}

func (h *Handler) performQC(w http.ResponseWriter, r *http.Request, ticketID string) {
	var req QCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Inspector == "" {
		writeError(w, http.StatusBadRequest, "inspector is required")
		return
	}

	ticket, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if ticket.Status != models.StatusProcessing {
		writeError(w, http.StatusBadRequest, "invalid status for QC", string(ticket.Status))
		return
	}

	if ticket.Type != models.TicketTypeReturn {
		writeError(w, http.StatusBadRequest, "QC only applicable for return tickets")
		return
	}

	qcReq := &qc.QCRequest{
		TicketID:  ticketID,
		Passed:    req.Passed,
		Reason:    req.Reason,
		Inspector: req.Inspector,
	}

	_, err = h.qcSvc.PerformQC(qcReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if req.Passed {
		order, _ := h.store.GetOrder(ticket.OrderID)
		h.qcSvc.ProcessRefund(&qc.RefundRequest{
			OrderID: order.ID,
			Amount:  order.TotalAmount,
			Method:  order.PaymentMethod,
		})
		if err := h.store.UpdateTicketStatus(ticketID, models.StatusPendingConfirm, ""); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		closedReason := "QC failed: " + req.Reason
		if err := h.store.UpdateTicketStatus(ticketID, models.StatusClosed, closedReason); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	ticket, _ = h.store.GetTicket(ticketID)
	writeJSON(w, http.StatusOK, ticket)
}

type RepairRecordRequest struct {
	Description string       `json:"description"`
	PartsUsed   []models.Part `json:"parts_used"`
}

func (h *Handler) addRepairRecord(w http.ResponseWriter, r *http.Request, ticketID string) {
	var req RepairRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "description is required")
		return
	}

	ticket, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if ticket.Status != models.StatusProcessing {
		writeError(w, http.StatusBadRequest, "invalid status for repair", string(ticket.Status))
		return
	}

	if ticket.Type != models.TicketTypeRepair {
		writeError(w, http.StatusBadRequest, "repair records only applicable for repair tickets")
		return
	}

	rr := &models.RepairRecord{
		ID:          generateTicketID(),
		TicketID:    ticketID,
		Description: req.Description,
		PartsUsed:   req.PartsUsed,
	}

	if err := h.store.CreateRepairRecord(rr); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, rr)
}

func (h *Handler) listRepairRecords(w http.ResponseWriter, r *http.Request, ticketID string) {
	_, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	records, err := h.store.GetRepairRecords(ticketID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, records)
}

type StatusUpdateRequest struct {
	Status models.TicketStatus `json:"status"`
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request, ticketID string) {
	var req StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := h.store.GetTicket(ticketID)
	if err != nil {
		if err.Error() == "ticket not found" {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	canTransition, err := ticket.CanTransitionTo(req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), string(ticket.Status))
		return
	}

	if !canTransition {
		writeError(w, http.StatusBadRequest, "invalid status transition", string(ticket.Status))
		return
	}

	if err := h.store.UpdateTicketStatus(ticketID, req.Status, ""); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ticket, _ = h.store.GetTicket(ticketID)
	writeJSON(w, http.StatusOK, ticket)
}
