package core

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	OverCapacityThreshold  = 1.50
	UnderCapacityThreshold = 0.30
	IdleHoursThreshold     = 72
)

type DispatchService struct {
	store *Store
}

func NewDispatchService(store *Store) *DispatchService {
	return &DispatchService{store: store}
}

type FenceCapacityStatus struct {
	FenceID    string
	FenceName  string
	Current    int
	Capacity   int
	Ratio      float64
	Status     string
	NeedBikes  int
	ExcessBikes int
}

func (ds *DispatchService) AnalyzeCapacity() []FenceCapacityStatus {
	fences := ds.store.GetFences()
	result := make([]FenceCapacityStatus, 0)

	for _, f := range fences {
		if f.Type != FenceTypeRecommended && f.Type != FenceTypeDispatch {
			continue
		}

		current := ds.store.CountBikesInFence(f.ID)
		capacity := f.Capacity
		if capacity <= 0 {
			capacity = 1
		}

		ratio := float64(current) / float64(capacity)

		status := "normal"
		needBikes := 0
		excessBikes := 0

		if ratio > OverCapacityThreshold {
			status = "over"
			excessBikes = current - int(float64(capacity)*OverCapacityThreshold)
		} else if ratio < UnderCapacityThreshold {
			status = "under"
			needBikes = int(float64(capacity)*UnderCapacityThreshold) - current
		}

		result = append(result, FenceCapacityStatus{
			FenceID:     f.ID,
			FenceName:   f.Name,
			Current:     current,
			Capacity:    capacity,
			Ratio:       ratio,
			Status:      status,
			NeedBikes:   needBikes,
			ExcessBikes: excessBikes,
		})
	}

	return result
}

func (ds *DispatchService) GenerateDispatchTasks() []*DispatchTask {
	statuses := ds.AnalyzeCapacity()

	var overFences, underFences []FenceCapacityStatus
	for _, s := range statuses {
		if s.Status == "over" {
			overFences = append(overFences, s)
		} else if s.Status == "under" {
			underFences = append(underFences, s)
		}
	}

	sort.Slice(overFences, func(i, j int) bool {
		return overFences[i].Ratio > overFences[j].Ratio
	})
	sort.Slice(underFences, func(i, j int) bool {
		return underFences[i].Ratio < underFences[j].Ratio
	})

	inactiveBikes := ds.store.GetInactiveBikes(time.Hour * IdleHoursThreshold)
	inactiveMap := make(map[string]bool)
	for _, b := range inactiveBikes {
		inactiveMap[b.ID] = true
	}

	tasks := make([]*DispatchTask, 0)

	for _, over := range overFences {
		bikes := ds.store.GetBikes()
		availableInFence := make([]*Bike, 0)

		for _, b := range bikes {
			if b.FenceID == over.FenceID && b.Status == BikeStatusIdle {
				availableInFence = append(availableInFence, b)
			}
		}

		sort.Slice(availableInFence, func(i, j int) bool {
			iInactive := inactiveMap[availableInFence[i].ID]
			jInactive := inactiveMap[availableInFence[j].ID]
			if iInactive && !jInactive {
				return true
			}
			if !iInactive && jInactive {
				return false
			}
			return availableInFence[i].LastUsedAt.Before(availableInFence[j].LastUsedAt)
		})

		remainingExcess := over.ExcessBikes
		bikeIdx := 0

		for remainingExcess > 0 && bikeIdx < len(availableInFence) && len(underFences) > 0 {
			under := &underFences[0]
			if under.NeedBikes <= 0 {
				underFences = underFences[1:]
				continue
			}

			transferCount := min(remainingExcess, under.NeedBikes, len(availableInFence)-bikeIdx)
			transferBikes := make([]string, 0, transferCount)

			for i := 0; i < transferCount; i++ {
				transferBikes = append(transferBikes, availableInFence[bikeIdx+i].ID)
			}

			task := &DispatchTask{
				ID:          uuid.New().String(),
				FromFenceID: over.FenceID,
				ToFenceID:   under.FenceID,
				BikeIDs:     transferBikes,
				Reason:      "capacity_balance",
				CreatedAt:   time.Now(),
			}
			ds.store.saveTask(task)
			tasks = append(tasks, task)

			remainingExcess -= transferCount
			under.NeedBikes -= transferCount
			bikeIdx += transferCount
		}
	}

	return tasks
}

func min(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func (ds *DispatchService) CompleteTask(taskID string) (*DispatchTask, error) {
	task := ds.store.GetTask(taskID)
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.CompletedAt != nil {
		return nil, errors.New("task already completed")
	}

	now := time.Now()
	task.CompletedAt = &now
	ds.store.saveTask(task)

	return task, nil
}

func (ds *DispatchService) GetPendingTasks() []*DispatchTask {
	tasks := ds.store.GetTasks()
	result := make([]*DispatchTask, 0)
	for _, t := range tasks {
		if t.CompletedAt == nil {
			result = append(result, t)
		}
	}
	return result
}
