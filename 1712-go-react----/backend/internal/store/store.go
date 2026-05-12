package store

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"health-archive/internal/models"
	"health-archive/internal/utils"

	"github.com/google/uuid"
)

type Store struct {
	residents    map[string]*models.Resident
	checkups     map[string]*models.Checkup
	followups    map[string]*models.FollowupRecord
	families     map[string]*models.Family
	idCardIndex  map[string]string

	mu sync.RWMutex
}

func New() *Store {
	return &Store{
		residents:   make(map[string]*models.Resident),
		checkups:    make(map[string]*models.Checkup),
		followups:   make(map[string]*models.FollowupRecord),
		families:    make(map[string]*models.Family),
		idCardIndex: make(map[string]string),
	}
}

func (s *Store) AddResident(resident *models.Resident) (*models.Resident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idCard := utils.NormalizeIDCard(resident.IDCard)

	if !utils.ValidateIDCard(idCard) {
		return nil, errors.New("invalid ID card format")
	}

	if _, exists := s.idCardIndex[idCard]; exists {
		return nil, errors.New("resident with same ID card already exists")
	}

	genderFromID, err := utils.ExtractGenderFromIDCard(idCard)
	if err != nil {
		return nil, err
	}

	birthDateFromID, err := utils.ExtractBirthDateFromIDCard(idCard)
	if err != nil {
		return nil, err
	}

	resident.IDCard = idCard
	resident.Gender = genderFromID
	resident.BirthDate = birthDateFromID

	if utils.IsUnderage(birthDateFromID) {
		if resident.Guardian == nil || resident.Guardian.Name == "" || resident.Guardian.Phone == "" {
			return nil, errors.New("guardian information is required for minors")
		}
	}

	resident.ID = uuid.New().String()
	resident.CreateDate = time.Now()

	s.residents[resident.ID] = resident
	s.idCardIndex[idCard] = resident.ID

	return resident, nil
}

func (s *Store) GetResident(id string) (*models.Resident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resident, exists := s.residents[id]
	if !exists {
		return nil, errors.New("resident not found")
	}
	return resident, nil
}

func (s *Store) SearchResidents(name, idCard, community string) []*models.Resident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.Resident
	nameLower := strings.ToLower(strings.TrimSpace(name))
	idCardNorm := utils.NormalizeIDCard(idCard)
	communityLower := strings.ToLower(strings.TrimSpace(community))

	for _, resident := range s.residents {
		match := true

		if nameLower != "" && !strings.Contains(strings.ToLower(resident.Name), nameLower) {
			match = false
		}

		if idCardNorm != "" && !strings.Contains(resident.IDCard, idCardNorm) {
			match = false
		}

		if communityLower != "" && !strings.Contains(strings.ToLower(resident.Community), communityLower) {
			match = false
		}

		if match {
			results = append(results, resident)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreateDate.After(results[j].CreateDate)
	})

	return results
}

func (s *Store) AddCheckup(checkup *models.Checkup) (*models.Checkup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.residents[checkup.ResidentID]; !exists {
		return nil, errors.New("resident not found")
	}

	if checkup.Height <= 0 || checkup.Weight <= 0 {
		return nil, errors.New("height and weight must be positive")
	}

	heightMeters := checkup.Height / 100.0
	bmi := checkup.Weight / (heightMeters * heightMeters)
	checkup.BMI = float64(int(bmi*10)) / 10

	checkup.ID = uuid.New().String()
	s.checkups[checkup.ID] = checkup

	return checkup, nil
}

func (s *Store) GetCheckups(residentID string) []*models.Checkup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var checkups []*models.Checkup
	for _, c := range s.checkups {
		if c.ResidentID == residentID {
			checkups = append(checkups, c)
		}
	}

	sort.Slice(checkups, func(i, j int) bool {
		return checkups[i].Date.After(checkups[j].Date)
	})

	return checkups
}

func (s *Store) AddFollowup(followup *models.FollowupRecord) (*models.FollowupRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.residents[followup.ResidentID]; !exists {
		return nil, errors.New("resident not found")
	}

	if followup.Date.After(time.Now()) {
		return nil, errors.New("followup date cannot be in the future")
	}

	switch followup.Disease {
	case models.ChronicHypertension:
		if followup.Hypertension == nil || followup.Hypertension.SystolicBP == 0 || followup.Hypertension.DiastolicBP == 0 {
			return nil, errors.New("hypertension followup requires blood pressure data")
		}
		followup.NextDueDate = followup.Date.AddDate(0, 3, 0)

	case models.ChronicDiabetes:
		if followup.Diabetes == nil {
			return nil, errors.New("diabetes followup requires blood sugar data")
		}
		followup.NextDueDate = followup.Date.AddDate(0, 1, 0)

	case models.ChronicCoronary:
		if followup.Coronary == nil {
			return nil, errors.New("coronary followup requires cardiac function data")
		}
		followup.NextDueDate = followup.Date.AddDate(0, 6, 0)

	case models.ChronicStroke, models.ChronicCOPD:
		followup.NextDueDate = followup.Date.AddDate(0, 3, 0)
	}

	followup.IsOverdue = followup.NextDueDate.Before(time.Now())
	followup.ID = uuid.New().String()
	s.followups[followup.ID] = followup

	return followup, nil
}

func (s *Store) GetFollowups(residentID string, disease models.ChronicDisease) []*models.FollowupRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var followups []*models.FollowupRecord
	for _, f := range s.followups {
		if f.ResidentID == residentID && (disease == "" || f.Disease == disease) {
			f.IsOverdue = f.NextDueDate.Before(time.Now())
			followups = append(followups, f)
		}
	}

	sort.Slice(followups, func(i, j int) bool {
		return followups[i].Date.After(followups[j].Date)
	})

	return followups
}

func (s *Store) GetAllChronicPatients() []*models.FollowupRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patientMap := make(map[string]*models.FollowupRecord)

	for _, f := range s.followups {
		key := f.ResidentID + "-" + string(f.Disease)
		if existing, exists := patientMap[key]; !exists || f.Date.After(existing.Date) {
			f.IsOverdue = f.NextDueDate.Before(time.Now())
			patientMap[key] = f
		}
	}

	var results []*models.FollowupRecord
	for _, f := range patientMap {
		results = append(results, f)
	}

	return results
}

func (s *Store) AddFamily(family *models.Family) (*models.Family, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, member := range family.Members {
		if _, exists := s.residents[member.ResidentID]; !exists {
			return nil, errors.New("resident not found: " + member.ResidentID)
		}
	}

	family.ID = uuid.New().String()
	s.families[family.ID] = family

	for _, member := range family.Members {
		if resident, exists := s.residents[member.ResidentID]; exists {
			resident.FamilyID = family.ID
		}
	}

	return family, nil
}

func (s *Store) GetFamily(id string) (*models.Family, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	family, exists := s.families[id]
	if !exists {
		return nil, errors.New("family not found")
	}
	return family, nil
}

func (s *Store) GetAllFamilies() []*models.Family {
	s.mu.RLock()
	defer s.mu.RUnlock()

	families := make([]*models.Family, 0, len(s.families))
	for _, f := range s.families {
		families = append(families, f)
	}
	return families
}

func (s *Store) GetCommunityStats(community string) *models.CommunityStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &models.CommunityStats{
		AgeGroups: make(map[string]int),
	}

	communityLower := strings.ToLower(strings.TrimSpace(community))

	for _, resident := range s.residents {
		if communityLower != "" && !strings.Contains(strings.ToLower(resident.Community), communityLower) {
			continue
		}

		stats.TotalCount++
		if resident.Gender == models.GenderMale {
			stats.MaleCount++
		} else {
			stats.FemaleCount++
		}

		age := utils.CalculateAge(resident.BirthDate)
		var group string
		switch {
		case age < 18:
			group = "0-17"
		case age < 35:
			group = "18-34"
		case age < 50:
			group = "35-49"
		case age < 65:
			group = "50-64"
		default:
			group = "65+"
		}
		stats.AgeGroups[group]++
	}

	return stats
}
