package store

import (
	"approval-flow/pkg/db"
	"approval-flow/pkg/util"
)

type TimeoutRecord struct {
	ID            string
	ApplicationID string
	Level         int
	Reminded48h   bool
	Transferred72h bool
}

func CreateTimeoutRecord(appID string, level int) error {
	id := util.NewUUID()
	query := `INSERT INTO timeouts (id, application_id, level) VALUES (?, ?, ?)`
	_, err := db.DB.Exec(query, id, appID, level)
	return err
}

func GetTimeoutRecord(appID string, level int) (*TimeoutRecord, error) {
	query := `SELECT id, application_id, level, reminded_48h, transferred_72h FROM timeouts WHERE application_id = ? AND level = ?`
	row := db.DB.QueryRow(query, appID, level)

	record := &TimeoutRecord{}
	var reminded48h, transferred72h int
	err := row.Scan(&record.ID, &record.ApplicationID, &record.Level, &reminded48h, &transferred72h)
	if err != nil {
		return nil, err
	}
	record.Reminded48h = util.IntToBool(reminded48h)
	record.Transferred72h = util.IntToBool(transferred72h)
	return record, nil
}

func MarkReminded48h(appID string, level int) error {
	query := `UPDATE timeouts SET reminded_48h = 1, updated_at = CURRENT_TIMESTAMP WHERE application_id = ? AND level = ?`
	_, err := db.DB.Exec(query, appID, level)
	return err
}

func MarkTransferred72h(appID string, level int) error {
	query := `UPDATE timeouts SET transferred_72h = 1, updated_at = CURRENT_TIMESTAMP WHERE application_id = ? AND level = ?`
	_, err := db.DB.Exec(query, appID, level)
	return err
}

func DeleteTimeoutRecords(appID string) error {
	query := `DELETE FROM timeouts WHERE application_id = ?`
	_, err := db.DB.Exec(query, appID)
	return err
}
