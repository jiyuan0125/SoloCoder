package repository

import (
	"database/sql"
	"time"

	"points-mall/db"
	"points-mall/models"
)

func CreateExchange(userID int, productID int, points int, lockExpiresAt *time.Time) (*models.Exchange, error) {
	var expiresAt interface{}
	if lockExpiresAt != nil {
		expiresAt = lockExpiresAt.Format("2006-01-02 15:04:05")
	}

	result, err := db.DB.Exec(
		"INSERT INTO exchanges (user_id, product_id, status, points, lock_expires_at) VALUES (?, ?, ?, ?, ?)",
		userID, productID, models.StatusNew, points, expiresAt,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	err = AddExchangeDetail(int(id), "create", nil, models.StatusNew, "创建兑换记录")
	if err != nil {
		return nil, err
	}

	return GetExchangeByID(int(id))
}

func GetExchangeByID(id int) (*models.Exchange, error) {
	e := &models.Exchange{}
	var lockExpiresAt sql.NullString
	err := db.DB.QueryRow(
		"SELECT id, user_id, product_id, status, points, lock_expires_at, created_at, updated_at FROM exchanges WHERE id = ?",
		id,
	).Scan(&e.ID, &e.UserID, &e.ProductID, &e.Status, &e.Points, &lockExpiresAt, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lockExpiresAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lockExpiresAt.String)
		e.LockExpiresAt = &t
	}

	return e, nil
}

func UpdateExchangeStatus(exchangeID int, oldStatus string, newStatus string, actionType string, description string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"UPDATE exchanges SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?",
		newStatus, exchangeID, oldStatus,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		exchangeID, actionType, oldStatus, newStatus, description,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func ClearExchangeLock(exchangeID int) error {
	_, err := db.DB.Exec(
		"UPDATE exchanges SET lock_expires_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		exchangeID,
	)
	return err
}

func GetExpiredLocks() ([]*models.Exchange, error) {
	rows, err := db.DB.Query(
		"SELECT id, user_id, product_id, status, points, lock_expires_at, created_at, updated_at FROM exchanges WHERE status = ? AND lock_expires_at IS NOT NULL AND lock_expires_at < datetime('now')",
		models.StatusNew,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exchanges []*models.Exchange
	for rows.Next() {
		e := &models.Exchange{}
		var lockExpiresAt sql.NullString
		err := rows.Scan(&e.ID, &e.UserID, &e.ProductID, &e.Status, &e.Points, &lockExpiresAt, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if lockExpiresAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", lockExpiresAt.String)
			e.LockExpiresAt = &t
		}
		exchanges = append(exchanges, e)
	}
	return exchanges, nil
}

func AddExchangeDetail(exchangeID int, actionType string, oldStatus *string, newStatus string, description string) error {
	var old interface{}
	if oldStatus != nil {
		old = *oldStatus
	}
	_, err := db.DB.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		exchangeID, actionType, old, newStatus, description,
	)
	return err
}

func GetExchangeDetails(exchangeID int) ([]*models.ExchangeDetail, error) {
	rows, err := db.DB.Query(
		"SELECT id, exchange_id, action_type, old_status, new_status, description, created_at FROM exchange_details WHERE exchange_id = ? ORDER BY created_at ASC",
		exchangeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []*models.ExchangeDetail
	for rows.Next() {
		d := &models.ExchangeDetail{}
		var oldStatus sql.NullString
		err := rows.Scan(&d.ID, &d.ExchangeID, &d.ActionType, &oldStatus, &d.NewStatus, &d.Description, &d.CreatedAt)
		if err != nil {
			return nil, err
		}
		if oldStatus.Valid {
			d.OldStatus = &oldStatus.String
		}
		details = append(details, d)
	}
	return details, nil
}

func GetUserExchanges(userID int) ([]*models.Exchange, error) {
	rows, err := db.DB.Query(
		"SELECT id, user_id, product_id, status, points, lock_expires_at, created_at, updated_at FROM exchanges WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exchanges []*models.Exchange
	for rows.Next() {
		e := &models.Exchange{}
		var lockExpiresAt sql.NullString
		err := rows.Scan(&e.ID, &e.UserID, &e.ProductID, &e.Status, &e.Points, &lockExpiresAt, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if lockExpiresAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", lockExpiresAt.String)
			e.LockExpiresAt = &t
		}
		exchanges = append(exchanges, e)
	}
	return exchanges, nil
}
