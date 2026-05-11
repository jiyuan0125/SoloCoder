package core

import "time"

type CargoType string

const (
	CargoTypeFrozen    CargoType = "frozen"
	CargoTypeRefrigerated CargoType = "refrigerated"
	CargoTypeMedicine  CargoType = "medicine"
	CargoTypeNormal    CargoType = "normal"
)

type SensorData struct {
	DeviceID    string    `json:"device_id"`
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Speed       float64   `json:"speed"`
}

type TransportTask struct {
	ID          string
	DeviceID    string
	CargoType   CargoType
	StartWarehouse string
	EndWarehouse  string
	StartTime   time.Time
	EndTime     *time.Time
	Status      TaskStatus
	TempMin     float64
	TempMax     float64
	HumidityMin float64
	HumidityMax float64
}

type TaskStatus string

const (
	TaskStatusActive  TaskStatus = "active"
	TaskStatusEnded   TaskStatus = "ended"
)

type Alert struct {
	ID          string
	TaskID      string
	Timestamp   time.Time
	Temperature float64
	Level       AlertLevel
}

type AlertLevel string

const (
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

type TaskStats struct {
	TaskID          string
	TotalSamples    int
	ValidSamples    int
	PassRate        float64
	NeedRecheck     bool
	TempMax         float64
	TempMin         float64
	TempAvg         float64
	TempStdDev      float64
}

type TaskReport struct {
	TaskID     string
	Stats      TaskStats
	Alerts     []Alert
	ExportedAt time.Time
}
