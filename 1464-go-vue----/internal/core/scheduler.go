package core

import (
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	tasks   map[string]*scheduledTask
	running bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopCh  chan struct{}

	deviceService     *DeviceService
	inspectionService *InspectionService
	drillService      *DrillService
}

type scheduledTask struct {
	id          string
	name        string
	interval    time.Duration
	handler     func() error
	maxRetries  int
	retryDelays []time.Duration
	lastRun     time.Time
}

func NewScheduler(ds *DeviceService, is *InspectionService, drs *DrillService) *Scheduler {
	return &Scheduler{
		tasks:             make(map[string]*scheduledTask),
		stopCh:            make(chan struct{}),
		deviceService:     ds,
		inspectionService: is,
		drillService:      drs,
	}
}

func (s *Scheduler) RegisterTask(id, name string, interval time.Duration, handler func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[id] = &scheduledTask{
		id:          id,
		name:        name,
		interval:    interval,
		handler:     handler,
		maxRetries:  3,
		retryDelays: []time.Duration{1 * time.Second, 5 * time.Second, 10 * time.Second},
		lastRun:     time.Time{},
	}
}

func (s *Scheduler) RegisterDefaultTasks() {
	s.RegisterTask("daily_inspection", "每日巡检任务生成", 1*time.Hour, s.handleDailyInspection)
	s.RegisterTask("device_expiry", "设备到期提醒检查", 24*time.Hour, s.handleDeviceExpiry)
	s.RegisterTask("quarterly_drill", "季度演练检查", 24*time.Hour, s.handleQuarterlyDrill)
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
	go s.runLoop()
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

func (s *Scheduler) runLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.runDueTasks()
		}
	}
}

func (s *Scheduler) runDueTasks() {
	s.mu.Lock()
	tasks := make([]*scheduledTask, 0, len(s.tasks))
	now := time.Now()
	for _, t := range s.tasks {
		if now.Sub(t.lastRun) >= t.interval {
			tasks = append(tasks, t)
		}
	}
	s.mu.Unlock()

	for _, task := range tasks {
		go s.executeTask(task)
	}
}

func (s *Scheduler) executeTask(task *scheduledTask) {
	s.mu.Lock()
	task.lastRun = time.Now()
	s.mu.Unlock()

	var err error
	for i := 0; i <= task.maxRetries; i++ {
		err = task.handler()
		if err == nil {
			log.Printf("[Scheduler] Task %s executed successfully", task.name)
			return
		}

		if i < task.maxRetries {
			delay := task.retryDelays[i]
			log.Printf("[Scheduler] Task %s failed (attempt %d/%d): %v, retrying in %v",
				task.name, i+1, task.maxRetries+1, err, delay)
			select {
			case <-s.stopCh:
				return
			case <-time.After(delay):
			}
		}
	}

	log.Printf("[Scheduler] Task %s failed after %d attempts: %v", task.name, task.maxRetries+1, err)
}

func (s *Scheduler) handleDailyInspection() error {
	log.Println("[Scheduler] Running daily inspection task generation...")
	return s.inspectionService.GenerateDailyTasks()
}

func (s *Scheduler) handleDeviceExpiry() error {
	log.Println("[Scheduler] Running device expiry check...")
	devices := s.deviceService.CheckExpiringDevices()
	for _, d := range devices {
		_, err := s.deviceService.CreateExpiryReminder(d.ID)
		if err != nil {
			log.Printf("[Scheduler] Failed to create expiry reminder for device %s: %v", d.ID, err)
		}
	}
	return nil
}

func (s *Scheduler) handleQuarterlyDrill() error {
	log.Println("[Scheduler] Running quarterly drill check...")

	now := time.Now()
	year, quarter := getCurrentQuarter(now)
	_, end := getQuarterRange(year, quarter)

	daysUntilEnd := end.Sub(now).Hours() / 24

	if daysUntilEnd <= 7 && daysUntilEnd > 0 {
		hasDrill, err := s.drillService.CheckQuarterlyDrillRequirement()
		if err != nil {
			return err
		}
		if !hasDrill {
			_, err := s.drillService.CreateQuarterlyReminder()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scheduler) TriggerTask(taskID string) error {
	s.mu.Lock()
	task, exists := s.tasks[taskID]
	s.mu.Unlock()

	if !exists {
		return ErrNotFound
	}

	go s.executeTask(task)
	return nil
}
