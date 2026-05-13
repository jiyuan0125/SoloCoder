package registry

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"serviceregistry/internal/events"
	"serviceregistry/internal/model"
	"serviceregistry/internal/storage"
)

var (
	ErrInstanceExists = errors.New("instance already exists")
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("not found")
)

type Registry struct {
	storage   *storage.SQLiteStorage
	eventBus  *events.EventBus
	instances map[string]*model.ServiceInstance
	lock      sync.RWMutex
}

func NewRegistry(storage *storage.SQLiteStorage, eventBus *events.EventBus) *Registry {
	return &Registry{
		storage:   storage,
		eventBus:  eventBus,
		instances: make(map[string]*model.ServiceInstance),
	}
}

func (r *Registry) Restore() error {
	instances, err := r.storage.GetAllInstances()
	if err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, inst := range instances {
		r.instances[inst.ID] = inst
	}

	return nil
}

func (r *Registry) Register(inst *model.ServiceInstance, resource *model.Resource) error {
	if err := validateInstance(inst); err != nil {
		return ErrInvalidRequest
	}

	inst.ID = generateInstanceID(inst)
	inst.LastHeartbeat = time.Now()
	inst.CreatedAt = time.Now()
	inst.Status = model.StatusHealthy

	if inst.Weight <= 0 {
		inst.Weight = 1
	}

	var existed bool

	err := r.storage.Transactional(func(tx *sql.Tx) error {
		existing, err := r.storage.GetInstanceTx(tx, inst.ID)
		if err != nil {
			return err
		}
		if existing != nil {
			existed = true
			return ErrInstanceExists
		}

		if err := r.storage.SaveInstanceTx(tx, inst); err != nil {
			return err
		}

		if resource != nil {
			if err := r.storage.SaveResourceTx(tx, resource); err != nil {
				return err
			}

			assoc := &model.ResourceAssociation{
				ResourceID: resource.ID,
				TargetID:   inst.ID,
				TargetType: model.ResourceInstance,
				Operation:  "register",
				Timestamp:  time.Now(),
			}
			if err := r.storage.AddAssociationTx(tx, assoc); err != nil {
				return err
			}
		}

		return nil
	})

	if existed {
		return ErrInstanceExists
	}

	if err != nil {
		return err
	}

	r.lock.Lock()
	r.instances[inst.ID] = inst
	r.lock.Unlock()

	r.eventBus.Publish(&model.ServiceEvent{
		Type:      model.EventRegistered,
		Instance:  inst,
		Timestamp: time.Now(),
	})

	return nil
}

func (r *Registry) Unregister(id string, resource *model.Resource) error {
	var inst *model.ServiceInstance
	var existed bool

	err := r.storage.Transactional(func(tx *sql.Tx) error {
		var err error
		inst, err = r.storage.GetInstanceTx(tx, id)
		if err != nil {
			return err
		}
		if inst == nil {
			return nil
		}
		existed = true

		if err := r.storage.DeleteInstanceTx(tx, id); err != nil {
			return err
		}

		if resource != nil {
			if err := r.storage.SaveResourceTx(tx, resource); err != nil {
				return err
			}

			assoc := &model.ResourceAssociation{
				ResourceID: resource.ID,
				TargetID:   id,
				TargetType: model.ResourceInstance,
				Operation:  "unregister",
				Timestamp:  time.Now(),
			}
			if err := r.storage.AddAssociationTx(tx, assoc); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if !existed {
		return nil
	}

	r.lock.Lock()
	delete(r.instances, id)
	r.lock.Unlock()

	if r.eventBus.HasSubscribers(inst.ServiceName) {
		r.eventBus.Publish(&model.ServiceEvent{
			Type:      model.EventUnregistered,
			Instance:  inst,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (r *Registry) GetInstance(id string) (*model.ServiceInstance, error) {
	r.lock.RLock()
	inst, ok := r.instances[id]
	r.lock.RUnlock()

	if ok {
		return inst, nil
	}

	return r.storage.GetInstance(id)
}

func (r *Registry) GetAllInstances() []*model.ServiceInstance {
	r.lock.RLock()
	defer r.lock.RUnlock()

	instances := make([]*model.ServiceInstance, 0, len(r.instances))
	for _, inst := range r.instances {
		instances = append(instances, inst)
	}
	return instances
}

func (r *Registry) GetInstances(serviceName string) []*model.ServiceInstance {
	r.lock.RLock()
	defer r.lock.RUnlock()

	instances := make([]*model.ServiceInstance, 0)
	for _, inst := range r.instances {
		if inst.ServiceName == serviceName {
			instances = append(instances, inst)
		}
	}
	return instances
}

func validateInstance(inst *model.ServiceInstance) error {
	if inst.ServiceName == "" {
		return errors.New("service name is required")
	}
	if inst.Address == "" {
		return errors.New("address is required")
	}
	if inst.Port <= 0 || inst.Port > 65535 {
		return errors.New("invalid port")
	}
	if inst.Version == "" {
		return errors.New("version is required")
	}
	return nil
}

func generateInstanceID(inst *model.ServiceInstance) string {
	return fmt.Sprintf("%s:%s:%d", inst.ServiceName, inst.Address, inst.Port)
}
