package core

import (
	"errors"
	"sort"
	"time"

	"repair-platform/common"
)

const (
	MaxDailyOrders = 6
	MinGapHours    = 1
)

type candidateTech struct {
	tech       common.Technician
	matchScore int
}

func (s *Service) AssignTechnician(reqID string) (string, error) {
	req, ok := s.storage.GetRepairRequest(reqID)
	if !ok {
		return "", errors.New("维修请求不存在")
	}
	if req.Status != common.StatusPending {
		return "", errors.New("该订单状态不允许派单")
	}

	tech, err := s.findBestCandidate(req)
	if err != nil {
		s.storage.PushWaitQueue(reqID)
		return "", err
	}

	now := time.Now()
	s.storage.UpdateTechnician(tech.ID, func(t *common.Technician) {
		t.DailyOrderCount++
		t.LastOrderTime = &now
	})
	s.storage.UpdateRepairRequest(reqID, func(r *common.RepairRequest) {
		r.AssignedTechID = tech.ID
		r.Status = common.StatusAssigned
	})

	return tech.ID, nil
}

func (s *Service) findBestCandidate(req *common.RepairRequest) (*common.Technician, error) {
	allTechs := s.storage.ListTechnicians()
	if len(allTechs) == 0 {
		return nil, errors.New("没有可用师傅")
	}

	var candidates []candidateTech
	now := time.Now()

	for _, t := range allTechs {
		if !specialtyMatches(t.Specialties, req.Appliance.Category) {
			continue
		}
		if !areaMatches(t.ServiceAreas, req.Area) {
			continue
		}
		if t.DailyOrderCount >= MaxDailyOrders {
			continue
		}
		if t.LastOrderTime != nil && !hasEnoughGap(*t.LastOrderTime, now) {
			continue
		}

		score := 0
		if specialtyMatches(t.Specialties, req.Appliance.Category) {
			score += 100
		}
		if areaMatches(t.ServiceAreas, req.Area) {
			score += 50
		}
		candidates = append(candidates, candidateTech{tech: t, matchScore: score})
	}

	if len(candidates) == 0 {
		return nil, errors.New("没有匹配的师傅")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].matchScore > candidates[j].matchScore
	})

	best := candidates[0].tech
	return &best, nil
}

func specialtyMatches(specs []common.ApplianceCategory, cat common.ApplianceCategory) bool {
	for _, s := range specs {
		if s == cat {
			return true
		}
	}
	return false
}

func areaMatches(areas []string, area string) bool {
	for _, a := range areas {
		if a == area {
			return true
		}
	}
	return false
}

func hasEnoughGap(last, now time.Time) bool {
	if !isSameDay(last, now) {
		return true
	}
	return now.Sub(last) >= time.Hour*time.Duration(MinGapHours)
}

func (s *Service) ProcessWaitingQueue() {
	for {
		reqID, ok := s.storage.PopWaitQueue()
		if !ok {
			return
		}
		_, err := s.AssignTechnician(reqID)
		if err != nil {
			s.storage.PushWaitQueue(reqID)
			break
		}
	}
}
