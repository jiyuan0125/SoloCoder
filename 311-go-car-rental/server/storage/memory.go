package storage

import (
	"carrental/common"
)

type memoryStorage struct {
	cars   map[string]*common.Car
	orders map[string]*common.Order
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		cars:   make(map[string]*common.Car),
		orders: make(map[string]*common.Order),
	}
}

func (m *memoryStorage) SaveCar(car *common.Car) error {
	m.cars[car.ID] = car
	return nil
}

func (m *memoryStorage) GetCar(id string) (*common.Car, error) {
	car, ok := m.cars[id]
	if !ok {
		return nil, nil
	}
	return car, nil
}

func (m *memoryStorage) ListCars() ([]*common.Car, error) {
	cars := make([]*common.Car, 0, len(m.cars))
	for _, car := range m.cars {
		cars = append(cars, car)
	}
	return cars, nil
}

func (m *memoryStorage) UpdateCar(car *common.Car) error {
	if _, ok := m.cars[car.ID]; !ok {
		return nil
	}
	m.cars[car.ID] = car
	return nil
}

func (m *memoryStorage) SaveOrder(order *common.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *memoryStorage) GetOrder(id string) (*common.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, nil
	}
	return order, nil
}

func (m *memoryStorage) ListOrders() ([]*common.Order, error) {
	orders := make([]*common.Order, 0, len(m.orders))
	for _, order := range m.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (m *memoryStorage) ListOrdersByPhone(phone string) ([]*common.Order, error) {
	orders := make([]*common.Order, 0)
	for _, order := range m.orders {
		if order.UserPhone == phone {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (m *memoryStorage) UpdateOrder(order *common.Order) error {
	if _, ok := m.orders[order.ID]; !ok {
		return nil
	}
	m.orders[order.ID] = order
	return nil
}
