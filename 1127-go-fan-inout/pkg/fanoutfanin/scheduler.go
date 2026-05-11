package fanoutfanin

import (
	"errors"
	"sort"
)

var (
	ErrSourceExists   = errors.New("source already exists")
	ErrSourceNotFound = errors.New("source not found")
	ErrSchedulerClosed = errors.New("scheduler is closed")
	ErrInvalidPolicy  = errors.New("invalid backpressure policy")
)

func NewScheduler(config SchedulerConfig) *Scheduler {
	if config.WorkerPoolSize <= 0 {
		config.WorkerPoolSize = 4
	}
	if config.WorkerBufferCap <= 0 {
		config.WorkerBufferCap = 8
	}
	if config.SourceBufferCap <= 0 {
		config.SourceBufferCap = 16
	}
	if config.AggBufferCap <= 0 {
		config.AggBufferCap = 32
	}
	if config.InitialPolicy != PolicyBlock && config.InitialPolicy != PolicyDrop {
		config.InitialPolicy = PolicyBlock
	}

	s := &Scheduler{
		sources:        make(map[string]*SourceInfo),
		aggregated:     make(chan DataItem, config.AggBufferCap),
		policy:         config.InitialPolicy,
		workerPoolSize: config.WorkerPoolSize,
		workerBufCap:   config.WorkerBufferCap,
		sourceBufCap:   config.SourceBufferCap,
		aggBufCap:      config.AggBufferCap,
	}

	s.workers = make([]*Worker, config.WorkerPoolSize)
	for i := 0; i < config.WorkerPoolSize; i++ {
		s.workers[i] = &Worker{
			ID:      i,
			Channel: make(chan DataItem, config.WorkerBufferCap),
		}
	}

	s.wg.Add(1)
	go s.fanOutLoop()

	return s
}

func (s *Scheduler) RegisterSource(name string) (chan<- DataItem, error) {
	if s.closed.Load() {
		return nil, ErrSchedulerClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sources[name]; exists {
		return nil, ErrSourceExists
	}

	ch := make(chan DataItem, s.sourceBufCap)
	info := &SourceInfo{
		Name:      name,
		Channel:   ch,
		BufferCap: s.sourceBufCap,
	}
	info.Active.Store(true)
	s.sources[name] = info

	s.wg.Add(1)
	go s.fanInSource(name, ch, info)

	return ch, nil
}

func (s *Scheduler) UnregisterSource(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.sources[name]
	if !exists {
		return ErrSourceNotFound
	}

	info.Active.Store(false)
	close(info.Channel)
	delete(s.sources, name)

	return nil
}

func (s *Scheduler) SetPolicy(policy BackpressurePolicy) error {
	if policy != PolicyBlock && policy != PolicyDrop {
		return ErrInvalidPolicy
	}
	s.mu.Lock()
	s.policy = policy
	s.mu.Unlock()
	return nil
}

func (s *Scheduler) GetPolicy() BackpressurePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.policy
}

func (s *Scheduler) Workers() []*Worker {
	return s.workers
}

func (s *Scheduler) SourceStats() []SourceStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]SourceStat, 0, len(s.sources))
	for name, info := range s.sources {
		backlog := 0
		if info.Channel != nil {
			backlog = len(info.Channel)
		}
		stats = append(stats, SourceStat{
			Name:        name,
			TotalRecv:   info.TotalRecv.Load(),
			TotalSent:   info.TotalSent.Load(),
			Dropped:     info.Dropped.Load(),
			Active:      info.Active.Load(),
			Backlog:     backlog,
			BufferCap:   info.BufferCap,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Name < stats[j].Name
	})

	return stats
}

func (s *Scheduler) AggregatedOutput() <-chan DataItem {
	return s.aggregated
}

func (s *Scheduler) Close() {
	if !s.closed.CompareAndSwap(false, true) {
		return
	}

	s.mu.Lock()
	for name, info := range s.sources {
		info.Active.Store(false)
		close(info.Channel)
		delete(s.sources, name)
	}
	s.mu.Unlock()

	s.wg.Wait()

	close(s.aggregated)

	for _, w := range s.workers {
		close(w.Channel)
	}
}

func (s *Scheduler) fanInSource(name string, ch <-chan DataItem, info *SourceInfo) {
	defer s.wg.Done()

	for item := range ch {
		info.TotalRecv.Add(1)

		s.mu.RLock()
		policy := s.policy
		s.mu.RUnlock()

		if policy == PolicyDrop {
			select {
			case s.aggregated <- item:
				info.TotalSent.Add(1)
			default:
				info.Dropped.Add(1)
			}
		} else {
			s.aggregated <- item
			info.TotalSent.Add(1)
		}
	}
}

func (s *Scheduler) fanOutLoop() {
	defer s.wg.Done()

	for item := range s.aggregated {
		workerIdx := s.selectLeastLoadedWorker()
		s.workers[workerIdx].Channel <- item
		s.workers[workerIdx].Processed.Add(1)
	}
}

func (s *Scheduler) selectLeastLoadedWorker() int {
	minLoad := 1<<31 - 1
	minIdx := 0

	for i, w := range s.workers {
		load := len(w.Channel)
		if load < minLoad {
			minLoad = load
			minIdx = i
		}
	}

	return minIdx
}

type SourceStat struct {
	Name        string
	TotalRecv   int64
	TotalSent   int64
	Dropped     int64
	Active      bool
	Backlog     int
	BufferCap   int
}

func (w *Worker) Output() <-chan DataItem {
	return w.Channel
}
