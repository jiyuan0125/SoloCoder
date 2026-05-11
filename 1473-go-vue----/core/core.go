package core

import "sync"

type Direction int

const (
	DirectionUp Direction = iota
	DirectionDown
)

func (d Direction) String() string {
	if d == DirectionUp {
		return "上行"
	}
	return "下行"
}

type StationType int

const (
	StationTypeNormal StationType = iota
	StationTypeTransfer
	StationTypeTerminal
)

func (t StationType) String() string {
	switch t {
	case StationTypeTransfer:
		return "换乘站"
	case StationTypeTerminal:
		return "首末站"
	default:
		return "普通站"
	}
}

type Station struct {
	Code      string
	Name      string
	Latitude  float64
	Longitude float64
	Type      StationType
}

type Line struct {
	Code          string
	Name          string
	Direction     Direction
	FirstHour     int
	FirstMinute   int
	LastHour      int
	LastMinute    int
	Stations      []*Station
	IsOperational bool
	HasError      bool
	ErrorReason   string
}

type VehicleStatus int

const (
	VehicleStatusIdle VehicleStatus = iota
	VehicleStatusRunning
	VehicleStatusAtStation
	VehicleStatusStopped
)

func (s VehicleStatus) String() string {
	switch s {
	case VehicleStatusRunning:
		return "运行中"
	case VehicleStatusAtStation:
		return "停靠中"
	case VehicleStatusStopped:
		return "已停止"
	default:
		return "空闲"
	}
}

type Vehicle struct {
	ID              string
	Line            *Line
	CurrentStation  int
	Status          VehicleStatus
	Latitude        float64
	Longitude       float64
	StayAtStationAt int64
	Mutex           sync.RWMutex
}

type TransitSystem struct {
	Lines     map[string]*Line
	Stations  map[string]*Station
	Vehicles  map[string]*Vehicle
	StationUsage map[string]map[string]bool
	Mutex     sync.RWMutex
	Scheduler *Scheduler
}

func NewTransitSystem() *TransitSystem {
	return &TransitSystem{
		Lines:        make(map[string]*Line),
		Stations:     make(map[string]*Station),
		Vehicles:     make(map[string]*Vehicle),
		StationUsage: make(map[string]map[string]bool),
	}
}

func (ts *TransitSystem) Start() {
	ts.Scheduler = NewScheduler(ts)
	ts.Scheduler.Start()
}

func (ts *TransitSystem) Stop() {
	if ts.Scheduler != nil {
		ts.Scheduler.Stop()
	}
}
