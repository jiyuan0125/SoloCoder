package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"inventory-alert/models"
)

var db *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

func createTables() error {
	productTable := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category TEXT NOT NULL,
		current_stock INTEGER NOT NULL,
		safety_stock INTEGER NOT NULL,
		purchase_lead_time INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	alertTable := `
	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER NOT NULL,
		product_name TEXT NOT NULL,
		category TEXT NOT NULL,
		current_stock INTEGER NOT NULL,
		safety_stock INTEGER NOT NULL,
		level TEXT NOT NULL,
		status TEXT NOT NULL,
		assigned_to TEXT DEFAULT '',
		suggested_qty INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (product_id) REFERENCES products(id)
	);`

	_, err := db.Exec(productTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(alertTable)
	if err != nil {
		return err
	}

	return nil
}

func CreateProduct(product *models.Product) (int64, error) {
	query := `
	INSERT INTO products (name, category, current_stock, safety_stock, purchase_lead_time, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := db.Exec(query, product.Name, product.Category, product.CurrentStock, 
		product.SafetyStock, product.PurchaseLeadTime, now, now)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetProductByID(id int64) (*models.Product, error) {
	query := `SELECT id, name, category, current_stock, safety_stock, purchase_lead_time, created_at, updated_at 
	          FROM products WHERE id = ?`
	
	product := &models.Product{}
	err := db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.Category, &product.CurrentStock,
		&product.SafetyStock, &product.PurchaseLeadTime, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return product, nil
}

func GetAllProducts() ([]*models.Product, error) {
	query := `SELECT id, name, category, current_stock, safety_stock, purchase_lead_time, created_at, updated_at 
	          FROM products`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []*models.Product{}
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Category, &product.CurrentStock,
			&product.SafetyStock, &product.PurchaseLeadTime, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func UpdateProduct(product *models.Product) error {
	query := `
	UPDATE products 
	SET name = ?, category = ?, current_stock = ?, safety_stock = ?, purchase_lead_time = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query, product.Name, product.Category, product.CurrentStock,
		product.SafetyStock, product.PurchaseLeadTime, time.Now(), product.ID)
	return err
}

func DeleteProduct(id int64) error {
	query := `DELETE FROM products WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func UpdateStock(productID int64, quantity int) error {
	query := `
	UPDATE products 
	SET current_stock = current_stock + ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query, quantity, time.Now(), productID)
	return err
}

func GetLastAlertTime(productID int64) (time.Time, error) {
	query := `SELECT created_at FROM alerts WHERE product_id = ? ORDER BY created_at DESC LIMIT 1`
	
	var lastAlertTime time.Time
	err := db.QueryRow(query, productID).Scan(&lastAlertTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	return lastAlertTime, nil
}

func CreateAlert(alert *models.Alert) (int64, error) {
	query := `
	INSERT INTO alerts (product_id, product_name, category, current_stock, safety_stock, level, status, assigned_to, suggested_qty, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := db.Exec(query, alert.ProductID, alert.ProductName, alert.Category,
		alert.CurrentStock, alert.SafetyStock, alert.Level, alert.Status, alert.AssignedTo,
		alert.SuggestedQty, now, now)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetAlertByID(id int64) (*models.Alert, error) {
	query := `SELECT id, product_id, product_name, category, current_stock, safety_stock, level, status, assigned_to, suggested_qty, created_at, updated_at 
	          FROM alerts WHERE id = ?`
	
	alert := &models.Alert{}
	err := db.QueryRow(query, id).Scan(
		&alert.ID, &alert.ProductID, &alert.ProductName, &alert.Category, &alert.CurrentStock,
		&alert.SafetyStock, &alert.Level, &alert.Status, &alert.AssignedTo, &alert.SuggestedQty,
		&alert.CreatedAt, &alert.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return alert, nil
}

func GetAllAlerts(level string) ([]*models.Alert, error) {
	query := `SELECT id, product_id, product_name, category, current_stock, safety_stock, level, status, assigned_to, suggested_qty, created_at, updated_at 
	          FROM alerts`
	args := []interface{}{}
	
	if level != "" {
		query += ` WHERE level = ?`
		args = append(args, level)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []*models.Alert{}
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.ProductID, &alert.ProductName, &alert.Category, &alert.CurrentStock,
			&alert.SafetyStock, &alert.Level, &alert.Status, &alert.AssignedTo, &alert.SuggestedQty,
			&alert.CreatedAt, &alert.UpdatedAt)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func UpdateAlert(alert *models.Alert) error {
	query := `
	UPDATE alerts 
	SET level = ?, status = ?, assigned_to = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.Exec(query, alert.Level, alert.Status, alert.AssignedTo, time.Now(), alert.ID)
	return err
}

func CloseProductAlerts(productID int64) error {
	query := `
	UPDATE alerts 
	SET status = ?, updated_at = ?
	WHERE product_id = ? AND status != ?`

	_, err := db.Exec(query, models.AlertStatusClosed, time.Now(), productID, models.AlertStatusClosed)
	return err
}

func GetCategoryStats() ([]*models.CategoryStats, error) {
	query := `
	SELECT category, COUNT(*) as count 
	FROM alerts 
	WHERE status != ?
	GROUP BY category
	ORDER BY count DESC`

	rows, err := db.Query(query, models.AlertStatusClosed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []*models.CategoryStats{}
	for rows.Next() {
		stat := &models.CategoryStats{}
		err := rows.Scan(&stat.Category, &stat.Count)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
