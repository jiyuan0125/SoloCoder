package services

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"expense-reimburse/database"
	"expense-reimburse/models"
)

const (
	maxModifyCount = 2
	expireDays     = 30
)

func YuanToCent(amountYuan float64) int64 {
	return int64(math.Round(amountYuan * 100))
}

func CentToYuan(amountCent int64) float64 {
	return float64(amountCent) / 100.0
}

func DetermineApprover(amountCent int64) models.ApprovalRole {
	amountYuan := CentToYuan(amountCent)
	switch {
	case amountYuan < 500:
		return models.RoleManager
	case amountYuan >= 500 && amountYuan < 5000:
		return models.RoleDepartment
	default:
		return models.RoleCFO
	}
}

func GetNextApprover(currentApprover models.ApprovalRole, amountCent int64) (models.ApprovalRole, bool) {
	amountYuan := CentToYuan(amountCent)

	switch currentApprover {
	case models.RoleManager:
		if amountYuan >= 500 {
			return models.RoleDepartment, true
		}
		return "", false
	case models.RoleDepartment:
		if amountYuan >= 5000 {
			return models.RoleCFO, true
		}
		return "", false
	case models.RoleCFO:
		return "", false
	}
	return "", false
}

func CheckDuplicate(employeeID string, occurredDate time.Time, amountCent int64, expenseType models.ExpenseType, excludeID int64) (*models.Reimbursement, error) {
	occurredDateStr := database.FormatDate(occurredDate)

	query := `
		SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
		       status, current_approver, modify_count, created_at, updated_at
		FROM reimbursements
		WHERE employee_id = ? 
		  AND occurred_date = ? 
		  AND amount_cent = ? 
		  AND expense_type = ? 
		  AND status != ?
		  AND id != ?
		LIMIT 1
	`

	row := database.DB.QueryRow(query, employeeID, occurredDateStr, amountCent, expenseType, models.StatusExpired, excludeID)

	rm, err := scanReimbursement(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rm, nil
}

func IsExpired(createdAt time.Time) bool {
	return time.Since(createdAt).Hours() > float64(expireDays*24)
}

func MarkAsExpiredIfNeeded(rm *models.Reimbursement) bool {
	if rm.Status == models.StatusPending && IsExpired(rm.CreatedAt) {
		rm.Status = models.StatusExpired
		return true
	}
	return false
}

func SubmitReimbursement(req *models.SubmitRequest) (*models.Reimbursement, error) {
	occurredDate, err := database.ParseDate(req.OccurredDate)
	if err != nil {
		return nil, fmt.Errorf("invalid occurred_date: %v", err)
	}

	amountCent := YuanToCent(req.AmountYuan)

	if dup, err := CheckDuplicate(req.EmployeeID, occurredDate, amountCent, req.ExpenseType, 0); err != nil {
		return nil, err
	} else if dup != nil {
		return nil, fmt.Errorf("409:已有报销单号 #%d", dup.ID)
	}

	currentApprover := DetermineApprover(amountCent)
	now := time.Now()

	query := `
		INSERT INTO reimbursements 
		(employee_id, amount_cent, expense_type, occurred_date, description, 
		 status, current_approver, modify_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query,
		req.EmployeeID,
		amountCent,
		req.ExpenseType,
		database.FormatDate(occurredDate),
		req.Description,
		models.StatusPending,
		currentApprover,
		0,
		database.FormatTime(now),
		database.FormatTime(now),
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err = addStatusHistory(id, "", models.StatusPending, "system", fmt.Sprintf("提交成功，当前审批人: %s", currentApprover)); err != nil {
		return nil, err
	}

	return GetReimbursementByID(id)
}

func UpdateReimbursement(req *models.UpdateRequest) (*models.Reimbursement, error) {
	rm, err := GetReimbursementByID(req.ReimbursementID)
	if err != nil {
		return nil, err
	}

	if rm.EmployeeID != req.EmployeeID {
		return nil, fmt.Errorf("无权修改他人的报销单")
	}

	if rm.Status != models.StatusRejected {
		return nil, fmt.Errorf("只有被驳回的报销单可以修改")
	}

	if rm.ModifyCount >= maxModifyCount {
		return nil, fmt.Errorf("400:修改次数已达上限")
	}

	newAmountCent := YuanToCent(req.AmountYuan)
	oldAmountCent := rm.AmountCent

	occurredDate := rm.OccurredDate
	if dup, err := CheckDuplicate(req.EmployeeID, occurredDate, newAmountCent, rm.ExpenseType, rm.ID); err != nil {
		return nil, err
	} else if dup != nil {
		return nil, fmt.Errorf("409:已有报销单号 #%d", dup.ID)
	}

	newApprover := DetermineApprover(newAmountCent)
	now := time.Now()

	query := `
		UPDATE reimbursements 
		SET amount_cent = ?, current_approver = ?, modify_count = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = database.DB.Exec(query, newAmountCent, newApprover, rm.ModifyCount+1, models.StatusPending, database.FormatTime(now), rm.ID)
	if err != nil {
		return nil, err
	}

	note := fmt.Sprintf("修改金额: %.2f元 -> %.2f元", CentToYuan(oldAmountCent), CentToYuan(newAmountCent))
	if err = addStatusHistory(rm.ID, models.StatusRejected, models.StatusPending, req.EmployeeID, note); err != nil {
		return nil, err
	}

	return GetReimbursementByID(rm.ID)
}

func ApproveReimbursement(id int64, req *models.ApprovalRequest) (*models.Reimbursement, error) {
	rm, err := GetReimbursementByID(id)
	if err != nil {
		return nil, err
	}

	if MarkAsExpiredIfNeeded(rm) {
		return nil, fmt.Errorf("已过期请重新提交")
	}

	if rm.Status != models.StatusPending {
		return nil, fmt.Errorf("只有待审批状态可以操作")
	}

	if req.ApproverRole != string(rm.CurrentApprover) {
		return nil, fmt.Errorf("无权审批该报销单")
	}

	now := time.Now()
	currentRole := rm.CurrentApprover

	if req.Action == "approve" {
		nextApprover, hasNext := GetNextApprover(currentRole, rm.AmountCent)

		var newStatus models.ReimbursementStatus
		var note string

		if hasNext {
			newStatus = models.StatusPending
			note = fmt.Sprintf("%s审批通过，流转至 %s", currentRole, nextApprover)
		} else {
			newStatus = models.StatusCompleted
			note = fmt.Sprintf("%s审批通过，流程完成", currentRole)
		}

		query := `UPDATE reimbursements SET status = ?, current_approver = ?, updated_at = ? WHERE id = ?`
		_, err = database.DB.Exec(query, newStatus, nextApprover, database.FormatTime(now), rm.ID)
		if err != nil {
			return nil, err
		}

		if err = addStatusHistory(rm.ID, models.StatusPending, newStatus, req.ApproverID, note); err != nil {
			return nil, err
		}

	} else if req.Action == "reject" {
		query := `UPDATE reimbursements SET status = ?, updated_at = ? WHERE id = ?`
		_, err = database.DB.Exec(query, models.StatusRejected, database.FormatTime(now), rm.ID)
		if err != nil {
			return nil, err
		}

		note := fmt.Sprintf("%s驳回", currentRole)
		if req.Note != "" {
			note += ": " + req.Note
		}

		if err = addStatusHistory(rm.ID, models.StatusPending, models.StatusRejected, req.ApproverID, note); err != nil {
			return nil, err
		}
	}

	return GetReimbursementByID(id)
}

func GetReimbursementByID(id int64) (*models.Reimbursement, error) {
	query := `
		SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
		       status, current_approver, modify_count, created_at, updated_at
		FROM reimbursements
		WHERE id = ?
	`

	row := database.DB.QueryRow(query, id)
	rm, err := scanReimbursement(row)
	if err != nil {
		return nil, err
	}

	if MarkAsExpiredIfNeeded(rm) {
		now := time.Now()
		query := `UPDATE reimbursements SET status = ?, updated_at = ? WHERE id = ?`
		database.DB.Exec(query, models.StatusExpired, database.FormatTime(now), rm.ID)
		addStatusHistory(rm.ID, models.StatusPending, models.StatusExpired, "system", "超过30天未审批，自动过期")
	}

	return rm, nil
}

func GetReimbursements(employeeID string) ([]*models.Reimbursement, error) {
	var query string
	var args []interface{}

	if employeeID != "" {
		query = `
			SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
			       status, current_approver, modify_count, created_at, updated_at
			FROM reimbursements
			WHERE employee_id = ?
			ORDER BY created_at DESC
		`
		args = append(args, employeeID)
	} else {
		query = `
			SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
			       status, current_approver, modify_count, created_at, updated_at
			FROM reimbursements
			ORDER BY created_at DESC
		`
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Reimbursement
	for rows.Next() {
		rm, err := scanReimbursementRows(rows)
		if err != nil {
			return nil, err
		}

		if MarkAsExpiredIfNeeded(rm) {
			now := time.Now()
			updateQuery := `UPDATE reimbursements SET status = ?, updated_at = ? WHERE id = ?`
			database.DB.Exec(updateQuery, models.StatusExpired, database.FormatTime(now), rm.ID)
			addStatusHistory(rm.ID, models.StatusPending, models.StatusExpired, "system", "超过30天未审批，自动过期")
		}

		list = append(list, rm)
	}

	return list, rows.Err()
}

func GetStatusHistory(reimbursementID int64) ([]*models.StatusHistory, error) {
	query := `
		SELECT id, reimbursement_id, from_status, to_status, changed_by, changed_at, note
		FROM status_history
		WHERE reimbursement_id = ?
		ORDER BY changed_at ASC
	`

	rows, err := database.DB.Query(query, reimbursementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.StatusHistory
	for rows.Next() {
		var h models.StatusHistory
		var fromStatus, toStatus, changedAt, note sql.NullString

		err := rows.Scan(&h.ID, &h.ReimbursementID, &fromStatus, &toStatus, &h.ChangedBy, &changedAt, &note)
		if err != nil {
			return nil, err
		}

		if fromStatus.Valid {
			h.FromStatus = models.ReimbursementStatus(fromStatus.String)
		}
		h.ToStatus = models.ReimbursementStatus(toStatus.String)
		if changedAt.Valid {
			h.ChangedAt, _ = database.ParseTime(changedAt.String)
		}
		if note.Valid {
			h.Note = note.String
		}

		list = append(list, &h)
	}

	return list, rows.Err()
}

func addStatusHistory(reimbursementID int64, fromStatus, toStatus models.ReimbursementStatus, changedBy, note string) error {
	now := time.Now()
	query := `
		INSERT INTO status_history (reimbursement_id, from_status, to_status, changed_by, changed_at, note)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var fromStr sql.NullString
	if fromStatus != "" {
		fromStr = sql.NullString{String: string(fromStatus), Valid: true}
	}

	_, err := database.DB.Exec(query, reimbursementID, fromStr, string(toStatus), changedBy, database.FormatTime(now), note)
	return err
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanReimbursement(scanner rowScanner) (*models.Reimbursement, error) {
	rm := &models.Reimbursement{}
	var occurredDate, createdAt, updatedAt string

	err := scanner.Scan(
		&rm.ID,
		&rm.EmployeeID,
		&rm.AmountCent,
		&rm.ExpenseType,
		&occurredDate,
		&rm.Description,
		&rm.Status,
		&rm.CurrentApprover,
		&rm.ModifyCount,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	rm.AmountYuan = CentToYuan(rm.AmountCent)
	rm.OccurredDate, _ = database.ParseDate(occurredDate)
	rm.CreatedAt, _ = database.ParseTime(createdAt)
	rm.UpdatedAt, _ = database.ParseTime(updatedAt)

	return rm, nil
}

func scanReimbursementRows(rows *sql.Rows) (*models.Reimbursement, error) {
	return scanReimbursement(rows)
}
