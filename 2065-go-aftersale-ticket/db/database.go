package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"aftersale-ticket/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_fk=1")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &Store{db: db}
	if err := store.initSchema(); err != nil {
		return nil, err
	}

	if err := store.seedData(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		product_sku TEXT NOT NULL,
		product_name TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		total_amount REAL NOT NULL,
		payment_method TEXT NOT NULL,
		completed_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tickets (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		order_id TEXT NOT NULL,
		status TEXT NOT NULL,
		reason TEXT NOT NULL,
		description TEXT,
		customer_id TEXT NOT NULL,
		assigned_to TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		closed_reason TEXT,
		closed_at DATETIME,
		FOREIGN KEY (order_id) REFERENCES orders(id)
	);

	CREATE TABLE IF NOT EXISTS logistics_orders (
		id TEXT PRIMARY KEY,
		ticket_id TEXT NOT NULL,
		type TEXT NOT NULL,
		carrier TEXT NOT NULL,
		tracking_no TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id)
	);

	CREATE TABLE IF NOT EXISTS qc_results (
		id TEXT PRIMARY KEY,
		ticket_id TEXT NOT NULL,
		passed INTEGER NOT NULL,
		reason TEXT,
		inspector TEXT NOT NULL,
		inspected_at DATETIME NOT NULL,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id)
	);

	CREATE TABLE IF NOT EXISTS repair_records (
		id TEXT PRIMARY KEY,
		ticket_id TEXT NOT NULL,
		description TEXT NOT NULL,
		parts_used TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id)
	);

	CREATE TABLE IF NOT EXISTS inventory (
		sku TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		stock INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS ticket_locks (
		ticket_id TEXT PRIMARY KEY,
		assigned_to TEXT NOT NULL,
		locked_at DATETIME NOT NULL,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id)
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) seedData() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	completed1 := time.Now().AddDate(0, 0, -3)
	completed2 := time.Now().AddDate(0, 0, -10)
	completed3 := time.Now().AddDate(0, 0, -100)

	orders := []struct {
		ID          string
		CustomerID  string
		ProductSKU  string
		ProductName string
		Quantity    int
		TotalAmount float64
		PaymentMethod string
		CompletedAt time.Time
	}{
		{"ORD0001", "CUST001", "PROD001", "蓝牙耳机", 1, 299.00, "alipay", completed1},
		{"ORD0002", "CUST002", "PROD002", "智能手表", 1, 999.00, "wechat", completed2},
		{"ORD0003", "CUST003", "PROD003", "手机屏幕", 1, 599.00, "alipay", completed3},
	}

	for _, o := range orders {
		_, err := tx.Exec(`
			INSERT INTO orders (id, customer_id, product_sku, product_name, quantity, total_amount, payment_method, completed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, o.ID, o.CustomerID, o.ProductSKU, o.ProductName, o.Quantity, o.TotalAmount, o.PaymentMethod, o.CompletedAt)
		if err != nil {
			return err
		}
	}

	inventory := []struct {
		SKU   string
		Name  string
		Stock int
	}{
		{"PROD001", "蓝牙耳机", 50},
		{"PROD002", "智能手表", 30},
		{"PROD003", "手机屏幕", 20},
		{"PART001", "电池", 100},
		{"PART002", "充电接口", 80},
		{"PART003", "扬声器", 60},
	}

	for _, inv := range inventory {
		_, err := tx.Exec(`
			INSERT INTO inventory (sku, name, stock)
			VALUES (?, ?, ?)
		`, inv.SKU, inv.Name, inv.Stock)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) GetOrder(orderID string) (*models.Order, error) {
	var order models.Order
	err := s.db.QueryRow(`
		SELECT id, customer_id, product_sku, product_name, quantity, total_amount, payment_method, completed_at
		FROM orders WHERE id = ?
	`, orderID).Scan(&order.ID, &order.CustomerID, &order.ProductSKU, &order.ProductName, &order.Quantity, &order.TotalAmount, &order.PaymentMethod, &order.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("order not found")
	}
	return &order, err
}

func (s *Store) CreateTicket(ticket *models.Ticket) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM orders WHERE id = ?", ticket.OrderID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("order not found")
	}

	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO tickets (id, type, order_id, status, reason, description, customer_id, assigned_to, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ticket.ID, ticket.Type, ticket.OrderID, ticket.Status, ticket.Reason, ticket.Description, ticket.CustomerID, ticket.AssignedTo, now, now)
	if err != nil {
		return err
	}

	ticket.CreatedAt = now
	ticket.UpdatedAt = now
	return tx.Commit()
}

func (s *Store) GetTicket(ticketID string) (*models.Ticket, error) {
	var ticket models.Ticket
	var closedReason sql.NullString
	var closedAt sql.NullTime
	var assignedTo sql.NullString
	var description sql.NullString

	err := s.db.QueryRow(`
		SELECT id, type, order_id, status, reason, description, customer_id, assigned_to, created_at, updated_at, closed_reason, closed_at
		FROM tickets WHERE id = ?
	`, ticketID).Scan(&ticket.ID, &ticket.Type, &ticket.OrderID, &ticket.Status, &ticket.Reason, &description, &ticket.CustomerID, &assignedTo, &ticket.CreatedAt, &ticket.UpdatedAt, &closedReason, &closedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("ticket not found")
	}

	if closedReason.Valid {
		ticket.ClosedReason = closedReason.String
	}
	if closedAt.Valid {
		ticket.ClosedAt = &closedAt.Time
	}
	if assignedTo.Valid {
		ticket.AssignedTo = assignedTo.String
	}
	if description.Valid {
		ticket.Description = description.String
	}

	return &ticket, err
}

func (s *Store) ListTickets() ([]*models.Ticket, error) {
	rows, err := s.db.Query(`
		SELECT id, type, order_id, status, reason, description, customer_id, assigned_to, created_at, updated_at, closed_reason, closed_at
		FROM tickets ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var closedReason sql.NullString
		var closedAt sql.NullTime
		var assignedTo sql.NullString
		var description sql.NullString
		err := rows.Scan(&ticket.ID, &ticket.Type, &ticket.OrderID, &ticket.Status, &ticket.Reason, &description, &ticket.CustomerID, &assignedTo, &ticket.CreatedAt, &ticket.UpdatedAt, &closedReason, &closedAt)
		if err != nil {
			return nil, err
		}
		if closedReason.Valid {
			ticket.ClosedReason = closedReason.String
		}
		if closedAt.Valid {
			ticket.ClosedAt = &closedAt.Time
		}
		if assignedTo.Valid {
			ticket.AssignedTo = assignedTo.String
		}
		if description.Valid {
			ticket.Description = description.String
		}
		tickets = append(tickets, &ticket)
	}
	return tickets, rows.Err()
}

func (s *Store) UpdateTicketStatus(ticketID string, status models.TicketStatus, closedReason string) error {
	now := time.Now()

	if status == models.StatusClosed {
		_, err := s.db.Exec(`
			UPDATE tickets SET status = ?, updated_at = ?, closed_reason = ?, closed_at = ?
			WHERE id = ?
		`, status, now, closedReason, now, ticketID)
		return err
	}

	_, err := s.db.Exec(`
		UPDATE tickets SET status = ?, updated_at = ?
		WHERE id = ?
	`, status, now, ticketID)
	return err
}

func (s *Store) ClaimTicket(ticketID, agentID string) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var existingAgent sql.NullString
	err = tx.QueryRow("SELECT assigned_to FROM tickets WHERE id = ?", ticketID).Scan(&existingAgent)
	if err == sql.ErrNoRows {
		return false, errors.New("ticket not found")
	}
	if err != nil {
		return false, err
	}

	if existingAgent.Valid && existingAgent.String != agentID {
		return false, nil
	}

	now := time.Now()
	_, err = tx.Exec("UPDATE tickets SET assigned_to = ?, updated_at = ? WHERE id = ?", agentID, now, ticketID)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO ticket_locks (ticket_id, assigned_to, locked_at)
		VALUES (?, ?, ?)
	`, ticketID, agentID, now)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (s *Store) GetInventory(sku string) (*models.InventoryItem, error) {
	var item models.InventoryItem
	err := s.db.QueryRow(`
		SELECT sku, name, stock FROM inventory WHERE sku = ?
	`, sku).Scan(&item.SKU, &item.Name, &item.Stock)
	if err == sql.ErrNoRows {
		return nil, errors.New("inventory item not found")
	}
	return &item, err
}

func (s *Store) DeductInventory(sku string, quantity int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStock int
	err = tx.QueryRow("SELECT stock FROM inventory WHERE sku = ?", sku).Scan(&currentStock)
	if err == sql.ErrNoRows {
		return errors.New("inventory item not found")
	}
	if err != nil {
		return err
	}

	if currentStock < quantity {
		return fmt.Errorf("insufficient stock: available %d, requested %d", currentStock, quantity)
	}

	_, err = tx.Exec("UPDATE inventory SET stock = stock - ? WHERE sku = ?", quantity, sku)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) CreateLogisticsOrder(lo *models.LogisticsOrder) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO logistics_orders (id, ticket_id, type, carrier, tracking_no, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, lo.ID, lo.TicketID, lo.Type, lo.Carrier, lo.TrackingNo, lo.Status, now)
	if err != nil {
		return err
	}
	lo.CreatedAt = now
	return nil
}

func (s *Store) GetLogisticsByTicket(ticketID string) (*models.LogisticsOrder, error) {
	var lo models.LogisticsOrder
	err := s.db.QueryRow(`
		SELECT id, ticket_id, type, carrier, tracking_no, status, created_at
		FROM logistics_orders WHERE ticket_id = ?
	`, ticketID).Scan(&lo.ID, &lo.TicketID, &lo.Type, &lo.Carrier, &lo.TrackingNo, &lo.Status, &lo.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("logistics order not found")
	}
	return &lo, err
}

func (s *Store) CreateQCResult(qc *models.QCResult) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO qc_results (id, ticket_id, passed, reason, inspector, inspected_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, qc.ID, qc.TicketID, qc.Passed, qc.Reason, qc.Inspector, now)
	if err != nil {
		return err
	}
	qc.InspectedAt = now
	return nil
}

func (s *Store) CreateRepairRecord(rr *models.RepairRecord) error {
	partsJSON, err := json.Marshal(rr.PartsUsed)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, part := range rr.PartsUsed {
		var stock int
		err = tx.QueryRow("SELECT stock FROM inventory WHERE sku = ?", part.SKU).Scan(&stock)
		if err == sql.ErrNoRows {
			return fmt.Errorf("part not found: %s", part.SKU)
		}
		if stock < part.Quantity {
			return fmt.Errorf("insufficient part stock for %s: available %d", part.SKU, stock)
		}
	}

	for _, part := range rr.PartsUsed {
		_, err := tx.Exec("UPDATE inventory SET stock = stock - ? WHERE sku = ?", part.Quantity, part.SKU)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO repair_records (id, ticket_id, description, parts_used, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, rr.ID, rr.TicketID, rr.Description, string(partsJSON), now)
	if err != nil {
		return err
	}

	rr.CreatedAt = now
	return tx.Commit()
}

func (s *Store) GetRepairRecords(ticketID string) ([]*models.RepairRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, ticket_id, description, parts_used, created_at
		FROM repair_records WHERE ticket_id = ? ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.RepairRecord
	for rows.Next() {
		var rr models.RepairRecord
		var partsJSON string
		err := rows.Scan(&rr.ID, &rr.TicketID, &rr.Description, &partsJSON, &rr.CreatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(partsJSON), &rr.PartsUsed)
		records = append(records, &rr)
	}
	return records, rows.Err()
}
