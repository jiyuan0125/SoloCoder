package handler

import (
	"encoding/json"
	"net/http"
	"order-matching/internal/engine"
	"order-matching/internal/model"
	"strings"
)

type Handler struct {
	engine *engine.MatchingEngine
}

func NewHandler(engine *engine.MatchingEngine) *Handler {
	return &Handler{engine: engine}
}

type SubmitOrderRequest struct {
	UserID   string  `json:"user_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
}

type SubmitOrderResponse struct {
	OrderID int64          `json:"order_id"`
	Status  model.OrderStatus `json:"status"`
	Trades []TradeResponse `json:"trades,omitempty"`
}

type TradeResponse struct {
	TradeID   int64   `json:"trade_id"`
	BuyerID   string  `json:"buyer_id"`
	SellerID  string  `json:"seller_id"`
	Price     float64 `json:"price"`
	Quantity  int64   `json:"quantity"`
}

type CancelOrderRequest struct {
	OrderID int64 `json:"order_id"`
}

type OrderBookResponse struct {
	Bids []OrderBookEntryResponse `json:"bids"`
	Asks []OrderBookEntryResponse `json:"asks"`
}

type OrderBookEntryResponse struct {
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
}

func (h *Handler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Price <= 0 {
		http.Error(w, "Price must be positive", http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be positive", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Symbol) == "" {
		http.Error(w, "Symbol cannot be empty", http.StatusBadRequest)
		return
	}

	var side model.OrderSide
	switch strings.ToLower(req.Side) {
	case "buy":
		side = model.Buy
	case "sell":
		side = model.Sell
	default:
		http.Error(w, "Invalid side", http.StatusBadRequest)
		return
	}

	order := &model.Order{
		UserID:   req.UserID,
		Symbol:   req.Symbol,
		Side:     side,
		Price:    req.Price,
		Quantity: req.Quantity,
	}

	trades, err := h.engine.SubmitOrder(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tradeResponses := make([]TradeResponse, 0, len(trades))
	for _, t := range trades {
		tradeResponses = append(tradeResponses, TradeResponse{
			TradeID:  t.ID,
			BuyerID:  t.BuyerID,
			SellerID: t.SellerID,
			Price:    t.Price,
			Quantity: t.Quantity,
		})
	}

	resp := SubmitOrderResponse{
		OrderID: order.ID,
		Status:  order.Status,
		Trades: tradeResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.engine.CancelOrder(req.OrderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	symbol := r.URL.Query().Get("symbol")

	bids, asks, err := h.engine.GetOrderBook(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := OrderBookResponse{
		Bids: make([]OrderBookEntryResponse, 0, 5),
		Asks: make([]OrderBookEntryResponse, 0, 5),
	}

	for _, b := range bids {
		resp.Bids = append(resp.Bids, OrderBookEntryResponse{
			Price:    b.Price,
			Quantity: b.Quantity,
		})
	}

	for _, a := range asks {
		resp.Asks = append(resp.Asks, OrderBookEntryResponse{
			Price:    a.Price,
			Quantity: a.Quantity,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
