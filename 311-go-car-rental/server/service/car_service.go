package service

import (
	"errors"
	"time"

	"carrental/common"
	"carrental/server/storage"
)

var (
	ErrBrandRequired   = errors.New("品牌不能为空")
	ErrModelRequired   = errors.New("车型不能为空")
	ErrDailyRateInvalid = errors.New("日租金必须大于0")
	ErrDepositInvalid  = errors.New("押金金额不能为负")
	ErrCarNotFound     = errors.New("车辆不存在")
	ErrInvalidStatus   = errors.New("无效的车辆状态")
)

type CarService struct {
	store storage.Storage
}

func NewCarService(store storage.Storage) *CarService {
	return &CarService{store: store}
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func (s *CarService) CreateCar(req *common.CreateCarRequest) (*common.Car, error) {
	if req.Brand == "" {
		return nil, ErrBrandRequired
	}
	if req.Model == "" {
		return nil, ErrModelRequired
	}
	if req.DailyRate <= 0 {
		return nil, ErrDailyRateInvalid
	}
	if req.Deposit < 0 {
		return nil, ErrDepositInvalid
	}

	car := &common.Car{
		ID:        generateID(),
		Brand:     req.Brand,
		Model:     req.Model,
		DailyRate: req.DailyRate,
		Deposit:   req.Deposit,
		Status:    common.CarStatusAvailable,
		CreatedAt: time.Now(),
	}

	if err := s.store.SaveCar(car); err != nil {
		return nil, err
	}

	return car, nil
}

func (s *CarService) GetCar(id string) (*common.Car, error) {
	return s.store.GetCar(id)
}

func (s *CarService) ListCars() ([]*common.Car, error) {
	return s.store.ListCars()
}

func (s *CarService) UpdateCarStatus(req *common.UpdateCarStatusRequest) (*common.Car, error) {
	car, err := s.store.GetCar(req.CarID)
	if err != nil {
		return nil, err
	}
	if car == nil {
		return nil, ErrCarNotFound
	}

	if req.Status != common.CarStatusAvailable && req.Status != common.CarStatusMaintenance {
		return nil, ErrInvalidStatus
	}

	car.Status = req.Status
	if err := s.store.UpdateCar(car); err != nil {
		return nil, err
	}

	return car, nil
}
