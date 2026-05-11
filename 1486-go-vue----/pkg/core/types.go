package core

import (
	"locker/pkg/api"
	"sync"
	"time"
)

type Compartment struct {
	ID             string
	Size           api.LockerSize
	Status         api.CompartmentStatus
	CourierID      string
	RecipientPhone string
	PickupCode     string
	StoreTime      time.Time
	ExpireTime     time.Time
	mu             sync.Mutex
}

type Locker struct {
	ID            string
	Address       string
	Compartments  map[string]*Compartment
	BusinessStart time.Time
	BusinessEnd   time.Time
	mu            sync.RWMutex
}

type Courier struct {
	ID        string
	Name      string
	Balance   float64
	LastUsed  map[string][]string
	mu        sync.Mutex
}

type System struct {
	Lockers  map[string]*Locker
	Couriers map[string]*Courier
	Overdue  []OverdueTask
	mu       sync.RWMutex
}

type OverdueTask struct {
	LockerID      string
	CompartmentID string
	CourierID     string
	RecipientPhone string
	StoreTime     time.Time
	ExpireTime    time.Time
	CreatedAt     time.Time
}

const (
	SmallPrice  = 0.3
	MediumPrice = 0.5
	LargePrice  = 0.8
	MaxConsecutiveUse = 3
	PickupValidHours  = 24
)

func NewSystem() *System {
	return &System{
		Lockers:  make(map[string]*Locker),
		Couriers: make(map[string]*Courier),
		Overdue:  []OverdueTask{},
	}
}
