package model

import (
	"hospital-bed/shared/protocol"
	"sync"
	"time"
)

type Bed struct {
	ID             string
	Type           protocol.BedType
	DepartmentName string
	WardName       string
	IsOccupied     bool
	PatientID      *string
	PatientName    *string
	IsExtra        bool
}

type Ward struct {
	Name           string
	DepartmentName string
	Beds           map[string]*Bed
	SingleCount    int
	DoubleCount    int
	TripleCount    int
	ExtraBedLimit  int
	ExtraBedCount  int
}

type Department struct {
	Name  string
	Wards map[string]*Ward
	Queue []*QueueEntry
}

type QueueEntry struct {
	PatientID     string
	PatientName   string
	WardName      *string
	PreferredType protocol.BedType
	QueueTime     time.Time
}

type PatientRecord struct {
	PatientID      string
	PatientName    string
	BedID          *string
	DepartmentName *string
	WardName       *string
	BedType        *protocol.BedType
	InQueue        bool
	QueuePosition  int
}

type SystemState struct {
	Departments map[string]*Department
	Patients    map[string]*PatientRecord
}

type Store struct {
	state SystemState
	mu    sync.RWMutex
}
