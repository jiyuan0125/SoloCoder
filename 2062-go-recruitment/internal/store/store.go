package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"recruitment/internal/models"
)

type Store struct {
	dataFile string
	mu       sync.RWMutex
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dataFile := filepath.Join(dataDir, "recruitment.json")

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		db := models.NewDatabase()
		data, err := json.MarshalIndent(db, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal initial database: %w", err)
		}
		if err := os.WriteFile(dataFile, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write initial database: %w", err)
		}
	}

	return &Store{dataFile: dataFile}, nil
}

func (s *Store) Load() (*models.Database, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read database file: %w", err)
	}

	var db models.Database
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to parse database: %w", err)
	}

	return &db, nil
}

func (s *Store) Save(db *models.Database) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validate(db); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	if err := os.WriteFile(s.dataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write database: %w", err)
	}

	return nil
}

func (s *Store) validate(db *models.Database) error {
	jobIDs := make(map[string]bool)
	for _, job := range db.Jobs {
		if job.ID == "" {
			return fmt.Errorf("job with empty ID found")
		}
		if jobIDs[job.ID] {
			return fmt.Errorf("duplicate job ID: %s", job.ID)
		}
		jobIDs[job.ID] = true
	}

	candidateIDs := make(map[string]bool)
	candidateKeyMap := make(map[string]bool)
	for _, cand := range db.Candidates {
		if cand.ID == "" {
			return fmt.Errorf("candidate with empty ID found")
		}
		if candidateIDs[cand.ID] {
			return fmt.Errorf("duplicate candidate ID: %s", cand.ID)
		}
		if !jobIDs[cand.JobID] {
			return fmt.Errorf("candidate %s references non-existent job: %s", cand.ID, cand.JobID)
		}

		key := fmt.Sprintf("%s_%s", cand.JobID, cand.ResumeID)
		if candidateKeyMap[key] {
			return fmt.Errorf("duplicate candidate: resume %s for job %s", cand.ResumeID, cand.JobID)
		}
		candidateKeyMap[key] = true
		candidateIDs[cand.ID] = true
	}

	interviewIDs := make(map[string]bool)
	for _, interview := range db.Interviews {
		if interview.ID == "" {
			return fmt.Errorf("interview with empty ID found")
		}
		if interviewIDs[interview.ID] {
			return fmt.Errorf("duplicate interview ID: %s", interview.ID)
		}
		if !candidateIDs[interview.CandidateID] {
			return fmt.Errorf("interview %s references non-existent candidate: %s", interview.ID, interview.CandidateID)
		}
		if !jobIDs[interview.JobID] {
			return fmt.Errorf("interview %s references non-existent job: %s", interview.ID, interview.JobID)
		}
		interviewIDs[interview.ID] = true
	}

	offerIDs := make(map[string]bool)
	for _, offer := range db.Offers {
		if offer.ID == "" {
			return fmt.Errorf("offer with empty ID found")
		}
		if offerIDs[offer.ID] {
			return fmt.Errorf("duplicate offer ID: %s", offer.ID)
		}
		if !candidateIDs[offer.CandidateID] {
			return fmt.Errorf("offer %s references non-existent candidate: %s", offer.ID, offer.CandidateID)
		}
		if !jobIDs[offer.JobID] {
			return fmt.Errorf("offer %s references non-existent job: %s", offer.ID, offer.JobID)
		}
		offerIDs[offer.ID] = true
	}

	budgetJobIDs := make(map[string]bool)
	for _, budget := range db.Budgets {
		if budget.JobID == "" {
			return fmt.Errorf("budget with empty job ID found")
		}
		if !jobIDs[budget.JobID] {
			return fmt.Errorf("budget references non-existent job: %s", budget.JobID)
		}
		if budgetJobIDs[budget.JobID] {
			return fmt.Errorf("duplicate budget for job: %s", budget.JobID)
		}
		budgetJobIDs[budget.JobID] = true
	}

	return nil
}

func (s *Store) Transact(fn func(*models.Database) error) error {
	db, err := s.Load()
	if err != nil {
		return err
	}

	if err := fn(db); err != nil {
		return err
	}

	return s.Save(db)
}

func (s *Store) ValidateConsistency() error {
	db, err := s.Load()
	if err != nil {
		return err
	}

	return s.validate(db)
}

func (s *Store) TriggerUpdateCheck(jobID string) error {
	db, err := s.Load()
	if err != nil {
		return err
	}

	relatedJobIDs := map[string]bool{jobID: true}

	for _, cand := range db.Candidates {
		if cand.JobID == jobID {
			relatedJobIDs[jobID] = true
		}
	}

	for _, interview := range db.Interviews {
		if interview.JobID == jobID {
			relatedJobIDs[jobID] = true
		}
	}

	for _, offer := range db.Offers {
		if offer.JobID == jobID {
			relatedJobIDs[jobID] = true
		}
	}

	now := time.Now()
	for i := range db.Offers {
		offer := &db.Offers[i]
		if offer.Status == models.OfferStatusPending && now.After(offer.ValidUntil) {
			offer.Status = models.OfferStatusAbandoned
			for j := range db.Candidates {
				if db.Candidates[j].ID == offer.CandidateID {
					db.Candidates[j].Status = models.CandidateStatusOfferAbandoned
					db.Candidates[j].RejectionReason = "Offer 超时未回复"
					rejectedAt := now
					db.Candidates[j].RejectedAt = &rejectedAt
				}
			}
		}
	}

	return s.Save(db)
}
