package repository

import (
	"database/sql"
	"delivery-tracking/models"
	"fmt"
	"os"

	_ "github.com/glebarez/sqlite"
)

type DeliveryRepository interface {
	InitDB() error
	InsertStatus(status *models.DeliveryStatus) error
	CheckDuplicate(orderNo string, statusName string, occurredAt int64) (bool, error)
	GetTrajectoryByOrderNo(orderNo string) ([]models.DeliveryStatus, error)
	Close() error
}

type SQLiteDeliveryRepository struct {
	db *sql.DB
}

func NewSQLiteDeliveryRepository(dbPath string) (*SQLiteDeliveryRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &SQLiteDeliveryRepository{db: db}
	if err := repo.InitDB(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteDeliveryRepository) InitDB() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS delivery_statuses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_no TEXT NOT NULL,
		status_name TEXT NOT NULL,
		longitude REAL NOT NULL,
		latitude REAL NOT NULL,
		occurred_at INTEGER NOT NULL,
		received_at INTEGER NOT NULL,
		UNIQUE(order_no, status_name, occurred_at)
	);
	CREATE INDEX IF NOT EXISTS idx_order_no ON delivery_statuses(order_no);
	CREATE INDEX IF NOT EXISTS idx_occurred_at ON delivery_statuses(occurred_at);
	`

	_, err := r.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (r *SQLiteDeliveryRepository) InsertStatus(status *models.DeliveryStatus) error {
	insertSQL := `
	INSERT OR IGNORE INTO delivery_statuses (
		order_no, status_name, longitude, latitude, occurred_at, received_at
	) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(insertSQL,
		status.OrderNo,
		status.StatusName,
		status.Longitude,
		status.Latitude,
		status.OccurredAt,
		status.ReceivedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("duplicate entry, status not inserted")
	}

	id, err := result.LastInsertId()
	if err == nil {
		status.ID = id
	}

	return nil
}

func (r *SQLiteDeliveryRepository) CheckDuplicate(orderNo string, statusName string, occurredAt int64) (bool, error) {
	checkSQL := `
	SELECT COUNT(*) FROM delivery_statuses 
	WHERE order_no = ? AND status_name = ? AND occurred_at = ?
	`

	var count int
	err := r.db.QueryRow(checkSQL, orderNo, statusName, occurredAt).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return count > 0, nil
}

func (r *SQLiteDeliveryRepository) GetTrajectoryByOrderNo(orderNo string) ([]models.DeliveryStatus, error) {
	querySQL := `
	SELECT id, order_no, status_name, longitude, latitude, occurred_at, received_at
	FROM delivery_statuses
	WHERE order_no = ?
	ORDER BY occurred_at ASC, id ASC
	`

	rows, err := r.db.Query(querySQL, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to query trajectory: %w", err)
	}
	defer rows.Close()

	var statuses []models.DeliveryStatus
	for rows.Next() {
		var status models.DeliveryStatus
		err := rows.Scan(
			&status.ID,
			&status.OrderNo,
			&status.StatusName,
			&status.Longitude,
			&status.Latitude,
			&status.OccurredAt,
			&status.ReceivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		statuses = append(statuses, status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return statuses, nil
}

func (r *SQLiteDeliveryRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func DBExists(dbPath string) bool {
	_, err := os.Stat(dbPath)
	return err == nil
}
