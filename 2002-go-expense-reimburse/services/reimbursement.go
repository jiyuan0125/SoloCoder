package services

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
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

func BuildApprovalChain(amountCent int64) []models.ApprovalRole {
	amountYuan := CentToYuan(amountCent)
	var chain []models.ApprovalRole

	chain = append(chain, models.RoleManager)

	if amountYuan >= 500 {
		chain = append(chain, models.RoleDepartment)
	}

	if amountYuan >= 5000 {
		chain = append(chain, models.RoleCFO)
	}

	return chain
}

func ApprovalChainToString(chain []models.ApprovalRole) string {
	strs := make([]string, len(chain))
	for i, role := range chain {
		strs[i] = string(role)
	}
	return strings.Join(strs, ",")
}

func StringToApprovalChain(s string) []models.ApprovalRole {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	chain := make([]models.ApprovalRole, 0, len(parts))
	for _, part := range parts {
		chain = append(chain, models.ApprovalRole(part))
	}
	return chain
}

func CalculateStageAmounts(amountCent int64, stageCount int) []int64 {
	if stageCount == 0 {
		return nil
	}
	baseAmount := amountCent / int64(stageCount)
	remainder := amountCent % int64(stageCount)

	amounts := make([]int64, stageCount)
	for i := 0; i < stageCount; i++ {
		amounts[i] = baseAmount
		if i == stageCount-1 {
			amounts[i] += remainder
		}
	}
	return amounts
}

func CheckDuplicate(employeeID string, occurredDate time.Time, amountCent int64, expenseType models.ExpenseType, excludeID int64) (*models.Reimbursement, error) {
	occurredDateStr := database.FormatDate(occurredDate)

	query := `
		SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
		       status, current_approval_step, approval_chain, modify_count, created_at, updated_at
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

	approvalChain := BuildApprovalChain(amountCent)
	chainStr := ApprovalChainToString(approvalChain)
	now := time.Now()

	query := `
		INSERT INTO reimbursements 
		(employee_id, amount_cent, expense_type, occurred_date, description, 
		 status, current_approval_step, approval_chain, modify_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query,
		req.EmployeeID,
		amountCent,
		req.ExpenseType,
		database.FormatDate(occurredDate),
		req.Description,
		models.StatusPending,
		0,
		chainStr,
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

	stageAmounts := CalculateStageAmounts(amountCent, len(approvalChain))
	for i, role := range approvalChain {
		_, err = database.DB.Exec(
			`INSERT INTO stages (reimbursement_id, approval_role, amount_cent, approval_step) VALUES (?, ?, ?, ?)`,
			id, string(role), stageAmounts[i], i,
		)
		if err != nil {
			return nil, err
		}
	}

	note := fmt.Sprintf("提交成功，审批链: %s，当前审批: %s", chainStr, approvalChain[0])
	if err = addStatusHistory(id, "", models.StatusPending, "system", note); err != nil {
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

	newApprovalChain := BuildApprovalChain(newAmountCent)
	newChainStr := ApprovalChainToString(newApprovalChain)
	now := time.Now()

	query := `
		UPDATE reimbursements 
		SET amount_cent = ?, approval_chain = ?, current_approval_step = ?, modify_count = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = database.DB.Exec(query, newAmountCent, newChainStr, 0, rm.ModifyCount+1, models.StatusPending, database.FormatTime(now), rm.ID)
	if err != nil {
		return nil, err
	}

	_, err = database.DB.Exec(`DELETE FROM stages WHERE reimbursement_id = ?`, rm.ID)
	if err != nil {
		return nil, err
	}

	stageAmounts := CalculateStageAmounts(newAmountCent, len(newApprovalChain))
	for i, role := range newApprovalChain {
		_, err = database.DB.Exec(
			`INSERT INTO stages (reimbursement_id, approval_role, amount_cent, approval_step) VALUES (?, ?, ?, ?)`,
			rm.ID, string(role), stageAmounts[i], i,
		)
		if err != nil {
			return nil, err
		}
	}

	note := fmt.Sprintf("修改金额: %.2f元 -> %.2f元，新审批链: %s", CentToYuan(oldAmountCent), CentToYuan(newAmountCent), newChainStr)
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

	if len(rm.ApprovalChain) == 0 {
		return nil, fmt.Errorf("审批链为空")
	}

	currentStep := rm.CurrentApprovalStep
	if currentStep >= len(rm.ApprovalChain) {
		return nil, fmt.Errorf("审批已完成")
	}

	currentApprover := rm.ApprovalChain[currentStep]
	if req.ApproverRole != string(currentApprover) {
		return nil, fmt.Errorf("当前审批人为 %s，无权审批", currentApprover)
	}

	now := time.Now()

	if req.Action == "approve" {
		nextStep := currentStep + 1
		hasNext := nextStep < len(rm.ApprovalChain)

		var newStatus models.ReimbursementStatus
		var note string

		if hasNext {
			newStatus = models.StatusPending
			nextApprover := rm.ApprovalChain[nextStep]
			note = fmt.Sprintf("%s审批通过，流转至 %s", currentApprover, nextApprover)
		} else {
			newStatus = models.StatusCompleted
			note = fmt.Sprintf("%s审批通过，流程完成", currentApprover)
		}

		query := `UPDATE reimbursements SET status = ?, current_approval_step = ?, updated_at = ? WHERE id = ?`
		_, err = database.DB.Exec(query, newStatus, nextStep, database.FormatTime(now), rm.ID)
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

		note := fmt.Sprintf("%s驳回", currentApprover)
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
		       status, current_approval_step, approval_chain, modify_count, created_at, updated_at
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

	stages, err := GetStagesByReimbursementID(id)
	if err != nil {
		return nil, err
	}
	rm.Stages = stages

	return rm, nil
}

func GetStagesByReimbursementID(reimbursementID int64) ([]*models.Stage, error) {
	query := `
		SELECT id, reimbursement_id, approval_role, amount_cent, approval_step
		FROM stages
		WHERE reimbursement_id = ?
		ORDER BY approval_step ASC
	`

	rows, err := database.DB.Query(query, reimbursementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []*models.Stage
	for rows.Next() {
		var s models.StageDB
		err := rows.Scan(&s.ID, &s.ReimbursementID, &s.ApprovalRole, &s.AmountCent, &s.ApprovalStep)
		if err != nil {
			return nil, err
		}

		stage := &models.Stage{
			ID:              s.ID,
			ReimbursementID: s.ReimbursementID,
			ApprovalRole:     models.ApprovalRole(s.ApprovalRole),
			AmountCent:     s.AmountCent,
			AmountYuan:   CentToYuan(s.AmountCent),
			ApprovalStep: s.ApprovalStep,
		}
		stages = append(stages, stage)
	}

	return stages, rows.Err()
}

func GetReimbursements(employeeID string) ([]*models.Reimbursement, error) {
	var query string
	var args []interface{}

	if employeeID != "" {
		query = `
			SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
			       status, current_approval_step, approval_chain, modify_count, created_at, updated_at
			FROM reimbursements
			WHERE employee_id = ?
			ORDER BY created_at DESC
		`
		args = append(args, employeeID)
	} else {
		query = `
			SELECT id, employee_id, amount_cent, expense_type, occurred_date, description,
			       status, current_approval_step, approval_chain, modify_count, created_at, updated_at
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

		stages, err := GetStagesByReimbursementID(rm.ID)
		if err != nil {
			return nil, err
		}
		rm.Stages = stages

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
	var occurredDate, createdAt, updatedAt, approvalChainStr string

	err := scanner.Scan(
		&rm.ID,
		&rm.EmployeeID,
		&rm.AmountCent,
		&rm.ExpenseType,
		&occurredDate,
		&rm.Description,
		&rm.Status,
		&rm.CurrentApprovalStep,
		&approvalChainStr,
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
	rm.ApprovalChain = StringToApprovalChain(approvalChainStr)

	return rm, nil
}

func scanReimbursementRows(rows *sql.Rows) (*models.Reimbursement, error) {
	return scanReimbursement(rows)
}
