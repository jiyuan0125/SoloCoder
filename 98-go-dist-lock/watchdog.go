package distlock

import (
	"sync"
	"time"
)

const maxConsecutiveFailures = 2

type RenewalFailureCallback func(key string)

type Watchdog struct {
	key           string
	lease         *Lease
	interval      time.Duration
	stopChan      chan struct{}
	failureCount  int
	onFailure     RenewalFailureCallback
	mu            sync.Mutex
	running       bool
}

func NewWatchdog(key string, lease *Lease, onFailure RenewalFailureCallback) *Watchdog {
	interval := lease.LeaseTime() / 3
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	
	return &Watchdog{
		key:       key,
		lease:     lease,
		interval:  interval,
		stopChan:  make(chan struct{}),
		onFailure: onFailure,
		running:   false,
	}
}

func (w *Watchdog) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.failureCount = 0
	w.mu.Unlock()
	
	go w.run()
}

func (w *Watchdog) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := w.lease.Renew(); err != nil {
				w.mu.Lock()
				w.failureCount++
				count := w.failureCount
				w.mu.Unlock()
				
				if count >= maxConsecutiveFailures {
					if w.onFailure != nil {
						w.onFailure(w.key)
					}
					return
				}
			} else {
				w.mu.Lock()
				w.failureCount = 0
				w.mu.Unlock()
			}
			
			if w.lease.IsExpired() {
				if w.onFailure != nil {
					w.onFailure(w.key)
				}
				return
			}
		case <-w.stopChan:
			return
		}
	}
}

func (w *Watchdog) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if !w.running {
		return
	}
	
	select {
	case <-w.stopChan:
	default:
		close(w.stopChan)
	}
	w.running = false
}

func (w *Watchdog) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}
