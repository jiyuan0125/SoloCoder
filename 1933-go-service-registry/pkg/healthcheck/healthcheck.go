package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"registry/pkg/model"
	"registry/pkg/store"
)

const (
	successThreshold = 3
	failThreshold    = 3
)

type HealthChecker struct {
	store      *store.InstanceStore
	httpClient *http.Client
	tickers    map[string]*time.Ticker
	tickerMu   sync.Mutex
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewHealthChecker(s *store.InstanceStore) *HealthChecker {
	return &HealthChecker{
		store:      s,
		httpClient: &http.Client{},
		tickers:    make(map[string]*time.Ticker),
		stopCh:     make(chan struct{}),
	}
}

func (hc *HealthChecker) Start(inst *model.Instance) {
	hc.tickerMu.Lock()
	if _, exists := hc.tickers[inst.ID]; exists {
		hc.tickerMu.Unlock()
		return
	}

	ticker := time.NewTicker(inst.HealthCheck.Interval)
	hc.tickers[inst.ID] = ticker
	hc.tickerMu.Unlock()

	hc.wg.Add(1)
	go hc.run(inst, ticker)
}

func (hc *HealthChecker) Stop(instanceID string) {
	hc.tickerMu.Lock()
	if ticker, exists := hc.tickers[instanceID]; exists {
		ticker.Stop()
		delete(hc.tickers, instanceID)
	}
	hc.tickerMu.Unlock()
}

func (hc *HealthChecker) StopAll() {
	close(hc.stopCh)
	hc.tickerMu.Lock()
	for _, ticker := range hc.tickers {
		ticker.Stop()
	}
	hc.tickerMu.Unlock()
	hc.wg.Wait()
}

func (hc *HealthChecker) run(inst *model.Instance, ticker *time.Ticker) {
	defer hc.wg.Done()

	for {
		select {
		case <-hc.stopCh:
			return
		case <-ticker.C:
			if stopped := hc.checkInstance(inst); stopped {
				return
			}
		}
	}
}

func (hc *HealthChecker) checkInstance(inst *model.Instance) (stopped bool) {
	status := inst.GetStatus()
	if status == model.StatusOffline {
		hc.Stop(inst.ID)
		return true
	}

	url := fmt.Sprintf("http://%s:%d%s", inst.Address, inst.Port, inst.HealthCheck.Path)

	ctx, cancel := context.WithTimeout(context.Background(), inst.HealthCheck.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		hc.handleFail(inst)
		return false
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		hc.handleFail(inst)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		hc.handleSuccess(inst)
	} else {
		hc.handleFail(inst)
	}

	return false
}

func (hc *HealthChecker) handleSuccess(inst *model.Instance) {
	inst.HealthSuccess()

	if inst.GetHealthSuccessCount() >= successThreshold {
		currentStatus := inst.GetStatus()
		if currentStatus == model.StatusPending {
			inst.SetStatus(model.StatusCanary)
		}
	}
}

func (hc *HealthChecker) handleFail(inst *model.Instance) {
	inst.HealthFail()

	if inst.GetHealthFailCount() >= failThreshold {
		currentStatus := inst.GetStatus()
		if currentStatus == model.StatusCanary || currentStatus == model.StatusActive {
			inst.SetStatus(model.StatusDraining)
		}
	}
}
