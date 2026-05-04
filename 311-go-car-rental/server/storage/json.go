package storage

import (
	"encoding/json"
	"os"
	"sync"

	"carrental/common"
)

type dataContainer struct {
	Cars   map[string]*common.Car   `json:"cars"`
	Orders map[string]*common.Order `json:"orders"`
}

type JSONStorage struct {
	*memoryStorage
	filePath string
	mu       sync.RWMutex
}

func NewJSONStorage(filePath string) (*JSONStorage, error) {
	s := &JSONStorage{
		memoryStorage: newMemoryStorage(),
		filePath:      filePath,
	}

	if err := s.Load(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *JSONStorage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	var container dataContainer
	if err := json.NewDecoder(file).Decode(&container); err != nil {
		return err
	}

	if container.Cars != nil {
		s.memoryStorage.cars = container.Cars
	}
	if container.Orders != nil {
		s.memoryStorage.orders = container.Orders
	}

	return nil
}

func (s *JSONStorage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	container := dataContainer{
		Cars:   s.memoryStorage.cars,
		Orders: s.memoryStorage.orders,
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(container); err != nil {
		return err
	}

	return nil
}

func (s *JSONStorage) SaveCar(car *common.Car) error {
	if err := s.memoryStorage.SaveCar(car); err != nil {
		return err
	}
	return s.Save()
}

func (s *JSONStorage) UpdateCar(car *common.Car) error {
	if err := s.memoryStorage.UpdateCar(car); err != nil {
		return err
	}
	return s.Save()
}

func (s *JSONStorage) SaveOrder(order *common.Order) error {
	if err := s.memoryStorage.SaveOrder(order); err != nil {
		return err
	}
	return s.Save()
}

func (s *JSONStorage) UpdateOrder(order *common.Order) error {
	if err := s.memoryStorage.UpdateOrder(order); err != nil {
		return err
	}
	return s.Save()
}
