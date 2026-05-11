package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrBikeNotFound       = errors.New("bike not found")
	ErrBikeNotIdle        = errors.New("bike is not idle")
	ErrBikeNotInUse       = errors.New("bike is not in use")
	ErrBikeAlreadyScrapped = errors.New("bike is already scrapped")
)

func (s *Store) AddBike(lat, lng float64) (*Bike, error) {
	fence := s.GetFenceByPoint(lat, lng)
	fenceID := ""
	if fence != nil {
		fenceID = fence.ID
	}

	bike := &Bike{
		ID:         uuid.New().String(),
		Status:     BikeStatusIdle,
		Lat:        lat,
		Lng:        lng,
		FenceID:    fenceID,
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	s.saveBike(bike)
	return bike, nil
}

func (s *Store) GetIdleBikes() []*Bike {
	bikes := s.GetBikes()
	result := make([]*Bike, 0)
	for _, b := range bikes {
		if b.Status == BikeStatusIdle {
			result = append(result, b)
		}
	}
	return result
}

func (s *Store) GetInactiveBikes(hours time.Duration) []*Bike {
	bikes := s.GetBikes()
	cutoff := time.Now().Add(-hours)
	result := make([]*Bike, 0)
	for _, b := range bikes {
		if b.Status == BikeStatusIdle && b.LastUsedAt.Before(cutoff) {
			result = append(result, b)
		}
	}
	return result
}

func (s *Store) UpdateBikeLocation(bikeID string, lat, lng float64) (*Bike, error) {
	bike := s.GetBike(bikeID)
	if bike == nil {
		return nil, ErrBikeNotFound
	}

	bike.Lat = lat
	bike.Lng = lng
	fence := s.GetFenceByPoint(lat, lng)
	if fence != nil {
		bike.FenceID = fence.ID
	} else {
		bike.FenceID = ""
	}

	s.saveBike(bike)
	return bike, nil
}

func (s *Store) SetBikeStatus(bikeID string, status BikeStatus) (*Bike, error) {
	bike := s.GetBike(bikeID)
	if bike == nil {
		return nil, ErrBikeNotFound
	}
	if bike.Status == BikeStatusScrapped {
		return nil, ErrBikeAlreadyScrapped
	}

	bike.Status = status
	s.saveBike(bike)
	return bike, nil
}

func (s *Store) ScrapeBike(bikeID string) (*Bike, error) {
	bike := s.GetBike(bikeID)
	if bike == nil {
		return nil, ErrBikeNotFound
	}
	if bike.Status == BikeStatusScrapped {
		return nil, ErrBikeAlreadyScrapped
	}
	bike.Status = BikeStatusScrapped
	s.saveBike(bike)
	return bike, nil
}

func (s *Store) CountBikesInFence(fenceID string) int {
	bikes := s.GetBikes()
	count := 0
	for _, b := range bikes {
		if b.FenceID == fenceID && b.Status != BikeStatusScrapped {
			count++
		}
	}
	return count
}
