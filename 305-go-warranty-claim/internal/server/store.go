package server

import (
	"encoding/json"
	"os"
	"sync"
	"warranty-claim/pkg/common"
)

type Store struct {
	applications map[string]common.WarrantyApplication
	appeals      map[string]common.Appeal
	appealsByApp map[string][]string
	dataFile     string
	mu           sync.RWMutex
}

func NewStore(dataFile string) *Store {
	s := &Store{
		applications: make(map[string]common.WarrantyApplication),
		appeals:      make(map[string]common.Appeal),
		appealsByApp: make(map[string][]string),
		dataFile:     dataFile,
	}
	s.Load()
	return s
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := map[string]interface{}{
		"applications": s.applications,
		"appeals":      s.appeals,
		"appealsByApp": s.appealsByApp,
	}

	file, err := os.Create(s.dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.dataFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	var data struct {
		Applications map[string]common.WarrantyApplication `json:"applications"`
		Appeals      map[string]common.Appeal              `json:"appeals"`
		AppealsByApp map[string][]string                   `json:"appealsByApp"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	if data.Applications != nil {
		s.applications = data.Applications
	}
	if data.Appeals != nil {
		s.appeals = data.Appeals
	}
	if data.AppealsByApp != nil {
		s.appealsByApp = data.AppealsByApp
	}

	return nil
}

func (s *Store) CreateApplication(app common.WarrantyApplication) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.applications[app.ID] = app
	return s.Save()
}

func (s *Store) GetApplication(id string) (common.WarrantyApplication, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, exists := s.applications[id]
	return app, exists
}

func (s *Store) GetApplicationsByUser(userID string) []common.WarrantyApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var apps []common.WarrantyApplication
	for _, app := range s.applications {
		if app.UserID == userID {
			apps = append(apps, app)
		}
	}
	return apps
}

func (s *Store) GetAllApplications() []common.WarrantyApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var apps []common.WarrantyApplication
	for _, app := range s.applications {
		apps = append(apps, app)
	}
	return apps
}

func (s *Store) GetPendingApplications() []common.WarrantyApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var apps []common.WarrantyApplication
	for _, app := range s.applications {
		if app.Status == common.ApplicationStatusPending || app.Status == common.ApplicationStatusAutoApproved {
			apps = append(apps, app)
		}
	}
	return apps
}

func (s *Store) UpdateApplication(app common.WarrantyApplication) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.applications[app.ID] = app
	return s.Save()
}

func (s *Store) HasActiveApplication(serialNumber string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, app := range s.applications {
		if app.SerialNumber == serialNumber {
			if app.Status == common.ApplicationStatusPending ||
				app.Status == common.ApplicationStatusAutoApproved ||
				app.Status == common.ApplicationStatusAppealed {
				return true
			}
		}
	}
	return false
}

func (s *Store) CreateAppeal(appeal common.Appeal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.appeals[appeal.ID] = appeal
	s.appealsByApp[appeal.ApplicationID] = append(s.appealsByApp[appeal.ApplicationID], appeal.ID)
	return s.Save()
}

func (s *Store) GetAppeal(id string) (common.Appeal, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	appeal, exists := s.appeals[id]
	return appeal, exists
}

func (s *Store) GetAppealsByApplication(applicationID string) []common.Appeal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var appeals []common.Appeal
	if appealIDs, exists := s.appealsByApp[applicationID]; exists {
		for _, id := range appealIDs {
			if appeal, exists := s.appeals[id]; exists {
				appeals = append(appeals, appeal)
			}
		}
	}
	return appeals
}

func (s *Store) GetPendingAppeals() []common.Appeal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var appeals []common.Appeal
	for _, appeal := range s.appeals {
		if appeal.Status == common.AppealStatusPending {
			appeals = append(appeals, appeal)
		}
	}
	return appeals
}

func (s *Store) UpdateAppeal(appeal common.Appeal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.appeals[appeal.ID] = appeal
	return s.Save()
}

func (s *Store) GetStatistics() common.StatisticsResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := common.StatisticsResponse{
		Success: true,
	}

	for _, app := range s.applications {
		stats.TotalApplications++
		switch app.Status {
		case common.ApplicationStatusPending:
			stats.PendingApplications++
		case common.ApplicationStatusAutoApproved:
			stats.AutoApprovedApplications++
			stats.PendingApplications++
		case common.ApplicationStatusApproved:
			stats.ApprovedApplications++
		case common.ApplicationStatusRejected:
			stats.RejectedApplications++
		case common.ApplicationStatusAppealed:
			stats.AppealedApplications++
		}
	}

	return stats
}
