package service

import (
	"errors"
	"fmt"
	"recruit-flow/models"
	"recruit-flow/repository"
	"time"
)

type ServiceError struct {
	Code    int
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(code int, msg string) error {
	return &ServiceError{Code: code, Message: msg}
}

func getNextStage(currentStage models.Stage) (models.Stage, bool) {
	for i, stage := range models.StageOrder {
		if stage == currentStage {
			if i+1 < len(models.StageOrder) {
				return models.StageOrder[i+1], true
			}
			return currentStage, false
		}
	}
	return currentStage, false
}

func CreateCandidate(name, email, phone string, positionID int64, operator string, interviewers []string) (*models.Candidate, error) {
	_, err := repository.GetPosition(positionID)
	if err != nil {
		return nil, NewServiceError(404, "position not found")
	}

	if email != "" {
		existing, _ := repository.GetCandidateByEmailPosition(email, positionID)
		if existing != nil {
			if existing.LastRejectedAt != nil {
				sixMonthsAgo := time.Now().AddDate(0, -6, 0)
				if existing.LastRejectedAt.After(sixMonthsAgo) {
					return nil, NewServiceError(409, "candidate was rejected within 6 months, cannot reapply")
				}
			}
		}
	}

	candidate := &models.Candidate{
		Name:         name,
		Email:        email,
		Phone:        phone,
		PositionID:   positionID,
		CurrentStage: models.StageResumeScreening,
	}
	if err := repository.CreateCandidate(candidate); err != nil {
		return nil, err
	}

	stageRecord := &models.StageRecord{
		CandidateID: candidate.ID,
		Stage:       models.StageResumeScreening,
		Owner:       operator,
		DueDate:     time.Now().AddDate(0, 0, 7),
	}
	if err := repository.CreateStageRecord(stageRecord); err != nil {
		return nil, err
	}

	if len(interviewers) > 0 {
		for _, interviewer := range interviewers {
			score := &models.TechInterviewerScore{
				CandidateID: candidate.ID,
				Interviewer: interviewer,
				Submitted:   false,
			}
			if err := repository.AddTechInterviewerScore(score); err != nil {
				return nil, err
			}
		}
	}

	_ = repository.AddHistoryRecord(&models.HistoryRecord{
		CandidateID: candidate.ID,
		Stage:       candidate.CurrentStage,
		Action:      "CREATE",
		Operator:    operator,
		Remark:      "Candidate created",
	})

	return candidate, nil
}

func GetCandidate(id int64) (*models.Candidate, error) {
	candidate, err := repository.GetCandidate(id)
	if err != nil {
		return nil, NewServiceError(404, "candidate not found")
	}
	return candidate, nil
}

func AdvanceStage(candidateID int64, operator string, extraData map[string]interface{}) (models.Stage, error) {
	candidate, err := repository.GetCandidate(candidateID)
	if err != nil {
		return "", NewServiceError(404, "candidate not found")
	}

	if candidate.CurrentStage == models.StageRejected || candidate.CurrentStage == models.StageOfferExpired {
		return candidate.CurrentStage, NewServiceError(400, "cannot advance rejected or expired candidate")
	}

	nextStage, hasNext := getNextStage(candidate.CurrentStage)
	if !hasNext {
		return candidate.CurrentStage, NewServiceError(400, "already at final stage")
	}

	switch candidate.CurrentStage {
	case models.StageWrittenTest:
		if extraData == nil {
			return candidate.CurrentStage, NewServiceError(400, "written test score is required")
		}
		scoreVal, ok := extraData["score"]
		if !ok {
			return candidate.CurrentStage, NewServiceError(400, "written test score is required")
		}
		var score float64
		switch v := scoreVal.(type) {
		case float64:
			score = v
		case int:
			score = float64(v)
		default:
			return candidate.CurrentStage, NewServiceError(400, "invalid score format")
		}

		if score < 0 || score > 100 {
			return candidate.CurrentStage, NewServiceError(400, "score must be between 0 and 100")
		}

		position, err := repository.GetPosition(candidate.PositionID)
		if err != nil {
			return candidate.CurrentStage, err
		}

		_ = repository.UpdateStageRecordScore(candidateID, candidate.CurrentStage, score)
		_ = repository.AddHistoryRecord(&models.HistoryRecord{
			CandidateID: candidateID,
			Stage:       candidate.CurrentStage,
			Action:      "SCORE_SUBMITTED",
			Operator:    operator,
			Score:       &score,
		})

		if score < position.PassScore {
			_ = repository.MarkCandidateRejected(candidateID)
			_ = repository.UpdateStageRecordStatus(candidateID, candidate.CurrentStage, "rejected", fmt.Sprintf("Written test score %.1f below pass score %.1f", score, position.PassScore))
			_ = repository.AddHistoryRecord(&models.HistoryRecord{
				CandidateID: candidateID,
				Stage:       candidate.CurrentStage,
				Action:      "REJECTED",
				Operator:    "SYSTEM",
				Remark:      fmt.Sprintf("Written test score %.1f below pass score %.1f", score, position.PassScore),
				Score:       &score,
			})
			return models.StageRejected, nil
		}

	case models.StageTechInterview:
		scores, err := repository.GetTechInterviewerScores(candidateID)
		if err != nil {
			return candidate.CurrentStage, err
		}

		if len(scores) == 0 {
			return candidate.CurrentStage, NewServiceError(400, "no interviewers assigned")
		}

		var totalScore float64
		allSubmitted := true
		for _, s := range scores {
			if !s.Submitted {
				allSubmitted = false
				break
			}
			totalScore += s.Score
		}

		if !allSubmitted {
			return candidate.CurrentStage, NewServiceError(400, "not all interviewers have submitted scores")
		}

		position, err := repository.GetPosition(candidate.PositionID)
		if err != nil {
			return candidate.CurrentStage, err
		}

		avgScore := totalScore / float64(len(scores))
		_ = repository.UpdateStageRecordScore(candidateID, candidate.CurrentStage, avgScore)
		_ = repository.AddHistoryRecord(&models.HistoryRecord{
			CandidateID: candidateID,
			Stage:       candidate.CurrentStage,
			Action:      "SCORE_SUBMITTED",
			Operator:    operator,
			Score:       &avgScore,
		})

		if avgScore < position.TechThreshold {
			_ = repository.MarkCandidateRejected(candidateID)
			_ = repository.UpdateStageRecordStatus(candidateID, candidate.CurrentStage, "rejected", fmt.Sprintf("Average tech score %.1f below threshold %.1f", avgScore, position.TechThreshold))
			_ = repository.AddHistoryRecord(&models.HistoryRecord{
				CandidateID: candidateID,
				Stage:       candidate.CurrentStage,
				Action:      "REJECTED",
				Operator:    "SYSTEM",
				Remark:      fmt.Sprintf("Average tech score %.1f below threshold %.1f", avgScore, position.TechThreshold),
				Score:       &avgScore,
			})
			return models.StageRejected, nil
		}

	case models.StageHRInterview:
		position, err := repository.GetPosition(candidate.PositionID)
		if err != nil {
			return candidate.CurrentStage, err
		}

		if position.UsedQuota >= position.TotalQuota {
			return candidate.CurrentStage, NewServiceError(400, "no quota available for this position")
		}

		offer := &models.Offer{
			CandidateID: candidateID,
			PositionID:  candidate.PositionID,
			ValidUntil:  time.Now().AddDate(0, 0, 7),
			Accepted:    false,
			Cancelled:   false,
		}
		if err := repository.CreateOffer(offer); err != nil {
			return candidate.CurrentStage, err
		}

		if err := repository.UpdatePositionQuota(candidate.PositionID, 1); err != nil {
			return candidate.CurrentStage, err
		}

		_ = repository.AddHistoryRecord(&models.HistoryRecord{
			CandidateID: candidateID,
			Stage:       models.StageOffer,
			Action:      "OFFER_CREATED",
			Operator:    operator,
			Remark:      fmt.Sprintf("Offer valid until %s", offer.ValidUntil.Format("2006-01-02")),
		})

	case models.StageOffer:
		offer, err := repository.GetOffer(candidateID)
		if err != nil {
			return candidate.CurrentStage, NewServiceError(400, "no offer found")
		}
		if offer.Cancelled {
			return candidate.CurrentStage, NewServiceError(400, "offer has been cancelled")
		}
		if time.Now().After(offer.ValidUntil) {
			return candidate.CurrentStage, NewServiceError(400, "offer has expired")
		}
		if err := repository.AcceptOffer(candidateID); err != nil {
			return candidate.CurrentStage, err
		}

		_ = repository.AddHistoryRecord(&models.HistoryRecord{
			CandidateID: candidateID,
			Stage:       models.StageOffer,
			Action:      "OFFER_ACCEPTED",
			Operator:    operator,
		})
	}

	if err := repository.UpdateCandidateStage(candidateID, nextStage); err != nil {
		return candidate.CurrentStage, err
	}

	var owner string
	if extraData != nil {
		if ow, ok := extraData["owner"].(string); ok {
			owner = ow
		}
	}
	if owner == "" {
		owner = operator
	}

	stageRecord := &models.StageRecord{
		CandidateID: candidateID,
		Stage:       nextStage,
		Owner:       owner,
		DueDate:     time.Now().AddDate(0, 0, 7),
	}
	if err := repository.CreateStageRecord(stageRecord); err != nil {
		return candidate.CurrentStage, err
	}

	_ = repository.AddHistoryRecord(&models.HistoryRecord{
		CandidateID: candidateID,
		Stage:       nextStage,
		Action:      "STAGE_ADVANCED",
		Operator:    operator,
		Remark:      fmt.Sprintf("Advanced from %s to %s", candidate.CurrentStage, nextStage),
	})

	return nextStage, nil
}

func RejectCandidate(candidateID int64, reason, operator string) error {
	candidate, err := repository.GetCandidate(candidateID)
	if err != nil {
		return NewServiceError(404, "candidate not found")
	}

	if reason == "" {
		return NewServiceError(400, "rejection reason is required")
	}

	if err := repository.MarkCandidateRejected(candidateID); err != nil {
		return err
	}

	if candidate.CurrentStage != models.StageRejected && candidate.CurrentStage != models.StageOfferExpired {
		_ = repository.UpdateStageRecordStatus(candidateID, candidate.CurrentStage, "rejected", reason)
	}

	_ = repository.AddHistoryRecord(&models.HistoryRecord{
		CandidateID: candidateID,
		Stage:       candidate.CurrentStage,
		Action:      "REJECTED",
		Operator:    operator,
		Remark:      reason,
	})

	return nil
}

func SubmitTechInterviewScore(candidateID int64, interviewer string, score float64, operator string) error {
	if score < 0 || score > 100 {
		return NewServiceError(400, "score must be between 0 and 100")
	}

	candidate, err := repository.GetCandidate(candidateID)
	if err != nil {
		return NewServiceError(404, "candidate not found")
	}

	if candidate.CurrentStage != models.StageTechInterview {
		return NewServiceError(400, "candidate not in tech interview stage")
	}

	scores, err := repository.GetTechInterviewerScores(candidateID)
	if err != nil {
		return err
	}

	found := false
	for _, s := range scores {
		if s.Interviewer == interviewer {
			found = true
			break
		}
	}

	if !found {
		return NewServiceError(400, "interviewer not assigned to this candidate")
	}

	if err := repository.SubmitTechInterviewerScore(candidateID, interviewer, score); err != nil {
		return err
	}

	_ = repository.AddHistoryRecord(&models.HistoryRecord{
		CandidateID: candidateID,
		Stage:       models.StageTechInterview,
		Action:      "INTERVIEWER_SCORE",
		Operator:    operator,
		Remark:      fmt.Sprintf("Interviewer %s scored %.1f", interviewer, score),
		Score:       &score,
	})

	return nil
}

func ProcessExpiredOffers() error {
	expiredOffers, err := repository.GetExpiredOffers()
	if err != nil {
		return err
	}

	for _, offer := range expiredOffers {
		if err := repository.CancelOffer(offer.CandidateID); err != nil {
			continue
		}

		candidate, err := repository.GetCandidate(offer.CandidateID)
		if err != nil {
			continue
		}

		if candidate.CurrentStage != models.StageOfferExpired && candidate.CurrentStage != models.StageOnboarding {
			_ = repository.UpdateCandidateStage(offer.CandidateID, models.StageOfferExpired)
			_ = repository.UpdatePositionQuota(offer.PositionID, -1)
			_ = repository.AddHistoryRecord(&models.HistoryRecord{
				CandidateID: offer.CandidateID,
				Stage:       models.StageOffer,
				Action:      "OFFER_EXPIRED",
				Operator:    "SYSTEM",
				Remark:      "Offer expired and cancelled, quota released",
			})
		}
	}

	return nil
}

func CheckOverdueStages() error {
	records, err := repository.GetOverdueStageRecords()
	if err != nil {
		return err
	}

	for _, record := range records {
		candidate, err := repository.GetCandidate(record.CandidateID)
		if err != nil {
			continue
		}
		if candidate.CurrentStage != record.Stage {
			continue
		}
		fmt.Printf("REMINDER: Candidate %d at stage %s is overdue (owner: %s, due: %s)\n",
			record.CandidateID, record.Stage, record.Owner, record.DueDate.Format("2006-01-02"))
	}

	return nil
}

func GetCandidateHistory(candidateID int64) ([]*models.HistoryRecord, error) {
	_, err := repository.GetCandidate(candidateID)
	if err != nil {
		return nil, NewServiceError(404, "candidate not found")
	}
	return repository.GetCandidateHistory(candidateID)
}

func ValidateData(candidateID int64) error {
	candidate, err := repository.GetCandidate(candidateID)
	if err != nil {
		return NewServiceError(404, "candidate not found")
	}

	_, _ = repository.GetStageRecord(candidateID, candidate.CurrentStage)

	_, _ = repository.GetCandidateHistory(candidateID)

	if candidate.CurrentStage == models.StageTechInterview {
		_, _ = repository.GetTechInterviewerScores(candidateID)
	}

	if candidate.CurrentStage == models.StageOffer || candidate.CurrentStage == models.StageOnboarding {
		_, _ = repository.GetOffer(candidateID)
	}

	if candidate.PositionID > 0 {
		position, err := repository.GetPosition(candidate.PositionID)
		if err == nil {
			if position.UsedQuota < 0 || position.UsedQuota > position.TotalQuota {
				return NewServiceError(500, "quota consistency error")
			}
		}
	}

	return nil
}

func GetServiceErrorCode(err error) int {
	var se *ServiceError
	if errors.As(err, &se) {
		return se.Code
	}
	return 500
}
