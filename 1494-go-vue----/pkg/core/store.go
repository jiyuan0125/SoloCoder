package core

import (
	"sync"
)

type Store interface {
	GetZones() []*Zone
	GetZone(id string) (*Zone, bool)
	SaveZone(zone *Zone)

	GetPlants() []*Plant
	GetPlant(id string) (*Plant, bool)
	GetPlantsByZone(zoneID string) []*Plant
	SavePlant(plant *Plant)
	DeletePlant(id string)

	GetWorkers() []*Worker
	GetWorker(id string) (*Worker, bool)
	GetWorkersByZone(zoneID string) []*Worker
	SaveWorker(worker *Worker)

	GetMaintenancePlans() []*MaintenancePlan
	GetMaintenancePlan(id string) (*MaintenancePlan, bool)
	GetPlansByZoneAndType(zoneID string, plantType string) []*MaintenancePlan
	SaveMaintenancePlan(plan *MaintenancePlan)

	GetTasks() []*Task
	GetTask(id string) (*Task, bool)
	GetTasksByZone(zoneID string) []*Task
	GetTasksByWorker(workerID string) []*Task
	SaveTask(task *Task)
}

type InMemoryStore struct {
	zones             sync.Map
	plants            sync.Map
	workers           sync.Map
	maintenancePlans  sync.Map
	tasks             sync.Map
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (s *InMemoryStore) GetZones() []*Zone {
	var zones []*Zone
	s.zones.Range(func(_, value interface{}) bool {
		zones = append(zones, value.(*Zone))
		return true
	})
	return zones
}

func (s *InMemoryStore) GetZone(id string) (*Zone, bool) {
	value, ok := s.zones.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*Zone), true
}

func (s *InMemoryStore) SaveZone(zone *Zone) {
	s.zones.Store(zone.ID, zone)
}

func (s *InMemoryStore) GetPlants() []*Plant {
	var plants []*Plant
	s.plants.Range(func(_, value interface{}) bool {
		plants = append(plants, value.(*Plant))
		return true
	})
	return plants
}

func (s *InMemoryStore) GetPlant(id string) (*Plant, bool) {
	value, ok := s.plants.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*Plant), true
}

func (s *InMemoryStore) GetPlantsByZone(zoneID string) []*Plant {
	var plants []*Plant
	s.plants.Range(func(_, value interface{}) bool {
		plant := value.(*Plant)
		if plant.ZoneID == zoneID {
			plants = append(plants, plant)
		}
		return true
	})
	return plants
}

func (s *InMemoryStore) SavePlant(plant *Plant) {
	s.plants.Store(plant.ID, plant)
}

func (s *InMemoryStore) DeletePlant(id string) {
	s.plants.Delete(id)
}

func (s *InMemoryStore) GetWorkers() []*Worker {
	var workers []*Worker
	s.workers.Range(func(_, value interface{}) bool {
		workers = append(workers, value.(*Worker))
		return true
	})
	return workers
}

func (s *InMemoryStore) GetWorker(id string) (*Worker, bool) {
	value, ok := s.workers.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*Worker), true
}

func (s *InMemoryStore) GetWorkersByZone(zoneID string) []*Worker {
	var workers []*Worker
	s.workers.Range(func(_, value interface{}) bool {
		worker := value.(*Worker)
		if worker.ZoneID == zoneID {
			workers = append(workers, worker)
		}
		return true
	})
	return workers
}

func (s *InMemoryStore) SaveWorker(worker *Worker) {
	s.workers.Store(worker.ID, worker)
}

func (s *InMemoryStore) GetMaintenancePlans() []*MaintenancePlan {
	var plans []*MaintenancePlan
	s.maintenancePlans.Range(func(_, value interface{}) bool {
		plans = append(plans, value.(*MaintenancePlan))
		return true
	})
	return plans
}

func (s *InMemoryStore) GetMaintenancePlan(id string) (*MaintenancePlan, bool) {
	value, ok := s.maintenancePlans.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*MaintenancePlan), true
}

func (s *InMemoryStore) GetPlansByZoneAndType(zoneID string, plantType string) []*MaintenancePlan {
	var plans []*MaintenancePlan
	s.maintenancePlans.Range(func(_, value interface{}) bool {
		plan := value.(*MaintenancePlan)
		if plan.ZoneID == zoneID && string(plan.PlantType) == plantType {
			plans = append(plans, plan)
		}
		return true
	})
	return plans
}

func (s *InMemoryStore) SaveMaintenancePlan(plan *MaintenancePlan) {
	s.maintenancePlans.Store(plan.ID, plan)
}

func (s *InMemoryStore) GetTasks() []*Task {
	var tasks []*Task
	s.tasks.Range(func(_, value interface{}) bool {
		tasks = append(tasks, value.(*Task))
		return true
	})
	return tasks
}

func (s *InMemoryStore) GetTask(id string) (*Task, bool) {
	value, ok := s.tasks.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*Task), true
}

func (s *InMemoryStore) GetTasksByZone(zoneID string) []*Task {
	var tasks []*Task
	s.tasks.Range(func(_, value interface{}) bool {
		task := value.(*Task)
		if task.ZoneID == zoneID {
			tasks = append(tasks, task)
		}
		return true
	})
	return tasks
}

func (s *InMemoryStore) GetTasksByWorker(workerID string) []*Task {
	var tasks []*Task
	s.tasks.Range(func(_, value interface{}) bool {
		task := value.(*Task)
		if task.AssignedWorkerID == workerID {
			tasks = append(tasks, task)
		}
		return true
	})
	return tasks
}

func (s *InMemoryStore) SaveTask(task *Task) {
	s.tasks.Store(task.ID, task)
}
