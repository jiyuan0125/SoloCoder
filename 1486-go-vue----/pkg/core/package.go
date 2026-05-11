package core

import (
	"errors"
	"locker/pkg/api"
	"sort"
	"time"
)

func (s *System) StorePackage(req api.StorePackageRequest) (*api.StorePackageResponse, error) {
	s.mu.RLock()
	locker, lockerExists := s.Lockers[req.LockerID]
	courier, courierExists := s.Couriers[req.CourierID]
	s.mu.RUnlock()

	if !lockerExists {
		return nil, errors.New("locker not found")
	}
	if !courierExists {
		return nil, errors.New("courier not found")
	}

	var price float64
	switch req.Size {
	case api.SizeSmall:
		price = SmallPrice
	case api.SizeMedium:
		price = MediumPrice
	case api.SizeLarge:
		price = LargePrice
	default:
		return nil, errors.New("invalid locker size")
	}

	courier.mu.Lock()
	if courier.Balance < price {
		courier.mu.Unlock()
		return nil, errors.New("insufficient balance")
	}
	courier.mu.Unlock()

	locker.mu.Lock()
	defer locker.mu.Unlock()

	compartment, err := s.findAvailableCompartment(locker, req.Size, req.CourierID)
	if err != nil {
		return nil, err
	}

	compartment.mu.Lock()
	defer compartment.mu.Unlock()

	if compartment.Status != api.StatusFree {
		return nil, errors.New("compartment already occupied")
	}

	courier.mu.Lock()
	courier.Balance -= price
	if courier.LastUsed == nil {
		courier.LastUsed = make(map[string][]string)
	}
	lockerKey := req.LockerID
	courier.LastUsed[lockerKey] = append(courier.LastUsed[lockerKey], compartment.ID)
	if len(courier.LastUsed[lockerKey]) > MaxConsecutiveUse {
		courier.LastUsed[lockerKey] = courier.LastUsed[lockerKey][1:]
	}
	courier.mu.Unlock()

	now := time.Now()
	var expireTime time.Time
	if locker.isBusinessHours(now) {
		expireTime = now.Add(PickupValidHours * time.Hour)
	} else {
		nextStart := locker.nextBusinessStart(now)
		expireTime = nextStart.Add(PickupValidHours * time.Hour)
	}

	compartment.Status = api.StatusOccupied
	compartment.CourierID = req.CourierID
	compartment.RecipientPhone = req.RecipientPhone
	compartment.PickupCode = generatePickupCode()
	compartment.StoreTime = now
	compartment.ExpireTime = expireTime

	return &api.StorePackageResponse{
		CompartmentID: compartment.ID,
		PickupCode:    compartment.PickupCode,
	}, nil
}

func (s *System) findAvailableCompartment(locker *Locker, size api.LockerSize, courierID string) (*Compartment, error) {
	ids := make([]string, 0)
	for id, c := range locker.Compartments {
		if c.Size == size && c.Status == api.StatusFree {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("no available compartment")
	}
	sort.Strings(ids)

	for _, id := range ids {
		if s.canUseCompartment(locker.ID, courierID, id) {
			return locker.Compartments[id], nil
		}
	}

	return nil, errors.New("no available compartment after consecutive use check")
}

func (s *System) canUseCompartment(lockerID, courierID, compartmentID string) bool {
	s.mu.RLock()
	courier, exists := s.Couriers[courierID]
	s.mu.RUnlock()
	if !exists {
		return true
	}

	courier.mu.Lock()
	defer courier.mu.Unlock()

	if courier.LastUsed == nil {
		return true
	}
	history := courier.LastUsed[lockerID]
	if len(history) < MaxConsecutiveUse {
		return true
	}

	for _, cid := range history {
		if cid != compartmentID {
			return true
		}
	}
	return false
}

func (s *System) PickupPackage(req api.PickupPackageRequest) (*api.PickupPackageResponse, error) {
	s.mu.RLock()
	locker, exists := s.Lockers[req.LockerID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("locker not found")
	}

	now := time.Now()
	if !locker.isBusinessHours(now) {
		return nil, errors.New("outside business hours")
	}

	locker.mu.Lock()
	defer locker.mu.Unlock()

	var targetCompartment *Compartment
	for _, c := range locker.Compartments {
		c.mu.Lock()
		if c.Status == api.StatusOccupied && c.PickupCode == req.PickupCode {
			targetCompartment = c
		}
		c.mu.Unlock()
		if targetCompartment != nil {
			break
		}
	}

	if targetCompartment == nil {
		return nil, errors.New("invalid pickup code")
	}

	targetCompartment.mu.Lock()
	defer targetCompartment.mu.Unlock()

	if now.After(targetCompartment.ExpireTime) {
		return nil, errors.New("pickup code expired")
	}

	targetCompartment.Status = api.StatusFree
	targetCompartment.CourierID = ""
	targetCompartment.RecipientPhone = ""
	targetCompartment.PickupCode = ""
	targetCompartment.StoreTime = time.Time{}
	targetCompartment.ExpireTime = time.Time{}

	return &api.PickupPackageResponse{
		CompartmentID: targetCompartment.ID,
		Success:       true,
	}, nil
}
