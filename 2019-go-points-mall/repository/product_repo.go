package repository

import (
	"database/sql"

	"points-mall/db"
	"points-mall/models"
)

func CreateProduct(name string, points int, stock int, dailyLimit int) (*models.Product, error) {
	result, err := db.DB.Exec(
		"INSERT INTO products (name, points, stock, daily_limit) VALUES (?, ?, ?, ?)",
		name, points, stock, dailyLimit,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetProductByID(int(id))
}

func GetOnlineProducts() ([]*models.Product, error) {
	rows, err := db.DB.Query(
		"SELECT id, name, points, stock, daily_limit, is_online, created_at, updated_at FROM products WHERE is_online = 1",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		p := &models.Product{}
		var isOnline int
		err := rows.Scan(&p.ID, &p.Name, &p.Points, &p.Stock, &p.DailyLimit, &isOnline, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		p.IsOnline = isOnline == 1
		products = append(products, p)
	}
	return products, nil
}

func GetAllProducts() ([]*models.Product, error) {
	rows, err := db.DB.Query(
		"SELECT id, name, points, stock, daily_limit, is_online, created_at, updated_at FROM products",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		p := &models.Product{}
		var isOnline int
		err := rows.Scan(&p.ID, &p.Name, &p.Points, &p.Stock, &p.DailyLimit, &isOnline, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		p.IsOnline = isOnline == 1
		products = append(products, p)
	}
	return products, nil
}

func GetProductByID(id int) (*models.Product, error) {
	p := &models.Product{}
	var isOnline int
	err := db.DB.QueryRow(
		"SELECT id, name, points, stock, daily_limit, is_online, created_at, updated_at FROM products WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.Points, &p.Stock, &p.DailyLimit, &isOnline, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.IsOnline = isOnline == 1
	return p, nil
}

func UpdateProductStatus(id int, isOnline bool) error {
	var status int
	if isOnline {
		status = 1
	}
	_, err := db.DB.Exec(
		"UPDATE products SET is_online = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	return err
}

func UpdateProductPoints(id int, points int) error {
	_, err := db.DB.Exec(
		"UPDATE products SET points = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		points, id,
	)
	return err
}

func DecreaseStock(id int, quantity int) (bool, error) {
	result, err := db.DB.Exec(
		"UPDATE products SET stock = stock - ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND stock >= ?",
		quantity, id, quantity,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func IncreaseStock(id int, quantity int) error {
	_, err := db.DB.Exec(
		"UPDATE products SET stock = stock + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		quantity, id,
	)
	return err
}

func SetStockToZero(id int) error {
	_, err := db.DB.Exec(
		"UPDATE products SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	return err
}

func CheckDailyRedemption(userID int, productID int) (bool, error) {
	var count int
	err := db.DB.QueryRow(
		"SELECT COUNT(*) FROM daily_redemptions WHERE user_id = ? AND product_id = ? AND redemption_date = DATE('now')",
		userID, productID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func AddDailyRedemption(userID int, productID int) error {
	_, err := db.DB.Exec(
		"INSERT OR IGNORE INTO daily_redemptions (user_id, product_id, redemption_date) VALUES (?, ?, DATE('now'))",
		userID, productID,
	)
	return err
}
