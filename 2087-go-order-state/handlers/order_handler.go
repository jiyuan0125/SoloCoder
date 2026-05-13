package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"order-state-machine/models"
	"order-state-machine/statemachine"
	"strconv"
	"strings"
	"time"
)

type OrderHandler struct {
	db *sql.DB
}

type CreateOrderRequest struct {
	OrderNo    string  `json:"order_no"`
	ResourceID *int    `json:"resource_id,omitempty"`
	Amount     float64 `json:"amount"`
}

type StatusFlowRequest struct {
	OrderID    int    `json:"order_id"`
	ToStatus   string `json:"to_status"`
	Operator   string `json:"operator"`
	Reason     string `json:"reason"`
	ResourceID *int   `json:"resource_id,omitempty"`
}

type CancelOrderRequest struct {
	Operator   string `json:"operator"`
	Reason     string `json:"reason"`
	ResourceID *int   `json:"resource_id,omitempty"`
}

func NewOrderHandler(db *sql.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

func (h *OrderHandler) HandleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listOrders(w, r)
	case http.MethodPost:
		h.createOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OrderHandler) listOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = h.db.Query(`
			SELECT id, order_no, resource_id, amount, current_status, is_cancelled, 
			       cancellation_type, created_at, updated_at
			FROM orders
			WHERE current_status = ?
			ORDER BY id DESC
		`, status)
	} else {
		rows, err = h.db.Query(`
			SELECT id, order_no, resource_id, amount, current_status, is_cancelled, 
			       cancellation_type, created_at, updated_at
			FROM orders
			ORDER BY id DESC
		`)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	orders := []*models.Order{}
	for rows.Next() {
		order := &models.Order{}
		var resourceID sql.NullInt64
		var cancellationType sql.NullString
		err := rows.Scan(
			&order.ID, &order.OrderNo, &resourceID, &order.Amount,
			&order.CurrentStatus, &order.IsCancelled, &cancellationType,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resourceID.Valid {
			id := int(resourceID.Int64)
			order.ResourceID = &id
		}
		if cancellationType.Valid {
			order.CancellationType = &cancellationType.String
		}
		order.StatusName = models.GetStatusName(order.CurrentStatus)
		orders = append(orders, order)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.OrderNo == "" {
		http.Error(w, "订单号不能为空", http.StatusBadRequest)
		return
	}

	initialStatus := models.OrderStatusPendingPayment

	result, err := h.db.Exec(`
		INSERT INTO orders (order_no, resource_id, amount, current_status, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, req.OrderNo, req.ResourceID, req.Amount, initialStatus, time.Now())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	order := &models.Order{
		ID:            int(id),
		OrderNo:       req.OrderNo,
		ResourceID:    req.ResourceID,
		Amount:        req.Amount,
		CurrentStatus: initialStatus,
		StatusName:    models.GetStatusName(initialStatus),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) HandleOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的订单ID", http.StatusBadRequest)
		return
	}

	order, err := h.getOrderByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, "订单不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) getOrderByID(id int) (*models.Order, error) {
	order := &models.Order{}
	var resourceID sql.NullInt64
	var cancellationType sql.NullString

	err := h.db.QueryRow(`
		SELECT id, order_no, resource_id, amount, current_status, is_cancelled, 
		       cancellation_type, created_at, updated_at
		FROM orders WHERE id = ?
	`, id).Scan(
		&order.ID, &order.OrderNo, &resourceID, &order.Amount,
		&order.CurrentStatus, &order.IsCancelled, &cancellationType,
		&order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if resourceID.Valid {
		rid := int(resourceID.Int64)
		order.ResourceID = &rid
	}
	if cancellationType.Valid {
		order.CancellationType = &cancellationType.String
	}
	order.StatusName = models.GetStatusName(order.CurrentStatus)
	return order, nil
}

func (h *OrderHandler) HandleOrderHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/history/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的订单ID", http.StatusBadRequest)
		return
	}

	_, err = h.getOrderByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, "订单不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, order_id, from_status, to_status, operator, reason, resource_id, created_at
		FROM order_history
		WHERE order_id = ?
		ORDER BY id ASC
	`, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := []*models.OrderHistory{}
	for rows.Next() {
		h := &models.OrderHistory{}
		var resourceID sql.NullInt64
		err := rows.Scan(
			&h.ID, &h.OrderID, &h.FromStatus, &h.ToStatus,
			&h.Operator, &h.Reason, &resourceID, &h.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resourceID.Valid {
			rid := int(resourceID.Int64)
			h.ResourceID = &rid
		}
		h.FromName = models.GetStatusName(h.FromStatus)
		h.ToName = models.GetStatusName(h.ToStatus)
		history = append(history, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *OrderHandler) HandleOrderStatusFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StatusFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Operator == "" {
		http.Error(w, "操作人不能为空", http.StatusBadRequest)
		return
	}

	order, err := h.getOrderByID(req.OrderID)
	if err == sql.ErrNoRows {
		http.Error(w, "订单不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := statemachine.ValidateOrderStatusTransition(order.CurrentStatus, req.ToStatus); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO order_history (order_id, from_status, to_status, operator, reason, resource_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, order.ID, order.CurrentStatus, req.ToStatus, req.Operator, req.Reason, req.ResourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		UPDATE orders
		SET current_status = ?, updated_at = ?
		WHERE id = ?
	`, req.ToStatus, time.Now(), order.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedOrder, _ := h.getOrderByID(order.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}

func (h *OrderHandler) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/cancel/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的订单ID", http.StatusBadRequest)
		return
	}

	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Operator == "" {
		http.Error(w, "操作人不能为空", http.StatusBadRequest)
		return
	}

	order, err := h.getOrderByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, "订单不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order.IsCancelled {
		http.Error(w, "订单已取消，无法再次取消", http.StatusBadRequest)
		return
	}

	if order.CurrentStatus == models.OrderStatusCompleted {
		http.Error(w, "已完成的订单不能取消", http.StatusBadRequest)
		return
	}

	var cancellationType string
	switch order.CurrentStatus {
	case models.OrderStatusPendingPayment:
		cancellationType = models.CancellationTypeDirect
	case models.OrderStatusPaid:
		cancellationType = models.CancellationTypeRefund
	case models.OrderStatusPreparing, models.OrderStatusShipped:
		cancellationType = models.CancellationTypeReturn
	case models.OrderStatusSigned:
		cancellationType = models.CancellationTypeAftersale
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO order_history (order_id, from_status, to_status, operator, reason, resource_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, order.ID, order.CurrentStatus, models.OrderStatusCancelled, req.Operator, req.Reason, req.ResourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		UPDATE orders
		SET current_status = ?, is_cancelled = 1, cancellation_type = ?, updated_at = ?
		WHERE id = ?
	`, models.OrderStatusCancelled, cancellationType, time.Now(), order.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedOrder, _ := h.getOrderByID(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order":             updatedOrder,
		"cancellation_type": cancellationType,
	})
}
