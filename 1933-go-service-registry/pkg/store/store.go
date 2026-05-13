package store

import (
	"errors"
	"sync"

	"registry/pkg/model"
)

type InstanceStore struct {
	mu        sync.RWMutex
	instances map[string]*model.Instance
	byService map[string]map[string]struct{}
}

func NewInstanceStore() *InstanceStore {
	return &InstanceStore{
		instances: make(map[string]*model.Instance),
		byService: make(map[string]map[string]struct{}),
	}
}

func (s *InstanceStore) Add(inst *model.Instance) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.instances[inst.ID] = inst
	if _, ok := s.byService[inst.ServiceName]; !ok {
		s.byService[inst.ServiceName] = make(map[string]struct{})
	}
	s.byService[inst.ServiceName][inst.ID] = struct{}{}
}

func (s *InstanceStore) Get(id string) (*model.Instance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst, ok := s.instances[id]
	return inst, ok
}

func (s *InstanceStore) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if inst, ok := s.instances[id]; ok {
		delete(s.instances, id)
		if svcMap, ok := s.byService[inst.ServiceName]; ok {
			delete(svcMap, id)
			if len(svcMap) == 0 {
				delete(s.byService, inst.ServiceName)
			}
		}
	}
}

func (s *InstanceStore) ListByService(serviceName string, includeCanary bool) []*model.Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Instance
	if ids, ok := s.byService[serviceName]; ok {
		for id := range ids {
			if inst, ok := s.instances[id]; ok {
				status := inst.GetStatus()
				if status == model.StatusActive {
					result = append(result, inst)
				} else if includeCanary && status == model.StatusCanary {
					result = append(result, inst)
				}
			}
		}
	}
	return result
}

func (s *InstanceStore) ListAll() []*model.Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*model.Instance, 0, len(s.instances))
	for _, inst := range s.instances {
		result = append(result, inst)
	}
	return result
}

func (s *InstanceStore) CanPromote(id string) error {
	inst, ok := s.Get(id)
	if !ok {
		return errors.New("instance not found")
	}
	if inst.GetStatus() != model.StatusCanary {
		return errors.New("instance is not in canary status")
	}
	inst.SetStatus(model.StatusActive)
	return nil
}

func (s *InstanceStore) StartDraining(id string) error {
	inst, ok := s.Get(id)
	if !ok {
		return errors.New("instance not found")
	}
	status := inst.GetStatus()
	if status != model.StatusCanary && status != model.StatusActive {
		return errors.New("instance is not in canary or active status")
	}
	inst.SetStatus(model.StatusDraining)
	return nil
}
