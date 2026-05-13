package store

import (
	"database/sql"
	"errors"

	"approval-flow/pkg/db"
	"approval-flow/pkg/model"
	"approval-flow/pkg/util"
)

var ErrApplicationNotFound = errors.New("application not found")

func CreateApplication(app *model.Application) error {
	if app.ID == "" {
		app.ID = util.NewUUID()
	}

	dataJSON := util.MapToJSON(app.Data)
	approverIDsJSON := util.SliceToJSON(app.CurrentApproverIDs)

	query := `INSERT INTO applications (id, chain_id, applicant_id, title, description, data, current_level, current_approver_ids, status, submission_count) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.DB.Exec(query, app.ID, app.ChainID, app.ApplicantID, app.Title, app.Description, dataJSON, app.CurrentLevel, approverIDsJSON, app.Status, app.SubmissionCount)
	return err
}

func UpdateApplication(app *model.Application) error {
	dataJSON := util.MapToJSON(app.Data)
	approverIDsJSON := util.SliceToJSON(app.CurrentApproverIDs)

	query := `UPDATE applications SET title = ?, description = ?, data = ?, current_level = ?, current_approver_ids = ?, status = ?, reject_reason = ?, submission_count = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.DB.Exec(query, app.Title, app.Description, dataJSON, app.CurrentLevel, approverIDsJSON, app.Status, app.RejectReason, app.SubmissionCount, app.ID)
	return err
}

func GetApplicationByID(id string) (*model.Application, error) {
	query := `SELECT id, chain_id, applicant_id, title, description, data, current_level, current_approver_ids, status, reject_reason, submission_count, created_at, updated_at FROM applications WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	app := &model.Application{}
	var dataJSON, approverIDsJSON string
	var rejectReason sql.NullString
	err := row.Scan(&app.ID, &app.ChainID, &app.ApplicantID, &app.Title, &app.Description, &dataJSON, &app.CurrentLevel, &approverIDsJSON, &app.Status, &rejectReason, &app.SubmissionCount, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}

	app.Data, _ = util.JSONToMap(dataJSON)
	app.CurrentApproverIDs, _ = util.JSONToSlice[string](approverIDsJSON)
	if rejectReason.Valid {
		app.RejectReason = rejectReason.String
	}

	return app, nil
}

func ListApplications() ([]*model.Application, error) {
	query := `SELECT id, chain_id, applicant_id, title, description, data, current_level, current_approver_ids, status, reject_reason, submission_count, created_at, updated_at FROM applications ORDER BY created_at DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*model.Application{}
	for rows.Next() {
		app := &model.Application{}
		var dataJSON, approverIDsJSON string
		var rejectReason sql.NullString
		err := rows.Scan(&app.ID, &app.ChainID, &app.ApplicantID, &app.Title, &app.Description, &dataJSON, &app.CurrentLevel, &approverIDsJSON, &app.Status, &rejectReason, &app.SubmissionCount, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, err
		}
		app.Data, _ = util.JSONToMap(dataJSON)
		app.CurrentApproverIDs, _ = util.JSONToSlice[string](approverIDsJSON)
		if rejectReason.Valid {
			app.RejectReason = rejectReason.String
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func ListApplicationsByUser(userID string) ([]*model.Application, error) {
	query := `SELECT id, chain_id, applicant_id, title, description, data, current_level, current_approver_ids, status, reject_reason, submission_count, created_at, updated_at FROM applications WHERE applicant_id = ? ORDER BY created_at DESC`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*model.Application{}
	for rows.Next() {
		app := &model.Application{}
		var dataJSON, approverIDsJSON string
		var rejectReason sql.NullString
		err := rows.Scan(&app.ID, &app.ChainID, &app.ApplicantID, &app.Title, &app.Description, &dataJSON, &app.CurrentLevel, &approverIDsJSON, &app.Status, &rejectReason, &app.SubmissionCount, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, err
		}
		app.Data, _ = util.JSONToMap(dataJSON)
		app.CurrentApproverIDs, _ = util.JSONToSlice[string](approverIDsJSON)
		if rejectReason.Valid {
			app.RejectReason = rejectReason.String
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func ListPendingApplications() ([]*model.Application, error) {
	query := `SELECT id, chain_id, applicant_id, title, description, data, current_level, current_approver_ids, status, reject_reason, submission_count, created_at, updated_at FROM applications WHERE status = 'pending' ORDER BY created_at ASC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*model.Application{}
	for rows.Next() {
		app := &model.Application{}
		var dataJSON, approverIDsJSON string
		var rejectReason sql.NullString
		err := rows.Scan(&app.ID, &app.ChainID, &app.ApplicantID, &app.Title, &app.Description, &dataJSON, &app.CurrentLevel, &approverIDsJSON, &app.Status, &rejectReason, &app.SubmissionCount, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, err
		}
		app.Data, _ = util.JSONToMap(dataJSON)
		app.CurrentApproverIDs, _ = util.JSONToSlice[string](approverIDsJSON)
		if rejectReason.Valid {
			app.RejectReason = rejectReason.String
		}
		apps = append(apps, app)
	}
	return apps, nil
}
