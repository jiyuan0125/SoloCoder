package core

import (
	"errors"
	"locker/pkg/api"
	"time"
)

func (s *System) ScanOverdue(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, locker := range s.Lockers {
		locker.mu.Lock()
		for _, c := range locker.Compartments {
			c.mu.Lock()
			if c.Status == api.StatusOccupied && now.After(c.ExpireTime) {
				c.Status = api.StatusOverdue
				task := OverdueTask{
					LockerID:       locker.ID,
					CompartmentID:  c.ID,
					CourierID:      c.CourierID,
					RecipientPhone: c.RecipientPhone,
					StoreTime:      c.StoreTime,
					ExpireTime:     c.ExpireTime,
					CreatedAt:      now,
				}
				s.Overdue = append(s.Overdue, task)
			}
			c.mu.Unlock()
		}
		locker.mu.Unlock()
	}
	return nil
}

func (s *System) ListOverdue() (*api.ListOverdueResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]api.OverdueInfo, 0, len(s.Overdue))
	for _, task := range s.Overdue {
		result = append(result, api.OverdueInfo{
			LockerID:       task.LockerID,
			CompartmentID:  task.CompartmentID,
			CourierID:      task.CourierID,
			RecipientPhone: task.RecipientPhone,
			StoreTime:      task.StoreTime.Format(time.RFC3339),
			ExpireTime:     task.ExpireTime.Format(time.RFC3339),
		})
	}
	return &api.ListOverdueResponse{OverduePackages: result}, nil
}

func (s *System) HandleOverdue(req api.HandleOverdueRequest) error {
	s.mu.Lock()
	locker, lockerExists := s.Lockers[req.LockerID]
	s.mu.Unlock()

	if !lockerExists {
		return errors.New("locker not found")
	}

	locker.mu.Lock()
	defer locker.mu.Unlock()

	compartment, exists := locker.Compartments[req.CompartmentID]
	if !exists {
		return errors.New("compartment not found")
	}

	compartment.mu.Lock()
	defer compartment.mu.Unlock()

	if compartment.Status != api.StatusOverdue {
		return errors.New("compartment is not overdue")
	}

	if compartment.CourierID != req.CourierID {
		return errors.New("only the original courier can handle this overdue package")
	}

	compartment.Status = api.StatusFree
	compartment.CourierID = ""
	compartment.RecipientPhone = ""
	compartment.PickupCode = ""
	compartment.StoreTime = time.Time{}
	compartment.ExpireTime = time.Time{}

	s.mu.Lock()
	defer s.mu.Unlock()
	newOverdue := []OverdueTask{}
	for _, task := range s.Overdue {
		if task.LockerID != req.LockerID || task.CompartmentID != req.CompartmentID {
			newOverdue = append(newOverdue, task)
		}
	}
	s.Overdue = newOverdue

	return nil
}

func (s *System) StartOverdueScanner() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			wait := next.Sub(now)
			time.Sleep(wait)
			_ = s.ScanOverdue(time.Now())
		}
	}()
}
