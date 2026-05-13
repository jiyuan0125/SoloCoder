package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type OrderStatus string

const (
	StatusDraft      OrderStatus = "draft"
	StatusApproved   OrderStatus = "approved"
	StatusExecuted   OrderStatus = "executed"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusClosed     OrderStatus = "closed"
)

var statusTransition = map[OrderStatus][]OrderStatus{
	StatusDraft:     {StatusApproved},
	StatusApproved:  {StatusExecuted},
	StatusExecuted:  {StatusConfirmed},
	StatusConfirmed: {StatusClosed},
}

type DiscountType string

const (
	DiscountFullMinus DiscountType = "full_minus"
	DiscountCategory  DiscountType = "category"
	DiscountVIP       DiscountType = "vip"
)

type Product struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Price    int64  `json:"price"`
}

type Discount struct {
	Type           DiscountType `json:"type"`
	FullMinusLimit int64        `json:"full_minus_limit,omitempty"`
	FullMinusAmount int64       `json:"full_minus_amount,omitempty"`
	Category       string       `json:"category,omitempty"`
	CategoryRate   float64      `json:"category_rate,omitempty"`
	VIPLevel       int          `json:"vip_level,omitempty"`
	VIPRate        float64      `json:"vip_rate,omitempty"`
}

type Order struct {
	ID          int64        `json:"id"`
	Products    []Product    `json:"products"`
	Discounts   []Discount   `json:"discounts"`
	Status      OrderStatus  `json:"status"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

type OrderDetail struct {
	ID        int64       `json:"id"`
	OrderID   int64       `json:"order_id"`
	EventType string      `json:"event_type"`
	NewStatus OrderStatus `json:"new_status"`
	OldStatus OrderStatus `json:"old_status,omitempty"`
	Details   string      `json:"details"`
	CreatedAt string      `json:"created_at"`
}

type PriceBreakdown struct {
	ProductID        string `json:"product_id"`
	OriginalPrice    int64  `json:"original_price"`
	AfterCategory    int64  `json:"after_category"`
	AfterVIP         int64  `json:"after_vip"`
}

type DiscountDetail struct {
	OrderID           int64           `json:"order_id"`
	OriginalTotal     int64           `json:"original_total"`
	ProductBreakdowns []PriceBreakdown `json:"product_breakdowns"`
	AfterCategory     int64           `json:"after_category"`
	AfterVIP          int64           `json:"after_vip"`
	FullMinusAmount   int64           `json:"full_minus_amount"`
	FinalPrice        int64           `json:"final_price"`
}

var db *sql.DB
var mu sync.Mutex

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./discount_stack.db")
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		products TEXT NOT NULL,
		discounts TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS order_details (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL,
		event_type TEXT NOT NULL,
		old_status TEXT,
		new_status TEXT NOT NULL,
		details TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (order_id) REFERENCES orders(id)
	)`)
	if err != nil {
		return err
	}

	return nil
}

func roundDown(x float64) int64 {
	return int64(x)
}

func roundHalfUp(x float64) int64 {
	if x < 0 {
		return int64(x - 0.5)
	}
	return int64(x + 0.5)
}

func calculateDiscount(order Order) (*DiscountDetail, error) {
	if countFullMinus(order.Discounts) > 1 {
		return nil, fmt.Errorf("每单限用一张满减券")
	}

	breakdowns := make([]PriceBreakdown, 0, len(order.Products))
	originalTotal := int64(0)
	afterCategoryTotal := int64(0)
	afterVIPTotal := int64(0)

	for _, p := range order.Products {
		originalTotal += p.Price
		price := p.Price

		afterCat := price
		for _, d := range order.Discounts {
			if d.Type == DiscountCategory && d.Category == p.Category && d.CategoryRate > 0 {
				afterCat = roundDown(float64(price) * d.CategoryRate)
			}
		}
		afterCategoryTotal += afterCat

		afterV := afterCat
		for _, d := range order.Discounts {
			if d.Type == DiscountVIP && d.VIPRate > 0 {
				afterV = roundDown(float64(afterCat) * d.VIPRate)
			}
		}
		afterVIPTotal += afterV

		breakdowns = append(breakdowns, PriceBreakdown{
			ProductID:     p.ID,
			OriginalPrice: price,
			AfterCategory: afterCat,
			AfterVIP:      afterV,
		})
	}

	fullMinusAmount := int64(0)
	for _, d := range order.Discounts {
		if d.Type == DiscountFullMinus && afterVIPTotal >= d.FullMinusLimit {
			fullMinusAmount = d.FullMinusAmount
		}
	}

	finalPrice := afterVIPTotal - fullMinusAmount
	if finalPrice < 0 {
		finalPrice = 0
	}

	return &DiscountDetail{
		OriginalTotal:     originalTotal,
		ProductBreakdowns: breakdowns,
		AfterCategory:     afterCategoryTotal,
		AfterVIP:          afterVIPTotal,
		FullMinusAmount:   fullMinusAmount,
		FinalPrice:        finalPrice,
	}, nil
}

func canTransition(from, to OrderStatus) bool {
	allowed, ok := statusTransition[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func countFullMinus(discounts []Discount) int {
	count := 0
	for _, d := range discounts {
		if d.Type == DiscountFullMinus {
			count++
		}
	}
	return count
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseProducts(s string) []Product {
	var products []Product
	json.Unmarshal([]byte(s), &products)
	return products
}

func parseDiscounts(s string) []Discount {
	var discounts []Discount
	json.Unmarshal([]byte(s), &discounts)
	return discounts
}

func scanOrder(rows *sql.Rows) (*Order, error) {
	var id int64
	var productsJSON, discountsJSON, status, createdAt, updatedAt string
	if err := rows.Scan(&id, &productsJSON, &discountsJSON, &status, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return &Order{
		ID:        id,
		Products:  parseProducts(productsJSON),
		Discounts: parseDiscounts(discountsJSON),
		Status:    OrderStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func getOrderByID(id int64) (*Order, error) {
	rows, err := db.Query("SELECT id, products, discounts, status, created_at, updated_at FROM orders WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanOrder(rows)
	}
	return nil, nil
}

func addDetail(orderID int64, eventType string, oldStatus, newStatus OrderStatus, details string) error {
	_, err := db.Exec(
		`INSERT INTO order_details (order_id, event_type, old_status, new_status, details) VALUES (?, ?, ?, ?, ?)`,
		orderID, eventType, oldStatus, newStatus, details,
	)
	return err
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Products  []Product  `json:"products"`
		Discounts []Discount `json:"discounts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Products) == 0 {
		writeError(w, http.StatusBadRequest, "products required")
		return
	}

	if countFullMinus(req.Discounts) > 1 {
		writeError(w, http.StatusBadRequest, "每单限用一张满减券")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	productsJSON, _ := json.Marshal(req.Products)
	discountsJSON, _ := json.Marshal(req.Discounts)

	result, err := tx.Exec(
		`INSERT INTO orders (products, discounts, status) VALUES (?, ?, ?)`,
		string(productsJSON), string(discountsJSON), StatusDraft,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()

	_, err = tx.Exec(
		`INSERT INTO order_details (order_id, event_type, new_status, details) VALUES (?, ?, ?, ?)`,
		id, "create", StatusDraft, "订单创建",
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	order, _ := getOrderByID(id)
	writeJSON(w, http.StatusCreated, order)
}

func listOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, products, discounts, status, created_at, updated_at FROM orders ORDER BY id DESC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		orders = append(orders, *order)
	}

	writeJSON(w, http.StatusOK, orders)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := getOrderByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req struct {
		Products  []Product  `json:"products"`
		Discounts []Discount `json:"discounts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	order, err := getOrderByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	if order.Status != StatusDraft {
		writeError(w, http.StatusBadRequest, "only draft orders can be updated")
		return
	}

	products := order.Products
	if len(req.Products) > 0 {
		products = req.Products
	}
	discounts := order.Discounts
	if req.Discounts != nil {
		discounts = req.Discounts
	}

	fullMinusCount := 0
	for _, d := range discounts {
		if d.Type == DiscountFullMinus {
			fullMinusCount++
		}
	}
	if fullMinusCount > 1 {
		writeError(w, http.StatusBadRequest, "每单限用一张满减券")
		return
	}

	productsJSON, _ := json.Marshal(products)
	discountsJSON, _ := json.Marshal(discounts)

	_, err = db.Exec(
		`UPDATE orders SET products = ?, discounts = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(productsJSON), string(discountsJSON), id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	details := fmt.Sprintf("更新商品: %d 个, 优惠: %d 个", len(products), len(discounts))
	addDetail(id, "update", order.Status, order.Status, details)

	order, _ = getOrderByID(id)
	writeJSON(w, http.StatusOK, order)
}

func listOrderDetails(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/details"), "/api/orders/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := getOrderByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	rows, err := db.Query(
		"SELECT id, order_id, event_type, old_status, new_status, details, created_at FROM order_details WHERE order_id = ? ORDER BY id ASC",
		id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	details := make([]OrderDetail, 0)
	for rows.Next() {
		var d OrderDetail
		var oldStatus sql.NullString
		if err := rows.Scan(&d.ID, &d.OrderID, &d.EventType, &oldStatus, &d.NewStatus, &d.Details, &d.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if oldStatus.Valid {
			d.OldStatus = OrderStatus(oldStatus.String)
		}
		details = append(details, d)
	}

	writeJSON(w, http.StatusOK, details)
}

func advanceOrder(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	action := parts[1]
	var targetStatus OrderStatus
	switch action {
	case "approve":
		targetStatus = StatusApproved
	case "execute":
		targetStatus = StatusExecuted
	case "confirm":
		targetStatus = StatusConfirmed
	case "close":
		targetStatus = StatusClosed
	default:
		writeError(w, http.StatusBadRequest, "invalid action")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	order, err := getOrderByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	if !canTransition(order.Status, targetStatus) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("cannot transition from %s to %s", order.Status, targetStatus))
		return
	}

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE orders SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		targetStatus, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	eventNames := map[string]string{
		"approve": "审批",
		"execute": "执行",
		"confirm": "确认",
		"close":   "关闭",
	}
	eventName := eventNames[action]

	_, err = tx.Exec(
		`INSERT INTO order_details (order_id, event_type, old_status, new_status, details) VALUES (?, ?, ?, ?, ?)`,
		id, action, order.Status, targetStatus, fmt.Sprintf("%s订单", eventName),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	order, _ = getOrderByID(id)
	writeJSON(w, http.StatusOK, order)
}

func getDiscountDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	idStr := strings.TrimSuffix(path, "/price")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := getOrderByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if order == nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	detail, err := calculateDiscount(*order)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	detail.OrderID = id

	writeJSON(w, http.StatusOK, detail)
}

func main() {
	if err := initDB(); err != nil {
		panic(fmt.Sprintf("failed to init db: %v", err))
	}
	defer db.Close()

	http.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createOrder(w, r)
		case http.MethodGet:
			listOrders(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	http.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/details") && r.Method == http.MethodGet {
			listOrderDetails(w, r)
			return
		}

		if strings.HasSuffix(path, "/price") && r.Method == http.MethodGet {
			getDiscountDetail(w, r)
			return
		}

		if strings.Contains(path, "/approve") || strings.Contains(path, "/execute") ||
			strings.Contains(path, "/confirm") || strings.Contains(path, "/close") {
			if r.Method == http.MethodPost {
				advanceOrder(w, r)
				return
			}
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		switch r.Method {
		case http.MethodGet:
			getOrder(w, r)
		case http.MethodPut:
			updateOrder(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	fmt.Println("Server starting on :9601")
	if err := http.ListenAndServe(":9601", nil); err != nil {
		panic(err)
	}
}
