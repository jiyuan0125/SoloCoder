package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	dbPath := filepath.Join("/tmp", "return_exchange.db")
	
	// 检查数据库文件是否存在
	_, err := os.Stat(dbPath)
	dbExists := !os.IsNotExist(err)
	
	// 打开数据库连接（如果不存在会自动创建）
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	// 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	// 如果数据库文件不存在，创建表结构
	if !dbExists {
		if err = createTables(); err != nil {
			return fmt.Errorf("failed to create tables: %w", err)
		}
	}
	
	return nil
}

func createTables() error {
	// 创建订单表（用于存储订单信息）
	orderTable := `
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_no TEXT NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		sku TEXT NOT NULL,
		specification TEXT,
		original_price REAL NOT NULL,
		actual_payment REAL NOT NULL,
		shipping_fee REAL NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	// 创建退换货申请表
	returnExchangeTable := `
	CREATE TABLE IF NOT EXISTS return_exchanges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_no TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL, -- 'return' 或 'exchange'
		reason TEXT NOT NULL,
		reason_detail TEXT,
		status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected, processing, completed
		original_sku TEXT NOT NULL,
		original_specification TEXT,
		original_price REAL NOT NULL,
		actual_payment REAL NOT NULL,
		shipping_fee REAL NOT NULL DEFAULT 0,
		shipping_bearer TEXT, -- 'buyer' 或 'platform'
		-- 换货相关字段
		new_sku TEXT,
		new_specification TEXT,
		new_price REAL,
		price_difference REAL,
		-- 审核相关
		admin_id INTEGER,
		review_comment TEXT,
		reviewed_at TIMESTAMP,
		-- 时间戳
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (order_no) REFERENCES orders(order_no)
	);`
	
	// 创建凭证图片表
	evidenceTable := `
	CREATE TABLE IF NOT EXISTS evidences (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		return_exchange_id INTEGER NOT NULL,
		image_base64 TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (return_exchange_id) REFERENCES return_exchanges(id)
	);`
	
	// 创建退款记录表
	refundTable := `
	CREATE TABLE IF NOT EXISTS refunds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		return_exchange_id INTEGER NOT NULL,
		refund_no TEXT NOT NULL UNIQUE,
		refund_amount REAL NOT NULL,
		shipping_fee_refund REAL NOT NULL DEFAULT 0,
		total_refund REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending', -- pending, processing, completed
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		FOREIGN KEY (return_exchange_id) REFERENCES return_exchanges(id)
	);`
	
	// 创建发货单表
	shipmentTable := `
	CREATE TABLE IF NOT EXISTS shipments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		return_exchange_id INTEGER NOT NULL,
		shipment_no TEXT NOT NULL UNIQUE,
		sku TEXT NOT NULL,
		specification TEXT,
		quantity INTEGER NOT NULL DEFAULT 1,
		shipping_address TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending', -- pending, shipped, delivered
		shipped_at TIMESTAMP,
		delivered_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (return_exchange_id) REFERENCES return_exchanges(id)
	);`
	
	// 执行表创建
	tables := []string{orderTable, returnExchangeTable, evidenceTable, refundTable, shipmentTable}
	
	for _, tableSQL := range tables {
		_, err := DB.Exec(tableSQL)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	
	// 插入一些测试订单数据
	testOrders := []struct {
		orderNo        string
		userID         int
		sku            string
		specification  string
		originalPrice  float64
		actualPayment  float64
		shippingFee    float64
	}{
		{"ORD20260501001", 1001, "SKU001", "红色/L", 199.99, 179.99, 15.00},
		{"ORD20260501002", 1002, "SKU002", "蓝色/M", 299.99, 269.99, 20.00},
		{"ORD20260501003", 1001, "SKU003", "黑色/XL", 159.99, 149.99, 10.00},
	}
	
	insertOrderSQL := `
	INSERT INTO orders (order_no, user_id, sku, specification, original_price, actual_payment, shipping_fee)
	VALUES (?, ?, ?, ?, ?, ?, ?);`
	
	for _, order := range testOrders {
		_, err := DB.Exec(insertOrderSQL, 
			order.orderNo, order.userID, order.sku, order.specification, 
			order.originalPrice, order.actualPayment, order.shippingFee)
		if err != nil {
			return fmt.Errorf("failed to insert test order: %w", err)
		}
	}
	
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
