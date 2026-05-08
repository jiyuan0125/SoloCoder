package core

import (
	"context"
	"sync"
	"time"

	"github.com/health-aggregator/pkg/common"
)

type Aggregator struct {
	store       *StateStore
	concurrency int
}

func NewAggregator(store *StateStore) *Aggregator {
	return &Aggregator{
		store:       store,
		concurrency: store.GetGlobalConfig().Concurrency,
	}
}

func (a *Aggregator) CalculateAggregate() *common.AggregateStatus {
	statuses := a.store.GetAllStatuses()
	total := len(statuses)

	healthyCount := 0
	unhealthyCount := 0
	for _, s := range statuses {
		if s.Healthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	groupMap := make(map[string][]*common.ServiceStatus)
	for _, s := range statuses {
		group := s.Group
		if group == "" {
			group = "default"
		}
		groupMap[group] = append(groupMap[group], s)
	}

	groupAggregates := make(map[string]*common.GroupAggregate)
	for group, services := range groupMap {
		groupHealthy := 0
		groupUnhealthy := 0
		for _, s := range services {
			if s.Healthy {
				groupHealthy++
			} else {
				groupUnhealthy++
			}
		}
		groupAggregates[group] = calculateGroupStatus(group, services, groupHealthy, groupUnhealthy, a.store.GetGlobalConfig().UnhealthyThreshold)
	}

	var overallStatus common.HealthStatus
	if total == 0 {
		overallStatus = common.StatusHealthy
	} else if unhealthyCount == 0 {
		overallStatus = common.StatusHealthy
	} else {
		unhealthyRatio := float64(unhealthyCount) / float64(total)
		if unhealthyRatio >= a.store.GetGlobalConfig().UnhealthyThreshold {
			overallStatus = common.StatusUnhealthy
		} else {
			overallStatus = common.StatusDegraded
		}
	}

	return &common.AggregateStatus{
		OverallStatus:   overallStatus,
		TotalServices:   total,
		HealthyCount:    healthyCount,
		UnhealthyCount:  unhealthyCount,
		ServiceStatuses: statuses,
		GroupAggregates: groupAggregates,
	}
}

func calculateGroupStatus(groupName string, services []*common.ServiceStatus, healthy, unhealthy int, threshold float64) *common.GroupAggregate {
	total := len(services)
	var status common.HealthStatus
	if total == 0 {
		status = common.StatusHealthy
	} else if unhealthy == 0 {
		status = common.StatusHealthy
	} else {
		ratio := float64(unhealthy) / float64(total)
		if ratio >= threshold {
			status = common.StatusUnhealthy
		} else {
			status = common.StatusDegraded
		}
	}
	return &common.GroupAggregate{
		GroupName:      groupName,
		OverallStatus:  status,
		TotalServices:  total,
		HealthyCount:   healthy,
		UnhealthyCount: unhealthy,
	}
}

type Scheduler struct {
	store      *StateStore
	aggregator *Aggregator
	stopCh     chan struct{}
	wg         sync.WaitGroup
	running    bool
	mu         sync.Mutex
}

func NewScheduler(store *StateStore) *Scheduler {
	return &Scheduler{
		store:      store,
		aggregator: NewAggregator(store),
		stopCh:     make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	s.runAllOnce()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastProbe := make(map[string]time.Time)

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			names := s.store.GetAllServiceNames()
			toProbe := make([]string, 0)

			for _, name := range names {
				cfg, ok := s.store.GetServiceConfig(name)
				if !ok {
					continue
				}
				last, exists := lastProbe[name]
				if !exists || now.Sub(last) >= cfg.Interval {
					toProbe = append(toProbe, name)
					lastProbe[name] = now
				}
			}

			if len(toProbe) > 0 {
				s.probeMultiple(toProbe)
			}
		}
	}
}

func (s *Scheduler) runAllOnce() {
	names := s.store.GetAllServiceNames()
	if len(names) > 0 {
		s.probeMultiple(names)
	}
}

func (s *Scheduler) probeMultiple(names []string) {
	concurrency := s.store.GetGlobalConfig().Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, name := range names {
		wg.Add(1)
		sem <- struct{}{}
		go func(serviceName string) {
			defer wg.Done()
			defer func() { <-sem }()
			s.probeOne(serviceName)
		}(name)
	}

	wg.Wait()
}

func (s *Scheduler) probeOne(name string) {
	cfg, ok := s.store.GetServiceConfig(name)
	if !ok {
		return
	}

	prober := s.store.GetProber(cfg.ProbeType)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout+1*time.Second)
	defer cancel()

	result := prober.Probe(ctx, cfg)
	s.store.UpdateServiceStatus(name, result)
}

func (s *Scheduler) GetAggregator() *Aggregator {
	return s.aggregator
}

func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
