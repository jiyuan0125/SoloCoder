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

type AfterSalesHandler struct {
	db *sql.DB
}

type CreateAfterSalesRequest struct {
	OrderID    int    `json:"order_id"`
	ResourceID *int   `json:"resource_id,omitempty"`
	Type       string `json:"type"`
	Reason     string `json:"reason"`
}

type AfterSalesStatusFlowRequest struct {
	AfterSalesID int    `json:"aftersales_id"`
	ToStatus     string `json:"to_status"`
	Operator     string `json:"operator"`
	Reason       string `json:"reason"`
	ResourceID   *int   `json:"resource_id,omitempty"`
}

func NewAfterSalesHandler(db *sql.DB) *AfterSalesHandler {
	return &AfterSalesHandler{db: db}
}

func (h *AfterSalesHandler) HandleAfterSales(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAfterSales(w, r)
	case http.MethodPost:
		h.createAfterSales(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AfterSalesHandler) listAfterSales(w http.ResponseWriter, r *http.Request) {
	afterSalesType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	query := `
		SELECT id, order_id, resource_id, type, current_status, created_at, updated_at
		FROM aftersales
		WHERE 1=1
	`
	args := []interface{}{}

	if afterSalesType != "" {
		query += " AND type = ?"
		args = append(args, afterSalesType)
	}
	if status != "" {
		query += " AND current_status = ?"
		args = append(args, status)
	}
	query += " ORDER BY id DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	afterSalesList := []*models.AfterSales{}
	for rows.Next() {
		as := &models.AfterSales{}
		var resourceID sql.NullInt64
		err := rows.Scan(
			&as.ID, &as.OrderID, &resourceID, &as.Type,
			&as.CurrentStatus, &as.CreatedAt, &as.UpdatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resourceID.Valid {
			rid := int(resourceID.Int64)
			as.ResourceID = &rid
		}
		as.StatusName = models.GetStatusName(as.CurrentStatus)
		afterSalesList = append(afterSalesList, as)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(afterSalesList)
}

func (h *AfterSalesHandler) createAfterSales(w http.ResponseWriter, r *http.Request) {
	var req CreateAfterSalesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.OrderID == 0 {
		http.Error(w, "订单ID不能为空", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "售后类型不能为空", http.StatusBadRequest)
		return
	}

	initialStatus, err := statemachine.GetAfterSalesInitialStatus(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO aftersales (order_id, resource_id, type, current_status, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, req.OrderID, req.ResourceID, req.Type, initialStatus, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	_, err = tx.Exec(`
		INSERT INTO aftersales_history (aftersales_id, from_status, to_status, operator, reason, resource_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, "", initialStatus, "system", req.Reason, req.ResourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	afterSales := &models.AfterSales{
		ID:            int(id),
		OrderID:       req.OrderID,
		ResourceID:    req.ResourceID,
		Type:          req.Type,
		CurrentStatus: initialStatus,
		StatusName:    models.GetStatusName(initialStatus),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(afterSales)
}

func (h *AfterSalesHandler) HandleAfterSalesByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/aftersales/")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	if len(parts) == 1 && parts[0] != "" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.getAfterSalesByID(w, parts[0])
	} else if len(parts) == 2 && parts[1] == "history" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.getAfterSalesHistory(w, parts[0])
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *AfterSalesHandler) getAfterSalesByID(w http.ResponseWriter, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的售后ID", http.StatusBadRequest)
		return
	}

	as, err := h.getAfterSales(id)
	if err == sql.ErrNoRows {
		http.Error(w, "售后记录不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(as)
}

func (h *AfterSalesHandler) getAfterSalesHistory(w http.ResponseWriter, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的售后ID", http.StatusBadRequest)
		return
	}

	_, err = h.getAfterSales(id)
	if err == sql.ErrNoRows {
		http.Error(w, "售后记录不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, aftersales_id, from_status, to_status, operator, reason, resource_id, created_at
		FROM aftersales_history
		WHERE aftersales_id = ?
		ORDER BY id ASC
	`, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := []*models.AfterSalesHistory{}
	for rows.Next() {
		h := &models.AfterSalesHistory{}
		var resourceID sql.NullInt64
		err := rows.Scan(
			&h.ID, &h.AfterSalesID, &h.FromStatus, &h.ToStatus,
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

func (h *AfterSalesHandler) getAfterSales(id int) (*models.AfterSales, error) {
	as := &models.AfterSales{}
	var resourceID sql.NullInt64

	err := h.db.QueryRow(`
		SELECT id, order_id, resource_id, type, current_status, created_at, updated_at
		FROM aftersales WHERE id = ?
	`, id).Scan(
		&as.ID, &as.OrderID, &resourceID, &as.Type,
		&as.CurrentStatus, &as.CreatedAt, &as.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if resourceID.Valid {
		rid := int(resourceID.Int64)
		as.ResourceID = &rid
	}
	as.StatusName = models.GetStatusName(as.CurrentStatus)
	return as, nil
}

func (h *AfterSalesHandler) HandleAfterSalesStatusFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AfterSalesStatusFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Operator == "" {
		http.Error(w, "操作人不能为空", http.StatusBadRequest)
		return
	}

	as, err := h.getAfterSales(req.AfterSalesID)
	if err == sql.ErrNoRows {
		http.Error(w, "售后记录不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := statemachine.ValidateAfterSalesStatusTransition(as.Type, as.CurrentStatus, req.ToStatus); err != nil {
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
		INSERT INTO aftersales_history (aftersales_id, from_status, to_status, operator, reason, resource_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, as.ID, as.CurrentStatus, req.ToStatus, req.Operator, req.Reason, req.ResourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		UPDATE aftersales
		SET current_status = ?, updated_at = ?
		WHERE id = ?
	`, req.ToStatus, time.Now(), as.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedAS, _ := h.getAfterSales(as.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedAS)
}
