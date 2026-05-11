package core

import (
	"errors"
	"fmt"
	"locker/pkg/api"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *System) CreateLocker(req api.CreateLockerRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Lockers[req.ID]; exists {
		return errors.New("locker already exists")
	}

	start, err := parseBusinessTime(req.BusinessStart)
	if err != nil {
		return err
	}
	end, err := parseBusinessTime(req.BusinessEnd)
	if err != nil {
		return err
	}

	locker := &Locker{
		ID:            req.ID,
		Address:       req.Address,
		Compartments:  make(map[string]*Compartment),
		BusinessStart: start,
		BusinessEnd:   end,
	}

	compartmentIdx := 1
	for i := 0; i < req.SmallCount; i++ {
		locker.Compartments[strconv.Itoa(compartmentIdx)] = &Compartment{
			ID:     strconv.Itoa(compartmentIdx),
			Size:   api.SizeSmall,
			Status: api.StatusFree,
		}
		compartmentIdx++
	}

	for i := 0; i < req.MediumCount; i++ {
		locker.Compartments[strconv.Itoa(compartmentIdx)] = &Compartment{
			ID:     strconv.Itoa(compartmentIdx),
			Size:   api.SizeMedium,
			Status: api.StatusFree,
		}
		compartmentIdx++
	}

	for i := 0; i < req.LargeCount; i++ {
		locker.Compartments[strconv.Itoa(compartmentIdx)] = &Compartment{
			ID:     strconv.Itoa(compartmentIdx),
			Size:   api.SizeLarge,
			Status: api.StatusFree,
		}
		compartmentIdx++
	}

	s.Lockers[req.ID] = locker
	return nil
}

func (s *System) ListCompartments(req api.ListCompartmentsRequest) (*api.ListCompartmentsResponse, error) {
	s.mu.RLock()
	locker, exists := s.Lockers[req.LockerID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("locker not found")
	}

	locker.mu.RLock()
	defer locker.mu.RUnlock()

	ids := make([]string, 0, len(locker.Compartments))
	for id := range locker.Compartments {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]api.CompartmentInfo, 0, len(ids))
	for _, id := range ids {
		c := locker.Compartments[id]
		info := api.CompartmentInfo{
			ID:     c.ID,
			Size:   c.Size,
			Status: c.Status,
		}
		if c.Status != api.StatusFree {
			info.CourierID = c.CourierID
			info.RecipientPhone = c.RecipientPhone
			info.PickupCode = c.PickupCode
			info.StoreTime = c.StoreTime.Format(time.RFC3339)
			info.ExpireTime = c.ExpireTime.Format(time.RFC3339)
		}
		result = append(result, info)
	}

	return &api.ListCompartmentsResponse{Compartments: result}, nil
}

func parseBusinessTime(t string) (time.Time, error) {
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return time.Time{}, errors.New("invalid time format, expected HH:MM")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return time.Time{}, errors.New("invalid hour")
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, errors.New("invalid minute")
	}
	return time.Date(0, 1, 1, hour, minute, 0, 0, time.Local), nil
}

func (l *Locker) isBusinessHours(now time.Time) bool {
	start := time.Date(now.Year(), now.Month(), now.Day(), l.BusinessStart.Hour(), l.BusinessStart.Minute(), 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), l.BusinessEnd.Hour(), l.BusinessEnd.Minute(), 0, 0, now.Location())
	return now.After(start) && now.Before(end)
}

func (l *Locker) nextBusinessStart(now time.Time) time.Time {
	start := time.Date(now.Year(), now.Month(), now.Day(), l.BusinessStart.Hour(), l.BusinessStart.Minute(), 0, 0, now.Location())
	if now.After(start) {
		return start.Add(24 * time.Hour)
	}
	return start
}

func generatePickupCode() string {
	now := time.Now().UnixNano()
	code := (now % 1000000)
	if code < 100000 {
		code += 100000
	}
	return fmt.Sprintf("%06d", code)
}
