package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"event-ticketing/pkg/common"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := common.ValidateCreateEventRequest(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	tiers := make([]common.Tier, len(req.Tiers))
	for i, t := range req.Tiers {
		tiers[i] = common.Tier{
			Name:      t.Name,
			Price:     t.Price,
			Capacity:  t.Capacity,
			Available: t.Capacity,
		}
	}

	event := &common.Event{
		ID:        common.GenerateEventID(),
		Name:      req.Name,
		Time:      req.Time,
		Location:  req.Location,
		Tiers:     tiers,
		CreatedAt: time.Now(),
	}

	if err := h.store.CreateEvent(event); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.CreateEventResponse{
		Success: true,
		EventID: event.ID,
	})
}

func (h *Handler) PurchaseTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.PurchaseTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := common.ValidatePurchaseRequest(req.Quantity); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ticketNumbers, err := h.store.PurchaseTickets(req.EventID, req.TierName, req.Quantity)
	if err != nil {
		switch err {
		case common.ErrEventNotFound:
			h.sendError(w, err.Error(), http.StatusNotFound)
		case common.ErrEventEnded, common.ErrInvalidTier, common.ErrInsufficientStock:
			h.sendError(w, err.Error(), http.StatusBadRequest)
		default:
			h.sendError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.PurchaseTicketResponse{
		Success:       true,
		TicketNumbers: ticketNumbers,
	})
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TicketNumber == "" {
		h.sendError(w, "票号不能为空", http.StatusBadRequest)
		return
	}

	if err := h.store.CheckIn(req.TicketNumber); err != nil {
		switch err {
		case common.ErrTicketNotFound:
			h.sendError(w, err.Error(), http.StatusNotFound)
		case common.ErrTicketAlreadyUsed, common.ErrTicketRefunded:
			h.sendError(w, err.Error(), http.StatusBadRequest)
		default:
			h.sendError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.CheckInResponse{
		Success: true,
		Message: "签到成功",
	})
}

func (h *Handler) RefundTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RefundTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TicketNumber == "" {
		h.sendError(w, "票号不能为空", http.StatusBadRequest)
		return
	}

	if err := h.store.RefundTicket(req.TicketNumber); err != nil {
		switch err {
		case common.ErrTicketNotFound, common.ErrEventNotFound:
			h.sendError(w, err.Error(), http.StatusNotFound)
		case common.ErrTicketAlreadyUsed, common.ErrTicketRefunded:
			h.sendError(w, err.Error(), http.StatusBadRequest)
		default:
			h.sendError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.RefundTicketResponse{
		Success: true,
		Message: "退票成功",
	})
}

func (h *Handler) GetEventStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventID := strings.TrimPrefix(r.URL.Path, "/events/stats/")
	if eventID == "" {
		h.sendError(w, "活动ID不能为空", http.StatusBadRequest)
		return
	}

	stats, err := h.store.GetEventStats(eventID)
	if err != nil {
		if err == common.ErrEventNotFound {
			h.sendError(w, err.Error(), http.StatusNotFound)
		} else {
			h.sendError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetEventStatsResponse{
		Success: true,
		Stats:   stats,
	})
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events := h.store.ListEvents()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ListEventsResponse{
		Success: true,
		Events:  events,
	})
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
