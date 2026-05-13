package heartbeat

import (
	"context"
	"sync"
	"time"

	"serviceregistry/internal/events"
	"serviceregistry/internal/model"
	"serviceregistry/internal/storage"
)

type Storage interface {
	GetAllInstances() ([]*model.ServiceInstance, error)
	UpdateHeartbeat(id string, t time.Time) error
	UpdateInstanceStatus(id string, status model.InstanceStatus) error
	DeleteInstance(id string) error
	GetInstancesNeedingHeartbeatCheck(now time.Time, threshold time.Duration) ([]*model.ServiceInstance, error)
}

type Checker struct {
	storage   Storage
	eventBus  *events.EventBus
	interval  time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	lock      sync.Mutex
}

func NewChecker(storage Storage, eventBus *events.EventBus) *Checker {
	return &Checker{
		storage:  storage,
		eventBus: eventBus,
		interval: 5 * time.Second,
	}
}

func (c *Checker) Start() {
	c.lock.Lock()
	if c.running {
		c.lock.Unlock()
		return
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.running = true
	c.lock.Unlock()

	c.wg.Add(1)
	go c.run()
}

func (c *Checker) Stop() {
	c.lock.Lock()
	if !c.running {
		c.lock.Unlock()
		return
	}
	c.cancel()
	c.lock.Unlock()
	c.wg.Wait()
}

func (c *Checker) run() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.checkInstances()
		}
	}
}

func (c *Checker) checkInstances() {
	now := time.Now()

	instances, err := c.storage.GetInstancesNeedingHeartbeatCheck(now, model.UnhealthyThreshold)
	if err != nil {
		return
	}

	for _, inst := range instances {
		elapsed := now.Sub(inst.LastHeartbeat)

		if elapsed >= model.TimeoutThreshold {
			c.handleTimeout(inst)
		} else if elapsed >= model.UnhealthyThreshold && inst.Status == model.StatusHealthy {
			c.markUnhealthy(inst)
		}
	}
}

func (c *Checker) markUnhealthy(inst *model.ServiceInstance) {
	if err := c.storage.UpdateInstanceStatus(inst.ID, model.StatusUnhealthy); err != nil {
		return
	}

	c.eventBus.Publish(&model.ServiceEvent{
		Type:      model.EventUnhealthy,
		Instance:  inst,
		Timestamp: time.Now(),
	})
}

func (c *Checker) handleTimeout(inst *model.ServiceInstance) {
	if err := c.storage.DeleteInstance(inst.ID); err != nil {
		return
	}

	if c.eventBus.HasSubscribers(inst.ServiceName) {
		c.eventBus.Publish(&model.ServiceEvent{
			Type:      model.EventUnregistered,
			Instance:  inst,
			Timestamp: time.Now(),
		})
	}
}

func (c *Checker) RecordHeartbeat(id string) error {
	return c.storage.UpdateHeartbeat(id, time.Now())
}

func (c *Checker) CheckAll() error {
	instances, err := c.storage.GetAllInstances()
	if err != nil {
		return err
	}

	now := time.Now()

	for _, inst := range instances {
		elapsed := now.Sub(inst.LastHeartbeat)

		if elapsed >= model.TimeoutThreshold {
			c.handleTimeout(inst)
		} else if elapsed >= model.UnhealthyThreshold {
			if inst.Status == model.StatusHealthy {
				c.markUnhealthy(inst)
			}
		} else {
			if inst.Status != model.StatusHealthy {
				if err := c.storage.UpdateInstanceStatus(inst.ID, model.StatusHealthy); err != nil {
					continue
				}
				c.eventBus.Publish(&model.ServiceEvent{
					Type:      model.EventHealthy,
					Instance:  inst,
					Timestamp: time.Now(),
				})
			}
		}
	}

	return nil
}

type HeartbeatStorage interface {
	Storage
	GetInstance(id string) (*model.ServiceInstance, error)
}

func NewCheckerWithStorage(storage HeartbeatStorage, eventBus *events.EventBus) *Checker {
	return &Checker{
		storage:  storage,
		eventBus: eventBus,
		interval: 5 * time.Second,
	}
}

var _ Storage = (*storage.SQLiteStorage)(nil)
var _ HeartbeatStorage = (*storage.SQLiteStorage)(nil)
