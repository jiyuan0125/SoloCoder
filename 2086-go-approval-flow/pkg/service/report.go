package service

import (
	"time"

	"approval-flow/pkg/model"
	"approval-flow/pkg/store"
)

func (s *ApprovalService) UpdateReport(chainID string) error {
	reportDate := time.Now().Format("2006-01-02")

	applications, err := store.ListApplications()
	if err != nil {
		return err
	}

	var totalAmount float64
	var approvedCount, rejectedCount, pendingCount int
	detailStats := map[string]map[string]int{
		"by_level":    {},
		"by_submitter": {},
	}

	for _, app := range applications {
		if app.ChainID != chainID {
			continue
		}

		if amount, ok := app.Data["amount"]; ok {
			if f, ok := amount.(float64); ok {
				totalAmount += f
			}
		}

		switch app.Status {
		case "approved":
			approvedCount++
		case "rejected":
			rejectedCount++
		case "pending":
			pendingCount++
		}

		levelKey := "level_unknown"
		if app.Status == "approved" || app.Status == "rejected" {
			levelKey = "completed"
		} else {
			levelKey = "pending"
		}
		detailStats["by_level"][levelKey]++

		if app.ApplicantID != "" {
			detailStats["by_submitter"][app.ApplicantID]++
		}
	}

	report := &model.Report{
		ReportDate:    reportDate,
		ChainID:       chainID,
		TotalAmount:   totalAmount,
		ApprovedCount: approvedCount,
		RejectedCount: rejectedCount,
		PendingCount:  pendingCount,
		DetailStats:   detailStats,
	}

	existing, err := store.GetReportByDateAndChain(reportDate, chainID)
	if err == nil && existing != nil {
		report.ID = existing.ID
	}

	return store.CreateOrUpdateReport(report)
}

func (s *ApprovalService) ReconcileReport(chainID string) error {
	reportDate := time.Now().Format("2006-01-02")

	applications, err := store.ListApplications()
	if err != nil {
		return err
	}

	verifiedTotal := float64(0)
	verifiedApproved := 0
	verifiedRejected := 0
	verifiedPending := 0
	detailStats := map[string]map[string]int{
		"by_level":    {},
		"by_submitter": {},
	}

	for _, app := range applications {
		if app.ChainID != chainID {
			continue
		}

		if amount, ok := app.Data["amount"]; ok {
			if f, ok := amount.(float64); ok {
				verifiedTotal += f
			}
		}

		switch app.Status {
		case "approved":
			verifiedApproved++
		case "rejected":
			verifiedRejected++
		case "pending":
			verifiedPending++
		}

		levelKey := "level_unknown"
		if app.Status == "approved" || app.Status == "rejected" {
			levelKey = "completed"
		} else {
			levelKey = "pending"
		}
		detailStats["by_level"][levelKey]++

		if app.ApplicantID != "" {
			detailStats["by_submitter"][app.ApplicantID]++
		}
	}

	report := &model.Report{
		ReportDate:    reportDate,
		ChainID:       chainID,
		TotalAmount:   verifiedTotal,
		ApprovedCount: verifiedApproved,
		RejectedCount: verifiedRejected,
		PendingCount:  verifiedPending,
		DetailStats:   detailStats,
	}

	existing, err := store.GetReportByDateAndChain(reportDate, chainID)
	if err == nil && existing != nil {
		report.ID = existing.ID
	}

	return store.CreateOrUpdateReport(report)
}
