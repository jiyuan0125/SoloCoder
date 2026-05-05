package server

import (
	"event-collector/common"
)

type Aggregator struct {
	storage *MemoryStorage
}

func NewAggregator(storage *MemoryStorage) *Aggregator {
	return &Aggregator{
		storage: storage,
	}
}

func (a *Aggregator) Aggregate(eventNames []string, startTime, endTime int64) (*common.AggregationResult, error) {
	events := a.storage.GetEventsByTimeRange(startTime, endTime)

	eventUsers := make(map[string]map[string]struct{})
	eventCounts := make(map[string]int)
	allEventUsers := make(map[string]struct{})

	for _, name := range eventNames {
		eventUsers[name] = make(map[string]struct{})
		eventCounts[name] = 0
	}

	for _, event := range events {
		if _, exists := eventUsers[event.Name]; exists {
			eventUsers[event.Name][event.UserID] = struct{}{}
			eventCounts[event.Name]++
			allEventUsers[event.UserID] = struct{}{}
		}
	}

	individualStats := make([]common.EventStat, 0, len(eventNames))
	for _, name := range eventNames {
		individualStats = append(individualStats, common.EventStat{
			EventName:  name,
			TotalCount: eventCounts[name],
			UserCount:  len(eventUsers[name]),
		})
	}

	overlapUsers := a.findOverlapUsers(eventNames, eventUsers)

	return &common.AggregationResult{
		IndividualStats: individualStats,
		OverlapUsers:    overlapUsers,
		OverlapCount:    len(overlapUsers),
	}, nil
}

func (a *Aggregator) findOverlapUsers(eventNames []string, eventUsers map[string]map[string]struct{}) []string {
	if len(eventNames) == 0 {
		return []string{}
	}

	firstEventUsers := eventUsers[eventNames[0]]
	if len(eventNames) == 1 {
		users := make([]string, 0, len(firstEventUsers))
		for user := range firstEventUsers {
			users = append(users, user)
		}
		return users
	}

	overlapUsers := make([]string, 0)
	for user := range firstEventUsers {
		inAllEvents := true
		for i := 1; i < len(eventNames); i++ {
			if _, exists := eventUsers[eventNames[i]][user]; !exists {
				inAllEvents = false
				break
			}
		}
		if inAllEvents {
			overlapUsers = append(overlapUsers, user)
		}
	}

	return overlapUsers
}
