package repository

import (
	"database/sql"
	"time"

	"flashsale/internal/model"
)

type ActivityRepository struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(activity *model.Activity) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO activities (product_name, original_price, flash_price, total_stock, start_time)
		VALUES (?, ?, ?, ?, ?)`,
		activity.ProductName, activity.OriginalPrice, activity.FlashPrice,
		activity.TotalStock, activity.StartTime,
	)
	if err != nil {
		return err
	}

	activityID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	activity.ID = activityID

	if _, err := tx.Exec(
		`INSERT INTO stocks (activity_id, available_stock, version) VALUES (?, ?, 0)`,
		activityID, activity.TotalStock,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ActivityRepository) GetByID(id int64) (*model.Activity, error) {
	row := r.db.QueryRow(
		`SELECT id, product_name, original_price, flash_price, total_stock, start_time
		FROM activities WHERE id = ?`, id,
	)

	activity := &model.Activity{}
	var startTimeStr string
	err := row.Scan(
		&activity.ID, &activity.ProductName, &activity.OriginalPrice,
		&activity.FlashPrice, &activity.TotalStock, &startTimeStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	activity.StartTime, err = time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		activity.StartTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, err
		}
	}

	return activity, nil
}

func (r *ActivityRepository) GetByIDWithStatus(id int64) (*model.Activity, error) {
	return r.GetByID(id)
}

func (r *ActivityRepository) ListActive() ([]*model.Activity, error) {
	rows, err := r.db.Query(
		`SELECT id, product_name, original_price, flash_price, total_stock, start_time
		FROM activities ORDER BY start_time DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]*model.Activity, 0)
	for rows.Next() {
		activity := &model.Activity{}
		var startTimeStr string
		err := rows.Scan(
			&activity.ID, &activity.ProductName, &activity.OriginalPrice,
			&activity.FlashPrice, &activity.TotalStock, &startTimeStr,
		)
		if err != nil {
			return nil, err
		}

		activity.StartTime, err = time.Parse("2006-01-02 15:04:05", startTimeStr)
		if err != nil {
			activity.StartTime, _ = time.Parse(time.RFC3339, startTimeStr)
		}

		activities = append(activities, activity)
	}

	return activities, nil
}
