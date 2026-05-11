package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrFenceOverlap    = errors.New("fence overlaps with existing fence")
	ErrFenceNotFound   = errors.New("fence not found")
	ErrInvalidFence    = errors.New("invalid fence coordinates")
	ErrFenceHasBikes   = errors.New("fence has bikes, cannot delete")
)

func (s *Store) CreateFence(name string, fenceType FenceType, minLat, maxLat, minLng, maxLng float64, capacity int) (*Fence, error) {
	if minLat >= maxLat || minLng >= maxLng {
		return nil, ErrInvalidFence
	}

	newFence := &Fence{
		MinLat: minLat,
		MaxLat: maxLat,
		MinLng: minLng,
		MaxLng: maxLng,
	}

	existingFences := s.GetFences()
	for _, ef := range existingFences {
		if rectanglesOverlap(ef, newFence) {
			return nil, ErrFenceOverlap
		}
	}

	fence := &Fence{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      fenceType,
		MinLat:    minLat,
		MaxLat:    maxLat,
		MinLng:    minLng,
		MaxLng:    maxLng,
		Capacity:  capacity,
		CreatedAt: time.Now(),
	}

	s.saveFence(fence)
	return fence, nil
}

func rectanglesOverlap(a, b *Fence) bool {
	return !(a.MaxLat <= b.MinLat || a.MinLat >= b.MaxLat ||
		a.MaxLng <= b.MinLng || a.MinLng >= b.MaxLng)
}

func (s *Store) GetFenceByPoint(lat, lng float64) *Fence {
	fences := s.GetFences()
	for _, f := range fences {
		if lat >= f.MinLat && lat <= f.MaxLat && lng >= f.MinLng && lng <= f.MaxLng {
			return f
		}
	}
	return nil
}

func (s *Store) UpdateFenceCapacity(fenceID string, capacity int) (*Fence, error) {
	fence := s.GetFence(fenceID)
	if fence == nil {
		return nil, ErrFenceNotFound
	}
	fence.Capacity = capacity
	s.saveFence(fence)
	return fence, nil
}

func (s *Store) DeleteFence(fenceID string) error {
	fence := s.GetFence(fenceID)
	if fence == nil {
		return ErrFenceNotFound
	}

	bikes := s.GetBikes()
	for _, b := range bikes {
		if b.FenceID == fenceID {
			return ErrFenceHasBikes
		}
	}

	s.removeFence(fenceID)
	return nil
}
