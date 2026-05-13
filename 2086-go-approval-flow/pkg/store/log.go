package store

import (
	"approval-flow/pkg/db"
	"approval-flow/pkg/model"
	"approval-flow/pkg/util"
)

func CreateOperationLog(op *model.ApprovalOperation) error {
	if op.ID == "" {
		op.ID = util.NewUUID()
	}

	query := `INSERT INTO approval_operations (id, application_id, operator_id, level, operation, reason, target_user_id) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, op.ID, op.ApplicationID, op.OperatorID, op.Level, op.Operation, op.Reason, op.TargetUserID)
	return err
}

func GetOperationsByApplication(appID string) ([]*model.ApprovalOperation, error) {
	query := `SELECT id, application_id, operator_id, level, operation, reason, target_user_id, created_at FROM approval_operations WHERE application_id = ? ORDER BY created_at ASC`
	rows, err := db.DB.Query(query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ops := []*model.ApprovalOperation{}
	for rows.Next() {
		op := &model.ApprovalOperation{}
		var reason, targetUserID *string
		err := rows.Scan(&op.ID, &op.ApplicationID, &op.OperatorID, &op.Level, &op.Operation, &reason, &targetUserID, &op.CreatedAt)
		if err != nil {
			return nil, err
		}
		if reason != nil {
			op.Reason = *reason
		}
		if targetUserID != nil {
			op.TargetUserID = *targetUserID
		}
		ops = append(ops, op)
	}
	return ops, nil
}
