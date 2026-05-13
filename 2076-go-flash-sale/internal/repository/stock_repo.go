package repository

import (
	"database/sql"

	"flashsale/internal/model"
)

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) GetByActivityID(activityID int64) (*model.Stock, error) {
	row := r.db.QueryRow(
		`SELECT id, activity_id, available_stock, version
		FROM stocks WHERE activity_id = ?`, activityID,
	)

	stock := &model.Stock{}
	err := row.Scan(&stock.ID, &stock.ActivityID, &stock.AvailableStock, &stock.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return stock, nil
}

func (r *StockRepository) DecrementStock(activityID int64, quantity int) (bool, error) {
	result, err := r.db.Exec(
		`UPDATE stocks SET available_stock = available_stock - ?, version = version + 1
		WHERE activity_id = ? AND available_stock >= ?`,
		quantity, activityID, quantity,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *StockRepository) IncrementStock(activityID int64, quantity int) error {
	_, err := r.db.Exec(
		`UPDATE stocks SET available_stock = available_stock + ?, version = version + 1
		WHERE activity_id = ?`,
		quantity, activityID,
	)
	return err
}
