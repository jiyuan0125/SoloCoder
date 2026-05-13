package location

import (
	"errors"
	"sort"

	"warehouse-mgmt/internal/models"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateLocation(db *models.Database, code string, capacity int) error {
	if capacity <= 0 {
		return errors.New("capacity must be greater than 0")
	}
	if db.GetLocationByCode(code) != nil {
		return errors.New("location already exists")
	}
	db.Locations = append(db.Locations, models.Location{
		Code:     code,
		Capacity: capacity,
		Used:     0,
	})
	return nil
}

func (s *Service) AllocateLocation(db *models.Database, productID string, quantity int) (string, error) {
	if len(db.Locations) == 0 {
		return "", errors.New("no available locations")
	}

	existingLocs := db.GetLocationsWithProduct(productID)

	var candidates []models.Location
	for _, loc := range db.Locations {
		available := loc.Capacity - loc.Used
		if available >= quantity {
			candidates = append(candidates, loc)
		}
	}

	if len(candidates) == 0 {
		return "", errors.New("no location has enough capacity")
	}

	existingSet := make(map[string]bool)
	for _, code := range existingLocs {
		existingSet[code] = true
	}

	type scored struct {
		loc   models.Location
		score int
	}
	var scoredList []scored
	for _, loc := range candidates {
		score := 0
		if existingSet[loc.Code] {
			score = 2
		} else {
			for existingCode := range existingSet {
				if isAdjacent(existingCode, loc.Code) {
					score = 1
					break
				}
			}
		}
		scoredList = append(scoredList, scored{loc: loc, score: score})
	}

	sort.SliceStable(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		return models.LocationPriority(scoredList[i].loc.Code, scoredList[j].loc.Code)
	})

	return scoredList[0].loc.Code, nil
}

func (s *Service) UpdateLocationUsage(db *models.Database, code string, delta int) error {
	loc := db.GetLocationByCode(code)
	if loc == nil {
		return errors.New("location not found")
	}
	newUsed := loc.Used + delta
	if newUsed < 0 {
		return errors.New("usage cannot be negative")
	}
	if newUsed > loc.Capacity {
		return errors.New("exceeds location capacity")
	}
	loc.Used = newUsed
	return nil
}

func isAdjacent(code1, code2 string) bool {
	return code1[:len(code1)-1] == code2[:len(code2)-1]
}
