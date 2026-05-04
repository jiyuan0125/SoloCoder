package server

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"realestate/pkg/common"
)

type Store struct {
	mu         sync.RWMutex
	properties map[string]*common.Property
	favorites  map[string]*common.Favorite
	dataFile   string
}

func NewStore(dataFile string) *Store {
	s := &Store{
		properties: make(map[string]*common.Property),
		favorites:  make(map[string]*common.Favorite),
		dataFile:   dataFile,
	}
	s.load()
	return s
}

type storedData struct {
	Properties []*common.Property `json:"properties"`
	Favorites  []*common.Favorite `json:"favorites"`
}

func (s *Store) load() {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return
	}

	var stored storedData
	if err := json.Unmarshal(data, &stored); err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range stored.Properties {
		s.properties[p.ID] = p
	}
	for _, f := range stored.Favorites {
		s.favorites[f.ID] = f
	}
}

func (s *Store) saveLocked() error {
	var stored storedData
	for _, p := range s.properties {
		stored.Properties = append(stored.Properties, p)
	}
	for _, f := range s.favorites {
		stored.Favorites = append(stored.Favorites, f)
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}

func (s *Store) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveLocked()
}

func (s *Store) AddProperty(p *common.Property) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.properties {
		if existing.LandlordID == p.LandlordID &&
			existing.Community == p.Community &&
			existing.HouseType == p.HouseType &&
			existing.Status != common.StatusSold {
			return ErrorDuplicateProperty
		}
	}

	now := time.Now().Unix()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.Status = common.StatusActive

	s.properties[p.ID] = p
	return s.saveLocked()
}

func (s *Store) GetProperty(id string) (*common.Property, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.properties[id]
	return p, ok
}

func (s *Store) GetPropertiesByLandlord(landlordID string) []*common.Property {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Property
	for _, p := range s.properties {
		if p.LandlordID == landlordID {
			result = append(result, p)
		}
	}
	return result
}

func (s *Store) UpdateProperty(id string, updateFn func(*common.Property) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.properties[id]
	if !ok {
		return ErrorPropertyNotFound
	}

	if p.Status == common.StatusSold {
		return ErrorPropertySold
	}

	if err := updateFn(p); err != nil {
		return err
	}

	p.UpdatedAt = time.Now().Unix()
	return s.saveLocked()
}

func (s *Store) FilterProperties(filter *common.FilterRequest) []*common.Property {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Property
	for _, p := range s.properties {
		if p.Status != common.StatusActive {
			continue
		}

		if !matchPrice(p, filter.MinPrice, filter.MaxPrice) {
			continue
		}

		if !matchArea(p, filter.MinArea, filter.MaxArea) {
			continue
		}

		if !matchHouseType(p, filter.HouseTypes) {
			continue
		}

		if !common.MatchFloorLevel(p.Floor, filter.FloorLevels) {
			continue
		}

		result = append(result, p)
	}
	return result
}

func matchPrice(p *common.Property, minPrice, maxPrice *float64) bool {
	if p.PriceType == common.PriceTypeNegotiable {
		return true
	}

	if minPrice != nil && p.Price < *minPrice {
		return false
	}

	if maxPrice != nil && p.Price > *maxPrice {
		return false
	}

	return true
}

func matchArea(p *common.Property, minArea, maxArea *float64) bool {
	if minArea != nil && p.Area < *minArea {
		return false
	}

	if maxArea != nil && p.Area > *maxArea {
		return false
	}

	return true
}

func matchHouseType(p *common.Property, houseTypes []string) bool {
	if len(houseTypes) == 0 {
		return true
	}

	for _, t := range houseTypes {
		if t == p.HouseType {
			return true
		}
	}
	return false
}

func (s *Store) AddFavorite(f *common.Favorite) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.favorites {
		if existing.UserID == f.UserID && existing.PropertyID == f.PropertyID {
			return ErrorAlreadyFavorited
		}
	}

	now := time.Now().Unix()
	f.CreatedAt = now

	s.favorites[f.ID] = f
	return s.saveLocked()
}

func (s *Store) RemoveFavorite(userID, propertyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, f := range s.favorites {
		if f.UserID == userID && f.PropertyID == propertyID {
			delete(s.favorites, id)
			return s.saveLocked()
		}
	}

	return ErrorFavoriteNotFound
}

func (s *Store) GetFavoritesByUser(userID string) []*common.Favorite {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Favorite
	for _, f := range s.favorites {
		if f.UserID == userID {
			result = append(result, f)
		}
	}
	return result
}

func (s *Store) IsFavorited(userID, propertyID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.favorites {
		if f.UserID == userID && f.PropertyID == propertyID {
			return true
		}
	}
	return false
}
