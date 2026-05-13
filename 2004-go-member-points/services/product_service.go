package services

import (
	"database/sql"
	"errors"

	"member-points/database"
	"member-points/models"
)

type InsufficientPointsError struct {
	CurrentPoints int64
	RequiredPoints int64
}

func (e *InsufficientPointsError) Error() string {
	return "insufficient points"
}

type AlreadyExchangedError struct{}

func (e *AlreadyExchangedError) Error() string {
	return "product already exchanged by this member"
}

func CreateProduct(name string, pointsCost int64, description string) (*models.Product, error) {
	if pointsCost <= 0 {
		return nil, errors.New("points cost must be positive")
	}

	result, err := database.DB.Exec(
		"INSERT INTO products (name, points_cost, description) VALUES (?, ?, ?)",
		name, pointsCost, description,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetProductByID(id)
}

func GetProductByID(id int64) (*models.Product, error) {
	var product models.Product
	err := database.DB.QueryRow(
		"SELECT id, name, points_cost, description FROM products WHERE id = ?",
		id,
	).Scan(&product.ID, &product.Name, &product.PointsCost, &product.Description)

	if err == sql.ErrNoRows {
		return nil, errors.New("product not found")
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func ListProducts() ([]*models.Product, error) {
	rows, err := database.DB.Query("SELECT id, name, points_cost, description FROM products ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.PointsCost, &product.Description); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func ExchangeProduct(memberID int64, productID int64) (*models.Exchange, *models.Member, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	member, err := getMemberByID(tx, memberID)
	if err != nil {
		return nil, nil, err
	}

	product, err := getProductByIDTx(tx, productID)
	if err != nil {
		return nil, nil, err
	}

	if member.Points < product.PointsCost {
		return nil, nil, &InsufficientPointsError{
			CurrentPoints: member.Points,
			RequiredPoints: product.PointsCost,
		}
	}

	exchangeID, err := insertExchangeTx(tx, memberID, productID, product.PointsCost)
	if err != nil {
		return nil, nil, err
	}

	_, err = tx.Exec(
		"UPDATE members SET points = points - ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		product.PointsCost, memberID,
	)
	if err != nil {
		return nil, nil, err
	}

	err = syncAndMarkExchange(tx, exchangeID, memberID, productID)
	if err != nil {
		return nil, nil, err
	}

	exchange, err := getExchangeByID(tx, exchangeID)
	if err != nil {
		return nil, nil, err
	}

	finalMember, err := getMemberByID(tx, memberID)
	if err != nil {
		return nil, nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}

	return exchange, finalMember, nil
}

func getProductByIDTx(tx *sql.Tx, id int64) (*models.Product, error) {
	var product models.Product
	err := tx.QueryRow(
		"SELECT id, name, points_cost, description FROM products WHERE id = ?",
		id,
	).Scan(&product.ID, &product.Name, &product.PointsCost, &product.Description)

	if err == sql.ErrNoRows {
		return nil, errors.New("product not found")
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func insertExchangeTx(tx *sql.Tx, memberID, productID, pointsUsed int64) (int64, error) {
	result, err := tx.Exec(
		`INSERT INTO exchanges (member_id, product_id, points_used, processed)
		 VALUES (?, ?, ?, 0)`,
		memberID, productID, pointsUsed,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return 0, &AlreadyExchangedError{}
		}
		return 0, err
	}

	return result.LastInsertId()
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "UNIQUE") || contains(errStr, "unique") || contains(errStr, "constraint")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func syncAndMarkExchange(tx *sql.Tx, exchangeID, memberID, productID int64) error {
	var exchange models.Exchange
	err := tx.QueryRow(
		`SELECT id, member_id, product_id, points_used, processed, created_at
		 FROM exchanges WHERE id = ? AND processed = 0`,
		exchangeID,
	).Scan(&exchange.ID, &exchange.MemberID, &exchange.ProductID, &exchange.PointsUsed,
		&exchange.Processed, &exchange.CreatedAt)

	if err != nil {
		return err
	}

	var points int64
	err = tx.QueryRow("SELECT points FROM members WHERE id = ?", memberID).Scan(&points)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE exchanges SET processed = 1 WHERE id = ?", exchangeID)
	return err
}

func getExchangeByID(tx *sql.Tx, id int64) (*models.Exchange, error) {
	var exchange models.Exchange
	err := tx.QueryRow(
		`SELECT id, member_id, product_id, points_used, processed, created_at
		 FROM exchanges WHERE id = ?`,
		id,
	).Scan(&exchange.ID, &exchange.MemberID, &exchange.ProductID, &exchange.PointsUsed,
		&exchange.Processed, &exchange.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("exchange not found")
	}

	if err != nil {
		return nil, err
	}

	return &exchange, nil
}
