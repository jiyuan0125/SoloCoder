package repository

import (
	"database/sql"
	"time"

	"flashsale/internal/model"
	"flashsale/internal/util"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *model.Order) error {
	_, err := r.db.Exec(
		`INSERT INTO orders (order_no, activity_id, user_id, status, price, expire_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		order.OrderNo, order.ActivityID, order.UserID,
		order.Status, order.Price, order.ExpireAt,
	)
	return err
}

func (r *OrderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	row := r.db.QueryRow(
		`SELECT id, order_no, activity_id, user_id, status, price, created_at, paid_at, cancelled_at, expire_at
		FROM orders WHERE order_no = ?`, orderNo,
	)

	order := &model.Order{}
	var createdAtStr, expireAtStr string
	var paidAtStr, cancelledAtStr sql.NullString
	err := row.Scan(
		&order.ID, &order.OrderNo, &order.ActivityID, &order.UserID,
		&order.Status, &order.Price, &createdAtStr, &paidAtStr, &cancelledAtStr, &expireAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	order.CreatedAt, _ = util.ParseTime(createdAtStr)
	order.ExpireAt, _ = util.ParseTime(expireAtStr)
	if paidAtStr.Valid {
		t, _ := util.ParseTime(paidAtStr.String)
		order.PaidAt = &t
	}
	if cancelledAtStr.Valid {
		t, _ := util.ParseTime(cancelledAtStr.String)
		order.CancelledAt = &t
	}

	return order, nil
}

func (r *OrderRepository) Exists(activityID int64, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM orders WHERE activity_id = ? AND user_id = ?`,
		activityID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *OrderRepository) UpdateStatus(orderNo string, status model.OrderStatus, isPayment bool) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.QueryRow(
		`SELECT status FROM orders WHERE order_no = ?`, orderNo,
	).Scan(&currentStatus)
	if err != nil {
		return false, err
	}

	currentOrderStatus := model.OrderStatus(currentStatus)
	if currentOrderStatus != model.OrderStatusPending {
		return false, nil
	}

	var result sql.Result
	if isPayment {
		result, err = tx.Exec(
			`UPDATE orders SET status = ?, paid_at = CURRENT_TIMESTAMP WHERE order_no = ? AND status = ?`,
			status, orderNo, model.OrderStatusPending,
		)
	} else {
		result, err = tx.Exec(
			`UPDATE orders SET status = ?, cancelled_at = CURRENT_TIMESTAMP WHERE order_no = ? AND status = ?`,
			status, orderNo, model.OrderStatusPending,
		)
	}

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, tx.Commit()
	}

	return true, tx.Commit()
}

func (r *OrderRepository) GetExpiredOrders(now time.Time) ([]*model.Order, error) {
	rows, err := r.db.Query(
		`SELECT id, order_no, activity_id, user_id, status, price, created_at, paid_at, cancelled_at, expire_at
		FROM orders WHERE status = ? AND expire_at <= ?`,
		model.OrderStatusPending, now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*model.Order, 0)
	for rows.Next() {
		order := &model.Order{}
		var createdAtStr, expireAtStr string
		var paidAtStr, cancelledAtStr sql.NullString
		err := rows.Scan(
			&order.ID, &order.OrderNo, &order.ActivityID, &order.UserID,
			&order.Status, &order.Price, &createdAtStr, &paidAtStr, &cancelledAtStr, &expireAtStr,
		)
		if err != nil {
			return nil, err
		}

		order.CreatedAt, _ = util.ParseTime(createdAtStr)
		order.ExpireAt, _ = util.ParseTime(expireAtStr)
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepository) GetByActivityID(activityID int64) ([]*model.Order, error) {
	rows, err := r.db.Query(
		`SELECT id, order_no, activity_id, user_id, status, price, created_at, paid_at, cancelled_at, expire_at
		FROM orders WHERE activity_id = ?`, activityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*model.Order, 0)
	for rows.Next() {
		order := &model.Order{}
		var createdAtStr, expireAtStr string
		var paidAtStr, cancelledAtStr sql.NullString
		err := rows.Scan(
			&order.ID, &order.OrderNo, &order.ActivityID, &order.UserID,
			&order.Status, &order.Price, &createdAtStr, &paidAtStr, &cancelledAtStr, &expireAtStr,
		)
		if err != nil {
			return nil, err
		}

		order.CreatedAt, _ = util.ParseTime(createdAtStr)
		order.ExpireAt, _ = util.ParseTime(expireAtStr)
		if paidAtStr.Valid {
			t, _ := util.ParseTime(paidAtStr.String)
			order.PaidAt = &t
		}
		if cancelledAtStr.Valid {
			t, _ := util.ParseTime(cancelledAtStr.String)
			order.CancelledAt = &t
		}

		orders = append(orders, order)
	}

	return orders, nil
}
