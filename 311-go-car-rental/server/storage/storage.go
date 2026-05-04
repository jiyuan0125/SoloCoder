package storage

import "carrental/common"

type Storage interface {
	SaveCar(car *common.Car) error
	GetCar(id string) (*common.Car, error)
	ListCars() ([]*common.Car, error)
	UpdateCar(car *common.Car) error

	SaveOrder(order *common.Order) error
	GetOrder(id string) (*common.Order, error)
	ListOrders() ([]*common.Order, error)
	ListOrdersByPhone(phone string) ([]*common.Order, error)
	UpdateOrder(order *common.Order) error

	Load() error
	Save() error
}
