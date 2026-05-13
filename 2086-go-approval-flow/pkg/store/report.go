package store

import (
	"approval-flow/pkg/db"
	"approval-flow/pkg/model"
	"approval-flow/pkg/util"
)

func CreateOrUpdateReport(report *model.Report) error {
	if report.ID == "" {
		report.ID = util.NewUUID()
	}

	detailStatsJSON := util.ToJSON(report.DetailStats)

	query := `INSERT INTO reports (id, report_date, chain_id, total_amount, approved_count, rejected_count, pending_count, detail_stats) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET 
			total_amount = excluded.total_amount,
			approved_count = excluded.approved_count,
			rejected_count = excluded.rejected_count,
			pending_count = excluded.pending_count,
			detail_stats = excluded.detail_stats,
			updated_at = CURRENT_TIMESTAMP`
	_, err := db.DB.Exec(query, report.ID, report.ReportDate, report.ChainID, report.TotalAmount, report.ApprovedCount, report.RejectedCount, report.PendingCount, detailStatsJSON)
	return err
}

func GetReportByDateAndChain(reportDate, chainID string) (*model.Report, error) {
	query := `SELECT id, report_date, chain_id, total_amount, approved_count, rejected_count, pending_count, detail_stats, created_at, updated_at FROM reports WHERE report_date = ? AND chain_id = ?`
	row := db.DB.QueryRow(query, reportDate, chainID)

	report := &model.Report{}
	var detailStatsJSON string
	err := row.Scan(&report.ID, &report.ReportDate, &report.ChainID, &report.TotalAmount, &report.ApprovedCount, &report.RejectedCount, &report.PendingCount, &detailStatsJSON, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if detailStatsJSON != "" && detailStatsJSON != "null" {
		report.DetailStats = map[string]map[string]int{}
		_ = util.FromJSON(detailStatsJSON, &report.DetailStats)
	}

	return report, nil
}

func ListReports() ([]*model.Report, error) {
	query := `SELECT id, report_date, chain_id, total_amount, approved_count, rejected_count, pending_count, detail_stats, created_at, updated_at FROM reports ORDER BY report_date DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := []*model.Report{}
	for rows.Next() {
		report := &model.Report{}
		var detailStatsJSON string
		err := rows.Scan(&report.ID, &report.ReportDate, &report.ChainID, &report.TotalAmount, &report.ApprovedCount, &report.RejectedCount, &report.PendingCount, &detailStatsJSON, &report.CreatedAt, &report.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if detailStatsJSON != "" && detailStatsJSON != "null" {
			report.DetailStats = map[string]map[string]int{}
			_ = util.FromJSON(detailStatsJSON, &report.DetailStats)
		}
		reports = append(reports, report)
	}
	return reports, nil
}
