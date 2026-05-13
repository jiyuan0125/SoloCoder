package model

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"time"
)

type OrderSide string

const (
	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPartial   OrderStatus = "partial"
	OrderFilled    OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID             int64
	UserID         string
	Symbol         string
	Side           OrderSide
	Price          float64
	Quantity       int64
	FilledQuantity int64
	Status         OrderStatus
	CreatedAt      time.Time
	Version        int64
}

type Trade struct {
	ID         int64
	BuyOrderID int64
	SellOrderID int64
	BuyerID    string
	SellerID   string
	Symbol     string
	Price      float64
	Quantity   int64
	CreatedAt  time.Time
}

type OrderBookEntry struct {
	Price    float64
	Quantity int64
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{db: db}
	if err := d.initSchema(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL,
		price REAL NOT NULL,
		quantity INTEGER NOT NULL,
		filled_quantity INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME NOT NULL,
		version INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders(symbol);
	CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
	CREATE INDEX IF NOT EXISTS idx_orders_symbol_side ON orders(symbol, side);

	CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		buy_order_id INTEGER NOT NULL,
		sell_order_id INTEGER NOT NULL,
		buyer_id TEXT NOT NULL,
		seller_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		price REAL NOT NULL,
		quantity INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
	`
	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *Database) InsertOrder(tx *sql.Tx, order *Order) error {
	query := `
	INSERT INTO orders (user_id, symbol, side, price, quantity, filled_quantity, status, created_at, version)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query,
		order.UserID,
		order.Symbol,
		order.Side,
		order.Price,
		order.Quantity,
		order.FilledQuantity,
		order.Status,
		order.CreatedAt,
		order.Version,
	)
	if err != nil {
		return err
	}
	order.ID, err = result.LastInsertId()
	return err
}

func (d *Database) UpdateOrder(tx *sql.Tx, order *Order) error {
	query := `
	UPDATE orders 
	SET filled_quantity = ?, status = ?, version = version + 1
	WHERE id = ? AND version = ?
	`
	result, err := tx.Exec(query,
		order.FilledQuantity,
		order.Status,
		order.ID,
		order.Version,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("optimistic lock failed for order %d", order.ID)
	}
	order.Version++
	return nil
}

func (d *Database) GetOrder(tx *sql.Tx, id int64) (*Order, error) {
	query := `
	SELECT id, user_id, symbol, side, price, quantity, filled_quantity, status, created_at, version
	FROM orders WHERE id = ?
	`
	row := tx.QueryRow(query, id)
	order := &Order{}
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Symbol,
		&order.Side,
		&order.Price,
		&order.Quantity,
		&order.FilledQuantity,
		&order.Status,
		&order.CreatedAt,
		&order.Version,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (d *Database) GetPendingOrders(tx *sql.Tx, symbol string, side OrderSide) ([]*Order, error) {
	ordering := "ASC"
	if side == Buy {
		ordering = "DESC"
	}
	query := fmt.Sprintf(`
	SELECT id, user_id, symbol, side, price, quantity, filled_quantity, status, created_at, version
	FROM orders 
	WHERE symbol = ? AND side = ? AND status IN ('pending', 'partial')
	ORDER BY price %s, created_at ASC
	`, ordering)

	rows, err := tx.Query(query, symbol, side)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		order := &Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Symbol,
			&order.Side,
			&order.Price,
			&order.Quantity,
			&order.FilledQuantity,
			&order.Status,
			&order.CreatedAt,
			&order.Version,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (d *Database) InsertTrade(tx *sql.Tx, trade *Trade) error {
	query := `
	INSERT INTO trades (buy_order_id, sell_order_id, buyer_id, seller_id, symbol, price, quantity, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query,
		trade.BuyOrderID,
		trade.SellOrderID,
		trade.BuyerID,
		trade.SellerID,
		trade.Symbol,
		trade.Price,
		trade.Quantity,
		trade.CreatedAt,
	)
	if err != nil {
		return err
	}
	trade.ID, err = result.LastInsertId()
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}
