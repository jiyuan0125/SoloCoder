package core

import (
	"firemanagement/pkg/api"
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	devices           map[string]api.Device
	deviceCodes       map[string]struct{}

	inspectionPoints  map[string]api.InspectionPoint
	inspectionRoutes  map[string]api.InspectionRoute
	inspectionPlans   map[string]api.InspectionPlan
	inspectionTasks   map[string]api.InspectionTask

	drillPlans        map[string]api.DrillPlan
	drillRecords      map[string]api.DrillRecord

	reminders         map[string]api.Reminder
}

func NewStore() *Store {
	return &Store{
		devices:          make(map[string]api.Device),
		deviceCodes:      make(map[string]struct{}),
		inspectionPoints: make(map[string]api.InspectionPoint),
		inspectionRoutes: make(map[string]api.InspectionRoute),
		inspectionPlans:  make(map[string]api.InspectionPlan),
		inspectionTasks:  make(map[string]api.InspectionTask),
		drillPlans:       make(map[string]api.DrillPlan),
		drillRecords:     make(map[string]api.DrillRecord),
		reminders:        make(map[string]api.Reminder),
	}
}

func (s *Store) generateID() string {
	return time.Now().Format("20060102150405.000000000")
}

func (s *Store) AddDevice(device api.Device) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.deviceCodes[device.Code]; exists {
		return "", ErrDuplicateDeviceCode
	}

	for _, d := range s.devices {
		if d.Location == device.Location && d.Type == device.Type && d.Status == api.DeviceStatusNormal {
			return "", ErrDuplicateActiveDevice
		}
	}

	id := s.generateID()
	device.ID = id
	s.devices[id] = device
	s.deviceCodes[device.Code] = struct{}{}
	return id, nil
}

func (s *Store) UpdateDevice(id string, updater func(d *api.Device)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, exists := s.devices[id]
	if !exists {
		return ErrNotFound
	}

	oldLocation := d.Location
	oldType := d.Type
	oldStatus := d.Status
	oldCode := d.Code

	updater(&d)

	if d.Code != oldCode {
		if _, exists := s.deviceCodes[d.Code]; exists {
			return ErrDuplicateDeviceCode
		}
		delete(s.deviceCodes, oldCode)
		s.deviceCodes[d.Code] = struct{}{}
	}

	if d.Status == api.DeviceStatusNormal {
		for _, dev := range s.devices {
			if dev.ID != id && dev.Location == d.Location && dev.Type == d.Type && dev.Status == api.DeviceStatusNormal {
				return ErrDuplicateActiveDevice
			}
		}
	}

	_ = oldLocation
	_ = oldType
	_ = oldStatus

	s.devices[id] = d
	return nil
}

func (s *Store) GetDevice(id string) (*api.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, exists := s.devices[id]
	if !exists {
		return nil, false
	}
	cp := d
	return &cp, true
}

func (s *Store) ListDevices(filter func(d api.Device) bool) []api.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []api.Device
	for _, d := range s.devices {
		if filter == nil || filter(d) {
			result = append(result, d)
		}
	}
	return result
}

func (s *Store) GetAllDevices() []api.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	devices := make([]api.Device, 0, len(s.devices))
	for _, d := range s.devices {
		devices = append(devices, d)
	}
	return devices
}

func (s *Store) AddInspectionPoint(point api.InspectionPoint) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	point.ID = id
	s.inspectionPoints[id] = point
	return id
}

func (s *Store) GetInspectionPoint(id string) (*api.InspectionPoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.inspectionPoints[id]
	if !exists {
		return nil, false
	}
	cp := p
	return &cp, true
}

func (s *Store) GetAllInspectionPoints() []api.InspectionPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	points := make([]api.InspectionPoint, 0, len(s.inspectionPoints))
	for _, p := range s.inspectionPoints {
		points = append(points, p)
	}
	return points
}

func (s *Store) AddInspectionRoute(route api.InspectionRoute) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	route.ID = id
	s.inspectionRoutes[id] = route
	return id
}

func (s *Store) GetInspectionRoute(id string) (*api.InspectionRoute, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, exists := s.inspectionRoutes[id]
	if !exists {
		return nil, false
	}
	cp := r
	return &cp, true
}

func (s *Store) GetAllInspectionRoutes() []api.InspectionRoute {
	s.mu.RLock()
	defer s.mu.RUnlock()
	routes := make([]api.InspectionRoute, 0, len(s.inspectionRoutes))
	for _, r := range s.inspectionRoutes {
		routes = append(routes, r)
	}
	return routes
}

func (s *Store) AddInspectionPlan(plan api.InspectionPlan) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	plan.ID = id
	s.inspectionPlans[id] = plan
	return id
}

func (s *Store) GetInspectionPlan(id string) (*api.InspectionPlan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.inspectionPlans[id]
	if !exists {
		return nil, false
	}
	cp := p
	return &cp, true
}

func (s *Store) GetAllInspectionPlans() []api.InspectionPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	plans := make([]api.InspectionPlan, 0, len(s.inspectionPlans))
	for _, p := range s.inspectionPlans {
		plans = append(plans, p)
	}
	return plans
}

func (s *Store) AddInspectionTask(task api.InspectionTask) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	task.ID = id
	s.inspectionTasks[id] = task
	return id
}

func (s *Store) UpdateInspectionTask(id string, updater func(t *api.InspectionTask)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, exists := s.inspectionTasks[id]
	if !exists {
		return ErrNotFound
	}
	updater(&t)
	s.inspectionTasks[id] = t
	return nil
}

func (s *Store) GetInspectionTask(id string) (*api.InspectionTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, exists := s.inspectionTasks[id]
	if !exists {
		return nil, false
	}
	cp := t
	return &cp, true
}

func (s *Store) ListInspectionTasks(filter func(t api.InspectionTask) bool) []api.InspectionTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []api.InspectionTask
	for _, t := range s.inspectionTasks {
		if filter == nil || filter(t) {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) AddDrillPlan(plan api.DrillPlan) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	plan.ID = id
	s.drillPlans[id] = plan
	return id
}

func (s *Store) GetDrillPlan(id string) (*api.DrillPlan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.drillPlans[id]
	if !exists {
		return nil, false
	}
	cp := p
	return &cp, true
}

func (s *Store) GetAllDrillPlans() []api.DrillPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	plans := make([]api.DrillPlan, 0, len(s.drillPlans))
	for _, p := range s.drillPlans {
		plans = append(plans, p)
	}
	return plans
}

func (s *Store) AddDrillRecord(record api.DrillRecord) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	record.ID = id
	s.drillRecords[id] = record
	return id
}

func (s *Store) GetDrillRecord(id string) (*api.DrillRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, exists := s.drillRecords[id]
	if !exists {
		return nil, false
	}
	cp := r
	return &cp, true
}

func (s *Store) ListDrillRecords(filter func(r api.DrillRecord) bool) []api.DrillRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []api.DrillRecord
	for _, r := range s.drillRecords {
		if filter == nil || filter(r) {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) AddReminder(reminder api.Reminder) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.generateID()
	reminder.ID = id
	s.reminders[id] = reminder
	return id
}

func (s *Store) UpdateReminder(id string, updater func(r *api.Reminder)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, exists := s.reminders[id]
	if !exists {
		return ErrNotFound
	}
	updater(&r)
	s.reminders[id] = r
	return nil
}

func (s *Store) GetReminder(id string) (*api.Reminder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, exists := s.reminders[id]
	if !exists {
		return nil, false
	}
	cp := r
	return &cp, true
}

func (s *Store) ListReminders(filter func(r api.Reminder) bool) []api.Reminder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []api.Reminder
	for _, r := range s.reminders {
		if filter == nil || filter(r) {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) ReminderExists(refID string, rType api.ReminderType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.reminders {
		if r.ReferenceID == refID && r.Type == rType {
			return true
		}
	}
	return false
}
