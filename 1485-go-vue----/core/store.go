package core

import (
	"sort"
	"sync"
)

type Store struct {
	mu             sync.RWMutex
	tasks          map[string]*TransportTask
	sensorData     map[string][]*SensorData
	alerts         map[string][]*Alert
	activeTasks    map[string]*TransportTask
}

func NewStore() *Store {
	return &Store{
		tasks:       make(map[string]*TransportTask),
		sensorData:  make(map[string][]*SensorData),
		alerts:      make(map[string][]*Alert),
		activeTasks: make(map[string]*TransportTask),
	}
}

func (s *Store) CreateTask(task *TransportTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[task.ID]; exists {
		return &TaskExistsError{TaskID: task.ID}
	}
	
	s.tasks[task.ID] = task
	s.sensorData[task.ID] = make([]*SensorData, 0)
	s.alerts[task.ID] = make([]*Alert, 0)
	
	if task.Status == TaskStatusActive {
		s.activeTasks[task.DeviceID] = task
	}
	
	return nil
}

func (s *Store) GetTask(taskID string) (*TransportTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return nil, &TaskNotFoundError{TaskID: taskID}
	}
	
	return task, nil
}

func (s *Store) GetActiveTaskByDevice(deviceID string) (*TransportTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.activeTasks[deviceID]
	return task, exists
}

func (s *Store) EndTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return &TaskNotFoundError{TaskID: taskID}
	}
	
	if task.Status == TaskStatusEnded {
		return &TaskAlreadyEndedError{TaskID: taskID}
	}
	
	task.Status = TaskStatusEnded
	delete(s.activeTasks, task.DeviceID)
	
	return nil
}

func (s *Store) AddSensorData(taskID string, data *SensorData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.sensorData[taskID] = append(s.sensorData[taskID], data)
	
	sort.Slice(s.sensorData[taskID], func(i, j int) bool {
		return s.sensorData[taskID][i].Timestamp.Before(s.sensorData[taskID][j].Timestamp)
	})
}

func (s *Store) GetSensorData(taskID string) ([]*SensorData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	data, exists := s.sensorData[taskID]
	if !exists {
		return nil, &TaskNotFoundError{TaskID: taskID}
	}
	
	result := make([]*SensorData, len(data))
	copy(result, data)
	return result, nil
}

func (s *Store) GetRecentSensorData(taskID string, count int) ([]*SensorData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	data, exists := s.sensorData[taskID]
	if !exists {
		return nil, &TaskNotFoundError{TaskID: taskID}
	}
	
	start := 0
	if len(data) > count {
		start = len(data) - count
	}
	
	result := make([]*SensorData, len(data)-start)
	copy(result, data[start:])
	return result, nil
}

func (s *Store) AddAlert(taskID string, alert *Alert) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, existing := range s.alerts[taskID] {
		if existing.Timestamp.Equal(alert.Timestamp) {
			return false
		}
	}
	
	s.alerts[taskID] = append(s.alerts[taskID], alert)
	return true
}

func (s *Store) GetAlerts(taskID string) ([]*Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	alerts, exists := s.alerts[taskID]
	if !exists {
		return nil, &TaskNotFoundError{TaskID: taskID}
	}
	
	result := make([]*Alert, len(alerts))
	copy(result, alerts)
	return result, nil
}

func (s *Store) ListTasks() []*TransportTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*TransportTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		result = append(result, task)
	}
	return result
}
