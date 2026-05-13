package repository

import (
	"database/sql"
	"errors"
	"recruit-flow/database"
	"recruit-flow/models"
	"time"
)

func CreatePosition(pos *models.Position) error {
	now := time.Now()
	pos.CreatedAt = now
	pos.UpdatedAt = now
	result, err := database.DB.Exec(
		`INSERT INTO positions (name, total_quota, used_quota, pass_score, tech_threshold, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pos.Name, pos.TotalQuota, pos.UsedQuota, pos.PassScore, pos.TechThreshold, now, now,
	)
	if err != nil {
		return err
	}
	pos.ID, err = result.LastInsertId()
	return err
}

func GetPosition(id int64) (*models.Position, error) {
	pos := &models.Position{}
	err := database.DB.QueryRow(
		`SELECT id, name, total_quota, used_quota, pass_score, tech_threshold, created_at, updated_at FROM positions WHERE id = ?`, id,
	).Scan(&pos.ID, &pos.Name, &pos.TotalQuota, &pos.UsedQuota, &pos.PassScore, &pos.TechThreshold, &pos.CreatedAt, &pos.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("position not found")
	}
	return pos, err
}

func UpdatePositionQuota(id int64, delta int) error {
	_, err := database.DB.Exec(
		`UPDATE positions SET used_quota = used_quota + ?, updated_at = ? WHERE id = ?`,
		delta, time.Now(), id,
	)
	return err
}

func CreateCandidate(candidate *models.Candidate) error {
	now := time.Now()
	candidate.CreatedAt = now
	candidate.UpdatedAt = now
	result, err := database.DB.Exec(
		`INSERT INTO candidates (name, email, phone, position_id, current_stage, created_at, updated_at, last_rejected_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		candidate.Name, candidate.Email, candidate.Phone, candidate.PositionID, candidate.CurrentStage, now, now, candidate.LastRejectedAt,
	)
	if err != nil {
		return err
	}
	candidate.ID, err = result.LastInsertId()
	return err
}

func GetCandidate(id int64) (*models.Candidate, error) {
	c := &models.Candidate{}
	err := database.DB.QueryRow(
		`SELECT id, name, email, phone, position_id, current_stage, created_at, updated_at, last_rejected_at FROM candidates WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.PositionID, &c.CurrentStage, &c.CreatedAt, &c.UpdatedAt, &c.LastRejectedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("candidate not found")
	}
	return c, err
}

func GetCandidateByEmailPosition(email string, positionID int64) (*models.Candidate, error) {
	c := &models.Candidate{}
	err := database.DB.QueryRow(
		`SELECT id, name, email, phone, position_id, current_stage, created_at, updated_at, last_rejected_at FROM candidates WHERE email = ? AND position_id = ?`, email, positionID,
	).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.PositionID, &c.CurrentStage, &c.CreatedAt, &c.UpdatedAt, &c.LastRejectedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("candidate not found")
	}
	return c, err
}

func UpdateCandidateStage(candidateID int64, stage models.Stage) error {
	_, err := database.DB.Exec(
		`UPDATE candidates SET current_stage = ?, updated_at = ? WHERE id = ?`,
		stage, time.Now(), candidateID,
	)
	return err
}

func MarkCandidateRejected(candidateID int64) error {
	now := time.Now()
	_, err := database.DB.Exec(
		`UPDATE candidates SET current_stage = ?, updated_at = ?, last_rejected_at = ? WHERE id = ?`,
		models.StageRejected, now, now, candidateID,
	)
	return err
}

func CreateStageRecord(record *models.StageRecord) error {
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now
	result, err := database.DB.Exec(
		`INSERT INTO stage_records (candidate_id, stage, owner, due_date, score, status, remark, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.CandidateID, record.Stage, record.Owner, record.DueDate, record.Score, record.Status, record.Remark, now, now,
	)
	if err != nil {
		return err
	}
	record.ID, err = result.LastInsertId()
	return err
}

func GetStageRecord(candidateID int64, stage models.Stage) (*models.StageRecord, error) {
	record := &models.StageRecord{}
	err := database.DB.QueryRow(
		`SELECT id, candidate_id, stage, owner, due_date, score, status, remark, created_at, updated_at FROM stage_records WHERE candidate_id = ? AND stage = ?`,
		candidateID, stage,
	).Scan(&record.ID, &record.CandidateID, &record.Stage, &record.Owner, &record.DueDate, &record.Score, &record.Status, &record.Remark, &record.CreatedAt, &record.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("stage record not found")
	}
	return record, err
}

func UpdateStageRecordScore(candidateID int64, stage models.Stage, score float64) error {
	_, err := database.DB.Exec(
		`UPDATE stage_records SET score = ?, updated_at = ? WHERE candidate_id = ? AND stage = ?`,
		score, time.Now(), candidateID, stage,
	)
	return err
}

func UpdateStageRecordStatus(candidateID int64, stage models.Stage, status string, remark string) error {
	_, err := database.DB.Exec(
		`UPDATE stage_records SET status = ?, remark = ?, updated_at = ? WHERE candidate_id = ? AND stage = ?`,
		status, remark, time.Now(), candidateID, stage,
	)
	return err
}

func AddTechInterviewerScore(score *models.TechInterviewerScore) error {
	now := time.Now()
	score.CreatedAt = now
	score.UpdatedAt = now
	result, err := database.DB.Exec(
		`INSERT INTO tech_interviewer_scores (candidate_id, interviewer, score, submitted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		score.CandidateID, score.Interviewer, score.Score, score.Submitted, now, now,
	)
	if err != nil {
		return err
	}
	score.ID, err = result.LastInsertId()
	return err
}

func GetTechInterviewerScores(candidateID int64) ([]*models.TechInterviewerScore, error) {
	rows, err := database.DB.Query(
		`SELECT id, candidate_id, interviewer, score, submitted, created_at, updated_at FROM tech_interviewer_scores WHERE candidate_id = ?`,
		candidateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []*models.TechInterviewerScore
	for rows.Next() {
		score := &models.TechInterviewerScore{}
		err := rows.Scan(&score.ID, &score.CandidateID, &score.Interviewer, &score.Score, &score.Submitted, &score.CreatedAt, &score.UpdatedAt)
		if err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	return scores, nil
}

func SubmitTechInterviewerScore(candidateID int64, interviewer string, score float64) error {
	_, err := database.DB.Exec(
		`UPDATE tech_interviewer_scores SET score = ?, submitted = 1, updated_at = ? WHERE candidate_id = ? AND interviewer = ?`,
		score, time.Now(), candidateID, interviewer,
	)
	return err
}

func CreateOffer(offer *models.Offer) error {
	now := time.Now()
	offer.CreatedAt = now
	offer.UpdatedAt = now
	result, err := database.DB.Exec(
		`INSERT INTO offers (candidate_id, position_id, valid_until, accepted, cancelled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		offer.CandidateID, offer.PositionID, offer.ValidUntil, offer.Accepted, offer.Cancelled, now, now,
	)
	if err != nil {
		return err
	}
	offer.ID, err = result.LastInsertId()
	return err
}

func GetOffer(candidateID int64) (*models.Offer, error) {
	offer := &models.Offer{}
	err := database.DB.QueryRow(
		`SELECT id, candidate_id, position_id, valid_until, accepted, cancelled, created_at, updated_at FROM offers WHERE candidate_id = ?`,
		candidateID,
	).Scan(&offer.ID, &offer.CandidateID, &offer.PositionID, &offer.ValidUntil, &offer.Accepted, &offer.Cancelled, &offer.CreatedAt, &offer.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("offer not found")
	}
	return offer, err
}

func AcceptOffer(candidateID int64) error {
	_, err := database.DB.Exec(
		`UPDATE offers SET accepted = 1, updated_at = ? WHERE candidate_id = ?`,
		time.Now(), candidateID,
	)
	return err
}

func CancelOffer(candidateID int64) error {
	_, err := database.DB.Exec(
		`UPDATE offers SET cancelled = 1, updated_at = ? WHERE candidate_id = ?`,
		time.Now(), candidateID,
	)
	return err
}

func GetExpiredOffers() ([]*models.Offer, error) {
	now := time.Now()
	rows, err := database.DB.Query(
		`SELECT id, candidate_id, position_id, valid_until, accepted, cancelled, created_at, updated_at FROM offers WHERE valid_until < ? AND accepted = 0 AND cancelled = 0`,
		now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*models.Offer
	for rows.Next() {
		offer := &models.Offer{}
		err := rows.Scan(&offer.ID, &offer.CandidateID, &offer.PositionID, &offer.ValidUntil, &offer.Accepted, &offer.Cancelled, &offer.CreatedAt, &offer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func AddHistoryRecord(record *models.HistoryRecord) error {
	now := time.Now()
	record.CreatedAt = now
	_, err := database.DB.Exec(
		`INSERT INTO history_records (candidate_id, stage, action, operator, remark, score, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.CandidateID, record.Stage, record.Action, record.Operator, record.Remark, record.Score, now,
	)
	return err
}

func GetCandidateHistory(candidateID int64) ([]*models.HistoryRecord, error) {
	rows, err := database.DB.Query(
		`SELECT id, candidate_id, stage, action, operator, remark, score, created_at FROM history_records WHERE candidate_id = ? ORDER BY created_at DESC`,
		candidateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.HistoryRecord
	for rows.Next() {
		record := &models.HistoryRecord{}
		err := rows.Scan(&record.ID, &record.CandidateID, &record.Stage, &record.Action, &record.Operator, &record.Remark, &record.Score, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func GetOverdueStageRecords() ([]*models.StageRecord, error) {
	now := time.Now()
	rows, err := database.DB.Query(
		`SELECT id, candidate_id, stage, owner, due_date, score, status, remark, created_at, updated_at FROM stage_records WHERE due_date < ? AND (status IS NULL OR status = '')`,
		now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.StageRecord
	for rows.Next() {
		record := &models.StageRecord{}
		err := rows.Scan(&record.ID, &record.CandidateID, &record.Stage, &record.Owner, &record.DueDate, &record.Score, &record.Status, &record.Remark, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
