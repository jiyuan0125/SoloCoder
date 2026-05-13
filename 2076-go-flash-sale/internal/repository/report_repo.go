package repository

import (
	"database/sql"

	"flashsale/internal/model"
	"flashsale/internal/util"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Upsert(report *model.Report) error {
	_, err := r.db.Exec(
		`INSERT INTO reports 
		(activity_id, total_orders, paid_orders, unpaid_orders, cancelled_orders, total_revenue, generated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(activity_id) DO UPDATE SET
		total_orders = excluded.total_orders,
		paid_orders = excluded.paid_orders,
		unpaid_orders = excluded.unpaid_orders,
		cancelled_orders = excluded.cancelled_orders,
		total_revenue = excluded.total_revenue,
		generated_at = CURRENT_TIMESTAMP`,
		report.ActivityID, report.TotalOrders, report.PaidOrders,
		report.UnpaidOrders, report.CancelledOrders, report.TotalRevenue,
	)
	return err
}

func (r *ReportRepository) GetByActivityID(activityID int64) (*model.Report, error) {
	row := r.db.QueryRow(
		`SELECT id, activity_id, total_orders, paid_orders, unpaid_orders, cancelled_orders, total_revenue, generated_at, verified_at
		FROM reports WHERE activity_id = ?`, activityID,
	)

	report := &model.Report{}
	var generatedAtStr string
	var verifiedAtStr sql.NullString
	err := row.Scan(
		&report.ID, &report.ActivityID, &report.TotalOrders, &report.PaidOrders,
		&report.UnpaidOrders, &report.CancelledOrders, &report.TotalRevenue,
		&generatedAtStr, &verifiedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	report.GeneratedAt, _ = util.ParseTime(generatedAtStr)
	if verifiedAtStr.Valid {
		t, _ := util.ParseTime(verifiedAtStr.String)
		report.VerifiedAt = &t
	}

	return report, nil
}

func (r *ReportRepository) MarkVerified(activityID int64) error {
	_, err := r.db.Exec(
		`UPDATE reports SET verified_at = CURRENT_TIMESTAMP WHERE activity_id = ?`,
		activityID,
	)
	return err
}
